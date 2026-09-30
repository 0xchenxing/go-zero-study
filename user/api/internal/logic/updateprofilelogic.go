// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"
	"encoding/json"
	"user/api/internal/svc"
	"user/api/internal/types"
	"user/rpc/userclient"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/x/errors"
)

type UpdateProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProfileLogic {
	return &UpdateProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateProfileLogic) UpdateProfile(req *types.UpdateProfileReq) (resp *types.UpdateProfileResp, err error) {
	userID, ok := l.ctx.Value("userId").(json.Number)
	if !ok {
		return nil, errors.New(401, "未认证")
	}

	id, err := userID.Int64()
	if err != nil {
		return nil, errors.New(401, "无效的用户 ID")
	}
	if id != req.UserId {
		return nil, errors.New(401, "无权访问")
	}

	rpcResp, err := l.svcCtx.UserRpc.UpdateUser(
		l.svcCtx.RpcContext(l.ctx),
		&userclient.UpdateUserReq{
			Id:       req.UserId,
			Username: req.Username,
		},
	)
	if err != nil {
		return nil, err
	}

	return &types.UpdateProfileResp{Success: rpcResp.GetSuccess()}, nil
}
