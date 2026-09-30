# user 项目整合演示汇报

## 1. 项目概述

本次工作将 `postgres-gormgen-9.20` 中的 PostgreSQL + GORM Gen 数据访问能力整合到 `user` 项目中，同时保留原有的登录、用户增删改查、个人资料、刷新 Token、登出、SSE、文件上传和 Kafka 演示功能。

本次目标不是构建生产级平台，而是完成一个结构清晰、功能完整、可以本地启动和演示的 Go-zero 微服务示例。

## 2. 最终架构

```text
┌──────────────┐
│  HTTP Client  │
└──────┬───────┘
       │ HTTP :8888
       ▼
┌──────────────────────────────┐
│ user/api                     │
│                              │
│ 登录、JWT、权限、Profile、   │
│ 文件上传、SSE、刷新、登出    │
└──────┬───────────────────────┘
       │ zRPC :8080
       ▼
┌──────────────────────────────┐
│ user/rpc                     │
│                              │
│ 用户 RPC、GORM Gen、密码哈希 │
│ PostgreSQL 数据访问          │
└──────┬───────────────────────┘
       │ PostgreSQL :5432
       ▼
┌──────────────────────────────┐
│ PostgreSQL                   │
│ users、todos 表              │
└──────────────────────────────┘

辅助组件：
- Redis：JWT 黑名单、缓存相关演示
- etcd：zRPC 服务注册与发现演示
- Kafka：消息生产者、消费者演示
- Kafka UI：Kafka 消息查看
```

## 3. 模块职责

### 3.1 API 模块

目录：`api/`

主要职责：

- 对外提供 HTTP REST 接口
- 处理请求参数和响应格式
- 解析和校验 JWT
- 处理角色权限和用户资源归属
- 生成 Access Token 和 Refresh Token
- 通过 RPC 调用用户数据服务
- 实现文件上传、SSE、登出和 Token 黑名单

API 不再直接访问用户数据库，用户相关数据统一通过 `user/rpc` 获取和修改。

### 3.2 RPC 模块

目录：`rpc/`

主要职责：

- 提供用户相关 gRPC 接口
- 使用 GORM Gen 生成数据库访问代码
- 访问 PostgreSQL
- 对密码进行 bcrypt 哈希
- 支持用户查询、认证、插入、更新和软删除
- 提供 Kafka 消息生产和消费演示

### 3.3 PostgreSQL

数据库连接配置位于：

```text
rpc/etc/user.yaml
```

默认连接信息：

```text
host=127.0.0.1
port=5432
database=appdb
user=devuser
password=devpass
```

初始化脚本位于：

```text
rpc/migrations/0001_init.sql
```

## 4. 已完成的核心功能

### 4.1 用户数据访问迁移到 RPC

原来的 API 用户逻辑直接访问数据库；现在改为：

```text
API Logic → userclient → zRPC → RPC Logic → GORM Gen → PostgreSQL
```

API 通过 `ServiceContext.RpcContext` 自动添加内部调用 metadata：

```text
token: internal
```

### 4.2 用户角色 role

用户表增加了角色字段：

```sql
role VARCHAR(32) NOT NULL DEFAULT 'user'
```

角色会参与以下流程：

1. 用户登录时从 RPC 获取角色
2. API 将角色写入 JWT Claims
3. JWT 中包含：

```json
{
  "userId": 1,
  "role": "admin"
}
```

4. 管理员接口从 JWT 中读取 `role`
5. Profile 可以查看当前用户角色，但普通用户不能通过 Profile 修改角色

### 4.3 密码安全

注册或插入用户时使用 bcrypt 哈希密码：

```text
明文密码 → bcrypt.GenerateFromPassword → 数据库存储哈希
```

登录时使用：

```text
bcrypt.CompareHashAndPassword
```

数据库和普通用户查询接口不会直接返回密码哈希。

### 4.4 PostgreSQL 软删除

用户模型使用 `deleted_at` 实现软删除。

用户名和邮箱使用部分唯一索引：

```sql
CREATE UNIQUE INDEX idx_users_username
ON users (username) WHERE deleted_at IS NULL;
```

因此，软删除后的用户名和邮箱可以重新使用。

### 4.5 etcd

etcd 用于演示服务注册与发现。

RPC 配置中已经配置：

```yaml
Etcd:
  Hosts:
    - 127.0.0.1:2379
  Key: user.rpc
```

含义：

- RPC 服务将服务地址注册到 etcd
- API 可以通过 etcd 的服务名发现 RPC
- 在单机演示中，也可以继续使用固定地址 `127.0.0.1:8080`

当前 API 使用固定 Endpoint：

```yaml
UserRpc:
  Endpoints:
    - 127.0.0.1:8080
```

因此 etcd 当前主要用于展示 go-zero 服务发现配置，API 的实际调用仍走固定地址。

## 5. HTTP 接口职责

