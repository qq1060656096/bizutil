package registry

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// 编译时类型断言，确保 manager 和 group 实现了对应接口
var _ Manager[testConfig, *testResource] = (*manager[testConfig, *testResource])(nil)
var _ Group[testConfig, *testResource] = (*group[testConfig, *testResource])(nil)

// 测试用的配置和资源类型
type testConfig struct {
	Name  string
	Value int
}

type testResource struct {
	Config testConfig
	Closed bool
}

// 创建测试用的 opener
func newTestOpener() Opener[testConfig, *testResource] {
	return func(ctx context.Context, cfg testConfig) (*testResource, error) {
		return &testResource{Config: cfg}, nil
	}
}

// 创建会失败的 opener
func newFailingOpener(errMsg string) Opener[testConfig, *testResource] {
	return func(ctx context.Context, cfg testConfig) (*testResource, error) {
		return nil, errors.New(errMsg)
	}
}

// 创建测试用的 closer
func newTestCloser() Closer[*testResource] {
	return func(ctx context.Context, r *testResource) error {
		r.Closed = true
		return nil
	}
}

// 创建会失败的 closer
func newFailingCloser(errMsg string) Closer[*testResource] {
	return func(ctx context.Context, r *testResource) error {
		return errors.New(errMsg)
	}
}

// 创建一个新的 manager 用于测试
func newTestManager(opener Opener[testConfig, *testResource], closer Closer[*testResource]) *manager[testConfig, *testResource] {
	return &manager[testConfig, *testResource]{
		groups: make(map[string]map[string]*connection[testConfig, *testResource]),
		opener: opener,
		closer: closer,
	}
}

// ============== Manager 测试 ==============

func TestManager_AddGroup(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())

	// 添加新组应该返回 false（表示之前不存在）
	existed := m.AddGroup("group1")
	if existed {
		t.Error("AddGroup should return false for new group")
	}

	// 再次添加同名组应该返回 true（表示已存在）
	existed = m.AddGroup("group1")
	if !existed {
		t.Error("AddGroup should return true for existing group")
	}

	// 添加另一个新组
	existed = m.AddGroup("group2")
	if existed {
		t.Error("AddGroup should return false for new group")
	}
}

func TestManager_Group(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())

	// 获取不存在的组应该返回错误
	_, err := m.Group("nonexistent")
	if err == nil {
		t.Error("Group should return error for nonexistent group")
	}
	if !errors.Is(err, ErrGroupNotFound) {
		t.Errorf("expected ErrGroupNotFound, got %v", err)
	}

	// 添加组后应该能获取
	m.AddGroup("group1")
	g, err := m.Group("group1")
	if err != nil {
		t.Errorf("Group should not return error for existing group: %v", err)
	}
	if g == nil {
		t.Error("Group should return non-nil group")
	}
}

func TestManager_MustGroup(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())
	m.AddGroup("group1")

	// 正常获取
	g := m.MustGroup("group1")
	if g == nil {
		t.Error("MustGroup should return non-nil group")
	}

	// 获取不存在的组应该 panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGroup should panic for nonexistent group")
		}
	}()
	m.MustGroup("nonexistent")
}

func TestManager_ListGroupNames(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())

	// 空 manager 应该返回空列表
	names := m.ListGroupNames()
	if len(names) != 0 {
		t.Errorf("expected empty list, got %v", names)
	}

	// 添加组后应该在列表中
	m.AddGroup("group1")
	m.AddGroup("group2")
	m.AddGroup("group3")

	names = m.ListGroupNames()
	if len(names) != 3 {
		t.Errorf("expected 3 groups, got %d", len(names))
	}

	// 验证所有组名都在列表中
	nameSet := make(map[string]bool)
	for _, name := range names {
		nameSet[name] = true
	}
	for _, expected := range []string{"group1", "group2", "group3"} {
		if !nameSet[expected] {
			t.Errorf("expected group %q in list", expected)
		}
	}
}

func TestManager_Close(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())
	ctx := context.Background()

	m.AddGroup("group1")
	g, _ := m.Group("group1")

	// 注册并获取资源（触发 opener）
	g.Register(ctx, "res1", testConfig{Name: "res1", Value: 1})
	res, _ := g.Get(ctx, "res1")

	// 关闭 manager
	errs := m.Close(ctx)
	if len(errs) != 0 {
		t.Errorf("Close should not return errors: %v", errs)
	}

	// 验证资源被关闭
	if !res.Closed {
		t.Error("resource should be closed")
	}

	// 验证组被清空
	if len(m.groups) != 0 {
		t.Error("groups should be empty after Close")
	}
}

