package logic

import (
	"context"
	"errors"

	"user/rpc/internal/svc"
	"user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type AuthenticateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAuthenticateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthenticateUserLogic {
	return &AuthenticateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AuthenticateUserLogic) AuthenticateUser(in *user.AuthenticateUserReq) (*user.AuthenticateUserResp, error) {
	if in.GetUsername() == "" {
		return nil, status.Error(codes.InvalidArgument, "用户名不能为空")
	}
	item, err := l.svcCtx.PostgresQuery.User.WithContext(l.ctx).
		Where(l.svcCtx.PostgresQuery.User.Username.Eq(in.GetUsername())).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "用户不存在")
		}
		l.Errorf("authenticate query failed: username=%s err=%v", in.GetUsername(), err)
		return nil, status.Error(codes.Internal, "查询用户失败")
	}
	return &user.AuthenticateUserResp{Id: item.ID, Username: item.Username, Email: item.Email, Role: item.Role, Password: item.Password}, nil
}
