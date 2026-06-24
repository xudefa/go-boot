# boot 包 — 应用启动器

## 概述

`boot` 包是 go-boot 框架的应用启动核心，参考 Spring Boot 的 `SpringApplication` 设计，提供了应用启动的全生命周期管理，包括：

- **自动配置**（AutoConfiguration）：各模块通过 `RegisterAutoConfig` 注册自动配置
- **启动器管理**（Starter）：管理组件的启动和停止生命周期
- **横幅打印**（Banner）：应用启动时显示 ASCII 艺术横幅
- **失败分析**（FailureAnalyzer）：启动失败时提供友好的错误提示

## 应用创建与启动

### NewApplication

`NewApplication` 是 go-boot 框架的推荐入口，替代旧的 `start.BuildApp()`。

```go
func NewApplication(opts ...BootOption) (*Boot, error)
```

创建流程：
1. 应用启动选项，生成 `BootConfig`
2. 创建 IoC 容器（`core.New()`）
3. 创建 Environment（`environment.NewEnvironment()`）
4. 创建应用上下文（`DefaultApplicationContext`）
5. 注入激活的 Profile
6. 返回 `Boot` 实例

### 函数式选项

| 选项 | 说明 |
|------|------|
| `WithAppName(name)` | 设置应用名称 |
| `WithVersion(version)` | 设置版本号 |
| `WithProfiles(profiles...)` | 设置激活的 Profile |
| `WithConfigLocation(location)` | 设置配置文件路径 |
| `WithConfigType(configType)` | 设置配置文件类型 (json/yaml) |
| `WithPropertySource(source)` | 添加自定义配置源 |
| `WithoutAutoConfig()` | 禁用自动配置执行 |
| `WithoutStarters()` | 禁用启动器自动管理 |

### Boot 结构体

```go
type Boot struct {
    ctx           *contextpkg.DefaultApplicationContext
    config        *BootConfig
    configLoader  *environment.ConfigLoader
    starters      []Starter
}
```

主要方法：

| 方法 | 说明 |
|------|------|
| `Start()` | 启动应用，执行完整的生命周期 |
| `Stop()` | 停止应用，优雅关闭 |
| `IsRunning()` | 检查应用是否运行中 |
| `WaitForSignal()` | 等待终止信号（SIGINT/SIGTERM），自动优雅关闭 |
| `Context()` | 返回应用上下文 |
| `Container()` | 返回 IoC 容器 |
| `Environment()` | 返回环境配置 |

### 完整示例

```go
package main

import (
    "log"
    "github.com/xudefa/go-boot/boot"
)

func main() {
    app, err := boot.NewApplication(
        boot.WithAppName("my-app"),
        boot.WithVersion("1.0.0"),
        boot.WithProfiles("dev"),
    )
    if err != nil {
        log.Fatal(err)
    }

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
    defer app.Stop()

    app.WaitForSignal()
}
```

## 配置加载

go-boot 支持灵活的配置加载机制，参考 Spring Boot 的设计思想。

### 配置文件格式

支持 JSON 格式。

### 配置文件命名

- 基础配置：`application.json`
- 环境特定配置：`application-{profile}.json`

### 配置文件搜索路径

按以下优先级搜索配置文件：
按以下优先级搜索配置文件：
1. `/etc/config`
2. 当前目录
3. `./config`
4. 可执行文件所在目录
5. 可执行文件所在目录的 `./config`

### 配置加载顺序

1. 加载基础配置
2. 加载环境特定配置
3. 环境特定配置覆盖基础配置

### 使用示例

```go
// 使用默认配置
app, err := boot.NewApplication(
    boot.WithProfiles("dev"),
)

// 指定配置文件路径
app, err := boot.NewApplication(
    boot.WithConfigLocation("/path/to/config.json"),
)

// 添加自定义配置源
customSource := environment.NewMapPropertySource(
    "custom",
    environment.PriorityHigh,
    map[string]any{
        "app.name": "custom-app",
    },
)
app, err := boot.NewApplication(
    boot.WithPropertySource(customSource),
)
```

## 启动流程

Boot.Start() 按照以下阶段顺序执行：

```
PhaseInitializing
    ↓
PhaseConfiguring
    ├── 执行 AutoConfiguration（按顺序）
    ├── 注册 Starter（拓扑排序）
    ├── 调用 Starter.Configure()
    └── 发布 EventEnvironmentPrepared
    ↓
PhaseContextRefreshed
    ├── 发布 EventContextRefreshed
    ↓
PhaseReady
    ├── 调用 Starter.Start()
    ├── 打印 Banner
    ↓
PhaseRunning
    ├── 发布 EventApplicationStarted
    └── 发布 EventApplicationReady
```

停止流程（Boot.Stop()）：

```
PhaseRunning
    ↓
PhaseStopping
    ├── 逆序停止 Starter（反向依赖顺序）
    ↓
PhaseStopped
    └── 发布 EventApplicationStopped
```

## 自动配置（AutoConfiguration）

### AutoConfiguration 接口

```go
type AutoConfiguration interface {
    Configure(ctx ApplicationContext) error
}
```

各模块实现此接口，通过 `RegisterAutoConfig` 注册。

### 注册自动配置

```go
func init() {
    boot.RegisterAutoConfig(&GinAutoConfiguration{},
        condition.OnProperty("gin.enabled", "true"),
    )
}
```

或使用带选项的注册：

```go
func init() {
    boot.RegisterAutoConfigWith(&GinAutoConfiguration{},
        boot.WithOrder(10),
        boot.WithDependsOn("dbConfig", "redisConfig"),
    )
}
```

### AutoConfigEntry

```go
type AutoConfigEntry struct {
    Config       AutoConfiguration
    Conditions   []condition.Condition
    Order        int
    Dependencies []string
}
```

