// Package cache 定义缓存操作的核心接口。
//
// 该包提供缓存访问的抽象层，支持不同的缓存实现
// (如 Redis、内存缓存) 而不需要修改使用方的代码。
//
// 主要接口：
//
//   - Cache: 缓存操作接口（Get/Set/Del/Exists/TTL/Close）
//   - Getter: 缓存旁路模式的值加载函数
//   - MemoryCache: 内存缓存实现
package cache

import (
	"context"
	"time"
)

// Cache 缓存操作接口
//
// 所有缓存实现都应该实现此接口，
// 以便与 go-boot 的依赖注入系统集成。
type Cache interface {
	// Get 获取指定键的缓存值，不存在返回 ErrNotFound
	Get(ctx context.Context, key string) (any, error)

	// Set 设置缓存键值，ttl<=0 表示永不过期
	Set(ctx context.Context, key string, value any, ttl time.Duration) error

	// Del 删除指定的缓存键
	Del(ctx context.Context, keys ...string) error

	// Exists 检查键是否存在且未过期
	Exists(ctx context.Context, key string) (bool, error)

	// TTL 获取键的剩余过期时间
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Close 关闭缓存连接并释放资源
	Close() error
}
