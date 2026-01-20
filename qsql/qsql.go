/*
Package qsql 提供基于 text/template 的 SQL 占位符引擎。

qsql 允许开发者使用模板语法动态生成安全的预编译 SQL 语句，
支持条件裁剪、逻辑组合和参数绑定，有效防止 SQL 注入。

# 核心特性

  - SQL 主体固定，条件占位符化 - 安全可控的 SQL 生成
  - 支持嵌套组合 - AND / OR / expr / if / range / val 任意嵌套
  - 自动裁剪 - 空参数自动忽略，不生成冗余 SQL
  - 预编译安全 - 输出标准的预编译 SQL + args，防止 SQL 注入

# 快速开始

创建引擎并解析 SQL 模板：

	engine := qsql.NewEngine()

	tmpl := `SELECT * FROM user WHERE 1=1
	{if not (isEmpty (getValue "$.params.name" .))}
	AND {expr "name" "=" "$.params.name" .}
	{end}`

	if err := engine.Parse("query", tmpl); err != nil {
	    log.Fatal(err)
	}

执行模板生成 SQL：

	paramsJSON := `{"params": {"name": "张三"}}`
	result, err := engine.Execute(paramsJSON)
	if err != nil {
	    log.Fatal(err)
	}

	fmt.Println(result.SQL)  // SELECT * FROM user WHERE 1=1 AND name = (?)
	fmt.Println(result.Args) // [张三]

# 模板语法

qsql 使用单花括号 { } 作为模板分隔符（区别于 Go 默认的 {{ }}）。

内置函数：

  - expr: 生成原子条件表达式，如 {expr "field" "op" "$.params.xxx" .}
  - and:  组合多个条件（AND 逻辑），如 {and . (expr ...) (expr ...)}
  - or:   组合多个条件（OR 逻辑），如 {or (expr ...) (expr ...)}
  - val:  插入动态值并生成占位符，如 {val "$.params.xxx" .}
  - getValue: 获取参数值，如 {getValue "$.params.xxx" .}
  - isEmpty: 检查值是否为空，如 {isEmpty value}

# 参数路径

参数使用 gjson 路径语法访问，支持三个命名空间：

  - $.params.xxx - 用户传入的查询参数
  - $.sys.xxx    - 系统参数（如当前用户ID）
  - $.users.xxx  - 用户相关信息

省略前缀时默认从 params 命名空间获取。

# 安全建议

  - 优先使用 expr 函数生成预编译占位符
  - val 函数会生成占位符并绑定值，确保来源可信
  - 对动态字段名使用白名单验证
*/
package qsql

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/tidwall/gjson"
)

// Engine 是 SQL 占位符引擎的核心类型。
//
// Engine 负责解析 SQL 模板并在执行时根据传入的参数动态生成
// 预编译的 SQL 语句和对应的参数列表。
//
// Engine 是并发不安全的，每个 goroutine 应该使用独立的 Engine 实例，
// 或者在调用 Execute 时进行适当的同步。
//
// 零值的 Engine 不可用，必须通过 [NewEngine] 创建并调用 [Engine.Parse] 后才能使用。
type Engine struct {
	template *template.Template
	rawSQL   string
}

// NewEngine 创建并返回一个新的 SQL 引擎实例。
//
// 返回的 Engine 需要调用 [Engine.Parse] 方法解析模板后才能执行。
//
// 示例：
//
//	engine := qsql.NewEngine()
//	engine.Parse("myQuery", "SELECT * FROM users WHERE id = {val \"$.params.id\" .}")
func NewEngine() *Engine {
	return &Engine{}
}

