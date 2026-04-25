# Log 模块

统一日志接口,支持多种日志库(slog/zap/zerolog)。

## 概述

log模块定义统一的`Logger`接口,提供通用日志操作:

- Debug/Info/Warn/Error 不同级别日志
- DPanic/Panic/Fatal 致命错误处理
- Sync 同步缓冲区
- With 添加上下文字段

支持多种实现:

- slog: Go标准库日志
- zap: Uber的zap日志库
- zero: rs/zerolog日志库

## 使用方法

### 基本用法

```go
import "github.com/xudefa/go-boot/log"

ctx := context.Background()

// 记录不同级别日志
log.Debug(ctx, "debug message", log.KeyValue{Key: "key", Value: "value"})
log.Info(ctx, "info message", log.KeyValue{Key: "key", Value: "value"})
log.Warn(ctx, "warning message", log.KeyValue{Key: "key", Value: "value"})
log.Error(ctx, "error message", log.KeyValue{Key: "error", Value: err})

// 同步缓冲区
log.Sync()
```

### 设置日志库

```go
// 使用 slog
cfg := &slog.Config{
    Level:  "info",
    Format: "json",
}
logger := slog.NewSlogLogger(cfg)
log.SetDefault(logger)

// 使用 zap
cfg := &zap.ZapConfig{
    Level:  "info",
    Format: "json",
}
logger := zap.NewZapAdapter(cfg)
log.SetDefault(logger)

// 使用 zerolog
cfg := &zero.ZerologConfig{
    Level:  "info",
    Format: "json",
}
logger := zero.NewZerologAdapter(cfg)
log.SetDefault(logger)
```

### 使用Builder

```go
// 使用 zap Builder
logger := zap.NewZapLoggerBuilder().
    Level("info").
    Format("json").
    AddCaller(true).
    Output("app.log").
    Build()

log.SetDefault(logger)
```

### 创建带上下文的日志

```go
ctx := context.Background()

// 添加额外字段
logger := log.With(ctx, log.KeyValue{Key: "user_id", Value: 123})
logger.Info(ctx, "user logged in")
```

## 配置说明

### slog.Config

| 字段         | 类型     | 说明                          | 默认值                 |
|------------|--------|-----------------------------|---------------------|
| Level      | string | 日志级别: debug/info/warn/error | info                |
| Format     | string | 输出格式: json/text             | json                |
| TimeFormat | string | 时间格式                        | 2006-01-02 15:04:05 |
| AddSource  | bool   | 是否添加源码位置                    | false               |
| Output     | string | 输出文件路径                      | stdout              |

### zap.ZapConfig

| 字段         | 类型     | 说明                          | 默认值                 |
|------------|--------|-----------------------------|---------------------|
| Level      | string | 日志级别: debug/info/warn/error | info                |
| Format     | string | 输出格式: json/console          | json                |
| TimeFormat | string | 时间格式                        | 2006-01-02 15:04:05 |
| AddCaller  | bool   | 是否添加调用者信息                   | false               |
| CallerSkip | int    | 调用者跳过帧数                     | 1                   |
| Output     | string | 输出文件路径                      | stdout              |

### zero.ZerologConfig

| 字段         | 类型     | 说明                          | 默认值                 |
|------------|--------|-----------------------------|---------------------|
| Level      | string | 日志级别: debug/info/warn/error | info                |
| Format     | string | 输出格式: json/console/text     | json                |
| TimeFormat | string | 时间格式                        | 2006-01-02 15:04:05 |
| Output     | string | 输出文件路径                      | stdout              |

## 接口定义

### Logger 接口

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

### 扩展接口

- `LoggerWithLevel`: 支持自定义级别日���
- `LoggerWithName`: 支持日志命名
- `LoggerWithCaller`: 支持调用者信息
- `LoggerWithTimeout`: 支持超时日志