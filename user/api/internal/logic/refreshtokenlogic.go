// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"

	"user/api/internal/svc"
	"user/api/internal/types"
	"user/rpc/userclient"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/x/errors"
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
		return nil, errors.New(401, "Refresh Token 已失效或非法")
	}
	claims := token.Claims.(jwt.MapClaims)
	userId := int64(claims["userId"].(float64))

	rpcUser, err := l.svcCtx.UserRpc.QueryUser(
		l.svcCtx.RpcContext(l.ctx),
		&userclient.QueryUserReq{Id: userId},
	)
	if err != nil {
		return nil, errors.New(401, "用户不存在")
	}

	newAccess, err := generateAccessToken(
		l.svcCtx.Config.Auth.AccessSecret, userId, rpcUser.GetRole())
	if err != nil {
		return nil, errors.New(401, "生成 token 失败")
	}
	newRefresh, err := generateRefreshToken(
		l.svcCtx.Config.Auth.RefreshSecret, l.svcCtx.Config.Auth.RefreshExpire, userId)
	if err != nil {
		return nil, errors.New(401, "生成 refresh token 失败")
	}

	return &types.LoginResp{
		AccessToken:  newAccess,
		RefreshToken: newRefresh,
		ExpiresIn:    l.svcCtx.Config.Auth.AccessExpire,
	}, nil
}
