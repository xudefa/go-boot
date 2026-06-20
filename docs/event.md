# event 包 — 事件系统

## 概述

`event` 包提供应用事件驱动支持，参考 Spring 的 `ApplicationEvent` / `ApplicationListener` 模式。核心功能包括：

- **ApplicationEvent** — 应用事件接口，所有自定义事件需实现此接口
- **EventBus** — 事件总线，支持事件的发布和订阅
- **BaseEvent** — 基础事件实现，可直接使用或嵌入自定义事件结构体

## ApplicationEvent 接口

```go
type ApplicationEvent interface {
    Type() string
    Timestamp() time.Time
}
```

所有应用事件必须实现此接口：

- `Type()` — 返回事件类型字符串，用于事件路由和匹配
- `Timestamp()` — 返回事件发生的时间戳

## EventListener

```go
type EventListener func(event ApplicationEvent)
```

监听器是一个函数类型，接收 `ApplicationEvent` 参数，处理事件通知。

## EventBus

```go
type EventBus struct {
    mu        sync.RWMutex
    listeners map[string][]EventListener
}
```

`EventBus` 是事件总线的核心实现，负责事件的发布与订阅管理。线程安全，支持并发发布和订阅。

### 创建

## event.NewEventBus

创建事件总线。

```go
bus := event.NewEventBus()
```

### 使用场景
- 应用启动时创建事件总线
- 模块间通信
- 解耦业务逻辑

### 订阅事件

## event.Subscribe

订阅指定类型的事件。

```go
bus.Subscribe(event.EventApplicationStarted, func(e event.ApplicationEvent) {
    fmt.Println("应用已启动，时间:", e.Timestamp())
})
```

### 使用场景
- 监听应用生命周期事件
- 处理业务事件
- 实现观察者模式

### 取消订阅

## event.Unsubscribe

取消订阅指定类型的事件。

```go
handler := func(e event.ApplicationEvent) {
    fmt.Println("收到事件:", e.Type())
}

bus.Subscribe(event.EventApplicationReady, handler)
bus.Unsubscribe(event.EventApplicationReady, handler)
```

### 使用场景
- 动态管理监听器
- 避免重复订阅
- 清理资源

### 发布事件

## event.Publish

发布事件到事件总线。

```go
bus.Publish(&event.BaseEvent{EventType: event.EventApplicationStarted})
```

### 使用场景
- 触发业务流程
- 通知状态变化
- 实现事件驱动架构

发布流程：
1. 获取事件类型对应的监听器列表（读锁保护）
2. 如果无监听器则直接返回
3. 拷贝监听器快照（保证并发安全）
4. 遍历监听器并逐个同步调用

## BaseEvent

基础事件实现，可直接使用，也支持嵌入到自定义事件结构体中。

```go
type BaseEvent struct {
    EventType string
    EventTime time.Time
}

func (e *BaseEvent) Type() string
func (e *BaseEvent) Timestamp() time.Time
```

### 使用场景
- 创建简单事件
- 嵌入到自定义事件中
- 快速实现事件接口

## event.BaseEvent

创建基础事件。

```go
evt := &event.BaseEvent{EventType: "CustomEvent"}
```

### 使用场景
- 快速创建事件
- 内置事件发布
- 测试和调试

- `EventType` — 事件类型字符串
- `EventTime` — 事件发生时间，可选，为空时 `Timestamp()` 自动返回 `time.Now()`

### 自定义事件

```go
type MyCustomEvent struct {
    event.BaseEvent
    UserID   string
    Action   string
}

func main() {
    evt := &MyCustomEvent{
        BaseEvent: event.BaseEvent{EventType: "UserLogin"},
        UserID:    "12345",
        Action:    "login",
    }

    bus := event.NewEventBus()
    bus.Subscribe("UserLogin", func(e event.ApplicationEvent) {
        myEvt := e.(*MyCustomEvent)
        fmt.Printf("用户 %s 执行了操作: %s\n", myEvt.UserID, myEvt.Action)
    })

    bus.Publish(evt)
}
```

## 内置事件类型

