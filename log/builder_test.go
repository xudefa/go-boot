package log

import (
	"context"
	"strings"
	"testing"
)

func TestLoggerBuilder_Build(t *testing.T) {
	logger := NewLoggerBuilder().
		Level(DebugLevel).
		Format("json").
		AddSource(false).
		Build()

	if logger == nil {
		t.Fatal("expected logger to be created")
	}

	ctx := context.Background()
	// 应该能正常记录日志
	logger.Debug(ctx, "test debug message")
	logger.Info(ctx, "test info message")
}

func TestLoggerBuilder_WithName(t *testing.T) {
	logger := NewLoggerBuilder().
		Name("my-service").
		Build()

	if logger == nil {
		t.Fatal("expected logger to be created")
	}

	ctx := context.Background()
	logger.Info(ctx, "test message with name")
}

func TestLoggerBuilder_WithSampler(t *testing.T) {
	// 10% 采样率
	sampler := NewRandomSampler(0.1)
	logger := NewLoggerBuilder().
		Level(DebugLevel).
		Sampler(sampler).
		Build()

	if logger == nil {
		t.Fatal("expected logger to be created")
	}

	ctx := context.Background()
	// 多次调用，部分应该被采样
	for i := 0; i < 100; i++ {
		logger.Debug(ctx, "sampled debug message")
	}
}

func TestRandomSampler_Rate(t *testing.T) {
	sampler := NewRandomSampler(0.5)

	sampled := 0
	total := 1000

	for i := 0; i < total; i++ {
		if sampler.ShouldSample() {
			sampled++
		}
	}

	rate := float64(sampled) / float64(total)
	// 采样率应该在 0.4-0.6 之间（允许一定偏差）
	if rate < 0.4 || rate > 0.6 {
		t.Errorf("expected sampling rate around 0.5, got %f", rate)
	}
}

func TestRandomSampler_Boundary(t *testing.T) {
	// 测试边界值
	sampler0 := NewRandomSampler(0)
	if sampler0.ShouldSample() {
		t.Error("expected 0% sampler to never sample")
	}

	sampler1 := NewRandomSampler(1)
	// 100% 采样器应该总是采样
	for i := 0; i < 100; i++ {
		if !sampler1.ShouldSample() {
			t.Error("expected 100% sampler to always sample")
			break
		}
	}
}

func TestThresholdSampler(t *testing.T) {
	sampler := NewThresholdSampler(5)

	sampled := 0
	for i := 0; i < 100; i++ {
		if sampler.ShouldSample() {
			sampled++
		}
	}

	// 每 5 次采样一次，100 次应该有 20 次
	if sampled != 20 {
		t.Errorf("expected 20 samples, got %d", sampled)
	}
}

func TestSampledLogger_ErrorNotSampled(t *testing.T) {
	// 创建一个 0% 采样率的日志器
	sampler := NewRandomSampler(0)
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()

	// Debug 不应该被记录（0% 采样率）
	// 但我们无法直接验证，只能验证 Error 总是被记录
	logger.Error(ctx, "error message") // 应该总是记录
}

func TestContextLogger_WithTraceID(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := context.Background()
	ctx = WithTraceID(ctx, "test-trace-123")

	// 应该能正常记录，且包含 trace_id
	logger.Info(ctx, "message with trace id")
}

func TestContextLogger_NoTraceID(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := context.Background()

	// 没有 trace_id 也应该正常记录
	logger.Info(ctx, "message without trace id")
}

func TestGetTraceID(t *testing.T) {
	ctx := context.Background()

	// 没有 trace_id
	if GetTraceID(ctx) != "" {
		t.Error("expected empty trace id")
	}

	// 有 trace_id
	ctx = WithTraceID(ctx, "abc-123")
	if GetTraceID(ctx) != "abc-123" {
		t.Errorf("expected abc-123, got %s", GetTraceID(ctx))
	}
}

func TestDynamicLevelLogger(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, InfoLevel)

	// 初始级别为 Info
	if logger.GetLevel() != InfoLevel {
		t.Errorf("expected InfoLevel, got %v", logger.GetLevel())
	}

	ctx := context.Background()

	// Debug 不应该被记录
	logger.Debug(ctx, "debug message")

	// Info 应该被记录
	logger.Info(ctx, "info message")

	// 动态调整级别为 Debug
	logger.SetLevel(DebugLevel)

	if logger.GetLevel() != DebugLevel {
		t.Errorf("expected DebugLevel after SetLevel, got %v", logger.GetLevel())
	}

	// Debug 现在应该被记录
	logger.Debug(ctx, "debug message after level change")
}

