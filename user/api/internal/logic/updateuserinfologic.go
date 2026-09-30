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
)

type UpdateUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserInfoLogic {
	return &UpdateUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserInfoLogic) UpdateUserInfo(req *types.UserUpdateReq) (resp *types.UserUpdateResp, err error) {
	role, ok := l.ctx.Value("role").(string)
	if !ok || role != "admin" {
		return nil, errors.New(403, "无权访问")
	}

	_, err = l.svcCtx.UserRpc.UpdateUser(
		l.svcCtx.RpcContext(l.ctx),
		&userclient.UpdateUserReq{
			Id:       req.Id,
			Username: req.Username,
			Email:    req.Email,
		},
	)
	if err != nil {
		return nil, err
	}

	return &types.UserUpdateResp{Info: "success"}, nil
}
