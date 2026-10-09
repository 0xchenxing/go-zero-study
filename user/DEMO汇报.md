# user 项目整合演示汇报

## 1. 项目目标

本项目是一个 Go-zero 微服务示例，重点演示以下技术组合：

- Go-zero REST API
- Go-zero zRPC
- Protocol Buffers / gRPC
- PostgreSQL
- GORM 与 GORM Gen
- JWT 登录认证
- Redis Token 黑名单
- Kafka 消息生产与消费
- etcd 服务注册配置
- SSE 流式响应
- multipart 文件上传

项目的核心目标是将用户数据访问统一放到 RPC 服务中，由 RPC 服务通过 GORM Gen 访问 PostgreSQL。API 只负责接收 HTTP 请求、鉴权、调用 RPC 和组织响应。

本项目用于本地学习和功能演示，不应直接作为生产环境的完整认证、文件存储或服务治理方案。

## 2. 整体架构

### 2.1 核心调用链

```text
HTTP Client
    │ HTTP :8888
    ▼
user/api
    │ REST、JWT、权限、Redis 黑名单、SSE、文件上传
    │ zRPC，固定 Endpoint：127.0.0.1:8080
    ▼
user/rpc
    │ gRPC、业务逻辑、GORM Gen
    ▼
PostgreSQL :5432
```

### 2.2 辅助组件

```text
API ── Redis :6379
RPC ── etcd :2379
RPC ── Kafka：127.0.0.1:29092
Kafka UI ── kafka:9092
```

各组件职责如下：

| 组件 | 作用 | 是否属于用户主链路 |
|---|---|---|
| API | 对外提供 HTTP 接口、JWT 和权限处理 | 是 |
| RPC | 提供内部用户服务并访问数据库 | 是 |
| PostgreSQL | 保存用户数据 | 是 |
| Redis | 保存已注销 JWT 的黑名单 | 受保护接口需要 |
| etcd | RPC 服务注册相关配置 | 当前不是 API 实际发现路径 |
| Kafka | 消息生产与消费演示 | 否 |
| Kafka UI | 查看 Kafka 消息 | 否 |

当前 API 配置使用固定地址：

```yaml
UserRpc:
  Endpoints:
    - 127.0.0.1:8080
```

因此，当前 API 调用 RPC 时不会通过 etcd 动态发现服务。RPC 配置虽然包含 etcd 信息，但它不代表 API 当前已经使用 etcd 完成动态发现。

## 3. 代码模块职责

### 3.1 API 模块

目录：

```text
api/
```

主要职责：

- 提供 HTTP REST 接口；
- 解析请求参数并组织响应；
- 校验 Access Token；
- 检查 JWT 黑名单；
- 判断管理员权限和用户资源归属；
- 生成 Access Token 和 Refresh Token；
- 通过 `userclient` 调用 RPC；
- 实现登出、SSE 和文件上传。

API 不直接访问 PostgreSQL，用户数据统一通过 RPC 获取和修改。

### 3.2 RPC 模块

目录：

```text
rpc/
```

主要职责：

- 提供用户相关 gRPC 方法；
- 使用 GORM Gen 访问 PostgreSQL；
- 对密码执行 bcrypt 哈希；
- 查询、认证、插入、更新和软删除用户；
- 提供 Kafka Producer 和 Consumer 演示。

### 3.3 数据库模块

数据库连接代码：

```text
rpc/internal/db/db.go
```

GORM 初始化时配置了：

- PostgreSQL Driver；
- `TranslateError: true`；
- 最大连接数 50；
- 最大空闲连接数 10；
- 连接最大生命周期 1 小时；
- 空闲连接最大时间 10 分钟。

初始化 SQL：

```text
rpc/migrations/0001_init.sql
```

## 4. 启动顺序和端口

### 4.1 启动基础设施

在项目根目录执行：

```bash
docker compose up -d postgres etcd kafka kafka-ui
```

查看状态：

```bash
docker compose ps
```

Redis 当前不在 `docker-compose.yml` 中，需要另外准备，默认地址为：

