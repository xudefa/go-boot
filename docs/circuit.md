# circuit 包 — 熔断器实现

## 概述

`circuit` 包提供了熔断器（Circuit Breaker）实现，用于防止级联故障。当后端服务不可用时，熔断器可以快速失败，避免系统资源被耗尽。

## 核心概念

### 熔断器状态

熔断器有三种状态：

| 状态 | 常量 | 说明 |
|------|------|------|
| Closed | `StateClosed` | 关闭状态，正常处理请求 |
| Open | `StateOpen` | 打开状态，快速失败 |
| HalfOpen | `StateHalfOpen` | 半开状态，尝试恢复 |

### 状态转换

```
Closed ──(失败率超过阈值)──> Open
Open ──(等待时间过后)──> HalfOpen
HalfOpen ──(请求成功)──> Closed
HalfOpen ──(请求失败)──> Open
```

## 接口定义

### Breaker 接口

```go
type Breaker interface {
    // Allow 检查是否允许请求通过
    Allow() error
    
    // RecordSuccess 记录成功请求
    RecordSuccess()
    
    // RecordFailure 记录失败请求
    RecordFailure()
    
    // State 获取当前状态
    State() State
}
```

## 使用示例

### 基本使用

```go
package main

import (
    "time"
    "github.com/xudefa/go-boot/circuit"
)

func main() {
    // 创建熔断器
    breaker := circuit.NewDefaultBreaker(
        circuit.WithErrorThreshold(0.5),      // 错误率阈值 50%
        circuit.WithWaitDuration(30*time.Second), // 等待时间 30 秒
        circuit.WithMaxRequests(10),          // 半开状态最大请求数 10
    )
    
    // 使用熔断器保护请求
    if err := breaker.Allow(); err != nil {
        // 熔断器打开，快速失败
        return
    }
    
    // 执行实际请求
    err := doSomething()
    
    // 记录结果
    if err != nil {
        breaker.RecordFailure()
    } else {
        breaker.RecordSuccess()
    }
}
```

### 配置选项

| 选项 | 默认值 | 说明 |
|------|--------|------|
| `WithMaxRequests(max int)` | 10 | 半开状态下允许的最大请求数 |
| `WithErrorThreshold(threshold float64)` | 0.5 | 错误率阈值（0.0-1.0） |
| `WithWaitDuration(duration time.Duration)` | 30s | 打开状态等待时间 |

## 错误类型

| 错误 | 说明 |
|------|------|
| `ErrCircuitOpen` | 熔断器打开，请求被拒绝 |
| `ErrCircuitHalfOpen` | 熔断器半开，达到最大请求数 |

## 注意事项

1. **线程安全**: `DefaultBreaker` 是线程安全的，可以在多个 goroutine 中并发使用
2. **错误率计算**: 基于滑动窗口计算错误率，窗口大小由 `maxRequests` 控制
3. **状态恢复**: 在 HalfOpen 状态下，需要连续成功 `maxRequests` 次才能恢复到 Closed 状态