| 接口 | 方法 | 是否需要 JWT | 主要用途 |
|---|---:|---:|---|
| `/user/login` | POST | 否 | 用户登录，返回 Access Token 和 Refresh Token |
| `/user/refresh` | POST | 否 | 使用 Refresh Token 获取新 Token |
| `/user/:id` | GET | 否 | 公开基础用户查询 |
| `/user/:id` | POST | 是 | 管理员修改用户基础资料 |
| `/user/profile/:userId` | GET | 是 | 当前用户查看自己的资料 |
| `/user/profile/:userId` | PUT | 是 | 当前用户修改自己的用户名 |
| `/user/logout` | POST | 是 | 将当前 Access Token 加入 Redis 黑名单 |
| `/upload/file` | POST | 否 | 文件上传演示 |
| `/sse/with/req` | POST | 否 | SSE 请求响应演示 |
| `/sse/without/resp` | POST | 否 | SSE 无响应体演示 |

### 5.1 GetUserInfo

```http
GET /user/:id
```

公开的基础用户查询，只返回：

```json
{
  "id": 1,
  "username": "demo",
  "email": "demo@example.com"
}
```

不会返回密码和其他认证敏感数据。

### 5.2 UpdateUserInfo

```http
POST /user/:id
Authorization: Bearer <access-token>
```

只有 JWT 中 `role` 为 `admin` 的用户可以调用。

支持修改：

- username
- email

HTTP 请求不接受 `role` 字段，避免普通用户通过 API 修改自己的角色。

### 5.3 GetProfile

```http
GET /user/profile/:userId
Authorization: Bearer <access-token>
```

只允许 JWT 中的 `userId` 与路径参数一致。

返回当前用户的：

- userId
- username
- role

### 5.4 UpdateProfile

```http
PUT /user/profile/:userId
Authorization: Bearer <access-token>
```

只允许用户修改自己的资料，目前仅允许修改用户名。

请求示例：

```json
{
  "username": "new-name"
}
```

即使请求额外提交：

```json
{
  "username": "new-name",
  "role": "admin"
}
```

API 契约也不会将 `role` 传入 RPC 更新请求。

## 6. 数据库表

### 6.1 users

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | BIGSERIAL | 用户 ID |
| `username` | VARCHAR(64) | 用户名，未删除用户唯一 |
| `email` | VARCHAR(128) | 邮箱，未删除用户唯一 |
| `password` | VARCHAR(256) | bcrypt 密码哈希 |
| `role` | VARCHAR(32) | 用户角色，默认 `user` |
| `created_at` | TIMESTAMPTZ | 创建时间 |
| `updated_at` | TIMESTAMPTZ | 修改时间 |
| `deleted_at` | TIMESTAMPTZ | 软删除时间 |

### 6.2 todos

迁移脚本同时保留了 `todos` 示例表，用于展示 PostgreSQL 表结构、索引和外键配置。

## 7. GORM Gen 代码

生成器入口：

```text
rpc/cmd/gen/main.go
```

生成结果：

```text
rpc/internal/dal/model/
rpc/internal/dal/query/
```

核心生成内容：

- `model.User`
- `query.User`
- 字段类型和字段映射
- GORM Gen 查询方法

RPC 业务逻辑通过类似代码访问数据库：

```go
l.svcCtx.PostgresQuery.User.WithContext(l.ctx).
    Where(l.svcCtx.PostgresQuery.User.ID.Eq(id)).
    First()
```

## 8. 本地启动步骤

### 8.1 启动基础设施

在项目根目录执行：

```bash
docker compose up -d postgres etcd kafka kafka-ui
```

查看容器状态：

```bash
docker compose ps
```

当前端口：

| 服务 | 端口 | 用途 |
|---|---:|---|
| PostgreSQL | 5432 | 用户数据库 |
| etcd | 2379 | 客户端通信和服务发现 |
| etcd | 2380 | 节点通信 |
| Kafka | 9092/29092 | 消息队列 |
| Kafka UI | 8081 | Kafka 管理页面 |

Redis 需要本机另行启动，默认配置：

```text
127.0.0.1:6379
```

Redis 用于：

- JWT 黑名单
- 相关缓存演示

### 8.2 启动 RPC

在项目根目录执行：

```bash
go run ./rpc -f rpc/etc/user.yaml
```

RPC 地址：

```text
127.0.0.1:8080
```

### 8.3 启动 API

另开一个终端执行：

```bash
go run ./api -f api/etc/user-api.yaml
```

API 地址：

```text
127.0.0.1:8888
```

## 9. 演示前准备

### 9.1 查看用户数据

```bash
docker exec postgres psql -U devuser -d appdb \
  -c "select id, username, email, role from users;"
```

### 9.2 设置管理员用户

选择一个演示用户设置为管理员：

```bash
docker exec postgres psql -U devuser -d appdb \
  -c "update users set role = 'admin' where id = 1;"
```

