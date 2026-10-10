# user 项目：功能实现与调用流转

一个 go-zero 微服务示例，把用户数据访问收敛到一个 RPC 服务里，API 只做协议适配、鉴权和响应组织。

技术栈：`go-zero rest` / `go-zero zrpc` / `gRPC + protobuf` / `PostgreSQL 17` / `GORM Gen` / `Redis` / `etcd` / `Kafka` / `OpenTelemetry + Jaeger`。

---

## 1. 架构与调用链

### 1.1 核心链路

```text
HTTP Client
    │  :8888
    ▼
user-api（api/）
    │  REST 路由、参数校验、JWT、Redis 黑名单、业务权限、双 Token、SSE、文件上传
    │  zRPC 客户端，经 etcd 发现 user.rpc
    ▼
user-rpc（rpc/）
    │  gRPC 服务、业务校验、bcrypt、事务、GORM Gen
    ▼
PostgreSQL :5432
```

**API 不直接访问数据库。** 所有 SQL 都在 RPC 侧，API 只通过 `userclient` 调用 RPC。

### 1.2 组件职责

| 组件 | 作用 | 是否在主链路 |
|---|---|---|
| `user-api` :8888 | HTTP 接口、鉴权、响应组织 | 是 |
| `user-rpc` :8080 | gRPC 服务、唯一的数据访问方 | 是 |
| PostgreSQL :5432 | 用户与待办数据 | 是 |
| Redis :6379 | JWT 黑名单 | 受保护接口需要 |
| etcd :2379 | RPC 服务注册与发现 | 是 |
| Jaeger :16686 | 链路追踪 UI（OTLP 接入 4317） | 否 |
| Kafka :9092 / :29092 | 消息收发演示 | 否 |
| kafka-ui :8081 | 查看 Kafka 消息 | 否 |

### 1.3 服务发现

API 配置里 `UserRpc` 用的是 etcd，不是固定地址：

```yaml
UserRpc:
  Etcd:
    Hosts:
      - 127.0.0.1:2379     # 容器内为 etcd:2379
    Key: user.rpc
#  Endpoints:              # 已注释
#    - 127.0.0.1:8080
```

链路是 `API → etcd → RPC`。RPC 以 `user.rpc` 为 key 注册自身，API 订阅这个 key。

一个副作用值得注意：API 用 `zrpc.MustNewClient` 建客户端，`NonBlock` 默认 false，走 `WithBlock` 且拨号超时硬编码 3 秒。**RPC 尚未在 etcd 注册时，API 会直接 `os.Exit(1)`**，而不是返回业务错误。所以容器编排里 `user-api` 配了 `restart: unless-stopped` 兜底，首次启动重启一两次属正常。

---

## 2. 统一响应与错误码

所有接口共用同一个信封（含成功）：

```json
// 成功
{"code": 0, "msg": "成功", "data": {"id": 38}, "trace_id": "d104c9befb93..."}

// 失败
{"code": 1001, "msg": "用户名已被占用", "trace_id": "6ac3f9507524..."}
```

三条约定：

| 约定 | 说明 |
|---|---|
| HTTP 状态码恒为 200 | 业务结果只由 body 里的 `code` 表达，不用 HTTP 状态码区分业务分支 |
| `trace_id` | 从 OTel SpanContext 取，取不到则省略该字段 |
| `detail` 不出响应 | 内部细节（SQL、堆栈、gRPC 原始错误）通过 `WithDetail` 附加，**只进日志** |

错误在前缀上分段：`1xxx` 用户域，`9xxx` 系统通用。完整字典见 [`api/common/errcode/README.md`](api/common/errcode/README.md)。

### 2.1 单一错误出口

所有错误都汇到一个全局处理器：

```text
Logic 返回 errcode.*
  ↓
Handler: httpx.ErrorCtx(ctx, w, err)
  ↓
SetErrorHandlerCtx 注册的处理器（api/internal/handler/errorhandler.go）
  ↓
errcode.ToError(err) 归一 → response.Fail(ctx, err)
  ↓
HTTP 200 + {"code":..,"msg":..,"trace_id":..}
```

`ToError` 把非业务错误（数据库错误、gRPC 错误）统一收敛成 9001，避免实现细节外泄。

### 2.2 参数错误为什么单独定性

