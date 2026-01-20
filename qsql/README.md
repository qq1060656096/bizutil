# QSQL - SQL 占位符引擎

基于 Go `text/template` 实现的 SQL 占位符引擎，支持动态 SQL 生成、条件裁剪、逻辑组合和循环生成。

## ✨ 特性

- 🎯 **SQL 主体固定，条件占位符化** - 安全可控的 SQL 生成
- 🔄 **支持嵌套组合** - AND / OR / expr / if / for / val 任意嵌套
- 🔁 **循环生成** - 使用 Go template 的 range 遍历数组生成重复条件
- ✂️ **自动裁剪** - 空参数自动忽略，不生成冗余 SQL
- 💉 **动态值插入** - val 支持字面量、计算值、动态字段
- 🛡️ **预编译安全** - 输出标准的预编译 SQL + args，防止 SQL 注入
- 📊 **BI 查询能力** - 用户只需传 JSON，无需编写接口逻辑

## 📦 安装

```bash
go get github.com/qc/qsql
```

## 🚀 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/qc/qsql"
)

func main() {
    // 创建引擎
    engine := qsql.NewEngine()
    
    // 定义 SQL 模板
    tmpl := `SELECT * FROM user WHERE 1=1 {expr "name" "=" "$.params.name" }`
    
    // 解析模板
    engine.Parse("query", tmpl)
    
    // 执行查询
    params := map[string]interface{}{
        "params": map[string]interface{}{
            "name": "张三",
        },
    }
    
    result, _ := engine.ExecuteWithMap(params)
    
    fmt.Println(result.SQL)   // SELECT * FROM user WHERE 1=1 name = ?
    fmt.Println(result.Args)  // [张三]
}
```

## 📖 核心占位符

### 1️⃣ expr - 原子条件

生成单个字段的条件表达式。

**语法**：

```go
{expr "field" "op" "$.params.xxx" }
```

**支持的操作符**：
- `=`, `>`, `<`, `>=`, `<=`, `!=`, `<>` - 比较操作
- `in` - IN 查询（自动展开数组）
- `like` - 模糊匹配

**示例**：

```go
// 等值查询
{expr "name" "=" "$.params.name" }
// 生成: name = ?

// IN 查询
{expr "status" "in" "$.params.statuses" }
// 生成: status IN (?, ?, ?)

// 比较查询
{expr "age" ">" "$.params.min_age" }
// 生成: age > ?
```

**自动裁剪**：参数为空时，expr 不生成任何内容。

### 2️⃣ and / or - 逻辑组合

组合多个条件。

**语法**：

```go
{and . (expr1) (expr2) (expr3)}
{or (expr1) (expr2) (expr3)}
```

**示例**：

```go
{and .
    (expr "name" "=" "$.params.name" .)
    (expr "age" ">" "$.params.min_age" .)
}
// 生成: (name = ? AND age > ?)
// 如果没有有效条件，返回 1=1 并记录错误

{or 
    (expr "status" "=" "$.params.status1" .)
    (expr "status" "=" "$.params.status2" .)
}
// 生成: (status = ? OR status = ?)
```

**特性**：
- 支持任意层级嵌套
- 自动过滤空条件
- 只有一个有效条件时，不生成括号

### 3️⃣ if - 条件裁剪

控制整段 SQL 是否渲染。

**语法**：

```go
{if condition}
    SQL / 占位符
{end}
```

**示例**：

```go
{if not (_empty (_get "$.params.name" .))}
AND {expr "name" "=" "$.params.name" }
{end}
// 如果 name 参数存在，生成: AND name = ?
// 如果 name 参数为空，不生成任何内容
```

**辅助函数**：
- `_get` - 获取参数值
- `_empty` - 检查是否为空

### 4️⃣ for - 循环生成

使用 Go template 的 `range` 遍历数组生成重复条件。

**语法**：

```go
{range $item := _get "$.params.list" }
    SQL / 占位符
{end}
```

**示例**：

```go
{$ctx := }
{range $i, $uid := (_get "$.params.user_ids" .)}
{if $i} OR {end}user_id = {$uid}
{end}
// 生成: user_id = 1 OR user_id = 2 OR user_id = 3
```

### 5️⃣ val - 动态值插入

直接插入值到 SQL（不生成 `?` 占位符）。

**语法**：

```go
{val "$.params.xxx" }
```

**用途**：
- 常量
- 动态字段名
- 排序字段/方向
- 计算值

**示例**：

```go
SELECT * FROM user
WHERE id = {val "$.params.user_id"}
ORDER BY {val "$.params.sort_field"} {val "$.params.sort_order" }

