package logic

import (
	"context"

	"user/rpc/internal/dal/model"
	"user/rpc/internal/svc"
	"user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DeleteUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserLogic {
	return &DeleteUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteUserLogic) DeleteUser(in *user.DeleteUserReq) (*user.DeleteUserResp, error) {
	if in.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "用户 ID 无效")
	}
	result, err := l.svcCtx.PostgresQuery.User.WithContext(l.ctx).Delete(&model.User{ID: in.GetId()})
	if err != nil {
		return nil, status.Error(codes.Internal, "删除用户失败")
	}
	if result.RowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "用户不存在")
	}
	return &user.DeleteUserResp{Success: true}, nil
}