go-zero 的 `httpx.Parse` 返回的是裸错误，不带任何状态标记。若不管它，全局处理器只能按未知故障归到 9001，客户端会误以为是自己修不了的后端问题、不去改请求体。

所以 handler 统一改用 `parse()` 包一层（`errorhandler.go:41`），把解析失败定性成 9002：

```go
func parse(r *http.Request, v any) error {
	if err := httpx.Parse(r, v); err != nil {
		return errcode.ErrBadRequest.WithDetail(err.Error())
	}
	return nil
}
```

路径参数解析失败同理，例如 `GET /user/abc` → 9002。

### 2.3 JWT 未认证也是信封

go-zero 内置的 JWT 中间件是把 401 裸写在 `ResponseWriter` 上的（`rest/handler/authhandler.go` 的 `unauthorized`），**不经过 `httpx.Error`**，所以全局处理器收不到，客户端只能拿到 `Content-Length: 0` 的 401 —— 这不是 JSON，按信封解析 `code` 会直接抛异常。

现在由 `handler.UnauthorizedCallback` 接管，通过服务级选项注册：

```go
server := rest.MustNewServer(c.RestConf,
	rest.WithUnauthorizedCallback(handler.UnauthorizedCallback))
```

注意 `WithUnauthorizedCallback` 是 **RunOption（服务级）**，而 `WithJwt` 是 RouteOption，注册点必须在 `MustNewServer`。go-zero 会把它注入到每个 JWT 路由组。效果：未携带 / 格式错 / 签名错 / 已过期的 token 一律 `HTTP 200 + code 9003`。

---

## 3. 接口清单

13 个 HTTP 接口，按鉴权分三组。

### 公开（7）

| 方法 | 路径 | 职责 | 涉及组件 |
|---|---|---|---|
| `POST` | `/user/login` | 登录，签发双 Token | RPC → PG |
| `POST` | `/user/refresh` | 用 Refresh Token 换新 Token | RPC → PG |
| `GET` | `/user/:id` | 查询用户公开信息 | RPC → PG |
| `POST` | `/users` | 创建用户（场景 B） | RPC → PG |
| `POST` | `/users-with-todo` | 事务内创建用户 + 待办（场景 A） | RPC → PG |
| `POST` | `/users-with-todo/rollback-demo` | 回滚演示（场景 A 反例） | RPC → PG |
| `POST` | `/upload/file` | multipart 文件上传 | 本地磁盘 |

### SSE（2）

| 方法 | 路径 | 职责 |
|---|---|---|
| `POST` | `/sse/with/req` | 带请求体，推 5 条消息 |
| `POST` | `/sse/without/resp` | 不带请求体，推 5 条消息 |

### 受保护（4）

JWT 中间件 + `JwtBlacklist` 中间件同时生效。

| 方法 | 路径 | 职责 | 额外业务校验 |
|---|---|---|---|
| `GET` | `/user/profile/:userId` | 查自己的资料 | JWT `userId` == 路径 `userId` |
| `PUT` | `/user/profile/:userId` | 改自己的用户名 | JWT `userId` == 路径 `userId` |
| `POST` | `/user/:id` | 管理员改任意用户 | JWT `role` == `admin` |
| `POST` | `/user/logout` | 注销当前 Token | — |

---

## 4. 接口调用流转

以下按接口逐个说明从 HTTP 进来到数据落库的完整路径。

### 4.1 登录 `POST /user/login`

```text
POST /user/login {username, password}
  ↓ LoginHandler → parse() 校验参数
  ↓ LoginLogic.Login
  ↓ svc.RpcContext(ctx)  附加 metadata: token=internal
  ↓ userclient.AuthenticateUser  ──gRPC(etcd 发现)──▶
  │   AuthenticateUserLogic
  │   ├ username 为空 → InvalidArgument
  │   └ 按 username 查 users（软删记录自动排除）
  │       ├ 查不到 → NotFound
  │       └ 查到   → 返回 id/username/email/role/**password(哈希)**
  ◀───────────────────────────────────────────────
  ↓ 密码比对在这里：bcrypt.CompareHashAndPassword(hash, 明文)
  │   ├ 不匹配 → 1003
  │   └ RPC 返回 NotFound 也映射成 1003（见下）
  ↓ 签发 Access Token + Refresh Token
  ▼
{"code":0,"data":{"accessToken":"...","refreshToken":"...","expiresIn":86400}}
```

