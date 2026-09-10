// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"
	"time"
	"user/api/internal/svc"
	"user/api/internal/types"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/x/errors"
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
	// 查找用户
	user, err := l.svcCtx.UserModel.FindOneByUsername(l.ctx, req.Username)
	if err != nil || user == nil {
		return nil, errors.New(422, "用户名或密码错误")
	}

	// 验证密码（实际使用 bcrypt）
	if req.Password != user.Password {
		return nil, errors.New(422, "用户名或密码错误")
	}

	// 生成 Access Token
	accessToken, err := generateAccessToken(l.svcCtx.Config.Auth.AccessSecret, user.ID, user.Role)
	if err != nil {
		return nil, errors.New(422, "生成 token 失败")
	}

	// 生成 Refresh Token
	refreshToken, _ := generateRefreshToken(l.svcCtx.Config.Auth.RefreshSecret, l.svcCtx.Config.Auth.RefreshExpire, user.ID)
	if err != nil {
		return nil, errors.New(422, "生成 refresh token 失败")
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
