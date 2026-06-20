# aop-integration 包 — AOP 集成指南

## 概述

`aop-integration` 包提供了代码生成工具与现有 AOP 框架的深度集成，支持多种 AOP 模式，为开发者提供统一的 AOP 使用体验。

### Go 语言的 AOP 限制

由于 Go 语言的反射限制，运行时 AOP 有以下限制：

| 限制 | 说明 |
|------|------|
| **接口代理限制** | 无法在运行时创建实现接口的动态类型 |
| **结构体代理限制** | 无法在运行时替换结构体的方法实现 |
| **方法包装限制** | 即使有接口，也无法真正包装方法调用 |

### 解决方案

为了解决这些限制，框架提供了三种 AOP 模式：

| 模式 | 说明 | 优点 | 缺点 |
|------|------|------|------|
| **代码生成模式** | 使用 `goaop` 命令行工具生成静态代理代码 | 性能最优，无反射开销；支持所有方法类型；类型安全 | 需要额外的代码生成步骤；需要重新编译 |
| **运行时模式** | 使用反射和动态代理，运行时织入 AOP 逻辑 | 无需代码生成；灵活性高 | 性能较低；受 Go 语言限制严重；只能对接口类型进行有限代理 |
| **混合模式** | 优先使用代码生成的代理，回退到运行时代理 | 结合两者优点；自动选择最优方案 | 配置相对复杂 |

---

## 核心组件

### AopManager

统一管理代码生成和运行时 AOP。

| 方法 | 说明 |
|------|------|
| `NewAopManager(config *AopConfig)` | 创建 AOP 管理器 |
| `GetProxy(target any) any` | 获取代理对象 |
| `RegisterAspect(aspect *AspectMeta)` | 注册切面 |
| `GetAspects() []*AspectMeta` | 获取所有切面 |
| `ClearCache()` | 清理代理缓存 |

### AopIntegration

提供代码生成和运行时 AOP 的统一集成。

| 方法 | 说明 |
|------|------|
| `NewAopIntegration(config *AopConfig)` | 创建 AOP 集成器 |
| `CreateProxy(beanID string, target any) any` | 创建代理对象 |
| `RegisterAspect(aspect *AspectMeta)` | 注册切面 |
| `GetAspects() []*AspectMeta` | 获取所有切面 |

### AopContainer

集成 AOP 功能的 IoC 容器。

| 方法 | 说明 |
|------|------|
| `CreateAopContainer()` | 创建 AOP 容器 |
| `RegisterAopBean(beanDef *AopBeanDefinition, target any) error` | 注册 AOP Bean |
| `RegisterAopBeanWithID(beanID string, beanType reflect.Type, target any) error` | 注册带 ID 的 AOP Bean |
| `GetAopProxy(beanID string) (any, error)` | 获取 AOP 代理 |
| `RegisterAspect(aspect *AspectMeta)` | 注册切面 |

### GeneratedProxyRegistry

管理代码生成的代理对象。

| 方法 | 说明 |
|------|------|
| `NewGeneratedProxyRegistry()` | 创建代理注册表 |
| `Register(beanID string, proxyFunc func(target any) any)` | 注册代理工厂 |
| `Get(beanID string) (func(target any) any, bool)` | 获取代理工厂 |
| `Has(beanID string) bool` | 检查是否存在代理 |
| `List() []string` | 列出所有代理 ID |

---

## AOP 模式配置

### AopConfig

AOP 配置结构。

```go
type AopConfig struct {
    Mode        AopMode              // AOP 模式
    Registry    *AopRegistry         // AOP 注册表
    Weaver      *Weaver              // 织入器
    EnableCache bool                 // 是否启用代理缓存
}
```

### AopMode

AOP 模式枚举。

| 模式 | 说明 |
|------|------|
| `AopModeRuntime` | 运行时模式，使用反射和动态代理 |
| `AopModeGenerated` | 代码生成模式，使用编译时代码生成 |
| `AopModeMixed` | 混合模式，自动选择最优方案 |

### 配置示例

```go
// 运行时模式
config := &aop.AopConfig{
    Mode:        aop.AopModeRuntime,
    Registry:    aop.NewAopRegistry(),
    Weaver:      aop.NewWeaver(),
    EnableCache: true,
}

// 代码生成模式
config := &aop.AopConfig{
    Mode:        aop.AopModeGenerated,
    Registry:    aop.NewAopRegistry(),
    Weaver:      aop.NewWeaver(),
    EnableCache: true,
}

// 混合模式（推荐）
config := &aop.AopConfig{
    Mode:        aop.AopModeMixed,
    Registry:    aop.NewAopRegistry(),
    Weaver:      aop.NewWeaver(),
    EnableCache: true,
}
```