两个设计点：

**密码校验在 API 侧，不在 RPC。** RPC 只负责"按用户名取回认证数据"，把哈希交给 API 比对。

**账号不存在与密码错误返回同一个码 1003。** `mapAuthenticateError` 把 RPC 的 `NotFound` 也翻译成 1003：

```go
func mapAuthenticateError(err error) error {
	if status.Code(err) == codes.NotFound {
		return errcode.ErrUserPasswordInvalid
	}
	return mapRpcError(err)
}
```

否则攻击者能靠错误码枚举出哪些账号已注册。注意它只对 `NotFound` 做这个转换，RPC 抖动等真实故障仍走 `mapRpcError`，不会被伪装成"密码错误"而误导排障方向。

**双 Token 的差别：**

| | Access Token | Refresh Token |
|---|---|---|
| Secret | `Auth.AccessSecret` | `Auth.RefreshSecret` |
| Claims | `userId`, `role`, `iat`, `exp` | `userId`, `iat`, `exp` |
| 有效期 | 硬编码 24h | `Auth.RefreshExpire`（7 天） |

### 4.2 刷新 `POST /user/refresh`

```text
POST /user/refresh {refreshToken}
  ↓ RefreshTokenLogic
  ↓ 用 RefreshSecret 验签解析 → 失败 1005
  ↓ 取 claims["userId"]（带 ok 断言，畸形载荷 → 1005，不 panic）
  ↓ QueryUser(userId)  ──gRPC──▶ QueryUserLogic → PG
  │     取数据库里的**最新 role**（而不是沿用旧 Token 里的 role）
  ↓ 重新签发 Access + Refresh
  ▼ 返回结构与登录一致
```

刷新时重新查库拿 role，意味着用户角色被管理员改过之后，刷新一次就生效，不必等旧 Token 过期。

### 4.3 公开查询 `GET /user/:id`

```text
GET /user/1
  ↓ GetUserInfoHandler → parse()（路径参数 id 转 int64）
  ↓ GetUserInfoLogic
  │    id <= 0 → 9002「用户 ID 无效」
  ↓ QueryUser(id) ──gRPC──▶ QueryUserLogic
  ▼
{"code":0,"data":{"id":1,"username":"...","email":"..."}}
```

只返回 id / username / email。RPC 的 `QueryUserResp` 里带 `role`，但 API 的 `UserInfoResp` 没有这个字段，不会漏出去。

### 4.4 创建用户（场景 B）`POST /users`

```text
POST /users {username, email, password}
  ↓ CreateUserLogic
  │    任一字段为空 → 9002
  ↓ InsertUser ──gRPC──▶ InsertUserLogic
  │   ├ bcrypt.GenerateFromPassword(明文, DefaultCost)
  │   ├ role 为空时默认 "user"
  │   └ INSERT INTO users ... RETURNING id
  │       失败 → mapCreateUserError(err)
  ◀── InsertUserResp{Id}
  ▼
{"code":0,"data":{"id":38}}
```

**id 由数据库生成，API 只透传。** `users.id` 是 `BIGSERIAL`，gorm 发的是 `INSERT ... RETURNING "id"`，把库生成的值回填到 `item.ID`，再经 `InsertUserResp` → `CreateUserResp` 一路返回。

**唯一约束冲突要区分字段。** RPC 侧 `dberr` 解析 PostgreSQL SQLSTATE（23505 唯一冲突 / 23503 外键冲突）和约束名：

```text
idx_users_email    → AlreadyExists「邮箱已被占用」
idx_users_username → AlreadyExists「用户名已被占用」
其他               → Internal「创建失败，请稍后重试」
```

区分出字段名是因为这两条索引是**部分唯一索引**，都带 `WHERE deleted_at IS NULL` 谓词，软删后的用户名和邮箱可以重新使用。

### 4.5 创建用户 + 待办（场景 A）`POST /users-with-todo`

这是唯一一个需要跨表原子性的场景。

