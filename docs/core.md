# core 包 — IoC 依赖注入容器

## 概述

`core` 包提供了一个轻量级的依赖注入（Dependency Injection，DI）容器实现，设计灵感来源于 Spring Framework 的 IoC 容器。该包是 go-boot 框架的核心基石，零外部依赖，仅使用 Go 标准库。

### 核心功能

| 功能 | 说明 |
|------|------|
| **Bean 注册** | 支持通过实例、工厂函数或类型注册 Bean |
| **依赖注入** | 支持字段注入（`inject` 标签）和方法注入（`Invoke`） |
| **作用域管理** | 支持单例（Singleton）和原型（Prototype）作用域 |
| **组件扫描** | 自动扫描并注册带有 `@Service`、`@Repository`、`@Component` 注解的结构体 |
| **生命周期管理** | 支持初始化函数（`Init`）和 Bean 后置处理器（`BeanPostProcessor`） |
| **条件注册** | 支持根据条件动态决定是否注册 Bean |
| **泛型支持** | 提供类型安全的泛型 API（`BeanOf`、`FactoryOf`、`TypeOf`） |
| **循环依赖检测** | 自动检测并报告循环依赖 |

---

## 核心接口与类型

### Container（容器接口）

`Container` 是整个 IoC 容器的核心接口，管理 Bean 的注册、解析和生命周期：

```go
type Container interface {
    Register(beanID string, builder ...BuilderOption) error
    Inject(target any) error
    Get(beanID string, opts ...GetOption) (any, error)
    GetAll(beanType any) ([]any, error)
    Invoke(fn any, opts ...InvokeOption) ([]any, error)
    ListBeans() []BeanInfo
    Has(beanID string) bool
    Remove(beanID string) error
    Close() error
}
```

### BeanDefinition（Bean 定义）

描述一个 Bean 的完整元数据：

```go
type BeanDefinition struct {
    Instance         any                       // Bean 实例
    OriginalInstance any                       // 原始实例（用于复制）
    ConcreteType     reflect.Type              // Bean 的具体类型
    Scope            BeanScope                 // 作用域（Singleton/Prototype）
    Factory          func(Container) (any, error) // 工厂函数
    Fields           []FieldInjection          // 字段注入配置
    Init             func(any) error           // 初始化函数
    DependsOn        []string                  // 依赖的 Bean ID 列表
    Condition        func(Container) bool      // 条件注册函数
    PostProcessors   []BeanPostProcessor       // 后置处理器
    RefreshScope     bool                      // 是否支持刷新
    ConfigKeys       []string                  // Bean 依赖的配置键
}
```

### BeanScope（作用域）

```go
type BeanScope string

const (
    SingletonScope BeanScope = "singleton" // 单例，容器缓存同一实例
    PrototypeScope BeanScope = "prototype" // 原型，每次获取创建新实例
)
```

### Sentinel 错误

```go
var (
    ErrDuplicateBean = errors.New("duplicate bean registration") // Bean 重复注册
    ErrBeanNotFound  = errors.New("bean not found")              // Bean 未找到
    ErrCannotInject  = errors.New("cannot inject to non-pointer field")
    ErrInvalidScope  = errors.New("invalid scope")
    ErrCircularDep   = errors.New("circular dependency detected") // 循环依赖
)
```

### BeanInfo

Bean 摘要信息，用于 `ListBeans` 返回：

```go
type BeanInfo struct {
    ID        string   `json:"id"`
    Type      string   `json:"type"`
    Scope     string   `json:"scope"`
    DependsOn []string `json:"dependsOn,omitempty"`
}
```

---

## 快速开始

### 创建容器

```go
// 默认配置创建容器
container := core.New()

// 自定义配置创建容器（禁用字段标签注入）
container := core.New(core.EnableFieldTag(false))
```

### 注册 Bean

## Container.Register

注册一个 Bean 到容器中。

```go
container := core.New()
container.Register("userService", core.Bean(&UserService{}))
```

### 使用场景
- 应用启动时注册核心服务
- 动态注册插件模块

```go
// 方式 1：直接注册实例
container.Register("userService", core.Bean(&UserService{}))

// 方式 2：通过工厂函数注册（延迟创建）
container.Register("config",
    core.Factory(func(c core.Container) (any, error) {
        return &Config{Path: "/etc/app"}, nil
    }, reflect.TypeOf((*Config)(nil)).Elem()),
)

// 方式 3：泛型工厂函数（类型安全）
container.Register("config",
    core.FactoryOf(func(c core.Container) (*Config, error) {
        return &Config{Path: "/etc/app"}, nil
    }),
)
```

