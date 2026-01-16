# MapUtil - 泛型 Map 操作工具库

[![Go Reference](https://pkg.go.dev/badge/github.com/qq1060656096/bizutil/maputil.svg)](https://pkg.go.dev/github.com/qq1060656096/bizutil/maputil)

`maputil` 是一个 Go 语言泛型工具库，提供简洁高效的 map 操作函数。

## 特性

- 🎯 **泛型支持** - 支持任意键值类型
- 🔄 **类型转换** - 获取值时可进行类型转换
- 🛡️ **安全访问** - 安全处理空 map 和不存在的键
- 📦 **切片转 Map** - 一行代码将切片转换为 map

## 安装

```bash
go get github.com/qq1060656096/bizutil/maputil
```

## 函数列表

| 函数 | 说明 |
|------|------|
| `MapGet` | 从 map 中安全获取值，支持值转换 |
| `MapBy` | 将切片转换为 map |

## MapGet

从 map 中安全地获取值，并支持可选的值转换。

### 函数签名

```go
func MapGet[T any, K comparable, V any](m map[K]T, key K, value func(T) V) (V, bool)
```

### 参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `m` | `map[K]T` | 源 map |
| `key` | `K` | 要查找的键 |
| `value` | `func(T) V` | 值转换函数，传入 `nil` 时返回零值 |

### 返回值

| 返回值 | 说明 |
|--------|------|
| 第一个 | 转换后的值，若 key 不存在或 value 为 nil 则返回零值 |
| 第二个 | key 是否存在于 map 中 |

### 使用示例

**基础用法：获取并转换值**

```go
m := map[string]int{"a": 1, "b": 2, "c": 3}
v, ok := maputil.MapGet(m, "b", func(i int) int { return i * 10 })
// v = 20, ok = true
```

**从结构体 map 中提取字段**

```go
type User struct {
    Name string
    Age  int
}

users := map[int]User{
    1: {Name: "Alice", Age: 30},
    2: {Name: "Bob", Age: 25},
}

name, ok := maputil.MapGet(users, 1, func(u User) string { return u.Name })
// name = "Alice", ok = true
```

**类型转换：int 转 string**

```go
m := map[string]int{"count": 42}
v, ok := maputil.MapGet(m, "count", func(i int) string {
    if i > 10 {
        return "large"
    }
    return "small"
})
// v = "large", ok = true
```

**处理不存在的键**

```go
m := map[string]int{"a": 1}
v, ok := maputil.MapGet(m, "notexist", func(i int) int { return i * 10 })
// v = 0 (零值), ok = false
```

**安全处理 nil map**

```go
var m map[string]int
v, ok := maputil.MapGet(m, "any", func(i int) int { return i })
// v = 0, ok = false (不会 panic)
```

**仅检查键是否存在**

```go
m := map[string]int{"a": 1, "b": 2}
v, ok := maputil.MapGet[int, string, int](m, "a", nil)
// v = 0, ok = true (value 为 nil 时返回零值，但 ok 仍正确反映键是否存在)
```

## MapBy

将切片转换为 map，通过指定的函数分别提取键和值。

### 函数签名

```go
func MapBy[T any, K comparable, V any](list []T, key func(T) K, value func(T) V) map[K]V
```

### 参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `list` | `[]T` | 源切片 |
| `key` | `func(T) K` | 键提取函数，从切片元素中提取 map 的键 |
| `value` | `func(T) V` | 值提取函数，从切片元素中提取 map 的值 |

### 返回值

由切片元素构建的 map。

> **注意：** 若多个元素产生相同的键，后者会覆盖前者。

### 使用示例

**基础用法：结构体切片转 map**

```go
type User struct {
    ID   int
    Name string
}

users := []User{
    {ID: 1, Name: "Alice"},
    {ID: 2, Name: "Bob"},
}

m := maputil.MapBy(users, 
    func(u User) int { return u.ID }, 
    func(u User) string { return u.Name },
)
// m = map[int]string{1: "Alice", 2: "Bob"}
```

**构建 ID 到对象的索引**

```go
type Product struct {
    SKU   string
    Price float64
}

products := []Product{
    {SKU: "A001", Price: 9.99},
    {SKU: "B002", Price: 19.99},
}

m := maputil.MapBy(products,
    func(p Product) string { return p.SKU },
    func(p Product) Product { return p },
)
// m["A001"] = Product{SKU: "A001", Price: 9.99}
```

**字符串切片按首字母分组**

```go
list := []string{"apple", "banana", "cherry"}
m := maputil.MapBy(list,
    func(s string) string { return s[:1] },
    func(s string) int { return len(s) },
)
// m = map[string]int{"a": 5, "b": 6, "c": 6}
```

**处理重复键（后者覆盖前者）**

```go
type Item struct {
    ID   int
    Name string
}

list := []Item{
    {ID: 1, Name: "first"},
    {ID: 2, Name: "second"},
    {ID: 1, Name: "third"}, // 重复 ID
}

m := maputil.MapBy(list,
    func(i Item) int { return i.ID },
    func(i Item) string { return i.Name },
)
// m = map[int]string{1: "third", 2: "second"}
// ID=1 的 "third" 覆盖了 "first"
```

**处理空切片和 nil 切片**

```go
// 空切片
m1 := maputil.MapBy([]int{}, func(i int) int { return i }, func(i int) string { return "x" })
// m1 = map[int]string{} (空 map，非 nil)

// nil 切片
var list []int
m2 := maputil.MapBy(list, func(i int) int { return i }, func(i int) string { return "x" })
// m2 = map[int]string{} (空 map，非 nil)
```

**指针切片处理**

```go
type Data struct {
    Key   string
    Value int
}

list := []*Data{
    {Key: "x", Value: 1},
    {Key: "y", Value: 2},
}

m := maputil.MapBy(list,
    func(d *Data) string { return d.Key },
    func(d *Data) int { return d.Value },
)
// m = map[string]int{"x": 1, "y": 2}
```

## 完整示例

```go
package main

import (
    "fmt"
    "github.com/qq1060656096/bizutil/maputil"
)

type User struct {
    ID     int
    Name   string
    Email  string
    Active bool
}

func main() {
    // 模拟从数据库获取的用户列表
    users := []User{
        {ID: 1, Name: "Alice", Email: "alice@example.com", Active: true},
        {ID: 2, Name: "Bob", Email: "bob@example.com", Active: false},
        {ID: 3, Name: "Charlie", Email: "charlie@example.com", Active: true},
    }

    // 使用 MapBy 构建 ID -> User 索引
    userByID := maputil.MapBy(users,
        func(u User) int { return u.ID },
        func(u User) User { return u },
    )

    // 使用 MapBy 构建 Email -> Name 映射
    nameByEmail := maputil.MapBy(users,
        func(u User) string { return u.Email },
        func(u User) string { return u.Name },
    )

    // 使用 MapGet 安全获取用户名
    name, ok := maputil.MapGet(userByID, 1, func(u User) string { return u.Name })
    if ok {
        fmt.Printf("用户 1 的名字: %s\n", name) // 输出: 用户 1 的名字: Alice
    }

    // 使用 MapGet 检查用户是否活跃
    active, ok := maputil.MapGet(userByID, 2, func(u User) bool { return u.Active })
    if ok {
        fmt.Printf("用户 2 是否活跃: %v\n", active) // 输出: 用户 2 是否活跃: false
    }

    // 通过 Email 查找用户名
    fmt.Printf("charlie@example.com 的用户名: %s\n", nameByEmail["charlie@example.com"])
    // 输出: charlie@example.com 的用户名: Charlie
}
```

## License

Apache License 2.0