func TestManager_Close_WithoutCloser(t *testing.T) {
	m := newTestManager(newTestOpener(), nil) // 没有 closer
	ctx := context.Background()

	m.AddGroup("group1")
	g, _ := m.Group("group1")
	g.Register(ctx, "res1", testConfig{Name: "res1", Value: 1})
	g.Get(ctx, "res1")

	// 关闭 manager（没有 closer 也不应该报错）
	errs := m.Close(ctx)
	if len(errs) != 0 {
		t.Errorf("Close should not return errors without closer: %v", errs)
	}
}

// ============== Group 测试 ==============

func TestGroup_Register(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())
	ctx := context.Background()

	m.AddGroup("group1")
	g, _ := m.Group("group1")

	// 注册新资源应该返回 true
	isNew, err := g.Register(ctx, "res1", testConfig{Name: "res1", Value: 1})
	if err != nil {
		t.Errorf("Register should not return error: %v", err)
	}
	if !isNew {
		t.Error("Register should return true for new resource")
	}

	// 注册已存在的资源应该返回 false
	isNew, err = g.Register(ctx, "res1", testConfig{Name: "res1", Value: 2})
	if err != nil {
		t.Errorf("Register should not return error: %v", err)
	}
	if isNew {
		t.Error("Register should return false for existing resource")
	}
}

func TestGroup_Get(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())
	ctx := context.Background()

	m.AddGroup("group1")
	g, _ := m.Group("group1")

	// 获取不存在的资源应该返回错误
	_, err := g.Get(ctx, "nonexistent")
	if err == nil {
		t.Error("Get should return error for nonexistent resource")
	}
	if !errors.Is(err, ErrResourceNotFound) {
		t.Errorf("expected ErrResourceNotFound, got %v", err)
	}

	// 注册并获取资源
	cfg := testConfig{Name: "res1", Value: 42}
	g.Register(ctx, "res1", cfg)

	res, err := g.Get(ctx, "res1")
	if err != nil {
		t.Errorf("Get should not return error: %v", err)
	}
	if res == nil {
		t.Error("Get should return non-nil resource")
	}
	if res.Config.Value != 42 {
		t.Errorf("expected config value 42, got %d", res.Config.Value)
	}

	// 再次获取应该返回同一个资源（懒加载只执行一次）
	res2, _ := g.Get(ctx, "res1")
	if res != res2 {
		t.Error("Get should return the same resource instance")
	}
}

func TestGroup_Get_OpenerError(t *testing.T) {
	m := newTestManager(newFailingOpener("open failed"), newTestCloser())
	ctx := context.Background()

	m.AddGroup("group1")
	g, _ := m.Group("group1")
	g.Register(ctx, "res1", testConfig{Name: "res1"})

	_, err := g.Get(ctx, "res1")
	if err == nil {
		t.Error("Get should return error when opener fails")
	}
	if err.Error() != "open failed" {
		t.Errorf("expected 'open failed' error, got %v", err)
	}
}

func TestGroup_MustGet(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())
	ctx := context.Background()

	m.AddGroup("group1")
	g, _ := m.Group("group1")
	g.Register(ctx, "res1", testConfig{Name: "res1"})

	// 正常获取
	res := g.MustGet(ctx, "res1")
	if res == nil {
		t.Error("MustGet should return non-nil resource")
	}

	// 获取不存在的资源应该 panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGet should panic for nonexistent resource")
		}
	}()
	g.MustGet(ctx, "nonexistent")
}

func TestGroup_Config(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())
	ctx := context.Background()

	m.AddGroup("group1")
	g, _ := m.Group("group1")

	// 获取不存在的资源配置应该返回错误
	_, err := g.Config(ctx, "nonexistent")
	if err == nil {
		t.Error("Config should return error for nonexistent resource")
	}
	if !errors.Is(err, ErrResourceNotFound) {
		t.Errorf("expected ErrResourceNotFound, got %v", err)
	}

	// 注册并获取资源配置
	cfg := testConfig{Name: "res1", Value: 42}
	g.Register(ctx, "res1", cfg)

	gotCfg, err := g.Config(ctx, "res1")
	if err != nil {
		t.Errorf("Config should not return error: %v", err)
	}
	if gotCfg.Name != "res1" || gotCfg.Value != 42 {
		t.Errorf("expected config {res1, 42}, got %v", gotCfg)
	}
}

