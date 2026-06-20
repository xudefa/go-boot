# actuator 包 — 运维端点

## 概述

`actuator` 包提供类似 Spring Boot Actuator 的运维端点支持，便于在生产环境中监控和管理应用。内置端点包括健康检查、指标收集、环境信息和 Bean 列表。

## 端点列表

| 路径 | 处理器 | 说明 |
|------|--------|------|
| `/actuator/health` | `HealthHandler` | 聚合健康指标，返回整体健康状态 |
| `/actuator/metrics` | `MetricsHandler` | 收集并返回所有注册的指标 |
| `/actuator/env` | `EnvHandler` | 显示环境配置源及其属性 |
| `/actuator/beans` | `BeansHandler` | 列出 IoC 容器中注册的所有 Bean |

## Actuator 结构体

```go
type Actuator struct {
    healthAggregator *health.Aggregator
    metricsRegistry  metrics.MeterRegistry
    appContext       AppContext
    sanitizer        *Sanitizer
}
```

### AppContext 接口

```go
type AppContext interface {
    Container() core.Container
    Environment() *environment.Environment
}
```

同时被 `context.ApplicationContext` 和 `boot.ApplicationContext` 满足，避免 actuator 直接依赖 boot 或 context 包。

## Sanitizer（敏感信息检测器）

`Sanitizer` 使用策略模式管理敏感信息检测规则，支持自定义检测策略。

### 默认检测规则

- **关键词检测**：password, secret, token, key, auth, credential, private, api_key, access_token, client_secret, oauth, bearer, jwt
- **值格式检测**：私钥格式、JWT 令牌、长随机字符串（包含大小写字母、数字和特殊字符）

### 自定义检测策略

```go
sanitizer := actuator.NewSanitizer()

// 添加自定义策略
sanitizer.AddStrategy(&myCustomStrategy{})

// 掩盖敏感值
value := sanitizer.Sanitize("db.password", "secret123")
// 返回: "***REDACTED***"
```

### SanitizeStrategy 接口

```go
type SanitizeStrategy interface {
    IsSensitive(key string, value any) bool
}
```

## RouteRegistrar（路由注册器）

`RouteRegistrar` 接口解耦路由注册逻辑，支持不同的 HTTP 框架。

### 使用示例

```go
registrar := &actuator.StdRouteRegistrar{Mux: mux}
config := actuator.DefaultRouteConfig()
config.ExposeDebug = true

actuator.RegisterRoutes(registrar, config)
```

### RouteConfig 配置

```go
type RouteConfig struct {
    BasePath    string  // 基础路径，默认 "/actuator"
    ExposeDebug bool    // 是否暴露调试端点（pprof）
    Prefix      string  // 路径前缀
}
```

## HealthIndicatorBuilder（健康指示器构建器）

使用 Builder 模式简化健康指示器的创建。

### 使用示例

```go
indicator := actuator.NewHealthIndicatorBuilder().
    Name("database").
    CheckFunc(db.Check).
    Timeout(5 * time.Second).
    Detail("type", "postgres").
    Build()
```

### 内置方法

| 方法 | 说明 |
|------|------|
| `Name(name string)` | 设置指标名称 |
| `CheckFunc(fn)` | 设置检查函数 |
| `Timeout(d)` | 设置超时时间（默认 5s） |
| `Detail(key, value)` | 添加详细信息 |
| `Build()` | 构建健康指示器 |

## 创建与配置

```go
actuator := actuator.New(ctx)

// 可选：自定义健康聚合器
actuator.SetHealthAggregator(customAggregator)

// 可选：自定义指标注册表
actuator.SetMetricsRegistry(customRegistry)

// 获取指标注册表（直接使用）
registry := actuator.MetricsRegistry()
```

## 注册 HTTP 路由

```go
config := actuator.DefaultRouteConfig()
config.ExposeDebug = true  // 可选：暴露 pprof 调试端点

registrar := &actuator.StdRouteRegistrar{Mux: mux}
actuator.RegisterRoutes(registrar, config)
// 注册端点：
// - /actuator/health
// - /actuator/metrics
// - /actuator/env
// - /actuator/beans
// - /actuator/info
// - /metrics (Prometheus 格式)
// - /debug/pprof/* (当 ExposeDebug=true 时)
```

