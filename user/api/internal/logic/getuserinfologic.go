// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"
	"user/api/internal/svc"
	"user/api/internal/types"
	"user/rpc/userclient"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/x/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	rpcResp, err := l.svcCtx.UserRpc.QueryUser(
		l.svcCtx.RpcContext(l.ctx),
		&userclient.QueryUserReq{Id: req.Id},
	)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, errors.New(404, "用户不存在")
		}
		return nil, err
	}

	return &types.UserInfoResp{
		Id:       rpcResp.GetId(),
		Username: rpcResp.GetUsername(),
		Email:    rpcResp.GetEmail(),
	}, nil
}
