# log 包 — 日志抽象层

## 概述

`log` 包提供统一的日志记录接口抽象，允许不同的日志库实现无缝替换。内置基于 Go 标准库 `log/slog` 的默认实现 `SlogLogger`，支持 JSON 和 text 两种输出格式。

主要组件：

- `Logger` 接口 — 核心日志记录接口
- `LoggerWithLevel` / `LoggerWithName` / `LoggerWithCaller` / `LoggerWithTimeout` — 扩展接口
- `SlogLogger` — 基于 slog 的默认实现
- `Level` / `KeyValue` — 日志级别和结构化数据

---

## Level — 日志级别

```go
type Level int8

const (
    DebugLevel  Level = iota // 调试级别
    InfoLevel                // 信息级别
    WarnLevel                // 警告级别
    ErrorLevel               // 错误级别
    DPanicLevel              // 致命错误级别（开发环境 panic）
    PanicLevel               // panic 级别
    FatalLevel               // 致命级别（程序退出）
)
```

字符串表示：`"debug"`、`"info"`、`"warn"`、`"error"`、`"dpanic"`、`"panic"`、`"fatal"`。

### ToLevel

将字符串转换为日志级别：

```go
level := log.ToLevel("info") // log.InfoLevel
level := log.ToLevel("warn") // log.WarnLevel
```

---

## KeyValue — 结构化键值对

```go
type KeyValue struct {
    Key   string
    Value any
}
```

用于结构化日志记录：

```go
log.KeyValue{Key: "user_id", Value: 123}
log.KeyValue{Key: "duration", Value: time.Second}
```

---

## Logger 接口

```go
type Logger interface {
    Debug(ctx context.Context, msg string, keys ...KeyValue)
    Info(ctx context.Context, msg string, keys ...KeyValue)
    Warn(ctx context.Context, msg string, keys ...KeyValue)
    Error(ctx context.Context, msg string, keys ...KeyValue)
    DPanic(ctx context.Context, msg string, keys ...KeyValue)
    Panic(ctx context.Context, msg string, keys ...KeyValue)
    Fatal(ctx context.Context, msg string, keys ...KeyValue)
    Sync() error
    With(ctx context.Context, keys ...KeyValue) Logger
}
```

### 方法说明

| 方法 | 说明 |
|------|------|
| `Debug` | 调试日志 |
| `Info` | 信息日志 |
| `Warn` | 警告日志 |
| `Error` | 错误日志 |
| `DPanic` | 致命错误日志，开发环境触发 panic |
| `Panic` | 记录日志并 panic |
| `Fatal` | 记录日志（注意：不主动调用 `os.Exit(1)`） |
| `Sync` | 同步日志缓冲区 |
| `With` | 返回带有额外固定字段的新 Logger |

---

## 扩展接口

### LoggerWithLevel

```go
type LoggerWithLevel interface {
    Logger
    Log(ctx context.Context, level Level, msg string, keys ...KeyValue)
}
```

支持在运行时使用任意级别记录日志。

### LoggerWithName

```go
type LoggerWithName interface {
    Logger
    WithName(name string) Logger
}
```

支持为 Logger 设置名称（如模块名），便于日志分类。

### LoggerWithCaller

```go
type LoggerWithCaller interface {
    Logger
    WithCaller(skip int) Logger
}
```

支持记录调用者源码位置信息。

### LoggerWithTimeout

```go
type LoggerWithTimeout interface {
    Logger
    WithTimeout(d time.Duration) Logger
}
```

支持超时自动刷盘。

---

## Build() — 日志记录器构建

```go
func Build(opts ...LoggerOption) Logger
```

```go
type LoggerOption func(*loggerConfig)

func WithLogger(logger Logger) LoggerOption
```

---

## SlogLogger — 基于 slog 的实现

### 创建

```go
// 默认配置（JSON 格式，Info 级别，标准输出）
logger := log.NewSlogLogger()

// 自定义配置
logger := log.NewSlogLogger(
    log.WithLevel(log.DebugLevel),
    log.WithFormat("text"),
    log.WithTimeFormat("2006-01-02 15:04:05"),
    log.WithAddSource(true),
    log.WithOutput(os.Stderr),
    log.WithOutputPath("/var/log/app.log"),
)
```

### Option

| 选项 | 说明 |
|------|------|
| `WithLevel(level Level)` | 设置日志级别（默认 Info） |
| `WithFormat(format string)` | 输出格式：`"json"`（默认）或 `"text"` |
| `WithTimeFormat(timeFormat string)` | 时间格式（默认 `"2006-01-02 15:04:05"`） |
| `WithAddSource(addSource bool)` | 是否添加源码位置（默认 false） |
| `WithOutput(w io.Writer)` | 设置输出 Writer（默认 os.Stdout） |
| `WithOutputPath(path string)` | 设置日志文件输出路径 |

### Close

关闭日志文件句柄（仅在设置了文件输出时需要调用）：

```go
if closer, ok := logger.(*log.SlogLogger); ok {
    defer closer.Close()
}
```

---

## 使用示例

```go
logger := log.NewSlogLogger(
    log.WithLevel(log.DebugLevel),
    log.WithFormat("json"),
)

ctx := context.Background()

logger.Info(ctx, "服务启动", log.KeyValue{Key: "port", Value: 8080})
logger.Debug(ctx, "SQL 查询", log.KeyValue{Key: "sql", Value: "SELECT * FROM users"})
logger.Error(ctx, "请求失败",
    log.KeyValue{Key: "path", Value: "/api/users"},
    log.KeyValue{Key: "status", Value: 500},
)

// 使用 With 添加固定字段
logger2 := logger.With(ctx, log.KeyValue{Key: "service", Value: "user-svc"})
logger2.Info(ctx, "用户注册") // 自动携带 service=user-svc

// 使用 Build 构建
logger := log.Build(log.WithLogger(log.NewSlogLogger(
    log.WithFormat("text"),
)))
```

### 与依赖注入集成

```go
container.Register("logger",
    core.Bean(log.NewSlogLogger(
        log.WithLevel(log.InfoLevel),
        log.WithFormat("json"),
    )),
    core.Singleton(),
)

type UserService struct {
    Logger log.Logger `inject:"logger"`
}
```