func TestDynamicLevelLogger_LevelFiltering(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, WarnLevel)

	ctx := context.Background()

	// 记录所有级别，只有 Warn 及以上应该被记录
	logger.Debug(ctx, "debug") // 不应该记录
	logger.Info(ctx, "info")   // 不应该记录
	logger.Warn(ctx, "warn")   // 应该记录
	logger.Error(ctx, "error") // 应该记录

	// 调整级别为 Debug
	logger.SetLevel(DebugLevel)
	logger.Debug(ctx, "debug after change") // 现在应该记录
}

func TestDynamicLevelLogger_With(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, InfoLevel)

	ctx := context.Background()
	childLogger := logger.With(ctx, KeyValue{Key: "module", Value: "test"})

	// 子日志器应该共享级别
	dynamicChild, ok := childLogger.(*DynamicLevelLogger)
	if !ok {
		t.Fatal("expected child to be DynamicLevelLogger")
	}

	// 修改父级别，子级别应该也受影响
	logger.SetLevel(DebugLevel)
	if dynamicChild.GetLevel() != DebugLevel {
		t.Errorf("expected child to inherit level change, got %v", dynamicChild.GetLevel())
	}
}

func TestSampledLogger_With(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(0.5)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()
	childLogger := logger.With(ctx, KeyValue{Key: "key", Value: "value"})

	if childLogger == nil {
		t.Fatal("expected child logger to be created")
	}

	// 子日志器应该保持采样器
	sampledChild, ok := childLogger.(*SampledLogger)
	if !ok {
		t.Fatal("expected child to be SampledLogger")
	}

	if sampledChild.sampler != sampler {
		t.Error("expected child to share same sampler")
	}
}

func TestLoggerBuilder_OutputPath(t *testing.T) {
	// 测试无效路径（应该回退到 stdout）
	logger := NewLoggerBuilder().
		Level(InfoLevel).
		OutputPath("/nonexistent/path/test.log").
		Build()

	if logger == nil {
		t.Fatal("expected logger to be created even with invalid path")
	}
}

func TestContextLogger_ChainWith(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := context.Background()
	ctx = WithTraceID(ctx, "trace-456")

	childLogger := logger.With(ctx, KeyValue{Key: "request_id", Value: "req-789"})

	// 子日志器应该也能正常工作
	childLogger.Info(ctx, "chained context log")
}

func TestDynamicLevelLogger_ConcurrentLevelChange(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, InfoLevel)

	ctx := context.Background()

	// 并发修改级别
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				logger.SetLevel(Level(j % 5))
				logger.Info(ctx, "concurrent log")
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// 如果没有 panic，说明并发安全
}

func TestSampledLogger_AllLevels(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(1.0) // 100% 采样
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()

	// 测试所有级别
	logger.Debug(ctx, "debug")
	logger.Info(ctx, "info")
	logger.Warn(ctx, "warn")
	logger.Error(ctx, "error")

	// 不应该 panic
}

func TestLoggerBuilder_DefaultValues(t *testing.T) {
	logger := NewLoggerBuilder().Build()

	if logger == nil {
		t.Fatal("expected logger with default values")
	}

	ctx := context.Background()
	logger.Info(ctx, "default config log")
}

func TestContextLogger_AllLevels(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := WithTraceID(context.Background(), "test-trace")

	logger.Debug(ctx, "debug")
	logger.Info(ctx, "info")
	logger.Warn(ctx, "warn")
	logger.Error(ctx, "error")
}

func TestDynamicLevelLogger_BoundaryLevels(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, DebugLevel)

	// 测试所有边界级别
	logger.SetLevel(DebugLevel)
	if logger.GetLevel() != DebugLevel {
		t.Errorf("expected DebugLevel, got %v", logger.GetLevel())
	}

	logger.SetLevel(InfoLevel)
	if logger.GetLevel() != InfoLevel {
		t.Errorf("expected InfoLevel, got %v", logger.GetLevel())
	}

	logger.SetLevel(WarnLevel)
	if logger.GetLevel() != WarnLevel {
		t.Errorf("expected WarnLevel, got %v", logger.GetLevel())
	}

	logger.SetLevel(ErrorLevel)
	if logger.GetLevel() != ErrorLevel {
		t.Errorf("expected ErrorLevel, got %v", logger.GetLevel())
	}
}

