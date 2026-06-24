# loadbalancer 包 — 负载均衡器实现

## 概述

`loadbalancer` 包提供了多种负载均衡策略实现，用于在服务实例之间分配请求。支持轮询、随机、加权、最少连接、一致性哈希、会话保持、IP 哈希、健康感知、响应时间加权和自适应权重等策略。

## 核心概念

### 服务实例

每个后端服务由 `ServiceInstance` 表示：

```go
type ServiceInstance struct {
    ID       string            // 服务实例 ID
    URL      string            // 服务地址
    Weight   int               // 权重
    Metadata map[string]string // 元数据
    Health   HealthStatus      // 健康状态
    Active   int64             // 活跃连接数
}
```

### 健康状态

```go
const (
    HealthUp      HealthStatus = iota // 健康
    HealthDown                        // 不健康
    HealthUnknown                     // 未知
)
```

## 接口定义

### Balancer 接口

```go
type Balancer interface {
    // Next 选择下一个服务实例
    Next(backends []*ServiceInstance) (*ServiceInstance, error)
}
```

## 负载均衡策略

### 1. RoundRobin — 轮询

按顺序依次选择后端服务实例。

```go
lb := loadbalancer.NewRoundRobin()
backend, err := lb.Next(backends)
```

**特点**:
- 简单公平，每个后端被均匀选择
- 适合后端性能相近的场景

### 2. Random — 随机

随机选择一个后端服务实例。

```go
lb := loadbalancer.NewRandom()
backend, err := lb.Next(backends)
```

**特点**:
- 实现简单，无状态
- 在大量请求下分布均匀

### 3. WeightedRoundRobin — 加权轮询

根据后端权重进行选择，权重高的后端被选中概率更大。

```go
lb := loadbalancer.NewWeightedRoundRobin()
backends := []*loadbalancer.ServiceInstance{
    {ID: "1", URL: "http://localhost:8081", Weight: 3},
    {ID: "2", URL: "http://localhost:8082", Weight: 1},
}
backend, err := lb.Next(backends)
```

**特点**:
- 适合后端性能不一致的场景
- 高性能实例可以处理更多请求

### 4. LeastConnections — 最少连接

选择当前活跃连接数最少的后端。

```go
lb := loadbalancer.NewLeastConnections()
backend, err := lb.Next(backends)
```

**特点**:
- 动态感知后端负载
- 适合请求处理时间差异大的场景

### 5. ConsistentHash — 一致性哈希

根据键的哈希值选择后端，相同键总是选择相同的后端。

```go
lb := loadbalancer.NewConsistentHash(150) // 150 个虚拟节点

// 使用默认键
backend, err := lb.Next(backends)

// 使用指定键
backend, err := lb.NextByKey(backends, "user-123")
```

**特点**:
- 相同请求路由到相同后端
- 适合需要会话亲和性或缓存的场景
- 后端增减时影响最小

### 6. StickySession — 会话保持

根据会话 ID 保持会话亲和性。

```go
lb := loadbalancer.NewStickySession("SESSIONID")

// 根据会话 ID 选择后端
backend, err := lb.NextWithSession(backends, "session-123")

// 获取会话绑定的后端
backend, exists := lb.GetSessionBackend("session-123")

// 移除会话绑定
lb.RemoveSession("session-123")
```

**特点**:
- 同一会话的请求路由到相同后端
- 适合有状态应用

### 7. IPHash — IP 哈希

根据客户端 IP 的哈希值选择后端。

```go
lb := loadbalancer.NewIPHash()
backend, err := lb.NextByIP(backends, "192.168.1.100")
```

**特点**:
- 相同 IP 的请求路由到相同后端
- 适合需要 IP 亲和性的场景

### 8. HealthAware — 健康感知

包装其他负载均衡器，只选择健康的后端。

```go
inner := loadbalancer.NewRoundRobin()
lb := loadbalancer.NewHealthAware(inner)

// 记录后端失败
lb.RecordFailure("http://localhost:8081")

// 记录后端成功
lb.RecordSuccess("http://localhost:8081")

// 获取失败计数
count := lb.GetFailureCount("http://localhost:8081")
```

