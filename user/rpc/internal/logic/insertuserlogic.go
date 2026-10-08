package logic

import (
	"context"

	"user/rpc/internal/dal/model"
	"user/rpc/internal/svc"
	"user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InsertUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInsertUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InsertUserLogic {
	return &InsertUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *InsertUserLogic) InsertUser(in *user.InsertUserReq) (*user.InsertUserResp, error) {
	if in.GetUsername() == "" || in.GetEmail() == "" || in.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "用户名、邮箱和密码不能为空")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "密码处理失败")
	}
	role := in.GetRole()
	if role == "" {
		role = "user"
	}
	item := &model.User{Username: in.GetUsername(), Email: in.GetEmail(), Password: string(hash), Role: role}
	if err := l.svcCtx.PostgresQuery.User.WithContext(l.ctx).Create(item); err != nil {
		l.Errorf("insert user failed: username=%s email=%s err=%v", in.GetUsername(), in.GetEmail(), err)
		return nil, mapCreateUserError(err)
	}
	return &user.InsertUserResp{Id: item.ID}, nil
}
