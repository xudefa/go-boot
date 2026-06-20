package cache

import (
	"context"
	"fmt"
	"time"
)

// MemoryCacheBuilder 内存缓存构建器，支持链式配置
type MemoryCacheBuilder struct {
	initialCapacity int
}

// NewMemoryCacheBuilder 创建内存缓存构建器
func NewMemoryCacheBuilder() *MemoryCacheBuilder {
	return &MemoryCacheBuilder{
		initialCapacity: 1024, // 默认初始容量
	}
}

// InitialCapacity 设置初始容量
func (b *MemoryCacheBuilder) InitialCapacity(capacity int) *MemoryCacheBuilder {
	b.initialCapacity = capacity
	return b
}

// Build 构建内存缓存
func (b *MemoryCacheBuilder) Build() *MemoryCache {
	return &MemoryCache{
		data: make(map[string]cacheItem, b.initialCapacity),
	}
}

// MustBuild 构建内存缓存
func (b *MemoryCacheBuilder) MustBuild() *MemoryCache {
	return b.Build()
}

// CacheHelper 缓存辅助工具，简化常见缓存操作
type CacheHelper struct {
	cache Cache
}

// NewCacheHelper 创建缓存辅助工具
func NewCacheHelper(cache Cache) *CacheHelper {
	return &CacheHelper{cache: cache}
}

// Get 获取缓存值，调用方需自行类型断言
func (h *CacheHelper) Get(ctx context.Context, key string) (any, error) {
	return h.cache.Get(ctx, key)
}

// Set 设置缓存值
func (h *CacheHelper) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return h.cache.Set(ctx, key, value, ttl)
}

// GetOrSet 获取缓存值，如果不存在则使用提供的函数获取并缓存
func (h *CacheHelper) GetOrSet(ctx context.Context, key string, fn func() (any, error), ttl time.Duration) (any, error) {
	// 尝试从缓存获取
	val, err := h.cache.Get(ctx, key)
	if err == nil {
		return val, nil
	}

	// 缓存未命中，调用函数获取
	result, err := fn()
	if err != nil {
		return nil, err
	}

	// 存储到缓存
	if setErr := h.cache.Set(ctx, key, result, ttl); setErr != nil {
		return result, fmt.Errorf("failed to cache value: %w", setErr)
	}

	return result, nil
}

// Invalidate 使缓存失效
func (h *CacheHelper) Invalidate(ctx context.Context, key string) error {
	return h.cache.Del(ctx, key)
}

// InvalidateAll 使所有指定键的缓存失效
func (h *CacheHelper) InvalidateAll(ctx context.Context, keys ...string) error {
	return h.cache.Del(ctx, keys...)
}

// Clear 清空所有缓存
func (h *CacheHelper) Clear(ctx context.Context) error {
	if mc, ok := h.cache.(*MemoryCache); ok {
		return mc.Clear(ctx)
	}
	return fmt.Errorf("cache does not support Clear operation")
}

// Exists 检查键是否存在
func (h *CacheHelper) Exists(ctx context.Context, key string) (bool, error) {
	return h.cache.Exists(ctx, key)
}

// TTL 获取键的剩余过期时间
func (h *CacheHelper) TTL(ctx context.Context, key string) (time.Duration, error) {
	return h.cache.TTL(ctx, key)
}

// CacheTemplate 缓存模板，提供常用的缓存操作模板
type CacheTemplate struct {
	cache  Cache
	prefix string
}

// NewCacheTemplate 创建缓存模板
func NewCacheTemplate(cache Cache, prefix string) *CacheTemplate {
	return &CacheTemplate{
		cache:  cache,
		prefix: prefix,
	}
}

// Key 生成带前缀的键
func (t *CacheTemplate) Key(key string) string {
	if t.prefix == "" {
		return key
	}
	return t.prefix + ":" + key
}

// Get 获取缓存值
func (t *CacheTemplate) Get(ctx context.Context, key string) (any, error) {
	return t.cache.Get(ctx, t.Key(key))
}

// Set 设置缓存值
func (t *CacheTemplate) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return t.cache.Set(ctx, t.Key(key), value, ttl)
}

// Del 删除缓存键
func (t *CacheTemplate) Del(ctx context.Context, key string) error {
	return t.cache.Del(ctx, t.Key(key))
}

// Exists 检查键是否存在
func (t *CacheTemplate) Exists(ctx context.Context, key string) (bool, error) {
	return t.cache.Exists(ctx, t.Key(key))
}

// TTL 获取键的剩余过期时间
func (t *CacheTemplate) TTL(ctx context.Context, key string) (time.Duration, error) {
	return t.cache.TTL(ctx, t.Key(key))
}

// GetOrSet 获取或设置缓存值
func (t *CacheTemplate) GetOrSet(ctx context.Context, key string, fn func() (any, error), ttl time.Duration) (any, error) {
	fullKey := t.Key(key)

	val, err := t.cache.Get(ctx, fullKey)
	if err == nil {
		return val, nil
	}

	result, err := fn()
	if err != nil {
		return nil, err
	}

	if setErr := t.cache.Set(ctx, fullKey, result, ttl); setErr != nil {
		return result, fmt.Errorf("failed to cache value: %w", setErr)
	}

	return result, nil
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Enabled      bool          // 是否启用缓存
	DefaultTTL   time.Duration // 默认TTL
	MaxSize      int           // 最大缓存项数
	KeyPrefix    string        // 键前缀
	StatsEnabled bool          // 是否启用统计
}

// CacheOption 缓存配置选项
type CacheOption func(*CacheConfig)

// WithCacheEnabled 设置是否启用缓存
func WithCacheEnabled(enabled bool) CacheOption {
	return func(c *CacheConfig) {
		c.Enabled = enabled
	}
}

// WithDefaultTTL 设置默认TTL
func WithDefaultTTL(ttl time.Duration) CacheOption {
	return func(c *CacheConfig) {
		c.DefaultTTL = ttl
	}
}

// WithMaxSize 设置最大缓存项数
func WithMaxSize(size int) CacheOption {
	return func(c *CacheConfig) {
		c.MaxSize = size
	}
}

// WithKeyPrefix 设置键前缀
func WithKeyPrefix(prefix string) CacheOption {
	return func(c *CacheConfig) {
		c.KeyPrefix = prefix
	}
}

// WithStatsEnabled 设置是否启用统计
func WithStatsEnabled(enabled bool) CacheOption {
	return func(c *CacheConfig) {
		c.StatsEnabled = enabled
	}
}

// DefaultCacheConfig 返回默认缓存配置
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		Enabled:      true,
		DefaultTTL:   30 * time.Minute,
		MaxSize:      10000,
		StatsEnabled: false,
	}
}

// ApplyOptions 应用配置选项
func (c *CacheConfig) ApplyOptions(opts []CacheOption) {
	for _, opt := range opts {
		opt(c)
	}
}
