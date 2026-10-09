// Package response 定义全项目统一的 HTTP 响应结构。
//
// 成功与失败使用同一个信封，客户端只需判断 code：
// code 为 0 表示成功，data 携带业务数据；非 0 表示失败，msg 可直接展示。
package response

import (
	"context"

	"user/api/common/errcode"

	"go.opentelemetry.io/otel/trace"
)

// Response 统一响应结构体。
type Response struct {
	Code    int    `json:"code"`               // 业务状态码，0 表示成功
	Message string `json:"msg"`                // 状态描述，可直接展示给用户
	Data    any    `json:"data,omitempty"`     // 成功时的业务数据
	TraceID string `json:"trace_id,omitempty"` // 链路 ID，便于跨服务排查
}

// Success 构造成功响应。
func Success(ctx context.Context, data any) *Response {
	return &Response{
		Code:    errcode.ErrSuccess.Code,
		Message: errcode.ErrSuccess.Message,
		Data:    data,
		TraceID: traceID(ctx),
	}
}

// Fail 构造失败响应。
func Fail(ctx context.Context, err *errcode.Error) *Response {
	return &Response{
		Code:    err.Code,
		Message: err.Message,
		TraceID: traceID(ctx),
	}
}

// traceID 从上下文提取链路 ID，取不到时返回空串，由 omitempty 忽略该字段。
func traceID(ctx context.Context) string {
	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	if !spanCtx.HasTraceID() {
		return ""
	}
	return spanCtx.TraceID().String()
}
