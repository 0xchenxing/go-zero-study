// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"

	"user/api/internal/svc"
	"user/api/internal/types"

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

	user, _ := l.svcCtx.UserModel.FindOne(l.ctx, userId)
	newAccess, _ := generateAccessToken(
		l.svcCtx.Config.Auth.AccessSecret, userId, user.Role)
	newRefresh, _ := generateRefreshToken(
		l.svcCtx.Config.Auth.RefreshSecret, l.svcCtx.Config.Auth.RefreshExpire, userId)

	return &types.LoginResp{
		AccessToken:  newAccess,
		RefreshToken: newRefresh,
		ExpiresIn:    l.svcCtx.Config.Auth.AccessExpire,
	}, nil
}
