// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package svc

import (
	"context"
	"user/api/internal/config"
	"user/api/internal/middleware"
	"user/rpc/userclient"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc/metadata"
)

type ServiceContext struct {
	Config       config.Config
	UserRpc      userclient.User
	Redis        *redis.Redis
	JwtBlacklist rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	r := redis.New(c.Redis.Host, redis.WithPass(c.Redis.Pass))
	return &ServiceContext{
		Config:       c,
		UserRpc:      userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		Redis:        r,
		JwtBlacklist: middleware.NewJwtBlacklistMiddleware(r).Handle,
	}
}

func (s *ServiceContext) RpcContext(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "token", "internal")
}
