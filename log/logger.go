// Package log 提供统一日志接口,支持多种日志库(slog/zap/zerolog).
package log

import (
	"context"
	"fmt"
	"time"
)

// Level 定义日志级别
type Level int8

const (
	DebugLevel  Level = iota // 调试级别
	InfoLevel                // 信息级别
	WarnLevel                // 警告级别
	ErrorLevel               // 错误级别
	DPanicLevel              // 致命错误级别(开发环境 panic)
	PanicLevel               // panic 级别
	FatalLevel               // 致命级别(程序退出)
)

// String 返回日志级别的字符串表示
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case DPanicLevel:
		return "dpanic"
	case PanicLevel:
		return "panic"
	case FatalLevel:
		return "fatal"
	default:
		return "unknown"
	}
}

// KeyValue 定义日志键值对
type KeyValue struct {
	Key   string
	Value any
}

// Logger 是日志记录器接口,所有日志库都需实现此接口
type Logger interface {
	Debug(ctx context.Context, msg string, keys ...KeyValue)  // 记录调试日志
	Info(ctx context.Context, msg string, keys ...KeyValue)   // 记录信息日志
	Warn(ctx context.Context, msg string, keys ...KeyValue)   // 记录警告日志
	Error(ctx context.Context, msg string, keys ...KeyValue)  // 记录错误日志
	DPanic(ctx context.Context, msg string, keys ...KeyValue) // 记录致命错误日志并 panic
	Panic(ctx context.Context, msg string, keys ...KeyValue)  // 记录日志并 panic
	Fatal(ctx context.Context, msg string, keys ...KeyValue)  // 记录日志并退出程序
	Sync() error                                              // 同步日志缓冲区
	With(ctx context.Context, keys ...KeyValue) Logger        // 返回带有额外字段的日志记录器
}

// LoggerWithLevel 支持自定义日志级别
type LoggerWithLevel interface {
	Logger
	Log(ctx context.Context, level Level, msg string, keys ...KeyValue)
}

// LoggerWithName 支持日志命名
type LoggerWithName interface {
	Logger
	WithName(name string) Logger
}

// LoggerWithCaller 支持调用者信息
type LoggerWithCaller interface {
	Logger
	WithCaller(skip int) Logger
}

// LoggerWithTimeout 支持超时日志
type LoggerWithTimeout interface {
	Logger
	WithTimeout(d time.Duration) Logger
}

// Debug 使用默认日志记录器记录调试日志
func Debug(ctx context.Context, msg string, keys ...KeyValue) {
	defaultLogger.Debug(ctx, msg, keys...)
}

// Info 使用默认日志记录器记录信息日志
func Info(ctx context.Context, msg string, keys ...KeyValue) {
	defaultLogger.Info(ctx, msg, keys...)
}

// Warn 使用默认日志记录器记录警告日志
func Warn(ctx context.Context, msg string, keys ...KeyValue) {
	defaultLogger.Warn(ctx, msg, keys...)
}

// Error 使用默认日志记录器记录错误日志
func Error(ctx context.Context, msg string, keys ...KeyValue) {
	defaultLogger.Error(ctx, msg, keys...)
}

// DPanic 使用默认日��记录器记录致命错误日志并 panic
func DPanic(ctx context.Context, msg string, keys ...KeyValue) {
	defaultLogger.DPanic(ctx, msg, keys...)
}

// Panic 使用默认日志记录器记录日志并 panic
func Panic(ctx context.Context, msg string, keys ...KeyValue) {
	defaultLogger.Panic(ctx, msg, keys...)
}

// Fatal 使用默认日志记录器记录日志并退出程序
func Fatal(ctx context.Context, msg string, keys ...KeyValue) {
	defaultLogger.Fatal(ctx, msg, keys...)
}

// Sync 同步默认日志记录器的缓冲区
func Sync() error {
	return defaultLogger.Sync()
}

// With 使用默认日志记录器返回带有额外字段的日志记录器
func With(ctx context.Context, keys ...KeyValue) Logger {
	return defaultLogger.With(ctx, keys...)
}

// defaultLogger 是默认日志记录器,初始为 nopLogger(不输出日志)
var defaultLogger Logger = &nopLogger{}

// SetDefault 设置默认日志记录器
func SetDefault(logger Logger) {
	if logger == nil {
		defaultLogger = &nopLogger{}
		return
	}
	defaultLogger = logger
}

// DefaultLogger 返回当前默认日志记录器
func DefaultLogger() Logger {
	return defaultLogger
}

// nopLogger 是空日志记录器,不输出任何日志
type nopLogger struct{}

func (n *nopLogger) Debug(ctx context.Context, msg string, keys ...KeyValue)  {}
func (n *nopLogger) Info(ctx context.Context, msg string, keys ...KeyValue)   {}
func (n *nopLogger) Warn(ctx context.Context, msg string, keys ...KeyValue)   {}
func (n *nopLogger) Error(ctx context.Context, msg string, keys ...KeyValue)  {}
func (n *nopLogger) DPanic(ctx context.Context, msg string, keys ...KeyValue) {}
func (n *nopLogger) Panic(ctx context.Context, msg string, keys ...KeyValue)  {}
func (n *nopLogger) Fatal(ctx context.Context, msg string, keys ...KeyValue)  {}
func (n *nopLogger) Sync() error                                              { return nil }
func (n *nopLogger) With(ctx context.Context, keys ...KeyValue) Logger        { return n }

// Builder 是日志记录器构建器
type Builder struct {
	Logger
}

// NewBuilder 创建新的日志记录器构建器
func NewBuilder() *Builder {
	return &Builder{&nopLogger{}}
}

// Build 构建日志记录器
func (b *Builder) Build() Logger {
	if b.Logger == nil {
		return &nopLogger{}
	}
	return b.Logger
}

// LoggerOption 定义日志记录器配置选项
type LoggerOption func(*Builder)

// WithLogger 设置日志记录器
func WithLogger(logger Logger) LoggerOption {
	return func(b *Builder) {
		b.Logger = logger
	}
}

// Build 使用选项构建日志记录器
func Build(opts ...LoggerOption) Logger {
	b := NewBuilder()
	for _, opt := range opts {
		opt(b)
	}
	return b.Build()
}

// Must 如果 err 不为 nil 则 panic
func Must(err error) {
	if err != nil {
		panic(fmt.Sprintf("log init failed: %v", err))
	}
}

// ToLevel 将字符串转换为日志级别
func ToLevel(level string) Level {
	switch level {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "dpanic":
		return DPanicLevel
	case "panic":
		return PanicLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel
	}
}
