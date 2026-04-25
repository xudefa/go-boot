# Cache 模块

该模块提供灵活的缓存接口及多种实现。

## 概述

cache模块定义`Cache`接口,提供通用缓存操作:

- Get/Set 单个值
- Get/Set 批量值
- Delete 删除键
- Exists 检查存在性
- Clear 清空缓存

提供多种实现:

- MemoryCache: 内存缓存,支持TTL
- RedisCache: Redis缓存(开发中)

## 使用方法

### MemoryCache

```go
import "your-module/cache"

c := cache.NewMemoryCache()
ctx := context.Background()

// 设置带TTL的值
err := c.Set(ctx, "key", "value", time.Minute)

// 获取值
val, err := c.Get(ctx, "key")

// 删除值
err = c.Delete(ctx, "key")
```

### CacheWithGetter

`CacheWithGetter`接口扩展`Cache`,提供`GetWithGetter`方法,在缓存未命中时自动从数据源获取并缓存值:

```go
val, err := c.GetWithGetter(ctx, "key", func(ctx context.Context, key string) (any, error) {
    // 从数据源获取值(如数据库、API)
    return fetchValueFromSource(key)
})
```

## 实现说明

- 所有方法都是并发安全的
- 使用context传递取消和超时控制
- MemoryCache使用RWMutex实现高效的并发读写
- 过期项在访问时延迟删除
- 实现了缓存旁路模式(cache-aside pattern)