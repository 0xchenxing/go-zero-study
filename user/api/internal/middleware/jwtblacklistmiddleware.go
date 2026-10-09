package middleware

import (
	"net/http"
	"strings"
	"user/api/common/errcode"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest/httpx"
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
				httpx.ErrorCtx(r.Context(), w, errcode.ErrUserTokenInvalid)
				return
			}
		}

		next(w, r)
	}
}
