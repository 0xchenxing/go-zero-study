// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package handler

import (
	"context"
	"net/http"
	"strings"

	"user/api/internal/logic"
	"user/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func LogoutHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从 Authorization Header 中提取 Token
		auth := r.Header.Get("Authorization")
		tokenStr := ""
		if strings.HasPrefix(auth, "Bearer ") {
			tokenStr = strings.TrimPrefix(auth, "Bearer ")
		}

		// 将 token 放入 context，以便 logic 层获取
		ctx := context.WithValue(r.Context(), "jwtToken", tokenStr)
		l := logic.NewLogoutLogic(ctx, svcCtx)
		resp, err := l.Logout()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
