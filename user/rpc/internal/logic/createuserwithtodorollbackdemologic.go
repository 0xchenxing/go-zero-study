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

type CreateUserWithTodoRollbackDemoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateUserWithTodoRollbackDemoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserWithTodoRollbackDemoLogic {
	return &CreateUserWithTodoRollbackDemoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateUserWithTodoRollbackDemo 事务回滚演示：
// 第一步正常插入用户，第二步故意写入 user_id=0 的待办，必然违反 fk_todos_user 外键约束，
// 事务整体回滚，用户不会落库。事务失败属于本次演示的预期结果，因此返回 nil error，
// 由响应体说明"已回滚"。
func (l *CreateUserWithTodoRollbackDemoLogic) CreateUserWithTodoRollbackDemo(in *user.RollbackDemoReq) (*user.RollbackDemoResp, error) {
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

	txErr := l.svcCtx.PostgresQuery.Transaction(func(tx *query.Query) error {
		if err := tx.User.WithContext(l.ctx).Create(newUser); err != nil {
			return err
		}
		// 不存在的 user_id：外键 fk_todos_user 会拒绝写入，从而触发整个事务回滚。
		invalidTodo := &model.Todo{
			UserID: 0,
			Title:  in.GetTodoTitle(),
		}
		return tx.Todo.WithContext(l.ctx).Create(invalidTodo)
	})

	if txErr == nil {
		// 外键失效等异常情况下事务意外提交成功，说明回滚演示没有按预期失败。
		l.Errorf("rollback demo unexpectedly committed: username=%s user_id=%d", in.GetUsername(), newUser.ID)
		return nil, status.Error(codes.Internal, "回滚演示异常：事务未按预期失败")
	}

	l.Infof("rollback demo: transaction rolled back as expected, username=%s err=%v", in.GetUsername(), txErr)
	return &user.RollbackDemoResp{Message: "事务失败，已完成回滚"}, nil
}
