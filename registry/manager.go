package registry

import "context"

// Manager 是资源管理器的顶层接口，负责管理多个资源组。
//
// 类型参数:
//   - C: 配置类型，用于创建资源
//   - T: 资源类型，被管理的资源实例类型
type Manager[C any, T any] interface {
	// Group 根据名称获取资源组。
	// 如果组不存在，返回 ErrGroupNotFound 错误。
	Group(name string) (Group[C, T], error)

	// MustGroup 根据名称获取资源组。
	// 如果组不存在，会触发 panic。
	MustGroup(name string) Group[C, T]

	// AddGroup 添加一个新的资源组。
	// 返回值表示组是否已经存在：
	//   - false: 组是新创建的
	//   - true: 组已经存在（不会重新创建）
	AddGroup(name string) bool

	// ListGroupNames 返回所有已注册的组名列表。
	ListGroupNames() []string

	// Close 关闭管理器中所有已初始化的资源。
	// 返回关闭过程中遇到的所有错误。
	// 调用后，管理器将被重置为空状态。
	Close(ctx context.Context) []error
}
