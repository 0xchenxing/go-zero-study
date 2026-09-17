package mongo

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var _ ArticleModel = (*customArticleModel)(nil)

type (
	// ArticleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customArticleModel.
	ArticleModel interface {
		articleModel
		FindList(ctx context.Context) ([]*Article, error)
	}

	customArticleModel struct {
		*defaultArticleModel
	}
)

// NewArticleModel returns a model for the mongo.
func NewArticleModel(url, db, collection string) ArticleModel {
	conn := mon.MustNewModel(url, db, collection)
	return &customArticleModel{
		defaultArticleModel: newDefaultArticleModel(conn),
	}
}

func (m *customArticleModel) FindList(ctx context.Context) ([]*Article, error) {
	var list []*Article
	err := m.conn.Find(ctx, &list, bson.M{})
	if err != nil {
		return nil, err
	}
	return list, nil
}