确认角色：

```bash
docker exec postgres psql -U devuser -d appdb \
  -c "select id, username, role from users where id = 1;"
```

如果用户的密码未知，可以通过 RPC 插入一个新用户，或准备一个已知密码的演示用户。

## 10. 推荐演示流程

### 10.1 查询公开用户

```bash
curl http://127.0.0.1:8888/user/1
```

预期返回：

```json
{
  "id": 1,
  "username": "demo",
  "email": "demo@example.com"
}
```

### 10.2 登录

```bash
curl -X POST http://127.0.0.1:8888/user/login \
  -H "Content-Type: application/json" \
  -d '{"username":"demo","password":"demo123"}'
```

预期返回：

```json
{
  "accessToken": "...",
  "refreshToken": "...",
  "expiresIn": 86400
}
```

保存返回的 `accessToken`：

```bash
TOKEN="<accessToken>"
```

### 10.3 查看个人资料

```bash
curl http://127.0.0.1:8888/user/profile/1 \
  -H "Authorization: Bearer $TOKEN"
```

### 10.4 修改个人资料

```bash
curl -X PUT http://127.0.0.1:8888/user/profile/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"username":"new-demo"}'
```

### 10.5 管理员修改用户信息

使用管理员用户登录后：

```bash
curl -X POST http://127.0.0.1:8888/user/2 \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"username":"updated-user","email":"updated@example.com"}'
```

普通用户调用该接口应返回 403。

### 10.6 刷新 Token

```bash
curl -X POST http://127.0.0.1:8888/user/refresh \
  -H "Content-Type: application/json" \
  -d '{"refreshToken":"<refreshToken>"}'
```

### 10.7 登出和黑名单

```bash
curl -X POST http://127.0.0.1:8888/user/logout \
  -H "Authorization: Bearer $TOKEN"
```

登出后再次访问受保护接口，Token 应被 Redis 黑名单拦截。

## 11. 已完成的验证

本次实际验证结果：

```text
go test ./...       通过
go build ./...      通过
```

运行验证：

- PostgreSQL 容器可以启动
- `users` 表包含 `role` 字段
- RPC 服务监听 `0.0.0.0:8080`
- API 服务监听 `0.0.0.0:8888`
- `GET /user/1` 返回 HTTP 200
- 不存在用户返回业务码 404
- 错误登录返回业务码 422
- 未认证访问 Profile 返回 HTTP 401
- 未认证调用管理员更新接口返回 HTTP 401
- 未认证调用 Logout 返回 HTTP 401
- API 可以通过 RPC 查询 PostgreSQL 用户数据

## 12. 已知限制和演示说明

### 12.1 demo 级 RPC 内部校验

当前 RPC interceptor 主要检查请求是否包含 metadata：

```text
token: internal
```

这是为了演示 API 到 RPC 的内部调用保护，不是生产级服务认证方案。生产环境应进一步使用 mTLS、服务凭证或更严格的服务间鉴权。

### 12.2 默认配置仅用于本地演示

当前配置中使用了本地演示凭据，例如：

```text
PostgreSQL 用户名：devuser
PostgreSQL 密码：devpass
JWT Secret：示例字符串
```

这些配置只适用于本地开发和演示，不应直接用于生产环境。

### 12.3 etcd 当前未用于 API 动态发现

RPC 已配置 etcd，API 当前仍使用固定 RPC Endpoint：

```yaml
UserRpc:
  Endpoints:
    - 127.0.0.1:8080
```

如需完整演示服务注册发现，可以将 API 配置改为 Etcd Hosts + Key，并确保 RPC 服务注册到同一个 etcd。

### 12.4 Kafka 是独立演示能力

Kafka 主要用于消息生产者和消费者示例，不是用户登录和用户 CRUD 的核心依赖。

如果只演示用户服务，可以优先启动：

```bash
docker compose up -d postgres etcd
```

如果要演示 Kafka，再启动：

```bash
docker compose up -d kafka kafka-ui
```

## 13. 总结

本次整合完成了以下目标：

1. 将 PostgreSQL 和 GORM Gen 引入 `user/rpc`
2. 将 API 用户数据访问统一迁移到 RPC
3. 保留登录、刷新、登出、用户 CRUD、Profile、SSE、上传和 Kafka 等演示功能
4. 增加 `role` 字段并接入 JWT 权限控制
5. 区分公开用户查询、管理员更新和当前用户 Profile 操作
6. 使用 bcrypt 保存和验证密码
7. 使用 PostgreSQL 软删除和部分唯一索引
8. 增加 Docker Compose 中的 PostgreSQL、etcd、Kafka 和 Kafka UI 配置
9. 通过 `go test ./...`、`go build ./...` 和 HTTP 冒烟请求验证核心链路

整体项目可以作为 Go-zero REST + zRPC + PostgreSQL + GORM Gen + Redis + etcd + Kafka 的综合演示项目。