```text
POST /users-with-todo {username, email, password, todoTitle}
  ↓ CreateUserWithTodoLogic（API：非空校验 + 转发）
  ↓ CreateUserWithTodo ──gRPC──▶ CreateUserWithTodoLogic
  │
  │   hash = bcrypt(明文)
  │   Query.Transaction(func(tx) error {
  │       ① tx.User.Create(newUser)          ← 拿到自增主键
  │       ② newTodo.UserID = newUser.ID      ← 关键：不能预先知道
  │       ③ tx.Todo.Create(newTodo)
  │   })
  │   任一步失败 → 整体回滚 → mapCreateUserError
  ▼
{"code":0,"data":{"userId":N,"todoId":M,"username":..,"email":..,"todoTitle":..}}
```

关键点是第 ②：待办的 `user_id` 必须取自**事务内**刚插入用户返回的自增主键，否则外键 `fk_todos_user` 无从满足。

### 4.6 回滚演示 `POST /users-with-todo/rollback-demo`

场景 A 的反例，用来证明"用户不会残留"：

```text
Query.Transaction(func(tx) error {
    tx.User.Create(newUser)              ← 第一步正常，id 已分配
    tx.Todo.Create(&Todo{UserID: 0, ...}) ← 故意用不存在的 user_id
                                          ← 外键 fk_todos_user 拒绝
    return err                            ← 事务整体回滚
})
```

注意返回语义：**事务失败是本次演示的预期结果**，所以接口返回 `code:0`，由响应体说明结果：

```json
{"code": 0, "data": {"message": "事务失败，已完成回滚"}}
```

只有在"外键失效导致事务意外提交成功"时才返回 9001。这是刻意把演示结果和真实故障分开。

副作用：回滚**不退还序列值**，每跑一次这个接口会同时烧掉 `users_id_seq` 和 `todos_id_seq` 各一个值。所以 `users.id` 出现空洞是正常现象，不能当"第几个用户"用。

### 4.7 管理员更新 `POST /user/:id`

```text
POST /user/2 {username, email}   Authorization: Bearer <admin token>
  ↓ JWT 中间件        ← 验签 + 过期检查，claims 注入 ctx
  ↓ JwtBlacklist     ← Redis 查 jwt:revoked:<token>
  ↓ UpdateUserInfoHandler → parse()
  ↓ UpdateUserInfoLogic
  │    ctx.Value("role") 断言失败或 != "admin" → 9004
  ↓ UpdateUser ──gRPC──▶ UpdateUserLogic
  │   ├ id <= 0 → InvalidArgument
  │   ├ 拼动态 updates（只更新非空字段）
  │   ├ 一个字段都没有 → InvalidArgument「没有可更新的字段」
  │   ├ db 错误 → mapUpdateUserError
  │   └ RowsAffected == 0 → NotFound
  ▼
{"code":0,"data":{"info":"success"}}
```

`UpdateUser` 这个 RPC 本身支持更新 `role` 字段，但 HTTP 层不接收也不转发 role —— `UserUpdateReq` 只声明了 `id` / `username` / `email`，客户端传 role 会被忽略。

### 4.8 个人资料 `GET` / `PUT /user/profile/:userId`

查：

```text
GET /user/profile/38  Authorization: Bearer <token>
  ↓ JWT 中间件 → JwtBlacklist
  ↓ GetProfileHandler → parse()（userId 转为 int64）
  ↓ GetProfileLogic
  │   ├ ctx.Value("userId").(json.Number) 断言失败 → 1005
  │   ├ .Int64() 转换失败 → 1005
  │   ├ id != req.UserId → 9004      ← 只能看自己的
  ↓ QueryUser(userId) ──gRPC──▶ PG
  ▼
{"code":0,"data":{"userId":38,"username":"...","role":"user"}}
```

改同理，`UpdateProfileReq` 只声明了 `username`，所以这个接口改不了 email / password / role。

两层鉴权的分工：JWT 中间件负责"你是谁"（验签、过期），Logic 里的 `id != req.UserId` 负责"你能不能碰这个资源"。中间件本身不做资源归属判断。

**Claims 类型要当心。** go-zero 用 `json.Number` 解析数字，这里必须做**带 ok 的断言**再用 `.Int64()`，而不是直接 `float64` 断言——不同解析路径下的数字类型不一样，畸形载荷直接断言会 panic。

### 4.9 注销 `POST /user/logout`

