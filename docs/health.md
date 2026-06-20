# health 包 — 健康检查抽象

## 概述

`health` 包提供健康检查的核心接口和聚合器，**零外部依赖**，仅使用 Go 标准库。各集成模块（如数据库、Redis）通过实现 `Indicator` 接口来提供组件的健康状态检查能力。

## Indicator 接口

健康指标接口，由各集成模块实现：

```go
type Indicator interface {
    Name() string
    Health(ctx context.Context) Health
}
```

- `Name()` — 返回指标名称（如 `"database"`、`"redis"`）
- `Health(ctx)` — 执行健康检查并返回 `Health` 结果

### 实现示例

```go
type MyIndicator struct {
    name string
}

func (m *MyIndicator) Name() string {
    return m.name
}

func (m *MyIndicator) Health(ctx context.Context) health.Health {
    if err := checkSomething(ctx); err != nil {
        return health.Health{
            Status:  health.StatusDown,
            Details: map[string]any{"error": err.Error()},
        }
    }
    return health.Health{Status: health.StatusUp}
}
```

## Health 结构体

健康信息结构体，包含状态、详情和时间戳：

```go
type Health struct {
    Status    Status         // 健康状态枚举
    Details   map[string]any // 详细信息
    Error     error          // 错误信息（不参与 JSON 序列化）
    Timestamp time.Time      // 检查时间戳
}
```

## Status 枚举

| 值 | 名称 | 优先级 | 说明 |
|----|------|--------|------|
| `StatusUp` | `"UP"` | 最高 | 组件运行正常 |
| `StatusDegraded` | `"DEGRADED"` | ↓ | 组件部分功能可用 |
| `StatusDown` | `"DOWN"` | ↓ | 组件不可用 |
| `StatusOutage` | `"OUTAGE"` | ↓ | 组件完全停服 |
| `StatusUnknown` | `"UNKNOWN"` | 最低 | 无法确定组件状态 |

优先级从高到低：`StatusUp` > `StatusDegraded` > `StatusDown` > `StatusOutage` > `StatusUnknown`。

## Aggregator 聚合器

聚合所有 `Indicator` 的健康状态，支持并发安全。

### 聚合规则

- 遍历所有指标，跟踪最差状态
- 全部指标为 `UP` → 整体 `UP`
- 任一指标为 `OUTAGE` → 整体 `OUTAGE`
- 任一指标为 `DOWN` → 整体 `DOWN`
- 任一指标为 `DEGRADED` → 整体 `DEGRADED`
- 优先级顺序：`StatusOutage` > `StatusDown` > `StatusDegraded` > `StatusUp` > `StatusUnknown`

### 完整使用示例

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/xudefa/go-boot/health"
)

// 自定义健康指标
type PingIndicator struct {
    target string
}

func (p *PingIndicator) Name() string {
    return p.target
}

func (p *PingIndicator) Health(ctx context.Context) health.Health {
    // 模拟健康检查
    if p.target == "broken" {
        return health.Health{
            Status: health.StatusDown,
            Details: map[string]any{
                "error": "connection refused",
            },
        }
    }
    return health.Health{
        Status:  health.StatusUp,
        Details: map[string]any{"latency": "5ms"},
    }
}

func main() {
    agg := health.NewAggregator()
    agg.AddIndicator(&PingIndicator{target: "database"})
    agg.AddIndicator(&PingIndicator{target: "cache"})
    agg.AddIndicator(&PingIndicator{target: "broken"})

    result := agg.Aggregate(context.Background())
    fmt.Printf("Overall Status: %s\n", result.Status) // DOWN
    fmt.Printf("Time: %s\n", result.Timestamp.Format(time.RFC3339))
    for name, detail := range result.Details {
        fmt.Printf("%s: %v\n", name, detail)
    }
}
```

### 获取指标列表

```go
indicators := agg.Indicators()
for _, ind := range indicators {
    fmt.Println(ind.Name())
}
```

## 设计要点

- `Aggregator` 使用 `sync.RWMutex` 保证读写并发安全
- `Aggregate()` 内部使用 `Indicators()` 获取副本，避免持有锁期间调用外部 `Health()`
- `Health.Error` 字段使用 `json:"-"` 标记，不参与 JSON 序列化，避免敏感信息泄露
- 时间戳在聚合时由 `Aggregate()` 统一设置，不依赖各指标返回时间
