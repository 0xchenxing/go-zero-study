// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"
	"time"
	"user/api/common/errcode"
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
	// token 由 handler 从 Authorization 头提取后放入 context。
	// 拿不到 token 属于业务结果，由响应体告知客户端，不按错误处理。
	tokenStr, _ := l.ctx.Value("jwtToken").(string)
	if tokenStr == "" {
		return &types.LogoutResp{Success: false, Message: "无法获取 Token"}, nil
	}

	// 仅需读取过期时间用于计算黑名单 TTL，因此不校验签名。
	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return nil, errcode.ErrUserTokenInvalid.WithDetail(err.Error())
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errcode.ErrUserTokenInvalid
	}
	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return &types.LogoutResp{Success: false, Message: "Token 缺失过期时间"}, nil
	}

	if ttl := time.Until(time.Unix(int64(expFloat), 0)); ttl > 0 {
		if err := l.svcCtx.Redis.Setex("jwt:revoked:"+tokenStr, "1", int(ttl.Seconds())); err != nil {
			return nil, errcode.ErrServiceBusy.WithDetail(err.Error())
		}
	}

	return &types.LogoutResp{Success: true}, nil
}