// Parse 解析给定的 SQL 模板字符串。
//
// name 是模板的名称，用于错误报告和调试。
// sqlTemplate 是包含占位符语法的 SQL 模板字符串。
//
// 模板使用单花括号 { } 作为分隔符，支持以下内置函数：
//   - expr: 原子条件表达式
//   - and:  AND 逻辑组合
//   - or:   OR 逻辑组合
//   - val:  动态值插入（生成占位符）
//   - getValue: 获取参数值
//   - isEmpty: 检查值是否为空
//
// 解析成功返回 nil，失败返回解析错误。
// Parse 应该在程序初始化阶段调用，解析错误通常表示模板语法问题。
//
// 示例：
//
//	engine := qsql.NewEngine()
//
//	// 简单查询
//	err := engine.Parse("simple", "SELECT * FROM user WHERE {expr \"id\" \"=\" \"$.params.id\" .}")
//
//	// 带条件裁剪的查询
//	err = engine.Parse("conditional", `
//	    SELECT * FROM orders WHERE 1=1
//	    {if not (isEmpty (getValue "$.params.status" .))}
//	    AND {expr "status" "=" "$.params.status" .}
//	    {end}
//	`)
//
//	// 逻辑组合查询
//	err = engine.Parse("combined", `
//	    SELECT * FROM products WHERE
//	    {and
//	        (expr "category" "=" "$.params.category" .)
//	        (or
//	            (expr "name" "like" "$.params.search" .)
//	            (expr "description" "like" "$.params.search" .)
//	        )
//	    }
//	`)
func (e *Engine) Parse(name, sqlTemplate string) error {
	e.rawSQL = sqlTemplate
	tmpl := template.New(name)
	// 设置自定义分隔符，使用单花括号 { }
	tmpl.Delims("{", "}")

	// 注册所有自定义函数
	tmpl.Funcs(template.FuncMap{
		// 原子条件
		"expr":    exprFunc,
		"optExpr": optionalExprFunc,
		// 逻辑组合
		"and": andFunc,
		"or":  orFunc,
		// 动态值插入
		"val": valFunc,
		// 辅助函数
		"getValue":  getValueByPathForTemplate,
		"isEmpty":   isEmpty,
		"printf":    fmt.Sprintf,
		"vInt":      validatorIntFunc,
		"vFloat":    validatorFloatFunc,
		"vStr":      validatorStrFunc,
		"vReg":      validatorRegFunc,
		"vRequired": validatorRequiredFunc,
	})

	var err error
	e.template, err = tmpl.Parse(sqlTemplate)
	return err
}

// Execute 使用给定的 JSON 参数执行已解析的模板，生成 SQL 语句。
//
// paramsJSON 必须是有效的 JSON 对象字符串，通常包含以下结构：
//
//	{
//	    "params": { ... },  // 用户查询参数
//	    "sys": { ... },     // 系统参数（可选）
//	    "users": { ... }    // 用户信息（可选）
//	}
//
// 返回的 [SQLStmt] 包含生成的 SQL 语句和对应的参数列表，
// 可直接用于数据库查询。
//
// 如果 paramsJSON 不是有效的 JSON 或模板执行出错，返回相应的错误。
//
// 示例：
//
//	engine := qsql.NewEngine()
//	engine.Parse("query", "SELECT * FROM user WHERE {expr \"name\" \"=\" \"$.params.name\" .}")
//
//	// 使用 JSON 字符串执行
//	result, err := engine.Execute(`{"params": {"name": "张三"}}`)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 使用生成的 SQL 和参数查询数据库
//	rows, err := db.Query(result.SQL, result.Args...)
func (e *Engine) Execute(paramsJSON string) (*SQLStmt, error) {
	// 验证 JSON 格式
	if !json.Valid([]byte(paramsJSON)) {
		return nil, fmt.Errorf("invalid JSON: %s", paramsJSON)
	}

	// 创建执行状态
	state := &execState{
		data: gjson.Parse(paramsJSON),
		args: make([]interface{}, 0),
	}

	// 执行模板
	var buf strings.Builder
	if err := e.template.Execute(&buf, state); err != nil {
		return nil, fmt.Errorf("template execute error: %w", err)
	}

	// 返回结果
	return &SQLStmt{
		RawSQL:           e.rawSQL,
		SQL:              cleanSQL(buf.String()),
		Args:             state.args,
		Errors:           state.errors,
		ValidatorsErrors: state.validatorsErrors,
	}, nil
}

// ExecuteWithVars 使用实现了 [Vars] 接口的对象执行模板。
//
// 此方法是 [Engine.Execute] 的便捷封装，内部将 Vars 转换为 JSON 字符串后调用 Execute。
// 适用于需要类型安全的参数传递场景。
//
// 示例：
//
//	// 使用 ValueVars
//	vars := qsql.NewValueVars()
//	vars.SetParam("name", "张三")
//	vars.SetParam("age", 25)
//
//	result, err := engine.ExecuteWithVars(vars)
//
//	// 使用 JSONVars
//	jsonVars := qsql.NewJSONVars(`{"params": {"status": "active"}}`)
//	result, err = engine.ExecuteWithVars(jsonVars)
func (e *Engine) ExecuteWithVars(vars Vars) (*SQLStmt, error) {
	return e.Execute(vars.JSON())
}
