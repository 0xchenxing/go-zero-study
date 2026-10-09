package logic

import (
	"user/api/common/errcode"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mapRpcError 把 RPC 返回的 gRPC 状态码翻译成 API 层的业务错误。
//
// 这是 RPC 与 API 之间唯一的错误协议边界：RPC 侧只表达 gRPC 语义，
// 业务码和对外文案统一在 API 侧决定，避免两个服务各维护一套错误码。
// 原始错误通过 WithDetail 附加，只进日志不出响应。
func mapRpcError(err error) error {
	if err == nil {
		return nil
	}

	st := status.Convert(err)
	switch st.Code() {
	case codes.InvalidArgument:
		// RPC 的参数校验文案比通用文案更具体（如"用户 ID 无效"），直接沿用。
		return errcode.NewError(errcode.ErrBadRequest.Code, st.Message()).
			WithDetail(err.Error())
	case codes.NotFound:
		return errcode.ErrUserNotFound.WithDetail(err.Error())
	case codes.AlreadyExists:
		// 重复的是用户名还是邮箱由 RPC 判定，文案直接沿用。
		return errcode.NewError(errcode.ErrUserAlreadyExists.Code, st.Message()).
			WithDetail(err.Error())
	case codes.FailedPrecondition:
		return errcode.ErrBadRequest.WithDetail(err.Error())
	case codes.DeadlineExceeded:
		return errcode.ErrTimeout.WithDetail(err.Error())
	case codes.Unavailable:
		return errcode.ErrServiceBusy.WithDetail(err.Error())
	default:
		return errcode.ErrInternalServer.WithDetail(err.Error())
	}
}

// mapAuthenticateError 翻译登录场景的用户查询错误。
//
// 用户不存在与密码错误返回同一个业务码，避免暴露账号是否已注册；
// 其余错误属于服务故障，不能伪装成密码错误，否则会把 RPC 抖动
// 误报成账号问题，误导用户和排障方向。
func mapAuthenticateError(err error) error {
	if status.Code(err) == codes.NotFound {
		return errcode.ErrUserPasswordInvalid
	}
	return mapRpcError(err)
}