```text
127.0.0.1:6379
```

如果只准备 PostgreSQL 和 etcd，可以执行：

```bash
docker compose up -d postgres etcd
```

不过 RPC 配置会初始化 Kafka Producer 和 Consumer，实际启动 RPC 时还应确认 Kafka 配置和运行状态。

### 4.2 端口说明

| 服务 | 地址 | 用途 |
|---|---|---|
| API | `127.0.0.1:8888` | HTTP 请求 |
| RPC | `127.0.0.1:8080` | zRPC 请求 |
| PostgreSQL | `127.0.0.1:5432` | 用户数据库 |
| Redis | `127.0.0.1:6379` | JWT 黑名单 |
| etcd | `127.0.0.1:2379` | 服务注册配置 |
| Kafka 宿主机入口 | `127.0.0.1:29092` | 宿主机 Go 程序访问 |
| Kafka 容器内部入口 | `kafka:9092` | Kafka UI 等容器访问 |
| Kafka UI | `127.0.0.1:8081` | 查看 Kafka 消息 |

Kafka 有两个访问地址：

- 宿主机上运行的 RPC 程序使用 `127.0.0.1:29092`；
- Docker 容器内部使用 `kafka:9092`。

不能在宿主机 Go 程序中直接使用容器内部地址 `kafka:9092`。

### 4.3 启动 RPC

```bash
go run ./rpc -f rpc/etc/user.yaml
```

RPC 监听：

```text
0.0.0.0:8080
```

RPC 启动时会：

1. 加载配置；
2. 连接 PostgreSQL；
3. 初始化 GORM Gen 查询对象；
4. 创建 Redis 客户端；
5. 创建 Kafka Producer；
6. 注册 `UserServer`；
7. 启动 Kafka Consumer；
8. 开始监听 RPC 请求。

### 4.4 启动 API

另开一个终端执行：

```bash
go run ./api -f api/etc/user-api.yaml
```

API 监听：

```text
0.0.0.0:8888
```

API 启动时会：

1. 加载 REST 配置；
2. 创建 zRPC 客户端；
3. 创建 Redis 客户端；
4. 注册 JWT 和黑名单中间件；
5. 注册 HTTP 路由；
6. 开始监听 HTTP 请求。

## 5. HTTP 接口清单

| 方法 | 路径 | JWT | 主要职责 |
|---|---|---|---|
| `POST` | `/user/login` | 否 | 用户登录并返回双 Token |
| `POST` | `/user/refresh` | 否 | 使用 Refresh Token 重新获取 Token |
| `GET` | `/user/:id` | 否 | 公开查询用户基础信息 |
| `POST` | `/user/:id` | 是 | 管理员修改用户资料 |
| `GET` | `/user/profile/:userId` | 是 | 查询当前用户自己的资料 |
| `PUT` | `/user/profile/:userId` | 是 | 修改当前用户自己的用户名 |
| `POST` | `/user/logout` | 是 | 将当前 Access Token 加入黑名单 |
| `POST` | `/upload/file` | 否 | 上传本地文件 |
| `POST` | `/sse/with/req` | 否 | 接收请求并返回 SSE 消息 |
| `POST` | `/sse/without/resp` | 否 | 直接返回 SSE 消息 |

受保护路由同时使用：

```text
JWT 中间件 + JwtBlacklist 中间件
```

### 5.1 登录：`POST /user/login`

请求：

```json
{
  "username": "demo",
  "password": "demo123"
}
```

处理流程：

1. API 接收用户名和密码；
2. API 调用 RPC 的 `AuthenticateUser`；
3. RPC 根据用户名查询 PostgreSQL；
4. RPC 返回用户资料和 bcrypt 密码哈希；
5. API 使用 `bcrypt.CompareHashAndPassword` 校验密码；
6. API 生成 Access Token 和 Refresh Token；
7. 返回 Token 和过期时间。

响应结构：

```json
{
  "accessToken": "<access-token>",
  "refreshToken": "<refresh-token>",
  "expiresIn": 86400
}
```

Access Token 中包含：

