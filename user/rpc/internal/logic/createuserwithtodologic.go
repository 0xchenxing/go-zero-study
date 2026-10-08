package logic

import (
	"context"

	"user/rpc/internal/dal/model"
	"user/rpc/internal/dal/query"
	"user/rpc/internal/svc"
	"user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateUserWithTodoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateUserWithTodoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserWithTodoLogic {
	return &CreateUserWithTodoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateUserWithTodo 在同一个事务里创建用户与待办。
// 任一步失败都会回滚，用户不会残留，待办也不会引用到不存在的用户。
func (l *CreateUserWithTodoLogic) CreateUserWithTodo(in *user.CreateUserWithTodoReq) (*user.CreateUserWithTodoResp, error) {
	if in.GetUsername() == "" || in.GetEmail() == "" || in.GetPassword() == "" || in.GetTodoTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "用户名、邮箱、密码和待办标题不能为空")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "密码处理失败")
	}

	newUser := &model.User{
		Username: in.GetUsername(),
		Email:    in.GetEmail(),
		Password: string(hash),
		Role:     "user",
	}
	newTodo := &model.Todo{
		Title: in.GetTodoTitle(),
	}

	err = l.svcCtx.PostgresQuery.Transaction(func(tx *query.Query) error {
		if err := tx.User.WithContext(l.ctx).Create(newUser); err != nil {
			return err
		}
		// 待办必须挂在刚创建的用户上，user_id 取自事务内自增返回的主键。
		newTodo.UserID = newUser.ID
		return tx.Todo.WithContext(l.ctx).Create(newTodo)
	})
	if err != nil {
		l.Errorf("create user with todo rolled back: username=%s err=%v", in.GetUsername(), err)
		return nil, mapCreateUserError(err)
	}

	return &user.CreateUserWithTodoResp{
		UserId:    newUser.ID,
		Username:  newUser.Username,
		Email:     newUser.Email,
		TodoId:    newTodo.ID,
		TodoTitle: newTodo.Title,
	}, nil
}
