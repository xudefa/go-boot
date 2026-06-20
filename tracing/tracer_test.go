package tracing

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestNoopTracer_Start(t *testing.T) {
	tracer := &NoopTracer{}
	ctx := context.Background()

	newCtx, span := tracer.Start(ctx, "test-operation")

	if newCtx != ctx {
		t.Error("expected context to be returned unchanged")
	}
	if span == nil {
		t.Fatal("expected non-nil span")
	}

	_, ok := span.(*NoopSpan)
	if !ok {
		t.Error("expected span to be *NoopSpan")
	}
}

func TestNoopSpan_End_NoPanic(t *testing.T) {
	span := &NoopSpan{}
	span.End()
}

func TestNoopSpan_AddEvent_NoPanic(t *testing.T) {
	span := &NoopSpan{}
	span.AddEvent("test-event")
	span.AddEvent("test-event", WithEventAttribute("key", "value"))
}

func TestNoopSpan_SetAttribute_NoPanic(t *testing.T) {
	span := &NoopSpan{}
	span.SetAttribute("key", "value")
}

func TestNoopSpan_RecordError_NoPanic(t *testing.T) {
	span := &NoopSpan{}
	span.RecordError(errors.New("test error"))
}

func TestNoopSpan_SetError_NoPanic(t *testing.T) {
	span := &NoopSpan{}
	span.SetError(errors.New("test error"))
}

func TestNoopSpan_SetStatus_NoPanic(t *testing.T) {
	span := &NoopSpan{}
	span.SetStatus(SpanStatusOK)
	span.SetStatus(SpanStatusError)
	span.SetStatus(SpanStatusCanceled)
}

func TestSpanContext(t *testing.T) {
	sc := SpanContext{}
	if sc.TraceID != "" {
		t.Errorf("expected empty TraceID, got %q", sc.TraceID)
	}
	if sc.SpanID != "" {
		t.Errorf("expected empty SpanID, got %q", sc.SpanID)
	}
}

func TestNoopSpan_SpanContext_Defaults(t *testing.T) {
	span := &NoopSpan{}
	sc := span.SpanContext()

	if sc.TraceID != "" {
		t.Errorf("expected empty TraceID, got %q", sc.TraceID)
	}
	if sc.SpanID != "" {
		t.Errorf("expected empty SpanID, got %q", sc.SpanID)
	}
}

func TestNoopSpan_GetTraceID(t *testing.T) {
	span := &NoopSpan{}
	if id := span.GetTraceID(); id != "" {
		t.Errorf("expected empty TraceID, got %q", id)
	}
}

func TestNoopSpan_GetSpanID(t *testing.T) {
	span := &NoopSpan{}
	if id := span.GetSpanID(); id != "" {
		t.Errorf("expected empty SpanID, got %q", id)
	}
}

func TestNoopTracer_CurrentSpan(t *testing.T) {
	tracer := &NoopTracer{}
	ctx := context.Background()

	span := tracer.CurrentSpan(ctx)
	if span == nil {
		t.Fatal("expected non-nil span")
	}

	_, ok := span.(*NoopSpan)
	if !ok {
		t.Error("expected span to be *NoopSpan")
	}
}

func TestNoopTracer_Finish_NoPanic(t *testing.T) {
	tracer := &NoopTracer{}
	tracer.Finish()
}

func TestSpanStatusCode_Values(t *testing.T) {
	if SpanStatusUnset != 0 {
		t.Errorf("SpanStatusUnset should be 0, got %d", SpanStatusUnset)
	}
	if SpanStatusOK != 1 {
		t.Errorf("SpanStatusOK should be 1, got %d", SpanStatusOK)
	}
	if SpanStatusError != 2 {
		t.Errorf("SpanStatusError should be 2, got %d", SpanStatusError)
	}
	if SpanStatusCanceled != 3 {
		t.Errorf("SpanStatusCanceled should be 3, got %d", SpanStatusCanceled)
	}
}

func TestStartWithOptions(t *testing.T) {
	tracer := &NoopTracer{}
	ctx := context.Background()

	_, span := tracer.Start(ctx, "op", WithAttribute("key", "val"))
	if span == nil {
		t.Fatal("expected non-nil span")
	}
}

