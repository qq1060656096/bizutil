/*
Package registry 提供了一个通用的资源注册与管理框架。

# 概述

registry 包实现了一个泛型资源管理器，支持：
  - 资源分组管理：将资源按组进行分类管理
  - 惰性初始化：资源仅在首次访问时才会被创建
  - 并发安全：所有操作都是线程安全的
  - 自定义打开/关闭：支持自定义资源的创建和销毁逻辑

# 核心概念

## Manager（管理器）

Manager 是整个注册表的顶层管理接口，负责管理多个资源组。

主要功能：
  - AddGroup: 添加新的资源组
  - Group/MustGroup: 获取指定名称的资源组
  - ListGroupNames: 列出所有组名
  - Close: 关闭所有已初始化的资源

## Group（资源组）

Group 是一组相关资源的容器，每个资源通过唯一名称标识。

主要功能：
  - Register: 注册资源配置（此时不会创建资源）
  - Get/MustGet: 获取资源（首次调用时会触发惰性初始化）
  - Unregister: 注销资源并关闭
  - List: 列出组内所有资源名称
  - Close: 关闭组内所有资源

## Opener（打开器）

Opener 是一个函数类型，定义了如何根据配置创建资源：

	type Opener[C any, T any] func(ctx context.Context, cfg C) (T, error)

参数说明：
  - C: 配置类型
  - T: 资源类型
  - ctx: 上下文，用于超时控制和取消
  - cfg: 资源配置

## Closer（关闭器）

Closer 是一个函数类型，定义了如何关闭/销毁资源：

	type Closer[T any] func(ctx context.Context, t T) error

# 使用示例

## 基础用法

创建一个数据库连接管理器：

	// 定义配置结构
	type DBConfig struct {
	    DSN string
	}

	// 创建管理器
	mgr := registry.New[DBConfig, *sql.DB](
	    func(ctx context.Context, cfg DBConfig) (*sql.DB, error) {
	        return sql.Open("mysql", cfg.DSN)
	    },
	    func(ctx context.Context, db *sql.DB) error {
	        return db.Close()
	    },
	)

	// 添加资源组
	mgr.AddGroup("master")
	mgr.AddGroup("slave")

	// 获取组并注册资源
	masterGroup, _ := mgr.Group("master")
	masterGroup.Register(ctx, "db1", DBConfig{DSN: "user:pass@tcp(host1:3306)/db"})
	masterGroup.Register(ctx, "db2", DBConfig{DSN: "user:pass@tcp(host2:3306)/db"})

	// 获取资源（首次调用时会初始化连接）
	db1, err := masterGroup.Get(ctx, "db1")
	if err != nil {
	    log.Fatal(err)
	}

	// 使用 MustGet（失败时会 panic）
	db2 := masterGroup.MustGet(ctx, "db2")

## 惰性初始化

资源只有在首次通过 Get 或 MustGet 访问时才会被创建：

	// 注册时只保存配置，不创建连接
	group.Register(ctx, "redis", RedisConfig{Addr: "localhost:6379"})

	// 首次 Get 时才会调用 Opener 创建连接
	client, _ := group.Get(ctx, "redis")

	// 后续 Get 直接返回已创建的实例
	client, _ = group.Get(ctx, "redis") // 不会重复创建

## 资源清理

关闭单个资源：

	err := group.Unregister(ctx, "db1")

关闭整个组的资源：

	errs := group.Close(ctx)
	for _, err := range errs {
	    log.Printf("关闭失败: %v", err)
	}

关闭管理器中所有资源：

	errs := mgr.Close(ctx)

# 错误处理

包中定义了以下错误类型：

  - ErrGroupNotFound: 指定的组不存在
  - ErrResourceNotFound: 指定的资源不存在
  - ErrCloseResourceFailed: 关闭资源时发生错误

可以使用 errors.Is 进行错误类型判断。

# 并发安全

所有公开的方法都是并发安全的，内部使用读写锁（sync.RWMutex）保护：

  - 读操作（Get 已初始化资源、List）使用读锁，支持并发读取
  - 写操作（Register、Unregister、Close、惰性初始化）使用写锁

# 设计模式

本包采用了以下设计模式：

  - 注册表模式：集中管理和访问资源
  - 惰性初始化模式：延迟资源创建，减少启动时间和资源浪费
  - 双重检查锁定：在惰性初始化时避免重复创建

# 适用场景

  - 数据库连接池管理
  - 缓存客户端管理（Redis、Memcached 等）
  - 消息队列连接管理
  - 任何需要分组管理且支持惰性加载的资源
*/
package registry
