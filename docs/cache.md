# cache 包 — 缓存抽象层

## 概述

`cache` 包提供统一的缓存操作接口抽象，支持不同的缓存实现（如 Redis、内存缓存）无缝替换。核心设计模式包括 Cache-Aside（缓存旁路）模式，通过 `Getter` 函数实现缓存未命中时的数据加载。

主要组件：

- `Cache` 接口 — 缓存核心操作
- `Getter` — 缓存未命中时的数据加载函数
- `MemoryCache` — 内存缓存实现（适用于测试和轻量场景）
- 错误 sentinel — `ErrNotFound` / `ErrCacheMiss`

---

## Cache 接口

```go
type Cache interface {
    Get(ctx context.Context, key string) (any, error)
    Set(ctx context.Context, key string, value any, ttl time.Duration) error
    Del(ctx context.Context, keys ...string) error
    Exists(ctx context.Context, key string) (bool, error)
    TTL(ctx context.Context, key string) (time.Duration, error)
    Close() error
}
```

### 方法说明

| 方法 | 说明 |
|------|------|
| `Get` | 获取指定键的缓存值，键不存在返回 `ErrNotFound`，已过期返回 `ErrCacheMiss` |
| `Set` | 设置缓存键值，可指定 TTL 过期时间。`ttl <= 0` 表示永不过期 |
| `Del` | 删除一个或多个缓存键 |
| `Exists` | 检查键是否存在（已过期的键视为不存在） |
| `TTL` | 获取键的剩余过期时间。未设置过期时间返回 0，不存在返回 `ErrNotFound` |
| `Close` | 关闭缓存连接并释放资源 |

---

## Getter — 数据加载函数

```go
type Getter func(ctx context.Context, key string) (any, error)
```

用于 Cache-Aside（缓存旁路）模式：当缓存未命中时，通过 Getter 从数据源加载数据并回填缓存。

实现 `GetWithGetter` 方法：

```go
func (c *MemoryCache) GetWithGetter(ctx context.Context, key string, fn Getter) (any, error)
```

流程：
1. 尝试从缓存获取值
2. 命中 → 直接返回
3. 未命中（`ErrNotFound` 或 `ErrCacheMiss`）→ 调用 Getter 加载
4. Getter 成功返回非 nil 值 → 写入缓存（ttl=0 永不过期）→ 返回值

---

## 错误 sentinel

```go
var (
    ErrNotFound  = errors.New("cache: key not found")
    ErrCacheMiss = errors.New("cache: key expired or not found")
)
```

| 错误 | 说明 |
|------|------|
| `ErrNotFound` | 缓存键不存在（从未设置） |
| `ErrCacheMiss` | 缓存键已过期或未命中 |

---

## MemoryCache — 内存缓存实现

MemoryCache 是基于 `sync.RWMutex` 和 `map[string]cacheItem` 的并发安全内存缓存实现，支持 TTL 过期和延迟清理。

### 创建

## cache.NewMemoryCache

创建内存缓存实例。

```go
cache := cache.NewMemoryCache()
```

### 使用场景
- 测试环境
- 轻量级应用
- 缓存原型开发

### 扩展方法

## cache.Get

获取指定键的缓存值。

```go
val, err := cache.Get(ctx, "user:1")
if errors.Is(err, cache.ErrNotFound) {
    // 键不存在
}
```

### 使用场景
- 从缓存读取数据
- 实现缓存读取逻辑
- 处理缓存未命中

## cache.Set

设置缓存键值，可指定 TTL 过期时间。

```go
err := cache.Set(ctx, "user:1", userData, 5*time.Minute)
```

### 使用场景
- 写入缓存数据
- 设置缓存过期时间
- 更新缓存内容

## cache.Del

删除一个或多个缓存键。

```go
err := cache.Del(ctx, "user:1", "user:2")
```

### 使用场景
- 删除特定缓存
- 批量清理缓存
- 失效缓存数据

