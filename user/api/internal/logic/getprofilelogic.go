package logic

import (
	"context"
	"encoding/json"
	"user/api/common/errcode"
	"user/api/internal/svc"
	"user/api/internal/types"
	"user/rpc/userclient"
)

type GetProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProfileLogic {
	return &GetProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProfileLogic) GetProfile(req *types.ProfileReq) (*types.ProfileResp, error) {
	// 从上下文读取 JWT Claims
	userID, ok := l.ctx.Value("userId").(json.Number)
	if !ok {
		return nil, errcode.ErrUserTokenInvalid
	}

	id, err := userID.Int64()
	if err != nil {
		return nil, errcode.ErrUserTokenInvalid
	}

	// 权限校验：只能查看自己的信息
	if id != req.UserId {
		return nil, errcode.ErrForbidden
	}

	rpcUser, err := l.svcCtx.UserRpc.QueryUser(
		l.svcCtx.RpcContext(l.ctx),
		&userclient.QueryUserReq{Id: req.UserId},
	)
	if err != nil {
		// 用户不存在与服务故障要区分开，后者不能被伪装成"用户不存在"。
		return nil, mapRpcError(err)
	}

	return &types.ProfileResp{
		UserId:   rpcUser.GetId(),
		Username: rpcUser.GetUsername(),
		Role:     rpcUser.GetRole(),
	}, nil
}