**特点**:
- 自动过滤不健康后端
- 可与其他策略组合使用

### 9. ResponseTimeWeighted — 响应时间加权

根据后端响应时间动态调整权重，响应快的后端被选中概率更高。

```go
lb := loadbalancer.NewResponseTimeWeighted(0.9) // 衰减因子

// 记录响应时间（毫秒）
lb.RecordResponseTime("http://localhost:8081", 50)

// 获取平均响应时间
avgTime, exists := lb.GetAvgResponseTime("http://localhost:8081")

// 选择后端
backend, err := lb.Next(backends)
```

**特点**:
- 自动适应后端性能变化
- 使用衰减因子平滑历史数据

### 10. AdaptiveWeight — 自适应权重

综合考虑错误率、响应时间和连接数，动态计算后端权重。

```go
lb := loadbalancer.NewAdaptiveWeight()

// 记录请求统计
lb.RecordRequest("http://localhost:8081", 50, false)  // 响应时间 50ms，成功
lb.RecordRequest("http://localhost:8082", 200, true)  // 响应时间 200ms，失败

// 记录连接数变化
lb.RecordConnection("http://localhost:8081", 1)   // 连接 +1
lb.RecordConnection("http://localhost:8081", -1)  // 连接 -1

// 获取统计信息
stats, exists := lb.GetStats("http://localhost:8081")
allStats := lb.GetAllStats()
```

**特点**:
- 多维度评估后端性能
- 自动适应后端变化

## 辅助函数

### SortByResponseTime

按响应时间排序后端实例。

```go
stats := lb.GetAllStats()
sorted := loadbalancer.SortByResponseTime(backends, stats)
```

## 错误处理

```go
var ErrNoBackends = errors.New("no available backends")
```

当没有可用的后端时返回此错误。

## 使用示例

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/xudefa/go-boot/loadbalancer"
)

func main() {
    // 创建负载均衡器
    lb := loadbalancer.NewRoundRobin()
    
    // 定义后端
    backends := []*loadbalancer.ServiceInstance{
        {ID: "1", URL: "http://localhost:8081", Weight: 1},
        {ID: "2", URL: "http://localhost:8082", Weight: 1},
        {ID: "3", URL: "http://localhost:8083", Weight: 1},
    }
    
    // 选择后端
    backend, err := lb.Next(backends)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    
    fmt.Println("Selected:", backend.URL)
}
```

### 与健康检查结合

```go
// 创建健康感知的负载均衡器
inner := loadbalancer.NewWeightedRoundRobin()
lb := loadbalancer.NewHealthAware(inner)

// 模拟健康检查
for _, b := range backends {
    if isHealthy(b.URL) {
        b.Health = loadbalancer.HealthUp
    } else {
        b.Health = loadbalancer.HealthDown
    }
}

// 选择健康后端
backend, err := lb.Next(backends)
```

### 与监控结合

```go
// 创建自适应权重负载均衡器
lb := loadbalancer.NewAdaptiveWeight()

// 在请求完成后记录统计
start := time.Now()
err := sendRequest(backend.URL)
duration := time.Since(start).Milliseconds()

if err != nil {
    lb.RecordRequest(backend.URL, float64(duration), true)
} else {
    lb.RecordRequest(backend.URL, float64(duration), false)
}
```

## 线程安全

所有负载均衡器实现都是线程安全的，可以在多个 goroutine 中并发使用。

## 注意事项

1. **一致性哈希**: `ConsistentHash` 的 `Next()` 方法使用默认键，建议使用 `NextByKey()` 方法以获得更好的一致性
2. **会话保持**: `StickySession` 会维护会话到后端的映射，注意内存使用
3. **健康感知**: `HealthAware` 需要外部健康检查机制来更新 `Health` 字段
4. **权重配置**: `WeightedRoundRobin` 需要正确配置权重值，权重为 0 时按轮询处理