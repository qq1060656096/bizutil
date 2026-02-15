package eresp

import (
	"errors"
	"reflect"
	"testing"

	"github.com/qq1060656096/bizutil/errcode"
)

// TestOKResp 测试成功响应
func TestOKResp(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		msg      string
		expected Response
	}{
		{
			name: "带数据和消息",
			data: map[string]string{"key": "value"},
			msg:  "success",
			expected: Response{
				Code:    SuccessCode,
				Message: "success",
				Data:    map[string]string{"key": "value"},
			},
		},
		{
			name:     "带数据无消息",
			data:     "test data",
			msg:      "",
			expected: Response{Code: SuccessCode, Message: "OK", Data: "test data"},
		},
		{
			name:     "无数据有消息",
			data:     nil,
			msg:      "created",
			expected: Response{Code: SuccessCode, Message: "created", Data: nil},
		},
		{
			name:     "无数据无消息",
			data:     nil,
			msg:      "",
			expected: Response{Code: SuccessCode, Message: "OK", Data: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := OKResp(tt.data, tt.msg)
			if result.Code != tt.expected.Code {
				t.Errorf("Code = %v, want %v", result.Code, tt.expected.Code)
			}
			if result.Message != tt.expected.Message {
				t.Errorf("Message = %v, want %v", result.Message, tt.expected.Message)
			}
			// 使用reflect.DeepEqual比较复杂数据类型
			if !reflect.DeepEqual(result.Data, tt.expected.Data) {
				t.Errorf("Data = %v, want %v", result.Data, tt.expected.Data)
			}
			// 确保错误字段为空
			if result.Reason != "" {
				t.Errorf("Reason = %v, want empty", result.Reason)
			}
			if result.Details != nil {
				t.Errorf("Details = %v, want nil", result.Details)
			}
		})
	}
}

// TestOk 测试泛型成功响应
func TestOk(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		msg      string
		expected Response
	}{
		{
			name: "字符串数据",
			data: "hello world",
			msg:  "ok",
			expected: Response{
				Code:    SuccessCode,
				Message: "ok",
				Data:    "hello world",
			},
		},
		{
			name: "整数数据",
			data: 42,
			msg:  "",
			expected: Response{
				Code:    SuccessCode,
				Message: "OK",
				Data:    42,
			},
		},
		{
			name: "结构体数据",
			data: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{Name: "张三", Age: 25},
			msg: "user created",
			expected: Response{
				Code:    SuccessCode,
				Message: "user created",
				Data:    struct {
					Name string `json:"name"`
					Age  int    `json:"age"`
				}{Name: "张三", Age: 25},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Ok(tt.data, tt.msg)
			if result.Code != tt.expected.Code {
				t.Errorf("Code = %v, want %v", result.Code, tt.expected.Code)
			}
			if result.Message != tt.expected.Message {
				t.Errorf("Message = %v, want %v", result.Message, tt.expected.Message)
			}
			if !reflect.DeepEqual(result.Data, tt.expected.Data) {
				t.Errorf("Data = %v, want %v", result.Data, tt.expected.Data)
			}
		})
	}
}

// TestErrorResp 测试错误响应
func TestErrorResp(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		reason   string
		msg      string
		details  any
		expected Response
	}{
		{
			name:    "完整错误信息",
			code:    1001,
			reason:  "USER_NOT_FOUND",
			msg:     "用户不存在",
			details: map[string]string{"user_id": "123"},
			expected: Response{
				Code:    1001,
				Reason:  "USER_NOT_FOUND",
				Message: "用户不存在",
				Details: map[string]string{"user_id": "123"},
			},
		},
		{
			name:     "无消息自动填充",
			code:     400,
			reason:   "BAD_REQUEST",
			msg:      "",
			details:  nil,
			expected: Response{Code: 400, Reason: "BAD_REQUEST", Message: "error", Details: nil},
		},
		{
			name:     "无reason",
			code:     500,
			reason:   "",
			msg:      "服务器内部错误",
			details:  "stack trace",
			expected: Response{Code: 500, Reason: "", Message: "服务器内部错误", Details: "stack trace"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ErrorResp(tt.code, tt.reason, tt.msg, tt.details)
			if result.Code != tt.expected.Code {
				t.Errorf("Code = %v, want %v", result.Code, tt.expected.Code)
			}
			if result.Reason != tt.expected.Reason {
				t.Errorf("Reason = %v, want %v", result.Reason, tt.expected.Reason)
			}
			if result.Message != tt.expected.Message {
				t.Errorf("Message = %v, want %v", result.Message, tt.expected.Message)
			}
			if !reflect.DeepEqual(result.Details, tt.expected.Details) {
				t.Errorf("Details = %v, want %v", result.Details, tt.expected.Details)
			}
			// 确保成功字段为空
			if result.Data != nil {
				t.Errorf("Data = %v, want nil", result.Data)
			}
		})
	}
}