```text
POST /user/logout  Authorization: Bearer <token>
  ↓ JWT 中间件（验签通过）→ JwtBlacklist（此时还没拉黑，放行）
  ↓ LogoutHandler：从 Authorization 头取出 token，塞进 ctx["jwtToken"]
  ↓ LogoutLogic
  │   ├ token 为空 → 返回 success:false + message（不是错误）
  │   ├ jwt.ParseUnverified(token)  ← 只读 exp，故意不验签
  │   ├ claims["exp"] 缺失 → success:false + message
  │   └ ttl = exp - now；ttl > 0 时
  │       Redis.SETEX "jwt:revoked:<token>" "1" ttl秒
  │       失败 → 9006
  ▼
{"code":0,"data":{"success":true}}
```

三个细节：

- **不验签**是有意的。注销只需要 `exp` 算 TTL，而 token 的合法性已由上游 JWT 中间件保证过，重复验签没意义。
- **TTL 设为剩余有效期**，而不是固定值。Token 自然过期后黑名单条目也自动消失，不需要清理任务。
- **`success:false` 而不是错误码**：拿不到 Token 或 Token 缺 `exp` 属于"注销没做成"，但不是服务故障，所以走 `code:0` + body 里的 `success` 字段来说明。这是全项目里唯一用 body 字段而非错误码表达业务结果的地方。

后续带着同一个 Token 访问受保护接口，就会被 `JwtBlacklist` 拦截 → 1005。

### 4.10 文件上传 `POST /upload/file`

**不走 RPC，不碰数据库。** MIME 校验在 Handler，落盘在 Logic。

```text
POST /upload/file  multipart 字段名: file
  ↓ UploadFileHandler
  │   ① r.ParseMultipartForm(32<<20)  内存缓存 32MB，超出自动落盘
  │   ② r.FormFile("file")
  │   ③ 读前 512 字节 → http.DetectContentType
  │      ← 刻意不信任后缀名，按内容嗅探
  │      不在白名单 → 9008（detail 记实际 MIME）
  │      白名单：image/jpeg image/png image/gif application/pdf
  │   ④ file.Seek(0, io.SeekStart)   ← 必须回绕，否则后面读到空
  ↓ UploadFileLogic
  │   ├ header.Size > 10MB → 9009
  │   ├ 文件名清洗：filepath.Base(filepath.Clean(name)) 防路径穿越
  │   ├ 保存名 = <UnixNano>_<清洗后名>
  │   ├ UploadDir 不存在则 MkdirAll
  │   └ io.Copy 落盘
  ▼
{"code":0,"data":{"filename":"a.png","size":1024,"url":"/files/1730.._a.png"}}
```

第 ③ 到第 ④ 步之间的 `Seek(0)` 是个易错点：嗅探 MIME 时已经把文件指针读走了，不回绕的话 `io.Copy` 会拷到一个空文件。

返回的 `url` 是 `/files/<name>`，但**项目没有注册 `/files/*` 静态路由**，所以这个 URL 目前访问不到，字段只是演示用。

### 4.11 SSE `POST /sse/with/req`、`POST /sse/without/resp`

用 `rest.WithSSE()` 注册，是**唯一不使用统一信封的成功响应**——SSE 协议要求 `data: {...}` 帧格式，套不了 JSON 信封。

```text
POST /sse/with/req {body}
  ↓ sseHandler：parse()  ← 参数错仍然是 JSON 信封 9002
  ↓ client := make(chan *SseResp, 16)
  ↓ threading.GoSafeCtx 起 goroutine 跑 SseLogic，结束 close(client)
  ↓ 主循环 select：
  │   ├ 收到数据 → json.Marshal → "data: %s\n\n" → Flush
  │   ├ 收到 ctx.Done() → return
  │   └ client 关闭 → return
```

两个 Logic 都是循环 5 次、每次 `time.Sleep(time.Second)`，区别只是 `/with/req` 带上了请求内容：

```text
data: {"msg":"req: hi, message 0"}
data: {"msg":"req: hi, message 1"}
...
data: {"msg":"message 0"}      ← /sse/without/resp
```

`/sse/without/resp` 连请求参数都不接收（`.api` 里声明为 `returns (*SseResp)`），所以没有任何入参校验入口。

**已知不一致**：流建立之后的错误只写日志，不发帧给客户端。客户端无法区分"正常结束"和"中途失败"。若要修，应向客户端补一帧 `event: error`。

