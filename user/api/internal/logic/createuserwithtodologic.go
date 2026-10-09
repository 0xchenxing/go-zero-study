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

type CreateUserWithTodoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateUserWithTodoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserWithTodoLogic {
	return &CreateUserWithTodoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreateUserWithTodo 场景 A：在同一个事务里创建用户与待办。
// 任一步失败整体回滚，接口不会返回只落库一半的数据。
func (l *CreateUserWithTodoLogic) CreateUserWithTodo(req *types.CreateUserWithTodoReq) (resp *types.CreateUserWithTodoResp, err error) {
	if req.Username == "" || req.Email == "" || req.Password == "" || req.TodoTitle == "" {
		return nil, errcode.NewError(errcode.ErrBadRequest.Code, "用户名、邮箱、密码和待办标题不能为空")
	}

	rpcResp, err := l.svcCtx.UserRpc.CreateUserWithTodo(
		l.svcCtx.RpcContext(l.ctx),
		&userclient.CreateUserWithTodoReq{
			Username:  req.Username,
			Email:     req.Email,
			Password:  req.Password,
			TodoTitle: req.TodoTitle,
		},
	)
	if err != nil {
		return nil, mapRpcError(err)
	}

	return &types.CreateUserWithTodoResp{
		UserId:    rpcResp.GetUserId(),
		Username:  rpcResp.GetUsername(),
		Email:     rpcResp.GetEmail(),
		TodoId:    rpcResp.GetTodoId(),
		TodoTitle: rpcResp.GetTodoTitle(),
	}, nil
}
