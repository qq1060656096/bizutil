// Package errcode 提供结构化业务错误码实现（Google API 风格）。
//
// 错误码格式：
//
//	1位占位符 + 2位模块 + 3位HTTP状态码 + 4位顺序
//
//	示例：
//	1 01 404 0001
//	│ │  │   └── 顺序号
//	│ │  └────── HTTP状态码
//	│ └──────── 模块
//	└────────── 占位符（仅保证可转 int，1-9）
//
// 设计原则：
//   - 第一位仅占位，无业务含义
//   - HTTP状态码直接从错误码解析
//   - reason 可选（用于 API 语义错误）
//   - 与 errors.Is / errors.As 完全兼容
//   - 支持 errors.Unwrap
package errcode

import (
	"errors"
	"fmt"
	"strconv"
)

const codeLen = 10

// Error 表示应用错误。
type Error struct {
	code   string // 机器定位码（必有）
	reason string // 语义错误码（可选）
	msg    string // 用户提示信息
	cause  error  // 原始错误
}

//
// ===== error interface =====
//

// Error 实现 error 接口。
func (e *Error) Error() string {
	switch {
	case e.reason != "" && e.msg != "":
		return fmt.Sprintf("[%s:%s] %s", e.reason, e.code, e.msg)

	case e.reason != "":
		return fmt.Sprintf("[%s:%s]", e.reason, e.code)

	case e.msg != "":
		return fmt.Sprintf("[%s] %s", e.code, e.msg)

	default:
		return "[" + e.code + "]"
	}
}

// Unwrap 支持 errors.Unwrap。
func (e *Error) Unwrap() error {
	return e.cause
}

// Is 支持 errors.Is（按 code 判断）。
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	return ok && e.code == t.code
}

//
// ===== getter =====
//

// Code 返回完整错误码。
func (e *Error) Code() string {
	return e.code
}

// Reason 返回语义错误码。
func (e *Error) Reason() string {
	return e.reason
}

// Message 返回错误消息。
func (e *Error) Message() string {
	return e.msg
}

// CodeInt 返回错误码的 int 形式。
func (e *Error) CodeInt() int {
	n, _ := strconv.Atoi(e.code)
	return n
}

// HTTPStatus 返回错误对应的 HTTP 状态码。
func (e *Error) HTTPStatus() int {
	s, _ := strconv.Atoi(e.code[3:6])
	return s
}

//
// ===== constructor =====
//

// New 创建错误（无 reason）。
//
// code 支持：
//   - 9位 → 自动补占位符1
//   - 10位完整码
func New(code int, message string) error {
	c, err := normalize(code)
	if err != nil {
		return err
	}

	return &Error{
		code: c,
		msg:  message,
	}
}

// NewWithReason 创建带语义码的错误（推荐用于 API）。
func NewWithReason(code int, reason, message string) error {
	c, err := normalize(code)
	if err != nil {
		return err
	}

	return &Error{
		code:   c,
		reason: reason,
		msg:    message,
	}
}

// Wrap 包装已有错误（无 reason）。
func Wrap(code int, err error, message string) error {
	if err == nil {
		return nil
	}

	c, e := normalize(code)
	if e != nil {
		return e
	}

	return &Error{
		code:  c,
		msg:   message,
		cause: err,
	}
}

// WrapWithReason 包装已有错误（带 reason）。
func WrapWithReason(code int, reason string, err error, message string) error {
	if err == nil {
		return nil
	}

	c, e := normalize(code)
	if e != nil {
		return e
	}

	return &Error{
		code:   c,
		reason: reason,
		msg:    message,
		cause:  err,
	}
}

//
// ===== internal =====
//

// normalize 规范化错误码。
func normalize(code int) (string, error) {
	if code < 0 {
		return "", errors.New("errcode: invalid code")
	}

	raw := strconv.Itoa(code)

	switch len(raw) {
	case codeLen:
		return validate(raw)

	case codeLen - 1:
		return validate("1" + raw)

	default:
		return "", fmt.Errorf("errcode: code must be 9 or 10 digits")
	}
}

// validate 校验错误码格式。
func validate(code string) (string, error) {
	if len(code) != codeLen {
		return "", errors.New("errcode: invalid code length")
	}

	for _, r := range code {
		if r < '0' || r > '9' {
			return "", errors.New("errcode: code must be digits")
		}
	}

	// 校验HTTP状态码
	status, _ := strconv.Atoi(code[3:6])
	if status < 100 || status > 599 {
		return "", errors.New("errcode: invalid http status")
	}

	return code, nil
}
