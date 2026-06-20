# context 包 — 应用上下文

## 概述

`context` 包提供应用上下文（ApplicationContext），聚合了 go-boot 框架的四个核心子系统：

- **Container** — IoC 依赖注入容器
- **Environment** — 分层配置源管理
- **Lifecycle** — 应用生命周期阶段管理
- **EventBus** — 事件发布与订阅

应用上下文是框架的核心运行时入口，所有组件通过它进行交互。

## ApplicationContext 接口

```go
type ApplicationContext interface {
    Container() core.Container
    Environment() *environment.Environment
    Lifecycle() *life.LifecycleManager
    EventBus() *event.EventBus
    EventPublisher() EventPublisher

    Register(name string, opts ...core.BuilderOption) error
    Get(name string) (any, error)
    Invoke(fn any, opts ...core.InvokeOption) error

    Start() error
    Stop() error
    IsRunning() bool
}

// EventPublisher 事件发布器接口（解耦事件发布逻辑）
type EventPublisher interface {
    Publish(event event.ApplicationEvent)
}
```

### 方法说明

| 方法 | 说明 |
|------|------|
| `Container()` | 返回 IoC 容器，管理 Bean 的注册和获取 |
| `Environment()` | 返回环境配置，支持多级配置源 |
| `Lifecycle()` | 返回生命周期管理器，控制应用状态流转 |
| `EventBus()` | 返回事件总线，支持事件的发布与订阅 |
| `EventPublisher()` | 返回事件发布器接口，解耦事件发布逻辑，便于测试和替换实现 |
| `Register(name, opts...)` | 在容器中注册 Bean |
| `Get(name)` | 从容器中获取指定名称的 Bean |
| `Invoke(fn, opts...)` | 调用函数并自动注入依赖参数 |
| `Start()` | 启动应用，发布启动事件并切换至运行阶段 |
| `Stop()` | 停止应用，切换至停止阶段并发布停止事件 |
| `IsRunning()` | 检查应用是否处于运行状态 |

## DefaultApplicationContext

`DefaultApplicationContext` 是 `ApplicationContext` 的默认实现。

```go
type DefaultApplicationContext struct {
    container core.Container
    env       *environment.Environment
    lifecycle *life.LifecycleManager
    events    *event.EventBus
}
```

### 创建实例

```go
func NewApplicationContext(container core.Container, env *environment.Environment) *DefaultApplicationContext
```

创建时自动初始化：

- `lifecycle` — 创建一个新的 `LifecycleManager`（初始阶段为 `PhaseInitializing`）
- `events` — 创建一个新的 `EventBus`

### 启动与停止

`Start()` 流程：

1. 发布 `EventApplicationStarted`
2. 设置生命周期阶段为 `PhaseRunning`
3. 发布 `EventApplicationReady`

```go
func (c *DefaultApplicationContext) Start() error {
    c.events.Publish(&event.BaseEvent{EventType: event.EventApplicationStarted})
    if err := c.lifecycle.SetPhase(life.PhaseRunning); err != nil {
        return err
    }
    c.events.Publish(&event.BaseEvent{EventType: event.EventApplicationReady})
    return nil
}
```

`Stop()` 流程：

1. 设置生命周期阶段为 `PhaseStopping`
2. 设置生命周期阶段为 `PhaseStopped`
3. 发布 `EventApplicationStopped`

```go
func (c *DefaultApplicationContext) Stop() error {
    if err := c.lifecycle.SetPhase(life.PhaseStopping); err != nil {
        return err
    }
    if err := c.lifecycle.SetPhase(life.PhaseStopped); err != nil {
        return err
    }
    c.events.Publish(&event.BaseEvent{EventType: event.EventApplicationStopped})
    return nil
}
```

### 辅助方法

```go
func (c *DefaultApplicationContext) GetBean(beanID string) (any, bool)
func (c *DefaultApplicationContext) HasProperty(key string) bool
func (c *DefaultApplicationContext) GetProperty(key string) (any, bool)
func (c *DefaultApplicationContext) ClassLoader() interface{ HasClass(name string) bool }
```

## ClassLoader 缓存优化

`buildInfoClassLoader` 是 `ClassLoader` 的实现，使用 `runtime/debug.ReadBuildInfo()` 检查模块是否在编译依赖中。

### 优化策略

