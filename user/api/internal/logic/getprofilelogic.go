package logic

import (
	"context"
	"encoding/json"
	"user/api/internal/svc"
	"user/api/internal/types"
	"user/rpc/userclient"

	"github.com/zeromicro/x/errors"
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
		return nil, errors.New(401, "未认证")
	}

	id, err := userID.Int64()
	if err != nil {
		return nil, errors.New(401, "无效的用户 ID")
	}

	// 验证权限（只能查看自己的信息）
	if id != req.UserId {
		return nil, errors.New(401, "无权访问")
	}

	rpcUser, err := l.svcCtx.UserRpc.QueryUser(
		l.svcCtx.RpcContext(l.ctx),
		&userclient.QueryUserReq{Id: req.UserId},
	)
	if err != nil {
		return nil, errors.New(401, "用户不存在")
	}

	return &types.ProfileResp{
		UserId:   rpcUser.GetId(),
		Username: rpcUser.GetUsername(),
		Role:     rpcUser.GetRole(),
	}, nil
}
