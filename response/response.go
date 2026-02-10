package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeSuccess     = 0  // 成功
	CodeSystemError = -1 // 系统错误
)

// Response 通用 JSON 响应结构
type Response[T any] struct {
	Code    int    `json:"code"`    // 0 表示成功，非0为业务错误
	Message string `json:"message"` // 提示信息
	Data    T      `json:"data"`    // 数据
}

// -------------------- 成功响应 --------------------

// Success 返回成功响应 200
func Success[T any](c *gin.Context, data T, msg ...string) {
	message := "success"
	if len(msg) > 0 {
		message = msg[0]
	}
	c.JSON(http.StatusOK, Response[T]{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// Created 返回创建成功响应 201
func Created[T any](c *gin.Context, data T, msg ...string) {
	message := "created"
	if len(msg) > 0 {
		message = msg[0]
	}
	c.JSON(http.StatusCreated, Response[T]{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// NoContent 返回 204 无内容
func NoContent(c *gin.Context, msg ...string) {
	c.Status(http.StatusNoContent)
}

// -------------------- 失败响应 --------------------

// Error 根据 error 返回 JSON
// code != 0 表示业务错误，否则系统错误
func Error[T any](c *gin.Context, err error, code int, data T) {
	if err == nil {
		Success(c, data)
		return
	}

	msg := err.Error()
	if code != 0 {
		// 业务错误，HTTP 200 返回
		c.JSON(http.StatusOK, Response[T]{
			Code:    code,
			Message: msg,
			Data:    data,
		})
	} else {
		// 系统错误，HTTP 500
		c.JSON(http.StatusInternalServerError, Response[T]{
			Code:    -1,
			Message: msg,
			Data:    data,
		})
	}
}

// ErrorNoData error 不带数据
func ErrorNoData(c *gin.Context, err error, code int) {
	Error[any](c, err, code, nil)
}

// -------------------- HTTP 状态码快捷方法 --------------------

// NotFound 404 资源未找到
func NotFound(c *gin.Context, msg ...string) {
	message := "resource not found"
	if len(msg) > 0 {
		message = msg[0]
	}
	c.JSON(http.StatusNotFound, Response[any]{
		Code:    404,
		Message: message,
		Data:    nil,
	})
}

// Forbidden 403 权限不足
func Forbidden(c *gin.Context, msg ...string) {
	message := "forbidden"
	if len(msg) > 0 {
		message = msg[0]
	}
	c.JSON(http.StatusForbidden, Response[any]{
		Code:    403,
		Message: message,
		Data:    nil,
	})
}

// BadRequest 400 参数错误
func BadRequest(c *gin.Context, msg ...string) {
	message := "bad request"
	if len(msg) > 0 {
		message = msg[0]
	}
	c.JSON(http.StatusBadRequest, Response[any]{
		Code:    400,
		Message: message,
		Data:    nil,
	})
}

// UnprocessableEntity 422 校验失败
func UnprocessableEntity(c *gin.Context, msg ...string) {
	message := "unprocessable entity"
	if len(msg) > 0 {
		message = msg[0]
	}
	c.JSON(http.StatusUnprocessableEntity, Response[any]{
		Code:    422,
		Message: message,
		Data:    nil,
	})
}

// SystemError 系统错误 500
func SystemError[T any](c *gin.Context, err error, data T) {
	msg := "system error"
	if err != nil {
		msg = err.Error()
	}
	c.JSON(http.StatusInternalServerError, Response[T]{
		Code:    CodeSystemError,
		Message: msg,
		Data:    data,
	})
}

// SystemErrorNoData 500 不带数据
func SystemErrorNoData(c *gin.Context, err error) {
	SystemError[any](c, err, nil)
}