### 配置 Bean 属性

```go
container.Register("userService",
    core.Bean(&UserService{}),
    core.Singleton(),                         // 设置为单例（默认）
    core.DependsOn("db", "logger"),           // 声明依赖
    core.Init(func(s any) error {             // 初始化函数
        return s.(*UserService).Init()
    }),
    core.Condition(func(c core.Container) bool { // 条件注册
        return c.Has("db")
    }),
)
```

### 获取 Bean

## Container.Get

从容器中获取 Bean 实例。

```go
svc, err := container.Get("userService")
if err != nil {
    log.Fatal(err)
}
svc.(*UserService).DoSomething()
```

### 使用场景
- 获取已注册的服务实例
- 在应用运行时动态获取依赖

```go
// 按 ID 获取
svc, err := container.Get("userService")
if err != nil {
    log.Fatal(err)
}
svc.(*UserService).DoSomething()

// 获取所有实现某接口的 Bean
plugins, _ := container.GetAll((*Plugin)(nil))
for _, p := range plugins {
    p.(Plugin).Init()
}
```

### 字段注入（inject 标签）

## Container.Inject

自动注入带有 inject 标签的字段。

```go
type UserHandler struct {
    UserService *UserService `inject:"userService"`
    Logger      Logger       `inject:"logger"`
}

var handler UserHandler
container.Inject(&handler)
```

### 使用场景
- 自动装配结构体依赖
- 减少手动依赖注入代码

```go
type UserHandler struct {
    UserService *UserService `inject:"userService"`
    Logger      Logger       `inject:"logger"`
}

var handler UserHandler
container.Inject(&handler)
// handler.UserService 和 handler.Logger 已自动注入
```

### 方法注入（Invoke）

## Container.Invoke

调用函数并自动注入参数。

```go
result, err := container.Invoke(func(s *UserService, logger Logger) error {
    return s.DoSomething(logger)
})
```

### 使用场景
- 临时函数调用
- 集成测试中的依赖注入

```go
// 函数参数会自动从容器中获取对应类型的 Bean
result, err := container.Invoke(func(s *UserService, logger Logger) error {
    return s.DoSomething(logger)
})
```

### 列出与检查

```go
// 列出所有 Bean
for _, info := range container.ListBeans() {
    fmt.Printf("ID: %s, Type: %s, Scope: %s\n", info.ID, info.Type, info.Scope)
}

// 检查 Bean 是否存在
if container.Has("userService") {
    // ...
}
```

---

## BuilderOption 详解

`BuilderOption` 是配置 Bean 注册的函数式选项，可在 `Register` 时链式传入：

```go
type BuilderOption func(*BeanDefinition) error
```

### 完整选项列表

| 函数 | 类型参数 | 说明 |
|------|----------|------|
| `Bean(bean)` | 实例 | 使用已有实例注册 |
| `BeanOf[T](bean)` | 泛型实例 | 类型安全的实例注册 |
| `Factory(fn, type)` | 工厂 + 类型 | 工厂函数注册 |
| `FactoryOf[T](fn)` | 泛型工厂 | 类型安全的工厂注册 |
| `Type(t)` | 反射类型 | 仅指定类型 |
| `TypeOf[T]()` | 泛型类型 | 类型安全的类型指定 |
| `Singleton()` | — | 设置为单例作用域（默认） |
| `Prototype()` | — | 设置为原型作用域 |
| `Fields(fields...)` | 字段注入列表 | 字段注入 |
| `Field(name, value)` | 字段值注入 | 注入固定值 |
| `Ref(beanID, fieldName...)` | 字段引用注入 | 注入 Bean 引用 |
| `DependsOn(ids...)` | 依赖列表 | 声明依赖关系 |
| `Init(fn)` | 初始化函数 | 设置初始化回调 |
| `Condition(fn)` | 条件函数 | 设置条件注册 |
| `PostProcessor(pp...)` | 后置处理器 | 添加后置处理器 |

### BuilderOption API 示例

## core.Bean

使用已有实例注册 Bean。

```go
container.Register("userService", core.Bean(&UserService{}))
```

### 使用场景
- 注册已创建的实例
- 简单的服务注册

## core.Factory

通过工厂函数注册 Bean（延迟创建）。

```go
container.Register("config",
    core.Factory(func(c core.Container) (any, error) {
        return &Config{Path: "/etc/app"}, nil
    }, reflect.TypeOf((*Config)(nil)).Elem()),
)
```

### 使用场景
- 延迟初始化资源
- 动态创建复杂对象

## core.Singleton