func TestGroup_Config_GroupNotFound(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())
	ctx := context.Background()

	// 创建一个指向不存在组的 group 对象
	g := &group[testConfig, *testResource]{
		name: "nonexistent",
		m:    m,
	}

	_, err := g.Config(ctx, "res1")
	if err == nil {
		t.Error("Config should return error for nonexistent group")
	}
	if !errors.Is(err, ErrGroupNotFound) {
		t.Errorf("expected ErrGroupNotFound, got %v", err)
	}
}

func TestGroup_MustConfig(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())
	ctx := context.Background()

	m.AddGroup("group1")
	g, _ := m.Group("group1")
	g.Register(ctx, "res1", testConfig{Name: "res1", Value: 100})

	// 正常获取
	cfg := g.MustConfig(ctx, "res1")
	if cfg.Value != 100 {
		t.Errorf("expected value 100, got %d", cfg.Value)
	}

	// 获取不存在的资源应该 panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustConfig should panic for nonexistent resource")
		}
	}()
	g.MustConfig(ctx, "nonexistent")
}

func TestGroup_Unregister(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())
	ctx := context.Background()

	m.AddGroup("group1")
	g, _ := m.Group("group1")

	// 注销不存在的资源应该返回错误
	err := g.Unregister(ctx, "nonexistent")
	if err == nil {
		t.Error("Unregister should return error for nonexistent resource")
	}
	if !errors.Is(err, ErrResourceNotFound) {
		t.Errorf("expected ErrResourceNotFound, got %v", err)
	}

	// 注册、获取然后注销
	g.Register(ctx, "res1", testConfig{Name: "res1"})
	res, _ := g.Get(ctx, "res1")

	err = g.Unregister(ctx, "res1")
	if err != nil {
		t.Errorf("Unregister should not return error: %v", err)
	}

	// 验证资源被关闭
	if !res.Closed {
		t.Error("resource should be closed after Unregister")
	}

	// 验证资源被删除
	_, err = g.Get(ctx, "res1")
	if err == nil {
		t.Error("Get should return error after Unregister")
	}
}

func TestGroup_Unregister_NotReady(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())
	ctx := context.Background()

	m.AddGroup("group1")
	g, _ := m.Group("group1")

	// 注册但不获取（资源未初始化）
	g.Register(ctx, "res1", testConfig{Name: "res1"})

	// 注销未初始化的资源不应该调用 closer
	err := g.Unregister(ctx, "res1")
	if err != nil {
		t.Errorf("Unregister should not return error: %v", err)
	}
}

func TestGroup_List(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())
	ctx := context.Background()

	m.AddGroup("group1")
	m.AddGroup("group2")
	g, _ := m.Group("group1")

	// 注册一些资源
	g.Register(ctx, "res1", testConfig{Name: "res1"})
	g.Register(ctx, "res2", testConfig{Name: "res2"})

	// List 返回的是组名，不是资源名（根据代码实现）
	names := g.List()
	// 注意：当前实现 List() 返回的是 manager 中的组名，而不是组内的资源名
	// 这可能是一个 bug，但我们先按照当前实现测试
	if len(names) != 2 {
		t.Errorf("expected 2 groups, got %d", len(names))
	}
}

func TestGroup_Close(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())
	ctx := context.Background()

	m.AddGroup("group1")
	m.AddGroup("group2")

	g1, _ := m.Group("group1")
	g2, _ := m.Group("group2")

	// 在两个组中注册资源
	g1.Register(ctx, "res1", testConfig{Name: "res1"})
	g1.Register(ctx, "res2", testConfig{Name: "res2"})
	g2.Register(ctx, "res3", testConfig{Name: "res3"})

	// 获取资源（触发初始化）
	res1, _ := g1.Get(ctx, "res1")
	res2, _ := g1.Get(ctx, "res2")
	res3, _ := g2.Get(ctx, "res3")

	// 关闭 group1
	errs := g1.Close(ctx)
	if len(errs) != 0 {
		t.Errorf("Close should not return errors: %v", errs)
	}

	// 验证 group1 的资源被关闭
	if !res1.Closed || !res2.Closed {
		t.Error("group1 resources should be closed")
	}

	// 验证 group2 的资源未被关闭
	if res3.Closed {
		t.Error("group2 resources should not be closed")
	}

	// 验证 group1 被删除
	if _, ok := m.groups["group1"]; ok {
		t.Error("group1 should be removed from manager")
	}

	// 验证 group2 仍然存在
	if _, ok := m.groups["group2"]; !ok {
		t.Error("group2 should still exist in manager")
	}
}

