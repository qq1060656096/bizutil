package registry

import (
	"context"
	"sync"
)

// defaultGroupName 是使用 NewGroup 创建单组资源管理器时的默认组名。
const defaultGroupName = "defaultGroup"

// 类型断言检查在测试文件中进行
// var _ Manager[C, T] = (*manager[C, T])(nil)
// var _ Group[C, T] = (*group[C, T])(nil)

// New 创建一个新的资源管理器。
//
// 参数:
//   - opener: 资源打开器，用于根据配置创建资源实例
//   - closer: 资源关闭器，用于关闭/销毁资源（可以为 nil）
//
// 类型参数:
//   - C: 配置类型
//   - T: 资源类型
func New[C any, T any](opener Opener[C, T], closer Closer[T]) Manager[C, T] {
	return &manager[C, T]{
		groups: make(map[string]map[string]*connection[C, T]),
		opener: opener,
		closer: closer,
	}
}

// connection 表示一个资源连接的内部状态。
//
// 类型参数:
//   - C: 配置类型
//   - T: 资源类型
type connection[C any, T any] struct {
	cfg   C    // cfg 是创建资源所需的配置
	val   T    // val 是已创建的资源实例
	ready bool // ready 标记资源是否已通过 opener 完成初始化
}

// manager 是 Manager 接口的具体实现，负责管理多个资源组。
//
// 类型参数:
//   - C: 配置类型
//   - T: 资源类型
type manager[C any, T any] struct {
	mu     sync.RWMutex                            // mu 用于保护并发访问
	groups map[string]map[string]*connection[C, T] // groups 存储所有资源组，外层 key 为组名，内层 key 为资源名

	opener Opener[C, T] // opener 用于创建资源实例
	closer Closer[T]    // closer 用于关闭资源实例（可为 nil）
}

// Group 根据名称获取资源组。
//
// 如果指定名称的组不存在，返回 ErrGroupNotFound 错误。
// 返回的 Group 对象可用于在该组内注册和获取资源。
func (m *manager[C, T]) Group(name string) (Group[C, T], error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, ok := m.groups[name]; !ok {
		return nil, NewErrGroupNotFound(name)
	}

	return &group[C, T]{
		name: name,
		m:    m,
	}, nil
}

// Close 关闭管理器中所有已初始化的资源。
//
// 遍历所有组中的所有资源，对已初始化（ready=true）的资源调用 closer 进行关闭。
// 关闭完成后，管理器将被重置为空状态（所有组和资源配置都会被清除）。
//
// 返回值:
//   - []error: 关闭过程中遇到的所有错误，每个错误都包含组名和资源名信息
func (m *manager[C, T]) Close(ctx context.Context) []error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error

	for groupName, groupMap := range m.groups {
		for name, conn := range groupMap {
			if !conn.ready {
				continue
			}
			if m.closer == nil {
				continue
			}
			if err := m.closer(ctx, conn.val); err != nil {
				errs = append(errs, NewErrCloseResourceFailed(groupName, name, err))
			}
		}
	}

	// 清空所有组
	m.groups = make(map[string]map[string]*connection[C, T])
	return errs
}

// MustGroup 根据名称获取资源组，如果组不存在则触发 panic。
//
// 此方法是 Group 的便捷封装，适用于确定组一定存在的场景。
// 如果不确定组是否存在，请使用 Group 方法并处理返回的错误。
func (m *manager[C, T]) MustGroup(name string) Group[C, T] {
	g, err := m.Group(name)
	if err != nil {
		panic(err)
	}
	return g
}

// AddGroup 添加一个新的资源组。
//
// 如果指定名称的组不存在，则创建一个新的空组。
// 如果组已存在，不会进行任何操作。
//
// 返回值:
//   - false: 组是新创建的
//   - true: 组已经存在（未做任何修改）
func (m *manager[C, T]) AddGroup(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.groups[name]
	if !ok {
		m.groups[name] = make(map[string]*connection[C, T])
		return false
	}
	return true
}

