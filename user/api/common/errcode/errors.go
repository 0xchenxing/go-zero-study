// Package errcode 定义全项目统一的业务错误码。
//
// 业务逻辑只返回本包中的错误值，不关心 HTTP 状态码与响应格式：
// 前者的转换由 response 包负责，后者由全局错误处理器负责。
package errcode

import (
	"errors"
	"fmt"
)

// Error 业务错误。
//
// Code 对外暴露，供客户端判断具体分支；Message 可直接展示给用户；
// Detail 只用于日志排查，不会出现在响应体中。
type Error struct {
	Code    int
	Message string
	Detail  string
}

// Error 实现 error 接口，供日志记录使用。
func (e *Error) Error() string {
	if e.Detail == "" {
		return fmt.Sprintf("ErrCode:%d, ErrMsg:%s", e.Code, e.Message)
	}
	return fmt.Sprintf("ErrCode:%d, ErrMsg:%s, Detail:%s", e.Code, e.Message, e.Detail)
}

// Is 支持 errors.Is 按错误码比较，例如 errors.Is(err, errcode.ErrUserNotFound)。
func (e *Error) Is(target error) bool {
	var t *Error
	if errors.As(target, &t) {
		return e.Code == t.Code
	}
	return false
}

// NewError 使用错误码和文案创建业务错误。
// 需要覆盖字典文案时使用，例如把参数错误细化为"用户名不能为空"。
func NewError(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

// WithDetail 附加仅供日志使用的内部细节。
// 返回副本而不是修改自身，避免调用方改到包级共享的错误值。
func (e *Error) WithDetail(detail string) *Error {
	clone := *e
	clone.Detail = detail
	return &clone
}

// IsError 判断 error 链上是否带有业务错误。
func IsError(err error) bool {
	var e *Error
	return errors.As(err, &e)
}

// ToError 把任意 error 归一成 *Error。
//
// 已是业务错误的原样返回；数据库、gRPC 等底层错误收敛成系统内部错误，
// 避免把实现细节暴露给调用方。
func ToError(err error) *Error {
	if err == nil {
		return nil
	}

	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return ErrInternalServer.WithDetail(err.Error())
}
