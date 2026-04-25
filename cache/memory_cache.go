// Package cache 提供内存缓存实现,支持TTL.
// 并发安全.
package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

// cacheItem 表示单个缓存项,可选择过期时间.
type cacheItem struct {
	// value 是缓存的数据.
	value any
	// expireAt 是过期时间.
	// 零时间表示永不过期.
	expireAt time.Time
}

// MemoryCache 内存缓存实现.
// 使用RWMutex实现高效的并发读写.
// 过期项在访问时延迟删除.
type MemoryCache struct {
	// mu 保护data映射的访问.
	mu sync.RWMutex
	// data 保存缓存项.
	data map[string]cacheItem
}

// NewMemoryCache 创建并返回新的MemoryCache实例.
// 缓存初始为空.
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		data: make(map[string]cacheItem),
	}
}

// Get 根据键从缓存获取值.
// 如果键不存在返回ErrNotFound.
// 如果键存在但已过期返回ErrCacheMiss.
// 因过期导致的缓存未命中,会从缓存中删除该项.
func (c *MemoryCache) Get(ctx context.Context, key string) (any, error) {
	// 使用读锁实现并发访问.
	c.mu.RLock()
	item, ok := c.data[key]
	c.mu.RUnlock()

	if !ok {
		return nil, ErrNotFound
	}
	if item.expireAt.IsZero() {
		return item.value, nil
	}
	if time.Now().After(item.expireAt) {
		// 项已过期,使用写锁删除.
		c.mu.Lock()
		delete(c.data, key)
		c.mu.Unlock()
		return nil, ErrCacheMiss
	}
	return item.value, nil
}

// Set 将值存入缓存,可指定TTL.
// 如果ttl<=0,则永不过期.
// 使用写锁保护并发映射访问.
func (c *MemoryCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	expireAt := time.Time{}
	if ttl > 0 {
		expireAt = time.Now().Add(ttl)
	}
	c.data[key] = cacheItem{
		value:    value,
		expireAt: expireAt,
	}
	return nil
}

// Delete 从缓存中删除键.
// 使用写锁保护并发映射访问.
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
	return nil
}

// Exists 判断所有给定键是否存在于缓存中.
// 如果任意键不存在或已过期,返回false.
// 使用读锁实现并发访问.
func (c *MemoryCache) Exists(ctx context.Context, keys ...string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, key := range keys {
		item, ok := c.data[key]
		if !ok {
			return false, nil
		}
		if !item.expireAt.IsZero() && time.Now().After(item.expireAt) {
			return false, nil
		}
	}
	return true, nil
}

// GetMulti 批量获取给定键的值.
// 返回存在的键值对映射.
// 已过期的键视为不存在,从结果中省略.
// 使用读锁实现并发访问.
func (c *MemoryCache) GetMulti(ctx context.Context, keys []string) (map[string]any, error) {
	result := make(map[string]any)
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, key := range keys {
		item, ok := c.data[key]
		if !ok {
			continue
		}
		if !item.expireAt.IsZero() && time.Now().After(item.expireAt) {
			continue
		}
		result[key] = item.value
	}
	return result, nil
}

// SetMulti 批量存储键值对,使用相同的TTL.
// 如果ttl<=0,则永不过期.
// 使用写锁保护并发映射访问.
func (c *MemoryCache) SetMulti(ctx context.Context, items map[string]any, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	expireAt := time.Time{}
	if ttl > 0 {
		expireAt = time.Now().Add(ttl)
	}
	for key, value := range items {
		c.data[key] = cacheItem{
			value:    value,
			expireAt: expireAt,
		}
	}
	return nil
}

// DeleteMulti 批量删除给定键.
// 使用写锁保护并发映射访问.
func (c *MemoryCache) DeleteMulti(ctx context.Context, keys []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, key := range keys {
		delete(c.data, key)
	}
	return nil
}

// Clear 清空缓存中的所有项.
// 使用写锁保护并发映射访问.
func (c *MemoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]cacheItem)
	return nil
}

// CacheWithDeleteAll 扩展Cache,添加DeleteAll方法.
// 为兼容未来扩展而定义.
type CacheWithDeleteAll interface {
	Cache
	DeleteAll(ctx context.Context) error
}

// 验证*MemoryCache满足Cache接口.
var _ Cache = (*MemoryCache)(nil)

// GetWithGetter 从缓存获取值,如果未找到或已过期,
// 调用Getter函数获取值并存储到缓存中.
// 这是缓存旁路模式(懒加载).
// 如果Getter返回非nil值,则存入缓存且无过期时间(ttl=0).
// 返回Getter或缓存操作的任何错误.
func (c *MemoryCache) GetWithGetter(ctx context.Context, key string, fn Getter) (any, error) {
	val, err := c.Get(ctx, key)
	if err == nil {
		return val, nil
	}
	if !errors.Is(err, ErrNotFound) && !errors.Is(err, ErrCacheMiss) {
		return nil, err
	}

	val, err = fn(ctx, key)
	if err != nil {
		return nil, err
	}
	if val != nil {
		_ = c.Set(ctx, key, val, 0)
	}
	return val, nil
}
