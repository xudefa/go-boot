package zap

import (
	"context"
	"fmt"
	"github.com/xudefa/go-boot/log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapConfig 定义 zap 日志配置
type ZapConfig struct {
	Level       string                                                                        `mapstructure:"level" json:"level,omitempty"`             // 日志级别: debug/info/warn/error
	Format      string                                                                        `mapstructure:"format" json:"format,omitempty"`           // 输出格式: json/console
	TimeFormat  string                                                                        `mapstructure:"time-format" json:"time_format,omitempty"` // 时间格式
	AddCaller   bool                                                                          `mapstructure:"add-caller" json:"add_caller"`             // 是否添加调用者信息
	CallerSkip  int                                                                           `mapstructure:"caller-skip" json:"caller_skip"`           // 调用者跳过帧数
	Output      string                                                                        `mapstructure:"output" json:"output,omitempty"`           // 输出文件路径
	EncodeLevel func(zapcore.Level, zapcore.TimeEncoder, zapcore.PrimitiveArrayEncoder) error // 自定义级别编码器
}

// ZapOption 定义 zap 日志配置选项
type ZapOption func(*ZapAdapter)

// WithZapEncoder 设置自定义编码器
func WithZapEncoder(encoder zapcore.Encoder) ZapOption {
	return func(a *ZapAdapter) {
		a.encoder = encoder
	}
}

// WithZapLevel 设置日志级别
func WithZapLevel(level log.Level) ZapOption {
	return func(a *ZapAdapter) {
		a.level = level
	}
}

// WithZapAddCaller 设置是否添加调用者信息
func WithZapAddCaller(addCaller bool) ZapOption {
	return func(a *ZapAdapter) {
		a.addCaller = addCaller
	}
}

// WithZapCallerSkip 设置调用者跳过帧数
func WithZapCallerSkip(callerSkip int) ZapOption {
	return func(a *ZapAdapter) {
		a.callerSkip = callerSkip
	}
}

// WithZapOutput 设置输出 writer
func WithZapOutput(output zapcore.WriteSyncer) ZapOption {
	return func(a *ZapAdapter) {
		a.output = output
	}
}

// WithZapTimeFormat 设置时间格式
func WithZapTimeFormat(timeFormat string) ZapOption {
	return func(a *ZapAdapter) {
		a.timeFormat = timeFormat
	}
}

// ZapAdapter 是 zap 日志适配器,实现 Logger 接口
type ZapAdapter struct {
	logger     *zap.SugaredLogger
	level      log.Level
	format     string
	timeFormat string
	addCaller  bool
	callerSkip int
	encoder    zapcore.Encoder
	output     zapcore.WriteSyncer
}

// NewZapAdapter 创建 zap 日志适配器
func NewZapAdapter(cfg *ZapConfig, opts ...ZapOption) *ZapAdapter {
	a := &ZapAdapter{
		level:      log.InfoLevel,
		format:     "json",
		timeFormat: "2006-01-02 15:04:05",
		addCaller:  false,
		callerSkip: 1,
		output:     zapcore.AddSync(os.Stdout),
	}
	if cfg != nil {
		a.level = log.ToLevel(cfg.Level)
		if cfg.Format != "" {
			a.format = cfg.Format
		}
		if cfg.TimeFormat != "" {
			a.timeFormat = cfg.TimeFormat
		}
		a.addCaller = cfg.AddCaller
		a.callerSkip = cfg.CallerSkip
		if cfg.Output != "" {
			f, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err == nil {
				a.output = zapcore.AddSync(f)
			}
		} else {
			a.output = zapcore.AddSync(os.Stdout)
		}
	}
	for _, opt := range opts {
		opt(a)
	}

	var encoder zapcore.Encoder
	encodeConfig := &zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout(a.timeFormat),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	if a.format == "text" {
		encoder = zapcore.NewConsoleEncoder(*encodeConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(*encodeConfig)
	}
	if a.encoder != nil {
		encoder = a.encoder
	}

	level := a.toZapLevel(a.level)
	core := zapcore.NewCore(
		encoder,
		a.output,
		level,
	)
	l := zap.New(core, zap.AddCaller(), zap.Development())
	a.logger = l.Sugar()

	return a
}

// toZapLevel 将日志级别转换为 zap 级别
func (a *ZapAdapter) toZapLevel(level log.Level) zapcore.Level {
	switch level {
	case log.DebugLevel:
		return zapcore.DebugLevel
	case log.InfoLevel:
		return zapcore.InfoLevel
	case log.WarnLevel:
		return zapcore.WarnLevel
	case log.ErrorLevel:
		return zapcore.ErrorLevel
	case log.DPanicLevel:
		return zapcore.DebugLevel
	case log.PanicLevel:
		return zapcore.FatalLevel
	case log.FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// toZapFields 将 KeyValue 转换为 zap.Field
func (a *ZapAdapter) toZapFields(keys []log.KeyValue) []any {
	var fields []any
	for _, kv := range keys {
		fields = append(fields, zap.Any(kv.Key, kv.Value))
	}
	return fields
}

// log 记录日志
func (a *ZapAdapter) log(ctx context.Context, level log.Level, msg string, keys []log.KeyValue) {
	fields := a.toZapFields(keys)
	switch level {
	case log.DebugLevel:
		a.logger.Debugw(msg, fields...)
	case log.InfoLevel:
		a.logger.Infow(msg, fields...)
	case log.WarnLevel:
		a.logger.Warnw(msg, fields...)
	case log.ErrorLevel:
		a.logger.Errorw(msg, fields...)
	case log.DPanicLevel:
		a.logger.DPanicw(msg, fields...)
	case log.PanicLevel:
		a.logger.Panicw(msg, fields...)
	case log.FatalLevel:
		a.logger.Fatalw(msg, fields...)
	}
}

// Debug 记录调试日志
func (a *ZapAdapter) Debug(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.DebugLevel, msg, keys)
}

// Info 记录信息日志
func (a *ZapAdapter) Info(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.InfoLevel, msg, keys)
}

// Warn 记录警告日志
func (a *ZapAdapter) Warn(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.WarnLevel, msg, keys)
}

// Error 记录错误日志
func (a *ZapAdapter) Error(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.ErrorLevel, msg, keys)
}

