package tracing

import (
	"context"
	"errors"
	"testing"
)

func TestTracerBuilder_Defaults(t *testing.T) {
	builder := NewTracerBuilder()

	if builder.sampleRate != 1.0 {
		t.Errorf("expected default sampleRate 1.0, got %f", builder.sampleRate)
	}

	if builder.headers == nil {
		t.Error("expected non-nil headers")
	}
}

func TestTracerBuilder_ChainConfig(t *testing.T) {
	tracer, err := NewTracerBuilder().
		ServiceName("test-service").
		ServiceVersion("1.0.0").
		Environment("prod").
		SampleRate(0.5).
		ExporterType("otlp").
		Endpoint("http://localhost:4318").
		Header("Authorization", "Bearer token").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tracer == nil {
		t.Fatal("expected non-nil tracer")
	}
}

func TestTracerBuilder_MissingServiceName(t *testing.T) {
	_, err := NewTracerBuilder().Build()
	if err == nil {
		t.Error("expected error for missing service name")
	}
}

func TestTracerBuilder_WithProvider(t *testing.T) {
	provider := &NoopTracerProvider{}

	tracer, err := NewTracerBuilder().
		Provider(provider).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tracer == nil {
		t.Fatal("expected non-nil tracer")
	}
}

func TestTracerBuilder_MustBuild_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for missing service name")
		}
	}()

	NewTracerBuilder().MustBuild()
}

func TestTracerBuilder_BuildAndRegister(t *testing.T) {
	provider := &NoopTracerProvider{}

	tracer, err := NewTracerBuilder().
		ServiceName("test").
		Provider(provider).
		BuildAndRegister()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tracer == nil {
		t.Fatal("expected non-nil tracer")
	}

	// 验证全局提供者已设置
	globalProvider := GetTracerProvider()
	if globalProvider == nil {
		t.Error("expected global provider to be set")
	}
}

func TestSpanBuilder_BasicSpan(t *testing.T) {
	tracer := &NoopTracer{}

	ctx, span := NewSpanBuilder(tracer, "test-operation").
		Attribute("key1", "value1").
		Attribute("key2", 123).
		Kind(SpanKindClient).
		Start()

	if ctx == nil {
		t.Fatal("expected non-nil context")
	}

	if span == nil {
		t.Fatal("expected non-nil span")
	}

	// 验证Span已设置到上下文
	spanFromCtx := SpanFromContext(ctx)
	if spanFromCtx == nil {
		t.Error("expected span to be set in context")
	}
}

func TestSpanBuilder_WithContext(t *testing.T) {
	tracer := &NoopTracer{}
	type testKey struct{}
	parentCtx := context.WithValue(context.Background(), testKey{}, "test-value")

	ctx, span := NewSpanBuilder(tracer, "test-operation").
		Context(parentCtx).
		Start()

	if ctx == nil {
		t.Fatal("expected non-nil context")
	}

	if span == nil {
		t.Fatal("expected non-nil span")
	}

	// 验证父上下文的值仍然存在
	if ctx.Value(testKey{}) != "test-value" {
		t.Error("expected parent context value to be preserved")
	}
}

func TestSpanBuilder_DefaultKind(t *testing.T) {
	tracer := &NoopTracer{}

	_, span := NewSpanBuilder(tracer, "test-operation").Start()

	if span == nil {
		t.Fatal("expected non-nil span")
	}
}

func TestTraceHelper_Trace_Success(t *testing.T) {
	tracer := &NoopTracer{}
	helper := NewTraceHelper(tracer)

	called := false
	err := helper.Trace(context.Background(), "test-operation", func(ctx context.Context) error {
		called = true
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !called {
		t.Error("expected function to be called")
	}
}

func TestTraceHelper_Trace_Error(t *testing.T) {
	tracer := &NoopTracer{}
	helper := NewTraceHelper(tracer)

	expectedErr := errors.New("test error")

	err := helper.Trace(context.Background(), "test-operation", func(ctx context.Context) error {
		return expectedErr
	})

	if err == nil {
		t.Fatal("expected error")
	}

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestTraceHelper_TraceWithCallback_Success(t *testing.T) {
	tracer := &NoopTracer{}
	helper := NewTraceHelper(tracer)

	result := helper.TraceWithCallback(context.Background(), "test-operation", func(ctx context.Context) (any, error) {
		return "success", nil
	})

	if result.Err != nil {
		t.Errorf("unexpected error: %v", result.Err)
	}

	if result.Value != "success" {
		t.Errorf("expected result 'success', got %v", result.Value)
	}
}

func TestTraceHelper_TraceWithCallback_Error(t *testing.T) {
	tracer := &NoopTracer{}
	helper := NewTraceHelper(tracer)

	expectedErr := errors.New("test error")

	result := helper.TraceWithCallback(context.Background(), "test-operation", func(ctx context.Context) (any, error) {
		return "", expectedErr
	})

	if result.Err == nil {
		t.Fatal("expected error")
	}

	if result.Err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, result.Err)
	}

	if result.Value != "" {
		t.Errorf("expected empty result, got %v", result.Value)
	}
}

func TestTracerBuilder_SampleRate_Invalid(t *testing.T) {
	// 测试负数
	builder := NewTracerBuilder().SampleRate(-0.5)
	if builder.sampleRate != 1.0 {
		t.Errorf("expected sampleRate to remain 1.0 for negative value, got %f", builder.sampleRate)
	}

	// 测试大于1
	builder = NewTracerBuilder().SampleRate(1.5)
	if builder.sampleRate != 1.0 {
		t.Errorf("expected sampleRate to remain 1.0 for value > 1, got %f", builder.sampleRate)
	}
}

func TestTracerBuilder_MultipleHeaders(t *testing.T) {
	tracer, err := NewTracerBuilder().
		ServiceName("test").
		Header("Header1", "Value1").
		Header("Header2", "Value2").
		Header("Header3", "Value3").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tracer == nil {
		t.Fatal("expected non-nil tracer")
	}

	if len(NewTracerBuilder().Header("k", "v").headers) != 1 {
		t.Error("expected headers to be added")
	}
}
