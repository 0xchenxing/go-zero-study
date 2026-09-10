package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type User struct {
	ID        int64  `db:"id"`
	Username  string `db:"username"`
	Password  string `db:"password"`
	Role      string `db:"role"`
	CreatedAt string `db:"created_at"`
}

type UserModel interface {
	FindOne(ctx context.Context, id int64) (*User, error)
	FindOneByUsername(ctx context.Context, username string) (*User, error)
}

type userModel struct {
	conn sqlx.SqlConn
}

func NewUserModel(conn sqlx.SqlConn) UserModel {
	return &userModel{conn: conn}
}

func (m *userModel) FindOne(ctx context.Context, id int64) (*User, error) {
	var user User
	query := "SELECT id, username, password, role, created_at FROM users WHERE id = ?"
	err := m.conn.QueryRowCtx(ctx, &user, query, id)
	if err != nil {
		if err == sqlc.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (m *userModel) FindOneByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	query := "SELECT id, username, password, role, created_at FROM users WHERE username = ?"
	err := m.conn.QueryRowCtx(ctx, &user, query, username)
	if err != nil {
		if err == sqlc.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
