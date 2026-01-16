# Registry - 通用资源注册与管理框架

[![Go Reference](https://pkg.go.dev/badge/github.com/qq1060656096/bizutil/registry.svg)](https://pkg.go.dev/github.com/qq1060656096/bizutil/registry)

`registry` 是一个 Go 语言实现的泛型资源管理框架，提供资源的分组管理、惰性初始化和并发安全访问。

## 特性

- 🎯 **泛型支持** - 支持任意类型的配置和资源
- 📦 **分组管理** - 将资源按组进行分类管理
- ⏰ **惰性初始化** - 资源仅在首次访问时才会被创建，减少启动时间和资源浪费
- 🔒 **并发安全** - 所有操作都是线程安全的，采用读写锁保护
- 🔌 **自定义打开/关闭** - 支持自定义资源的创建（Opener）和销毁（Closer）逻辑

## 安装

```bash
go get github.com/qq1060656096/bizutil/registry
```

## 快速开始

### 单组模式（推荐简单场景）

如果你只需要管理一类资源，不需要分组，可以使用 `NewGroup` 快速创建：

```go
package main

import (
    "context"
    "database/sql"
    "log"

    "github.com/qq1060656096/bizutil/registry"
    _ "github.com/go-sql-driver/mysql"
)

// 定义配置结构
type DBConfig struct {
    DSN string
}

func main() {
    ctx := context.Background()

    // 创建单组资源管理器
    group := registry.NewGroup[DBConfig, *sql.DB](
        // Opener: 定义如何创建资源
        func(ctx context.Context, cfg DBConfig) (*sql.DB, error) {
            return sql.Open("mysql", cfg.DSN)
        },
        // Closer: 定义如何关闭资源
        func(ctx context.Context, db *sql.DB) error {
            return db.Close()
        },
    )

    // 注册资源配置（此时不会创建连接）
    group.Register(ctx, "main", DBConfig{DSN: "user:pass@tcp(localhost:3306)/db"})
    group.Register(ctx, "backup", DBConfig{DSN: "user:pass@tcp(localhost:3307)/db"})

    // 获取资源（首次调用时会初始化连接）
    db, err := group.Get(ctx, "main")
    if err != nil {
        log.Fatal(err)
    }

    // 使用数据库连接
    _ = db

    // 程序退出时关闭所有资源
    defer group.Close(ctx)
}
```

### 多组模式（适合复杂场景）

如果需要按组分类管理资源（如主从分离、多服务数据库），使用 `New` 创建管理器：

```go
package main

import (
    "context"
    "database/sql"
    "log"

    "github.com/qq1060656096/bizutil/registry"
    _ "github.com/go-sql-driver/mysql"
)

// 定义配置结构
type DBConfig struct {
    DSN string
}

func main() {
    ctx := context.Background()

    // 创建管理器
    mgr := registry.New[DBConfig, *sql.DB](
        // Opener: 定义如何创建资源
        func(ctx context.Context, cfg DBConfig) (*sql.DB, error) {
            return sql.Open("mysql", cfg.DSN)
        },
        // Closer: 定义如何关闭资源
        func(ctx context.Context, db *sql.DB) error {
            return db.Close()
        },
    )

    // 添加资源组
    mgr.AddGroup("master")
    mgr.AddGroup("slave")

    // 获取组并注册资源（此时不会创建连接）
    masterGroup, _ := mgr.Group("master")
    masterGroup.Register(ctx, "db1", DBConfig{DSN: "user:pass@tcp(host1:3306)/db"})
    masterGroup.Register(ctx, "db2", DBConfig{DSN: "user:pass@tcp(host2:3306)/db"})

    // 获取资源（首次调用时会初始化连接）
    db1, err := masterGroup.Get(ctx, "db1")
    if err != nil {
        log.Fatal(err)
    }

    // 使用数据库连接
    _ = db1

    // 程序退出时关闭所有资源
    defer mgr.Close(ctx)
}
```

## 核心概念

### Manager（管理器）

`Manager` 是整个注册表的顶层管理接口，负责管理多个资源组。

```go
type Manager[C any, T any] interface {
    // 添加新的资源组，返回是否已存在
    AddGroup(name string) bool

    // 获取指定名称的资源组
    Group(name string) (Group[C, T], error)

    // 获取资源组，不存在时 panic
    MustGroup(name string) Group[C, T]

    // 列出所有组名
    ListGroupNames() []string

    // 关闭所有已初始化的资源
    Close(ctx context.Context) []error
}
```

### Group（资源组）

`Group` 是一组相关资源的容器，每个资源通过唯一名称标识。

```go
type Group[C any, T any] interface {
    // 注册资源配置（此时不会创建资源）
    Register(ctx context.Context, name string, cfg C) (isNew bool, err error)

    // 获取资源（首次调用时会触发惰性初始化）
    Get(ctx context.Context, name string) (T, error)

    // 获取资源，失败时 panic
    MustGet(ctx context.Context, name string) T

    // 注销资源并关闭
    Unregister(ctx context.Context, name string) error

    // 列出组内所有资源名称
    List() []string

    // 关闭组内所有资源
    Close(ctx context.Context) []error
}
```

### Opener（打开器）

`Opener` 定义了如何根据配置创建资源实例：

```go
type Opener[C any, T any] func(ctx context.Context, cfg C) (T, error)
```

**示例：**

```go
// 数据库连接打开器
opener := func(ctx context.Context, cfg DBConfig) (*sql.DB, error) {
    return sql.Open("mysql", cfg.DSN)
}

// Redis 连接打开器
redisOpener := func(ctx context.Context, cfg RedisConfig) (*redis.Client, error) {
    return redis.NewClient(&redis.Options{
        Addr:     cfg.Addr,
        Password: cfg.Password,
    }), nil
}
```

### Closer（关闭器）

`Closer` 定义了如何关闭/销毁资源实例：

```go
type Closer[T any] func(ctx context.Context, t T) error
```

**示例：**

```go
// 数据库连接关闭器
closer := func(ctx context.Context, db *sql.DB) error {
    return db.Close()
}

// Redis 连接关闭器
redisCloser := func(ctx context.Context, client *redis.Client) error {
    return client.Close()
}
```

**注意：** Closer 可以为 `nil`，此时资源不会被主动关闭。

## 使用示例

### 惰性初始化

资源只有在首次通过 `Get` 或 `MustGet` 访问时才会被创建：

```go
// 注册时只保存配置，不创建连接
group.Register(ctx, "redis", RedisConfig{Addr: "localhost:6379"})

// 首次 Get 时才会调用 Opener 创建连接
client, _ := group.Get(ctx, "redis")

// 后续 Get 直接返回已创建的实例，不会重复创建
client, _ = group.Get(ctx, "redis")
```

### 资源清理

**关闭单个资源：**

```go
err := group.Unregister(ctx, "db1")
if err != nil {
    log.Printf("注销失败: %v", err)
}
```

**关闭整个组的资源：**

```go
errs := group.Close(ctx)
for _, err := range errs {
    log.Printf("关闭失败: %v", err)
}
```

**关闭管理器中所有资源：**

```go
errs := mgr.Close(ctx)
for _, err := range errs {
    log.Printf("关闭失败: %v", err)
}
```

### MustGet / MustGroup

当你确定资源/组一定存在时，可以使用 `Must` 系列方法简化代码：

```go
// 不需要处理 error，失败时会 panic
group := mgr.MustGroup("master")
db := group.MustGet(ctx, "db1")
```

### 完整示例：多数据库管理

```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"

    "github.com/qq1060656096/bizutil/registry"
    _ "github.com/go-sql-driver/mysql"
)

type DBConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    Database string
}

func (c DBConfig) DSN() string {
    return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
        c.User, c.Password, c.Host, c.Port, c.Database)
}

func main() {
    ctx := context.Background()

    // 创建数据库管理器
    dbManager := registry.New[DBConfig, *sql.DB](
        func(ctx context.Context, cfg DBConfig) (*sql.DB, error) {
            db, err := sql.Open("mysql", cfg.DSN())
            if err != nil {
                return nil, err
            }
            // 设置连接池参数
            db.SetMaxOpenConns(100)
            db.SetMaxIdleConns(10)
            return db, db.PingContext(ctx)
        },
        func(ctx context.Context, db *sql.DB) error {
            return db.Close()
        },
    )
    defer dbManager.Close(ctx)

    // 创建分组
    dbManager.AddGroup("user")    // 用户服务数据库
    dbManager.AddGroup("order")   // 订单服务数据库
    dbManager.AddGroup("product") // 商品服务数据库

    // 注册各服务的主从数据库
    userGroup := dbManager.MustGroup("user")
    userGroup.Register(ctx, "master", DBConfig{
        Host: "user-master.db.local", Port: 3306,
        User: "root", Password: "pass", Database: "user",
    })
    userGroup.Register(ctx, "slave-1", DBConfig{
        Host: "user-slave1.db.local", Port: 3306,
        User: "root", Password: "pass", Database: "user",
    })
    userGroup.Register(ctx, "slave-2", DBConfig{
        Host: "user-slave2.db.local", Port: 3306,
        User: "root", Password: "pass", Database: "user",
    })

    // 使用时按需获取（惰性初始化）
    masterDB, err := userGroup.Get(ctx, "master")
    if err != nil {
        log.Fatalf("获取主库失败: %v", err)
    }

    // 执行查询...
    _ = masterDB

    // 列出所有已注册的数据库组
    fmt.Println("已注册的数据库组:", dbManager.ListGroupNames())
}
```

## 错误处理

包中定义了以下哨兵错误，可使用 `errors.Is` 进行判断：

| 错误 | 说明 |
|------|------|
| `ErrGroupNotFound` | 指定的组不存在 |
| `ErrResourceNotFound` | 指定的资源在组中不存在 |
| `ErrCloseResourceFailed` | 关闭资源时发生错误 |

**示例：**

```go
import "errors"

// 处理组不存在
_, err := mgr.Group("nonexistent")
if errors.Is(err, registry.ErrGroupNotFound) {
    log.Println("组不存在，需要先添加")
    mgr.AddGroup("nonexistent")
}

// 处理资源不存在
_, err = group.Get(ctx, "unknown")
if errors.Is(err, registry.ErrResourceNotFound) {
    log.Println("资源未注册")
}

// 处理关闭失败
errs := mgr.Close(ctx)
for _, err := range errs {
    if errors.Is(err, registry.ErrCloseResourceFailed) {
        log.Printf("资源关闭失败: %v", err)
    }
}
```

## 并发安全

所有公开的方法都是并发安全的，内部使用读写锁（`sync.RWMutex`）保护：

- **读操作**（`Get` 已初始化资源、`List`、`ListGroupNames`）使用读锁，支持并发读取
- **写操作**（`Register`、`Unregister`、`Close`、惰性初始化）使用写锁
- **双重检查锁定**：在惰性初始化时避免重复创建资源

```go
// 可以安全地在多个 goroutine 中并发访问
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        db, _ := group.Get(ctx, "db1") // 并发安全
        _ = db
    }()
}
wg.Wait()
```

## 设计模式

本包采用了以下设计模式：

- **注册表模式（Registry Pattern）**：集中管理和访问资源
- **惰性初始化模式（Lazy Initialization）**：延迟资源创建，减少启动时间和资源浪费
- **双重检查锁定（Double-Checked Locking）**：在惰性初始化时确保只创建一次资源

## 适用场景

- 数据库连接池管理
- 缓存客户端管理（Redis、Memcached 等）
- 消息队列连接管理（Kafka、RabbitMQ 等）
- gRPC 客户端连接管理
- 任何需要分组管理且支持惰性加载的资源

## API 参考

### 创建管理器

```go
func New[C any, T any](opener Opener[C, T], closer Closer[T]) Manager[C, T]
```

### Manager 方法

| 方法 | 说明 |
|------|------|
| `AddGroup(name string) bool` | 添加资源组，返回是否已存在 |
| `Group(name string) (Group, error)` | 获取资源组 |
| `MustGroup(name string) Group` | 获取资源组，不存在时 panic |
| `ListGroupNames() []string` | 列出所有组名 |
| `Close(ctx context.Context) []error` | 关闭所有资源 |

### Group 方法

| 方法 | 说明 |
|------|------|
| `Register(ctx, name, cfg) (bool, error)` | 注册资源配置 |
| `Get(ctx, name) (T, error)` | 获取资源（惰性初始化） |
| `MustGet(ctx, name) T` | 获取资源，失败时 panic |
| `Unregister(ctx, name) error` | 注销并关闭资源 |
| `List() []string` | 列出所有资源名 |
| `Close(ctx) []error` | 关闭组内所有资源 |


