// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package handler

import (
	"net/http"
	"user/api/internal/logic"
	"user/api/internal/svc"
	"user/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginReq
		if err := parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewLoginLogic(r.Context(), svcCtx)
		resp, err := l.Login(&req)
		if err != nil {
			// 交给全局错误处理器，保证与其他接口使用同一套响应结构。
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			okJson(r.Context(), w, resp)
		}
	}
}