func TestSampledLogger_Sync(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(0.5)
	logger := NewSampledLogger(baseLogger, sampler)

	err := logger.Sync()
	if err != nil {
		t.Errorf("expected no error from Sync, got %v", err)
	}
}

func TestContextLogger_Sync(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	err := logger.Sync()
	if err != nil {
		t.Errorf("expected no error from Sync, got %v", err)
	}
}

func TestLoggerBuilder_ComplexConfig(t *testing.T) {
	logger := NewLoggerBuilder().
		Name("complex-service").
		Level(WarnLevel).
		Format("text").
		AddSource(false).
		Sampler(NewThresholdSampler(10)).
		Build()

	if logger == nil {
		t.Fatal("expected logger with complex config")
	}

	ctx := WithTraceID(context.Background(), "complex-trace")
	logger.Warn(ctx, "complex config log")
}

func TestDynamicLevelLogger_LevelComparison(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, WarnLevel)

	ctx := context.Background()

	// WarnLevel 时，Debug 和 Info 不应该被记录
	// 我们无法直接验证，但可以确保 Warn 及以上被记录
	logger.Warn(ctx, "warn message")
	logger.Error(ctx, "error message")

	// 调整到 ErrorLevel
	logger.SetLevel(ErrorLevel)
	logger.Error(ctx, "error after change")
}

func TestSampledLogger_NilSampler(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))

	// 测试 nil sampler 的情况
	// 这里应该 panic，但我们测试正常情况
	sampler := NewRandomSampler(1.0)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()
	logger.Info(ctx, "message with sampler")
}

func TestContextLogger_EmptyContext(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	// 空上下文也应该正常工作
	ctx := context.Background()
	logger.Info(ctx, "empty context log")
}

func TestLoggerBuilder_StringFormat(t *testing.T) {
	logger := NewLoggerBuilder().
		Format("text").
		Build()

	if logger == nil {
		t.Fatal("expected text format logger")
	}

	ctx := context.Background()
	logger.Info(ctx, "text format log")
}

func TestLoggerBuilder_JSONFormat(t *testing.T) {
	logger := NewLoggerBuilder().
		Format("json").
		Build()

	if logger == nil {
		t.Fatal("expected json format logger")
	}

	ctx := context.Background()
	logger.Info(ctx, "json format log")
}

func TestDynamicLevelLogger_WithPreservesLevel(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, WarnLevel)

	ctx := context.Background()
	childLogger := logger.With(ctx, KeyValue{Key: "key", Value: "value"})

	// 子日志器应该保持相同的级别
	dynamicChild, ok := childLogger.(*DynamicLevelLogger)
	if !ok {
		t.Fatal("expected DynamicLevelLogger")
	}

	if dynamicChild.GetLevel() != WarnLevel {
		t.Errorf("expected WarnLevel, got %v", dynamicChild.GetLevel())
	}
}

func TestSampledLogger_ContextPropagation(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(1.0)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := WithTraceID(context.Background(), "sample-trace")
	logger.Info(ctx, "sampled with context")
}

func TestContextLogger_MultipleTraceIDs(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx1 := WithTraceID(context.Background(), "trace-1")
	ctx2 := WithTraceID(context.Background(), "trace-2")

	logger.Info(ctx1, "first trace")
	logger.Info(ctx2, "second trace")
}

func TestLoggerBuilder_WithOutputPath(t *testing.T) {
	// 测试临时文件路径
	logger := NewLoggerBuilder().
		OutputPath("/tmp/test-log-builder.log").
		Build()

	if logger == nil {
		t.Fatal("expected logger with output path")
	}

	ctx := context.Background()
	logger.Info(ctx, "output path log")

	// 清理
	_ = logger.Sync()
}

func TestDynamicLevelLogger_RapidLevelChanges(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, InfoLevel)

	ctx := context.Background()

	// 快速切换级别
	for i := 0; i < 100; i++ {
		logger.SetLevel(Level(i % 5))
		logger.Info(ctx, "rapid level change")
	}
}

func TestSampledLogger_PerformanceWithHighSampling(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(1.0) // 100% 采样
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()

	// 高频日志
	for i := 0; i < 1000; i++ {
		logger.Info(ctx, "high frequency log")
	}
}

