// Package dberr 把 PostgreSQL / GORM 的底层错误归类为可判断的语义错误，
// 便于上层把数据库异常转换成业务错误，而不是把底层 SQL 细节暴露给调用方。
package dberr

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// PostgreSQL SQLSTATE 错误码。
const (
	uniqueViolation     = "23505"
	foreignKeyViolation = "23503"
)

// ConstraintName 返回触发冲突的约束（或唯一索引）名称，取不到时返回空串。
// 依赖 gorm 的 TranslateError：驱动会把原始 *pgconn.PgError 一并包进错误链。
func ConstraintName(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.ConstraintName
	}
	return ""
}

// IsDuplicateKey 判断错误是否为唯一约束冲突（SQLSTATE 23505）。
func IsDuplicateKey(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolation
}

// IsForeignKeyViolation 判断错误是否为外键约束冲突（SQLSTATE 23503）。
func IsForeignKeyViolation(err error) bool {
	if errors.Is(err, gorm.ErrForeignKeyViolated) {
		return true
	}
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == foreignKeyViolation
}