```json
{
  "userId": 1,
  "role": "user",
  "iat": 0,
  "exp": 0
}
```

其中 `iat` 和 `exp` 的实际值由程序运行时间决定。

### 5.2 刷新 Token：`POST /user/refresh`

请求：

```json
{
  "refreshToken": "<refresh-token>"
}
```

处理流程：

1. 使用 Refresh Secret 解析 Refresh Token；
2. 读取 Token 中的用户 ID；
3. 调用 `QueryUser` 获取数据库中的最新用户信息；
4. 读取当前角色；
5. 重新生成 Access Token 和 Refresh Token。

Refresh Token 和 Access Token 使用不同的 Secret 和过期配置。

### 5.3 公开查询：`GET /user/:id`

示例：

```bash
curl http://127.0.0.1:8888/user/1
```

API 调用 RPC 的 `QueryUser`，但只向客户端返回：

```json
{
  "id": 1,
  "username": "demo",
  "email": "demo@example.com"
}
```

不会返回密码哈希，也不会将内部认证字段暴露给 HTTP 客户端。

### 5.4 管理员更新：`POST /user/:id`

请求：

```json
{
  "username": "updated-user",
  "email": "updated@example.com"
}
```

要求：

- 必须携带有效 JWT；
- JWT 中的 `role` 必须为 `admin`；
- 可以修改任意用户的用户名和邮箱；
- HTTP 请求不会接收或转发 `role` 字段。

API 负责判断管理员权限，RPC 的通用 `UpdateUser` 方法本身还支持 role 字段，但当前这个 HTTP 接口没有把 role 传给 RPC。

### 5.5 查询个人资料：`GET /user/profile/:userId`

示例：

```bash
curl http://127.0.0.1:8888/user/profile/1 \
  -H "Authorization: Bearer $TOKEN"
```

除了 JWT 签名和过期时间检查外，API 还要求：

```text
JWT userId == URL 中的 userId
```

响应：

```json
{
  "userId": 1,
  "username": "demo",
  "role": "user"
}
```

### 5.6 修改个人资料：`PUT /user/profile/:userId`

请求：

```json
{
  "username": "new-demo"
}
```

只允许用户修改自己的用户名，不允许通过该接口修改：

- email；
- password；
- role；
- 其他用户的资料。

### 5.7 注销：`POST /user/logout`

示例：

```bash
curl -X POST http://127.0.0.1:8888/user/logout \
  -H "Authorization: Bearer $TOKEN"
```

注销逻辑会把当前 Token 写入 Redis：

```text
jwt:revoked:<token>
```

Redis Key 的 TTL 设置为 Token 剩余有效期。后续访问受保护接口时，黑名单中间件会拒绝这个 Token。

### 5.8 文件上传：`POST /upload/file`

Handler 使用 multipart 字段：

```text
file
```

当前行为：

- 最大文件大小为 10 MB；
- 允许 JPEG、PNG、GIF、PDF；
- 使用清理后的文件名和时间戳生成保存名；
- 保存到配置中的 `UploadDir`；
- 返回原文件名、文件大小和 URL。

响应结构：

```json
{
  "filename": "demo.png",
  "size": 1024,
  "url": "/files/1730000000000000000_demo.png"
}
```

当前文档只将它作为本地文件保存演示。项目中如果没有额外注册 `/files/*` 静态文件服务，返回的 URL 不一定可以直接访问。

### 5.9 SSE：`POST /sse/with/req`

请求：

```json
{
  "body": "hello"
}
```

服务端连续发送 5 条消息，每条消息包含请求内容和序号：

```text
req: hello, message 0
req: hello, message 1
...
req: hello, message 4
```

### 5.10 SSE：`POST /sse/without/resp`

该接口不需要普通响应体，服务端连续返回 5 条 SSE 消息：

```text
message 0
message 1
...
message 4
```

## 6. 完整请求链路

### 6.1 登录链路