---

## 5. RPC 接口

协议定义：[`rpc/user.proto`](rpc/user.proto)。8 个方法。

| RPC | 作用 | HTTP 入口 |
|---|---|---|
| `QueryUser` | 按 ID 查用户 | `GET /user/:id`、`GET /user/profile/:userId`、`/user/refresh` |
| `AuthenticateUser` | 按用户名取认证数据（含哈希） | `/user/login` |
| `InsertUser` | 建用户 + bcrypt | `POST /users` |
| `UpdateUser` | 按需更新 username/email/role | `POST /user/:id`、`PUT /user/profile/:userId` |
| `CreateUserWithTodo` | 事务内建用户 + 待办 | `POST /users-with-todo` |
| `CreateUserWithTodoRollbackDemo` | 事务回滚演示 | `POST /users-with-todo/rollback-demo` |
| `DeleteUser` | 软删除用户 | **无 HTTP 包装** |
| `Pusher` | 向 Kafka 推演示消息 | **无 HTTP 包装** |

`DeleteUser` 和 `Pusher` 只有 RPC 能力，需要用 gRPC 客户端直接调，HTTP 层没有路由。

### 5.1 错误协议边界

RPC 侧只表达 gRPC 语义码，业务码和对外文案统一由 API 决定，避免两个服务各维护一套错误码。映射在 [`api/internal/logic/rpcerror.go`](api/internal/logic/rpcerror.go)：

| gRPC code | 业务码 | 说明 |
|---|---|---|
| `InvalidArgument` | 9002 | 沿用 RPC 的具体校验文案 |
| `NotFound` | 1002 | |
| `AlreadyExists` | 1001 | 沿用 RPC 判定的字段文案 |
| `FailedPrecondition` | 9002 | 外键约束 |
| `DeadlineExceeded` | 9007 | |
| `Unavailable` | 9006 | RPC 掉线，客户端重连前 |
| 其他 | 9001 | |
| `Unauthenticated` | 9003 | **预留给 RPC 内部鉴权，当前未启用** |

RPC Logic 的写法约定：参数校验用 `InvalidArgument`，校验顺序是"先参数、后业务、再数据库"。

### 5.2 RPC 内部鉴权尚未启用

`rpc/user.go` 定义了 `authInterceptor`，检查 metadata 里有没有 `token`：

```go
func authInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (any, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if len(md["token"]) == 0 {
		return nil, status.Error(codes.Unauthenticated, "缺少 token")
	}
	return handler(ctx, req)
}
```

但注册代码是注释状态：

```go
//s.AddUnaryInterceptors(authInterceptor)
```

所以 API 通过 `RpcContext` 附加的 `token: internal` **只是演示性标记，不构成服务间认证**：

```go
func (s *ServiceContext) RpcContext(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "token", "internal")
}
```

实际生效的拦截器只有流式日志那个：`s.AddStreamInterceptors(loggingStreamInterceptor)`（项目里没有流式 RPC，等于空转）。

---

## 6. 数据模型

建表脚本 [`rpc/migrations/0001_init.sql`](rpc/migrations/0001_init.sql)，由 compose 挂到 `/docker-entrypoint-initdb.d/`。

### users

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | `BIGSERIAL` | 主键，序列生成 |
| `username` | `VARCHAR(64)` | |
| `email` | `VARCHAR(128)` | |
| `password` | `VARCHAR(256)` | bcrypt 哈希 |
| `role` | `VARCHAR(32)` | 默认 `user` |
| `created_at` / `updated_at` | `TIMESTAMPTZ` | |
| `deleted_at` | `TIMESTAMPTZ` | 软删除标记 |

### todos

`id` / `user_id` / `title` / `content` / `status` / 时间戳，`user_id` 外键指向 `users(id)`。

### 三条设计要点

**部分唯一索引。** 用户名和邮箱的唯一性只约束未删除记录：

```sql
CREATE UNIQUE INDEX idx_users_username ON users (username) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_users_email    ON users (email)    WHERE deleted_at IS NULL;
```

所以软删一个用户后，它的用户名和邮箱可以立刻被重新注册。这也是为什么唯一冲突时要靠**约束名**去区分是哪个字段被占用。

