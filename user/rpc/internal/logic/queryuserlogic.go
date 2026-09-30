package logic

import (
	"context"

	"user/rpc/internal/svc"
	"user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type QueryUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryUserLogic {
	return &QueryUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryUserLogic) QueryUser(in *user.QueryUserReq) (*user.QueryUserResp, error) {
	if in.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "用户 ID 无效")
	}

	item, err := l.svcCtx.PostgresQuery.User.WithContext(l.ctx).
		Where(l.svcCtx.PostgresQuery.User.ID.Eq(in.GetId())).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "用户不存在")
		}
		return nil, status.Error(codes.Internal, "查询用户失败")
	}

	return &user.QueryUserResp{Id: item.ID, Username: item.Username, Email: item.Email, Role: item.Role}, nil
}
