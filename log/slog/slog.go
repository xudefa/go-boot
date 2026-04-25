package slog

import (
	"context"
	"github.com/xudefa/go-boot/log"
	"io"
	"log/slog"
	"os"
)

// Config 定义 slog 日志配置
type Config struct {
	Level      string `mapstructure:"level" json:"level,omitempty"`             // 日志级别: debug/info/warn/error
	Format     string `mapstructure:"format" json:"format,omitempty"`           // 输出格式: json/text
	TimeFormat string `mapstructure:"time-format" json:"time_format,omitempty"` // 时间格式
	AddSource  bool   `mapstructure:"add-source" json:"add_source"`             // 是否添加源码位置
	Output     string `mapstructure:"output" json:"output,omitempty"`           // 输出文件路径
}

// LevelValue 返回 slog.Level 类型的日志级别
func (c *Config) LevelValue() slog.Level {
	switch c.Level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Option 定义 slog 日志配置选项
type Option func(*SlogLogger)

// WithLevel 设置日志级别
func WithLevel(level log.Level) Option {
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

// SlogLogger 是 slog 日志适配器,实现 Logger 接口
type SlogLogger struct {
	logger     *slog.Logger
	level      log.Level
	format     string
	timeFormat string
	addSource  bool
	output     io.Writer
}

// NewSlogLogger 创建 slog 日志适配器
func NewSlogLogger(cfg *Config, opts ...Option) *SlogLogger {
	l := &SlogLogger{
		level:      log.InfoLevel,
		format:     "json",
		timeFormat: "2006-01-02 15:04:05",
		addSource:  false,
		output:     os.Stdout,
	}
	if cfg != nil {
		l.level = toLevel(cfg.Level)
		if cfg.Format != "" {
			l.format = cfg.Format
		}
		if cfg.TimeFormat != "" {
			l.timeFormat = cfg.TimeFormat
		}
		l.addSource = cfg.AddSource
		if cfg.Output != "" {
			f, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				f = os.Stdout
			}
			l.output = f
		}
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

// toLevel 将字符串转换为日志级别
func toLevel(level string) log.Level {
	return log.ToLevel(level)
}

// toSlogLevel 将日志级别转换为 slog 级别
func (l *SlogLogger) toSlogLevel(level log.Level) slog.Level {
	switch level {
	case log.DebugLevel:
		return slog.LevelDebug
	case log.InfoLevel:
		return slog.LevelInfo
	case log.WarnLevel:
		return slog.LevelWarn
	case log.ErrorLevel:
		return slog.LevelError
	case log.DPanicLevel:
		return slog.LevelDebug
	case log.PanicLevel:
		return slog.LevelDebug
	case log.FatalLevel:
		return slog.LevelDebug
	default:
		return slog.LevelInfo
	}
}

// toSlogAttr 将键值对转换为 slog.Attr
func (l *SlogLogger) toSlogAttr(key string, value any) slog.Attr {
	return slog.Any(key, value)
}

// log 记录日志
func (l *SlogLogger) log(ctx context.Context, level log.Level, msg string, keys []log.KeyValue) {
	slogLevel := l.toSlogLevel(level)
	var attrs []any
	for _, kv := range keys {
		attrs = append(attrs, kv.Key, kv.Value)
	}
	l.logger.Log(ctx, slogLevel, msg, attrs...)
}

// Debug 记录调试日志
func (l *SlogLogger) Debug(ctx context.Context, msg string, keys ...log.KeyValue) {
	l.log(ctx, log.DebugLevel, msg, keys)
}

// Info 记录信息日志
func (l *SlogLogger) Info(ctx context.Context, msg string, keys ...log.KeyValue) {
	l.log(ctx, log.InfoLevel, msg, keys)
}

// Warn 记录警告日志
func (l *SlogLogger) Warn(ctx context.Context, msg string, keys ...log.KeyValue) {
	l.log(ctx, log.WarnLevel, msg, keys)
}

// Error 记录错误日志
func (l *SlogLogger) Error(ctx context.Context, msg string, keys ...log.KeyValue) {
	l.log(ctx, log.ErrorLevel, msg, keys)
}

// DPanic 记录致命错误日志并 panic
func (l *SlogLogger) DPanic(ctx context.Context, msg string, keys ...log.KeyValue) {
	l.log(ctx, log.DPanicLevel, msg, keys)
}

// Panic 记录日志并 panic
func (l *SlogLogger) Panic(ctx context.Context, msg string, keys ...log.KeyValue) {
	l.log(ctx, log.PanicLevel, msg, keys)
}

// Fatal 记录日志并退出程序
func (l *SlogLogger) Fatal(ctx context.Context, msg string, keys ...log.KeyValue) {
	l.log(ctx, log.FatalLevel, msg, keys)
}

// Sync 同步日志缓冲区
func (l *SlogLogger) Sync() error {
	return nil
}

// With 返回带有额外字段的日志记录器
func (l *SlogLogger) With(ctx context.Context, keys ...log.KeyValue) log.Logger {
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

var _ log.Logger = (*SlogLogger)(nil)

type slogTextWriter struct {
	timeFormat string
}

func (w *slogTextWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

type slogJSONWriter struct {
	timeFormat string
}

func (w *slogJSONWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