### AutoConfigRegistry

`AutoConfigRegistry` 管理全局自动配置条目：

- `RegisterAutoConfig(config, conditions...)` — 注册到全局注册表
- `RegisterAutoConfigWith(config, opts...)` — 注册（支持选项）
- `GetAll()` — 获取所有注册的配置
- `GetMatching(ctx)` — 获取匹配条件的配置（按 Order 排序）

匹配逻辑：所有 Condition 必须全部满足，结果按 Order 升序排列。

## 启动器（Starter）

### Starter 接口

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

每个集成的 Starter 管理其自身的生命周期，提供 Configure → Start → Stop 三阶段控制。

### StarterRegistry

通过 `RegisterStarter` 注册启动器：

```go
func init() {
    boot.RegisterStarter(MyStarter{})
}
```

### 依赖拓扑排序

`GetOrdered()` 使用 **Kahn 算法**对启动器进行拓扑排序：

1. 构建依赖关系有向图
2. 计算每个节点的入度
3. 从入度为 0 的节点开始，逐层移除并加入结果
4. 如果存在循环依赖，回退到原始顺序

```go
A → B → C
A → C

排序结果: [A, B, C]
```

逆序停止：停止时按排序结果的逆序执行，保证依赖方先于被依赖方停止。

## 横幅（Banner）

### Banner 接口

```go
type Banner interface {
    Print(ctx ApplicationContext)
}
```

### 内置实现

| 类型 | 说明 |
|------|------|
| `LegacyBanner` | 默认 ASCII 艺术横幅 |
| `TextBanner` | 文本横幅，支持属性模板 |
| `ASCIIArtBanner` | ASCII 艺术横幅 |
| `CustomTemplateBanner` | 自定义模板横幅 |

### 默认横幅

```go
var DefaultBanner = &LegacyBanner{...}
```

显示效果：

```
  ________                  ____             __
 /  _____/  ____    ____   / __ )  ____     / /_
/   \  ___ /  _ \  /  _ \ / __  | / __ \   / __ \
\    \_\  (  <_> )(  <_> ) /_/ / / /_/ /  / /_/ /
 \______  /\____/  \____/_____/  \____/   /___/
        \/
:: my-app :: v1.0.0 :: profiles(dev)
```

## 失败分析（FailureAnalyzer）

### FailureAnalyzer 接口

```go
type FailureAnalyzer interface {
    CanAnalyze(err error) bool
    Analyze(err error) *FailureReport
}
```

### FailureReport

```go
type FailureReport struct {
    Headline          string
    Description       string
    Action            string
    Cause             string
    Details           map[string]any
    StackTrace        string
    PossibleSolutions []string
}
```

### 注册失败分析器

```go
boot.RegisterFailureAnalyzer(MyAnalyzer{})
```

或使用 `SimpleFailureAnalyzer`：

```go
boot.RegisterFailureAnalyzer(
    boot.NewSimpleFailureAnalyzer(func(err error) *boot.FailureReport {
        if strings.Contains(err.Error(), "port") {
            return &boot.FailureReport{
                Description: "端口被占用",
                Action:      "请检查端口是否被其他进程占用",
                Cause:       err.Error(),
            }
        }
        return nil
    }),
)
```

### 失败报告输出

```
====================
APPLICATION FAILED TO START
====================

描述: 端口被占用

动作: 请检查端口是否被其他进程占用

原因: listen tcp :8080: bind: address already in use

可能的解决方案:
  1. 检查端口占用：lsof -i :8080
  2. 修改配置文件中的端口号
  3. 停止占用端口的进程
```

## 结构化启动错误（BootError）

`BootError` 是框架提供的结构化启动错误类型，包含错误发生的阶段、原始错误、分析结果和修复建议：

```go
type BootError struct {
    Phase       string   // 错误发生的阶段（如 "configuring", "starting"）
    Original    error    // 原始错误
    Analyzed    string   // FailureAnalyzer 分析结果
    Suggestions []string // 修复建议
}
```

### 错误阶段

| 阶段 | 说明 |
|------|------|
| `configuring` | 配置加载和自动配置阶段 |
| `context-refreshed` | 上下文刷新阶段 |
| `starting` | 启动器启动阶段 |
| `stopping` | 停止阶段 |

### 使用示例

```go
app, err := boot.NewApplication()
if err != nil {
    // 错误已包含阶段信息
    if bootErr, ok := err.(*boot.BootError); ok {
        fmt.Printf("启动失败阶段: %s\n", bootErr.Phase)
        fmt.Printf("原始错误: %v\n", bootErr.Original)
        if bootErr.Analyzed != "" {
            fmt.Printf("分析: %s\n", bootErr.Analyzed)
        }
        for _, s := range bootErr.Suggestions {
            fmt.Printf("建议: %s\n", s)
        }
    }
}
```

### 与 FailureAnalyzer 集成

`Boot.reportError` 方法自动使用 `FailureAnalyzer` 分析错误，并将分析结果填充到 `BootError` 中：

```go
func (b *Boot) reportError(phase string, err error) error {
    bootErr := NewBootError(phase, err)
    
    // 使用 FailureAnalyzer 分析
    report := globalAnalyzerRegistry.Analyze(err)
    if report != nil {
        bootErr.WithAnalysis(report.Description)
        bootErr.WithSuggestions(report.PossibleSolutions...)
        fmt.Fprintf(os.Stderr, "\n%s\n", formatFailure(report))
    }
    
    return bootErr
}
```

## 适配器模式

### appCtxAdapter

将 `DefaultApplicationContext` 适配为 `boot.ApplicationContext`，解决 Go 接口签名差异问题。

### conditionCtx

将 `DefaultApplicationContext` 适配为 `condition.ConditionContext`，供条件判断系统使用。