| 常量 | 值 | 说明 |
|------|----|------|
| `EventEnvironmentPrepared` | `"EnvironmentPrepared"` | 环境准备完成 |
| `EventContextRefreshed` | `"ContextRefreshed"` | 上下文刷新完成 |
| `EventApplicationStarted` | `"ApplicationStarted"` | 应用已启动 |
| `EventApplicationReady` | `"ApplicationReady"` | 应用已就绪 |
| `EventApplicationStopped` | `"ApplicationStopped"` | 应用已停止 |

## 使用示例

### 基本事件发布与订阅

```go
package main

import (
    "fmt"
    "github.com/xudefa/go-boot/event"
)

func main() {
    bus := event.NewEventBus()

    // 订阅多个事件
    bus.Subscribe(event.EventApplicationStarted, func(e event.ApplicationEvent) {
        fmt.Println("应用启动中...")
    })

    bus.Subscribe(event.EventApplicationReady, func(e event.ApplicationEvent) {
        fmt.Println("应用已就绪，可以提供服务")
    })

    bus.Subscribe(event.EventApplicationStopped, func(e event.ApplicationEvent) {
        fmt.Println("应用已停止")
    })

    // 发布事件
    bus.Publish(&event.BaseEvent{EventType: event.EventApplicationStarted})
    bus.Publish(&event.BaseEvent{EventType: event.EventApplicationReady})
    bus.Publish(&event.BaseEvent{EventType: event.EventApplicationStopped})
}
```

### 在 Boot 启动中的事件流

`Boot.Start()` 按顺序发布事件：

```go
// PhaseConfiguring 阶段
eventBus.Publish(&event.BaseEvent{EventType: event.EventEnvironmentPrepared})

// PhaseContextRefreshed 阶段
eventBus.Publish(&event.BaseEvent{EventType: event.EventContextRefreshed})

// PhaseRunning 阶段
eventBus.Publish(&event.BaseEvent{EventType: event.EventApplicationStarted})
eventBus.Publish(&event.BaseEvent{EventType: event.EventApplicationReady})
```

`Boot.Stop()` 发布：

```go
eventBus.Publish(&event.BaseEvent{EventType: event.EventApplicationStopped})
```

完整的事件时间线：

```
PhaseConfiguring    → EventEnvironmentPrepared
PhaseContextRefreshed → EventContextRefreshed
PhaseRunning        → EventApplicationStarted
PhaseRunning        → EventApplicationReady
PhaseStopped        → EventApplicationStopped
```

### 多监听器

```go
bus := event.NewEventBus()

// 多个监听器订阅同一事件
bus.Subscribe(event.EventApplicationStarted, func(e event.ApplicationEvent) {
    fmt.Println("监听器 1: 记录启动日志")
})

bus.Subscribe(event.EventApplicationStarted, func(e event.ApplicationEvent) {
    fmt.Println("监听器 2: 发送启动通知")
})

bus.Subscribe(event.EventApplicationStarted, func(e event.ApplicationEvent) {
    fmt.Println("监听器 3: 初始化监控指标")
})

// 发布后三个监听器按订阅顺序依次调用
bus.Publish(&event.BaseEvent{EventType: event.EventApplicationStarted})
```

### 自定义事件类型

```go
type UserRegisteredEvent struct {
    event.BaseEvent
    Username  string
    Email     string
}

func main() {
    bus := event.NewEventBus()

    bus.Subscribe("UserRegistered", func(e event.ApplicationEvent) {
        evt := e.(*UserRegisteredEvent)
        fmt.Printf("用户注册: %s (%s)\n", evt.Username, evt.Email)
    })

    bus.Publish(&UserRegisteredEvent{
        BaseEvent: event.BaseEvent{EventType: "UserRegistered"},
        Username:  "alice",
        Email:     "alice@example.com",
    })
}
```

### 使用事件进行模块解耦