---

## 使用示例

### 1. 定义接口和实现

```go
type IUserService interface {
    GetUser(id int) (*User, error)
    CreateUser(user *User) error
    DeleteUser(id int) error
}

type UserService struct{}

func (s *UserService) GetUser(id int) (*User, error) {
    return &User{ID: id, Name: fmt.Sprintf("User%d", id)}, nil
}

func (s *UserService) CreateUser(user *User) error {
    fmt.Printf("Creating user: %+v\n", user)
    return nil
}

func (s *UserService) DeleteUser(id int) error {
    fmt.Printf("Deleting user: %d\n", id)
    return nil
}
```

### 2. 创建切面

使用流式 API 创建切面：

```go
// 前置通知
aop.NewAspectBuilder().
    MatchInterface((*IUserService)(nil)).
    Before(func(jp aop.JoinPoint) {
        fmt.Printf("[Before] Method: %v, Args: %v\n", jp.Method(), jp.Args())
    }).
    Order(1).
    BuildAndRegister()

// 后置通知
aop.NewAspectBuilder().
    MatchInterface((*IUserService)(nil)).
    After(func(jp aop.JoinPoint) {
        fmt.Printf("[After] Method: %v\n", jp.Method())
    }).
    Order(2).
    BuildAndRegister()

// 环绕通知
aop.NewAspectBuilder().
    MatchInterface((*IUserService)(nil)).
    Around(func(jp aop.JoinPoint, proceed aop.ProceedFunc) any {
        start := time.Now()
        fmt.Printf("[Around] Before: %v\n", jp.Method())
        result := proceed(jp.Args()...)
        elapsed := time.Since(start)
        fmt.Printf("[Around] After: %v, Elapsed: %v\n", jp.Method(), elapsed)
        return result
    }).
    Order(3).
    BuildAndRegister()
```

### 3. 使用 AOP 容器

```go
// 创建 AOP 容器
aopContainer := aop.CreateAopContainer()

// 注册服务 Bean
userService := &UserService{}
err := aopContainer.RegisterAopBeanWithID("userService", 
    reflect.TypeOf((*IUserService)(nil)).Elem(), userService)
if err != nil {
    fmt.Printf("Error registering userService: %v\n", err)
    return
}

// 获取代理服务
proxyService, err := aopContainer.GetAopProxy("userService")
if err != nil {
    fmt.Printf("Error getting proxy: %v\n", err)
    return
}

// 使用代理服务
svc := proxyService.(IUserService)
user, err := svc.GetUser(1)
if err != nil {
    fmt.Printf("Error: %v\n", err)
} else {
    fmt.Printf("Got user: %+v\n", user)
}
```

### 4. 使用 AOP 管理器

```go
// 直接使用 AOP 管理器创建代理
proxy := aop.GlobalAopManager.GetProxy(userService)

if proxySvc, ok := proxy.(IUserService); ok {
    user, err := proxySvc.GetUser(1)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    } else {
        fmt.Printf("Got user: %+v\n", user)
    }
}
```

---

## 切面匹配

### 按方法名匹配

```go
aop.NewAspectBuilder().
    MatchByName("GetUser").
    Before(func(jp aop.JoinPoint) {
        fmt.Println("Before GetUser")
    }).
    BuildAndRegister()
```

### 按接口匹配

```go
aop.NewAspectBuilder().
    MatchInterface((*IUserService)(nil)).
    Before(func(jp aop.JoinPoint) {
        fmt.Println("Before any IUserService method")
    }).
    BuildAndRegister()
```

### 按正则表达式匹配

```go
aop.NewAspectBuilder().
    MatchByRegex("^Get.*").
    Before(func(jp aop.JoinPoint) {
        fmt.Println("Before any Get method")
    }).
    BuildAndRegister()
```

### 匹配所有方法

```go
aop.NewAspectBuilder().
    MatchAll().
    Before(func(jp aop.JoinPoint) {
        fmt.Println("Before any method")
    }).
    BuildAndRegister()
```

---

## 通知类型

### 前置通知

```go
aop.Before(func(jp aop.JoinPoint) {
    fmt.Printf("Before: %v\n", jp.Method())
})
```

### 后置通知