// DPanic 记录致命错误日志并 panic
func (a *ZapAdapter) DPanic(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.DPanicLevel, msg, keys)
}

// Panic 记录日志并 panic
func (a *ZapAdapter) Panic(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.PanicLevel, msg, keys)
}

// Fatal 记录日志并退出程序
func (a *ZapAdapter) Fatal(ctx context.Context, msg string, keys ...log.KeyValue) {
	a.log(ctx, log.FatalLevel, msg, keys)
}

// Sync 同步日志缓冲区
func (a *ZapAdapter) Sync() error {
	return a.logger.Sync()
}

// With 返回带有额外字段的日志记录器
func (a *ZapAdapter) With(ctx context.Context, keys ...log.KeyValue) log.Logger {
	fields := a.toZapFields(keys)
	return &ZapAdapter{
		logger:     a.logger.With(fields...),
		level:      a.level,
		format:     a.format,
		timeFormat: a.timeFormat,
		addCaller:  a.addCaller,
		callerSkip: a.callerSkip,
		encoder:    a.encoder,
		output:     a.output,
	}
}

var _ log.Logger = (*ZapAdapter)(nil)

// zapTextWriter 是文本写入器(实现 io.Writer 接口)
type zapTextWriter struct{}

func (w *zapTextWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// MustZap 如果 err 不为 nil 则 panic
func MustZap(err error) {
	if err != nil {
		panic(fmt.Sprintf("log zap init failed: %v", err))
	}
}

// ZapLoggerBuilder 是 zap 日志构建器
type ZapLoggerBuilder struct {
	cfg  *ZapConfig
	opts []ZapOption
}

// NewZapLoggerBuilder 创建新的日志构建器
func NewZapLoggerBuilder() *ZapLoggerBuilder {
	return &ZapLoggerBuilder{}
}

// Level 设置日志级别
func (b *ZapLoggerBuilder) Level(level string) *ZapLoggerBuilder {
	b.cfg.Level = level
	return b
}

// Format 设置输出格式
func (b *ZapLoggerBuilder) Format(format string) *ZapLoggerBuilder {
	b.cfg.Format = format
	return b
}

// TimeFormat 设置时间格式
func (b *ZapLoggerBuilder) TimeFormat(timeFormat string) *ZapLoggerBuilder {
	b.cfg.TimeFormat = timeFormat
	return b
}

// AddCaller 设置是否添加调用者信息
func (b *ZapLoggerBuilder) AddCaller(addCaller bool) *ZapLoggerBuilder {
	b.cfg.AddCaller = addCaller
	return b
}

// CallerSkip 设置调用者跳过帧数
func (b *ZapLoggerBuilder) CallerSkip(callerSkip int) *ZapLoggerBuilder {
	b.cfg.CallerSkip = callerSkip
	return b
}

// Output 设置输出文件
func (b *ZapLoggerBuilder) Output(output string) *ZapLoggerBuilder {
	b.cfg.Output = output
	return b
}

// Build 构建日志记录器
func (b *ZapLoggerBuilder) Build(opts ...ZapOption) log.Logger {
	if b.cfg == nil {
		b.cfg = &ZapConfig{}
	}
	for _, opt := range b.opts {
		opts = append(opts, opt)
	}
	return NewZapAdapter(b.cfg, opts...)
}
