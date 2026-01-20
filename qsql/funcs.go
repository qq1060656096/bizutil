package qsql

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/tidwall/gjson"
)

// execState 执行状态（模板执行时使用）
type execState struct {
	data             gjson.Result      // 预解析的 JSON 数据，包含 sys/users/params
	args             []interface{}     // 收集的 SQL 参数
	errors           []string          // 错误列表（记录缺失的参数等）
	validatorsErrors []*ValidatorError // 验证器错误列表
}

func (s *SQLStmt) addArgs(args ...interface{}) *SQLStmt {
	s.Args = append(s.Args, args...)
	return s
}
func (state *execState) addError(err string) {
	state.errors = append(state.errors, err)
}

func (state *execState) addValidatorError(err *ValidatorError) {
	state.validatorsErrors = append(state.validatorsErrors, err)
}

// getValueByPath 根据路径从执行状态中获取值
// 将多个路径片段用 "." 连接，然后从 JSON 数据中查找对应的值
// 返回值和是否存在的标志
func getValueByPath(state *execState, paths ...string) (interface{}, bool) {
	path := strings.Join(paths, ".")
	result := state.data.Get(path)
	if !result.Exists() {
		return nil, false
	}

	return result.Value(), true
}

// getValueByPathForTemplate 用于模板的 getValue 函数，仅返回值（nil 如果不存在）
func getValueByPathForTemplate(state *execState, paths ...string) interface{} {
	val, _ := getValueByPath(state, paths...)
	return val
}

// valFunc 值函数，用于模板中的 {{val "path"}} 语法
// 根据路径获取值并添加到 SQL 参数列表中，返回占位符 "?"
func valFunc(state *execState, paths ...string) (string, error) {
	val, _ := getValueByPath(state, paths...)
	state.args = append(state.args, val)
	return "?", nil
}

// exprFunc 必需表达式函数，用于模板中的 {{expr "field" "op" "path"}} 语法
// 构建 SQL 条件表达式，如果值不存在会记录错误
func exprFunc(state *execState, paths ...string) string {
	return exprRaw(state, true, paths...)
}

// optionalExprFunc 可选表达式函数，用于模板中的 {{optionalExpr "field" "op" "path"}} 语法
// 构建 SQL 条件表达式，如果值不存在则返回空字符串（不记录错误）
func optionalExprFunc(state *execState, paths ...string) string {
	return exprRaw(state, false, paths...)
}

// exprRaw 原始表达式构建函数
// 解析路径参数：paths[0] 为字段名，paths[1] 为操作符，paths[2:] 为值的路径
// required 参数决定当值不存在时是否记录错误
// 示例：exprRaw(state, true, "age", ">=", "params", "minAge") 生成 "age >= ?"
func exprRaw(state *execState, required bool, paths ...string) string {
	var field, op string
	if len(paths) > 0 {
		field = paths[0]
	}
	if len(paths) > 1 {
		op = paths[1]
	}
	l := len(paths)
	if l < 1 {
		return ""
	}
	if len(paths) < 3 {
		state.errors = append(state.errors, "expr: no values")
		return buildExpr(state, field, op, required, nil)
	}

	field, op = paths[0], paths[1]
	realPaths := paths[2:]
	val, ok := getValueByPath(state, realPaths...)
	if !ok && required {
		state.errors = append(state.errors, "expr: no values")
	}
	return buildExpr(state, field, op, required, val)
}

// buildExpr 构建 SQL 表达式
// 将字段名、操作符和值组合成 SQL 条件表达式，并将值添加到参数列表中
// 支持单值和数组值，生成对应的占位符（如 field IN (?, ?, ?)）
func buildExpr(state *execState, field string, op string, required bool, val interface{}) string {
	var values []interface{}

	switch v := val.(type) {
	case []interface{}:
		values = v
	case []string:
		for _, s := range v {
			values = append(values, s)
		}
	case []int:
		for _, i := range v {
			values = append(values, i)
		}
	case []int64:
		for _, i := range v {
			values = append(values, i)
		}
	default:
		// 单个值也转为 IN (?)
		values = []interface{}{v}
	}

	if len(values) == 0 {
		if required {
			values = []interface{}{nil}
			return buildSqlPlaceholder(state, field, op, values)
		}

		return ""
	}

	return buildSqlPlaceholder(state, field, op, values)
}

// buildSqlPlaceholder 根据操作符类型构建 SQL 占位符表达式
// 支持 IN/NOT IN、BETWEEN/NOT BETWEEN 以及普通比较操作符
func buildSqlPlaceholder(state *execState, field string, op string, values []interface{}) string {
	upperOp := strings.ToUpper(strings.TrimSpace(op))
	switch upperOp {
	case "IN", "NOT IN":
		placeholders := make([]string, len(values))
		for i, v := range values {
			state.args = append(state.args, v)
			placeholders[i] = "?"
		}
		return fmt.Sprintf("%s %s (%s)", field, op, strings.Join(placeholders, ", "))
	case "BETWEEN", "NOT BETWEEN":
		if len(values) < 2 {
			state.errors = append(state.errors, "between: not enough values")
			return ""
		}
		state.args = append(state.args, values[0], values[1])
		return fmt.Sprintf("%s %s ? AND ?", field, op)
	default:
		state.args = append(state.args, values[0])
		return fmt.Sprintf("%s %s ?", field, op)
	}
}

