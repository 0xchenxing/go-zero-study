// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"
	"time"
	"user/api/common/errcode"
	"user/api/internal/svc"
	"user/api/internal/types"
	"user/rpc/userclient"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	rpcUser, err := l.svcCtx.UserRpc.AuthenticateUser(l.svcCtx.RpcContext(l.ctx), &userclient.AuthenticateUserReq{Username: req.Username})
	if err != nil {
		return nil, mapAuthenticateError(err)
	}

	if bcrypt.CompareHashAndPassword([]byte(rpcUser.GetPassword()), []byte(req.Password)) != nil {
		return nil, errcode.ErrUserPasswordInvalid
	}

	accessToken, err := generateAccessToken(l.svcCtx.Config.Auth.AccessSecret, rpcUser.GetId(), rpcUser.GetRole())
	if err != nil {
		return nil, errcode.ErrUserTokenGenerate.WithDetail(err.Error())
	}
	refreshToken, err := generateRefreshToken(l.svcCtx.Config.Auth.RefreshSecret, l.svcCtx.Config.Auth.RefreshExpire, rpcUser.GetId())
	if err != nil {
		return nil, errcode.ErrUserTokenGenerate.WithDetail(err.Error())
	}

	return &types.LoginResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    l.svcCtx.Config.Auth.AccessExpire,
	}, nil
}

func generateAccessToken(secret string, userId int64, role string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"userId": userId,
		"role":   role,
		"iat":    now.Unix(),
		"exp":    now.Add(24 * time.Hour).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(secret))
}

func generateRefreshToken(secret string, expire int64, userID int64) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"userId": userID,
		"iat":    now.Unix(),
		"exp":    now.Add(time.Duration(expire) * time.Second).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(secret))
}
