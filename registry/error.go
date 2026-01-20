package registry

import (
	"errors"
	"fmt"
)

// 预定义的哨兵错误，可使用 errors.Is 进行判断。
//
// 示例:
//
//	_, err := mgr.Group("nonexistent")
//	if errors.Is(err, registry.ErrGroupNotFound) {
//	    // 处理组不存在的情况
//	}
var (
	// ErrGroupNotFound 表示请求的资源组不存在。
	// 当调用 Manager.Group 或 Group.Get 时，如果指定的组未被添加，将返回此错误。
	ErrGroupNotFound = errors.New("bizutil.registry: group not found")

	// ErrResourceNotFound 表示请求的资源在组中不存在。
	// 当调用 Group.Get 或 Group.Unregister 时，如果指定的资源未被注册，将返回此错误。
	ErrResourceNotFound = errors.New("bizutil.registry: resource not found")

	// ErrCloseResourceFailed 表示关闭资源时发生错误。
	// 当 Closer 函数返回错误时，将返回此错误。
	ErrCloseResourceFailed = errors.New("bizutil.registry: close resource failed")

	// ErrPingResourceFailed
	ErrPingResourceFailed = errors.New("bizutil.registry: ping resource failed")
)

// NewErrGroupNotFound 创建一个包含组名信息的组未找到错误。
//
// 返回的错误可以通过 errors.Is(err, ErrGroupNotFound) 进行判断。
func NewErrGroupNotFound(groupName string) error {
	return fmt.Errorf("group %q not found: %w", groupName, ErrGroupNotFound)
}

// NewErrResourceNotFound 创建一个包含组名和资源名信息的资源未找到错误。
//
// 返回的错误可以通过 errors.Is(err, ErrResourceNotFound) 进行判断。
func NewErrResourceNotFound(groupName, resourceName string) error {
	return fmt.Errorf("resource %q not found from group %q: %w", resourceName, groupName, ErrResourceNotFound)
}

// NewErrCloseResourceFailed 创建一个包含组名、资源名和原始错误的关闭失败错误。
//
// 返回的错误可以通过 errors.Is(err, ErrCloseResourceFailed) 进行判断，
// 同时也可以通过 errors.Is 判断原始错误。
func NewErrCloseResourceFailed(groupName, resourceName string, err error) error {
	return fmt.Errorf("close resource %q in group %q failed: %w: %w", resourceName, groupName, ErrCloseResourceFailed, err)
}

func NewErrPingResourceFailed(groupName, resourceName string, err error) error {
	return fmt.Errorf("ping resource %q in group %q failed: %w", resourceName, groupName, ErrPingResourceFailed)
}
