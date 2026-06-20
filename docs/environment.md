# environment 包 — 分层环境配置管理

## 概述

`environment` 包提供分层配置源（PropertySource）管理和 Profile 机制，参考 Spring Framework 的 `Environment` 抽象设计。核心职责：

- 从命令行参数、环境变量、配置文件等多种来源读取配置
- 按优先级覆盖，高优先级属性自动覆盖低优先级同名字段
- 支持 `${...}` 占位符递归解析
- 支持 Profile 激活匹配（类似 Spring 的 `@Profile`）
- 支持类型安全的属性获取和结构体绑定

### 配置源优先级（从高到低）

| 优先级 | 常量 | 配置源 | 说明 |
|--------|------|--------|------|
| 最高 (4) | `PriorityHighest` | `ArgsPropertySource` | 命令行参数 `--key=value` |
| 高 (3) | `PriorityHigh` | `EnvPropertySource` | 环境变量（默认 `GO_BOOT_` 前缀） |
| 中 (2) | `PriorityNormal` | `MapPropertySource` | 用户动态添加的配置源 |
| 中 (1) | `PriorityLow` | - | 保留 |
| 最低 (0) | `PriorityLowest` | 应用配置源 | `application.json` 或 `JSONPropertySource` |

---

## PropertySource 配置源

### PropertySource 接口

```go
type PropertySource interface {
    Name() string                              // 配置源名称
    GetProperty(key string) (any, bool)        // 获取属性值
    Priority() Priority                        // 优先级（值越大越高）
    Contains(key string) bool                  // 检查键是否存在
}
```

### Priority 优先级常量

```go
type Priority int

const (
    PriorityLowest  Priority = iota  // 最低，默认配置使用
    PriorityLow                      // 低
    PriorityNormal                   // 正常
    priorityHigh                     // 高，环境变量使用（外部可用）
    priorityHighest                  // 最高，命令行参数使用
)
```

### MapPropertySource

基于 `map[string]any` 的内存配置源，可用于注入默认配置或测试数据。

```go
// 创建普通配置源（自定义优先级）
src := environment.NewMapPropertySource("myConfig", environment.PriorityNormal, map[string]any{
    "server.port": 8080,
    "app.name":    "myapp",
})

// 创建最低优先级默认配置源
defaults := environment.NewDefaultPropertySource("defaults", map[string]any{
    "server.port": 3000,
    "app.name":    "default",
})

src.Name()                    // "myConfig"
src.GetProperty("server.port") // 8080, true
src.Contains("app.name")       // true
src.Priority()                 // PriorityNormal (2)
src.Keys()                     // ["server.port", "app.name"]
```

### EnvPropertySource

基于环境变量的配置源，支持前缀映射。例如前缀为 `GO_BOOT` 时，键 `server.port` 将映射为环境变量 `GO_BOOT_SERVER_PORT`。

```go
// 创建环境变量配置源，前缀 "GO_BOOT"
envSrc := environment.NewEnvPropertySource("env", "GO_BOOT")

// 设置环境变量 GO_BOOT_SERVER_PORT=9090
// 则 envSrc.GetProperty("server.port") 返回 "9090", true
envSrc.Name()     // "env"
envSrc.Priority() // priorityHigh (3)
```

键名转换规则（`toEnvKey`）：
- `.` 和 `-` 转换为 `_`
- 小写字母转换为大写
- 例如：`server.port` → `SERVER_PORT` → 加前缀 → `GO_BOOT_SERVER_PORT`

通过可替换的 `lookupEnv` 变量便于测试：

```go
// 在测试中替换
environment.LookupEnv = func(key string) (string, bool) {
    table := map[string]string{
        "GO_BOOT_SERVER_PORT": "9090",
    }
    val, ok := table[key]
    return val, ok
}
```

### ArgsPropertySource

基于命令行参数的配置源，优先级最高。解析 `--key=value` 格式的参数。

```go
// 从 os.Args 解析命令行参数
// 例如 os.Args = ["./app", "--server.port=8080", "--app.name=myapp"]
argsSrc := environment.NewArgsPropertySource("args", os.Args)

// 实际应用中，NewEnvironment() 自动包含 args 和 env 源
argsSrc.GetProperty("server.port") // "8080", true
argsSrc.Priority()                 // priorityHighest (4)
```

---

## Environment 环境管理器

### 创建

```go
// 创建默认环境（自动包含 ArgsPropertySource + EnvPropertySource）
env := environment.NewEnvironment()

// 添加额外配置源
env.AddPropertySource(environment.NewMapPropertySource("app", environment.PriorityNormal, data))
```

`NewEnvironment()` 默认注册：
1. `ArgsPropertySource("args", os.Args)` — 命令行参数
2. `EnvPropertySource("env", "GO_BOOT")` — 环境变量

