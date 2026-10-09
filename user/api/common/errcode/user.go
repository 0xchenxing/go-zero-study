package errcode

// 用户模块错误（1xxx 段）。
//
// 1xxx 整段划给用户域（注册、登录、资料、凭证），
// 新增用户相关错误请在该段内继续取号：
//
//	return nil, errcode.ErrUserNotFound
var (
	ErrUserAlreadyExists   = NewError(1001, "用户已存在")
	ErrUserNotFound        = NewError(1002, "用户不存在")
	ErrUserPasswordInvalid = NewError(1003, "用户名或密码错误")
	ErrUserTokenInvalid    = NewError(1005, "登录凭证已失效，请重新登录")
	ErrUserTokenGenerate   = NewError(1009, "生成 token 失败")
)
