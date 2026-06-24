# center 包 — 服务注册发现与负载均衡

## 概述

`center` 包提供服务注册与发现的抽象接口，以及内置的负载均衡选择器。核心组件：

- `Registry` — 注册中心抽象接口（Register / Deregister / Discover / Watch）
- `Selector` — 负载均衡选择器接口（Select）
- `InstanceInfo` — 服务实例信息结构体
- 内置选择器：随机、轮询、加权随机
- Sentinel 错误：`ErrNoInstances`、`ErrInvalidInstance`

---

## InstanceInfo — 服务实例信息

```go
type InstanceInfo struct {
    ServiceName string
    ID          string
    Host        string
    Port        int
    Weight      int
    Healthy     bool
    Metadata    map[string]string
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `ServiceName` | `string` | 服务名称，如 `"user-service"` |
| `ID` | `string` | 实例唯一标识，如 `"192.168.1.1:8080"` |
| `Host` | `string` | 实例主机地址 |
| `Port` | `int` | 实例端口号 |
| `Weight` | `int` | 权重，用于加权选择策略 |
| `Healthy` | `bool` | 健康状态 |
| `Metadata` | `map[string]string` | 附加元数据（版本号、区域等） |

---

## Registry — 注册中心接口

```go
type Registry interface {
    Register(ctx context.Context, info InstanceInfo) error
    Deregister(ctx context.Context, info InstanceInfo) error
    Discover(ctx context.Context, serviceName string) ([]InstanceInfo, error)
    Watch(ctx context.Context, serviceName string) (<-chan []InstanceInfo, error)
}
```

| 方法 | 说明 |
|------|------|
| `Register` | 注册一个服务实例，返回错误表示注册失败 |
| `Deregister` | 注销一个服务实例 |
| `Discover` | 发现指定服务的所有实例 |
| `Watch` | 监听指定服务的实例变化，通过 channel 推送最新列表 |

### 使用示例

```go
var reg center.Registry // 具体实现由注册中心提供（如 etcd、consul、nacos）

// 注册实例
err := reg.Register(ctx, center.InstanceInfo{
    ServiceName: "user-service",
    ID:          "192.168.1.1:8080",
    Host:        "192.168.1.1",
    Port:        8080,
    Weight:      10,
    Healthy:     true,
    Metadata:    map[string]string{"version": "v1.0"},
})

// 发现服务
instances, err := reg.Discover(ctx, "user-service")

// 监听变化
ch, err := reg.Watch(ctx, "user-service")
for instances := range ch {
    // 处理最新的实例列表
}

// 注销实例
err := reg.Deregister(ctx, center.InstanceInfo{
    ServiceName: "user-service",
    ID:          "192.168.1.1:8080",
})
```

---

## Selector — 负载均衡选择器

```go
type Selector interface {
    Select(instances []InstanceInfo) (InstanceInfo, error)
}
```

从一组服务实例中按策略选出一个目标实例。传入空列表时返回 `ErrNoInstances`。

### 内置选择器

#### RandomSelect — 随机选择

```go
var RandomSelect Selector = &randomSelector{...}
```

从实例列表中随机选取一个，内部使用独立的 `math/rand/v2` 实例并通过 Mutex 保证并发安全。

```go
inst, err := center.RandomSelect.Select(instances)
if err != nil {
    // 处理无可用实例
}
```

#### RoundRobinSelect — 轮询选择

```go
var RoundRobinSelect Selector = &roundRobinSelector{...}
```

按顺序依次选取实例，内部使用原子计数器（`atomic.Uint64`）保证并发安全，无锁开销。

```go
inst1, _ := center.RoundRobinSelect.Select(instances) // 返回第一个
inst2, _ := center.RoundRobinSelect.Select(instances) // 返回第二个
```

#### WeightedRandomSelect — 加权随机选择

```go
func NewWeightedRandomSelect() Selector
```

权重越高的实例被选中的概率越大。实例的 `Weight <= 0` 时按 1 处理。

算法：
1. 计算所有实例的总权重
2. 在 `[0, total)` 范围内生成随机数
3. 遍历实例，累减权重直到找到目标实例

```go
sel := center.NewWeightedRandomSelect()

instances := []center.InstanceInfo{
    {Host: "192.168.1.1", Port: 8080, Weight: 10},  // 10/20 = 50%
    {Host: "192.168.1.2", Port: 8080, Weight: 5},   // 5/20  = 25%
    {Host: "192.168.1.3", Port: 8080, Weight: 5},   // 5/20  = 25%
}

inst, err := sel.Select(instances)
```

---

## Sentinel 错误

```go
var ErrNoInstances    = errors.New("center: no available instances")
var ErrInvalidInstance = errors.New("center: invalid instance info")
```

| 错误 | 说明 |
|------|------|
| `ErrNoInstances` | 选择器传入空实例列表时返回 |
| `ErrInvalidInstance` | 实例信息无效时返回 |

```go
inst, err := sel.Select(nil)
if errors.Is(err, center.ErrNoInstances) {
    log.Println("无可用实例")
}
```

---

## 完整使用示例

### 服务发现 + 负载均衡

```go
package main

import (
    "context"
    "fmt"
    "github.com/xudefa/go-boot/center"
)

func callService(ctx context.Context, reg center.Registry, sel center.Selector, serviceName string) error {
    // 发现服务实例
    instances, err := reg.Discover(ctx, serviceName)
    if err != nil {
        return err
    }

    // 选择一个实例
    inst, err := sel.Select(instances)
    if err != nil {
        return err
    }

    fmt.Printf("调用服务: %s -> %s:%d\n", serviceName, inst.Host, inst.Port)
    return nil
}
```

### 自定义选择器

实现 `Selector` 接口即可自定义负载均衡策略：

```go
type HashSelector struct {
    key string
}

func (s *HashSelector) Select(instances []center.InstanceInfo) (center.InstanceInfo, error) {
    if len(instances) == 0 {
        return center.InstanceInfo{}, center.ErrNoInstances
    }
    // 基于一致性哈希选择实例
    idx := hash(s.key) % len(instances)
    return instances[idx], nil
}
```

---

## 完整 API 参考

| 类型/函数 | 说明 |
|-----------|------|
| `InstanceInfo` | 服务实例信息结构体 |
| `Registry` | 注册中心接口（Register / Deregister / Discover / Watch） |
| `Selector` | 负载均衡选择器接口 |
| `RandomSelect` | 随机选择器（全局变量） |
| `RoundRobinSelect` | 轮询选择器（全局变量） |
| `NewWeightedRandomSelect()` | 创建加权随机选择器 |
| `ErrNoInstances` | 无可用实例错误 |
| `ErrInvalidInstance` | 实例信息无效错误 |