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

type CreateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreateUser 场景 B：新增用户。账号或邮箱重复时由 RPC 返回唯一约束冲突，
// 这里映射成 409 Conflict 的业务错误。
func (l *CreateUserLogic) CreateUser(req *types.CreateUserReq) (resp *types.CreateUserResp, err error) {
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return nil, errcode.NewError(errcode.ErrBadRequest.Code, "用户名、邮箱和密码不能为空")
	}

	rpcResp, err := l.svcCtx.UserRpc.InsertUser(
		l.svcCtx.RpcContext(l.ctx),
		&userclient.InsertUserReq{
			Username: req.Username,
			Email:    req.Email,
			Password: req.Password,
		},
	)
	if err != nil {
		return nil, mapRpcError(err)
	}

	return &types.CreateUserResp{Id: rpcResp.GetId()}, nil
}