设置 Bean 为单例作用域（默认）。

```go
container.Register("dbPool", core.Bean(&DBPool{}), core.Singleton())
```

### 使用场景
- 数据库连接池
- 配置中心客户端
- 线程安全的服务

## core.Prototype

设置 Bean 为原型作用域（每次获取创建新实例）。

```go
container.Register("request", core.Bean(&Request{}), core.Prototype())
```

### 使用场景
- HTTP 请求处理器
- 任务执行器
- 有状态的对象

## core.DependsOn

声明 Bean 的依赖关系。

```go
container.Register("userService",
    core.Bean(&UserService{}),
    core.DependsOn("db", "logger"),
)
```

### 使用场景
- 确保依赖先初始化
- 明确声明依赖关系

## core.Condition

设置条件注册函数。

```go
container.Register("productionService",
    core.Bean(&ProductionService{}),
    core.Condition(func(c core.Container) bool {
        return os.Getenv("ENV") == "production"
    }),
)
```

### 使用场景
- 根据环境变量注册不同实现
- 条件性启用功能

### 字段注入方式对比

```go
// 方式 1：inject 标签自动注入（容器级别的 Inject 调用）
type ServiceA struct {
    ServiceB *ServiceB `inject:"serviceB"`
}

// 方式 2：Field 注入固定值
container.Register("config",
    core.Bean(&Config{}),
    core.Field("Port", 8080),
    core.Field("Debug", true),
)

// 方式 3：Ref 注入 Bean 引用
container.Register("handler",
    core.Bean(&Handler{}),
    core.Ref("userService"),         // 字段名 = beanID
    core.Ref("logger", "Log"),       // 字段名 = "Log"
)
```

---

## 泛型工具函数

```go
// ZeroOf[T] — 获取类型的零值
var s string = core.ZeroOf[string]() // ""
var i int = core.ZeroOf[int]()       // 0
var p *T = core.ZeroOf[*T]()         // nil

// TypeOfGeneric[T] — 获取泛型参数的反射类型
t := core.TypeOfGeneric[UserService]()
fmt.Println(t.Name()) // "UserService"

// ValueOfGeneric[T] — 获取泛型值的反射值
v := core.ValueOfGeneric(myService)

// Clone[T] — 深克隆对象（使用递归反射深拷贝）
original := &User{Name: "Alice", Age: 30}
cloned, err := core.DeepCopy(original)
```

---

## 组件扫描（Component Scanning）

### 注解标记

通过在结构体的 Go 注释中使用 `@Service`、`@Repository`、`@Component` 或 `@Configuration` 标签，配合 `ComponentScanner` 实现自动扫描注册：

```go
// @Service("userService")
type UserService struct {
    DB *sql.DB `inject:"db"`
}

// @Repository
type UserRepository struct {
    // ...
}

// @Component("myBean")
type MyComponent struct {
    // ...
}
```

### 使用扫描器

```go
// 创建扫描器，指定要扫描的目录
scanner := core.NewComponentScanner("./internal/service")
if err := scanner.Scan(container); err != nil {
    log.Fatal(err)
}

// 扫描后，Bean 已自动注册到容器
svc, _ := container.Get("userService")
```

### 组件命名规则

| 注解形式 | Bean ID |
|----------|---------|
| `@Service("userService")` | `userService`（指定名称） |
| `@Service` | `userService`（结构体名首字母小写） |
| `@Repository` | `userRepository` |
| `@Component` | `myComponent` |

### 支持的注解类型

| 注解 | 组件类型 | 常量 |
|------|----------|------|
| `@Component` | 通用组件 | `ComponentTypeComponent` |
| `@Configuration` | 配置组件 | `ComponentTypeConfiguration` |
| `@Service` | 服务组件 | `ComponentTypeService` |
| `@Repository` | 仓储组件 | `ComponentTypeRepository` |

---

## 字段注入原理

容器的字段注入通过 `inject` 标签实现，流程如下：

1. `Inject(target)` 接收结构体指针
2. 遍历结构体字段，查找带有 `inject:"beanID"` 标签的字段
3. 对每个标记字段，调用 `container.Get(beanID)` 获取依赖
4. 检查类型兼容性后设置字段值

```go
type UserController struct {
    UserService *UserService   `inject:"userService"`
    Logger      Logger         `inject:"logger"`
    Config      *Config        `inject:"config"`
}

func main() {
    c := core.New()
    c.Register("userService", core.Bean(&UserService{}))
    c.Register("logger", core.Bean(&MyLogger{}))
    c.Register("config", core.Bean(&Config{Port: 8080}))

    ctrl := &UserController{}
    c.Inject(ctrl) // 自动注入所有 inject 标签字段
}
```

