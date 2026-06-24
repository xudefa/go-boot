# go-boot 架构设计文档

## 一、概述

go-boot 是一个参考 Spring Boot 设计哲学、保留 Go 语言特性的工程化应用框架。核心使用 Go 标准库实现（零外部依赖），提供依赖注入（IoC）、面向切面编程（AOP）、数据访问抽象、缓存抽象、日志抽象、配置管理、健康检查、指标收集、服务注册发现、熔断器、负载均衡、数据验证、定时任务调度等开箱即用的能力。

### 设计目标

- **零依赖核心**：核心框架仅使用 Go 标准库，无任何第三方依赖
- **接口抽象**：通过接口抽象实现组件的灵活替换
- **约定优于配置**：提供合理的默认配置，减少样板代码
- **生命周期管理**：统一的组件启动、运行、停止生命周期
- **可测试性**：核心层不依赖外部服务，易于单元测试

### 模块分布

| 分类 | 说明 |
|------|------|
| 核心框架 | 24 个子包（core, aop, boot, context, circuit, loadbalancer, validation 等），零外部依赖 |

---

## 二、核心架构

### 2.1 整体架构图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        ApplicationContext                                │
│  ┌──────────┐ ┌──────────────┐ ┌──────────────┐ ┌─────────────┐        │
│  │ Container │ │ Environment  │ │ Lifecycle    │ │ EventBus    │        │
│  │  (IoC)    │ │ 配置源管理     │ │ 阶段管理      │ │ 事件发布/订阅 │        │
│  └──────────┘ └──────────────┘ └──────────────┘ └─────────────┘        │
│                         ┌────────────────────┐                          │
│                         │ AutoConfigRegistry  │                          │
│                         │   自动配置注册表      │                          │
│                         └────────────────────┘                          │
└─────────────────────────────────────────────────────────────────────────┘
         │              │              │              │
         ▼              ▼              ▼              ▼