### 属性获取

```go
env.GetProperty("server.port")         // any, bool
env.GetString("server.port", "8080")   // string（含默认值）
env.GetInt("server.port", 8080)        // int（含默认值）
env.GetBool("server.enabled", true)   // bool（含默认值）
env.GetFloat64("server.port", 8080.0) // float64（含默认值）
env.ContainsProperty("server.port")   // bool
env.IsPropertyEmpty("server.port")    // bool
env.GetRequiredProperty("server.port") // any, error（不存在时返回错误）
```

类型转换自动支持：
- `GetInt`：支持 `int`、`float64`、数字字符串
- `GetBool`：支持 `bool`、`"true"/"false"` 字符串
- `GetFloat64`：支持 `float64`、`int`、数字字符串

## TypeConverter（类型转换器）

`TypeConverter` 提供统一的类型转换逻辑，支持多种类型间的转换。

### 支持的转换

| 源类型 | 目标类型 | 说明 |
|--------|----------|------|
| int/float64/string | int | 自动解析字符串 |
| float64/int/string | float64 | 自动解析字符串 |
| bool/string | bool | 支持 "true"/"false" |
| any | string | 格式化为字符串（避免 ASCII 转换） |

### 使用示例

```go
converter := environment.NewTypeConverter()

// 转换为 int
result, err := converter.ConvertTo("123", reflect.TypeOf(int(0)))
// result = 123

// 转换为 bool
result, err := converter.ConvertTo("true", reflect.TypeOf(false))
// result = true

// 数值转字符串（避免 ASCII 转换）
result, err := converter.ConvertTo(42, reflect.TypeOf(""))
// result = "42"（不是 "*"）
```

### 设计特点

- **避免 ASCII 转换**：数值类型到 string 的转换使用格式化而非 ASCII 码
- **错误处理**：转换失败返回明确的错误信息
- **扩展性**：支持所有基本数值类型和 bool/string 之间的转换

### 配置源管理

```go
// 添加配置源（按优先级自动排序）
env.AddPropertySource(src)

// 添加最高优先级配置源
env.AddPropertySourceFirst(src)

// 按名称移除配置源
env.RemovePropertySource("args")

// 获取所有配置源副本（按优先级升序排列）
sources := env.GetPropertySources()

// 验证所有配置源
errs := env.Validate()
```

### 配置源排序

添加配置源后自动按 `Priority()` 升序排列。查找属性时从列表尾部（高优先级）向前遍历，返回第一个匹配值。

### Profile 支持

```go
// 激活 Profile
env.AddActiveProfile("dev")
env.AddActiveProfile("test")

// 检查 Profile 是否匹配（支持否定前缀 !）
env.AcceptsProfile("dev")      // true（已激活）
env.AcceptsProfile("!dev")     // false（已激活时取反）
env.AcceptsProfile("prod")     // false（未激活）
env.AcceptsProfile("!prod")    // true（未激活时取反）

// 获取激活的 Profile 列表
profiles := env.GetActiveProfiles() // ["dev", "test"]

// 移除 Profile
env.RemoveProfile("test")

// 获取 Profile 激活状态（从命令行或环境变量）
profile := environment.GetProfileActive(os.Args) // 优先 --profile=，其次 GO_BOOT_PROFILE

// 解析逗号分隔的 Profile 字符串
environment.ParseProfiles("dev,test") // ["dev", "test"]
```

### 占位符解析

支持 `${key}` 和 `${key:defaultValue}` 语法，自动递归解析：

```go
env.AddPropertySource(environment.NewMapPropertySource("app", environment.PriorityNormal, map[string]any{
    "app.name": "myapp",
    "app.host": "localhost",
    "app.url":  "http://${app.host}:${app.port:8080}",
}))

// 自动递归解析占位符
val, _ := env.GetProperty("app.url") // "http://localhost:8080"

// 显式解析
result := env.ResolvePlaceholders("connecting to ${app.host}:${app.port:9090}")
// "connecting to localhost:9090"
```

支持嵌套占位符和循环引用检测：

```go
// 嵌套：defaultValue 中也可以使用占位符
// ${host:${default.host}} — 先查 host，不存在则查 default.host

// 循环引用检测：A → B → A 时记录日志并保留原始文本
```

实现机制：`parsePlaceholder` 从 `$` 开始解析，支持嵌套 `${}`（depth 计数），`resolvePlaceholders` 递归解析并维护 `resolving` map 检测循环引用。

### 结构体绑定

支持将配置绑定到 Go 结构体，通过 `env` 标签指定键名：

