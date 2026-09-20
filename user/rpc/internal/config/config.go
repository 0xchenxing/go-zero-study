package config

import (
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DB         sqlx.SqlConf
	CacheRedis cache.CacheConf
	Mongo      struct {
		Uri      string
		Database string
	}
	UserRedis    redis.RedisConf
	KqPusherConf struct {
		Brokers []string
		Topic   string
	}
	KqConsumerConf kq.KqConf
}
