package log

import (
	"context"
	"sync/atomic"
)

// contextKey 上下文键类型
type contextKey struct{}

// TraceContextKey 追踪 ID 上下文键
var TraceContextKey = contextKey{}

// WithTraceID 将 trace_id 注入上下文
//
// 参数:
//   - ctx: 原始上下文
//   - traceID: 追踪 ID
//
// 返回:
//   - context.Context: 包含 trace_id 的新上下文
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceContextKey, traceID)
}

// GetTraceID 从上下文获取 trace_id
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - string: trace_id，不存在返回空字符串
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceContextKey).(string); ok {
		return traceID
	}
	return ""
}

// ContextLogger 上下文感知日志器
//
// 自动从 ctx 中提取 trace_id 等信息并添加到日志中。
type ContextLogger struct {
	logger Logger
}

// NewContextLogger 创建上下文感知日志器
//
// 参数:
//   - logger: 底层日志器
//
// 返回:
//   - *ContextLogger: 上下文日志器实例
func NewContextLogger(logger Logger) *ContextLogger {
	return &ContextLogger{logger: logger}
}

// Debug 记录调试日志
func (l *ContextLogger) Debug(ctx context.Context, msg string, keys ...KeyValue) {
	keys = appendContextKeys(ctx, keys)
	l.logger.Debug(ctx, msg, keys...)
}

// Info 记录信息日志
func (l *ContextLogger) Info(ctx context.Context, msg string, keys ...KeyValue) {
	keys = appendContextKeys(ctx, keys)
	l.logger.Info(ctx, msg, keys...)
}

// Warn 记录警告日志
func (l *ContextLogger) Warn(ctx context.Context, msg string, keys ...KeyValue) {
	keys = appendContextKeys(ctx, keys)
	l.logger.Warn(ctx, msg, keys...)
}

// Error 记录错误日志
func (l *ContextLogger) Error(ctx context.Context, msg string, keys ...KeyValue) {
	keys = appendContextKeys(ctx, keys)
	l.logger.Error(ctx, msg, keys...)
}

// DPanic 记录致命错误日志并 panic
func (l *ContextLogger) DPanic(ctx context.Context, msg string, keys ...KeyValue) {
	keys = appendContextKeys(ctx, keys)
	l.logger.DPanic(ctx, msg, keys...)
}

// Panic 记录日志并 panic
func (l *ContextLogger) Panic(ctx context.Context, msg string, keys ...KeyValue) {
	keys = appendContextKeys(ctx, keys)
	l.logger.Panic(ctx, msg, keys...)
}

// Fatal 记录致命级别日志
func (l *ContextLogger) Fatal(ctx context.Context, msg string, keys ...KeyValue) {
	keys = appendContextKeys(ctx, keys)
	l.logger.Fatal(ctx, msg, keys...)
}

// Sync 同步日志缓冲区
func (l *ContextLogger) Sync() error {
	return l.logger.Sync()
}

// With 返回带有额外字段的日志记录器
func (l *ContextLogger) With(ctx context.Context, keys ...KeyValue) Logger {
	return &ContextLogger{
		logger: l.logger.With(ctx, keys...),
	}
}

// appendContextKeys 从上下文提取键值对
func appendContextKeys(ctx context.Context, keys []KeyValue) []KeyValue {
	if traceID := GetTraceID(ctx); traceID != "" {
		keys = append(keys, KeyValue{Key: "trace_id", Value: traceID})
	}
	return keys
}

var _ Logger = (*ContextLogger)(nil)

// DynamicLevelLogger 动态级别日志器
//
// 支持运行时动态调整日志级别，无需重启服务。
type DynamicLevelLogger struct {
	logger Logger
	level  *atomic.Int32 // 当前日志级别（指针，子日志器共享）
}

// NewDynamicLevelLogger 创建动态级别日志器
//
// 参数:
//   - logger: 底层日志器
//   - initialLevel: 初始日志级别
//
// 返回:
//   - *DynamicLevelLogger: 动态级别日志器实例
func NewDynamicLevelLogger(logger Logger, initialLevel Level) *DynamicLevelLogger {
	d := &DynamicLevelLogger{
		logger: logger,
		level:  &atomic.Int32{},
	}
	d.level.Store(int32(initialLevel))
	return d
}

// SetLevel 动态设置日志级别
//
// 参数:
//   - level: 新的日志级别
func (d *DynamicLevelLogger) SetLevel(level Level) {
	d.level.Store(int32(level))
}

// GetLevel 获取当前日志级别
//
// 返回:
//   - Level: 当前日志级别
func (d *DynamicLevelLogger) GetLevel() Level {
	return Level(d.level.Load())
}

// shouldLog 判断是否应该记录指定级别的日志
func (d *DynamicLevelLogger) shouldLog(level Level) bool {
	return level >= Level(d.level.Load())
}

// Debug 记录调试日志
func (d *DynamicLevelLogger) Debug(ctx context.Context, msg string, keys ...KeyValue) {
	if d.shouldLog(DebugLevel) {
		d.logger.Debug(ctx, msg, keys...)
	}
}

// Info 记录信息日志
func (d *DynamicLevelLogger) Info(ctx context.Context, msg string, keys ...KeyValue) {
	if d.shouldLog(InfoLevel) {
		d.logger.Info(ctx, msg, keys...)
	}
}

// Warn 记录警告日志
func (d *DynamicLevelLogger) Warn(ctx context.Context, msg string, keys ...KeyValue) {
	if d.shouldLog(WarnLevel) {
		d.logger.Warn(ctx, msg, keys...)
	}
}

// Error 记录错误日志
func (d *DynamicLevelLogger) Error(ctx context.Context, msg string, keys ...KeyValue) {
	if d.shouldLog(ErrorLevel) {
		d.logger.Error(ctx, msg, keys...)
	}
}

// DPanic 记录致命错误日志并 panic
func (d *DynamicLevelLogger) DPanic(ctx context.Context, msg string, keys ...KeyValue) {
	d.logger.DPanic(ctx, msg, keys...)
}

// Panic 记录日志并 panic
func (d *DynamicLevelLogger) Panic(ctx context.Context, msg string, keys ...KeyValue) {
	d.logger.Panic(ctx, msg, keys...)
}

// Fatal 记录致命级别日志
func (d *DynamicLevelLogger) Fatal(ctx context.Context, msg string, keys ...KeyValue) {
	d.logger.Fatal(ctx, msg, keys...)
}

// Sync 同步日志缓冲区
func (d *DynamicLevelLogger) Sync() error {
	return d.logger.Sync()
}

// With 返回带有额外字段的日志记录器
func (d *DynamicLevelLogger) With(ctx context.Context, keys ...KeyValue) Logger {
	return &DynamicLevelLogger{
		logger: d.logger.With(ctx, keys...),
		level:  d.level, // 共享 atomic 指针
	}
}

var _ Logger = (*DynamicLevelLogger)(nil)