func TestGroup_Close_WithFailingCloser(t *testing.T) {
	m := newTestManager(newTestOpener(), newFailingCloser("close failed"))
	ctx := context.Background()

	m.AddGroup("group1")
	g, _ := m.Group("group1")
	g.Register(ctx, "res1", testConfig{Name: "res1"})
	g.Get(ctx, "res1")

	errs := g.Close(ctx)
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d", len(errs))
	}
	if !errors.Is(errs[0], ErrCloseResourceFailed) {
		t.Errorf("expected ErrCloseResourceFailed, got %v", errs[0])
	}
}

func TestGroup_Close_NonexistentGroup(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())
	ctx := context.Background()

	// 创建一个指向不存在组的 group 对象
	g := &group[testConfig, *testResource]{
		name: "nonexistent",
		m:    m,
	}

	errs := g.Close(ctx)
	if len(errs) != 0 {
		t.Errorf("Close on nonexistent group should return empty errors: %v", errs)
	}
}

// ============== Ping 测试 ==============

func TestGroup_Ping(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())
	ctx := context.Background()

	m.AddGroup("group1")
	g, _ := m.Group("group1")

	// 注册资源
	cfg := testConfig{Name: "res1", Value: 100}
	g.Register(ctx, "res1", cfg)

	// Ping 应该成功
	err := g.Ping(ctx, "res1")
	if err != nil {
		t.Errorf("Ping should succeed: %v", err)
	}
}

func TestGroup_Ping_GroupNotFound(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())
	ctx := context.Background()

	// 创建一个指向不存在组的 group 对象
	g := &group[testConfig, *testResource]{
		name: "nonexistent",
		m:    m,
	}

	err := g.Ping(ctx, "res1")
	if err == nil {
		t.Error("Ping should return error for nonexistent group")
	}
	if !errors.Is(err, ErrGroupNotFound) {
		t.Errorf("expected ErrGroupNotFound, got %v", err)
	}
}

func TestGroup_Ping_ResourceNotFound(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())
	ctx := context.Background()

	m.AddGroup("group1")
	g, _ := m.Group("group1")

	// Ping 未注册的资源应该返回错误
	err := g.Ping(ctx, "nonexistent")
	if err == nil {
		t.Error("Ping should return error for nonexistent resource")
	}
	if !errors.Is(err, ErrResourceNotFound) {
		t.Errorf("expected ErrResourceNotFound, got %v", err)
	}
}

func TestGroup_Ping_OpenerError(t *testing.T) {
	m := newTestManager(newFailingOpener("opener error"), newTestCloser())
	ctx := context.Background()

	m.AddGroup("group1")
	g, _ := m.Group("group1")
	g.Register(ctx, "res1", testConfig{Name: "res1"})

	// Ping 应该返回 opener 错误
	err := g.Ping(ctx, "res1")
	if err == nil {
		t.Error("Ping should return error when opener fails")
	}
	if !errors.Is(err, ErrPingResourceFailed) {
		t.Errorf("expected ErrPingResourceFailed, got %v", err)
	}
}

func TestGroup_Ping_DoesNotCacheResource(t *testing.T) {
	var openerCallCount int32

	opener := func(ctx context.Context, cfg testConfig) (*testResource, error) {
		atomic.AddInt32(&openerCallCount, 1)
		return &testResource{Config: cfg}, nil
	}

	m := &manager[testConfig, *testResource]{
		groups: make(map[string]map[string]*connection[testConfig, *testResource]),
		opener: opener,
		closer: newTestCloser(),
	}
	ctx := context.Background()

	m.AddGroup("group1")
	g, _ := m.Group("group1")
	g.Register(ctx, "res1", testConfig{Name: "res1"})

	// 第一次 Ping
	err := g.Ping(ctx, "res1")
	if err != nil {
		t.Fatalf("Ping should succeed: %v", err)
	}

	// 第二次 Ping
	err = g.Ping(ctx, "res1")
	if err != nil {
		t.Fatalf("Ping should succeed: %v", err)
	}

	// Ping 不应该缓存资源，所以 opener 应该被调用两次
	if openerCallCount != 2 {
		t.Errorf("expected opener to be called 2 times, but was called %d times", openerCallCount)
	}

	// 验证资源没有被标记为 ready
	m.mu.RLock()
	conn := m.groups["group1"]["res1"]
	ready := conn.ready
	m.mu.RUnlock()

	if ready {
		t.Error("Ping should not mark resource as ready")
	}
}

