// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"

	"fmt"
	"time"

	"user/api/internal/svc"
	"user/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SseLogic {
	return &SseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SseLogic) Sse(req *types.SseReq, client chan<- *types.SseResp) error {
	for i := 0; i < 5; i++ {
		client <- &types.SseResp{
			Msg: fmt.Sprintf("req: %s, message %d", req.Body, i),
		}
		time.Sleep(time.Second)
	}

	return nil
}
