# config 包 — 配置管理接口与工具

## 概述

`config` 包提供配置管理的核心接口和模型，定义了一套与具体配置源无关的抽象层。包内包含：

- **`Config` 接口** — 统一配置访问方式（`Get`、`GetString`、`Unmarshal`、`Watch` 等）
- **`ConfigOption` / `ConfigModel`** — 配置元数据模型和函数式选项
- **`Loader` 接口 / `LoaderChain`** — 可组合的配置加载器链
- **`Validator` 接口 / `DefaultValidator`** — 配置验证（必填、范围、正则、枚举）
- **`WatchManager`** — 配置热重载事件管理

各配置源实现（如 `viper/`）实现这些接口以接入不同配置源，实现运行时互换。

---

## Config 接口

### 接口定义

```go
type Config interface {
    // 值获取方法
    Get(key string) any
    GetAll() map[string]any
    GetString(key string) string
    GetStringMap(key string) map[string]any
    GetStringMapString(key string) map[string]string
    GetStringSlice(key string) []string
    GetInt(key string) int
    GetInt64(key string) int64
    GetIntSlice(key string) []int
    GetFloat64(key string) float64
    GetBool(key string) bool

    // 键操作方法
    HasKey(key string) bool

    // 结构体映射方法
    Unmarshal(target any) error
    UnmarshalKey(key string, target any) error

    // 热重载方法
    Watch(callback func(WatchEvent)) error
    StopWatch() error

    // 配置源信息
    GetSource() string
}
```

### 值获取方法

支持点分隔的层级键名（如 `server.host`）：

```go
cfg.Get("server.port")                  // any — 原始值
cfg.GetString("app.name")               // string
cfg.GetInt("server.port")               // int
cfg.GetInt64("timeout")                 // int64
cfg.GetFloat64("threshold")             // float64
cfg.GetBool("server.enabled")           // bool
cfg.GetStringSlice("allowed.origins")   // []string
cfg.GetStringMap("database")            // map[string]any
cfg.GetStringMapString("database")      // map[string]string
cfg.GetIntSlice("ports")                // []int
cfg.HasKey("server.host")               // bool
cfg.GetAll()                            // map[string]any — 所有配置
```

### 结构体映射

```go
type AppConfig struct {
    Name    string `mapstructure:"name"`
    Version string `mapstructure:"version"`
    Server  ServerConfig `mapstructure:"server"`
}

type ServerConfig struct {
    Host string `mapstructure:"host"`
    Port int    `mapstructure:"port"`
}

var appCfg AppConfig

// 全量映射
if err := cfg.Unmarshal(&appCfg); err != nil {
    log.Fatal(err)
}

// 指定前缀映射
var srvCfg ServerConfig
if err := cfg.UnmarshalKey("server", &srvCfg); err != nil {
    log.Fatal(err)
}
```

### 热重载

```go
// 注册变更监听
err := cfg.Watch(func(ev config.WatchEvent) {
    fmt.Printf("配置变更: %s %s = %v → %v\n",
        ev.Type, ev.Key, ev.OldValue, ev.NewValue)
})

// 停止监听
err := cfg.StopWatch()
```

`WatchEvent` 结构：

```go
type WatchEvent struct {
    Type     string   // "modify" | "delete" | "create"
    Key      string   // 变更的键名
    OldValue any      // 旧值
    NewValue any      // 新值
}
```

### 配置源信息

```go
source := cfg.GetSource() // 如 "viper", "env", "remote"
```

---

## ConfigModel 配置模型

### 创建

```go
// 使用函数式选项创建配置模型
cfg, err := config.New(
    loadFn,                          // 加载函数（最后执行）
    config.WithConfigName("app"),
    config.WithConfigPath("./config", "/etc/app"),
    config.WithConfigType("yaml"),
    config.WithEnvironment("prod"),
    config.WithEnvVariable("APP"),
    config.WithConfigFile("/etc/app/config.yaml"),
    config.WithDefaultEnv(),
)
```

### 选项说明

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `WithConfigName(name)` | 配置文件名（不含扩展名） | `"config"` |
| `WithConfigPath(paths...)` | 搜索路径 | `["./", "./config"]` |
| `WithConfigType(type)` | 配置类型 | `"json"` |
| `WithEnvironment(env)` | 环境名称（附加到文件名） | `""` |
| `WithConfigFile(path)` | 完整路径（优先于名称+路径） | `""` |
| `WithEnvVariable(name)` | 环境变量前缀 | `""` |
| `WithDefaultEnv()` | 自动检测环境（APP_ENV → GO_ENV → ENV） | - |

