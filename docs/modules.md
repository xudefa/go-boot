# go-boot 核心模块文档

## 快速导航

- **架构概览**：参见 [architecture.md](architecture.md)
- **各模块详细文档**：参见下方模块列表

---

## 核心模块

go-boot 核心框架零外部依赖，仅使用 Go 标准库实现。

| 包 | 说明 | 接口定义 | 详细文档 |
|------|------|----------|----------|
| **core** | IoC 容器（依赖注入核心） | `core.Container` | [core.md](core.md) |
| **aop** | AOP 框架（5 种通知类型 + 多种切点匹配） | `aop.Advice`, `aop.PointCut`, `aop.Weaver` | [aop.md](aop.md) |
| **boot** | 应用启动器、自动配置注册、横幅、失败分析 | `boot.AutoConfiguration`, `boot.Starter` | [boot.md](boot.md) |
| **context** | 应用上下文（聚合容器、环境、生命周期、事件） | `context.ApplicationContext` | [context.md](context.md) |
| **environment** | 环境配置管理（分层 PropertySource + Profile） | `environment.Environment` | [environment.md](environment.md) |
| **condition** | 条件判断（OnProperty / OnBean / OnClass 等） | `condition.Condition` | [condition.md](condition.md) |
| **event** | 事件驱动支持（发布/订阅） | `event.ApplicationEvent`, `event.EventPublisher` | [event.md](event.md) |
| **life** | 生命周期阶段管理（7 个阶段） | `life.Lifecycle`, `life.Phase` | [life.md](life.md) |
| **data** | 数据访问抽象（Repository[T] / Transactor） | `data.Repository[T]`, `data.Transactor` | [data.md](data.md) |
| **cache** | 缓存抽象（Cache 接口） | `cache.Cache` | [cache.md](cache.md) |
| **config** | 配置管理（Config 接口 + Loader 链 + Validator） | `config.Config`, `config.Loader` | [config.md](config.md) |
| **log** | 日志抽象（Logger 接口 + slog 默认实现） | `log.Logger` | [log.md](log.md) |
| **net** | HTTP 服务器/客户端抽象接口 | `net.Server`, `net.HttpClient` | [net.md](net.md) |
| **health** | 健康指标（Indicator + Aggregator） | `health.Indicator`, `health.HealthAggregator` | [health.md](health.md) |
| **metrics** | 指标收集（Counter + Gauge + Registry） | `metrics.Counter`, `metrics.Gauge`, `metrics.MeterRegistry` | [metrics.md](metrics.md) |
| **actuator** | 运维端点（健康、指标、环境信息） | `actuator.Endpoint` | [actuator.md](actuator.md) |
| **schedule** | 定时任务调度（Cron 解析、最小堆调度器、@Scheduled 注解） | `schedule.Task`, `schedule.Scheduler` | [schedule.md](schedule.md) |
| **center** | 注册中心抽象（Registry 接口 + Selector 接口 + 内置选择器） | `center.Registry`, `center.Selector` | [center.md](center.md) |
| **circuit** | 熔断器实现（防止级联故障） | `circuit.Breaker` | [circuit.md](circuit.md) |
| **loadbalancer** | 负载均衡器（多种策略实现） | `loadbalancer.Balancer` | [loadbalancer.md](loadbalancer.md) |
| **validation** | 数据验证（HTTP 请求验证 + 结构体验证） | `validation.RequestValidator` | [validation.md](validation.md) |
| **security** | 安全框架（认证、授权、过滤器链） | `security.Filter`, `security.Authentication` | [secure.md](secure.md) |
| **exception** | 异常处理（错误码、处理器、中间件） | `exception.Handler`, `exception.Resolver` | — |
| **refresh** | 配置热刷新（Scope、代理、事件） | `refresh.Scope`, `refresh.Config` | — |
| **constants** | 常量定义（配置键、条件值） | — | — |

---

## 模块分类

### 核心框架

| 模块 | 说明 |
|------|------|
| core | IoC 容器，依赖注入核心 |
| aop | 面向切面编程框架 |
| boot | 应用启动器和自动配置 |
| context | 应用上下文聚合 |
| condition | 条件装配机制 |

### 配置与环境

| 模块 | 说明 |
|------|------|
| environment | 环境配置管理 |
| config | 配置加载与管理 |
| refresh | 配置热刷新支持 |

### 数据与缓存

| 模块 | 说明 |
|------|------|
| data | 数据访问抽象 |
| cache | 缓存抽象接口 |

### 网络与通信

| 模块 | 说明 |
|------|------|
| net | HTTP 服务器/客户端接口 |
| center | 服务注册与发现 |
| loadbalancer | 负载均衡策略 |

### 可观测性

| 模块 | 说明 |
|------|------|
| log | 日志抽象 |
| health | 健康检查 |
| metrics | 指标收集 |
| actuator | 运维端点 |

### 安全与验证

| 模块 | 说明 |
|------|------|
| security | 安全框架 |
| validation | 数据验证 |
| exception | 异常处理 |

### 可靠性

| 模块 | 说明 |
|------|------|
| circuit | 熔断器 |
| event | 事件驱动 |
| life | 生命周期管理 |
| schedule | 定时任务 |