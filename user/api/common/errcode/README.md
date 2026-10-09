# 错误码字典

全项目统一的业务错误码。业务逻辑只返回本包的错误值，HTTP 状态码与响应格式由
`api/common/response` 和全局错误处理器负责。

## 使用方式

```go
// 直接复用已登记的错误码
return nil, errcode.ErrUserNotFound

// 需要更具体的文案时，复用错误码、覆盖文案
return nil, errcode.NewError(errcode.ErrBadRequest.Code, "用户 ID 无效")

// 附加仅供日志的内部细节（不会出现在响应体里）
return nil, errcode.ErrInternalServer.WithDetail(err.Error())
```

判断错误：

```go
if errors.Is(err, errcode.ErrUserNotFound) { ... }
```

## 编号规则

5 位以内，按模块分段，段内连续取号并预留扩展空间：

| 段位 | 模块 | 说明 |
| --- | --- | --- |
| `0` | 成功 | 唯一的成功码 |
| `1xxx` | 用户模块 | 注册、登录、资料、登录凭证 |
| `9xxx` | 系统级通用 | 与业务无关，所有模块可复用 |

> **兼容性约定**：错误码一经对外暴露，含义不可再修改。业务下线后错误码只能标注
> 废弃并保留，绝不能复用给新的含义，否则客户端的判断分支会静默出错。

## 成功

| 码 | 文案 | 说明 |
| --- | --- | --- |
| 0 | 成功 | 所有成功响应的 code，数据在 `data` 字段 |

## 系统级通用错误（9xxx）

| 码 | 名称 | 文案 | 触发条件 |
| --- | --- | --- | --- |
| 9001 | `ErrInternalServer` | 系统内部错误，请稍后再试 | 未预期的底层错误，兜底码 |
| 9002 | `ErrBadRequest` | 请求参数错误 | 参数缺失、格式非法、ID 无效 |
| 9003 | `ErrUnauthorized` | 身份验证失败，请重新登录 | 未认证（预留给需区分认证失败的场景） |
| 9004 | `ErrForbidden` | 权限不足，拒绝访问 | 已登录但无权操作该资源 |
| 9005 | `ErrNotFound` | 请求的资源不存在 | 通用资源不存在（用户域请用 1002） |
| 9006 | `ErrServiceBusy` | 系统繁忙，请稍后再试 | 依赖服务不可用、Redis 写入失败 |
| 9007 | `ErrTimeout` | 请求处理超时 | RPC 调用超时 |
| 9008 | `ErrUploadFileType` | 不支持的文件类型 | 文件 MIME 不在白名单内 |
| 9009 | `ErrUploadFileSize` | 文件大小超出限制 | 文件超过 10 MB |

## 用户模块错误（1xxx）

| 码 | 名称 | 文案 | 触发条件 |
| --- | --- | --- | --- |
| 1001 | `ErrUserAlreadyExists` | 用户名已被占用 / 邮箱已被占用 | 命中唯一索引 `idx_users_username` / `idx_users_email`，具体文案由 RPC 判定 |
| 1002 | `ErrUserNotFound` | 用户不存在 | 按 ID 或用户名查不到记录 |
| 1003 | `ErrUserPasswordInvalid` | 用户名或密码错误 | 密码校验失败。**账号不存在也返回此码**，避免暴露账号是否已注册 |
| 1005 | `ErrUserTokenInvalid` | 登录凭证已失效，请重新登录 | Refresh Token 非法或过期、Token 已被登出拉黑、Token 载荷缺失 |
| 1009 | `ErrUserTokenGenerate` | 生成 token 失败 | 签发 Access / Refresh Token 失败 |

## 响应格式

成功与失败使用同一个信封：

```json
// 成功
{"code": 0, "msg": "成功", "data": {"id": 18}, "trace_id": "d7f2..."}

// 失败
{"code": 1002, "msg": "用户不存在", "trace_id": "d7f2..."}
```

- HTTP 状态码固定为 `200`，业务结果只由 `code` 表达
- `trace_id` 取不到链路信息时省略该字段
- 内部细节（SQL、堆栈、gRPC 原始错误）通过 `WithDetail` 附加，只写入日志，不进入响应体

## RPC 边界的错误映射

RPC 侧只表达 gRPC 语义，业务码与对外文案统一在 API 侧决定，映射见
`api/internal/logic/rpcerror.go`：

| gRPC 状态 | 业务码 |
| --- | --- |
| `InvalidArgument` | 9002（沿用 RPC 的具体校验文案） |
| `NotFound` | 1002 |
| `AlreadyExists` | 1001（沿用 RPC 判定的具体字段文案） |
| `FailedPrecondition` | 9002 |
| `DeadlineExceeded` | 9007 |
| `Unavailable` | 9006 |
| 其他 | 9001 |