`RegisterRoutes` 接受任何实现了 `RouteRegistrar` 接口的路由器，兼容标准 `http.ServeMux`：

```go
mux := http.NewServeMux()
config := actuator.DefaultRouteConfig()
actuator.RegisterRoutes(&actuator.StdRouteRegistrar{Mux: mux}, config)
http.ListenAndServe(":8080", mux)
```

## 内置健康指标

### FuncHealthIndicator

基于函数的通用健康指标实现：

```go
indicator := actuator.NewFuncHealthIndicator(
    "my-service",
    func(ctx context.Context) error {
        resp, err := http.Get("http://localhost:8080/health")
        if err != nil {
            return err
        }
        resp.Body.Close()
        return nil
    },
)
```

### DatabaseHealthIndicator

数据库健康指标，名称为 `"database"`：

```go
indicator := actuator.NewDatabaseHealthIndicator(
    func(ctx context.Context) error {
        return db.PingContext(ctx)
    },
)
```

### RedisHealthIndicator

Redis 健康指标，名称为 `"redis"`：

```go
indicator := actuator.NewRedisHealthIndicator(
    func(ctx context.Context) error {
        return redisClient.Ping(ctx).Err()
    },
)
```

### checkHealth 检查逻辑

这三种指标统一使用 `checkHealth` 函数：

1. `checkFn` 为 `nil` → 返回 `StatusUnknown`
2. `checkFn` 返回错误 → 返回 `StatusDown`，详情中附带错误信息
3. 检查通过 → 返回 `StatusUp`

## 自动配置

通过 `boot.RegisterAutoConfig` 注册，由 `boot.Application` 在启动时自动执行：

```go
func init() {
    boot.RegisterAutoConfig(
        &ActuatorAutoConfiguration{},
        condition.OnProperty("actuator.enabled", "true"),
    )
}
```

### 配置属性

| 属性 | 默认值 | 说明 |
|------|--------|------|
| `actuator.enabled` | — | 设为 `"true"` 启用 Actuator |

### 自动配置行为

1. 创建 `Actuator` 实例
2. 从 IoC 容器中查找所有已注册的 `health.Indicator` 实现
3. 若有 Indicator，自动创建 `Aggregator` 并聚合，设置为 Actuator 的健康聚合器
4. 将 Actuator 实例以 `"actuator"` ID 注册到 IoC 容器

## 完整使用示例

```go
package main

import (
    "net/http"
    "github.com/xudefa/go-boot/actuator"
    "github.com/xudefa/go-boot/boot"
    "github.com/xudefa/go-boot/health"
)

func main() {
    app := boot.NewApplication(
        boot.WithProperty("actuator.enabled", "true"),
    )
    // 启动后 Actuator 自动配置完成

    // 注册自定义健康指标
    app.Context().Container().Register("myIndicator",
        core.Bean(&customIndicator{}))

    // 注册路由
    mux := http.NewServeMux()
    act, _ := app.Context().Get("actuator")
    act.(*actuator.Actuator).RegisterRoutes(mux)

    http.ListenAndServe(":8080", mux)
}

type customIndicator struct{}

func (c *customIndicator) Name() string { return "custom" }

func (c *customIndicator) Health(ctx context.Context) health.Health {
    return health.Health{Status: health.StatusUp}
}
```

### REST API 示例

```bash
# 健康检查
curl http://localhost:8080/actuator/health
# {"status":0,"details":{"database":{"status":"UP","detail":{}}},"timestamp":"2025-01-01T00:00:00Z"}

# 指标收集
curl http://localhost:8080/actuator/metrics
# [{"name":"requests_total","value":42,"tags":{"service":"api"}}]

# 环境信息
curl http://localhost:8080/actuator/env
# [{"name":"application.properties","priority":0,"properties":[{"name":"server.port","value":"8080"}]}]

# Bean 列表
curl http://localhost:8080/actuator/beans
# {"beans":[{"name":"actuator","type":"*actuator.Actuator","singleton":true}]}
```