// ListGroupNames 返回所有已注册的组名列表。
//
// 返回的列表顺序不保证固定（依赖 map 遍历顺序）。
func (m *manager[C, T]) ListGroupNames() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	groupNames := make([]string, 0, len(m.groups))
	for name := range m.groups {
		groupNames = append(groupNames, name)
	}
	return groupNames
}

// group 是 Group 接口的具体实现，代表一个资源组。
//
// group 通过持有 manager 的引用来访问和操作资源，
// 所有操作都会通过 manager 的锁来保证并发安全。
//
// 类型参数:
//   - C: 配置类型
//   - T: 资源类型
type group[C any, T any] struct {
	name string         // name 是该组的唯一标识名称
	m    *manager[C, T] // m 是所属的资源管理器
}

// Get 根据名称获取资源，支持惰性初始化。
//
// 实现采用双重检查锁定（Double-Checked Locking）模式：
//  1. 首先使用读锁检查资源是否已初始化
//  2. 如果已初始化，直接返回缓存的资源
//  3. 如果未初始化，升级为写锁并调用 opener 创建资源
//  4. 创建后标记为 ready，后续调用将直接返回
//
// 可能返回的错误:
//   - ErrGroupNotFound: 组不存在（可能已被关闭）
//   - ErrResourceNotFound: 资源未注册
//   - opener 返回的错误: 资源创建失败
func (g *group[C, T]) Get(ctx context.Context, name string) (T, error) {
	var zero T

	// 读锁：快速路径，检查资源是否已初始化
	g.m.mu.RLock()
	groupMap, ok := g.m.groups[g.name]
	if !ok {
		g.m.mu.RUnlock()
		return zero, NewErrGroupNotFound(g.name)
	}

	conn, ok := groupMap[name]
	if !ok {
		g.m.mu.RUnlock()
		return zero, NewErrResourceNotFound(g.name, name)
	}

	if conn.ready {
		val := conn.val
		g.m.mu.RUnlock()
		return val, nil
	}
	g.m.mu.RUnlock()

	// 写锁：慢速路径，惰性创建资源
	g.m.mu.Lock()
	defer g.m.mu.Unlock()

	// 双重检查：在获取写锁期间，其他 goroutine 可能已删除组或资源
	groupMap, ok = g.m.groups[g.name]
	if !ok {
		return zero, NewErrGroupNotFound(g.name)
	}

	conn, ok = groupMap[name]
	if !ok {
		return zero, NewErrResourceNotFound(g.name, name)
	}

	if conn.ready {
		return conn.val, nil
	}

	val, err := g.m.opener(ctx, conn.cfg)
	if err != nil {
		return zero, err
	}

	conn.val = val
	conn.ready = true
	return val, nil
}

// MustGet 根据名称获取资源，如果获取失败则触发 panic。
//
// 此方法是 Get 的便捷封装，适用于确定资源一定存在且能成功创建的场景。
// 如果不确定，请使用 Get 方法并处理返回的错误。
func (g *group[C, T]) MustGet(ctx context.Context, name string) T {
	val, err := g.Get(ctx, name)
	if err != nil {
		panic(err)
	}
	return val
}

// Register 向组中注册一个新的资源配置。
//
// 注意事项:
//   - 此方法只保存配置，不会立即创建资源实例
//   - 资源将在首次通过 Get 访问时惰性初始化
//   - 如果资源名已存在，不会覆盖原有配置
//   - 如果组不存在（已被关闭），会自动重新创建组
//
// 返回值:
//   - isNew: true 表示新注册成功，false 表示资源名已存在
//   - err: 目前始终为 nil，保留用于将来扩展
func (g *group[C, T]) Register(ctx context.Context, name string, cfg C) (bool, error) {
	g.m.mu.Lock()
	defer g.m.mu.Unlock()

	groupMap, ok := g.m.groups[g.name]
	if !ok {
		groupMap = make(map[string]*connection[C, T])
		g.m.groups[g.name] = groupMap
	}

	if _, exists := groupMap[name]; exists {
		return false, nil
	}

	groupMap[name] = &connection[C, T]{cfg: cfg}
	return true, nil
}