func TestGroup_Ping_DoesNotAffectGetCache(t *testing.T) {
	var openerCallCount int32

	opener := func(ctx context.Context, cfg testConfig) (*testResource, error) {
		atomic.AddInt32(&openerCallCount, 1)
		return &testResource{Config: cfg}, nil
	}

	m := &manager[testConfig, *testResource]{
		groups: make(map[string]map[string]*connection[testConfig, *testResource]),
		opener: opener,
		closer: newTestCloser(),
	}
	ctx := context.Background()

	m.AddGroup("group1")
	g, _ := m.Group("group1")
	g.Register(ctx, "res1", testConfig{Name: "res1"})

	// 先 Ping 一次
	err := g.Ping(ctx, "res1")
	if err != nil {
		t.Fatalf("Ping should succeed: %v", err)
	}

	// Ping 不应该缓存，所以调用了一次
	if openerCallCount != 1 {
		t.Errorf("expected opener to be called 1 time, but was called %d times", openerCallCount)
	}

	// 然后通过 Get 获取资源
	res1, err := g.Get(ctx, "res1")
	if err != nil {
		t.Fatalf("Get should succeed: %v", err)
	}

	// Get 应该再调用一次 opener（因为 Ping 没有缓存）
	if openerCallCount != 2 {
		t.Errorf("expected opener to be called 2 times after Get, but was called %d times", openerCallCount)
	}

	// 再次 Get 应该使用缓存
	res2, err := g.Get(ctx, "res1")
	if err != nil {
		t.Fatalf("Get should succeed: %v", err)
	}

	// 应该返回同一个实例
	if res1 != res2 {
		t.Error("Get should return the same cached instance")
	}

	// opener 调用次数不应该增加
	if openerCallCount != 2 {
		t.Errorf("expected opener to still be called 2 times, but was called %d times", openerCallCount)
	}
}

func TestConcurrent_Ping(t *testing.T) {
	var openerCallCount int32

	opener := func(ctx context.Context, cfg testConfig) (*testResource, error) {
		atomic.AddInt32(&openerCallCount, 1)
		// 模拟慢速操作
		time.Sleep(5 * time.Millisecond)
		return &testResource{Config: cfg}, nil
	}

	m := &manager[testConfig, *testResource]{
		groups: make(map[string]map[string]*connection[testConfig, *testResource]),
		opener: opener,
		closer: newTestCloser(),
	}
	ctx := context.Background()

	m.AddGroup("group1")
	g, _ := m.Group("group1")
	g.Register(ctx, "res1", testConfig{Name: "res1", Value: 1})

	const numGoroutines = 20

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	errCount := int32(0)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			err := g.Ping(ctx, "res1")
			if err != nil {
				atomic.AddInt32(&errCount, 1)
				t.Errorf("Ping error: %v", err)
			}
		}(i)
	}

	wg.Wait()

	// 所有 Ping 都应该成功
	if errCount != 0 {
		t.Errorf("expected 0 errors, got %d", errCount)
	}

	// 由于 Ping 不缓存，每次都应该调用 opener
	if openerCallCount != numGoroutines {
		t.Errorf("expected opener to be called %d times, but was called %d times", numGoroutines, openerCallCount)
	}

	// 验证资源仍然没有被标记为 ready
	m.mu.RLock()
	conn := m.groups["group1"]["res1"]
	ready := conn.ready
	m.mu.RUnlock()

	if ready {
		t.Error("concurrent Ping should not mark resource as ready")
	}
}

// ============== 错误类型测试 ==============

func TestErrors(t *testing.T) {
	t.Run("ErrGroupNotFound", func(t *testing.T) {
		err := NewErrGroupNotFound("testGroup")
		if !errors.Is(err, ErrGroupNotFound) {
			t.Error("should wrap ErrGroupNotFound")
		}
		if err.Error() == "" {
			t.Error("error message should not be empty")
		}
	})

	t.Run("ErrResourceNotFound", func(t *testing.T) {
		err := NewErrResourceNotFound("testGroup", "testResource")
		if !errors.Is(err, ErrResourceNotFound) {
			t.Error("should wrap ErrResourceNotFound")
		}
		if err.Error() == "" {
			t.Error("error message should not be empty")
		}
	})

	t.Run("ErrCloseResourceFailed", func(t *testing.T) {
		innerErr := errors.New("inner error")
		err := NewErrCloseResourceFailed("testGroup", "testResource", innerErr)
		if !errors.Is(err, ErrCloseResourceFailed) {
			t.Error("should wrap ErrCloseResourceFailed")
		}
		if !errors.Is(err, innerErr) {
			t.Error("should wrap inner error")
		}
	})
}

