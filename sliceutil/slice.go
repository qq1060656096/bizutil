package sliceutil

import (
	"errors"
	"fmt"
)

// FindDuplicateWithErr 查找切片中的重复元素，支持自定义错误生成函数
// 泛型参数：
//
//	T: 切片中元素的类型
//	K: 用于比较的键类型，必须是可比较的
//
// 参数：
//
//	list: 要检查的切片
//	keyFn: 从元素中提取比较键的函数
//	errFn: 自定义错误生成函数，可以为 nil
//
// 返回值：
//
//	error: 如果发现重复元素则返回错误，否则返回 nil
func FindDuplicateWithErr[T any, K comparable](list []T, keyFn func(T) K, errFn func(T) error) error {
	// 创建 map 用于记录已出现的键，初始容量为切片长度以提高性能
	m := make(map[K]struct{}, len(list))

	// 遍历切片中的每个元素
	for _, v := range list {
		// 提取当前元素的比较键
		k := keyFn(v)

		// 检查该键是否已经存在于 map 中
		if _, exists := m[k]; exists {
			// 如果存在重复，使用自定义错误函数生成错误
			if errFn != nil {
				return errFn(v)
			}
			// 如果没有提供自定义错误函数，使用默认错误格式
			return errors.New(fmt.Sprintf("重复值: %+v", v))
		}

		// 将当前键标记为已出现
		m[k] = struct{}{}
	}

	// 遍历完成未发现重复，返回 nil
	return nil
}