┌──────────────┐ ┌────────────┐ ┌───────────┐ ┌──────────────┐
│  Starters    │ │  Actuator  │ │  Failure  │ │    Banner    │
│  启动器注册表  │ │  健康/指标   │ │  Analyzer │ │    启动横幅    │
└──────────────┘ └────────────┘ └───────────┘ └──────────────┘
```

### 2.2 包依赖结构

```
go-boot/
├── core/           ← IoC 容器（依赖注入核心）
├── aop/            ← AOP 框架（引用 core 的类型信息）
├── data/           ← 数据访问抽象
├── cache/          ← 缓存抽象
├── config/         ← 配置管理接口
├── log/            ← 日志抽象接口
├── net/            ← HTTP 服务器/客户端抽象
├── circuit/        ← 熔断器（防止级联故障）
├── loadbalancer/   ← 负载均衡器（多种策略）
├── validation/     ← 数据验证（HTTP 请求验证）
├── security/       ← 安全框架（认证、授权）
├── actuator/       ← 运维端点（健康、指标、环境信息）
├── condition/      ← 条件判断（用于自动配置）
├── environment/    ← 环境配置管理（多层 PropertySource）
├── health/         ← 健康指标接口
├── metrics/        ← 指标收集接口
├── schedule/       ← 定时任务调度（Cron 解析、最小堆调度器、@Scheduled 注解）
├── center/         ← 注册中心抽象（服务注册发现）
├── event/          ← 事件驱动支持
├── exception/      ← 异常处理
├── refresh/        ← 配置热刷新
├── context/        ← 应用上下文（聚合容器、环境、生命周期、事件）
├── life/           ← 生命周期阶段管理
├── boot/           ← 启动器、自动配置注册、横幅、失败分析
├── constants/      ← 常量定义
└── root pkg        ← Starter 接口（简单启动器/上下文启动器）
```

---

## 三、核心模块详解

### 3.1 IoC 容器 (`core/`)

依赖注入容器是整个框架的基石，参考 Spring Framework 的 IoC 容器设计。

#### 核心接口：`Container`

```go
type Container interface {
    Register(beanID string, builder ...BuilderOption) error
    Inject(target any) error
    Get(beanID string, opts ...GetOption) (any, error)
    GetAll(beanType any) ([]any, error)
    Invoke(fn any, opts ...InvokeOption) ([]any, error)
    Has(beanID string) bool
    Remove(beanID string) error
    Close() error
}
```

#### 关键特性

- **Bean 定义** (`BeanDefinition`)：实例、工厂函数、具体类型、作用域、依赖、初始化、条件、后置处理器、RefreshScope、ConfigKeys
- **两种作用域**：Singleton（单例，默认）、Prototype（原型，每次获取新实例）
- **字段注入**：通过 `inject:"beanId"` 结构体标签自动注入，统一使用 `FieldInjector` 处理 tag 注入和注解注入
- **方法注入**：`Invoke(fn)` 自动解析函数参数类型并从容器获取对应的 Bean
- **类型查找**：`GetAll(interfaceType)` 获取所有实现指定接口的 Bean
- **条件注册**：`Condition(fn)` 动态决定是否注册
- **循环依赖检测**：通过 `BeanCreationGuard` 的 `creating` 映射检测循环引用，检测到循环依赖时立即返回错误
- **并发安全**：使用 `BeanCreationGuard` 管理 Bean 创建并发，细粒度锁 + 创建标记，避免竞态条件和循环依赖死锁
- **深拷贝优化**：原型模式使用 `core.DeepCopy` 递归深拷贝，支持结构体、slice、map 等类型，正确处理 sync 等不可复制类型
- **泛型 API**：`BeanOf[T]`、`FactoryOf[T]`、`TypeOf[T]` 提供编译期类型安全

#### 创建流程

```
Register() → BeanDefinition 存储 → Get() → createInstance() → injectFields() → Init() → PostProcessors → 缓存/返回
```

#### 容器链

支持通过 `parent` 字段形成容器链，子容器查找不到时自动向上查找父容器。

### 3.2 AOP 框架 (`aop/`)

面向切面编程框架，参考 Spring AOP 设计。

#### 核心概念

| 概念 | 说明 |
|------|------|
| Advice | 通知（增强逻辑），5 种类型 |
| PointCut | 切点（方法匹配规则） |
| Advisor | 顾问（PointCut + Advice 组合） |
| AspectMeta | 切面元数据（Instance + PointCut + Advice + Order） |
| Weaver | 织入器（将切面织入目标对象生成代理） |
| JoinPoint | 连接点（方法调用的上下文信息） |
| ProxyFactory | 代理工厂（创建 AOP 代理对象） |
| AopRegistry | AOP 注册表（管理切面和织入器） |
| InterfaceProxyWrapper | 接口代理包装器（通过反射转发接口方法调用，支持 AOP 切面织入） |
| ChainExecutor | 链执行器（执行通知链，按正确顺序执行各种类型的通知） |

#### 通知类型

- `AdviceBefore` — 前置通知，在目标方法执行前调用
- `AdviceAfter` — 后置通知，在目标方法执行后调用（无论是否异常）
- `AdviceAfterReturning` — 返回通知，目标方法正常返回后调用
- `AdviceAfterThrowing` — 异常通知，目标方法抛出异常后调用
- `AdviceAround` — 环绕通知，可完全控制方法执行

#### 切点匹配器

| 匹配器 | 说明 |
|--------|------|
| `MatchAll()` | 匹配所有方法 |
| `MatchByName(name)` | 按方法名精确匹配 |
| `MatchByNamePrefix(prefix)` | 按方法名前缀匹配 |
| `MatchByRegex(pattern)` | 正则表达式匹配 |
| `MatchByAnnotation(t)` | 按注解类型匹配 |
| `MatchClass(matcher)` | 自定义类匹配 |
| `MatchMethod(matcher)` | 自定义方法匹配 |
| `MatchInterface(iface)` | 匹配实现指定接口的类型 |

#### 通知执行顺序

```
Before 通知（按 Order 升序）
  → Around 通知链
    → 目标方法
  → After/AfterReturning 通知