## cache.Exists

检查键是否存在。

```go
exists, err := cache.Exists(ctx, "user:1")
```

### 使用场景
- 检查缓存是否存在
- 避免缓存穿透
- 实现缓存预热

## cache.TTL

获取键的剩余过期时间。

```go
ttl, err := cache.TTL(ctx, "user:1")
```

### 使用场景
- 监控缓存过期时间
- 实现缓存续期
- 缓存性能分析

## cache.GetWithGetter

Cache-Aside 模式：缓存未命中时从数据源加载。

```go
val, err := cache.GetWithGetter(ctx, "user:1", func(ctx context.Context, key string) (any, error) {
    var user User
    db.First(&user, 1)
    return user, nil
})
```

### 使用场景
- 缓存旁路模式
- 自动回填缓存
- 减少数据库访问

## cache.GetMulti

批量获取多个键的缓存值。

```go
result, err := cache.GetMulti(ctx, []string{"user:1", "user:2"})
```

### 使用场景
- 批量读取缓存
- 减少网络请求
- 提高缓存性能

## cache.SetMulti

批量设置多个缓存键值。

```go
items := map[string]any{"user:1": userData1, "user:2": userData2}
err := cache.SetMulti(ctx, items, time.Minute)
```

### 使用场景
- 批量写入缓存
- 批量预热缓存
- 提高缓存写入效率

## cache.DeleteMulti

批量删除多个缓存键。

```go
err := cache.DeleteMulti(ctx, []string{"user:1", "user:2"})
```

### 使用场景
- 批量清理缓存
- 批量失效缓存
- 提高缓存清理效率

## cache.Clear

清空所有缓存。

```go
err := cache.Clear(ctx)
```

### 使用场景
- 缓存重置
- 测试环境清理
- 紧急缓存清空

---

## 使用示例

### 基础操作

```go
c := cache.NewMemoryCache()
ctx := context.Background()

// 设置缓存，5 分钟过期
err := c.Set(ctx, "user:1", userData, 5*time.Minute)

// 获取缓存
val, err := c.Get(ctx, "user:1")
if errors.Is(err, cache.ErrNotFound) {
    // 键不存在
}
if errors.Is(err, cache.ErrCacheMiss) {
    // 键已过期
}

// 检查是否存在
exists, _ := c.Exists(ctx, "user:1")

// 获取剩余 TTL
ttl, _ := c.TTL(ctx, "user:1")

// 批量操作
items := map[string]any{"a": 1, "b": 2}
_ = c.SetMulti(ctx, items, time.Minute)
result, _ := c.GetMulti(ctx, []string{"a", "b"})
_ = c.DeleteMulti(ctx, []string{"a", "b"})

// 清空
_ = c.Clear(ctx)
```

### Cache-Aside 模式

```go
val, err := c.GetWithGetter(ctx, "user:1", func(ctx context.Context, key string) (any, error) {
    // 从数据库加载
    var user User
    db.First(&user, 1)
    return user, nil
})
```

### 与依赖注入集成

```go
// 注册为 Bean
container.Register("cache",
    core.Bean(cache.NewMemoryCache()),
    core.Singleton(),
)

// 注入使用
type UserService struct {
    Cache cache.Cache `inject:"cache"`
}
```

---

## 使用场景

### 场景 1：用户信息缓存

**描述**：缓存用户信息，减少数据库查询，提高查询性能。

```go
func (s *UserService) GetUser(id int) (*User, error) {
    key := fmt.Sprintf("user:%d", id)
    
    val, err := s.cache.GetWithGetter(ctx, key, func(ctx context.Context, key string) (any, error) {
        var user User
        if err := s.db.First(&user, id).Error; err != nil {
            return nil, err
        }
        return &user, nil
    })
    
    if err != nil {
        return nil, err
    }
    return val.(*User), nil
}

func (s *UserService) UpdateUser(user *User) error {
    if err := s.db.Save(user).Error; err != nil {
        return err
    }
    
    key := fmt.Sprintf("user:%d", user.ID)
    s.cache.Del(ctx, key)
    
    return nil
}
```

