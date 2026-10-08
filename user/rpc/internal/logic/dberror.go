package logic

import (
	"user/rpc/internal/dberr"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// users 表上与被占用字段相关的唯一索引名，来自 rpc/migrations/0001_init.sql。
const (
	constraintUsersUsername = "idx_users_username"
	constraintUsersEmail    = "idx_users_email"
)

// mapCreateUserError 把写用户时产生的数据库错误映射成业务错误。
// 唯一约束冲突要区分是用户名还是邮箱，才能给出准确的业务提示；
// 其余错误统一返回内部错误，约束名、SQLSTATE 等细节只写日志。
func mapCreateUserError(err error) error {
	switch {
	case dberr.IsDuplicateKey(err):
		if dberr.ConstraintName(err) == constraintUsersEmail {
			return status.Error(codes.AlreadyExists, "邮箱已被占用")
		}
		return status.Error(codes.AlreadyExists, "用户名已被占用")
	case dberr.IsForeignKeyViolation(err):
		return status.Error(codes.FailedPrecondition, "关联数据不存在")
	default:
		return status.Error(codes.Internal, "创建失败，请稍后重试")
	}
}
