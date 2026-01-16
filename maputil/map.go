package maputil

// MapGet 从 map 中获取指定 key 对应的值。
// 返回值 v 表示对应的值，ok 表示 key 是否存在。
//
// 功能相当于：
//
//	v, ok := m[key]
//
// 但可以在泛型或函数式场景下直接使用。
func MapGet[K comparable, V any](m map[K]V, key K) (V, bool) {
	v, ok := m[key]
	return v, ok
}

// MapBy 根据给定的 key 和 value 提取函数，将切片转换为 map。
//
// 返回的 map 中，每个切片元素都会生成一条记录。
// 如果多个元素生成相同的 key，后面的元素会覆盖前面的值。
func MapBy[T any, K comparable, V any](list []T, key func(T) K, value func(T) V) map[K]V {
	m := make(map[K]V, 0)
	for _, v := range list {
		m[key(v)] = value(v)
	}
	return m
}