```text
POST /user/login
  ↓
LoginHandler
  ↓
LoginLogic.Login
  ↓
ServiceContext.RpcContext
  ↓
userclient.AuthenticateUser
  ↓
zRPC 127.0.0.1:8080
  ↓
AuthenticateUserLogic
  ↓
GORM Gen Query
  ↓
PostgreSQL.users
  ↓
返回用户信息和 bcrypt 哈希
  ↓
API 校验明文密码
  ↓
生成 Access Token 和 Refresh Token
  ↓
HTTP 返回
```

API 调用 RPC 时会附加 metadata：

```text
token: internal
```

### 6.2 Profile 查询链路

```text
GET /user/profile/:userId
  ↓
JWT 中间件
  ↓
JwtBlacklist 中间件
  ↓
GetProfileHandler
  ↓
GetProfileLogic
  ↓
校验 JWT userId 与路径 userId
  ↓
userclient.QueryUser
  ↓
zRPC
  ↓
QueryUserLogic
  ↓
GORM Gen
  ↓
PostgreSQL
  ↓
返回 Profile
```

JWT 校验和 Redis 黑名单检查发生在 API 侧，数据库查询发生在 RPC 侧。

### 6.3 管理员更新链路

```text
POST /user/:id
  ↓
JWT 中间件
  ↓
JwtBlacklist 中间件
  ↓
UpdateUserInfoHandler
  ↓
UpdateUserInfoLogic
  ↓
检查 JWT role == admin
  ↓
userclient.UpdateUser
  ↓
zRPC
  ↓
UpdateUserLogic
  ↓
GORM Updates
  ↓
PostgreSQL.users
```

### 6.4 登出链路

```text
POST /user/logout
  ↓
JWT 验证
  ↓
JwtBlacklist 中间件
  ↓
LogoutLogic
  ↓
读取 JWT exp
  ↓
Redis SETEX jwt:revoked:<token>
  ↓
后续受保护请求被黑名单拦截
```

### 6.5 RPC 内部鉴权的当前状态

`rpc/user.go` 中定义了 `authInterceptor`，它可以检查请求 metadata 中是否存在 `token`：

```go
func authInterceptor(...) (any, error) {
    md, _ := metadata.FromIncomingContext(ctx)
    if len(md["token"]) == 0 {
        return nil, status.Error(codes.Unauthenticated, "缺少 token")
    }
    return handler(ctx, req)
}
```

但是实际注册代码当前仍然是注释状态：

```go
// s.AddUnaryInterceptors(authInterceptor)
```

因此，当前不能描述为“RPC Unary Interceptor 已经完成内部鉴权”。实际生效的是：

- API 侧 JWT 验证；
- API 侧 Redis 黑名单检查；
- API 调 RPC 时传递 `token: internal` metadata。

这个 metadata 目前只是演示性的内部标记，并不是生产级服务间认证方案。

## 7. RPC 接口

协议文件：

```text
rpc/user.proto
```

| RPC | 作用 | 当前使用情况 |
|---|---|---|
| `QueryUser` | 根据 ID 查询用户 | 公开查询、Profile、刷新 Token |
| `AuthenticateUser` | 根据用户名查询认证数据 | 登录 |
| `InsertUser` | 创建用户并生成 bcrypt 哈希 | RPC 能力，无 HTTP 包装 |
| `UpdateUser` | 更新用户名、邮箱或角色 | 管理员更新、Profile 更新 |
| `DeleteUser` | 删除用户 | RPC 能力，无 HTTP 包装 |
| `Pusher` | 向 Kafka 推送演示消息 | RPC 能力，无 HTTP 包装 |

每个 RPC Server 方法只负责创建对应 Logic 并委托执行，例如：

```text
UserServer.QueryUser
  ↓
NewQueryUserLogic
  ↓
QueryUserLogic.QueryUser
```

真正的数据库访问位于 `rpc/internal/logic/` 中。

## 8. PostgreSQL 与 GORM Gen

### 8.1 GORM Gen 目录

生成器入口：

```text
rpc/cmd/gen/main.go
```

生成结果：

```text
rpc/internal/dal/model/
rpc/internal/dal/query/
```

用户模型：

