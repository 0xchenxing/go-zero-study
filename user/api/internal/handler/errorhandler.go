package handler

import (
	"context"
	"net/http"

	"user/api/common/errcode"
	"user/api/common/response"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// RegisterErrorHandler 注册全局错误处理器，是所有接口错误响应的唯一出口。
//
// 业务逻辑只需返回 errcode 中的错误值，这里负责归一成统一响应结构：
// HTTP 状态码固定 200，业务结果只由响应体的 code 表达。
func RegisterErrorHandler() {
	httpx.SetErrorHandlerCtx(func(ctx context.Context, err error) (int, any) {
		codeErr := errcode.ToError(err)

		switch {
		case codeErr.Detail != "":
			// 业务错误附带的内部细节只进日志，不出响应。
			logx.WithContext(ctx).Errorf("error detail: code=%d, detail=%s",
				codeErr.Code, codeErr.Detail)
		case !errcode.IsError(err):
			// 未预期的底层错误，对外只会返回通用文案。
			logx.WithContext(ctx).Errorf("unhandled error: %+v", err)
		}

		return http.StatusOK, response.Fail(ctx, codeErr)
	})
}

// parse 解析请求参数，失败一律定性为 9002 参数错误。
//
// go-zero 的 httpx.Parse 返回裸错误，不带任何状态标记，若不在此处定性，
// 全局错误处理器只能把它当成未知的服务端故障（9001），客户端会误以为是自己无法修复的
// 后端问题而不会去修正请求体。路径参数解析失败同理（如 /user/abc）。
func parse(r *http.Request, v any) error {
	if err := httpx.Parse(r, v); err != nil {
		return errcode.ErrBadRequest.WithDetail(err.Error())
	}

	return nil
}

// okJson 以统一信封写出成功响应。
func okJson(ctx context.Context, w http.ResponseWriter, data any) {
	httpx.OkJsonCtx(ctx, w, response.Success(ctx, data))
}
