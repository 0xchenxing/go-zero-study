package mongo

import "github.com/zeromicro/go-zero/core/stores/mon"

var _ ArticleModel = (*customArticleModel)(nil)

type (
	// ArticleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customArticleModel.
	ArticleModel interface {
		articleModel
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
