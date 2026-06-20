// Package log 提供 slog 日志适配器实现。
//
// SlogLogger 实现 Logger 接口，基于 Go 标准库 log/slog 包，
// 支持 JSON 和 text 两种输出格式，以及多种日志级别。
package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
)

// Option 定义 slog 日志配置选项
type Option func(*SlogLogger)

// WithLevel 设置日志级别
func WithLevel(level Level) Option {
	return func(l *SlogLogger) {
		l.level = level
	}
}

// WithFormat 设置输出格式
func WithFormat(format string) Option {
	return func(l *SlogLogger) {
		l.format = format
	}
}

// WithTimeFormat 设置时间格式
func WithTimeFormat(timeFormat string) Option {
	return func(l *SlogLogger) {
		l.timeFormat = timeFormat
	}
}

// WithAddSource 设置是否添加源码位置
func WithAddSource(addSource bool) Option {
	return func(l *SlogLogger) {
		l.addSource = addSource
	}
}

// WithOutput 设置输出 writer
func WithOutput(output io.Writer) Option {
	return func(l *SlogLogger) {
		l.output = output
	}
}

// WithOutputPath 设置日志文件输出路径
func WithOutputPath(path string) Option {
	return func(l *SlogLogger) {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Printf("failed to open log file %s: %v, using stdout\n", path, err)
			return
		}
		l.output = f
		l.file = f
	}
}

// SlogLogger 是 slog 日志适配器,实现 Logger 接口
type SlogLogger struct {
	logger     *slog.Logger
	level      Level
	format     string
	timeFormat string
	addSource  bool
	output     io.Writer
	file       *os.File
}

// NewSlogLogger 创建 slog 日志适配器
func NewSlogLogger(opts ...Option) *SlogLogger {
	l := &SlogLogger{
		level:      InfoLevel,
		format:     "json",
		timeFormat: "2006-01-02 15:04:05",
		addSource:  false,
		output:     os.Stdout,
	}

	for _, opt := range opts {
		opt(l)
	}

	var handler slog.Handler
	handlerOptions := &slog.HandlerOptions{
		Level:     l.toSlogLevel(l.level),
		AddSource: l.addSource,
	}

	if l.format == "text" {
		handler = slog.NewTextHandler(l.output, handlerOptions)
	} else {
		handler = slog.NewJSONHandler(l.output, handlerOptions)
	}

	l.logger = slog.New(handler)
	return l
}

// 自定义 slog 级别，用于 panic/fatal（高于 slog.LevelError=8）
const (
	slogLevelDPanic = slog.LevelError + 2
	slogLevelPanic  = slog.LevelError + 4
	slogLevelFatal  = slog.LevelError + 6
)

// toSlogLevel 将日志级别转换为 slog 级别
func (l *SlogLogger) toSlogLevel(level Level) slog.Level {
	switch level {
	case DebugLevel:
		return slog.LevelDebug
	case InfoLevel:
		return slog.LevelInfo
	case WarnLevel:
		return slog.LevelWarn
	case ErrorLevel:
		return slog.LevelError
	case DPanicLevel:
		return slogLevelDPanic
	case PanicLevel:
		return slogLevelPanic
	case FatalLevel:
		return slogLevelFatal
	default:
		return slog.LevelInfo
	}
}

// log 记录日志
func (l *SlogLogger) log(ctx context.Context, level Level, msg string, keys []KeyValue) {
	slogLevel := l.toSlogLevel(level)
	var attrs []any

	for _, kv := range keys {
		attrs = append(attrs, kv.Key, kv.Value)
	}
	l.logger.Log(ctx, slogLevel, msg, attrs...)
}

// Debug 记录调试日志
func (l *SlogLogger) Debug(ctx context.Context, msg string, keys ...KeyValue) {
	l.log(ctx, DebugLevel, msg, keys)
}

// Info 记录信息日志
func (l *SlogLogger) Info(ctx context.Context, msg string, keys ...KeyValue) {
	l.log(ctx, InfoLevel, msg, keys)
}

// Warn 记录警告日志
func (l *SlogLogger) Warn(ctx context.Context, msg string, keys ...KeyValue) {
	l.log(ctx, WarnLevel, msg, keys)
}

// Error 记录错误日志
func (l *SlogLogger) Error(ctx context.Context, msg string, keys ...KeyValue) {
	l.log(ctx, ErrorLevel, msg, keys)
}

// DPanic 记录致命错误日志并 panic
func (l *SlogLogger) DPanic(ctx context.Context, msg string, keys ...KeyValue) {
	l.log(ctx, DPanicLevel, msg, keys)
	panic(msg)
}

// Panic 记录日志并 panic
func (l *SlogLogger) Panic(ctx context.Context, msg string, keys ...KeyValue) {
	l.log(ctx, PanicLevel, msg, keys)
	panic(msg)
}

// Fatal 记录致命级别日志
//
// 注意：与标准 log.Fatal 不同，此方法仅记录日志，不会调用 os.Exit(1)。
// 如需退出程序，调用方需自行处理。
func (l *SlogLogger) Fatal(ctx context.Context, msg string, keys ...KeyValue) {
	l.log(ctx, FatalLevel, msg, keys)
}

// Sync 同步日志缓冲区
func (l *SlogLogger) Sync() error {
	return nil
}

// Close 关闭日志文件句柄
func (l *SlogLogger) Close() error {
	if l.file != nil && l.file != os.Stdout {
		return l.file.Close()
	}
	return nil
}

// With 返回带有额外字段的日志记录器
func (l *SlogLogger) With(ctx context.Context, keys ...KeyValue) Logger {
	var attrs []any
	for _, kv := range keys {
		attrs = append(attrs, kv.Key, kv.Value)
	}
	return &SlogLogger{
		logger:     l.logger.With(attrs...),
		level:      l.level,
		format:     l.format,
		timeFormat: l.timeFormat,
		addSource:  l.addSource,
		output:     l.output,
	}
}

var _ Logger = (*SlogLogger)(nil)
