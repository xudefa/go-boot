package cache

import (
	"container/list"
	"context"
	"sync"
	"time"
)

// LRUCache LRU（最近最少使用）缓存实现
//
// 基于双向链表和哈希表实现 O(1) 时间复杂度的缓存操作。
// 支持 TTL 过期淘汰和容量限制淘汰。
//
// 使用示例：
//
//	cache := NewLRUCache(100) // 容量 100
//	cache.Set(context.Background(), "key", "value", time.Minute)
//	val, err := cache.Get(context.Background(), "key")
type LRUCache struct {
	mu        sync.RWMutex
	capacity  int
	items     map[string]*list.Element
	evictList *list.List
	onEvict   func(key string, value any)
}

// lruEntry LRU 缓存项
type lruEntry struct {
	key       string
	value     any
	expiresAt time.Time
}

// LRUOption LRU 缓存选项函数
type LRUOption func(*LRUCache)

// WithEvictCallback 设置淘汰回调函数
//
// 当缓存项因容量限制被淘汰时调用此回调。
func WithEvictCallback(fn func(key string, value any)) LRUOption {
	return func(c *LRUCache) {
		c.onEvict = fn
	}
}

// NewLRUCache 创建 LRU 缓存
//
// 参数:
//   - capacity: 缓存容量，必须大于 0
//   - opts: 可选配置项
//
// 返回:
//   - *LRUCache: LRU 缓存实例
func NewLRUCache(capacity int, opts ...LRUOption) *LRUCache {
	if capacity <= 0 {
		capacity = 100 // 默认容量
	}
	c := &LRUCache{
		capacity:  capacity,
		items:     make(map[string]*list.Element),
		evictList: list.New(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Get 获取缓存值
func (c *LRUCache) Get(ctx context.Context, key string) (any, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		return nil, ErrNotFound
	}

	entry := elem.Value.(*lruEntry)

	// 检查是否过期
	if time.Now().After(entry.expiresAt) {
		c.removeElement(elem)
		return nil, ErrNotFound
	}

	// 移到最前（最近使用）
	c.evictList.MoveToFront(elem)
	return entry.value, nil
}

// Set 设置缓存值
func (c *LRUCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果已存在，更新值并移到最前
	if elem, ok := c.items[key]; ok {
		c.evictList.MoveToFront(elem)
		entry := elem.Value.(*lruEntry)
		entry.value = value
		entry.expiresAt = time.Now().Add(ttl)
		return nil
	}

	// 如果超出容量，淘汰最久未使用的项
	for c.evictList.Len() >= c.capacity {
		c.evictOldest()
	}

	// 添加新项
	entry := &lruEntry{
		key:       key,
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
	elem := c.evictList.PushFront(entry)
	c.items[key] = elem

	return nil
}

// Del 删除缓存项
func (c *LRUCache) Del(ctx context.Context, keys ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, key := range keys {
		if elem, ok := c.items[key]; ok {
			c.removeElement(elem)
		}
	}
	return nil
}

// Exists 检查键是否存在且未过期
func (c *LRUCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		return false, nil
	}

	entry := elem.Value.(*lruEntry)
	if time.Now().After(entry.expiresAt) {
		c.removeElement(elem)
		return false, nil
	}

	return true, nil
}

// TTL 获取键的剩余过期时间
func (c *LRUCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		return 0, ErrNotFound
	}

	entry := elem.Value.(*lruEntry)
	remaining := time.Until(entry.expiresAt)
	if remaining <= 0 {
		c.removeElement(elem)
		return 0, ErrNotFound
	}

	return remaining, nil
}

// Close 关闭缓存
func (c *LRUCache) Close() error {
	c.Clear()
	return nil
}

// Len 返回缓存项数量
func (c *LRUCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.evictList.Len()
}

// Clear 清空缓存
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.evictList = list.New()
}

// evictOldest 淘汰最久未使用的项
func (c *LRUCache) evictOldest() {
	elem := c.evictList.Back()
	if elem != nil {
		c.removeElement(elem)
	}
}

// removeElement 移除元素
func (c *LRUCache) removeElement(elem *list.Element) {
	c.evictList.Remove(elem)
	entry := elem.Value.(*lruEntry)
	delete(c.items, entry.key)

	if c.onEvict != nil {
		c.onEvict(entry.key, entry.value)
	}
}
