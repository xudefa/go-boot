# go-boot 集成模块文档

## 快速导航

- **核心包文档**：参见 [docs/](/) 目录下各子包的详细说明
- **架构概览**：参见 [architecture.md](architecture.md)

---

## 一、核心模块

go-boot 核心框架零外部依赖，仅使用 Go 标准库。所有集成模块以独立 GitHub 仓库形式提供。

| 包 | 说明 | 接口定义 |
|------|------|----------|
| **core** | IoC 容器（依赖注入核心） | `core.Container` |
| **aop** | AOP 框架（5 种通知类型 + 多种切点匹配） | `aop.Advice`, `aop.PointCut`, `aop.Weaver` |
| **boot** | 应用启动器、自动配置注册、横幅、失败分析 | `boot.AutoConfiguration`, `boot.Starter` |
| **context** | 应用上下文（聚合容器、环境、生命周期、事件） | `context.ApplicationContext` |
| **environment** | 环境配置管理（分层 PropertySource + Profile） | `environment.Environment` |
| **condition** | 条件判断（OnProperty / OnBean / OnClass 等） | `condition.Condition` |
| **event** | 事件驱动支持（发布/订阅） | `event.ApplicationEvent`, `event.EventPublisher` |
| **life** | 生命周期阶段管理（7 个阶段） | `life.Lifecycle`, `life.Phase` |
| **data** | 数据访问抽象（Repository[T] / Transactor） | `data.Repository[T]`, `data.Transactor` |
| **cache** | 缓存抽象（Cache 接口） | `cache.Cache` |
| **config** | 配置管理（Config 接口 + Loader 链 + Validator） | `config.Config`, `config.Loader` |
| **log** | 日志抽象（Logger 接口 + slog 默认实现） | `log.Logger` |
| **net** | HTTP 服务器/客户端抽象接口 | `net.Server`, `net.HttpClient` |
| **health** | 健康指标（Indicator + Aggregator） | `health.Indicator`, `health.HealthAggregator` |
| **metrics** | 指标收集（Counter + Gauge + Registry） | `metrics.Counter`, `metrics.Gauge`, `metrics.MeterRegistry` |
| **tracing** | 分布式追踪抽象 + LocalTracer 实现 | `tracing.Tracer`, `tracing.Span` |
| **actuator** | 运维端点（健康、指标、环境信息） | `actuator.Endpoint` |
| **schedule** | 定时任务调度（Cron 解析、最小堆调度器、@Scheduled 注解） | `schedule.Task`, `schedule.Scheduler` |
| **center** | 注册中心抽象（Registry 接口 + Selector 接口 + 内置选择器） | `center.Registry`, `center.Selector` |

---

## 二、集成模块（独立仓库）

以下集成模块已从核心框架中拆分，以独立 GitHub 仓库形式提供：

| 模块 | 仓库 | 实现接口 | 说明 |
|------|------|----------|------|
| **gin** | `github.com/xudefa/go-boot-gin` | net.Server | Gin HTTP 服务器 |
| **gorm** | `github.com/xudefa/go-boot-gorm` | data.Transactor, data.Repository[T] | GORM ORM |
| **redis** | `github.com/xudefa/go-boot-redis` | cache.Cache | Redis 缓存 |
| **viper** | `github.com/xudefa/go-boot-viper` | config.Config | 配置管理 |
| **zap** | `github.com/xudefa/go-boot-zap` | log.Logger | Zap 日志 |
| **zerolog** | `github.com/xudefa/go-boot-zerolog` | log.Logger | Zerolog 日志 |
| **etcd** | `github.com/xudefa/go-boot-etcd` | center.Registry | Etcd 注册中心 |
| **nacos** | `github.com/xudefa/go-boot-nacos` | center.Registry | Nacos 注册中心 |
| **consul** | `github.com/xudefa/go-boot-consul` | center.Registry | Consul 注册中心 |
| **grpc** | `github.com/xudefa/go-boot-grpc` | — | gRPC 服务端/客户端 |
| **hertz** | `github.com/xudefa/go-boot-hertz` | net.Server, net.HttpClient | Hertz HTTP 服务/客户端 |
| **fasthttp** | `github.com/xudefa/go-boot-fasthttp` | net.HttpClient | Fasthttp HTTP 客户端 |
| **websocket** | `github.com/xudefa/go-boot-websocket` | net.WebSocketServer | WebSocket 服务器 |
| **jwt** | `github.com/xudefa/go-boot-jwt` | — | JWT 认证 |
| **casbin** | `github.com/xudefa/go-boot-casbin` | — | 访问控制 |
| **email** | `github.com/xudefa/go-boot-email` | — | 邮件发送 |
| **swagger** | `github.com/xudefa/go-boot-swagger` | — | Swagger 文档 |
| **xorm** | `github.com/xudefa/go-boot-xorm` | data.Transactor | XORM ORM |
| **kitex** | `github.com/xudefa/go-boot-kitex` | — | Kitex RPC |
| **opentelemetry** | `github.com/xudefa/go-boot-opentelemetry` | tracing.Tracer | OpenTelemetry 追踪 |

### 使用方式

集成模块通过 `go get` 安装，例如：

```bash
go get github.com/xudefa/go-boot-gin@latest
go get github.com/xudefa/go-boot-gorm@latest
go get github.com/xudefa/go-boot-redis@latest
```

每个集成模块都实现了核心框架中定义的接口，可以无缝替换：

```go
// 使用 Gin 实现 net.Server 接口
import "github.com/xudefa/go-boot-gin"

server := gin.New(
    gin.WithContainer(app.Container()),
    gin.WithHost(":8080"),
)

// 使用 GORM 实现 data.Repository[T] 接口
import "github.com/xudefa/go-boot-gorm"

db, _ := gorm.OpenMySQL(
    gorm.WithDSN("user:pass@tcp(localhost:3306)/db"),
)
repo := gorm.NewRepository[User](db.DB)
```

---

## 三、示例模块

**路径**: `examples/`

| 示例 | 说明 |
|------|------|
| `core-hello/` | 核心框架基础使用（IoC 容器 + 依赖注入） |

更多示例请参考各集成模块的独立仓库文档。