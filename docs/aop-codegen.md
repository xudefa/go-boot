# aop-codegen 包 — AOP 代码生成工具

## 概述

`aop-codegen` 包为 Go 语言提供编译时代理生成功能，解决运行时反射限制，实现真正的 AOP 拦截。通过编译时代码生成，创建静态代理类，实现完整的 AOP 功能。

### 核心特性

| 特性 | 说明 |
|------|------|
| **编译时织入** | 在编译时生成代理代码，无运行时反射开销 |
| **完整支持** | 支持所有方法类型，包括结构体方法、接口方法 |
| **类型安全** | 编译时类型检查，避免运行时错误 |
| **高性能** | 直接方法调用，性能接近原生代码 |
| **注解驱动** | 使用注解定义切面和代理，简洁直观 |

### Go 语言的 AOP 限制

由于 Go 语言的反射限制，运行时 AOP 有以下限制：

- 无法在运行时创建实现接口的动态类型
- 无法在运行时替换结构体的方法实现
- 即使有接口，也无法真正包装方法调用

代码生成工具通过编译时代码生成，完美解决了这些限制。

---

## 安装

```bash
cd cmd/goaop
go build -o goaop
```

---

## 快速开始

### 1. 定义服务

```go
// @AopProxy bean=userService
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

### 2. 定义切面

```go
// @Aspect order=1
type LoggingAspect struct{}

// @Before target=UserService.GetUser
func (a *LoggingAspect) BeforeLog(jp aop.JoinPoint) {
    fmt.Printf("[Before] Method: %v, Args: %v\n", jp.Method(), jp.Args())
}

// @After target=UserService.GetUser
func (a *LoggingAspect) AfterLog(jp aop.JoinPoint) {
    fmt.Printf("[After] Method: %v\n", jp.Method())
}

// @Around target=UserService.GetUser
func (a *LoggingAspect) AroundLog(jp aop.JoinPoint, proceed aop.ProceedFunc) any {
    start := time.Now()
    fmt.Printf("[Around] Before: %v\n", jp.Method())
    result := proceed(jp.Args()...)
    elapsed := time.Since(start)
    fmt.Printf("[Around] After: %v, Elapsed: %v\n", jp.Method(), elapsed)
    return result
}
```

### 3. 生成代理代码

```bash
# 生成代理代码
goaop generate -dir .

# 清理生成的代码
goaop clean -dir .
```

### 4. 编译和运行

```bash
# 使用 goaop 标签编译
go build -tags goaop -o app

# 运行
./app
```

---

## 注解语法

### @AopProxy

标记需要生成代理的结构体。

**参数**：

| 参数 | 说明 | 必需 |
|------|------|------|
| `bean` | Bean ID，用于在容器中注册 | 否，默认为结构体名小写 |

**示例**：

```go
// @AopProxy bean=userService
type UserService struct{}
```

### @Aspect

标记切面类。

**参数**：

| 参数 | 说明 | 必需 |
|------|------|------|
| `order` | 执行顺序，数字越小越先执行 | 否，默认为 0 |

**示例**：

```go
// @Aspect order=1
type LoggingAspect struct{}
```

### @Before

定义前置通知。

**参数**：

| 参数 | 说明 | 必需 |
|------|------|------|
| `target` | 目标方法，格式为 `Struct.Method` | 是 |

**示例**：

```go
// @Before target=UserService.GetUser
func (a *LoggingAspect) BeforeLog(jp aop.JoinPoint) {
    fmt.Printf("[Before] Method: %v\n", jp.Method())
}
```

### @After

定义后置通知。

**参数**：

| 参数 | 说明 | 必需 |
|------|------|------|
| `target` | 目标方法，格式为 `Struct.Method` | 是 |

**示例**：

```go
// @After target=UserService.GetUser
func (a *LoggingAspect) AfterLog(jp aop.JoinPoint) {
    fmt.Printf("[After] Method: %v\n", jp.Method())
}
```

### @Around

定义环绕通知。

**参数**：

| 参数 | 说明 | 必需 |
|------|------|------|
| `target` | 目标方法，格式为 `Struct.Method` | 是 |

**示例**：

```go
// @Around target=UserService.GetUser
func (a *LoggingAspect) AroundLog(jp aop.JoinPoint, proceed aop.ProceedFunc) any {
    start := time.Now()
    result := proceed(jp.Args()...)
    elapsed := time.Since(start)
    fmt.Printf("[Around] Elapsed: %v\n", elapsed)
    return result
}
```

### @AfterReturning

定义返回通知。

**参数**：

| 参数 | 说明 | 必需 |
|------|------|------|
| `target` | 目标方法，格式为 `Struct.Method` | 是 |

**示例**：

```go
// @AfterReturning target=UserService.GetUser
func (a *LoggingAspect) AfterReturningLog(jp aop.JoinPoint, result any) {
    fmt.Printf("[AfterReturning] Result: %v\n", result)
}
```

### @AfterThrowing

定义异常通知。

**参数**：

| 参数 | 说明 | 必需 |
|------|------|------|
| `target` | 目标方法，格式为 `Struct.Method` | 是 |

**示例**：

```go
// @AfterThrowing target=UserService.GetUser
func (a *LoggingAspect) AfterThrowingLog(jp aop.JoinPoint, err error) {
    fmt.Printf("[AfterThrowing] Error: %v\n", err)
}
```

---

## 命令行工具

### generate

生成代理代码。

**用法**：

```bash
goaop generate [options]
```

**选项**：

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `-dir, --dir` | 指定扫描目录 | 当前目录 |
| `-output, --output` | 指定输出目录 | 当前目录 |
| `-clean, --clean` | 生成前清理旧代码 | false |

**示例**：

```bash
# 生成当前目录的代理代码
goaop generate

