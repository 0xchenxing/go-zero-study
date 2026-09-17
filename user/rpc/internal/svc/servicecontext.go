package svc

import (
	"user/rpc/internal/config"
	"user/rpc/model/mongo"
	"user/rpc/model/mysql"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config       config.Config
	UserModel    mysql.UserModel
	ArticleModel mongo.ArticleModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DB.DataSource)
	return &ServiceContext{
		Config:       c,
		UserModel:    mysql.NewUserModel(conn, c.CacheRedis),
		ArticleModel: mongo.NewArticleModel(c.Mongo.Uri, c.Mongo.Database, "article"),
	}
}