**软删除。** 模型里 `DeletedAt gorm.DeletedAt`，`Delete` 默认只写 `deleted_at` 而**不是**物理删除，普通查询自动过滤已删记录。

**密码不出 JSON。** `Password` 字段标了 `json:"-"`，即使模型被序列化也不会带出哈希。

### GORM 初始化

[`rpc/internal/db/db.go`](rpc/internal/db/db.go)：`TranslateError: true`（把驱动错误转成 `gorm.ErrDuplicatedKey` 这类可判定的错误）、最大连接 50、最大空闲 10、连接生命周期 1h、空闲回收 10min。日志级别由 `DB_LOG_LEVEL` 环境变量控制。

生成代码：`rpc/cmd/gen/main.go` → `rpc/internal/dal/{model,query}/`（gorm/gen 产出，`DO NOT EDIT`）。

---

## 7. 其他实现要点（简要）

### 7.1 JWT 黑名单

Redis 只做一件事：记录已注销的 Token。`jwt:revoked:<token>` → `"1"`，TTL 为 Token 剩余有效期。不是缓存体系。

`JwtBlacklistMiddleware` 的失败策略是 **fail-open**：

```go
exists, err := m.Redis.Exists("jwt:revoked:" + tokenStr)
if err == nil && exists {     // ← Redis 报错时放行
	httpx.ErrorCtx(r.Context(), w, errcode.ErrUserTokenInvalid)
	return
}
```

`err != nil` 时请求继续往下走。可用性优先，但 Redis 挂掉期间已注销的 Token 会重新变得有效。生产环境需要按安全要求重新决定这里的策略。

### 7.2 链路追踪

两个服务都接 OTLP/gRPC 上报到 Jaeger。API 的响应体里回填 `trace_id`，可以直接拿它去 Jaeger 搜到完整链路：

```text
user 服务：  GET /users  →  user.User/InsertUser（client span）
user.rpc 服务：            user.User/InsertUser（server span）
```

默认是批量上报，**刚发完请求立刻去查可能只有一半 span，等几秒刷盘就齐了**。

宿主机的 RPC 配置（`rpc/etc/user.yaml`）**没有 Telemetry 段**，所以本机 `go run` 模式只能看到半条链路；容器配置补上了这一段。

### 7.3 Kafka

`Pusher` RPC 往 `payment-success` 推固定内容 `"zhangSan"`；`PaymentSuccess` 消费者只把收到的 key/value 打日志，**不写数据库、没有订单业务**。这个 topic 不参与登录、查询、注销任何主链路。

消费者组名在容器配置里改成了 `kqConsumer-docker`，避免和宿主机 `go run` 的 RPC 落到同一个组——单分区下后加入的消费者收不到消息。

### 7.4 Prometheus

API 在独立端口 9091 暴露 `/metrics`（主端口 8888 不冲突）。

### 7.5 中间件与路由注册顺序

`api/user.go` 启动时：

```text
conf.MustLoad
  ↓ rest.MustNewServer(..., WithUnauthorizedCallback)   ← 先挂 JWT 未认证回调
  ↓ handler.RegisterErrorHandler()                        ← 再注册统一错误处理器
  ↓ server.Use(中间件)
  ↓ svc.NewServiceContext(c)                              ← 建 RPC 客户端、Redis 客户端
  ↓ handler.RegisterHandlers(server, ctx)
  ↓ server.Start()
```

`RegisterHandlers` 按 `.api` 里的分组顺序注册：公开组 → SSE 组（`rest.WithSSE()`）→ 受保护组（`rest.WithMiddlewares([JwtBlacklist])` + `rest.WithJwt(AccessSecret)`）。

注意受保护组是**先 JWT、后黑名单**（见 `engine.go` 的 `appendAuthHandler`）。顺序很重要：JWT 先验签，无效 token 在 9003 就结束了，不会被黑名单中间件看到；只有**签名有效但已注销**的 token 才会走到黑名单拿到 1005。这两个码不冲突。

---

## 8. 运行方式

两步：

```bash
bash build.sh          # 宿主机交叉编译 → build/user-api、build/user-rpc
docker compose up -d   # 拉起全栈
```