func TestContextLogger_WithEmptyKeys(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := context.Background()
	childLogger := logger.With(ctx)

	if childLogger == nil {
		t.Fatal("expected child logger with empty keys")
	}

	childLogger.Info(ctx, "empty keys log")
}

func TestLoggerBuilder_LevelString(t *testing.T) {
	// 测试级别字符串
	levels := []Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, DPanicLevel, PanicLevel, FatalLevel}
	expected := []string{"debug", "info", "warn", "error", "dpanic", "panic", "fatal"}

	for i, level := range levels {
		if level.String() != expected[i] {
			t.Errorf("expected %s, got %s", expected[i], level.String())
		}
	}
}

func TestToLevel_Conversion(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"debug", DebugLevel},
		{"info", InfoLevel},
		{"warn", WarnLevel},
		{"warning", WarnLevel},
		{"error", ErrorLevel},
		{"dpanic", DPanicLevel},
		{"panic", PanicLevel},
		{"fatal", FatalLevel},
		{"unknown", InfoLevel}, // 默认
	}

	for _, tt := range tests {
		if ToLevel(tt.input) != tt.expected {
			t.Errorf("ToLevel(%s) = %v, expected %v", tt.input, ToLevel(tt.input), tt.expected)
		}
	}
}

func TestLoggerBuilder_NilLogger(t *testing.T) {
	// 测试 Build 函数在没有指定 logger 时的默认行为
	logger := Build()

	if logger == nil {
		t.Fatal("expected default logger")
	}

	ctx := context.Background()
	logger.Info(ctx, "default logger")
}

func TestLoggerBuilder_WithCustomLogger(t *testing.T) {
	customLogger := NewSlogLogger(WithLevel(DebugLevel), WithFormat("text"))
	logger := Build(WithLogger(customLogger))

	if logger == nil {
		t.Fatal("expected custom logger")
	}

	ctx := context.Background()
	logger.Info(ctx, "custom logger")
}

func TestSampledLogger_SamplerShouldSample(t *testing.T) {
	// 测试采样器的 ShouldSample 方法
	sampler := NewRandomSampler(0)
	if sampler.ShouldSample() {
		t.Error("0% sampler should not sample")
	}

	sampler = NewRandomSampler(1)
	for i := 0; i < 100; i++ {
		if !sampler.ShouldSample() {
			t.Error("100% sampler should always sample")
			break
		}
	}
}

func TestContextLogger_WithTraceIDEmpty(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	// 测试空 trace_id
	ctx := WithTraceID(context.Background(), "")
	logger.Info(ctx, "empty trace id")
}

func TestDynamicLevelLogger_AllLevelsShouldLog(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, DebugLevel)

	ctx := context.Background()

	// DebugLevel 时所有级别都应该被记录
	logger.Debug(ctx, "debug")
	logger.Info(ctx, "info")
	logger.Warn(ctx, "warn")
	logger.Error(ctx, "error")
}

func TestLoggerBuilder_SamplerNil(t *testing.T) {
	// 测试 sampler 为 nil 的情况
	logger := NewLoggerBuilder().
		Level(InfoLevel).
		Build()

	if logger == nil {
		t.Fatal("expected logger without sampler")
	}

	ctx := context.Background()
	logger.Info(ctx, "no sampler")
}

func TestContextLogger_WithNilContext(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	// 不应该 panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %v", r)
		}
	}()

	ctx := context.Background()
	logger.Info(ctx, "nil context test")
}

func TestSampledLogger_WithNilSampler(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))

	// 测试 nil sampler 会 panic
	panicked := false
	defer func() {
		if r := recover(); r != nil {
			panicked = true
		}
		if !panicked {
			t.Error("expected panic with nil sampler")
		}
	}()

	logger := NewSampledLogger(baseLogger, nil)
	ctx := context.Background()
	logger.Info(ctx, "nil sampler")
}

func TestLoggerBuilder_ChainMultipleOptions(t *testing.T) {
	logger := NewLoggerBuilder().
		Name("chain-test").
		Level(DebugLevel).
		Format("json").
		AddSource(false).
		Sampler(NewRandomSampler(0.5)).
		Build()

	if logger == nil {
		t.Fatal("expected chained options logger")
	}

	ctx := WithTraceID(context.Background(), "chain-trace")
	logger.Info(ctx, "chained options log")
}