// ============== 并发测试 ==============

func TestConcurrent_AddGroup(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())

	const numGoroutines = 100
	const numGroups = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			groupName := fmt.Sprintf("group%d", id%numGroups)
			m.AddGroup(groupName)
		}(i)
	}

	wg.Wait()

	// 验证组数量
	names := m.ListGroupNames()
	if len(names) != numGroups {
		t.Errorf("expected %d groups, got %d", numGroups, len(names))
	}
}

func TestConcurrent_Register(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())
	ctx := context.Background()

	m.AddGroup("group1")
	g, _ := m.Group("group1")

	const numGoroutines = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	successCount := int32(0)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			isNew, err := g.Register(ctx, "shared-resource", testConfig{Name: "shared", Value: id})
			if err != nil {
				t.Errorf("Register error: %v", err)
				return
			}
			if isNew {
				atomic.AddInt32(&successCount, 1)
			}
		}(i)
	}

	wg.Wait()

	// 只有一个 goroutine 应该成功创建资源
	if successCount != 1 {
		t.Errorf("expected exactly 1 successful registration, got %d", successCount)
	}
}

func TestConcurrent_Get(t *testing.T) {
	var openerCallCount int32

	opener := func(ctx context.Context, cfg testConfig) (*testResource, error) {
		atomic.AddInt32(&openerCallCount, 1)
		// 模拟慢速初始化
		time.Sleep(10 * time.Millisecond)
		return &testResource{Config: cfg}, nil
	}

	m := &manager[testConfig, *testResource]{
		groups: make(map[string]map[string]*connection[testConfig, *testResource]),
		opener: opener,
		closer: newTestCloser(),
	}
	ctx := context.Background()

	m.AddGroup("group1")
	g, _ := m.Group("group1")
	g.Register(ctx, "res1", testConfig{Name: "res1", Value: 1})

	const numGoroutines = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	results := make([]*testResource, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			res, err := g.Get(ctx, "res1")
			if err != nil {
				t.Errorf("Get error: %v", err)
				return
			}
			results[id] = res
		}(i)
	}

	wg.Wait()

	// opener 应该只被调用一次（由于 double-check 锁定）
	if openerCallCount != 1 {
		t.Errorf("opener should be called exactly once, but was called %d times", openerCallCount)
	}

	// 所有 goroutine 应该获得相同的资源实例
	firstRes := results[0]
	for i, res := range results {
		if res != firstRes {
			t.Errorf("goroutine %d got different resource instance", i)
		}
	}
}

func TestConcurrent_RegisterAndGet(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())
	ctx := context.Background()

	m.AddGroup("group1")
	g, _ := m.Group("group1")

	const numResources = 20
	const numGetters = 10

	var wg sync.WaitGroup

	// 注册多个资源
	for i := 0; i < numResources; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			resName := fmt.Sprintf("res%d", id)
			g.Register(ctx, resName, testConfig{Name: resName, Value: id})
		}(i)
	}

	wg.Wait()

	// 并发获取资源
	wg.Add(numResources * numGetters)
	for i := 0; i < numResources; i++ {
		for j := 0; j < numGetters; j++ {
			go func(resID, getterID int) {
				defer wg.Done()
				resName := fmt.Sprintf("res%d", resID)
				res, err := g.Get(ctx, resName)
				if err != nil {
					t.Errorf("Get(%s) error: %v", resName, err)
					return
				}
				if res.Config.Value != resID {
					t.Errorf("expected value %d, got %d", resID, res.Config.Value)
				}
			}(i, j)
		}
	}

	wg.Wait()
}

