// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"
	"fmt"
	"user/api/internal/svc"
	"user/api/internal/types"
	"user/rpc/user"

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
	rpcResp, err := l.svcCtx.UserRpc.Ping(l.ctx, &user.Request{
		Ping: fmt.Sprintf("user_id:%d", req.Id),
	})
	if err != nil {
		return nil, err
	}

	return &types.UserInfoResp{
		Id:       req.Id,
		Username: rpcResp.Pong,
		Email:    "cyk@example.com",
	}, nil
}