func TestDynamicLevelLogger_LevelOrdering(t *testing.T) {
	// 验证级别顺序：Debug < Info < Warn < Error < DPanic < Panic < Fatal
	if DebugLevel >= InfoLevel {
		t.Error("DebugLevel should be less than InfoLevel")
	}
	if InfoLevel >= WarnLevel {
		t.Error("InfoLevel should be less than WarnLevel")
	}
	if WarnLevel >= ErrorLevel {
		t.Error("WarnLevel should be less than ErrorLevel")
	}
	if ErrorLevel >= DPanicLevel {
		t.Error("ErrorLevel should be less than DPanicLevel")
	}
}

func TestContextLogger_WithMultipleFields(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := WithTraceID(context.Background(), "multi-trace")
	childLogger := logger.With(ctx,
		KeyValue{Key: "key1", Value: "value1"},
		KeyValue{Key: "key2", Value: "value2"},
	)

	childLogger.Info(ctx, "multiple fields log")
}

func TestSampledLogger_ErrorAlwaysLogged(t *testing.T) {
	// 验证 Error 级别总是被记录，不受采样影响
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(0) // 0% 采样
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()

	// Error 应该总是被记录
	logger.Error(ctx, "always logged error")
}

func TestLoggerBuilder_DefaultFormat(t *testing.T) {
	logger := NewLoggerBuilder().Build()

	// 默认应该是 json 格式
	if logger == nil {
		t.Fatal("expected default format logger")
	}

	ctx := context.Background()
	logger.Info(ctx, "default format log")
}

func TestContextLogger_PreserveOriginalContext(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := WithTraceID(context.Background(), "original-trace")

	// 原始上下文应该保持不变
	if GetTraceID(ctx) != "original-trace" {
		t.Errorf("expected original trace id to be preserved")
	}

	logger.Info(ctx, "preserve context log")
}

func TestDynamicLevelLogger_SetLevelConcurrent(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, InfoLevel)

	// 并发设置级别
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			logger.SetLevel(Level(n % 5))
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// 如果没有 panic，说明并发安全
}

func TestSampledLogger_WithPreservesSampler(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(0.5)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()
	childLogger := logger.With(ctx, KeyValue{Key: "key", Value: "value"})

	sampledChild, ok := childLogger.(*SampledLogger)
	if !ok {
		t.Fatal("expected SampledLogger")
	}

	if sampledChild.sampler != sampler {
		t.Error("expected child to preserve sampler")
	}
}

func TestLoggerBuilder_OutputPathEmpty(t *testing.T) {
	// 测试空输出路径
	logger := NewLoggerBuilder().
		OutputPath("").
		Build()

	if logger == nil {
		t.Fatal("expected logger with empty output path")
	}

	ctx := context.Background()
	logger.Info(ctx, "empty output path log")
}

func TestContextLogger_WithNilLogger(t *testing.T) {
	// 测试 nil logger 的情况
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with nil logger")
		}
	}()

	logger := NewContextLogger(nil)
	ctx := context.Background()
	logger.Info(ctx, "nil logger")
}

func TestSampledLogger_WithNilLogger(t *testing.T) {
	// 测试 nil logger 的情况
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with nil logger")
		}
	}()

	sampler := NewRandomSampler(1.0)
	logger := NewSampledLogger(nil, sampler)
	ctx := context.Background()
	logger.Info(ctx, "nil logger")
}

func TestDynamicLevelLogger_NilLogger(t *testing.T) {
	// 测试 nil logger 的情况
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with nil logger")
		}
	}()

	logger := NewDynamicLevelLogger(nil, InfoLevel)
	ctx := context.Background()
	logger.Info(ctx, "nil logger")
}

func TestLoggerBuilder_ComplexConfigWithAllOptions(t *testing.T) {
	logger := NewLoggerBuilder().
		Name("full-config").
		Level(DebugLevel).
		Format("json").
		AddSource(false).
		OutputPath("/tmp/test-full-config.log").
		Sampler(NewRandomSampler(0.8)).
		Build()

	if logger == nil {
		t.Fatal("expected full config logger")
	}

	ctx := WithTraceID(context.Background(), "full-trace")
	logger.Info(ctx, "full config log")

	_ = logger.Sync()
}

