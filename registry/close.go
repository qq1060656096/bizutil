package registry

import "context"

// Closer 是资源关闭器函数类型。
//
// Closer 定义了如何关闭/销毁资源实例。
// 在以下场景会被调用：
//   - Group.Unregister 注销资源时
//   - Group.Close 关闭整个组时
//   - Manager.Close 关闭整个管理器时
//
// 类型参数:
//   - T: 资源类型
//
// 参数:
//   - ctx: 上下文，可用于超时控制和取消操作
//   - t: 要关闭的资源实例
//
// 返回值:
//   - error: 关闭过程中的错误，nil 表示成功
//
// 注意:
//   - Closer 可以为 nil，此时资源不会被主动关闭
//   - 即使 Closer 返回错误，资源仍会从注册表中移除
//
// 示例:
//
//	closer := func(ctx context.Context, db *sql.DB) error {
//	    return db.Close()
//	}
type Closer[T any] func(ctx context.Context, t T) error
