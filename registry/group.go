package registry

import "context"

// Group 是资源组接口，用于管理一组相关的资源。
//
// 每个资源通过唯一的名称标识，资源采用惰性初始化策略，
// 即只有在首次通过 Get 或 MustGet 访问时才会创建。
//
// 类型参数:
//   - C: 配置类型，用于创建资源
//   - T: 资源类型，被管理的资源实例类型
type Group[C any, T any] interface {
	// Get 根据名称获取资源。
	//
	// 如果资源尚未初始化，会调用 Opener 进行惰性初始化。
	// 后续调用将直接返回已创建的资源实例。
	//
	// 可能返回的错误:
	//   - ErrGroupNotFound: 组不存在
	//   - ErrResourceNotFound: 资源未注册
	//   - Opener 返回的错误: 资源创建失败
	Get(ctx context.Context, name string) (T, error)

	// MustGet 根据名称获取资源。
	// 如果获取失败，会触发 panic。
	MustGet(ctx context.Context, name string) T

	Config(ctx context.Context, name string) (C, error)
	MustConfig(ctx context.Context, name string) C

	// Register 向组中注册一个新的资源配置。
	//
	// 注意：此方法只保存配置，不会立即创建资源。
	// 资源将在首次通过 Get 访问时惰性初始化。
	//
	// 返回值:
	//   - isNew: true 表示新注册成功，false 表示资源名已存在（不会覆盖）
	//   - err: 目前始终为 nil，保留用于将来扩展
	Register(ctx context.Context, name string, cfg C) (isNew bool, err error)

	// Unregister 从组中注销指定资源。
	//
	// 如果资源已初始化，会先调用 Closer 关闭资源。
	// 如果资源不存在，返回 ErrResourceNotFound 错误。
	Unregister(ctx context.Context, name string) error

	// List 返回组内所有已注册的资源名称列表。
	List() []string

	// Close 关闭组内所有已初始化的资源。
	// 返回关闭过程中遇到的所有错误。
	// 调用后，整个组将从管理器中移除。
	Close(ctx context.Context) []error

	// Ping 遍历组内所有已注册资源，尝试初始化以验证可用性。
	//
	// Ping 不会将资源保存到组中。
	// 返回的 errors 列表包含所有无法初始化的资源及其错误。
	Ping(ctx context.Context, name string) error
}
