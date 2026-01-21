# BizUtil

[![Go Reference](https://pkg.go.dev/badge/github.com/qq1060656096/bizutil.svg)](https://pkg.go.dev/github.com/qq1060656096/bizutil)
[![Go Version](https://img.shields.io/github/go-mod/go-version/qq1060656096/bizutil)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

BizUtil 是一个 Go 语言业务工具库，提供常用的泛型工具函数和资源管理框架，帮助你更高效地构建业务应用。

## 特性

- 🎯 **泛型支持** - 基于 Go 1.21+ 泛型，类型安全且灵活
- 🛡️ **并发安全** - 所有资源管理操作都是线程安全的
- ⏰ **惰性初始化** - 资源仅在首次访问时创建，减少启动时间
- 📦 **模块化设计** - 按需引入，避免不必要的依赖

## 安装

```bash
go get github.com/qq1060656096/bizutil
```

## 模块列表

| 模块 | 说明 | 文档 |
|------|------|------|
| [maputil](./maputil) | 泛型 Map 操作工具库 | [查看文档](./maputil/README.md) |
| [qsql](./qsql) | SQL 占位符引擎，支持动态 SQL 生成 | [查看文档](./qsql/README.md) |
| [registry](./registry) | 通用资源注册与管理框架 | [查看文档](./registry/README.md) |

## 快速开始

### maputil - Map 操作工具

提供简洁高效的 map 操作函数，支持安全访问和类型转换。

```go
import "github.com/qq1060656096/bizutil/maputil"

// 将切片转换为 map
type User struct {
    ID   int
    Name string
}

users := []User{
    {ID: 1, Name: "Alice"},
    {ID: 2, Name: "Bob"},
}

userMap := maputil.MapBy(users,
    func(u User) int { return u.ID },
    func(u User) string { return u.Name },
)
// userMap = map[int]string{1: "Alice", 2: "Bob"}

// 从 map 中安全获取值
name, ok := maputil.MapGet(userMap, 1, func(n string) string { return n })
// name = "Alice", ok = true
```

**主要函数：**

| 函数 | 说明 |
|------|------|
| `MapGet` | 从 map 中安全获取值，支持值转换 |
| `MapBy` | 将切片转换为 map |

### registry - 资源管理框架

通用的资源注册与管理框架，支持分组管理、惰性初始化和并发安全访问。

```go
import (
    "context"
    "database/sql"
    "github.com/qq1060656096/bizutil/registry"
)

type DBConfig struct {
    DSN string
}

// 创建资源管理器
group := registry.New[DBConfig, *sql.DB](
    // Opener: 定义如何创建资源
    func(ctx context.Context, cfg DBConfig) (*sql.DB, error) {
        return sql.Open("mysql", cfg.DSN)
    },
    // Closer: 定义如何关闭资源
    func(ctx context.Context, db *sql.DB) error {
        return db.Close()
    },
)

ctx := context.Background()

// 注册资源配置（此时不会创建连接）
group.Register(ctx, "main", DBConfig{DSN: "user:pass@tcp(localhost:3306)/db"})

// 获取资源（首次调用时会初始化连接）
db, err := group.Get(ctx, "main")

// 程序退出时关闭所有资源
defer group.Close(ctx)
```

**核心功能：**

| 功能 | 说明 |
|------|------|
| 分组管理 | 将资源按组进行分类管理 |
| 惰性初始化 | 资源仅在首次访问时创建 |
| 并发安全 | 所有操作都是线程安全的 |
| 自定义打开/关闭 | 支持自定义资源的创建和销毁逻辑 |

## 适用场景

- **maputil**: 数据转换、索引构建、切片与 map 互转
- **registry**: 数据库连接池、Redis 客户端、消息队列连接、gRPC 客户端等资源管理

## 要求

- Go 1.21 或更高版本

## License

[Apache License 2.0](LICENSE)
