package logic

import (
	"context"

	"user/rpc/internal/svc"
	"user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UpdateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateUserLogic) UpdateUser(in *user.UpdateUserReq) (*user.UpdateUserResp, error) {
	if in.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "用户 ID 无效")
	}
	updates := map[string]any{}
	if in.GetUsername() != "" {
		updates["username"] = in.GetUsername()
	}
	if in.GetEmail() != "" {
		updates["email"] = in.GetEmail()
	}
	if in.GetRole() != "" {
		updates["role"] = in.GetRole()
	}
	if len(updates) == 0 {
		return nil, status.Error(codes.InvalidArgument, "没有可更新的字段")
	}
	result, err := l.svcCtx.PostgresQuery.User.WithContext(l.ctx).
		Where(l.svcCtx.PostgresQuery.User.ID.Eq(in.GetId())).Updates(updates)
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, "用户名或邮箱已存在")
	}
	if result.RowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "用户不存在")
	}
	return &user.UpdateUserResp{Success: true}, nil
}
