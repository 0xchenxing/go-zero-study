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
)

type LogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LogoutLogic) Logout() (resp *types.LogoutResp, err error) {
	// 获取 Header 中的 Authorization Token (需在 handler 处理传入 context)
	// 在 go-zero 中，通常可以通过 jwt 中间件后的 context 获取
	tokenStr, _ := l.ctx.Value("jwtToken").(string)
	if tokenStr == "" {
		return &types.LogoutResp{Success: false, Message: "无法获取 Token"}, nil
	}

	token, _, _ := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	claims := token.Claims.(jwt.MapClaims)
	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return &types.LogoutResp{Success: false, Message: "Token 缺失过期时间"}, nil
	}
	exp := int64(expFloat)
	ttl := time.Until(time.Unix(exp, 0))

	if ttl > 0 {
		err = l.svcCtx.Redis.Setex("jwt:revoked:"+tokenStr, "1", int(ttl.Seconds()))
		if err != nil {
			return nil, err
		}
	}

	return &types.LogoutResp{Success: true}, nil
}
