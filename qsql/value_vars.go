package qsql

import "github.com/tidwall/sjson"

// ValueVars 是基于 Go 值的 Vars 实现。
// 内部使用 sjson 直接构建 JSON 字符串，
// 避免 map + json.Marshal 带来的中间结构和开销。
type ValueVars struct {
	raw string
}

// NewValueVars 创建一个空的 ValueVars。
// 返回值的零值同样可用。
func NewValueVars() *ValueVars {
	return &ValueVars{}
}

// Set 设置指定名称的变量，value 为 Go 值。
// value 会被自动序列化并写入 JSON。
func (v *ValueVars) Set(name string, value any) {
	v.raw, _ = sjson.Set(v.raw, name, value)
}

// Sys 设置系统级变量（$.sys）。
func (v *ValueVars) Sys(value any) {
	v.Set("sys", value)
}

// Users 设置用户级变量（$.users）。
func (v *ValueVars) Users(value any) {
	v.Set("users", value)
}

// Params 设置前端参数变量（$.params）。
func (v *ValueVars) Params(value any) {
	v.Set("params", value)
}

// JSON 返回变量的 JSON 字符串。
// 如果未设置任何变量，返回空对象 {}。
func (v ValueVars) JSON() string {
	if v.raw == "" {
		return "{}"
	}
	return v.raw
}
