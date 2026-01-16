package registry

import "context"

// Opener 是资源打开器函数类型。
//
// Opener 定义了如何根据配置创建资源实例。
// 在 Group.Get 首次访问资源时会被调用（惰性初始化）。
//
// 类型参数:
//   - C: 配置类型
//   - T: 资源类型
//
// 参数:
//   - ctx: 上下文，可用于超时控制和取消操作
//   - cfg: 资源配置，由 Register 时传入
//
// 返回值:
//   - T: 创建的资源实例
//   - error: 创建过程中的错误，nil 表示成功
//
// 示例:
//
//	opener := func(ctx context.Context, cfg DBConfig) (*sql.DB, error) {
//	    return sql.Open("mysql", cfg.DSN)
//	}
type Opener[C any, T any] func(ctx context.Context, cfg C) (T, error)
