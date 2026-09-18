package xerr

import "fmt"

// CodeError 自定义业务错误类型
type CodeError struct {
	errCode uint32
	errMsg  string
}

// GetErrCode 返回给前端的错误码
func (e *CodeError) GetErrCode() uint32 {
	return e.errCode
}

// GetErrMsg 返回给前端显示的错误信息
func (e *CodeError) GetErrMsg() string {
	return e.errMsg
}

// Error 实现 error 接口，用于日志记录
func (e *CodeError) Error() string {
	return fmt.Sprintf("ErrCode:%d, ErrMsg:%s", e.errCode, e.errMsg)
}

// NewErrCodeMsg 使用错误码和自定义错误信息创建 CodeError
func NewErrCodeMsg(errCode uint32, errMsg string) *CodeError {
	return &CodeError{errCode: errCode, errMsg: errMsg}
}

// NewErrCode 使用错误码自动匹配错误信息创建 CodeError
func NewErrCode(errCode uint32) *CodeError {
	return &CodeError{errCode: errCode, errMsg: MapErrMsg(errCode)}
}

// NewErrMsg 使用默认错误码和自定义错误信息创建 CodeError
func NewErrMsg(errMsg string) *CodeError {
	return &CodeError{errCode: SERVER_COMMON_ERROR, errMsg: errMsg}
}