func TestLocalTracer_Start(t *testing.T) {
	tracer := NewLocalTracer("test-tracer")
	ctx := context.Background()

	newCtx, span := tracer.Start(ctx, "test-operation", WithAttribute("key", "value"))

	if newCtx == nil {
		t.Fatal("expected non-nil context")
	}
	if span == nil {
		t.Fatal("expected non-nil span")
	}

	traceID := span.GetTraceID()
	if traceID == "" {
		t.Error("expected non-empty TraceID")
	}

	spanID := span.GetSpanID()
	if spanID == "" {
		t.Error("expected non-empty SpanID")
	}
}

func TestLocalTracer_CurrentSpan(t *testing.T) {
	tracer := NewLocalTracer("test-tracer")
	ctx := context.Background()

	newCtx, span := tracer.Start(ctx, "test-operation")
	retrievedSpan := tracer.CurrentSpan(newCtx)

	if retrievedSpan.GetTraceID() != span.GetTraceID() {
		t.Error("expected matching TraceID")
	}
}

func TestLocalSpan_SetAttribute(t *testing.T) {
	tracer := NewLocalTracer("test-tracer")
	ctx := context.Background()

	_, span := tracer.Start(ctx, "test-operation")
	span.SetAttribute("test-key", "test-value")

	attrs := span.GetAttributes()
	if attrs["test-key"] != "test-value" {
		t.Errorf("expected attribute test-key=test-value, got %v", attrs)
	}
}

func TestLocalSpan_SetError(t *testing.T) {
	tracer := NewLocalTracer("test-tracer")
	ctx := context.Background()

	_, span := tracer.Start(ctx, "test-operation")
	span.SetError(errors.New("test error"))

	attrs := span.GetAttributes()
	if attrs["error"] != "test error" {
		t.Errorf("expected error attribute, got %v", attrs)
	}
}

func TestLocalSpan_SpanContext(t *testing.T) {
	tracer := NewLocalTracer("test-tracer")
	ctx := context.Background()

	_, span := tracer.Start(ctx, "test-operation")
	sc := span.SpanContext()

	if sc.TraceID == "" {
		t.Error("expected non-empty TraceID")
	}
	if sc.SpanID == "" {
		t.Error("expected non-empty SpanID")
	}
	if sc.TraceFlags != 0x01 {
		t.Errorf("expected TraceFlags 0x01, got %d", sc.TraceFlags)
	}
}

func TestNoopTracerProvider(t *testing.T) {
	provider := &NoopTracerProvider{}
	tracer := provider.Tracer("test")

	if tracer == nil {
		t.Fatal("expected non-nil tracer")
	}

	err := provider.Shutdown(context.Background())
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestTracerProviderImpl(t *testing.T) {
	provider := &TracerProviderImpl{}
	tracer := provider.Tracer("test")

	if tracer == nil {
		t.Fatal("expected non-nil tracer")
	}

	err := provider.Shutdown(context.Background())
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestHTTPHeadersCarrier(t *testing.T) {
	headers := http.Header{}
	carrier := NewHTTPHeadersCarrier(headers)

	carrier.Set("Trace-Parent", "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01")
	if val := carrier.Get("Trace-Parent"); val != "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01" {
		t.Errorf("expected Trace-Parent value, got %q", val)
	}

	keys := carrier.Keys()
	if len(keys) != 1 {
		t.Errorf("expected 1 key, got %d", len(keys))
	}
}

func TestNoopPropagator(t *testing.T) {
	propagator := &NoopPropagator{}
	ctx := context.Background()
	carrier := NewHTTPHeadersCarrier(http.Header{})

	resultCtx := propagator.Extract(ctx, carrier)
	if resultCtx != ctx {
		t.Error("expected context to be returned unchanged")
	}

	propagator.Inject(ctx, carrier)
}

func TestHTTPStatusToSpanStatusCode(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   SpanStatusCode
	}{
		{http.StatusOK, SpanStatusOK},
		{http.StatusNotFound, SpanStatusUnset},
		{http.StatusBadRequest, SpanStatusUnset},
		{http.StatusInternalServerError, SpanStatusError},
		{http.StatusServiceUnavailable, SpanStatusError},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.statusCode), func(t *testing.T) {
			result := HTTPStatusToSpanStatusCode(tt.statusCode)
			if result != tt.expected {
				t.Errorf("HTTPStatusToSpanStatusCode(%d) = %v, want %v", tt.statusCode, result, tt.expected)
			}
		})
	}
}

