// Package eresp 提供企业级统一 API 响应结构（Google API 风格）。
//
// 设计原则：
//   - code=0 表示成功
//   - data 只用于成功返回
//   - details 只用于错误附加信息
//   - reason 为稳定机器可读标识
//   - 支持 error 自动转换
package eresp

import (
	"errors"

	"github.com/qq1060656096/bizutil/errcode"
)

// Response 统一 API 响应结构。
type Response struct {
	Code    int    `json:"code"`             // 业务码（0=成功）
	Reason  string `json:"reason,omitempty"` // 稳定错误标识
	Message string `json:"message"`          // 用户提示信息

	Data    any `json:"data,omitempty"`    // 成功返回数据（success only）
	Details any `json:"details,omitempty"` // 错误附加信息（error only）

	TraceID string `json:"trace_id,omitempty"` // 链路追踪ID
}

const (
	// 成功
	OkCode = 0
	// 未知错误码
	UnknownCode = -1
)

// OKResp 成功响应。
func OKResp(data any, msg string) Response {
	if msg == "" {
		msg = "OK"
	}

	return Response{
		Code:    OkCode,
		Message: msg,
		Data:    data,
	}
}

// Ok 泛型成功响应（推荐）。
func Ok[T any](data T, msg string) Response {
	if msg == "" {
		msg = "OK"
	}
	return Response{
		Code:    OkCode,
		Message: msg,
		Data:    data,
	}
}

// ErrorResp 错误响应。
func ErrorResp(code int, reason, msg string, details any) Response {
	if msg == "" {
		msg = "error"
	}

	return Response{
		Code:    code,
		Reason:  reason,
		Message: msg,
		Details: details,
	}
}

// WithTrace 设置 traceID。
func (r Response) WithTrace(traceID string) Response {
	r.TraceID = traceID
	return r
}

// FromError 从 error 类型转换为统一响应结构。
// 支持三种错误类型的自动转换：
//  1. errcode.Error - 业务错误，提取 CodeInt()、Reason()、Message()
//  2. httpError - HTTP错误，提取 StatusCode() 作为业务码
//  3. 其他错误 - 作为未知内部错误处理
func FromError(err error, details any) Response {
	if err == nil {
		return OKResp(nil, "")
	}

	// 1️⃣ errcode.Error（业务错误）
	var e *errcode.Error
	if errors.As(err, &e) {
		return ErrorResp(e.CodeInt(), e.Reason(), e.Message(), details)
	}

	// 2️⃣ HTTP error（可选）
	type httpError interface {
		StatusCode() int
	}

	if he, ok := err.(httpError); ok {
		return ErrorResp(he.StatusCode(), "HTTP_ERROR", err.Error(), details)
	}

	// 3️⃣ fallback：未知错误
	return ErrorResp(UnknownCode, "INTERNAL_ERROR", "internal server error", details)
}
