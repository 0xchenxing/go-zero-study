package logic

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/x/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mapRpcError 把 RPC 的 gRPC 状态码转换成 API 业务错误。
// 业务提示沿用 RPC 侧给出的文案（例如"用户名已被占用"），
// 无法识别的错误统一收敛成服务端错误，不把底层细节返回给调用方。
func mapRpcError(err error) error {
	st := status.Convert(err)
	switch st.Code() {
	// 唯一约束冲突：用户名或邮箱已被占用 -> 409 Conflict。
	case codes.AlreadyExists:
		return errors.New(http.StatusConflict, st.Message())
	// 外键约束冲突：关联数据不存在 -> 409 Conflict。
	case codes.FailedPrecondition:
		return errors.New(http.StatusConflict, st.Message())
	case codes.NotFound:
		return errors.New(http.StatusNotFound, "用户不存在")
	case codes.InvalidArgument:
		return errors.New(http.StatusBadRequest, st.Message())
	default:
		logx.Errorf("rpc call failed: %v", err)
		return errors.New(http.StatusInternalServerError, "系统繁忙，请稍后重试")
	}
}
