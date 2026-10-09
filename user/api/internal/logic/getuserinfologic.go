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

type GetUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserInfoLogic) GetUserInfo(req *types.UserInfoReq) (resp *types.UserInfoResp, err error) {
	if req.Id <= 0 {
		return nil, errcode.NewError(errcode.ErrBadRequest.Code, "用户 ID 无效")
	}

	rpcResp, err := l.svcCtx.UserRpc.QueryUser(
		l.svcCtx.RpcContext(l.ctx),
		&userclient.QueryUserReq{Id: req.Id},
	)
	if err != nil {
		return nil, mapRpcError(err)
	}

	return &types.UserInfoResp{
		Id:       rpcResp.GetId(),
		Username: rpcResp.GetUsername(),
		Email:    rpcResp.GetEmail(),
	}, nil
}
