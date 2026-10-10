package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"user/api/common/errcode"
	"user/api/common/response"
)

// TestUnauthorizedCallback 锁定 JWT 未认证响应的契约。
//
// go-zero 内置的 JWT 中间件原本把 401 裸写在 ResponseWriter 上、body 为空，
// UnauthorizedCallback 接管后必须是 HTTP 200 + 统一信封（code=9003）。
// 这里同时覆盖 err == nil 的分支：go-zero 在无可披露原因时会传 nil，
// 若不特判会在拼 Detail 时空指针 panic。
func TestUnauthorizedCallback(t *testing.T) {
	RegisterErrorHandler()

	cases := []struct {
		name string
		err  error
	}{
		{name: "附带原因", err: errors.New("invalid auth token")},
		{name: "原因为空", err: nil},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/user/profile/1", nil)

			UnauthorizedCallback(rec, req, c.err)

			if rec.Code != http.StatusOK {
				t.Fatalf("HTTP 状态码 = %d, 期望 200（错误一律用业务码表达）", rec.Code)
			}

			var got response.Response
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("响应不是合法 JSON: %v, body=%q", err, rec.Body.String())
			}
			if got.Code != errcode.ErrUnauthorized.Code {
				t.Errorf("code = %d, 期望 %d", got.Code, errcode.ErrUnauthorized.Code)
			}
			if got.Message != errcode.ErrUnauthorized.Message {
				t.Errorf("msg = %q, 期望 %q", got.Message, errcode.ErrUnauthorized.Message)
			}
		})
	}
}
