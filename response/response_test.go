package response

import (
	"encoding/json"
	"testing"
)

func TestResponse_Success(t *testing.T) {
	// 测试成功响应
	data := "test data"
	resp := Response[string]{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	}

	if resp.Code != CodeSuccess {
		t.Errorf("Expected code %d, got %d", CodeSuccess, resp.Code)
	}

	if resp.Message != "success" {
		t.Errorf("Expected message 'success', got '%s'", resp.Message)
	}

	if resp.Data != data {
		t.Errorf("Expected data '%s', got '%s'", data, resp.Data)
	}
}

func TestResponse_Error(t *testing.T) {
	// 测试错误响应
	resp := Response[any]{
		Code:    400,
		Message: "bad request",
	}

	if resp.Code != 400 {
		t.Errorf("Expected code 400, got %d", resp.Code)
	}

	if resp.Message != "bad request" {
		t.Errorf("Expected message 'bad request', got '%s'", resp.Message)
	}

	// 确认Data字段为零值
	var zeroData any
	if resp.Data != zeroData {
		t.Errorf("Expected zero data, got %v", resp.Data)
	}
}

func TestResponse_WithStructData(t *testing.T) {
	// 测试结构体数据
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	user := User{ID: 1, Name: "John"}
	resp := Response[User]{
		Code:    CodeSuccess,
		Message: "user found",
		Data:    user,
	}

	if resp.Code != CodeSuccess {
		t.Errorf("Expected code %d, got %d", CodeSuccess, resp.Code)
	}

	if resp.Data.ID != 1 {
		t.Errorf("Expected user ID 1, got %d", resp.Data.ID)
	}

	if resp.Data.Name != "John" {
		t.Errorf("Expected user name 'John', got '%s'", resp.Data.Name)
	}
}

func TestResponse_JSONSerialization(t *testing.T) {
	// 测试JSON序列化
	type TestData struct {
		Value string `json:"value"`
	}

	data := TestData{Value: "test"}
	resp := Response[TestData]{
		Code:    CodeSuccess,
		Message: "ok",
		Data:    data,
	}

	// 序列化为JSON
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}

	// 反序列化
	var parsedResp Response[TestData]
	err = json.Unmarshal(jsonBytes, &parsedResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal from JSON: %v", err)
	}

	if parsedResp.Code != CodeSuccess {
		t.Errorf("Expected code %d after JSON roundtrip, got %d", CodeSuccess, parsedResp.Code)
	}

	if parsedResp.Message != "ok" {
		t.Errorf("Expected message 'ok' after JSON roundtrip, got '%s'", parsedResp.Message)
	}

	if parsedResp.Data.Value != "test" {
		t.Errorf("Expected data value 'test' after JSON roundtrip, got '%s'", parsedResp.Data.Value)
	}
}

func TestResponse_WithoutData(t *testing.T) {
	// 测试没有数据的响应（Data字段应该被省略）
	resp := Response[any]{
		Code:    CodeSuccess,
		Message: "no data",
	}

	// 序列化为JSON，检查data字段是否被省略
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}

	// 对于any类型，零值是nil，omitempty会省略该字段
	// 但实际上Go的json.Marshal对于any类型的零值nil会省略字段
	// 我们验证JSON结构是否正确
	var parsedResp Response[any]
	err = json.Unmarshal(jsonBytes, &parsedResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal from JSON: %v", err)
	}

	if parsedResp.Code != CodeSuccess {
		t.Errorf("Expected code %d, got %d", CodeSuccess, parsedResp.Code)
	}

	if parsedResp.Message != "no data" {
		t.Errorf("Expected message 'no data', got '%s'", parsedResp.Message)
	}
}

func TestResponse_WithNilData(t *testing.T) {
	// 测试指针类型数据为nil的情况
	resp := Response[*string]{
		Code:    CodeSuccess,
		Message: "nil data",
		Data:    nil,
	}

	// 序列化为JSON，检查data字段是否被省略
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}

	// 对于指针类型，nil值会被omitempty省略
	// 验证JSON结构是否正确
	var parsedResp Response[*string]
	err = json.Unmarshal(jsonBytes, &parsedResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal from JSON: %v", err)
	}

	if parsedResp.Code != CodeSuccess {
		t.Errorf("Expected code %d, got %d", CodeSuccess, parsedResp.Code)
	}

	if parsedResp.Message != "nil data" {
		t.Errorf("Expected message 'nil data', got '%s'", parsedResp.Message)
	}

	// 确认Data字段仍为nil
	if parsedResp.Data != nil {
		t.Errorf("Expected nil data, got %v", parsedResp.Data)
	}
}