```go
// order 模块
bus.Subscribe("PaymentCompleted", func(e event.ApplicationEvent) {
    evt := e.(*PaymentCompletedEvent)
    fmt.Printf("订单 %s 支付完成，更新订单状态\n", evt.OrderID)
})

// notification 模块
bus.Subscribe("PaymentCompleted", func(e event.ApplicationEvent) {
    evt := e.(*PaymentCompletedEvent)
    fmt.Printf("发送支付成功通知给用户 %s\n", evt.UserID)
})

// analytics 模块
bus.Subscribe("PaymentCompleted", func(e event.ApplicationEvent) {
    evt := e.(*PaymentCompletedEvent)
    fmt.Printf("统计: 订单金额 %.2f\n", evt.Amount)
})
```

## 与 context 包的关系

`DefaultApplicationContext` 内部持有 `EventBus`：

```go
ctx := context.NewApplicationContext(container, env)

ctx.EventBus().Subscribe(event.EventApplicationReady, func(e event.ApplicationEvent) {
    fmt.Println("上下文已就绪")
})

ctx.Start() // 触发 EventApplicationStarted 和 EventApplicationReady
ctx.Stop()  // 触发 EventApplicationStopped
```

---

## 使用场景

### 场景 1：应用生命周期管理

**描述**：监听应用启动、就绪、停止等生命周期事件，执行相应的初始化和清理操作。

```go
bus := event.NewEventBus()

bus.Subscribe(event.EventApplicationStarted, func(e event.ApplicationEvent) {
    fmt.Println("应用启动中，初始化资源...")
})

bus.Subscribe(event.EventApplicationReady, func(e event.ApplicationEvent) {
    fmt.Println("应用已就绪，开始处理请求...")
})

bus.Subscribe(event.EventApplicationStopped, func(e event.ApplicationEvent) {
    fmt.Println("应用停止中，清理资源...")
})
```

**最佳实践**：
- 在应用启动时订阅生命周期事件
- 使用事件进行资源初始化和清理
- 避免在监听器中执行耗时操作

### 场景 2：模块间解耦

**描述**：使用事件实现模块间松耦合通信，避免直接依赖。

```go
// 订单模块
bus.Subscribe("PaymentCompleted", func(e event.ApplicationEvent) {
    evt := e.(*PaymentCompletedEvent)
    fmt.Printf("订单 %s 支付完成，更新订单状态\n", evt.OrderID)
})

// 通知模块
bus.Subscribe("PaymentCompleted", func(e event.ApplicationEvent) {
    evt := e.(*PaymentCompletedEvent)
    fmt.Printf("发送支付成功通知给用户 %s\n", evt.UserID)
})

// 分析模块
bus.Subscribe("PaymentCompleted", func(e event.ApplicationEvent) {
    evt := e.(*PaymentCompletedEvent)
    fmt.Printf("统计: 订单金额 %.2f\n", evt.Amount)
})
```

**最佳实践**：
- 使用事件实现发布-订阅模式
- 每个模块独立订阅感兴趣的事件
- 避免在事件处理中引入循环依赖

### 场景 3：审计日志

**描述**：记录关键业务操作的事件日志，用于审计和追溯。

```go
bus.Subscribe("UserCreated", func(e event.ApplicationEvent) {
    evt := e.(*UserCreatedEvent)
    log.Printf("用户创建: ID=%s, Username=%s, Time=%s",
        evt.UserID, evt.Username, evt.Timestamp())
})

bus.Subscribe("OrderPlaced", func(e event.ApplicationEvent) {
    evt := e.(*OrderPlacedEvent)
    log.Printf("订单创建: OrderID=%s, UserID=%s, Amount=%.2f",
        evt.OrderID, evt.UserID, evt.Amount)
})
```

**最佳实践**：
- 使用事件记录关键业务操作
- 包含足够的上下文信息
- 异步处理审计日志，避免影响业务性能

### 场景 4：缓存失效

**描述**：当数据变更时，通过事件通知相关模块清理缓存。

```go
bus.Subscribe("DataUpdated", func(e event.ApplicationEvent) {
    evt := e.(*DataUpdatedEvent)
    cacheKey := fmt.Sprintf("%s:%s", evt.DataType, evt.DataID)
    cache.Delete(cacheKey)
    fmt.Printf("缓存已清理: %s\n", cacheKey)
})
```

**最佳实践**：
- 使用事件触发缓存清理
- 确保缓存清理的幂等性
- 考虑使用事件版本号避免重复处理