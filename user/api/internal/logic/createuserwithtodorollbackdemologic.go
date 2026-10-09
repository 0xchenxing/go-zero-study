// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"
	"user/api/common/errcode"
	"user/api/internal/svc"
	"user/api/internal/types"
	"user/rpc/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateUserWithTodoRollbackDemoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateUserWithTodoRollbackDemoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserWithTodoRollbackDemoLogic {
	return &CreateUserWithTodoRollbackDemoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreateUserWithTodoRollbackDemo 场景 A 的回滚演示：RPC 侧第二步故意违反外键，
// 事务整体回滚。接口返回成功，由 message 说明"已回滚"，用户不会残留。
func (l *CreateUserWithTodoRollbackDemoLogic) CreateUserWithTodoRollbackDemo(req *types.CreateUserWithTodoReq) (resp *types.RollbackDemoResp, err error) {
	if req.Username == "" || req.Email == "" || req.Password == "" || req.TodoTitle == "" {
		return nil, errcode.NewError(errcode.ErrBadRequest.Code, "用户名、邮箱、密码和待办标题不能为空")
	}

	rpcResp, err := l.svcCtx.UserRpc.CreateUserWithTodoRollbackDemo(
		l.svcCtx.RpcContext(l.ctx),
		&userclient.RollbackDemoReq{
			Username:  req.Username,
			Email:     req.Email,
			Password:  req.Password,
			TodoTitle: req.TodoTitle,
		},
	)
	if err != nil {
		return nil, mapRpcError(err)
	}

	return &types.RollbackDemoResp{Message: rpcResp.GetMessage()}, nil
}
