package logic

import (
	"context"

	"user/rpc/internal/svc"
	"user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type PusherLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPusherLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PusherLogic {
	return &PusherLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PusherLogic) Pusher(in *user.PusherReq) (*user.PusherResp, error) {
	data := "zhangSan"
	if err := l.svcCtx.KqPusherClient.Push(l.ctx, data); err != nil {
		logx.Errorf("KqPusherClient Push Error , err :%v", err)
		return nil, err
	}

	return &user.PusherResp{}, nil
}
