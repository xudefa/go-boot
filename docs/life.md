# life 包 — 生命周期管理

## 概述

`life` 包定义应用的完整生命周期阶段和阶段变更监听机制。生命周期按固定顺序正向流转，仅允许向前转换，确保应用状态的可预测性。

```
PhaseInitializing
    ↓
PhaseConfiguring
    ↓
PhaseContextRefreshed
    ↓
PhaseReady
    ↓
PhaseRunning
    ↓
PhaseStopping
    ↓
PhaseStopped
```

## ApplicationPhase

```go
type ApplicationPhase int
```

### 阶段定义

| 常量 | 值 | 说明 |
|------|----|------|
| `PhaseInitializing` | 0 | 初始化阶段：创建容器和基础组件 |
| `PhaseConfiguring` | 1 | 配置阶段：加载配置、注册 Bean |
| `PhaseContextRefreshed` | 2 | 上下文刷新完成：所有 Bean 已注册 |
| `PhaseReady` | 3 | 就绪阶段：应用准备就绪但尚未开始服务 |
| `PhaseRunning` | 4 | 运行阶段：应用正常运行，处理请求 |
| `PhaseStopping` | 5 | 停止阶段：应用正在停止，释放资源 |
| `PhaseStopped` | 6 | 已停止：应用完全停止 |

### 阶段流转约束

```go
func isForwardTransition(oldPhase, newPhase ApplicationPhase) bool
```

流转规则：

- `newPhase > oldPhase`：新阶段的值必须大于旧阶段
- `oldPhase >= PhaseInitializing`：从有效的初始阶段开始
- `newPhase <= PhaseStopped`：新阶段不能超出已停止阶段

任何违反正向流转规则的转换都会返回错误：

```go
err := manager.SetPhase(PhaseConfiguring)  // 可以：PhaseInitializing → PhaseConfiguring
err := manager.SetPhase(PhaseInitializing) // 错误：不能回退
err := manager.SetPhase(PhaseStopped)      // 可以：可跳转到任意后续阶段
```

### 阶段名称

每个阶段有对应的字符串表示，用于日志和错误输出：

```go
phaseNames[PhaseInitializing]     = "INITIALIZING"
phaseNames[PhaseConfiguring]      = "CONFIGURING"
phaseNames[PhaseContextRefreshed] = "CONTEXT_REFRESHED"
phaseNames[PhaseReady]            = "READY"
phaseNames[PhaseRunning]          = "RUNNING"
phaseNames[PhaseStopping]         = "STOPPING"
phaseNames[PhaseStopped]          = "STOPPED"
```

## PhaseListener

```go
type PhaseListener interface {
    OnPhaseChange(oldPhase, newPhase ApplicationPhase) error
}
```

`PhaseListener` 允许外部组件监听阶段变更事件。当生命周期发生阶段转换时，所有注册的监听器会收到 `OnPhaseChange` 回调，包含旧阶段和新阶段信息。

## LifecycleManager

```go
type LifecycleManager struct {
    mu        sync.RWMutex
    phase     ApplicationPhase
    listeners []PhaseListener
}
```

### 创建

```go
func NewLifecycleManager() *LifecycleManager
```

初始阶段为 `PhaseInitializing`。

### 方法

| 方法 | 说明 |
|------|------|
| `GetPhase()` | 返回当前阶段 |
| `SetPhase(newPhase)` | 设置新阶段，验证转换合法性，通知监听器 |
| `AddListener(listener)` | 添加阶段监听器 |

### SetPhase 流程

1. 加写锁，获取旧阶段
2. 验证 `isForwardTransition(oldPhase, newPhase)`
3. 更新当前阶段
4. 拷贝监听器列表（读快照）
5. 释放写锁
6. 遍历监听器，逐个调用 `OnPhaseChange`
7. 返回第一个错误（如果有多个错误，仅返回第一个）