### WithDefaultEnv 检测顺序

1. 检查 `APP_ENV` 环境变量
2. 检查 `GO_ENV` 环境变量
3. 检查 `ENV` 环境变量
4. 均未设置时使用空字符串（即无需环境后缀）

---

## Loader 加载器接口

### Loader 接口

```go
type Loader interface {
    Load(opts ...LoaderOption) (Config, error)  // 加载配置
    Priority() int                               // 优先级（值越小越先加载）
    Name() string                                // 加载器名称
    SupportsWatch() bool                         // 是否支持热重载
}
```

### LoaderOption 选项

```go
type LoaderModel struct {
    Paths      []string   // 搜索路径
    FileName   string     // 文件名
    FileType   string     // 文件类型 (yaml/json/toml)
    Env        string     // 环境名称
    Prefix     string     // 环境变量前缀
    RemoteType string     // 远程配置类型 (etcd/consul/apollo)
    Endpoints  []string   // 远程端点
    Key        string     // 配置键
}
```

加载选项函数：

| 选项 | 说明 |
|------|------|
| `WithPaths(paths...)` | 配置搜索路径 |
| `WithFileName(name)` | 文件名 |
| `WithLoaderFileType(typ)` | 文件类型 |
| `WithLoaderEnv(env)` | 环境 |
| `WithPrefix(prefix)` | 环境变量前缀 |
| `WithRemoteType(typ)` | 远程配置类型 |
| `WithEndpoints(endpoints)` | 远程端点 |
| `WithLoaderKey(key)` | 配置键 |

### LoaderChain 加载器链

支持多个加载器按优先级排序，低优先级先加载，高优先级后加载并覆盖同名配置：

```go
chain := &config.LoaderChain{}

// 添加加载器
chain.Add(fileLoader)    // 文件加载器（低优先级）
chain.Add(envLoader)     // 环境变量加载器（中优先级）
chain.Add(remoteLoader)  // 远程配置加载器（高优先级）

// 获取按优先级排序的加载器列表
sorted := chain.Sorted()
```

`LoaderChain` 实现了 `sort.Interface` 接口：

```go
chain.Len()              // 加载器数量
chain.Less(i, j)         // 优先级比较
chain.Swap(i, j)         // 交换位置
chain.Sorted()           // 返回排序后的副本
```

### 实现示例（Viper 加载器）

```go
type ViperFileLoader struct{}

func (l *ViperFileLoader) Load(opts ...config.LoaderOption) (config.Config, error) {
    model := &config.LoaderModel{}
    for _, opt := range opts {
        opt(model)
    }
    v := viper.New()
    v.SetConfigName(model.FileName)
    v.SetConfigType(model.FileType)
    // ... 加载配置
    return &ViperConfig{v: v}, nil
}

func (l *ViperFileLoader) Priority() int   { return 10 }
func (l *ViperFileLoader) Name() string     { return "viper-file" }
func (l *ViperFileLoader) SupportsWatch() bool { return true }
```

---

## Validator 配置验证

### Validator 接口

```go
type Validator interface {
    Validate(target any) error
}
```

### DefaultValidator 使用

```go
v := config.NewValidator()

// 链式构建验证规则
v.AddRequired("db.host", "db.port")        // 必填字段
 .AddMin("db.port", 1)                       // 最小值
 .AddMax("db.port", 65535)                   // 最大值
 .AddRegex("app.email", `^.+@.+\..+$`)      // 正则匹配
 .AddEnum("app.logLevel", "debug", "info", "warn", "error") // 枚举值
 .AddCustomRule("app.rate", func(val any) error {            // 自定义验证
        if v, ok := val.(float64); ok && v > 0 && v <= 1.0 {
            return nil
        }
        return fmt.Errorf("must be 0 < rate <= 1.0")
    })

// 执行验证
data := map[string]any{
    "db.host":    "localhost",
    "db.port":    5432,
    "app.email":  "admin@example.com",
    "app.logLevel": "info",
    "app.rate":   0.5,
}

if err := v.Validate(data); err != nil {
    log.Fatal(err) // 返回 ValidationError
}
```

### Rules 验证规则

```go
type Rules struct {
    Required []string                    // 必填字段
    Min      map[string]int              // 最小值限制
    Max      map[string]int              // 最大值限制
    Regex    map[string]string           // 正则表达式
    Enum     map[string][]any            // 枚举值限制
    Custom   map[string]func(any) error  // 自定义验证函数
}
```

