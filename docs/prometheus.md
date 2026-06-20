# prometheus 包 — Prometheus 监控集成

## 概述

`prometheus` 包提供与 Prometheus 监控系统的深度集成，允许 go-boot 应用将内部指标数据转换为 Prometheus 格式并通过标准端点暴露。该模块支持 Counter、Gauge 等常见指标类型，并提供了自动配置功能。

## 快速开始

### 1. 启用 Prometheus 集成

通过配置文件启用 Prometheus 集成：

```yaml
prometheus:
  enabled: true
  endpoint: /metrics
  port: 9090
```

或者通过环境变量：

```bash
export PROMETHEUS_ENABLED=true
export PROMETHEUS_ENDPOINT=/metrics
export PROMETHEUS_PORT=9090
```

### 2. 在代码中使用

```go
package main

import (
    "github.com/xudefa/go-boot/boot"
    "github.com/xudefa/go-boot/prometheus"
)

func main() {
    app := boot.NewApplication(
        boot.WithProperty("prometheus.enabled", "true"),
        boot.WithProperty("prometheus.endpoint", "/metrics"),
    )
    
    // 启动应用
    app.Run()
}
```

### 3. 访问指标端点

启动应用后，可以通过以下 URL 访问 Prometheus 格式的指标：

```
http://localhost:9090/metrics
```

## 核心组件

### Exporter

Prometheus 导出器负责将 go-boot 指标转换为 Prometheus 格式：

```go
exporter := prometheus.NewExporter(bootMetrics)
err := exporter.Start()
if err != nil {
    // 处理错误
}
defer exporter.Stop()

// 获取 HTTP 处理器
handler := exporter.Handler()
```

### BootMetricsCollector

实现 Prometheus Collector 接口，将 go-boot 指标数据转换为 Prometheus 格式：

```go
collector := &prometheus.BootMetricsCollector{
    registry: bootMetricsRegistry,
}
```

## 配置选项

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `prometheus.enabled` | boolean | false | 是否启用 Prometheus 集成 |
| `prometheus.endpoint` | string | /metrics | 指标暴露端点 |
| `prometheus.port` | int | 9090 | HTTP 服务器端口 |

## 使用示例

### 完整的监控应用示例

```go
package main

import (
    "net/http"
    "time"

    "github.com/xudefa/go-boot/boot"
    "github.com/xudefa/go-boot/prometheus"
    "github.com/xudefa/go-boot/metrics"
)

func main() {
    app := boot.NewApplication(
        boot.WithProperty("prometheus.enabled", "true"),
        boot.WithProperty("prometheus.endpoint", "/metrics"),
    )

    // 启动应用
    app.Start()

    // 获取 Prometheus 导出器
    var exporter *prometheus.Exporter
    if err := app.Context().Get("prometheus.exporter", &exporter); err != nil {
        panic(err)
    }

    // 启动 HTTP 服务器
    http.Handle("/metrics", exporter.Handler())
    http.ListenAndServe(":9090", nil)
}
```

### 指标收集示例

```go
// 获取指标注册表
var registry metrics.MeterRegistry
app.Context().Get("meterRegistry", &registry)

// 记录请求计数
counter := registry.Counter("http_requests_total", "method", "GET", "handler", "api")
counter.Inc()

// 记录内存使用情况
gauge := registry.Gauge("memory_usage_bytes", "unit", "MB")
gauge.Set(1024.5)
```