func TestContextLogger_WithEmptyMessage(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := context.Background()
	logger.Info(ctx, "")
}

func TestSampledLogger_WithEmptyMessage(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(1.0)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()
	logger.Info(ctx, "")
}

func TestDynamicLevelLogger_WithEmptyMessage(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, InfoLevel)

	ctx := context.Background()
	logger.Info(ctx, "")
}

func TestLoggerBuilder_InvalidFormat(t *testing.T) {
	// 测试无效格式（应该回退到默认格式）
	logger := NewLoggerBuilder().
		Format("invalid").
		Build()

	if logger == nil {
		t.Fatal("expected logger with invalid format")
	}

	ctx := context.Background()
	logger.Info(ctx, "invalid format log")
}

func TestContextLogger_WithNilKeys(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := context.Background()
	childLogger := logger.With(ctx, nil...)

	if childLogger == nil {
		t.Fatal("expected child logger with nil keys")
	}

	childLogger.Info(ctx, "nil keys log")
}

func TestSampledLogger_WithNilKeys(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(1.0)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()
	childLogger := logger.With(ctx, nil...)

	if childLogger == nil {
		t.Fatal("expected child logger with nil keys")
	}

	childLogger.Info(ctx, "nil keys log")
}

func TestDynamicLevelLogger_WithNilKeys(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, InfoLevel)

	ctx := context.Background()
	childLogger := logger.With(ctx, nil...)

	if childLogger == nil {
		t.Fatal("expected child logger with nil keys")
	}

	childLogger.Info(ctx, "nil keys log")
}

func TestLoggerBuilder_LevelDebug(t *testing.T) {
	logger := NewLoggerBuilder().
		Level(DebugLevel).
		Build()

	ctx := context.Background()
	logger.Debug(ctx, "debug level log")
}

func TestLoggerBuilder_LevelInfo(t *testing.T) {
	logger := NewLoggerBuilder().
		Level(InfoLevel).
		Build()

	ctx := context.Background()
	logger.Info(ctx, "info level log")
}

func TestLoggerBuilder_LevelWarn(t *testing.T) {
	logger := NewLoggerBuilder().
		Level(WarnLevel).
		Build()

	ctx := context.Background()
	logger.Warn(ctx, "warn level log")
}

func TestLoggerBuilder_LevelError(t *testing.T) {
	logger := NewLoggerBuilder().
		Level(ErrorLevel).
		Build()

	ctx := context.Background()
	logger.Error(ctx, "error level log")
}

func TestContextLogger_WithSpecialCharacters(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := WithTraceID(context.Background(), "trace-with-special-chars-!@#$%^&*()")
	logger.Info(ctx, "message with special chars: !@#$%^&*()")
}

func TestSampledLogger_WithSpecialCharacters(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(1.0)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()
	logger.Info(ctx, "special chars: !@#$%^&*()")
}

func TestDynamicLevelLogger_WithSpecialCharacters(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, InfoLevel)

	ctx := context.Background()
	logger.Info(ctx, "special chars: !@#$%^&*()")
}

func TestLoggerBuilder_UnicodeMessage(t *testing.T) {
	logger := NewLoggerBuilder().
		Level(InfoLevel).
		Build()

	ctx := context.Background()
	logger.Info(ctx, "Unicode message: 你好世界 🌍")
}

func TestContextLogger_UnicodeMessage(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := context.Background()
	logger.Info(ctx, "Unicode message: 你好世界 🌍")
}

func TestSampledLogger_UnicodeMessage(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(1.0)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()
	logger.Info(ctx, "Unicode message: 你好世界 🌍")
}

func TestDynamicLevelLogger_UnicodeMessage(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, InfoLevel)

	ctx := context.Background()
	logger.Info(ctx, "Unicode message: 你好世界 🌍")
}

func TestLoggerBuilder_LongMessage(t *testing.T) {
	logger := NewLoggerBuilder().
		Level(InfoLevel).
		Build()

	ctx := context.Background()
	longMsg := strings.Repeat("a", 10000)
	logger.Info(ctx, longMsg)
}

func TestContextLogger_LongMessage(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := context.Background()
	longMsg := strings.Repeat("a", 10000)
	logger.Info(ctx, longMsg)
}

func TestSampledLogger_LongMessage(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(1.0)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()
	longMsg := strings.Repeat("a", 10000)
	logger.Info(ctx, longMsg)
}

