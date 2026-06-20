# go-boot

[![Go Version](https://img.shields.io/github/go-mod/go-version/xudefa/go-boot)](https://go.dev/) [![License](https://img.shields.io/github/license/xudefa/go-boot)](./LICENSE) [![Build Status](https://img.shields.io/github/actions/workflow/status/xudefa/go-boot/test.yml?branch=master)](https://github.com/xudefa/go-boot/actions) [![Go Reference](https://pkg.go.dev/badge/github.com/xudefa/go-boot.svg)](https://pkg.go.dev/github.com/xudefa/go-boot) [![Go Report Card](https://goreportcard.com/badge/github.com/xudefa/go-boot)](https://goreportcard.com/report/github.com/xudefa/go-boot)

Go 语言轻量级应用开发框架 — 提供依赖注入（IoC）、面向切面编程（AOP）、数据访问层抽象及常用中间件集成，帮助开发者快速构建可测试、松耦合的 Go 应用程序。

> 设计理念：零外部依赖的核心框架 + 可插拔的集成模块，借鉴 Spring Boot 的设计思想，为 Go 开发者提供熟悉的开发体验。

## 整体架构

```
┌───────────────────────────────────────────────────────────────────────┐
│                    boot.ApplicationContext                            │
│  ┌───────────┐ ┌──────────────┐ ┌───────────┐ ┌───────────┐           │
│  │ Container │ │  Environment │ │ Lifecycle │ │ EventBus  │           │
│  └───────────┘ └──────────────┘ └───────────┘ └───────────┘           │
│                       ┌─────────────────────┐                         │
│                       │ AutoConfig Registry │                         │
│                       └─────────────────────┘                         │
└───────────────────────────────────────────────────────────────────────┘
         │                        │                          │
         ▼                        ▼                          ▼
┌──────────────┐     ┌──────────────────┐      ┌──────────────────────────┐
│   Starters   │     │    Actuator      │      │   Failure Analyzers      │
└──────────────┘     └──────────────────┘      └──────────────────────────┘
```

## 目录

- [快速开始](#快速开始)
- [功能特性](#功能特性)
- [接口定义](#接口定义)
- [应用启动器](#应用启动器)
- [项目结构](#项目结构)
- [开发指南](#开发指南)
- [代码规范](#代码规范)
- [设计文档](#设计文档)
- [示例文档](#示例文档)

## 快速开始

### 安装

```bash
go get github.com/xudefa/go-boot
```

### IoC 容器

```go
package main

import (
    "fmt"
    "github.com/xudefa/go-boot/core"
)

type HelloService struct {
    Message string `inject:"message"`
}

func (s *HelloService) Say() string {
    return s.Message
}

func main() {
    c := core.New()
    c.Register("message", core.Bean("Hello World"))
    c.Register("hello", core.Bean(&HelloService{}))

    svc := c.Get("hello").(*HelloService)
    fmt.Println(svc.Say()) // Output: Hello World
}
```

### AOP 切面编程

```go
package main

import (
    "fmt"
    "github.com/xudefa/go-boot/aop"
    "github.com/xudefa/go-boot/core"
)

type LoggingAspect struct{}

func (l *LoggingAspect) Before(ctx aop.JoinPoint) {
    fmt.Printf("Before: %s\n", ctx.Method().Name)
}

func (l *LoggingAspect) After(ctx aop.JoinPoint, result any, err error) {
    fmt.Printf("After: %s\n", ctx.Method().Name)
}

type UserService struct{}

func (u *UserService) GetUser(id int) string {
    return fmt.Sprintf("User %d", id)
}

func main() {
    c := core.New()
    c.Register("userService", core.Bean(&UserService{}))

    aop.RegisterAspect(&LoggingAspect{}, aop.WithPointcut("*UserService.*"))

    svc := c.Get("userService").(*UserService)
    svc.GetUser(1)
}
```

### 应用启动器（推荐）

```go
package main

import (
    "github.com/xudefa/go-boot/boot"
    "github.com/xudefa/go-boot/core"
)

type MyService struct {
    Message string `inject:"message"`
}

func (m *MyService) Run() {
    println(m.Message)
}

func main() {
    app, err := boot.NewApplication(
        boot.WithAppName("my-app"),
        boot.WithVersion("1.0.0"),
        boot.WithProfiles("dev"),
    )
    if err != nil {
        panic(err)
    }
    defer app.Stop()

    // 注册 Bean
    app.Container().Register("message", core.Bean("Hello from go-boot!"))
    app.Container().Register("myService", core.Bean(&MyService{}))

    // 启动应用（显示横幅、发布事件）
    app.Start()

    // 获取服务并使用
    svc := app.Container().Get("myService").(*MyService)
    svc.Run()

    // 等待终止信号
    app.WaitForSignal()
}
```

### 定时任务

```go
package main

import (
    "fmt"
    "github.com/xudefa/go-boot/schedule"
)

func main() {
    scheduler := schedule.NewScheduler()

    // 每5秒执行一次
    scheduler.Schedule("*/5 * * * * ?", func() {
        fmt.Println("Task executed at:", schedule.Now())
    })

    // 每分钟执行一次
    scheduler.Schedule("0 * * * * ?", func() {
        fmt.Println("Minutely task executed")
    })

    scheduler.Start()
    defer scheduler.Stop()
}
```

### 事件驱动

```go
package main

import (
    "fmt"
    "github.com/xudefa/go-boot/event"
)

type UserCreatedEvent struct {
    UserID int
}

func (e *UserCreatedEvent) EventName() string {
    return "user.created"
}

func main() {
    bus := event.NewEventBus()

    // 订阅事件
    bus.Subscribe("user.created", func(e event.ApplicationEvent) {
        evt := e.(*UserCreatedEvent)
        fmt.Printf("User created: %d\n", evt.UserID)
    })

    // 发布事件
    bus.Publish(&UserCreatedEvent{UserID: 123})
}
```

## 功能特性

### 核心模块

| 模块 | 路径 | 说明 |
|------|------|------|
| core | [core/](core/) | IoC 容器：依赖注入、组件扫描、自动装配 |
| aop | [aop/](aop/) | 切面框架：切点匹配、通知类型、动态代理 |
| boot | [boot/](boot/) | 应用启动器、自动配置注册、横幅、失败分析 |
| context | [context/](context/) | 应用上下文：聚合容器、环境、生命周期、事件 |
| environment | [environment/](environment/) | 环境配置：分层 PropertySource、Profile |
| condition | [condition/](condition/) | 条件判断：OnProperty / OnBean / OnClass |
| event | [event/](event/) | 事件驱动：发布/订阅、内置生命周期事件 |
| life | [life/](life/) | 生命周期阶段管理（7 个阶段） |
| data | [data/](data/) | 数据访问：泛型 Repository[T]、Transactor 接口 |
| cache | [cache/](cache/) | 缓存抽象：统一缓存接口 |
| config | [config/](config/) | 配置管理：Config 接口、Loader 链、Validator |
| log | [log/](log/) | 日志接口：Logger 抽象、slog 默认实现 |
| net | [net/](net/) | HTTP 入口：服务器/客户端统一接口 |
| health | [health/](health/) | 健康指标：Indicator、Aggregator |
| metrics | [metrics/](metrics/) | 指标收集：Counter、Gauge、Registry |
| tracing | [tracing/](tracing/) | 分布式追踪：Tracer、Span 抽象 |
| actuator | [actuator/](actuator/) | 运维端点：健康检查、指标、环境信息 |
| schedule | [schedule/](schedule/) | 定时任务调度：Cron 解析、最小堆调度器 |
| center | [center/](center/) | 注册中心抽象：服务注册发现、客户端负载均衡 |

## 集成模块

go-boot 核心框架零外部依赖。以下集成模块提供对常用第三方库的支持，每个模块都是独立的 GitHub 仓库：

| 模块 | 仓库 | 说明 |
|------|------|------|
| Gin | [go-boot-gin](https://github.com/xudefa/go-boot-gin) | Gin HTTP 框架集成 |
| Hertz | [go-boot-hertz](https://github.com/xudefa/go-boot-hertz) | CloudWeGo Hertz 集成 |
| FastHTTP | [go-boot-fasthttp](https://github.com/xudefa/go-boot-fasthttp) | 高性能 HTTP 客户端 |
| gRPC | [go-boot-grpc](https://github.com/xudefa/go-boot-grpc) | gRPC 框架集成 |
| Kitex | [go-boot-kitex](https://github.com/xudefa/go-boot-kitex) | CloudWeGo Kitex RPC |
| GORM | [go-boot-gorm](https://github.com/xudefa/go-boot-gorm) | GORM ORM 集成 |
| XORM | [go-boot-xorm](https://github.com/xudefa/go-boot-xorm) | XORM ORM 集成 |
| Redis | [go-boot-redis](https://github.com/xudefa/go-boot-redis) | Redis 缓存集成 |
| Viper | [go-boot-viper](https://github.com/xudefa/go-boot-viper) | Viper 配置集成 |
| Zap | [go-boot-zap](https://github.com/xudefa/go-boot-zap) | Uber Zap 日志集成 |
| Zerolog | [go-boot-zerolog](https://github.com/xudefa/go-boot-zerolog) | RS Zerolog 日志集成 |
| Etcd | [go-boot-etcd](https://github.com/xudefa/go-boot-etcd) | Etcd 注册中心集成 |
| Nacos | [go-boot-nacos](https://github.com/xudefa/go-boot-nacos) | Nacos 注册中心集成 |
| Consul | [go-boot-consul](https://github.com/xudefa/go-boot-consul) | Consul 注册中心集成 |
| Casbin | [go-boot-casbin](https://github.com/xudefa/go-boot-casbin) | Casbin 权限控制 |
| JWT | [go-boot-jwt](https://github.com/xudefa/go-boot-jwt) | JWT 认证集成 |
| WebSocket | [go-boot-websocket](https://github.com/xudefa/go-boot-websocket) | WebSocket 支持 |
| Swagger | [go-boot-swagger](https://github.com/xudefa/go-boot-swagger) | API 文档生成 |
| Prometheus | [go-boot-prometheus](https://github.com/xudefa/go-boot-prometheus) | Prometheus 指标集成 |
| OpenTelemetry | [go-boot-opentelemetry](https://github.com/xudefa/go-boot-opentelemetry) | 分布式追踪集成 |
| Email | [go-boot-email](https://github.com/xudefa/go-boot-email) | 邮件发送集成 |
| Validation | [go-boot-validation](https://github.com/xudefa/go-boot-validation) | 参数验证集成 |

### 使用集成模块

```bash
# 安装核心框架
go get github.com/xudefa/go-boot@v1.0.0

# 按需安装集成模块
go get github.com/xudefa/go-boot-gin@v1.0.0
go get github.com/xudefa/go-boot-redis@v1.0.0
go get github.com/xudefa/go-boot-gorm@v1.0.0
```

## 接口定义

### `net.Server`

HTTP 服务器统一接口（由 go-boot-gin、go-boot-hertz 实现）：

- `Starter() error` — 启动服务器，支持优雅关闭
- `Use(m any) Server` — 添加路由级中间件
- `Stop(ctx) error` — 优雅停止
- `SetConfig(opts ...ServerOption)` - 设置服务器配置参数

### `net.HttpClient`

HTTP 客户端统一接口（由 go-boot-hertz、go-boot-fasthttp 实现）：

- `Get / Post / Put / Delete / Do` — RESTful 请求方法
- `Close() error` — 关闭客户端

支持请求选项：Header、Query、Timeout、AuthToken、BasicAuth。

### `data.Repository[T]`

泛型 CRUD 接口（由 go-boot-gorm 实现）：

| 方法 | 说明 |
|------|------|
| `Create(entity *T) error` | 创建记录 |
| `CreateBatch(entities []T) error` | 批量创建 |
| `Delete(id any) error` | 根据 ID 删除 |
| `DeleteByCondition(where, args...) error` | 条件删除 |
| `Update(entity *T) error` | 更新记录 |
| `UpdateByCondition(where, args...) (int64, error)` | 条件更新 |
| `FindByID(id any) (*T, error)` | 根据 ID 查询 |
| `FindOne(where, args...) (*T, error)` | 单条查询 |
| `FindAll(where, args...) ([]T, error)` | 多条查询 |
| `Count(where, args...) (int64, error)` | 计数查询 |
| `Raw(sql string, args...) ([]T, error)` | 原始 SQL 查询 |

### `data.Transactor` / `data.Transaction`

- `Query / QueryRow / Exec` — SQL 执行
- `Begin() (Transaction, error)` — 开始事务
- `Commit() / Rollback()` — 事务控制
- `Stats() DBStats` — 连接池统计
- `Close()` — 关闭连接

### `cache.Cache`

- `Get / Set / Del` — 基本缓存操作
- `Exists / TTL` — 缓存检查
- `Close()` — 关闭连接

### `log.Logger`

- `Debug / Info / Warn / Error / DPanic / Panic / Fatal` — 日志级别
- `With(ctx, keys...) Logger` — 带上下文的日志记录器
- `Sync()` — 同步缓冲区

### `health.Indicator`

```go
type Indicator interface {
    Name() string
    Health(ctx context.Context) Health
}
```

通过 Aggregator 聚合多个 Indicator，任一 DOWN 则整体 DOWN。

### `config.Config`

提供 Get、GetString、GetInt、GetBool、Unmarshal、Watch 等 20+ 配置访问方法。

## 应用启动器

推荐通过 `boot.NewApplication()` 创建应用，该入口替代了旧版入口：

```go
package main

import (
    "github.com/xudefa/go-boot/boot"
)

func main() {
    app, err := boot.NewApplication(
        boot.WithAppName("my-app"),
        boot.WithVersion("1.0.0"),
    )
    if err != nil {
        panic(err)
    }
    defer app.Stop()

    // 注册 Bean
    app.Container().Register("myService", core.Bean(&MyService{}))

    // 启动
    app.Start()
    // 等待信号
    app.WaitForSignal()
}
```

## 项目结构

```
go-boot/                    # 核心框架（零外部依赖）
├── core/                   # IoC 容器
├── aop/                    # 切面框架
├── boot/                   # 应用启动器
├── context/                # 应用上下文
├── environment/            # 环境配置
├── config/                 # 配置管理接口
├── condition/              # 条件判断
├── event/                  # 事件系统
├── life/                   # 生命周期
├── data/                   # 数据访问抽象
├── cache/                  # 缓存抽象
├── log/                    # 日志接口
├── net/                    # HTTP 入口接口
├── health/                 # 健康指标
├── metrics/                # 指标收集
├── tracing/                # 分布式追踪
├── actuator/               # 运维端点
├── schedule/               # 定时任务调度
├── center/                 # 注册中心抽象
├── constants/              # 常量定义
├── exception/              # 异常处理
├── refresh/                # 配置刷新
└── docs/                   # 文档
```

## 开发指南

### 构建

```bash
go build ./...
```

### 测试

```bash
go test ./...
go test -cover ./...       # 带覆盖率
go test -race ./...        # 数据竞争检测
```

### 代码规范

```bash
go fmt ./...
golangci-lint run
```

## 代码规范

### 1. 总体原则

- **清晰优于巧妙**：代码应该易于理解和维护
- **简单优于复杂**：优先选择简单直接的实现方式
- **可读性第一**：代码首先是给人阅读的，其次才是给机器执行的
- **零外部依赖**：核心框架不引入外部依赖，仅使用Go标准库

### 2. 命名规范

#### 2.1 包命名
- 包名全部小写
- 多个单词用连字符连接（如 `user-service`）
- 除 `main` 包外，其他包名应与最内层目录名保持一致

#### 2.2 标识符命名
- **导出标识符**：大写驼峰（`UserID`, `GetUser`）
- **非导出标识符**：小写驼峰（`userID`, `getUser`）
- **常量**：使用驼峰命名（`MaxConnections`, `DefaultTimeout`），避免使用全大写下划线
- **测试函数**：`TestFunctionName_Condition_ExpectedBehavior`
- **错误变量**：以 `Err` 前缀（`ErrNotFound`, `ErrInvalidInput`）

### 3. 代码结构

#### 3.1 包内文件组织
- `package` 声明后是包级别的文档注释
- 常量定义
- 类型定义（struct, interface, type alias）
- 变量声明
- 公共函数
- 私有函数
- 方法定义（按接收者分组）

#### 3.2 导入规范
```go
import (
    // 标准库
    "context"
    "fmt"
    "sync"

    // 项目内部包
    "github.com/xudefa/go-boot/core"
    "github.com/xudefa/go-boot/log"
)
```

### 4. 注释与文档

#### 4.1 注释语言
- 使用中文注释，保持国际化友好
- 重点注释应说明"为什么这样做"而不是"做了什么"

#### 4.2 类型和函数注释
```go
// CalculateDiscount 计算应用分级折扣后的最终价格。
// 折扣根据订单数量逐步应用：每个等级解锁额外的百分比减免。
// 如果数量无效或基础价格在应用折扣后会导致负值，则返回错误。
//
// 参数:
//   - basePrice: 任何折扣前的原始价格（必须为非负数）
//   - quantity: 订单的数量（必须为正数）
//   - tiers: 按最小数量阈值排序的折扣等级切片
//
// 返回最终折扣价格，四舍五入到小数点后两位。
// 如果 basePrice 为负数，返回 ErrInvalidPrice。
// 如果 quantity 为零或负数，返回 ErrInvalidQuantity。
func CalculateDiscount(basePrice float64, quantity int, tiers []DiscountTier) (float64, error) {
    // implementation
}
```

### 5. 错误处理

- 不忽略任何返回的错误
- 使用 `fmt.Errorf` 和 `%w` 包装错误
- 提供清晰的错误信息
- 使用哨兵错误（sentinel errors）表示框架错误

```go
var (
    ErrNotFound = errors.New("item not found")
    ErrInvalidInput = errors.New("invalid input provided")
)

// Good - 包装错误
if err := validate(input); err != nil {
    return fmt.Errorf("validation failed: %w", err)
}
```

### 6. 特定领域规范

#### 6.1 IoC 容器规范
- 使用 `core.New()` 创建容器
- 启用字段注入：`core.EnableFieldTag(true)`
- Bean 注册使用 `container.Register("id", core.Bean(value))`
- 字段注入使用 `inject:"beanId"` 结构体标签
- 工厂函数使用 `core.Factory(func(c core.Container) (any, error))`

#### 6.2 AOP 规范
- 通知类型：`aop.Before`, `aop.After`, `aop.Around`, `aop.AfterReturning`, `aop.AfterThrowing`
- 切点匹配器：`aop.MatchByName`, `aop.MatchByPrefix`, `aop.MatchByRegex` 等
- 通过 `aop.WithOrder(n)` 控制通知执行顺序
- Around 通知必须调用 `proceed` 使调用链继续

#### 6.3 函数式选项模式
```go
// Good - 函数式选项
container.Register("service",
    core.Bean(&Service{}),
    core.Singleton(),
    core.DependsOn("db"),
    core.Init(func(s *Service) error { return s.Start() }),
    core.Condition(func(c core.Container) bool { return c.Has("db") }),
)
```

## 贡献

欢迎提交 Issue 和 Pull Request！详细贡献指南请参阅 [CONTRIBUTING.md](./CONTRIBUTING.md)。

## 许可证

本项目采用 MIT 许可证 — 详情请参阅 [LICENSE](./LICENSE) 文件。