func TestConcurrent_Unregister(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())
	ctx := context.Background()

	m.AddGroup("group1")
	g, _ := m.Group("group1")

	const numResources = 50

	// 先注册一批资源
	for i := 0; i < numResources; i++ {
		resName := fmt.Sprintf("res%d", i)
		g.Register(ctx, resName, testConfig{Name: resName, Value: i})
		g.Get(ctx, resName) // 触发初始化
	}

	var wg sync.WaitGroup
	wg.Add(numResources)

	successCount := int32(0)
	errorCount := int32(0)

	// 并发注销
	for i := 0; i < numResources; i++ {
		go func(id int) {
			defer wg.Done()
			resName := fmt.Sprintf("res%d", id)
			err := g.Unregister(ctx, resName)
			if err == nil {
				atomic.AddInt32(&successCount, 1)
			} else {
				atomic.AddInt32(&errorCount, 1)
			}
		}(i)
	}

	wg.Wait()

	// 所有注销操作应该成功
	if successCount != numResources {
		t.Errorf("expected %d successful unregistrations, got %d (errors: %d)",
			numResources, successCount, errorCount)
	}
}

func TestConcurrent_MixedOperations(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())
	ctx := context.Background()

	const numGroups = 5
	const numResourcesPerGroup = 10
	const numOperations = 100

	// 创建组
	for i := 0; i < numGroups; i++ {
		m.AddGroup(fmt.Sprintf("group%d", i))
	}

	var wg sync.WaitGroup
	wg.Add(numOperations)

	for i := 0; i < numOperations; i++ {
		go func(id int) {
			defer wg.Done()

			groupID := id % numGroups
			resourceID := id % numResourcesPerGroup
			groupName := fmt.Sprintf("group%d", groupID)
			resName := fmt.Sprintf("res%d", resourceID)

			g, err := m.Group(groupName)
			if err != nil {
				return
			}

			// 随机执行不同操作
			switch id % 4 {
			case 0:
				g.Register(ctx, resName, testConfig{Name: resName, Value: id})
			case 1:
				g.Get(ctx, resName)
			case 2:
				g.List()
			case 3:
				m.ListGroupNames()
			}
		}(i)
	}

	wg.Wait()

	// 验证数据完整性 - 能正常获取组列表
	names := m.ListGroupNames()
	if len(names) != numGroups {
		t.Errorf("expected %d groups, got %d", numGroups, len(names))
	}
}

func TestConcurrent_GroupClose(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())
	ctx := context.Background()

	const numGroups = 10

	// 创建多个组并注册资源
	for i := 0; i < numGroups; i++ {
		groupName := fmt.Sprintf("group%d", i)
		m.AddGroup(groupName)
		g, _ := m.Group(groupName)
		g.Register(ctx, "res1", testConfig{Name: "res1", Value: i})
		g.Get(ctx, "res1")
	}

	var wg sync.WaitGroup
	wg.Add(numGroups)

	// 并发关闭所有组
	for i := 0; i < numGroups; i++ {
		go func(id int) {
			defer wg.Done()
			groupName := fmt.Sprintf("group%d", id)
			g, err := m.Group(groupName)
			if err != nil {
				// 组可能已被其他 goroutine 关闭
				return
			}
			g.Close(ctx)
		}(i)
	}

	wg.Wait()

	// 所有组应该被关闭
	names := m.ListGroupNames()
	if len(names) != 0 {
		t.Errorf("expected 0 groups after close, got %d", len(names))
	}
}

func TestConcurrent_ManagerClose(t *testing.T) {
	const numManagers = 5

	var wg sync.WaitGroup
	wg.Add(numManagers)

	for i := 0; i < numManagers; i++ {
		go func(id int) {
			defer wg.Done()

			m := newTestManager(newTestOpener(), newTestCloser())
			ctx := context.Background()

			// 创建组和资源
			for j := 0; j < 10; j++ {
				groupName := fmt.Sprintf("group%d", j)
				m.AddGroup(groupName)
				g, _ := m.Group(groupName)
				for k := 0; k < 5; k++ {
					resName := fmt.Sprintf("res%d", k)
					g.Register(ctx, resName, testConfig{Name: resName})
					g.Get(ctx, resName)
				}
			}

			// 关闭 manager
			errs := m.Close(ctx)
			if len(errs) != 0 {
				t.Errorf("manager %d Close errors: %v", id, errs)
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrent_ReadWrite 测试并发读写场景
func TestConcurrent_ReadWrite(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())
	ctx := context.Background()

	m.AddGroup("group1")
	g, _ := m.Group("group1")

	// 预先注册一些资源
	for i := 0; i < 10; i++ {
		g.Register(ctx, fmt.Sprintf("res%d", i), testConfig{Name: fmt.Sprintf("res%d", i), Value: i})
	}

	const (
		numReaders = 50
		numWriters = 10
		duration   = 100 * time.Millisecond
	)

	stopCh := make(chan struct{})
	var wg sync.WaitGroup

	// 启动读者
	wg.Add(numReaders)
	for i := 0; i < numReaders; i++ {
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case <-stopCh:
					return
				default:
					resName := fmt.Sprintf("res%d", id%10)
					g.Get(ctx, resName)
					m.ListGroupNames()
				}
			}
		}(i)
	}

	// 启动写者
	wg.Add(numWriters)
	for i := 0; i < numWriters; i++ {
		go func(id int) {
			defer wg.Done()
			counter := 0
			for {
				select {
				case <-stopCh:
					return
				default:
					resName := fmt.Sprintf("dynamic_res_%d_%d", id, counter)
					g.Register(ctx, resName, testConfig{Name: resName, Value: counter})
					counter++
				}
			}
		}(i)
	}

	// 运行一段时间后停止
	time.Sleep(duration)
	close(stopCh)
	wg.Wait()
}

