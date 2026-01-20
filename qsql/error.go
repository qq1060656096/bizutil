package qsql

import (
	"fmt"
	"strings"
)

const (
	ErrValidatorRequired   = "required"
	ErrValidatorTypeStr    = "str"
	ErrValidatorTypeInt    = "int"
	ErrValidatorTypeStrLen = "strLen"
	ErrValidatorTypeArrLen = "arrLen"
	ErrValidatorTypeReg    = "reg"
)

type ValidatorError struct {
	Type      string `json:"type"`    // 校验类型：required / type / range / enum / custom
	FieldName string `json:"field"`   // 字段名（业务字段）
	Code      string `json:"code"`    // 错误码（machine readable）
	Msg       string `json:"message"` // 错误文案（human readable）
	Paths     string `json:"path"`    // JSONPath / DSL 路径
}

func (e *ValidatorError) SetPaths(paths ...string) *ValidatorError {
	e.Paths = strings.Join(paths, ".")
	return e
}

func (e *ValidatorError) Error() string {
	return fmt.Sprintf("validator error: %s, code: %s, msg: %s, paths: %s", e.Type, e.Code, e.Msg, e.Paths)
}

func NewValidatorError(typ string, fieldName, code, msg string) *ValidatorError {
	return &ValidatorError{
		Type:      typ,
		FieldName: fieldName,
		Code:      code,
		Msg:       msg,
	}
}