- **sync.Once 延迟初始化**：首次调用 `HasClass` 时读取构建信息，避免启动时不必要的开销
- **依赖列表缓存**：将构建信息中的依赖模块列表缓存到内存，避免每次调用都重复读取和解析
- **全局共享实例**：使用 `globalClassLoader` 全局共享实例，提升性能并减少重复初始化

### 实现细节

```go
// 全局共享 ClassLoader 实例
var globalClassLoader = &buildInfoClassLoader{}

type buildInfoClassLoader struct {
    once     sync.Once
    deps     []string // 缓存的依赖路径列表
    mainPath string
    err      error
}

func (b *buildInfoClassLoader) HasClass(name string) bool {
    b.once.Do(b.init)
    if b.err != nil {
        return false
    }

    pkgPath := extractPkgPath(name)
    if pkgPath == "" {
        return false
    }

    // 先检查主模块
    if b.mainPath != "" && pathMatches(b.mainPath, pkgPath) {
        return true
    }

    // 再检查缓存的依赖列表
    for _, dep := range b.deps {
        if pathMatches(dep, pkgPath) {
            return true
        }
    }
    return false
}
```

### 性能提升

- **首次调用**：读取构建信息（约 1-5ms）
- **后续调用**：直接查询缓存（约 10-100ns）
- **内存开销**：每个依赖模块约 50 字节字符串，100 个依赖约 5KB

## 使用示例

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/xudefa/go-boot/context"
    "github.com/xudefa/go-boot/core"
    "github.com/xudefa/go-boot/environment"
)

func main() {
    container := core.New()
    env := environment.NewEnvironment()
    ctx := context.NewApplicationContext(container, env)

    // 注册 Bean
    _ = ctx.Register("myService",
        core.Bean(&MyService{Name: "hello"}),
        core.Singleton(),
    )

    // 获取 Bean
    svc, _ := ctx.Get("myService")
    fmt.Println(svc.(*MyService).Name)

    // 启动
    ctx.Start()
    defer ctx.Stop()
}

type MyService struct {
    Name string
}
```

### 事件订阅

```go
ctx.EventBus().Subscribe(event.EventApplicationStarted, func(e event.ApplicationEvent) {
    fmt.Println("应用已启动，时间:", e.Timestamp())
})

ctx.Start()
```

### 生命周期监听

```go
ctx.Lifecycle().AddListener(&myPhaseListener{})

ctx.Start() // 触发 PhaseRunning 变更通知
```

### 方法注入

```go
_ = ctx.Invoke(func(svc *MyService) {
    fmt.Println("注入的 service:", svc.Name)
})
```

## 四子系统集成关系

```
┌──────────────────────────────────────────────────────┐
│              DefaultApplicationContext               │
│                                                      │
│  ┌──────────────┐  ┌──────────────────────────────┐  │
│  │   Container  │  │        Environment           │  │
│  │  (core.Cont) │  │  (environment.Environment)   │  │
│  │              │  │                              │  │
│  │  Register()  │  │  GetProperty()               │  │
│  │  Get()       │  │  AddPropertySource()         │  │
│  │  Invoke()    │  │  GetActiveProfiles()         │  │
│  └──────────────┘  └──────────────────────────────┘  │
│                                                      │
│  ┌──────────────┐  ┌──────────────────────────────┐  │
│  │  Lifecycle   │  │          EventBus            │  │
│  │  (life.Man)  │  │      (event.EventBus)        │  │
│  │              │  │                              │  │
│  │  SetPhase()  │  │  Publish()                   │  │
│  │  GetPhase()  │  │  Subscribe()                 │  │
│  │  AddListen() │  │  Unsubscribe()               │  │
│  └──────────────┘  └──────────────────────────────┘  │
└──────────────────────────────────────────────────────┘
```

## 与 Boot 的关系

`boot.Boot` 在内部持有 `DefaultApplicationContext` 并进行更细粒度的生命周期控制：

- `Boot.Start()` 在 `PhaseConfiguring` → `PhaseReady` 之间插入自动配置执行、启动器配置等步骤
- `DefaultApplicationContext.Start()` 仅处理 `PhaseRunning` 阶段切换和事件发布

`DefaultApplicationContext` 同时实现了 `condition.ConditionContext` 所需的辅助方法（`GetBean`、`HasProperty`、`GetProperty`），通过 `conditionCtx` 适配器供条件系统使用。