```text
rpc/internal/dal/model/users.gen.go
```

查询示例：

```go
item, err := l.svcCtx.PostgresQuery.User.
    WithContext(l.ctx).
    Where(l.svcCtx.PostgresQuery.User.ID.Eq(in.GetId())).
    First()
```

API 不直接执行这条查询，调用关系是：

```text
API Logic → userclient → RPC Logic → GORM Gen → PostgreSQL
```

### 8.2 users 表

初始化脚本：

```text
rpc/migrations/0001_init.sql
```

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | `BIGSERIAL` | 用户 ID |
| `username` | `VARCHAR(64)` | 用户名 |
| `email` | `VARCHAR(128)` | 邮箱 |
| `password` | `VARCHAR(256)` | bcrypt 密码哈希 |
| `role` | `VARCHAR(32)` | 用户角色，默认 `user` |
| `created_at` | `TIMESTAMPTZ` | 创建时间 |
| `updated_at` | `TIMESTAMPTZ` | 更新时间 |
| `deleted_at` | `TIMESTAMPTZ` | 软删除时间 |

用户名和邮箱使用部分唯一索引：

```sql
CREATE UNIQUE INDEX idx_users_username
ON users (username)
WHERE deleted_at IS NULL;
```

含义是：

- 未删除用户之间的用户名唯一；
- 未删除用户之间的邮箱唯一；
- 用户软删除后，原用户名和邮箱可以重新使用。

### 8.3 密码处理

创建用户时：

```text
明文密码
  ↓
bcrypt.GenerateFromPassword
  ↓
数据库保存密码哈希
```

登录时：

```text
请求明文密码 + 数据库密码哈希
  ↓
bcrypt.CompareHashAndPassword
```

模型中的密码字段使用：

```go
json:"-"
```

避免通过模型 JSON 序列化直接返回密码。

### 8.4 软删除

`User` 模型包含：

```go
DeletedAt gorm.DeletedAt
```

因此调用 GORM `Delete` 时默认是软删除，而不是立即物理删除。普通查询会自动忽略 `deleted_at` 不为空的记录。

### 8.5 todos 表

迁移脚本还包含 `todos` 表、索引和用户外键，用于展示数据库结构设计。但是当前项目没有：

- Todo HTTP 路由；
- Todo API Logic；
- Todo RPC 方法；
- Todo 可演示业务流程。

因此，`todos` 目前只是数据库结构示例，不能描述为已经完成的 Todo 业务功能。

## 9. 扩展功能

### 9.1 Redis JWT 黑名单

API 在注销时将 Token 写入 Redis：

```text
jwt:revoked:<token>
```

受保护路由会通过 `JwtBlacklist` 中间件检查该 Key 是否存在。Redis 主要用于 JWT 失效控制，当前不是完整缓存体系。

### 9.2 etcd

RPC 配置中包含：

```yaml
Etcd:
  Hosts:
    - 127.0.0.1:2379
  Key: user.rpc
```

它用于演示 go-zero 服务注册相关配置。但当前 API 仍配置固定 RPC Endpoint：

```yaml
UserRpc:
  Endpoints:
    - 127.0.0.1:8080
```

所以当前主链路不是：

```text
API → etcd → RPC
```

而是：

```text
API → 127.0.0.1:8080
```

### 9.3 Kafka

Kafka 配置：

```text
Topic: payment-success
```

Producer 由 `Pusher` RPC 使用，Consumer 由 `PaymentSuccess` 处理。目前 Consumer 主要记录收到的 key 和 value 日志，没有继续更新数据库或执行订单业务。

Kafka 是独立的消息队列演示能力，不参与：

- 登录；
- 用户查询；
- Profile 查询和更新；
- JWT 注销。

当前 HTTP API 没有 Kafka 业务路由，因此不能直接通过普通 HTTP curl 调用 `Pusher`。

### 9.4 SSE

SSE 接口使用 Go-zero 的 `rest.WithSSE()` 注册，适合演示服务端连续推送消息。两个 SSE 接口都会发送 5 条消息，每条消息之间间隔约 1 秒。

