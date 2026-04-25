package zero

import (
	"context"
	"fmt"
	"github.com/xudefa/go-boot/log"
	"os"

	"github.com/rs/zerolog"
)

// ZerologConfig 定义 zerolog 日志配置
type ZerologConfig struct {
	Level      string `mapstructure:"level" json:"level,omitempty"`             // 日志级别: debug/info/warn/error
	Format     string `mapstructure:"format" json:"format,omitempty"`           // 输出格式: json/console/text
	TimeFormat string `mapstructure:"time-format" json:"time_format,omitempty"` // 时间格式
	Output     string `mapstructure:"output" json:"output,omitempty"`           // 输出文件路径
}

// ZerologOption 定义 zerolog 日志配置选项
type ZerologOption func(*ZerologAdapter)

// WithZerologLevel 设置日志级别
func WithZerologLevel(level log.Level) ZerologOption {
	return func(a *ZerologAdapter) {
		a.level = level
	}
}

// WithZerologFormat 设置输出格式
func WithZerologFormat(format string) ZerologOption {
	return func(a *ZerologAdapter) {
		a.format = format
	}
}

// WithZerologTimeFormat 设置时间格式
func WithZerologTimeFormat(timeFormat string) ZerologOption {
	return func(a *ZerologAdapter) {
		a.timeFormat = timeFormat
	}
}

// WithZerologOutput 设置输出文件
func WithZerologOutput(output *os.File) ZerologOption {
	return func(a *ZerologAdapter) {
		a.output = output
	}
}

// ZerologAdapter 是 zerolog 日志适配器,实现 Logger 接口
type ZerologAdapter struct {
	logger     zerolog.Logger
	level      log.Level
	format     string
	timeFormat string
	output     *os.File
}

// NewZerologAdapter 创建 zerolog 日志适配器
func NewZerologAdapter(cfg *ZerologConfig, opts ...ZerologOption) *ZerologAdapter {
	a := &ZerologAdapter{
		level:      log.InfoLevel,
		format:     "json",
		timeFormat: "2006-01-02 15:04:05",
		output:     os.Stdout,
	}
	if cfg != nil {
		a.level = log.ToLevel(cfg.Level)
		if cfg.Format != "" {
			a.format = cfg.Format
		}
		if cfg.TimeFormat != "" {
			a.timeFormat = cfg.TimeFormat
		}
		if cfg.Output != "" {
			f, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err == nil {
				a.output = f
			}
		}
	}
	for _, opt := range opts {
		opt(a)
	}

	var logger zerolog.Logger
	if a.format == "console" || a.format == "text" {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: a.output}).With().Timestamp().Logger()
	} else {
		logger = zerolog.New(a.output).With().Timestamp().Logger()
	}
	logger = logger.Level(a.toZerologLevel(a.level))
	a.logger = logger
	return a
}

// toZerologLevel 将日志级别转换为 zerolog 级别
func (a *ZerologAdapter) toZerologLevel(level log.Level) zerolog.Level {
	switch level {
	case log.DebugLevel:
		return zerolog.DebugLevel
	case log.InfoLevel:
		return zerolog.InfoLevel
	case log.WarnLevel:
		return zerolog.WarnLevel
	case log.ErrorLevel:
		return zerolog.ErrorLevel
	case log.DPanicLevel:
		return zerolog.DebugLevel
	case log.PanicLevel:
		return zerolog.PanicLevel
	case log.FatalLevel:
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// toZerologFields 将 KeyValue 转换为 map
func (a *ZerologAdapter) toZerologFields(keys []log.KeyValue) map[string]any {
	fields := make(map[string]any)
	for _, kv := range keys {
		fields[kv.Key] = kv.Value
	}
	return fields
}

// log 记录日志
func (a *ZerologAdapter) log(ctx context.Context, level log.Level, msg string, keys []log.KeyValue) {
	fields := a.toZerologFields(keys)
	switch level {
	case log.DebugLevel:
		a.logger.Debug().Fields(fields).Msg(msg)
	case log.InfoLevel:
		a.logger.Info().Fields(fields).Msg(msg)
	case log.WarnLevel:
		a.logger.Warn().Fields(fields).Msg(msg)
	case log.ErrorLevel:
		a.logger.Error().Fields(fields).Msg(msg)
	case log.DPanicLevel:
		a.logger.Debug().Fields(fields).Msg(msg)
	case log.PanicLevel:
		a.logger.Panic().Fields(fields).Msg(msg)
	case log.FatalLevel:
		a.logger.Fatal().Fields(fields).Msg(msg)
	default:
		a.logger.Info().Fields(fields).Msg(msg)
	}
}

// Debug 记录调试日志
func (a *ZerologAdapter) Debug(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.DebugLevel, msg, keys)
}

// Info 记录信息日志
func (a *ZerologAdapter) Info(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.InfoLevel, msg, keys)
}

