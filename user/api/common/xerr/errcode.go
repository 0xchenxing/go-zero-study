package xerr

// 成功返回
const OK uint32 = 200

// 全局错误码
// 前3位代表业务,后三位代表具体功能
const (
	SERVER_COMMON_ERROR  uint32 = 100001
	REQUEST_PARAM_ERROR  uint32 = 100002
	TOKEN_EXPIRE_ERROR   uint32 = 100003
	TOKEN_GENERATE_ERROR uint32 = 100004
	DB_ERROR             uint32 = 100005
)

// 用户模块
const (
	USER_NOT_FOUND      uint32 = 200001
	USER_ALREADY_EXISTS uint32 = 200002
	PASSWORD_ERROR      uint32 = 200003
)
