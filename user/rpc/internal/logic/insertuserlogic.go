package logic

import (
	"context"
	"user/rpc/internal/svc"
	"user/rpc/model/mysql"
	"user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type InsertUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInsertUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InsertUserLogic {
	return &InsertUserLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *InsertUserLogic) InsertUser(in *user.InsertUserReq) (*user.InsertUserResp, error) {
	data := &mysql.User{
		Username: in.Username,
		Password: in.Password,
		Mobile:   in.Mobile,
	}
	_, err := l.svcCtx.UserModel.Insert(l.ctx, data)
	if err != nil {
		return nil, err
	}
	return &user.InsertUserResp{Id: data.Id}, nil
}
