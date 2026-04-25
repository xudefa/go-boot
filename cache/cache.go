// Package cache 提供通用缓存接口及错误类型定义.
// 多种实现(memory、redis等)可以满足Cache接口.
package cache

import (
	"context"
	"errors"
	"time"
)

// 缓存实现返回的通用错误.
var (
	// ErrCacheMiss 表示键存在但已过期.
	ErrCacheMiss = errors.New("cache miss")
	// ErrNotFound 表示键在缓存中不存在.
	ErrNotFound = errors.New("key not found")
)

// Cache 定义通用缓存接口.
// 实现应该是并发安全的.
type Cache interface {
	// Get 根据键获取值.
	// 如果键不存在返回ErrNotFound.
	// 如果键存在但已过期返回ErrCacheMiss.
	Get(ctx context.Context, key string) (any, error)

	// Set 存储键值对,可指定TTL.
	// 如果ttl<=0,则永不过期.
	Set(ctx context.Context, key string, value any, ttl time.Duration) error

	// Delete 从缓存中删除键.
	Delete(ctx context.Context, key string) error

	// Exists 判断所有给定键是否存在于缓存中.
	// 如果任意键不存在或已过期,返回false.
	Exists(ctx context.Context, keys ...string) (bool, error)

	// GetMulti 批量获取给定键的值.
	// 返回存在的键值对映射.
	// 已过期的键视为不存在.
	GetMulti(ctx context.Context, keys []string) (map[string]any, error)

	// SetMulti 批量存储键值对,使用相同的TTL.
	// 如果ttl<=0,则永不过期.
	SetMulti(ctx context.Context, items map[string]any, ttl time.Duration) error

	// DeleteMulti 批量删除给定键.
	DeleteMulti(ctx context.Context, keys []string) error

	// Clear 清空缓存.
	Clear(ctx context.Context) error
}

// Getter 是从数据源获取值的函数类型.
// 用于GetWithGetter在缓存未命中时填充缓存.
type Getter func(ctx context.Context, key string) (any, error)

// CacheWithGetter 扩展Cache,提供GetWithGetter方法.
// 该方法从缓存获取值,如果未找到或已过期,
// 调用Getter函数获取值并存储到缓存中.
type CacheWithGetter interface {
	Cache
	GetWithGetter(ctx context.Context, key string, fn Getter) (any, error)
}