`build.sh` 必须先在宿主机跑。原因：本机装了 E-SafeNet 透明加密，**仓库里的 `.go` 是密文**，只有白名单进程（`go.exe`、`git.exe`、编辑器）读到明文，Docker 不在白名单里。所以 `api/Dockerfile` 和 `rpc/Dockerfile` 都改成**只打包**——COPY 宿主机编译好的二进制 + 一份配置，绝不在镜像内 `go build`。

配置分两份：`api/etc/user-api.yaml` 和 `rpc/etc/user.yaml` 用 `127.0.0.1`（供宿主机 `go run`），同目录的 `*.docker.yaml` 换成 compose 服务名（供容器）。

端口映射：

| 服务 | 端口 |
|---|---|
| API | 8888 |
| API Prometheus | 9091 |
| RPC | 8080 |
| PostgreSQL | 5432 |
| Redis | 6379 |
| etcd | 2379 |
| Kafka（宿主机 / 容器） | 29092 / 9092 |
| kafka-ui | 8081 |
| Jaeger UI | 16686 |

Kafka 有两个监听器：宿主机程序用 `29092`，容器内用 `kafka:9092`。**不能在宿主机 Go 程序里用 `kafka:9092`。**

宿主机 `go run` 的老流程仍然可用，但和容器抢 8080/8888/9091，两者只能二选一。

---

## 9. 已知限制

按影响排序：

| # | 限制 | 说明 |
|---|---|---|
| 1 | **RPC 内部鉴权未启用** | `authInterceptor` 是注释状态，`token: internal` 只是标记，不是认证。任何能连上 8080 的客户端都能直接调 RPC |
| 2 | **Redis 黑名单 fail-open** | Redis 故障期间已注销 Token 恢复有效 |
| 3 | **明文密码进日志** | go-zero 客户端拦截器在 RPC 报错时把整个请求体打进日志，含明文密码。重复用户名、RPC 掉线都会触发 |
| 4 | **Bearer Token 进日志** | go-zero 鉴权失败时会 `DumpRequest` 整个请求（含 Authorization 头） |
| 5 | **Access Token 有效期硬编码** | `generateAccessToken` 固定 `now + 24h`，但响应里的 `expiresIn` 返回的是配置值 `Auth.AccessExpire`。当前两者恰好都是 86400 所以一致，**配置一改就会不一致** |
| 6 | **SSE 错误被吞** | 流建立后的错误只落日志，客户端无法区分正常结束与失败 |
| 7 | **上传文件不可访问** | 返回 `/files/<name>`，但没有静态路由 |
| 8 | **上传无内容安全检测** | 只有 MIME 嗅探 + 大小限制，没有病毒扫描；`Config.S3Bucket` 存在但 `uploadToS3` 是注释的，未接对象存储 |
| 9 | **迁移不自动重跑** | `docker-entrypoint-initdb.d` 只在数据卷首次创建时执行，改 SQL 不会生效（需 `down -v` 重置或手工执行） |
| 10 | **无 Todo 独立接口** | `todos` 表和模型、query 都有，也通过场景 A 被写入，但没有独立的 Todo CRUD，只能在建用户时顺带创建 |
| 11 | **默认配置仅演示用** | 库密码和 JWT Secret 明文提交在 `etc/` 里 |

---

## 10. 小结

主链路是清晰的单向依赖：

```text
HTTP → API（协议/鉴权/响应）→ RPC（业务/唯一数据访问）→ PostgreSQL
```

分工上，**API 负责"能不能进"，RPC 负责"数据怎么变"**：

- 协议层的事（参数解析定性、统一信封、双 Token 签发、JWT 与黑名单、路径参数与资源归属校验）全在 API；
- 数据层的事（参数校验、bcrypt、事务、约束错误判定、序列主键）全在 RPC。

两个场景是这个项目的核心演示点：

- **场景 A**（`/users-with-todo`）：跨表原子写入，待办外键依赖事务内自增主键，任一步失败整体回滚；
- **场景 B**（`/users`）：唯一约束冲突在 RPC 侧按约束名区分字段，在 API 侧映射成 1001 并沿用具体文案。

而 `rollback-demo` 是场景 A 的**可执行反例**——它把"事务失败"从异常变成了可预期的返回值。

配套的横切能力：统一错误码分段（`1xxx` 用户 / `9xxx` 系统）、单一错误出口、OTel 链路追踪（`trace_id` 回填响应）、Prometheus 指标、以及一条命令拉起全栈的 Docker 编排。