### 9.5 文件上传

文件上传是本地磁盘保存演示，当前并未接入对象存储服务。文件类型、大小和文件名会在 API Handler/Logic 中进行基本处理。

## 10. 可复现演示流程

以下命令是演示模板。实际返回内容和状态码应以运行环境为准，未执行的命令不能标记为“已通过”。

### 10.1 查看用户数据

```bash
docker exec postgres psql -U devuser -d appdb \
  -c "select id, username, email, role from users;"
```

如需准备管理员用户：

```bash
docker exec postgres psql -U devuser -d appdb \
  -c "update users set role = 'admin' where id = 1;"
```

确认角色：

```bash
docker exec postgres psql -U devuser -d appdb \
  -c "select id, username, role from users where id = 1;"
```

### 10.2 公开查询

```bash
curl http://127.0.0.1:8888/user/1
```

预期是返回用户 ID、用户名和邮箱。

### 10.3 登录

```bash
curl -X POST http://127.0.0.1:8888/user/login \
  -H "Content-Type: application/json" \
  -d '{"username":"demo","password":"demo123"}'
```

保存返回的 Token：

```bash
TOKEN="<access-token>"
REFRESH_TOKEN="<refresh-token>"
```

### 10.4 查询个人资料

```bash
curl http://127.0.0.1:8888/user/profile/1 \
  -H "Authorization: Bearer $TOKEN"
```

预期是返回用户自己的 Profile。如果路径中的用户 ID 与 Token 中的用户 ID 不一致，API 应拒绝请求。

### 10.5 修改个人资料

```bash
curl -X PUT http://127.0.0.1:8888/user/profile/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"username":"new-demo"}'
```

### 10.6 管理员修改用户

使用管理员用户登录后执行：

```bash
ADMIN_TOKEN="<admin-access-token>"

curl -X POST http://127.0.0.1:8888/user/2 \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"username":"updated-user","email":"updated@example.com"}'
```

普通用户使用自己的 Token 调用该接口，预期会被业务权限判断拒绝。

### 10.7 刷新 Token

```bash
curl -X POST http://127.0.0.1:8888/user/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refreshToken\":\"$REFRESH_TOKEN\"}"
```

### 10.8 登出并验证黑名单

```bash
curl -X POST http://127.0.0.1:8888/user/logout \
  -H "Authorization: Bearer $TOKEN"
```

然后再次访问受保护接口：

```bash
curl http://127.0.0.1:8888/user/profile/1 \
  -H "Authorization: Bearer $TOKEN"
```

预期是被 Redis 黑名单中间件拒绝。

### 10.9 SSE 演示

```bash
curl -N -X POST http://127.0.0.1:8888/sse/with/req \
  -H "Content-Type: application/json" \
  -d '{"body":"hello"}'
```

```bash
curl -N -X POST http://127.0.0.1:8888/sse/without/resp
```

### 10.10 文件上传演示

```bash
curl -X POST http://127.0.0.1:8888/upload/file \
  -F "file=@./demo.png"
```

检查返回结果和 `UploadDir` 中是否生成文件。由于当前文档不保证存在 `/files/*` 静态路由，返回的 URL 应视实际路由配置验证。

### 10.11 Kafka 演示

Kafka Producer 和 Consumer 在 RPC 启动过程中初始化。`Pusher` 是 RPC 方法，不是 HTTP 路由，需要使用 RPC 客户端调用。消息发送后，可以通过：

- RPC 服务日志；
- Consumer 日志；
- Kafka UI：`http://127.0.0.1:8081`

观察 `payment-success` Topic 的消息。

## 11. 验证说明

本文只记录源码和配置中能够确认的事实，不把未执行的命令写成已通过。

建议在目标环境执行以下验证：

```bash
go test ./...
go build ./...
docker compose ps
```

核心接口验证：

```bash
curl http://127.0.0.1:8888/user/1
curl -X POST http://127.0.0.1:8888/user/login ...
curl http://127.0.0.1:8888/user/profile/1 ...
curl -X PUT http://127.0.0.1:8888/user/profile/1 ...
curl -X POST http://127.0.0.1:8888/user/refresh ...
curl -X POST http://127.0.0.1:8888/user/logout ...
```