// TestWithTrace 测试设置TraceID
func TestWithTrace(t *testing.T) {
	resp := OKResp("test", "success")
	traceID := "trace-123-456"

	result := resp.WithTrace(traceID)

	// 验证返回的新响应包含traceID
	if result.TraceID != traceID {
		t.Errorf("TraceID = %v, want %v", result.TraceID, traceID)
	}

	// 验证原响应未被修改
	if resp.TraceID != "" {
		t.Errorf("Original response TraceID should be empty, got %v", resp.TraceID)
	}

	// 验证其他字段未被改变
	if result.Code != resp.Code || result.Message != resp.Message || !reflect.DeepEqual(result.Data, resp.Data) {
		t.Error("Other fields should remain unchanged")
	}
}

// TestFromError 测试从error转换
func TestFromError(t *testing.T) {
	// 创建测试用的 errcode.Error
	bizErr := errcode.NewWithReason(1014040001, "USER_NOT_FOUND", "用户不存在").(*errcode.Error)

	// 创建测试用的 HTTP error
	httpErr := &testHTTPError{status: 404, msg: "Not Found"}

	tests := []struct {
		name     string
		err      error
		details  any
		expected Response
	}{
		{
			name:    "nil error",
			err:     nil,
			details: nil,
			expected: Response{
				Code:    SuccessCode,
				Message: "OK",
				Data:    nil,
			},
		},
		{
			name:    "errcode.Error",
			err:     bizErr,
			details: map[string]string{"user_id": "123"},
			expected: Response{
				Code:    1014040001,
				Reason:  "USER_NOT_FOUND",
				Message: "用户不存在",
				Details: map[string]string{"user_id": "123"},
			},
		},
		{
			name:     "HTTP error",
			err:      httpErr,
			details:  "request details",
			expected: Response{Code: 404, Reason: "HTTP_ERROR", Message: "Not Found", Details: "request details"},
		},
		{
			name:     "普通错误",
			err:      errors.New("something went wrong"),
			details:  nil,
			expected: Response{Code: UnknownCode, Reason: "INTERNAL_ERROR", Message: "internal server error", Details: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromError(tt.err, tt.details)
			if result.Code != tt.expected.Code {
				t.Errorf("Code = %v, want %v", result.Code, tt.expected.Code)
			}
			if result.Reason != tt.expected.Reason {
				t.Errorf("Reason = %v, want %v", result.Reason, tt.expected.Reason)
			}
			if result.Message != tt.expected.Message {
				t.Errorf("Message = %v, want %v", result.Message, tt.expected.Message)
			}
			if !reflect.DeepEqual(result.Details, tt.expected.Details) {
				t.Errorf("Details = %v, want %v", result.Details, tt.expected.Details)
			}
		})
	}
}

// testHTTPError 实现 httpError 接口用于测试
type testHTTPError struct {
	status int
	msg    string
}

func (e *testHTTPError) Error() string {
	return e.msg
}

func (e *testHTTPError) StatusCode() int {
	return e.status
}

// TestResponseJSON 测试响应的JSON序列化
func TestResponseJSON(t *testing.T) {
	t.Run("成功响应JSON", func(t *testing.T) {
		resp := Ok(map[string]string{"name": "张三"}, "创建成功")
		if resp.Code != SuccessCode {
			t.Errorf("Expected success code, got %d", resp.Code)
		}
	})

	t.Run("错误响应JSON", func(t *testing.T) {
		resp := ErrorResp(400, "BAD_REQUEST", "请求参数错误", map[string]string{"field": "name"})
		if resp.Code != 400 {
			t.Errorf("Expected error code 400, got %d", resp.Code)
		}
	})
}

// TestConstants 测试常量定义
func TestConstants(t *testing.T) {
	if SuccessCode != 0 {
		t.Errorf("SuccessCode = %d, want 0", SuccessCode)
	}
	if UnknownCode != -1 {
		t.Errorf("UnknownCode = %d, want -1", UnknownCode)
	}
}

// TestResponseImmutability 测试响应的不可变性
func TestResponseImmutability(t *testing.T) {
	// 测试WithTrace方法不会修改原始响应
	original := OKResp("data", "success")
	modified := original.WithTrace("trace-id")
	
	// 原始响应应该保持不变
	if original.TraceID != "" {
		t.Error("Original response should not be modified")
	}
	
	// 修改后的响应应该有trace id
	if modified.TraceID != "trace-id" {
		t.Error("Modified response should have trace id")
	}
	
	// 其他字段应该相同
	if original.Code != modified.Code || original.Message != modified.Message {
		t.Error("Other fields should be the same")
	}
}

// BenchmarkOKResp 性能测试
func BenchmarkOKResp(b *testing.B) {
	data := map[string]string{"key": "value"}
	msg := "success"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		OKResp(data, msg)
	}
}

// BenchmarkOk 泛型版本性能测试
func BenchmarkOk(b *testing.B) {
	data := "test data"
	msg := "ok"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Ok(data, msg)
	}
}

// BenchmarkErrorResp 错误响应性能测试
func BenchmarkErrorResp(b *testing.B) {
	code := 500
	reason := "INTERNAL_ERROR"
	msg := "服务器内部错误"
	details := "error details"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ErrorResp(code, reason, msg, details)
	}
}

// BenchmarkFromError 错误转换性能测试
func BenchmarkFromError(b *testing.B) {
	err := errors.New("test error")
	details := "error details"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FromError(err, details)
	}
}
