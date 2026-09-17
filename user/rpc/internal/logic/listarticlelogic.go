package logic

import (
	"context"

	"user/rpc/internal/svc"
	"user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListArticleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListArticleLogic {
	return &ListArticleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListArticleLogic) ListArticle(in *user.ListArticleReq) (*user.ListArticleResp, error) {
	list, err := l.svcCtx.ArticleModel.FindList(l.ctx)
	if err != nil {
		return nil, err
	}
	articles := make([]*user.Article, 0, len(list))
	for _, a := range list {
		articles = append(articles, &user.Article{
			Id:        a.ID.Hex(),
			Title:     a.Title,
			Content:   a.Content,
			Published: a.Published,
		})
	}
	return &user.ListArticleResp{Articles: articles}, nil
}