```go
func (m *LifecycleManager) SetPhase(newPhase ApplicationPhase) error {
    m.mu.Lock()
    oldPhase := m.phase
    if !isForwardTransition(oldPhase, newPhase) {
        m.mu.Unlock()
        return fmt.Errorf("invalid phase transition from %s to %s", oldPhase, newPhase)
    }
    m.phase = newPhase
    listeners := make([]PhaseListener, len(m.listeners))
    copy(listeners, m.listeners)
    m.mu.Unlock()

    for _, listener := range listeners {
        if e := listener.OnPhaseChange(oldPhase, newPhase); e != nil {
            return e
        }
    }
    return nil
}
```

## 使用示例

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/xudefa/go-boot/life"
)

func main() {
    manager := life.NewLifecycleManager()

    fmt.Println(manager.GetPhase()) // INITIALIZING

    _ = manager.SetPhase(life.PhaseConfiguring)
    fmt.Println(manager.GetPhase()) // CONFIGURING

    _ = manager.SetPhase(life.PhaseRunning)
    fmt.Println(manager.GetPhase()) // RUNNING

    _ = manager.SetPhase(life.PhaseStopping)
    _ = manager.SetPhase(life.PhaseStopped)
}
```

### 非法转换

```go
manager := life.NewLifecycleManager()

// 尝试回退到初始化阶段
err := manager.SetPhase(life.PhaseInitializing)
fmt.Println(err) // invalid phase transition from INITIALIZING to RUNNING

// 先前进到运行阶段
_ = manager.SetPhase(life.PhaseRunning)

// 再回退
err = manager.SetPhase(life.PhaseConfiguring)
fmt.Println(err) // invalid phase transition from RUNNING to CONFIGURING
```

### 阶段监听

```go
type StartupLogger struct{}

func (l *StartupLogger) OnPhaseChange(oldPhase, newPhase life.ApplicationPhase) error {
    fmt.Printf("阶段变更: %s → %s\n", oldPhase, newPhase)
    return nil
}

type StartupGuard struct{}

func (g *StartupGuard) OnPhaseChange(oldPhase, newPhase life.ApplicationPhase) error {
    if newPhase == life.PhaseRunning {
        fmt.Println("应用已进入运行阶段，开始接受请求")
    }
    return nil
}

func main() {
    manager := life.NewLifecycleManager()
    manager.AddListener(&StartupLogger{})
    manager.AddListener(&StartupGuard{})

    _ = manager.SetPhase(life.PhaseConfiguring)
    _ = manager.SetPhase(life.PhaseContextRefreshed)
    _ = manager.SetPhase(life.PhaseReady)
    _ = manager.SetPhase(life.PhaseRunning)
    _ = manager.SetPhase(life.PhaseStopping)
    _ = manager.SetPhase(life.PhaseStopped)
}
```

## 在 Boot 中的使用

`Boot.Start()` 和 `Boot.Stop()` 对生命周期阶段有精确控制，确保自动配置、Starter 等组件在正确的阶段执行：

```
Boot.Start():
  1. SetPhase(PhaseConfiguring)          → 执行 AutoConfiguration
  2. SetPhase(PhaseContextRefreshed)     → 上下文就绪
  3. SetPhase(PhaseReady)                → Starter 就绪
  4. SetPhase(PhaseRunning)              → 正式运行

Boot.Stop():
  1. SetPhase(PhaseStopping)             → 开始停止
  2. SetPhase(PhaseStopped)              → 完全停止
```

## 与 context 包的关系

`DefaultApplicationContext` 内部持有 `LifecycleManager`：

```go
ctx := context.NewApplicationContext(container, env)
ctx.Lifecycle().SetPhase(life.PhaseRunning)
fmt.Println(ctx.Lifecycle().GetPhase()) // RUNNING
ctx.IsRunning() // true
```

`IsRunning()` 直接检查当前阶段是否为 `PhaseRunning`：

```go
func (c *DefaultApplicationContext) IsRunning() bool {
    return c.lifecycle.GetPhase() == life.PhaseRunning
}
```
