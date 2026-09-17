package logic

import (
	"context"

	"user/rpc/internal/svc"
	"user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryUserLogic {
	return &QueryUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryUserLogic) QueryUser(in *user.QueryUserReq) (*user.QueryUserResp, error) {
	u, err := l.svcCtx.UserModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &user.QueryUserResp{Id: u.Id, Username: u.Username, Mobile: u.Mobile}, nil
}
