# tracing 包 — 分布式追踪

## 概述

`tracing` 包提供分布式追踪的抽象接口及 OpenTelemetry 实现。核心接口包括 `Tracer`（追踪器）、`Span`（追踪跨度）和 `SpanContext`（上下文信息），内置 `NoopTracer` / `NoopSpan` 空实现避免空指针问题。

## 核心接口

### Tracer 接口

追踪器，负责创建和管理 Span：

```go
type Tracer interface {
    Start(ctx context.Context, spanName string, opts ...SpanOption) (context.Context, Span)
    CurrentSpan(ctx context.Context) Span
    Finish()
}
```

- `Start()` — 创建一个新的 Span 并注入到 context 中
- `CurrentSpan()` — 从 context 中获取当前 Span
- `Finish()` — 刷新/关闭追踪器

### TracerProvider 接口

追踪器提供者：

```go
type TracerProvider interface {
    Tracer(name string) Tracer
}
```

### Span 接口

追踪跨度，记录操作的时间范围和元数据：

```go
type Span interface {
    End()                                   // 结束 Span
    AddEvent(name string, opts ...EventOption) // 添加事件
    SetAttribute(key string, value any)     // 设置属性
    RecordError(err error)                  // 记录错误
    SetError(err error)                     // 设置错误（同 RecordError）
    SetStatus(code SpanStatusCode)          // 设置状态码
    SpanContext() SpanContext               // 获取上下文
    GetTraceID() string                     // 获取 TraceID
    GetSpanID() string                      // 获取 SpanID
}
```

### SpanContext 结构体

```go
type SpanContext struct {
    TraceID string
    SpanID  string
}
```

### SpanStatusCode 状态码

| 值 | 说明 |
|----|------|
| `SpanStatusUnset` | 状态未设置（默认值） |
| `SpanStatusOK` | 操作成功完成 |
| `SpanStatusError` | 操作发生错误 |
| `SpanStatusCanceled` | 操作被取消 |

## 选项模式

### SpanOption

创建 Span 时的选项：

```go
// 创建 Span 时添加初始属性
tracer.Start(ctx, "operation", tracing.WithAttribute("key", "value"))
```

### EventOption

添加事件时的选项：

```go
span.AddEvent("cache.miss",
    tracing.WithEventAttribute("key", "user:123"),
)
```

## NoopTracer 和 NoopSpan

默认的空实现，所有方法均为空操作，避免 nil 判断：

```go
// 空操作
tracer := &tracing.NoopTracer{}
ctx, span := tracer.Start(context.Background(), "op")
span.End()
span.SetAttribute("key", "value")
```

## OpenTelemetryTracer 实现

基于 OpenTelemetry 的追踪器实现，使用 `context.WithValue` 在 context 中传递 Span：

```go
tracer := tracing.NewTracer("my-service")
ctx, span := tracer.Start(context.Background(), "handle-request")
defer span.End()

span.SetAttribute("http.method", "GET")
span.SetAttribute("http.path", "/api/users")

span.AddEvent("db.query.start")
// ... 执行数据库查询
span.AddEvent("db.query.end")

span.SetStatus(tracing.SpanStatusOK)
```

### generateID

OpenTelemetryTracer 使用 `crypto/rand` 生成随机 TraceID（32 字符）和 SpanID（16 字符）。

## 自动配置

通过 `boot.RegisterAutoConfig` 注册，由 `boot.Application` 在启动时自动执行：

```go
func init() {
    boot.RegisterAutoConfig(
        &TracingAutoConfiguration{},
        condition.OnProperty("tracing.enabled", "true"),
    )
}
```

### 配置属性

| 属性 | 默认值 | 说明 |
|------|--------|------|
| `tracing.enabled` | — | 设为 `"true"` 启用 Tracing |

### 自动配置行为

1. 创建 `TracerProviderImpl` 实例
2. 以 `"tracerProvider"` ID 注册到 IoC 容器（单例模式）

## 完整使用示例

```go
package main

import (
    "context"
    "fmt"
    "github.com/xudefa/go-boot/tracing"
)

func main() {
    tracer := tracing.NewTracer("order-service")

    // 创建根 Span
    ctx, span := tracer.Start(context.Background(), "create-order",
        tracing.WithAttribute("user_id", "u123"),
        tracing.WithAttribute("order_id", "ord-456"),
    )
    defer span.End()

    // 获取 TraceID
    fmt.Println("TraceID:", span.GetTraceID())
    fmt.Println("SpanID:", span.GetSpanID())

    // 记录事件
    span.AddEvent("payment.started",
        tracing.WithEventAttribute("amount", "99.99"),
    )

    // 设置属性
    span.SetAttribute("payment.method", "credit_card")

    // 子 Span 示例
    subTracer := tracing.NewTracer("db-client")
    subCtx, subSpan := subTracer.Start(ctx, "insert-order")
    subSpan.SetAttribute("table", "orders")
    subSpan.End()

    // 在当前 context 中获取 Span
    currentSpan := tracer.CurrentSpan(subCtx)
    fmt.Println("Current:", currentSpan.GetSpanID())

    // 错误记录
    span.RecordError(fmt.Errorf("payment timeout"))
    span.SetStatus(tracing.SpanStatusError)

    tracer.Finish()
}
```

### 自动配置 + IoC 集成

```go
package main

import (
    "context"
    "github.com/xudefa/go-boot/boot"
    "github.com/xudefa/go-boot/tracing"
)

func main() {
    app := boot.NewApplication(
        boot.WithProperty("tracing.enabled", "true"),
    )
    ctx := app.Context()

    var provider tracing.TracerProvider
    ctx.Get("tracerProvider", &provider)

    tracer := provider.Tracer("my-service")
    _, span := tracer.Start(context.Background(), "main")
    defer span.End()
}
```

## 设计要点

- `OpenTelemetryTracer` 使用 `context.WithValue` 将 Span 存储在 context 中，通过 `spanContextKey{}` 类型键隔离命名空间
- `NoopTracer` / `NoopSpan` 提供安全的空实现，适合测试和 tracing 未启用时的场景
- `SetError` 和 `RecordError` 功能相同，均将错误信息写入 `"error"` 属性
- Span 状态码支持 `Unset`、`OK`、`Error`、`Canceled` 四种状态
- `SpanContext` 不依赖于 tracing 系统，为纯数据结构的 TraceID/SpanID 载体
