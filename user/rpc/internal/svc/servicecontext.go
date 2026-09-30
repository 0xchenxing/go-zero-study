package svc

import (
	"user/rpc/internal/config"
	"user/rpc/internal/dal/query"
	"user/rpc/internal/db"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config         config.Config
	Redis          *redis.Redis
	KqPusherClient *kq.Pusher
	PostgresDB     *gorm.DB
	PostgresQuery  *query.Query
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	postgresDB, err := db.Init(c.PostgresDSN)
	if err != nil {
		return nil, err
	}

	return &ServiceContext{
		Config:         c,
		Redis:          redis.MustNewRedis(c.UserRedis),
		KqPusherClient: kq.NewPusher(c.KqPusherConf.Brokers, c.KqPusherConf.Topic),
		PostgresDB:     postgresDB,
		PostgresQuery:  query.Use(postgresDB),
	}, nil
}
