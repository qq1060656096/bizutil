package response

// 统一响应结构
type Response[T any] struct {
	Code    int    `json:"code"`    // 状态码，0表示成功，非0表示错误
	Message string `json:"message"` // 提示信息
	Data    T      `json:"data"`    // 返回的数据，泛型支持任意类型
}

const (
	CodeSuccess = 0 // 成功
)