# 生成指定目录的代理代码
goaop generate -dir ./services

# 生成并清理旧代码
goaop generate -dir . -clean
```

### clean

清理生成的代理代码。

**用法**：

```bash
goaop clean [options]
```

**选项**：

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `-dir, --dir` | 指定清理目录 | 当前目录 |

**示例**：

```bash
# 清理当前目录的代理代码
goaop clean

# 清理指定目录的代理代码
goaop clean -dir ./services
```

### version

显示版本信息。

**用法**：

```bash
goaop version
```

---

## 生成的代码结构

代码生成工具会为每个标记为 `@AopProxy` 的结构体生成对应的代理类。

### 代理类结构

```go
// UserServiceProxy 生成的代理类
type UserServiceProxy struct {
    target  *UserService
    aspects []aop.AspectMeta
}

// NewUserServiceProxy 创建代理实例
func NewUserServiceProxy(target *UserService) *UserServiceProxy {
    proxy := &UserServiceProxy{
        target:  target,
        aspects: []aop.AspectMeta{
            // 切面元数据
        },
    }
    return proxy
}

// GetUser 代理方法
func (p *UserServiceProxy) GetUser(id int) (*User, error) {
    // 创建方法调用信息
    jp := &aop.MethodInvocation{
        Func:   p.target.GetUser,
        Params: []any{id},
        Object: p.target,
    }
    
    // 执行通知链
    result := aop.ExecuteChain(jp, p.aspects)
    
    // 处理返回值
    if tuple, ok := result.([]any); ok {
        if len(tuple) > 0 {
            if err, ok := tuple[len(tuple)-1].(error); ok && err != nil {
                return tuple[0].(*User), err
            }
        }
        return tuple[0].(*User)
    }
    return nil, nil
}
```

### 切面适配器

为每个切面方法生成适配器，将切面方法转换为 `aop.Advice` 接口。

```go
// LoggingAspectBeforeLogAdapter 前置通知适配器
type LoggingAspectBeforeLogAdapter struct {
    aspect *LoggingAspect
}

func (a *LoggingAspectBeforeLogAdapter) Type() aop.AdviceType {
    return aop.AdviceBefore
}

func (a *LoggingAspectBeforeLogAdapter) Apply(jp aop.JoinPoint, proceed aop.ProceedFunc) any {
    a.aspect.BeforeLog(jp)
    return nil
}
```

### Bean 注册

自动生成 Bean 注册代码，将代理注册到 IoC 容器。

```go
func init() {
    globalContainer.Register("userService", core.Factory(func(c core.Container) (any, error) {
        target := &UserService{}
        return NewUserServiceProxy(target), nil
    }, reflect.TypeOf(&UserService{})), core.Singleton())
}
```

---

## 高级特性

### 多切面组合

一个方法可以应用多个切面，切面按 `order` 参数排序执行。

```go
// @Aspect order=1
type LoggingAspect struct{}

// @Before target=UserService.GetUser
func (a *LoggingAspect) BeforeLog(jp aop.JoinPoint) {
    fmt.Printf("[Logging] Before: %v\n", jp.Method())
}

// @Aspect order=2
type TransactionAspect struct{}