// 生成: 
// SELECT * FROM user
// WHERE id = 123
// ORDER BY created_at DESC
```

⚠️ **安全提醒**：`val` 会直接插入 SQL，必须保证来源可信或已转义。

## 🎯 完整示例

### 示例 1: 动态查询

```go
engine := qsql.NewEngine()

tmpl := `SELECT * FROM user 
WHERE 1=1
{if not (_empty (_get "$.params.name" .))}
AND {expr "name" "=" "$.params.name" }
{end}
{if not (_empty (_get "$.params.min_age" .))}
AND {expr "age" ">=" "$.params.min_age" }
{end}
{if not (_empty (_get "$.params.statuses" .))}
AND {expr "status" "in" "$.params.statuses" }
{end}
ORDER BY {val "$.params.sort_field" } {val "$.params.sort_order" }
LIMIT {val "$.params.limit" }`

engine.Parse("dynamic_query", tmpl)

params := map[string]interface{}{
    "params": map[string]interface{}{
        "name":       "张三",
        "min_age":    18,
        "statuses":   []string{"active", "pending"},
        "sort_field": "created_at",
        "sort_order": "DESC",
        "limit":      10,
    },
}

result, _ := engine.ExecuteWithMap(params)

// SQL: SELECT * FROM user WHERE 1=1 AND name = ? AND age >= ? AND status IN (?, ?) ORDER BY created_at DESC LIMIT 10
// Args: [张三 18 active pending]
```

### 示例 2: 复杂嵌套

```go
tmpl := `SELECT * FROM orders
WHERE 1=1
{and
    (or
        (expr "order_no" "like" "$.params.search" .)
        (expr "customer_name" "like" "$.params.search" .)
    )
    (expr "status" "in" "$.params.statuses" .)
    (expr "total" ">=" "$.params.min_total" .)
}`

params := map[string]interface{}{
    "params": map[string]interface{}{
        "search":    "%ABC%",
        "statuses":  []string{"completed", "shipped"},
        "min_total": 100,
    },
}

// SQL: SELECT * FROM orders WHERE 1=1 ((order_no LIKE ? OR customer_name LIKE ?) AND status IN (?, ?) AND total >= ?)
// Args: [%ABC% %ABC% completed shipped 100]
```

### 示例 3: 循环生成

```go
tmpl := `SELECT * FROM user 
WHERE 1=1 AND (
{$ctx := }
{range $i, $uid := (_get "$.params.user_ids" .)}
{if $i} OR {end}user_id = {$uid}
{end}
)`

params := map[string]interface{}{
    "params": map[string]interface{}{
        "user_ids": []interface{}{1, 2, 3},
    },
}

// SQL: SELECT * FROM user WHERE 1=1 AND ( user_id = 1 OR user_id = 2 OR user_id = 3 )
```

## 📚 更多示例

查看 `examples/` 目录：
- `examples/basic/` - 基础用法示例
- `examples/advanced/` - 高级场景示例（包含 PRD 中的所有示例）

运行示例：

```bash
go run examples/basic/main.go
go run examples/advanced/main.go
```

## 🧪 测试

```bash
# 运行测试
go test -v

# 运行基准测试
go test -bench=. -benchmem

# 查看覆盖率
go test -cover
```

## 📋 API 参考

### Engine

```go
// 创建新引擎
func NewEngine() *Engine

// 解析模板
func (e *Engine) Parse(name, sqlTemplate string) error

// 执行模板（JSON 参数）
func (e *Engine) Execute(paramsJSON string) (*SQLStmt, error)

// 执行模板（map 参数）
func (e *Engine) ExecuteWithMap(params map[string]interface{}) (*SQLStmt, error)
```

### SQLStmt

```go
type SQLStmt struct {
    SQL  string        // 生成的 SQL
    Args []interface{} // 参数列表
}
```

## 🎨 设计原则

1. **SQL 主体固定** - 模板定义了 SQL 的结构，只有参数是动态的
2. **安全第一** - 默认使用预编译占位符 `?`，避免 SQL 注入
3. **自动裁剪** - 空参数自动忽略，生成最简洁的 SQL
4. **可组合** - 所有占位符可以任意嵌套组合
5. **标准兼容** - 基于 Go `text/template`，学习成本低

## 🔒 安全建议

1. **优先使用 `expr`**：生成预编译占位符，最安全
2. **谨慎使用 `val`**：只用于可信来源（如配置项、枚举值）
3. **验证输入**：对用户输入进行验证和清理
4. **白名单机制**：对动态字段名使用白名单验证

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可

MIT License

## 🔗 相关资源

- [PRD 文档](./prd.md) - 完整的产品需求文档
- [Go text/template 文档](https://pkg.go.dev/text/template)

---

**Made with ❤️ by qc**
Quick SQL