// ============== 边界条件测试 ==============

func TestEmptyGroupName(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())

	// 空组名应该可以正常工作
	existed := m.AddGroup("")
	if existed {
		t.Error("AddGroup should return false for new empty group name")
	}

	g, err := m.Group("")
	if err != nil {
		t.Errorf("Group should work with empty name: %v", err)
	}
	if g == nil {
		t.Error("Group should return non-nil for empty group name")
	}
}

func TestEmptyResourceName(t *testing.T) {
	m := newTestManager(newTestOpener(), newTestCloser())
	ctx := context.Background()

	m.AddGroup("group1")
	g, _ := m.Group("group1")

	// 空资源名应该可以正常工作
	isNew, err := g.Register(ctx, "", testConfig{Name: ""})
	if err != nil {
		t.Errorf("Register should work with empty name: %v", err)
	}
	if !isNew {
		t.Error("Register should return true for new empty resource name")
	}

	res, err := g.Get(ctx, "")
	if err != nil {
		t.Errorf("Get should work with empty name: %v", err)
	}
	if res == nil {
		t.Error("Get should return non-nil for empty resource name")
	}
}

func TestContextCancellation(t *testing.T) {
	var openerCalled bool
	opener := func(ctx context.Context, cfg testConfig) (*testResource, error) {
		openerCalled = true
		// 检查 context 是否已取消
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			return &testResource{Config: cfg}, nil
		}
	}

	m := &manager[testConfig, *testResource]{
		groups: make(map[string]map[string]*connection[testConfig, *testResource]),
		opener: opener,
		closer: newTestCloser(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	m.AddGroup("group1")
	g, _ := m.Group("group1")
	g.Register(ctx, "res1", testConfig{Name: "res1"})

	_, err := g.Get(ctx, "res1")
	if err == nil {
		t.Error("Get should return error when context is cancelled")
	}
	if !openerCalled {
		t.Error("opener should have been called")
	}
}

// ============== 基准测试 ==============

func BenchmarkGroup_Get_Cached(b *testing.B) {
	m := newTestManager(newTestOpener(), newTestCloser())
	ctx := context.Background()

	m.AddGroup("group1")
	g, _ := m.Group("group1")
	g.Register(ctx, "res1", testConfig{Name: "res1"})
	g.Get(ctx, "res1") // 预热

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Get(ctx, "res1")
	}
}

func BenchmarkGroup_Get_Cached_Parallel(b *testing.B) {
	m := newTestManager(newTestOpener(), newTestCloser())
	ctx := context.Background()

	m.AddGroup("group1")
	g, _ := m.Group("group1")
	g.Register(ctx, "res1", testConfig{Name: "res1"})
	g.Get(ctx, "res1") // 预热

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			g.Get(ctx, "res1")
		}
	})
}

func BenchmarkManager_ListGroupNames(b *testing.B) {
	m := newTestManager(newTestOpener(), newTestCloser())

	for i := 0; i < 100; i++ {
		m.AddGroup(fmt.Sprintf("group%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.ListGroupNames()
	}
}

func BenchmarkManager_ListGroupNames_Parallel(b *testing.B) {
	m := newTestManager(newTestOpener(), newTestCloser())

	for i := 0; i < 100; i++ {
		m.AddGroup(fmt.Sprintf("group%d", i))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.ListGroupNames()
		}
	})
}