AfterThrowing 通知（异常时）
```

### 3.3 应用启动与自动配置 (`boot/`)

#### 启动入口

```go
boot.NewApplication(
    boot.WithConfigLocation("application.yml"),
    boot.WithProfiles("dev"),
)
```

`NewApplication` 创建 包含 Container、Environment、ApplicationContext 的应用实例，通过 `boot.Start()` 显示横幅并启动。

#### 自动配置机制

参考 Spring Boot 的 `@Configuration + @Bean` 模式：

1. **AutoConfiguration 接口**：每个模块实现 `Configure(ctx ApplicationContext)` 方法
2. **AutoConfigRegistry**：全局注册表，所有自动配置通过 `init()` 函数注册
3. **条件评估**：支持 `condition.OnProperty()`、`condition.OnBean()` 等条件控制
4. **排序与依赖**：支持 `WithOrder()` 和 `WithDependsOn()` 控制执行顺序

#### Starter 机制

```go
type Starter interface {
    Name() string
    Dependencies() []string
    Configure(ctx ApplicationContext) error
    Start(ctx ApplicationContext) error
    Stop(ctx ApplicationContext) error
    GetCondition() condition.Condition
}
```

StarterRegistry 管理所有启动器，支持按依赖关系拓扑排序。

#### 失败分析器

参考 Spring Boot 的 FailureAnalyzer，在启动失败时提供友好的错误提示和可能的解决方案。

#### 结构化启动错误（BootError）

`BootError` 是框架提供的结构化启动错误类型，包含错误发生的阶段、原始错误、分析结果和修复建议：

```go
type BootError struct {
    Phase       string   // 错误发生的阶段（如 "configuring", "starting"）
    Original    error    // 原始错误
    Analyzed    string   // FailureAnalyzer 分析结果
    Suggestions []string // 修复建议
}
```

`Boot.reportError` 方法自动使用 `FailureAnalyzer` 分析错误，并将分析结果填充到 `BootError` 中。

### 3.4 应用上下文 (`context/`)

`ApplicationContext` 是使用 go-boot 框架的核心入口，统一封装了：

- **Container**：IoC 容器（Bean 管理）
- **Environment**：环境配置（PropertySource、Profile）
- **Lifecycle**：生命周期管理器（阶段流转）
- **EventBus**：事件总线（发布/订阅）
- **EventPublisher**：事件发布器接口（解耦事件发布逻辑，便于测试和替换实现）

#### ClassLoader 缓存优化

`buildInfoClassLoader` 使用 `runtime/debug.ReadBuildInfo()` 检查模块是否在编译依赖中，通过 `sync.Once` 延迟初始化并缓存依赖列表，避免每次调用都读取构建信息。

#### 生命周期阶段

```
PhaseInitializing → PhaseConfiguring → PhaseContextRefreshed
  → PhaseReady → PhaseRunning → PhaseStopping → PhaseStopped
