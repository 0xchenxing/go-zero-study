// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"
	"encoding/json"
	"user/api/common/errcode"
	"user/api/internal/svc"
	"user/api/internal/types"
	"user/rpc/userclient"

	"github.com/zeromicro/go-zero/core/logx"
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
		return nil, errcode.ErrUserTokenInvalid
	}

	id, err := userID.Int64()
	if err != nil {
		return nil, errcode.ErrUserTokenInvalid
	}
	if id != req.UserId {
		return nil, errcode.ErrForbidden
	}

	rpcResp, err := l.svcCtx.UserRpc.UpdateUser(
		l.svcCtx.RpcContext(l.ctx),
		&userclient.UpdateUserReq{
			Id:       req.UserId,
			Username: req.Username,
		},
	)
	if err != nil {
		return nil, mapRpcError(err)
	}

	return &types.UpdateProfileResp{Success: rpcResp.GetSuccess()}, nil
}
