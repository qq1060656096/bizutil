package qsql

import "github.com/tidwall/sjson"

// JSONVars 是基于 JSON 字符串的 Vars 实现。
// 内部使用 sjson 进行 JSON 的增量构建。
type JSONVars struct {
	raw string
}

// NewJSONVars 创建一个空的 JSONVars。
func NewJSONVars() *JSONVars {
	return &JSONVars{}
}

// Set 设置指定名称的变量，value 为 Go 值。
// value 会被自动序列化为 JSON。
func (v *JSONVars) Set(name string, value string) error {
	var err error
	v.raw, err = sjson.SetRaw(v.raw, name, value)
	return err
}

// SetRaw 使用原始 JSON 字符串设置变量。
// rawJSON 必须是合法的 JSON。
func (v *JSONVars) SetRaw(name string, rawJSON string) error {
	var err error
	v.raw, err = sjson.SetRaw(v.raw, name, rawJSON)
	return err
}

// Sys 设置系统级变量（$.sys）。
func (v *JSONVars) Sys(value string) error {
	return v.Set("sys", value)
}

// Users 设置用户级变量（$.users）。
func (v *JSONVars) Users(value string) error {
	return v.Set("users", value)
}

// Params 设置前端参数变量（$.params）。
func (v *JSONVars) Params(value string) error {
	return v.Set("params", value)
}

// JSON 返回完整的变量 JSON。
// 如果未设置任何变量，返回空对象 {}。
func (v JSONVars) JSON() string {
	if v.raw == "" {
		return "{}"
	}
	return v.raw
}