验证结果应记录：

- 实际 HTTP 状态码；
- 响应体中的业务码和消息；
- RPC 服务日志；
- PostgreSQL 查询结果；
- Redis 黑名单 Key；
- Kafka Consumer 日志。

## 12. 已知限制

### 12.1 Access Token 有效期配置

配置中存在 `AccessExpire`，但当前 Access Token 生成逻辑使用固定的 24 小时计算过期时间。返回的 `expiresIn` 与 Token 实际 `exp` 可能出现不一致。

### 12.2 JWT Claims 类型处理

部分逻辑直接对 JWT Claims 做类型断言，例如将 `userId` 断言为 `float64` 或 `json.Number`。不同解析路径下的数字类型可能不同，异常 Token Claims 也可能造成类型断言问题。

### 12.3 RPC interceptor 未启用

`authInterceptor` 已定义，但 `s.AddUnaryInterceptors(authInterceptor)` 当前被注释。API 传递的 `token: internal` 目前不能等同于已经生效的 RPC 身份认证。

### 12.4 Redis 故障处理策略

JWT 黑名单查询发生 Redis 错误时，当前中间件可能继续放行请求。这是 fail-open 行为，生产环境需要根据安全要求重新确定 Redis 故障时的处理策略。

### 12.5 文件访问

上传文件保存在本地目录，但返回的 `/files/...` URL 是否可访问取决于项目是否额外配置静态文件路由。文件上传也没有接入对象存储、病毒扫描或更完整的内容安全检测。

### 12.6 数据库迁移

Docker Compose 的初始化 SQL 只会在 PostgreSQL 数据卷首次创建时自动执行。数据卷已经存在时，修改迁移脚本不会自动重新执行。

### 12.7 默认配置仅用于本地演示

项目中的 PostgreSQL 账号、密码和 JWT Secret 是演示配置，不应直接用于生产环境。生产环境应使用安全的密钥管理和服务凭据配置。

### 12.8 当前没有完整用户 CRUD HTTP 闭环

RPC 中存在 `InsertUser` 和 `DeleteUser`，但当前 HTTP API 没有对应的注册和删除路由。HTTP 层实际暴露的是查询、登录、刷新、更新、Profile、注销、上传和 SSE 等接口。

## 13. 总结

本项目已经形成了清晰的用户服务调用链：

```text
HTTP API
  → JWT / 权限 / Redis 黑名单
  → userclient
  → zRPC
  → RPC Logic
  → GORM Gen
  → PostgreSQL
```

主要完成内容包括：

1. 将用户数据访问集中到 RPC 服务；
2. 使用 PostgreSQL 保存用户数据；
3. 使用 GORM Gen 生成模型和类型化查询代码；
4. 使用 bcrypt 保存和校验密码；
5. 使用 JWT 实现 Access Token 和 Refresh Token；
6. 使用 role 区分管理员和普通用户；
7. 使用 JWT userId 限制用户只能访问自己的 Profile；
8. 使用 Redis 保存已注销 Token 的黑名单；
9. 使用 GORM 软删除和部分唯一索引；
10. 提供 SSE、文件上传、Kafka 和 etcd 的扩展演示能力。

需要准确区分核心链路和扩展能力：

- 登录、用户查询、Profile、管理员更新和注销属于 API → RPC → PostgreSQL 的核心链路；
- Redis 参与受保护接口的 Token 黑名单检查；
- SSE、文件上传和 Kafka 是独立演示功能；
- etcd 当前主要体现 RPC 注册配置，API 实际使用固定 Endpoint；
- `todos` 目前只有数据库结构示例，没有业务接口。

因此，该项目适合作为 Go-zero REST + zRPC + PostgreSQL + GORM Gen 的综合学习和本地演示项目，但仍需要补充更严格的服务间鉴权、统一错误处理、完整测试、生产级文件存储和服务发现配置，才能进一步用于生产环境。