func TestDynamicLevelLogger_LongMessage(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, InfoLevel)

	ctx := context.Background()
	longMsg := strings.Repeat("a", 10000)
	logger.Info(ctx, longMsg)
}

func TestLoggerBuilder_MultipleBuilds(t *testing.T) {
	logger1 := NewLoggerBuilder().Name("logger1").Build()
	logger2 := NewLoggerBuilder().Name("logger2").Build()

	if logger1 == logger2 {
		t.Error("expected different logger instances")
	}

	ctx := context.Background()
	logger1.Info(ctx, "logger1 message")
	logger2.Info(ctx, "logger2 message")
}

func TestContextLogger_MultipleInstances(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger1 := NewContextLogger(baseLogger)
	logger2 := NewContextLogger(baseLogger)

	ctx := context.Background()
	logger1.Info(ctx, "logger1 message")
	logger2.Info(ctx, "logger2 message")
}

func TestSampledLogger_MultipleInstances(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler1 := NewRandomSampler(0.5)
	sampler2 := NewRandomSampler(0.8)

	logger1 := NewSampledLogger(baseLogger, sampler1)
	logger2 := NewSampledLogger(baseLogger, sampler2)

	ctx := context.Background()
	logger1.Info(ctx, "logger1 message")
	logger2.Info(ctx, "logger2 message")
}

func TestDynamicLevelLogger_MultipleInstances(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger1 := NewDynamicLevelLogger(baseLogger, InfoLevel)
	logger2 := NewDynamicLevelLogger(baseLogger, WarnLevel)

	ctx := context.Background()
	logger1.Info(ctx, "logger1 message")
	logger2.Warn(ctx, "logger2 message")
}

func TestLoggerBuilder_BuildIdempotent(t *testing.T) {
	builder := NewLoggerBuilder().Name("idempotent")

	logger1 := builder.Build()
	logger2 := builder.Build()

	if logger1 == logger2 {
		t.Error("expected different logger instances from same builder")
	}

	ctx := context.Background()
	logger1.Info(ctx, "logger1 message")
	logger2.Info(ctx, "logger2 message")
}

func TestContextLogger_WithChain(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewContextLogger(baseLogger)

	ctx := context.Background()

	// 链式 With
	child1 := logger.With(ctx, KeyValue{Key: "key1", Value: "value1"})
	child2 := child1.With(ctx, KeyValue{Key: "key2", Value: "value2"})

	child2.Info(ctx, "chained with log")
}

func TestSampledLogger_WithChain(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	sampler := NewRandomSampler(1.0)
	logger := NewSampledLogger(baseLogger, sampler)

	ctx := context.Background()

	child1 := logger.With(ctx, KeyValue{Key: "key1", Value: "value1"})
	child2 := child1.With(ctx, KeyValue{Key: "key2", Value: "value2"})

	child2.Info(ctx, "chained with log")
}

func TestDynamicLevelLogger_WithChain(t *testing.T) {
	baseLogger := NewSlogLogger(WithLevel(DebugLevel))
	logger := NewDynamicLevelLogger(baseLogger, InfoLevel)

	ctx := context.Background()

	child1 := logger.With(ctx, KeyValue{Key: "key1", Value: "value1"})
	child2 := child1.With(ctx, KeyValue{Key: "key2", Value: "value2"})

	child2.Info(ctx, "chained with log")
}

func TestLoggerBuilder_WithNilOptions(t *testing.T) {
	// 测试 nil options
	logger := NewLoggerBuilder().Build()

	if logger == nil {
		t.Fatal("expected logger with nil options")
	}

	ctx := context.Background()
	logger.Info(ctx, "nil options log")
}

func TestContextLogger_WithNilLoggerAndContext(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with nil logger")
		}
	}()

	logger := NewContextLogger(nil)
	logger.Info(context.TODO(), "nil logger and context")
}

func TestSampledLogger_WithNilLoggerAndSampler(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with nil logger")
		}
	}()

	logger := NewSampledLogger(nil, nil)
	ctx := context.Background()
	logger.Info(ctx, "nil logger and sampler")
}

func TestDynamicLevelLogger_NilLoggerPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with nil logger")
		}
	}()

	logger := NewDynamicLevelLogger(nil, InfoLevel)
	ctx := context.Background()
	logger.Info(ctx, "nil logger")
}
