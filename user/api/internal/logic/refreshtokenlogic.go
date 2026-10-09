// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"
	"user/api/common/errcode"
	"user/api/internal/svc"
	"user/api/internal/types"
	"user/rpc/userclient"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshTokenLogic) RefreshToken(req *types.RefreshReq) (*types.LoginResp, error) {
	token, err := jwt.ParseWithClaims(req.RefreshToken, jwt.MapClaims{},
		func(t *jwt.Token) (any, error) {
			return []byte(l.svcCtx.Config.Auth.RefreshSecret), nil
		})
	if err != nil || !token.Valid {
		return nil, errcode.ErrUserTokenInvalid
	}

	// 签名合法不代表载荷完整，此处必须做带 ok 的断言，避免畸形 token 触发 panic。
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errcode.ErrUserTokenInvalid
	}
	rawID, ok := claims["userId"].(float64)
	if !ok {
		return nil, errcode.ErrUserTokenInvalid
	}
	userId := int64(rawID)

	rpcUser, err := l.svcCtx.UserRpc.QueryUser(
		l.svcCtx.RpcContext(l.ctx),
		&userclient.QueryUserReq{Id: userId},
	)
	if err != nil {
		return nil, mapRpcError(err)
	}

	newAccess, err := generateAccessToken(
		l.svcCtx.Config.Auth.AccessSecret, userId, rpcUser.GetRole())
	if err != nil {
		return nil, errcode.ErrUserTokenGenerate.WithDetail(err.Error())
	}
	newRefresh, err := generateRefreshToken(
		l.svcCtx.Config.Auth.RefreshSecret, l.svcCtx.Config.Auth.RefreshExpire, userId)
	if err != nil {
		return nil, errcode.ErrUserTokenGenerate.WithDetail(err.Error())
	}

	return &types.LoginResp{
		AccessToken:  newAccess,
		RefreshToken: newRefresh,
		ExpiresIn:    l.svcCtx.Config.Auth.AccessExpire,
	}, nil
}
