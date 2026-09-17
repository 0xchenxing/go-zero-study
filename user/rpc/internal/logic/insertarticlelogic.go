package logic

import (
	"context"
	"user/rpc/internal/svc"
	"user/rpc/model/mongo"
	"user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type InsertArticleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInsertArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InsertArticleLogic {
	return &InsertArticleLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *InsertArticleLogic) InsertArticle(in *user.InsertArticleReq) (*user.InsertArticleResp, error) {
	article := &mongo.Article{
		Title:     in.Title,
		Content:   in.Content,
		Published: in.Published,
	}
	err := l.svcCtx.ArticleModel.Insert(l.ctx, article)
	if err != nil {
		return nil, err
	}
	return &user.InsertArticleResp{Id: article.ID.Hex()}, nil
}
