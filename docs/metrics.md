# metrics 包 — 指标收集抽象

## 概述

`metrics` 包提供轻量级的指标收集接口和默认实现，**零外部依赖**，仅使用 Go 标准库。核心概念参考 Micrometer，提供 `Counter`（计数器）和 `Gauge`（仪表盘）两种指标类型，通过 `MeterRegistry` 统一管理。

## Counter 接口

计数器，只增不减，适用于记录请求总数、错误次数等单调递增的数值：

```go
type Counter interface {
    Inc()          // 加 1
    Add(v float64) // 增加指定值
    Value() float64 // 获取当前值
}
```

使用示例：

```go
counter := metrics.NewSimpleCounter()
counter.Inc()
counter.Add(10)
fmt.Println(counter.Value()) // 11
```

## Gauge 接口

仪表盘，可增可减，适用于记录当前连接数、内存使用量、CPU 使用率等：

```go
type Gauge interface {
    Set(v float64)  // 设置当前值
    Add(v float64)  // 增加指定值（负数即减少）
    Value() float64 // 获取当前值
}
```

使用示例：

```go
gauge := metrics.NewSimpleGauge()
gauge.Set(100)
gauge.Add(-5)
fmt.Println(gauge.Value()) // 95
```

## MeterRegistry 接口

指标注册表，管理 Counter 和 Gauge 的创建与收集：

```go
type MeterRegistry interface {
    Counter(name string, tags ...string) Counter // 获取或创建计数器
    Gauge(name string, tags ...string) Gauge     // 获取或创建仪表盘
    Collect() []Metric                            // 收集所有指标快照
}
```

- `tags` 参数为偶数个的键值对序列（如 `"service", "auth", "version", "v1"`）
- `Collect()` 返回所有已注册指标的 `Metric` 快照

## SimpleRegistry 实现

`SimpleRegistry` 是 `MeterRegistry` 的默认实现，使用 `map` 存储指标，`sync.Mutex` 保证并发安全：

```go
registry := metrics.NewSimpleRegistry()

counter := registry.Counter("http_requests_total", "method", "GET")
counter.Inc()

gauge := registry.Gauge("memory_usage", "type", "heap")
gauge.Set(1024.5)

// 收集所有指标
allMetrics := registry.Collect()
for _, m := range allMetrics {
    fmt.Printf("%s = %.2f (tags: %v)\n", m.Name, m.Value, m.Tags)
}
```

## Metric 结构体

指标快照结构体，用于采集和上报：

```go
type Metric struct {
    Name  string            // 指标名称
    Value float64           // 指标当前值
    Tags  map[string]string // 指标标签
}
```

## 完整使用示例

```go
package main

import (
    "fmt"
    "github.com/xudefa/go-boot/metrics"
)

func main() {
    registry := metrics.NewSimpleRegistry()

    // 创建计数器
    reqCounter := registry.Counter("requests_total", "service", "api")
    errCounter := registry.Counter("errors_total", "service", "api")

    // 创建仪表盘
    connGauge := registry.Gauge("active_connections", "pool", "main")

    // 模拟业务
    reqCounter.Inc()
    reqCounter.Inc()
    errCounter.Inc()
    connGauge.Set(42)
    connGauge.Add(-1)

    // 采集
    for _, m := range registry.Collect() {
        fmt.Printf("Metric: %s = %.0f\n", m.Name, m.Value)
    }
    // 输出:
    // Metric: requests_total = 2
    // Metric: errors_total = 1
    // Metric: active_connections = 41
}
```

## 独立使用 Counter / Gauge

`SimpleCounter` 和 `SimpleGauge` 可独立于 Registry 使用：

```go
counter := metrics.NewSimpleCounter()
counter.Inc()

gauge := metrics.NewSimpleGauge()
gauge.Set(100)
```

## 设计要点

- `SimpleCounter` 使用 `sync.RWMutex`：读操作（Value）使用 RLock，写操作（Inc/Add）使用 Lock
- `SimpleGauge` 使用 `sync.Mutex`（无区分读写锁的必要）
- `SimpleRegistry` 按名称索引，同名 Counter 或 Gauge 在注册表中共享同一实例
- `tags` 参数通过 `parseTags()` 解析为 `map[string]string`，偶数索引为 key，奇数索引为 value
- 所有实现均为并发安全，适用于多 goroutine 场景