```go
aop.After(func(jp aop.JoinPoint) {
    fmt.Printf("After: %v\n", jp.Method())
})
```

### 环绕通知

```go
aop.Around(func(jp aop.JoinPoint, proceed aop.ProceedFunc) any {
    start := time.Now()
    result := proceed(jp.Args()...)
    elapsed := time.Since(start)
    fmt.Printf("Elapsed: %v\n", elapsed)
    return result
})
```

### 返回通知

```go
aop.AfterReturning(func(jp aop.JoinPoint, result any) {
    fmt.Printf("Result: %v\n", result)
})
```

### 异常通知

```go
aop.AfterThrowing(func(jp aop.JoinPoint, err error) {
    fmt.Printf("Error: %v\n", err)
})
```

---

## 性能监控

```go
// 获取 AOP 指标
metrics := aop.GetGlobalAopMetrics()
for key, value := range metrics {
    fmt.Printf("%s: %v\n", key, value)
}

// 输出示例：
// total_proxies: 10
// generated_proxies: 8
// runtime_proxies: 2
// total_aspects: 15
// total_interceptions: 1000
// average_latency: 0.5ms
```

### AOP 指标说明

| 指标 | 说明 |
|------|------|
| `total_proxies` | 总代理数量 |
| `generated_proxies` | 代码生成代理数量 |
| `runtime_proxies` | 运行时代理数量 |
| `total_aspects` | 总切面数量 |
| `total_interceptions` | 总拦截次数 |
| `average_latency` | 平均延迟 |

---

## 最佳实践

### 1. 优先使用代码生成模式

对于生产环境，推荐使用代码生成模式以获得最佳性能。

```go
config := &aop.AopConfig{
    Mode: aop.AopModeGenerated,
}
```

### 2. 使用接口定义

定义清晰的接口，便于 AOP 切面匹配。

```go
type IUserService interface {
    GetUser(id int) (*User, error)
    CreateUser(user *User) error
}
```

### 3. 合理设置切面顺序

使用 Order 参数控制切面执行顺序。

```go
// 日志切面先执行
loggingAspect := &aop.AspectMeta{
    Order: 1,
}

// 事务切面后执行
transactionAspect := &aop.AspectMeta{
    Order: 2,
}
```

### 4. 避免过度使用 AOP

AOP 会增加系统复杂度，只在需要横切关注点时使用。

### 5. 监控性能

使用 AOP 指标监控性能，及时发现性能问题。

```go
metrics := aop.GetGlobalAopMetrics()
fmt.Printf("Average latency: %v\n", metrics["average_latency"])
```

---

## 限制和注意事项

### 1. Go 语言限制

| 限制 | 说明 |
|------|------|
| 无法在运行时创建实现接口的动态类型 | 需要使用代码生成模式 |
| 无法在运行时替换结构体的方法实现 | 需要使用代码生成模式 |
| 即使有接口，也无法真正包装方法调用 | 需要使用代码生成模式 |

### 2. 性能考虑

| 考虑 | 说明 |
|------|------|
| 运行时模式性能较低 | 推荐使用代码生成模式 |
| 混合模式需要额外判断 | 适合一般应用场景 |
| 避免在性能关键路径上使用过多的切面 | 合理使用切面 |

---

## 故障排除

### 切面未执行

**问题**：切面定义正确，但未执行。

**解决方案**：

1. 检查切面目标是否正确
2. 检查切面顺序是否正确
3. 确认是否使用了正确的代理实例
4. 检查方法名是否匹配（区分大小写）

### 性能问题

**问题**：AOP 导致性能下降。

**解决方案**：

1. 使用代码生成模式
2. 启用代理缓存
3. 减少切面数量
4. 优化切面逻辑

### 类型断言错误

**问题**：运行时出现类型断言错误。

**解决方案**：

1. 检查返回值类型是否正确
2. 检查错误处理是否正确
3. 确保代理类正确实现了目标接口

---

## 相关文档

- [aop 包 — 面向切面编程（AOP）框架](./aop.md)
- [aop-codegen 包 — AOP 代码生成工具](./aop-codegen.md)

---

## 示例项目

完整的示例项目位于 `examples/aop-integration/` 目录，包括：

- 使用集成 API 创建切面
- 使用 AOP 容器管理 Bean
- 使用 AOP 管理器创建代理
- 各种切面匹配方式
- 性能监控

---

## 贡献

欢迎贡献代码和文档！

---

## 许可证

MIT License