// Warn 记录警告日志
func (a *ZerologAdapter) Warn(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.WarnLevel, msg, keys)
}

// Error 记录错误日志
func (a *ZerologAdapter) Error(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.ErrorLevel, msg, keys)
}

// DPanic 记录致命错误日志并 panic
func (a *ZerologAdapter) DPanic(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.DPanicLevel, msg, keys)
}

// Panic 记录日志并 panic
func (a *ZerologAdapter) Panic(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.PanicLevel, msg, keys)
}

// Fatal 记录日志并退出程序
func (a *ZerologAdapter) Fatal(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.FatalLevel, msg, keys)
}

// Sync 同步日志缓冲区
func (a *ZerologAdapter) Sync() error {
	if a.output != nil {
		return a.output.Sync()
	}
	return nil
}

// With 返回带有额外字段的日志记录器
func (a *ZerologAdapter) With(ctx context.Context, keys ...log.KeyValue) log.Logger {
	fields := a.toZerologFields(keys)
	return &ZerologAdapter{
		logger:     a.logger.With().Fields(fields).Logger(),
		level:      a.level,
		format:     a.format,
		timeFormat: a.timeFormat,
		output:     a.output,
	}
}

var _ log.Logger = (*ZerologAdapter)(nil)

type zerologTextWriter struct{}

func (w *zerologTextWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// MustZerolog 如果 err 不为 nil 则 panic
func MustZerolog(err error) {
	if err != nil {
		panic(fmt.Sprintf("log zerolog init failed: %v", err))
	}
}

// ZerologLoggerBuilder 是 zerolog 日志构建器
type ZerologLoggerBuilder struct {
	cfg  *ZerologConfig
	opts []ZerologOption
}

// NewZerologLoggerBuilder 创建新的日志构建器
func NewZerologLoggerBuilder() *ZerologLoggerBuilder {
	return &ZerologLoggerBuilder{}
}

// Level 设置日志级别
func (b *ZerologLoggerBuilder) Level(level string) *ZerologLoggerBuilder {
	b.cfg.Level = level
	return b
}

// Format 设置输出格式
func (b *ZerologLoggerBuilder) Format(format string) *ZerologLoggerBuilder {
	b.cfg.Format = format
	return b
}

// TimeFormat 设置时间格式
func (b *ZerologLoggerBuilder) TimeFormat(timeFormat string) *ZerologLoggerBuilder {
	b.cfg.TimeFormat = timeFormat
	return b
}

// Output 设置输出文件
func (b *ZerologLoggerBuilder) Output(output string) *ZerologLoggerBuilder {
	b.cfg.Output = output
	return b
}

// Build 构建日志记录器
func (b *ZerologLoggerBuilder) Build(opts ...ZerologOption) log.Logger {
	if b.cfg == nil {
		b.cfg = &ZerologConfig{}
	}
	for _, opt := range b.opts {
		opts = append(opts, opt)
	}
	return NewZerologAdapter(b.cfg, opts...)
}
