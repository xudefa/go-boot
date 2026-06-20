package log

import (
	"context"
	"sync/atomic"
)

// LoggerBuilder 日志器构建器
//
// 使用 Builder 模式简化日志器创建流程，支持链式配置。
//
// 使用示例：
//
//	logger := NewLoggerBuilder().
//	    Level(InfoLevel).
//	    Format("json").
//	    AddSource(true).
//	    OutputPath("/var/log/app.log").
//	    Sampler(NewRandomSampler(0.1)). // 10% 采样率
//	    Build()
type LoggerBuilder struct {
	level      Level
	format     string
	addSource  bool
	outputPath string
	sampler    Sampler
	name       string
}

// NewLoggerBuilder 创建日志器构建器
//
// 返回:
//   - *LoggerBuilder: 构建器实例，默认 INFO 级别、JSON 格式
func NewLoggerBuilder() *LoggerBuilder {
	return &LoggerBuilder{
		level:  InfoLevel,
		format: "json",
	}
}

// Level 设置日志级别
func (b *LoggerBuilder) Level(level Level) *LoggerBuilder {
	b.level = level
	return b
}

// Format 设置输出格式（json 或 text）
func (b *LoggerBuilder) Format(format string) *LoggerBuilder {
	b.format = format
	return b
}

// AddSource 设置是否添加源码位置
func (b *LoggerBuilder) AddSource(addSource bool) *LoggerBuilder {
	b.addSource = addSource
	return b
}

// OutputPath 设置日志文件输出路径
func (b *LoggerBuilder) OutputPath(path string) *LoggerBuilder {
	b.outputPath = path
	return b
}

// Sampler 设置采样策略
func (b *LoggerBuilder) Sampler(sampler Sampler) *LoggerBuilder {
	b.sampler = sampler
	return b
}

// Name 设置日志器名称
func (b *LoggerBuilder) Name(name string) *LoggerBuilder {
	b.name = name
	return b
}

// Build 构建日志器
func (b *LoggerBuilder) Build() Logger {
	opts := []Option{
		WithLevel(b.level),
		WithFormat(b.format),
		WithAddSource(b.addSource),
	}

	if b.outputPath != "" {
		opts = append(opts, WithOutputPath(b.outputPath))
	}

	var logger Logger = NewSlogLogger(opts...)

	// 包装采样器
	if b.sampler != nil {
		logger = NewSampledLogger(logger, b.sampler)
	}

	// 包装命名
	if b.name != "" {
		logger = logger.With(context.Background(), KeyValue{Key: "logger", Value: b.name})
	}

	return logger
}

// Sampler 采样策略接口
//
// 用于控制高频日志的采样率，避免日志过多影响性能。
type Sampler interface {
	// ShouldSample 判断是否应该记录本次日志
	ShouldSample() bool
}

// RandomSampler 随机采样器
//
// 按固定概率采样，适用于高频日志场景。
type RandomSampler struct {
	rate float64 // 采样率 0.0-1.0
}

// NewRandomSampler 创建随机采样器
//
// 参数:
//   - rate: 采样率，0.0 表示不采样，1.0 表示全部采样
func NewRandomSampler(rate float64) *RandomSampler {
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	return &RandomSampler{rate: rate}
}

// ShouldSample 判断是否采样
func (s *RandomSampler) ShouldSample() bool {
	return randomFloat() < s.rate
}

// ThresholdSampler 阈值采样器
//
// 每 N 次日志记录一次，适用于计数场景。
type ThresholdSampler struct {
	threshold int64
	counter   atomic.Int64
}

// NewThresholdSampler 创建阈值采样器
//
// 参数:
//   - threshold: 每 N 次记录一次
func NewThresholdSampler(threshold int64) *ThresholdSampler {
	if threshold <= 0 {
		threshold = 1
	}
	return &ThresholdSampler{threshold: threshold}
}

// ShouldSample 判断是否采样
func (s *ThresholdSampler) ShouldSample() bool {
	return s.counter.Add(1)%s.threshold == 0
}

// SampledLogger 带采样的日志器
type SampledLogger struct {
	logger  Logger
	sampler Sampler
}

// NewSampledLogger 创建带采样的日志器
func NewSampledLogger(logger Logger, sampler Sampler) *SampledLogger {
	return &SampledLogger{
		logger:  logger,
		sampler: sampler,
	}
}

// Debug 记录调试日志
func (l *SampledLogger) Debug(ctx context.Context, msg string, keys ...KeyValue) {
	if l.sampler.ShouldSample() {
		l.logger.Debug(ctx, msg, keys...)
	}
}

// Info 记录信息日志
func (l *SampledLogger) Info(ctx context.Context, msg string, keys ...KeyValue) {
	if l.sampler.ShouldSample() {
		l.logger.Info(ctx, msg, keys...)
	}
}

// Warn 记录警告日志
func (l *SampledLogger) Warn(ctx context.Context, msg string, keys ...KeyValue) {
	if l.sampler.ShouldSample() {
		l.logger.Warn(ctx, msg, keys...)
	}
}

// Error 记录错误日志（错误日志不采样，全部记录）
func (l *SampledLogger) Error(ctx context.Context, msg string, keys ...KeyValue) {
	l.logger.Error(ctx, msg, keys...)
}

// DPanic 记录致命错误日志并 panic
func (l *SampledLogger) DPanic(ctx context.Context, msg string, keys ...KeyValue) {
	l.logger.DPanic(ctx, msg, keys...)
}

// Panic 记录日志并 panic
func (l *SampledLogger) Panic(ctx context.Context, msg string, keys ...KeyValue) {
	l.logger.Panic(ctx, msg, keys...)
}

// Fatal 记录致命级别日志
func (l *SampledLogger) Fatal(ctx context.Context, msg string, keys ...KeyValue) {
	l.logger.Fatal(ctx, msg, keys...)
}

// Sync 同步日志缓冲区
func (l *SampledLogger) Sync() error {
	return l.logger.Sync()
}

// With 返回带有额外字段的日志记录器
func (l *SampledLogger) With(ctx context.Context, keys ...KeyValue) Logger {
	return &SampledLogger{
		logger:  l.logger.With(ctx, keys...),
		sampler: l.sampler,
	}
}

var _ Logger = (*SampledLogger)(nil)

// randomFloat 生成 0.0-1.0 之间的随机数
// 使用简单的线性同余生成器，避免引入 math/rand 包
func randomFloat() float64 {
	// 使用 atomic 操作保证并发安全
	return float64(atomic.AddUint64(&randomSeed, 1)%1000) / 1000.0
}

var randomSeed uint64
