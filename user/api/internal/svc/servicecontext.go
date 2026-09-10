// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package svc

import (
	"user/api/internal/config"
	"user/api/model"
	"user/rpc/userclient"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config    config.Config
	UserRpc   userclient.User
	UserModel model.UserModel
	Redis     *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DB.DataSource)
	return &ServiceContext{
		Config:    c,
		UserRpc:   userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		UserModel: model.NewUserModel(conn),
		Redis:     redis.New(c.Redis.Host, redis.WithPass(c.Redis.Pass)),
	}
}