// andFunc AND 逻辑连接函数，用于模板中的 {{and "cond1" "cond2"}} 语法
// 将多个条件用 AND 连接，并用括号包裹
// 示例：andFunc(state, "", "a = 1", "b = 2") 生成 "(a = 1 and b = 2)"
func andFunc(state *execState, funcName string, conditions ...string) string {
	return andOrFunc(state, "and", conditions...)
}

// orFunc OR 逻辑连接函数，用于模板中的 {{or "cond1" "cond2"}} 语法
// 将多个条件用 OR 连接，并用括号包裹
// 示例：orFunc(state, "a = 1", "b = 2") 生成 "(a = 1 or b = 2)"
func orFunc(state *execState, conditions ...string) string {
	return andOrFunc(state, "or", conditions...)
}

// andOrFunc AND/OR 逻辑连接的通用函数
// 过滤空条件，将有效条件用指定的逻辑操作符连接
// 如果没有有效条件，记录错误并返回空字符串
func andOrFunc(state *execState, logic string, conditions ...string) string {
	var valid []string

	for _, cond := range conditions {
		cond = strings.TrimSpace(cond)
		if cond != "" {
			valid = append(valid, cond)
		}
	}

	if len(valid) == 0 {
		state.errors = append(state.errors, "and: no valid conditions")
		return ""
	}

	return "(" + strings.Join(valid, " "+logic+" ") + ")"
}

func validatorIntFunc(state *execState, fieldName string, code string, msg string, paths ...string) string {
	val, ok := getValueByPath(state, paths...)
	if !ok {
		return ""
	}
	if _, ok := val.(int64); !ok {
		err := NewValidatorError(ErrValidatorTypeInt, fieldName, code, msg)
		err.SetPaths(paths...)
		state.addValidatorError(err)
		return ""
	}
	return ""
}

func validatorFloatFunc(state *execState, fieldName string, code string, msg string, paths ...string) string {
	val, ok := getValueByPath(state, paths...)
	if !ok {
		return ""
	}
	if _, ok := val.(float64); !ok {
		err := NewValidatorError(ErrValidatorTypeInt, fieldName, code, msg)
		err.SetPaths(paths...)
		state.addValidatorError(err)
		return ""
	}
	return ""
}

func validatorRequiredFunc(state *execState, fieldName string, code string, msg string, paths ...string) string {
	_, ok := getValueByPath(state, paths...)
	if !ok {
		err := NewValidatorError(ErrValidatorRequired, fieldName, code, msg)
		err.SetPaths(paths...)
		state.addValidatorError(err)
		return ""
	}
	return ""
}

func validatorStrFunc(state *execState, fieldName string, code string, msg string, paths ...string) string {
	val, ok := getValueByPath(state, paths...)
	if !ok {
		return ""
	}
	if _, ok := val.(string); !ok {
		err := NewValidatorError(ErrValidatorTypeStr, fieldName, code, msg)
		err.SetPaths(paths...)
		state.addValidatorError(err)
		return ""
	}
	return ""
}

func validatorStrLenFunc(state *execState, min *int, max *int, fieldName string, code string, msg string, paths ...string) string {
	val, ok := getValueByPath(state, paths...)
	if !ok {
		return ""
	}
	v, ok := val.(string)
	if !ok {
		err := NewValidatorError(ErrValidatorTypeStrLen, fieldName, code, msg)
		err.SetPaths(paths...)
		state.addValidatorError(err)
		return ""
	}
	l := len(v)
	if min != nil && l < *min {
		err := NewValidatorError(ErrValidatorTypeStrLen, fieldName, code, msg)
		err.SetPaths(paths...)
		state.addValidatorError(err)
		return ""
	}
	if max != nil && l > *max {
		err := NewValidatorError(ErrValidatorTypeStrLen, fieldName, code, msg)
		err.SetPaths(paths...)
		state.addValidatorError(err)
		return ""
	}
	return ""
}

func validatorArrLenFunc(state *execState, min *int, max *int, fieldName string, code string, msg string, paths ...string) string {
	val, ok := getValueByPath(state, paths...)
	if !ok {
		return ""
	}
	v, ok := val.([]interface{})
	if !ok {
		err := NewValidatorError(ErrValidatorTypeArrLen, fieldName, code, msg)
		err.SetPaths(paths...)
		state.addValidatorError(err)
		return ""
	}
	l := len(v)
	if min != nil && l < *min {
		err := NewValidatorError(ErrValidatorTypeArrLen, fieldName, code, msg)
		err.SetPaths(paths...)
		state.addValidatorError(err)
		return ""
	}
	if max != nil && l > *max {
		err := NewValidatorError(ErrValidatorTypeArrLen, fieldName, code, msg)
		err.SetPaths(paths...)
		state.addValidatorError(err)
		return ""
	}
	return ""
}

func validatorRegFunc(state *execState, pattern string, fieldName string, code string, msg string, paths ...string) string {
	val, ok := getValueByPath(state, paths...)
	if !ok {
		return ""
	}
	v, ok := val.(string)
	if !ok {
		err := NewValidatorError(ErrValidatorTypeReg, fieldName, code, msg)
		err.SetPaths(paths...)
		state.addValidatorError(err)
		return ""
	}
	if !regexp.MustCompile(pattern).MatchString(v) {
		err := NewValidatorError(ErrValidatorTypeReg, fieldName, code, msg)
		err.SetPaths(paths...)
		state.addValidatorError(err)
		return ""
	}
	return ""
}
