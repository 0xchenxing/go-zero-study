package errcode

// 系统级通用错误（9xxx 段）。
//
// 这类错误与具体业务无关，所有模块都可以直接复用：
//
//	return nil, errcode.ErrBadRequest
//
// 错误码一经对外暴露就不可再修改含义，废弃时只能标注保留。
// 新增通用错误请在该段内继续取号。
var (
	ErrSuccess = NewError(0, "成功")

	ErrInternalServer = NewError(9001, "系统内部错误，请稍后再试")
	ErrBadRequest     = NewError(9002, "请求参数错误")
	ErrUnauthorized   = NewError(9003, "身份验证失败，请重新登录")
	ErrForbidden      = NewError(9004, "权限不足，拒绝访问")
	ErrNotFound       = NewError(9005, "请求的资源不存在")
	ErrServiceBusy    = NewError(9006, "系统繁忙，请稍后再试")
	ErrTimeout        = NewError(9007, "请求处理超时")

	// 文件上传
	ErrUploadFileType = NewError(9008, "不支持的文件类型")
	ErrUploadFileSize = NewError(9009, "文件大小超出限制")
)