// @Before target=UserService.GetUser
func (a *TransactionAspect) BeginTransaction(jp aop.JoinPoint) {
    fmt.Printf("[Transaction] Begin: %v\n", jp.Method())
}
```

执行顺序：

```
1. LoggingAspect.BeforeLog (order=1)
2. TransactionAspect.BeginTransaction (order=2)
3. 目标方法
4. TransactionAspect.CommitTransaction (order=2)
5. LoggingAspect.AfterLog (order=1)
```

### 返回值处理

代码生成工具自动处理各种返回值类型：

| 返回值类型 | 处理方式 |
|------------|----------|
| 无返回值 | 直接调用，不处理返回值 |
| 单个返回值 | 直接返回 |
| 多个返回值 | 转换为 `[]any` 处理 |
| 错误返回值 | 检查最后一个返回值是否为 `error` |

### 错误处理

自动处理方法调用的错误，确保错误正确传播。

```go
func (p *UserServiceProxy) GetUser(id int) (*User, error) {
    // ... 执行通知链
    
    if tuple, ok := result.([]any); ok {
        if len(tuple) > 0 {
            // 检查错误
            if err, ok := tuple[len(tuple)-1].(error); ok && err != nil {
                return tuple[0].(*User), err
            }
        }
        return tuple[0].(*User)
    }
    return nil, nil
}
```

---

## 性能优化

### 1. 编译时织入

代理代码在编译时生成，无运行时反射开销。

### 2. 直接方法调用

生成的代理代码直接调用目标方法，性能接近原生代码。

### 3. 类型安全

编译时类型检查，避免运行时类型转换错误。

### 4. 最小化包装

只包装需要 AOP 的方法，其他方法直接调用。

---

## 最佳实践

### 1. 合理使用注解

只在需要 AOP 的方法上添加注解，避免过度使用。

### 2. 设置切面顺序

合理设置切面的 `order` 参数，确保正确的执行顺序。

### 3. 性能监控

使用 AOP 指标监控性能，及时发现性能问题。

```go
metrics := aop.GetGlobalAopMetrics()
fmt.Printf("Average latency: %v\n", metrics["average_latency"])
```

### 4. 错误处理

在切面中正确处理错误，避免吞没异常。

### 5. 避免循环依赖

避免在切面中调用可能触发其他切面的方法，防止循环调用。

---

## 限制和注意事项

### 1. Go 语言限制

| 限制 | 说明 |
|------|------|
| 无法在运行时创建实现接口的动态类型 | 需要编译时代码生成 |
| 无法在运行时替换结构体的方法实现 | 需要编译时代码生成 |
| 需要编译时代码生成 | 必须重新编译才能生效 |

### 2. 代码生成限制

| 限制 | 说明 |
|------|------|
| 需要重新编译才能生效 | 修改切面后需要重新生成代码 |
| 生成的代码需要手动维护 | 生成的代码不应手动修改 |
| 不支持动态修改切面 | 切面在编译时确定 |

### 3. 性能考虑

| 考虑 | 说明 |
|------|------|
| 虽然性能很高，但仍有一定开销 | 避免在性能关键路径上使用过多的切面 |
| 避免在性能关键路径上使用过多的切面 | 合理使用缓存减少重复计算 |

---

## 故障排除

### 代理代码未生成

**问题**：运行 `goaop generate` 后没有生成代理代码。

**解决方案**：

1. 检查是否正确添加了 `@AopProxy` 注解
2. 检查注解格式是否正确
3. 查看生成日志，确认是否有错误

### 编译错误

**问题**：生成的代码编译失败。

**解决方案**：

1. 检查目标方法的签名是否正确
2. 检查切面方法的签名是否正确
3. 确保所有依赖都已正确导入

### 运行时错误

**问题**：运行时出现类型断言错误。

**解决方案**：

1. 检查返回值类型是否正确
2. 检查错误处理是否正确
3. 确保代理类正确实现了目标接口

### 切面未执行

**问题**：切面定义正确，但未执行。

**解决方案**：

1. 检查切面目标是否正确
2. 检查切面顺序是否正确
3. 确认是否使用了正确的代理实例

---

## 相关文档

- [aop 包 — 面向切面编程（AOP）框架](./aop.md)
- [aop-integration 包 — AOP 集成指南](./aop-integration.md)

---

## 示例项目

完整的示例项目位于 `examples/aop-generated/` 目录，包括：

- 服务定义
- 切面定义
- 代理代码生成
- 编译和运行

---

## 贡献

欢迎贡献代码和文档！

---

## 许可证

MIT License