---

## 完整示例

```go
package main

import (
    "fmt"
    "log"
    "github.com/xudefa/go-boot/core"
)

// @Service("greeter")
type GreeterService struct {
    Prefix string
}

func (g *GreeterService) Greet(name string) string {
    return g.Prefix + ", " + name
}

// Handler 使用 inject 标签声明依赖
type Handler struct {
    Greeter *GreeterService `inject:"greeter"`
}

func main() {
    c := core.New()

    // 注册服务
    c.Register("greeter",
        core.Bean(&GreeterService{Prefix: "Hello"}),
        core.Singleton(),
    )

    // 手动注入
    h := &Handler{}
    if err := c.Inject(h); err != nil {
        log.Fatal(err)
    }
    fmt.Println(h.Greeter.Greet("World")) // 输出: Hello, World

    // 方法注入
    result, _ := c.Invoke(func(g *GreeterService) string {
        return g.Greet("Invoke")
    })
    fmt.Println(result[0]) // 输出: Hello, Invoke

    // 列出所有 Bean
    for _, info := range c.ListBeans() {
        fmt.Printf("Bean: %s (%s)\n", info.ID, info.Type)
    }
}
```

---

## 原型作用域示例

```go
// 注册原型 Bean（每次获取都创建新实例）
container.Register("counter",
    core.Bean(&Counter{Count: 0}),
    core.Prototype(),
)

c1, _ := container.Get("counter") // 新实例 Count=0
c2, _ := container.Get("counter") // 另一个新实例 Count=0
```

## 条件注册示例

```go
container.Register("productionService",
    core.Bean(&ProductionService{}),
    core.Condition(func(c core.Container) bool {
        return os.Getenv("ENV") == "production"
    }),
)
```

## 性能特性

- **并发安全**：使用 `BeanCreationGuard` 管理 Bean 创建并发，细粒度锁 + 创建标记，避免竞态条件和循环依赖死锁
- **缓存机制**：单例 Bean 创建后缓存，避免重复反射
- **循环依赖检测**：通过 `BeanCreationGuard` 的 `creating` 映射检测循环引用，检测到循环依赖时立即返回错误
- **深拷贝优化**：原型模式使用反射深拷贝（`core.DeepCopy`），支持结构体、slice、map 等类型，正确处理 sync 等不可复制类型

---

## 使用场景

### 场景 1：单例服务管理

**描述**：管理数据库连接池、配置中心客户端等单例服务。

```go
container.Register("dbPool", core.Bean(&DBPool{}), core.Singleton())
pool, _ := container.Get("dbPool")
```

**最佳实践**：
- 单例服务应保证线程安全
- 在应用启动时注册，避免运行时动态创建
- 使用 Init 函数进行初始化

### 场景 2：原型对象创建

**描述**：每次获取都创建新实例，适用于 HTTP 请求处理器、任务执行器等。

```go
container.Register("request", core.Bean(&Request{}), core.Prototype())
r1, _ := container.Get("request")
r2, _ := container.Get("request")
// r1 和 r2 是不同的实例
```

**最佳实践**：
- 原型对象应轻量级，避免重复创建昂贵资源
- 适用于有状态的对象
- 注意并发安全性

### 场景 3：条件 Bean 注册

**描述**：根据环境变量或配置动态决定是否注册 Bean。

```go
container.Register("prodService",
    core.Bean(&ProductionService{}),
    core.Condition(func(c core.Container) bool {
        return os.Getenv("ENV") == "production"
    }),
)

container.Register("devService",
    core.Bean(&DevService{}),
    core.Condition(func(c core.Container) bool {
        return os.Getenv("ENV") == "development"
    }),
)
```

**最佳实践**：
- 使用环境变量控制不同环境的实现
- 条件函数应简单高效
- 避免在条件函数中进行复杂逻辑

### 场景 4：容器链

**描述**：使用多个容器实现多租户隔离、模块化应用。

```go
// 创建父容器（共享服务）
parentContainer := core.New()
parentContainer.Register("sharedService", core.Bean(&SharedService{}))

// 创建子容器（租户特定服务）
tenantContainer := core.New()
tenantContainer.SetParent(parentContainer)
tenantContainer.Register("tenantService", core.Bean(&TenantService{}))

// 子容器可以访问父容器的 Bean
svc, _ := tenantContainer.Get("sharedService")
```

**最佳实践**：
- 父容器存放共享服务
- 子容器存放租户特定服务
- 合理设计容器层次结构