```go
type AppConfig struct {
    Host    string `env:"server.host"`
    Port    int    `env:"server.port"`
    Enabled bool   `env:"server.enabled"`
}

var cfg AppConfig

// 全量绑定（按字段名或 env 标签匹配）
if err := env.Bind(&cfg); err != nil {
    log.Fatal(err)
}

// 前缀绑定（自动添加前缀）
// 相当于绑定 server.host、server.port、server.enabled
if err := env.BindPrefix("server", &cfg); err != nil {
    log.Fatal(err)
}

// 指定键绑定到目标
var port int
if err := env.BindKey("server.port", &port); err != nil {
    log.Fatal(err)
}
```

绑定流程：
1. 遍历结构体字段，检查 `env` 标签，未标记者使用字段名
2. 递归处理嵌套结构体（键名用 `.` 连接）
3. 自动类型转换：`float64 → int`（JSON/YAML 反序列化常见）、`int → float64`、`string` 转换等

### 完整示例

```go
package main

import (
    "fmt"
    "log"
    "github.com/xudefa/go-boot/environment"
)

type ServerConfig struct {
    Host string `env:"host"`
    Port int    `env:"port"`
}

func main() {
    env := environment.NewEnvironment()

    // 添加默认配置源（最低优先级）
    env.AddPropertySource(environment.NewDefaultPropertySource("defaults", map[string]any{
        "server.host": "0.0.0.0",
        "server.port": 3000,
    }))

    // 添加应用配置源
    env.AddPropertySource(environment.NewMapPropertySource("app", environment.PriorityNormal, map[string]any{
        "server.port": 8080,
    }))

    // 激活 Profile
    env.AddActiveProfile("dev")

    // 获取属性
    fmt.Println(env.GetString("server.host", "localhost")) // "0.0.0.0"
    fmt.Println(env.GetInt("server.port", 8080))          // 8080

    // 绑定到结构体
    var cfg ServerConfig
    if err := env.BindPrefix("server", &cfg); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%+v\n", cfg) // {Host:0.0.0.0 Port:8080}

    // 占位符解析
    fmt.Println(env.ResolvePlaceholders("http://${server.host}:${server.port}"))
    // "http://0.0.0.0:8080"
}
```

### 线程安全

`Environment` 使用 `sync.RWMutex` 保护所有字段，读写操作均加锁。`GetProperty` 等读操作使用 `RLock`，`AddPropertySource` 等写操作使用 `Lock`。

## 配置加载器 (ConfigLoader)

ConfigLoader 提供了统一的配置文件加载接口，支持多种配置格式和环境特定配置。

### 创建配置加载器

```go
loader := environment.NewConfigLoader(
    "application",           // 配置文件名
    environment.ConfigTypeJSON, // 配置类型
    "",                      // 配置文件路径（空则自动搜索）
    []string{"dev"},         // 激活的 Profile
)
```

### 加载配置

```go
sources, err := loader.Load()
if err != nil {
    log.Fatal(err)
}

for _, source := range sources {
    env.AddPropertySource(source)
}
```

### 支持的配置类型

- `ConfigTypeJSON`: JSON 格式
- `ConfigTypeYAML`: YAML 格式

### 配置文件搜索

ConfigLoader 会按优先级在多个位置搜索配置文件，详见 [配置文件搜索路径](#配置文件搜索路径)。

### 配置加载顺序

1. 如果指定了 `configLocation`，直接加载该配置文件
2. 否则，按优先级搜索并加载基础配置文件（如 `application.json`）
3. 对于每个激活的 Profile，搜索并加载环境特定配置文件（如 `application-dev.json`）
4. 环境特定配置会覆盖基础配置中的同名属性

### 配置文件搜索路径

按以下优先级搜索配置文件：
1. `/etc/config`
2. 当前目录
3. `./config`
4. 可执行文件所在目录
5. 可执行文件所在目录的 `./config`

### 使用示例

```go
// 创建配置加载器
loader := environment.NewConfigLoader(
    "application",
    environment.ConfigTypeJSON,
    "",
    []string{"dev", "test"},
)

// 加载配置
sources, err := loader.Load()
if err != nil {
    log.Fatal(err)
}

// 将配置源添加到环境
for _, source := range sources {
    env.AddPropertySource(source)
}

// 访问配置
port := env.GetInt("app.port", 8080)
debug := env.GetBool("app.debug", false)
```

### 配置文件类型解析

```go
// 从文件路径解析配置类型
configType := environment.ParseConfigType("/path/to/config.yaml")
// 返回 ConfigTypeYAML

configType = environment.ParseConfigType("/path/to/config.json")
// 返回 ConfigTypeJSON

// 获取配置文件扩展名
ext := environment.GetConfigFileExtension(environment.ConfigTypeYAML)
// 返回 "yaml"
```