**最佳实践**：
- 使用 Cache-Aside 模式自动回填缓存
- 数据更新时主动失效缓存
- 设置合理的 TTL 避免缓存雪崩

### 场景 2：会话缓存

**描述**：缓存用户会话信息，减少数据库访问，提高认证性能。

```go
func (s *SessionService) CreateSession(userID string) (*Session, error) {
    session := &Session{
        ID:        generateSessionID(),
        UserID:    userID,
        CreatedAt: time.Now(),
        ExpiresAt: time.Now().Add(24 * time.Hour),
    }
    
    key := fmt.Sprintf("session:%s", session.ID)
    if err := s.cache.Set(ctx, key, session, 24*time.Hour); err != nil {
        return nil, err
    }
    
    return session, nil
}

func (s *SessionService) GetSession(sessionID string) (*Session, error) {
    key := fmt.Sprintf("session:%s", sessionID)
    
    val, err := s.cache.Get(ctx, key)
    if err != nil {
        return nil, err
    }
    
    return val.(*Session), nil
}
```

**最佳实践**：
- 设置合理的会话过期时间
- 使用随机 Session ID 避免冲突
- 定期清理过期会话

### 场景 3：配置缓存

**描述**：缓存系统配置信息，减少配置读取开销，提高系统性能。

```go
func (s *ConfigService) GetConfig(key string) (string, error) {
    cacheKey := fmt.Sprintf("config:%s", key)
    
    val, err := s.cache.GetWithGetter(ctx, cacheKey, func(ctx context.Context, key string) (any, error) {
        var config Config
        if err := s.db.Where("key = ?", key).First(&config).Error; err != nil {
            return nil, err
        }
        return config.Value, nil
    })
    
    if err != nil {
        return "", err
    }
    return val.(string), nil
}

func (s *ConfigService) UpdateConfig(key, value string) error {
    if err := s.db.Model(&Config{}).Where("key = ?", key).Update("value", value).Error; err != nil {
        return err
    }
    
    cacheKey := fmt.Sprintf("config:%s", key)
    s.cache.Del(ctx, cacheKey)
    
    return nil
}
```

**最佳实践**：
- 配置变更时主动失效缓存
- 使用较长的 TTL 减少缓存失效
- 考虑配置热更新机制

### 场景 4：计数器缓存

**描述**：缓存计数器信息，减少数据库更新频率，提高性能。

```go
func (s *CounterService) Increment(key string, delta int) (int, error) {
    cacheKey := fmt.Sprintf("counter:%s", key)
    
    val, err := s.cache.GetWithGetter(ctx, cacheKey, func(ctx context.Context, key string) (any, error) {
        var counter Counter
        if err := s.db.Where("key = ?", key).First(&counter).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                counter = Counter{Key: key, Value: 0}
                s.db.Create(&counter)
            } else {
                return nil, err
            }
        }
        return counter.Value, nil
    })
    
    if err != nil {
        return 0, err
    }
    
    count := val.(int) + delta
    s.cache.Set(ctx, cacheKey, count, 5*time.Minute)
    
    return count, nil
}

func (s *CounterService) PersistCounters() error {
    keys := []string{"counter:views", "counter:likes", "counter:shares"}
    result, err := s.cache.GetMulti(ctx, keys)
    if err != nil {
        return err
    }
    
    for cacheKey, val := range result {
        key := strings.TrimPrefix(cacheKey, "counter:")
        count := val.(int)
        s.db.Model(&Counter{}).Where("key = ?", key).Update("value", count)
    }
    
    return nil
}
```

**最佳实践**：
- 使用批量操作提高性能
- 定期持久化计数器到数据库
- 设置合理的 TTL 平衡性能和一致性