```

### 3.5 环境配置 (`environment/`)

分层配置源管理，参考 Spring 的 Environment 抽象：

#### 配置源优先级（从高到低）

| 优先级 | 配置源 | 说明 |
|--------|--------|------|
| 最高 | ArgsPropertySource | 命令行参数 `--key=value` |
| 高 | EnvPropertySource | 环境变量（如 `GO_BOOT_SERVER_PORT`） |
| 中 | MapPropertySource | 动态添加的配置源 |
| 低 | 文件配置源 | 通过 viper 集成 |

#### Profile 机制

支持通过 `--profile=dev` 命令行参数或 `GO_BOOT_PROFILE` 环境变量激活 Profile，支持否定前缀 `!dev`。

#### TypeConverter（类型转换器）

统一的类型转换逻辑，支持多种类型间的转换，避免重复代码：
- 数值类型到 string 的转换使用格式化而非 ASCII 码
- 支持所有基本数值类型和 bool/string 之间的转换
- 转换失败返回明确的错误信息

#### 配置绑定

```go
type Config struct {
    Host string `env:"server.host"`
    Port int    `env:"server.port"`
}
var cfg Config
env.Bind(&cfg)        // 全量绑定
env.BindPrefix("server", &cfg) // 前缀绑定
```

### 3.6 事件系统 (`event/`)

参考 Spring 的 ApplicationEvent/ApplicationListener 模式。

#### 内置事件类型

| 事件 | 说明 |
|------|------|
| EventEnvironmentPrepared | 环境准备完成 |
| EventContextRefreshed | 上下文刷新完成 |
| EventApplicationStarted | 应用启动 |
| EventApplicationReady | 应用就绪 |
| EventApplicationStopped | 应用停止 |

### 3.7 条件判断 (`condition/`)

参考 Spring Boot 的 @Conditional 注解体系，用于自动配置的条件控制。

#### 内置条件

- `OnProperty(key, value)` — 属性条件
- `OnBean(beanID)` / `OnMissingBean(beanID)` — Bean 存在条件
- `OnClass(className)` — 类路径条件
- `OnProfile(profile)` — Profile 条件
- `All(conditions...)` / `Any(conditions...)` — 组合条件

### 3.8 接口抽象层

所有集成层通过根模块中的接口抽象定义，实现运行时互换。

#### HTTP 服务器接口 (`net.Server`)

```go
type Server interface {
    Starter() error
    Use(m any) Server
    UseGlobal(m any) Server
    Register(fn func(core.Container) error) Server
    Container() core.Container
    Stop(ctx context.Context) error
}
```

#### HTTP 客户端接口 (`net.HttpClient`)

提供 Get/Post/Put/Delete/Do 方法，支持请求选项（Header、Query、Timeout、AuthToken、BasicAuth）。

#### 数据访问接口 (`data.Repository[T]`)

泛型 CRUD 接口：Create、CreateBatch、Read（FindByID/FindOne/FindAll）、Update、Delete、Count、Raw。

#### 事务接口 (`data.Transactor` / `data.Transaction`)

Query、QueryRow、Exec、Begin、Commit、Rollback、Stats、Close。

#### 缓存接口 (`cache.Cache`)

Get、Set、Del、Exists、TTL、Close。

#### 日志接口 (`log.Logger`)

Debug、Info、Warn、Error、DPanic、Panic、Fatal + With/Sync 以及扩展接口（WithLevel、WithName、WithCaller、WithTimeout）。

#### 健康指标接口 (`health.Indicator`)

Name() + Health(ctx) Health，通过 Aggregator 聚合所有指标。

#### 配置接口 (`config.Config`)

Get、GetString、GetInt、GetBool、Unmarshal、Watch 等 20+ 方法。

#### 指标接口 (`metrics.MeterRegistry`)

Counter/Gauge 创建和 Collect 收集。

#### 注册中心接口 (`center.Registry` / `center.Selector`)

- `Registry` — 注册发现：Register、Deregister、Discover、Watch
- `Selector` — 客户端负载均衡：Select 从实例列表中选择目标
- 内置选择器：RandomSelect、RoundRobinSelect、NewWeightedRandomSelect

---

## 四、核心模块文档索引

### 核心子包文档

每个核心子包有独立的详细说明文档，涵盖接口定义、核心类型、使用示例：

| 包 | 文档 | 说明 |
|----|------|------|
| core/ | [docs/core.md](core.md) | IoC 容器（依赖注入核心） |
| aop/ | [docs/aop.md](aop.md) | AOP 框架（5 种通知类型 + 多种切点匹配） |
| boot/ | [docs/boot.md](boot.md) | 应用启动器、自动配置注册、横幅、失败分析 |
| context/ | [docs/context.md](context.md) | 应用上下文（聚合容器、环境、生命周期、事件） |
| life/ | [docs/life.md](life.md) | 7 个生命周期阶段管理 |
| event/ | [docs/event.md](event.md) | 事件驱动支持（发布/订阅） |
| environment/ | [docs/environment.md](environment.md) | 分层 PropertySource + Profile |
| condition/ | [docs/condition.md](condition.md) | 条件判断（OnProperty / OnBean / OnClass 等） |
| config/ | [docs/config.md](config.md) | 配置管理（Config 接口 + Loader 链 + Validator） |
| data/ | [docs/data.md](data.md) | 数据访问抽象（Repository[T] / Transactor） |
| cache/ | [docs/cache.md](cache.md) | 缓存抽象（Cache 接口 + MemoryCache） |
| log/ | [docs/log.md](log.md) | 日志抽象（Logger 接口 + slog 默认实现） |
| net/ | [docs/net.md](net.md) | HTTP 服务器/客户端抽象接口 |
| health/ | [docs/health.md](health.md) | 健康指标（Indicator + Aggregator） |
| metrics/ | [docs/metrics.md](metrics.md) | 指标收集（Counter + Gauge + Registry） |
| actuator/ | [docs/actuator.md](actuator.md) | 运维端点（健康、指标、环境信息） |
| schedule/ | [docs/schedule.md](schedule.md) | 定时任务调度（Cron 解析 + 最小堆调度器 + @Scheduled） |
| security/ | [docs/secure.md](secure.md) | 安全框架（认证、授权、上下文） |
| center/ | [docs/center.md](center.md) | 注册中心抽象（Registry + Selector + 内置选择器） |
| circuit/ | [docs/circuit.md](circuit.md) | 熔断器（防止级联故障） |
| loadbalancer/ | [docs/loadbalancer.md](loadbalancer.md) | 负载均衡器（多种策略实现） |
| validation/ | [docs/validation.md](validation.md) | 数据验证（HTTP 请求验证 + 结构体验证） |
| https/ | [docs/https.md](https.md) | HTTPS 客户端 + AES/RSA 加解密 + TLS 配置 |

---

## 五、编码规范

### 函数式选项模式

整个框架优先使用函数式选项模式：

```go
container.Register("service",
    core.Bean(&Service{}),
    core.Singleton(),
    core.DependsOn("db"),
    core.Init(func(s *Service) error { return s.Start() }),
    core.Condition(func(c core.Container) bool { return c.Has("db") }),
)
```

### 组件扫描标签

```go
// @Service("userService")
type UserService struct {
    DB *Database `inject:"database"`
}
// @Repository
// @Configuration
```

### 包命名

核心框架包名采用简洁语义化命名（如 `core`, `aop`, `boot`），小写字母。

---

## 六、测试策略

- **核心框架**：单元测试，不依赖外部服务，使用表驱动测试
- **并行测试**：使用 `t.Parallel()`
- **命名规范**：`TestFunctionName_Condition_ExpectedBehavior`

---

## 七、核心包引用关系

```
go-boot/ (零外部依赖)
  ├── core/           ← 基础类型，被所有包引用
  ├── aop/            → core (reflect.Type)
  ├── net/            → core (Container 接口)
  ├── condition/      ← 独立，仅引用标准库
  ├── event/          ← 独立
  ├── life/           ← 独立
  ├── health/         ← 独立
  ├── metrics/        ← 独立
  ├── circuit/        ← 独立，仅引用标准库
  ├── loadbalancer/   ← 独立，仅引用标准库
  ├── validation/     ← 独立，仅引用标准库
  ├── center/         ← 纯接口，无内部依赖
  ├── context/        → core, environment, event, life
  ├── boot/           → context, core, environment, condition
  ├── data/           ← 纯接口，无内部依赖
  ├── actuator/       → context, health, metrics, boot
  ├── schedule/       → core, boot, condition
  ├── security/       → core, boot, condition
  ├── exception/      → core, log
  └── refresh/        → core, context, event
```

---

## 八、关键设计决策

### 为什么核心零外部依赖？

确保框架核心的稳定性和可移植性，用户无需引入不必要的传递依赖。所有功能通过接口抽象实现，便于扩展和替换。

### 为什么优先函数式选项模式？

相比建造者模式，函数式选项更符合 Go 语言的函数式编程风格，语义清晰，易于扩展（新增选项只需添加一个函数，无需修改接口）。

### 为什么支持自动配置机制？

降低框架使用门槛，让开发者专注于业务逻辑而非组件组装。通过注册表 + 条件判断实现灵活控制。