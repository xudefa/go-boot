# Refresh Scope

配置/Bean 热重载机制，支持配置变更时自动刷新相关 Bean。

## 功能特性

- **自动刷新**: 配置变更时自动触发相关 Bean 的重新初始化
- **代理模式**: 使用代理实现平滑切换，旧请求继续使用旧实例，新请求使用新实例
- **事件驱动**: 统一的事件接口，支持多种配置源
- **指标收集**: 内置刷新指标，监控刷新性能
- **可配置**: 支持自定义刷新配置选项

## 快速开始

```go
import (
    "github.com/xudefa/go-boot/context"
    "github.com/xudefa/go-boot/core"
    "github.com/xudefa/go-boot/environment"
    "github.com/xudefa/go-boot/refresh"
)

// 创建应用上下文
container := core.New()
env := environment.NewEnvironment()
ctx := context.NewApplicationContext(container, env)

// 注册 RefreshScope Bean
ctx.Register("databaseConfig",
    core.Bean(&DatabaseConfig{
        Host:     "localhost",
        Port:     3306,
        Username: "root",
    }),
    core.RefreshScope(),
    core.ConfigKeys("db.host", "db.port", "db.username"),
)

// 启动应用
ctx.Start()
defer ctx.Stop()
```

## 核心概念

### RefreshScope

标记需要热重载的 Bean，当配置变更时自动刷新。

```go
ctx.Register("config",
    core.Bean(&Config{}),
    core.RefreshScope(),
    core.ConfigKeys("app.timeout", "app.retry"),
)
```

### ConfigKeys

指定 Bean 依赖的配置键，用于检测配置变更。

```go
core.ConfigKeys("db.host", "db.port", "db.username")
```

### RefreshableBean

实现 `RefreshableBean` 接口的 Bean 可以自定义配置变更处理逻辑。

```go
type MyService struct {
    config *Config
}

func (s *MyService) OnConfigChange(event refresh.ConfigChangeEvent) error {
    // 自定义配置变更处理
    return nil
}
```

## 配置选项

```go
ctx := context.NewApplicationContext(container, env,
    refresh.WithRefreshEnabled(true),
    refresh.WithRefreshDelay(100*time.Millisecond),
    refresh.WithMaxRefreshAttempts(3),
)
```

### 可用选项

- `WithRefreshEnabled(bool)`: 启用/禁用刷新功能
- `WithRefreshDelay(time.Duration)`: 设置刷新延迟
- `WithMaxRefreshAttempts(int)`: 设置最大刷新尝试次数
- `WithRefreshLogger(*slog.Logger)`: 设置日志记录器

## 指标监控

```go
mgr := ctx.RefreshScopeManager()
metrics := mgr.Metrics()

fmt.Printf("Total refreshes: %d\n", metrics.TotalRefreshes())
fmt.Printf("Successful: %d\n", metrics.SuccessfulRefreshes())
fmt.Printf("Failed: %d\n", metrics.FailedRefreshes())
fmt.Printf("Average time: %v\n", metrics.AverageRefreshTime())
```

## 事件系统

### ConfigChangeEvent

配置变更事件，包含变更的键、旧值和新值。

```go
event := refresh.ConfigChangeEvent{
    EventType: "modify",
    Keys:      []string{"server.port"},
    OldValues: map[string]any{"server.port": 8080},
    NewValues: map[string]any{"server.port": 9090},
    Source:    "viper",
}
```

### BeanRefreshedEvent

Bean 刷新完成事件。

```go
event := refresh.BeanRefreshedEvent{
    BeanID:      "config",
    OldVersion:  1,
    NewVersion:  2,
    RefreshTime: time.Now(),
    Success:     true,
}
```

## 最佳实践

1. **合理使用 RefreshScope**: 只在需要热重载的 Bean 上使用
2. **配置键精确匹配**: ConfigKeys 应该精确指定依赖的配置键
3. **错误处理**: 实现 RefreshableBean 接口时做好错误处理
4. **监控指标**: 定期检查刷新指标，确保系统健康

## 示例

完整示例请参考 [go-boot-examples](https://github.com/xudefa/go-boot-examples) 项目。

## API 参考

### RefreshScopeManager

刷新作用域管理器，提供刷新功能的核心接口。

```go
type RefreshScopeManager struct {
    // ...
}

// 标记 Bean 为需要刷新
func (m *RefreshScopeManager) MarkBeanForRefresh(beanID string)

// 获取刷新后的 Bean
func (m *RefreshScopeManager) GetRefreshedBean(beanID string) (any, error)

// 注册可刷新 Bean
func (m *RefreshScopeManager) RegisterRefreshableBean(beanID string, bean RefreshableBean)

// 返回刷新指标
func (m *RefreshScopeManager) Metrics() *RefreshMetrics
```

### RefreshMetrics

刷新指标，记录刷新操作的性能数据。

```go
type RefreshMetrics struct {
    // ...
}

// 总刷新次数
func (m *RefreshMetrics) TotalRefreshes() int64

// 成功刷新次数
func (m *RefreshMetrics) SuccessfulRefreshes() int64

// 失败刷新次数
func (m *RefreshMetrics) FailedRefreshes() int64

// 平均刷新时间
func (m *RefreshMetrics) AverageRefreshTime() time.Duration

// 最后刷新时间
func (m *RefreshMetrics) LastRefreshTime() time.Time
```

## 注意事项

1. **循环依赖**: 避免在 RefreshScope Bean 之间创建循环依赖
2. **性能影响**: 频繁的配置变更可能影响性能
3. **线程安全**: RefreshScopeManager 是线程安全的
4. **延迟刷新**: 配置变更后，Bean 会在下次访问时刷新