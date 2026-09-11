package middleware

import (
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest/httpx"
	errorx "github.com/zeromicro/x/errors"
)

type JwtBlacklistMiddleware struct {
	Redis *redis.Redis
}

func NewJwtBlacklistMiddleware(r *redis.Redis) *JwtBlacklistMiddleware {
	return &JwtBlacklistMiddleware{
		Redis: r,
	}
}

func (m *JwtBlacklistMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		tokenStr := ""
		if strings.HasPrefix(auth, "Bearer ") {
			tokenStr = strings.TrimPrefix(auth, "Bearer ")
		}

		if tokenStr != "" {
			exists, err := m.Redis.Exists("jwt:revoked:" + tokenStr)
			if err == nil && exists {
				httpx.ErrorCtx(r.Context(), w, errorx.New(http.StatusUnauthorized, "Token 已失效"))
				return
			}
		}

		next(w, r)
	}
}