### ValidationError

```go
type ValidationError struct {
    Field   string // 验证失败的字段名
    Message string // 错误描述
}

err := v.Validate(data)
if verr, ok := err.(*config.ValidationError); ok {
    fmt.Printf("字段 %s 验证失败: %s\n", verr.Field, verr.Message)
}
```

验证顺序：必填 → 最小值 → 最大值 → 正则 → 枚举 → 自定义，返回第一个发现的错误。

---

## WatchManager 热重载管理器

```go
// 创建热重载管理器
wm := config.NewWatchManager()
defer wm.Close()

// 注册监听器
wm.Register("logger", func(ev config.WatchEvent) {
    fmt.Printf("[变更] %s %s: %v → %v\n", ev.Type, ev.Key, ev.OldValue, ev.NewValue)
})

// 注册配置源（chan WatchEvent）
ch := make(chan config.WatchEvent, 10)
wm.AddSource("file-watcher", ch)

// 通知所有监听器
wm.Notify(config.WatchEvent{
    Type:     config.EventModify,
    Key:      "server.port",
    OldValue: 8080,
    NewValue: 9090,
})

// 取消监听
wm.Unregister("logger")

// 关闭管理器（关闭所有源通道，清空回调）
wm.Close()
```

### 事件类型常量

```go
config.EventModify = "modify"   // 配置修改
config.EventDelete = "delete"   // 配置删除
config.EventCreate = "create"   // 配置创建
```

### 线程安全

`WatchManager` 使用 `sync.RWMutex` 保护所有操作。已关闭的管理器自动忽略 `Register`、`Unregister`、`AddSource` 和 `Notify` 调用。

---

## 完整示例

### 使用 Viper 集成

```go
package main

import (
    "fmt"
    "log"
    "github.com/xudefa/go-boot/config"
    vipercfg "github.com/xudefa/go-boot/viper"
)

func main() {
    // 使用加载器链
    chain := &config.LoaderChain{}
    chain.Add(&vipercfg.FileLoader{})

    loaders := chain.Sorted()
    var cfg config.Config

    for _, l := range loaders {
        c, err := l.Load(
            config.WithFileName("app"),
            config.WithLoaderFileType("yaml"),
            config.WithLoaderEnv("prod"),
            config.WithPrefix("APP"),
        )
        if err != nil {
            log.Printf("加载器 %s 失败: %v", l.Name(), err)
            continue
        }
        cfg = c
        break
    }

    // 读取配置
    port := cfg.GetInt("server.port")
    host := cfg.GetString("server.host")
    enabled := cfg.GetBool("server.enabled")

    fmt.Printf("Server: %s:%d (enabled=%v)\n", host, port, enabled)

    // 结构体映射
    type DBConfig struct {
        Host     string `mapstructure:"host"`
        Port     int    `mapstructure:"port"`
        Database string `mapstructure:"database"`
    }
    var dbCfg DBConfig
    if err := cfg.UnmarshalKey("database", &dbCfg); err != nil {
        log.Fatal(err)
    }

    // 配置验证
    v := config.NewValidator()
    v.AddRequired("server.host", "server.port")
    v.AddMin("server.port", 1)
    v.AddMax("server.port", 65535)

    if err := v.Validate(cfg.GetAll()); err != nil {
        log.Fatalf("配置验证失败: %v", err)
    }
}
```

### 热重载监听

```go
func watchConfig(cfg config.Config) {
    err := cfg.Watch(func(ev config.WatchEvent) {
        switch ev.Type {
        case config.EventModify:
            fmt.Printf("配置更新: %s = %v\n", ev.Key, ev.NewValue)
        case config.EventDelete:
            fmt.Printf("配置删除: %s\n", ev.Key)
        case config.EventCreate:
            fmt.Printf("配置新增: %s = %v\n", ev.Key, ev.NewValue)
        }
    })
    if err != nil {
        log.Printf("热重载启动失败: %v", err)
    }
}
```

### config.New 创建配置模型

```go
loadFn := func(cfg *config.ConfigModel) error {
    // 自定义加载逻辑
    cfg.Config = map[string]any{
        "server.host": "0.0.0.0",
        "server.port": 8080,
    }
    return nil
}

model, err := config.New(
    loadFn,
    config.WithConfigName("app"),
    config.WithConfigType("yaml"),
    config.WithDefaultEnv(),
)
if err != nil {
    log.Fatal(err)
}
fmt.Println(model.Config)   // 加载后的配置数据
fmt.Println(model.Source)   // 配置源描述
```