// Package maputil 提供了一组泛型 map 操作工具函数。
package maputil

// MapGet 从 map 中安全地获取值，并支持可选的值转换。
//
// 参数:
//   - m: 源 map
//   - key: 要查找的键
//   - value: 值转换函数，用于将 map 中的原始值转换为目标类型；传入 nil 时返回零值
//
// 返回值:
//   - 第一个返回值为转换后的值，若 key 不存在或 value 为 nil 则返回零值
//   - 第二个返回值表示 key 是否存在于 map 中
//
// 示例:
//
//	users := map[int]User{1: {Name: "Alice"}}
//	name, ok := MapGet(users, 1, func(u User) string { return u.Name })
//	// name = "Alice", ok = true
func MapGet[T any, K comparable, V any](m map[K]T, key K, value func(T) V) (V, bool) {
	var zero V
	v, ok := m[key]
	if !ok {
		return zero, false
	}
	if value == nil {
		return zero, ok
	}
	return value(v), ok
}

// MapBy 将切片转换为 map，通过指定的函数分别提取键和值。
//
// 参数:
//   - list: 源切片
//   - key: 键提取函数，用于从切片元素中提取 map 的键
//   - value: 值提取函数，用于从切片元素中提取 map 的值
//
// 返回值:
//   - 由切片元素构建的 map
//
// 注意: 若多个元素产生相同的键，后者会覆盖前者。
//
// 示例:
//
//	users := []User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
//	m := MapBy(users, func(u User) int { return u.ID }, func(u User) string { return u.Name })
//	// m = map[int]string{1: "Alice", 2: "Bob"}
func MapBy[T any, K comparable, V any](list []T, key func(T) K, value func(T) V) map[K]V {
	m := make(map[K]V, len(list))
	for _, v := range list {
		m[key(v)] = value(v)
	}
	return m
}
