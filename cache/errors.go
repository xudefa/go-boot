// Package cache 定义缓存相关的错误 sentinel。
//
// 提供缓存操作的标准错误值：
//   - ErrNotFound: 缓存键不存在
//   - ErrCacheMiss: 缓存键过期或未命中
package cache

import "errors"

var (
	ErrNotFound  = errors.New("cache: key not found")
	ErrCacheMiss = errors.New("cache: key expired or not found")
)