// Unregister 从组中注销指定资源。
//
// 如果资源已初始化（ready=true），会先调用 closer 关闭资源。
// 关闭时的错误会被忽略，资源仍会被移除。
//
// 返回值:
//   - ErrResourceNotFound: 资源不存在
//   - nil: 注销成功
func (g *group[C, T]) Unregister(ctx context.Context, name string) error {
	g.m.mu.Lock()
	defer g.m.mu.Unlock()

	groupMap, ok := g.m.groups[g.name]
	if !ok {
		return NewErrGroupNotFound(g.name)
	}

	conn, ok := groupMap[name]
	if !ok {
		return NewErrResourceNotFound(g.name, name)
	}

	if conn.ready && g.m.closer != nil {
		_ = g.m.closer(ctx, conn.val)
	}

	delete(groupMap, name)
	return nil
}

// List 返回组内所有已注册的资源名称列表。
//
// 返回的列表顺序不保证固定（依赖 map 遍历顺序）。
// 如果组不存在（已被关闭），返回空列表。
func (g *group[C, T]) List() []string {
	g.m.mu.RLock()
	defer g.m.mu.RUnlock()

	groupMap, ok := g.m.groups[g.name]
	if !ok {
		return nil
	}

	names := make([]string, 0, len(groupMap))
	for name := range groupMap {
		names = append(names, name)
	}
	return names
}

// Close 关闭组内所有已初始化的资源，并从管理器中移除整个组。
//
// 遍历组内所有资源，对已初始化（ready=true）的资源调用 closer 进行关闭。
// 关闭完成后，整个组将从管理器中删除。
//
// 返回值:
//   - []error: 关闭过程中遇到的所有错误，每个错误都包含组名和资源名信息
//   - nil: 组不存在（可能已被关闭）
func (g *group[C, T]) Close(ctx context.Context) []error {
	g.m.mu.Lock()
	defer g.m.mu.Unlock()

	groupMap, ok := g.m.groups[g.name]
	if !ok {
		return nil
	}

	var errs []error
	for name, conn := range groupMap {
		if !conn.ready {
			continue
		}
		if g.m.closer == nil {
			continue
		}
		if err := g.m.closer(ctx, conn.val); err != nil {
			err = NewErrCloseResourceFailed(g.name, name, err)
			errs = append(errs, err)
		}
	}

	delete(g.m.groups, g.name)
	return errs
}

// NewGroup 创建一个独立的资源组（单组模式）。
//
// 此函数是 New 的简化版本，适用于不需要多组管理的场景。
// 它会创建一个内部 manager 并预创建一个默认组，直接返回该组的引用。
//
// 使用场景:
//   - 应用只需要管理一类资源，不需要分组
//   - 快速原型开发，简化 API 调用
//
// 参数:
//   - opener: 资源打开器，用于根据配置创建资源实例
//   - closer: 资源关闭器，用于关闭/销毁资源（可以为 nil）
//
// 类型参数:
//   - C: 配置类型
//   - T: 资源类型
//
// 示例:
//
//	group := NewGroup(dbOpener, dbCloser)
//	group.Register(ctx, "main", dbConfig)
//	db, err := group.Get(ctx, "main")
func NewGroup[C any, T any](
	opener Opener[C, T],
	closer Closer[T],
) Group[C, T] {
	m := &manager[C, T]{
		groups: make(map[string]map[string]*connection[C, T]),
		opener: opener,
		closer: closer,
	}

	// 预创建默认 group，使用 defaultGroupName 作为组名
	m.groups[defaultGroupName] = make(map[string]*connection[C, T])
	return &group[C, T]{
		name: defaultGroupName,
		m:    m,
	}
}