func TestHTTPTraceAttributes_SetAttributes(t *testing.T) {
	tracer := NewLocalTracer("test")
	ctx := context.Background()
	_, span := tracer.Start(ctx, "test")

	attrs := &HTTPTraceAttributes{
		Method:     "GET",
		Target:     "/api/test",
		Host:       "localhost:8080",
		StatusCode: http.StatusOK,
	}
	attrs.SetAttributes(span)

	spanAttrs := span.GetAttributes()
	if spanAttrs["http.method"] != "GET" {
		t.Errorf("expected http.method=GET, got %v", spanAttrs["http.method"])
	}
}

func TestStartHTTPServerSpan(t *testing.T) {
	tracer := NewLocalTracer("test")
	ctx := context.Background()

	newCtx, span := StartHTTPServerSpan(ctx, tracer, "GET /test", "GET", "/test", "localhost")
	if newCtx == nil {
		t.Fatal("expected non-nil context")
	}
	if span == nil {
		t.Fatal("expected non-nil span")
	}

	attrs := span.GetAttributes()
	if attrs["http.method"] != "GET" {
		t.Errorf("expected http.method=GET, got %v", attrs["http.method"])
	}
}

func TestStartHTTPClientSpan(t *testing.T) {
	tracer := NewLocalTracer("test")
	ctx := context.Background()

	newCtx, span := StartHTTPClientSpan(ctx, tracer, "GET /api", "GET", "/api", "api.example.com")
	if newCtx == nil {
		t.Fatal("expected non-nil context")
	}
	if span == nil {
		t.Fatal("expected non-nil span")
	}
}

func TestWithSpanKind(t *testing.T) {
	config := &SpanConfig{}
	WithSpanKind(SpanKindServer)(config)
	if config.Kind != SpanKindServer {
		t.Errorf("expected SpanKindServer, got %v", config.Kind)
	}
}

func TestSetAndGetTracerProvider(t *testing.T) {
	original := GetTracerProvider()

	customProvider := &NoopTracerProvider{}
	SetTracerProvider(customProvider)

	if GetTracerProvider() != customProvider {
		t.Error("expected custom provider")
	}

	SetTracerProvider(original)
}

func TestGetTracer(t *testing.T) {
	tracer := GetTracer("test")
	if tracer == nil {
		t.Fatal("expected non-nil tracer")
	}
}

func TestSetAndGetPropagator(t *testing.T) {
	original := GetPropagator()

	customPropagator := &NoopPropagator{}
	SetPropagator(customPropagator)

	if GetPropagator() != customPropagator {
		t.Error("expected custom propagator")
	}

	SetPropagator(original)
}

func TestExtractAndInjectTraceContext(t *testing.T) {
	ctx := context.Background()
	carrier := NewHTTPHeadersCarrier(http.Header{})

	ExtractTraceContext(ctx, carrier)
	InjectTraceContext(ctx, carrier)
}

func TestLocalSpan_End(t *testing.T) {
	tracer := NewLocalTracer("test")
	ctx := context.Background()
	_, span := tracer.Start(ctx, "test")
	span.End()
}

func TestLocalSpan_AddEvent(t *testing.T) {
	tracer := NewLocalTracer("test")
	ctx := context.Background()
	_, span := tracer.Start(ctx, "test")
	span.AddEvent("test-event")
}

func TestLocalSpan_RecordError(t *testing.T) {
	tracer := NewLocalTracer("test")
	ctx := context.Background()
	_, span := tracer.Start(ctx, "test")
	span.RecordError(errors.New("test error"))

	attrs := span.GetAttributes()
	if attrs["error"] != "test error" {
		t.Errorf("expected error attribute, got %v", attrs)
	}
}

func TestLocalSpan_SetStatus(t *testing.T) {
	tracer := NewLocalTracer("test")
	ctx := context.Background()
	_, span := tracer.Start(ctx, "test")
	span.SetStatus(SpanStatusError)
}

func TestNoopSpan_GetAttributes(t *testing.T) {
	span := &NoopSpan{}
	attrs := span.GetAttributes()
	if attrs != nil {
		t.Error("expected nil attributes")
	}
}
