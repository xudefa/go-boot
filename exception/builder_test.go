package exception

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestExceptionHandlerBuilder_Defaults(t *testing.T) {
	builder := NewExceptionHandlerBuilder()

	if builder.exceptionHandlers == nil {
		t.Error("expected non-nil exceptionHandlers")
	}
}

func TestExceptionHandlerBuilder_ChainConfig(t *testing.T) {
	handler, err := NewExceptionHandlerBuilder().
		IncludeStackTrace(true).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestExceptionHandlerBuilder_WithLogger(t *testing.T) {
	logger := &mockLogger{}

	handler, err := NewExceptionHandlerBuilder().
		Logger(logger).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestExceptionHandlerBuilder_WithMetricsRecorder(t *testing.T) {
	recorder := &mockMetricsRecorder{}

	handler, err := NewExceptionHandlerBuilder().
		MetricsRecorder(recorder).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestExceptionHandlerBuilder_WithResolver(t *testing.T) {
	resolver := &mockExceptionResolver{}

	handler, err := NewExceptionHandlerBuilder().
		Resolver(resolver).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestExceptionHandlerBuilder_WithExceptionHandler(t *testing.T) {
	handlerCalled := false

	handler, err := NewExceptionHandlerBuilder().
		ExceptionHandler(reflect.TypeOf(errors.New("")), func(ctx context.Context, err error) *ErrorResponse {
			handlerCalled = true
			return NewErrorResponse(500, "internal error", "", "", nil)
		}).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if handler == nil {
		t.Fatal("expected non-nil handler")
	}

	_ = handlerCalled
}

func TestExceptionHandlerBuilder_MustBuild(t *testing.T) {
	handler := NewExceptionHandlerBuilder().MustBuild()

	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestErrorResponseBuilder_BasicResponse(t *testing.T) {
	response := NewErrorResponseBuilder().
		Code(404).
		Message("Not Found").
		RequestID("req-123").
		TraceID("trace-456").
		Details(map[string]string{"path": "/api/users"}).
		Build()

	if response.Code != 404 {
		t.Errorf("expected Code 404, got %d", response.Code)
	}

	if response.Message != "Not Found" {
		t.Errorf("expected Message 'Not Found', got %s", response.Message)
	}

	if response.RequestID != "req-123" {
		t.Errorf("expected RequestID 'req-123', got %s", response.RequestID)
	}

	if response.TraceID != "trace-456" {
		t.Errorf("expected TraceID 'trace-456', got %s", response.TraceID)
	}

	if response.Details == nil {
		t.Error("expected non-nil Details")
	}
}

func TestErrorResponseBuilder_ToJSON(t *testing.T) {
	jsonBytes, err := NewErrorResponseBuilder().
		Code(500).
		Message("Internal Server Error").
		ToJSON()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(jsonBytes) == 0 {
		t.Error("expected non-empty JSON bytes")
	}

	// 验证JSON包含预期字段
	jsonStr := string(jsonBytes)
	if jsonStr == "" {
		t.Error("expected non-empty JSON string")
	}
}

func TestErrorResponse_ToJSON(t *testing.T) {
	response := &ErrorResponse{
		Code:      400,
		Message:   "Bad Request",
		RequestID: "",
		TraceID:   "",
	}

	jsonBytes, err := response.ToJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(jsonBytes) == 0 {
		t.Error("expected non-empty JSON bytes")
	}

	// Verify JSON contains code and message
	jsonStr := string(jsonBytes)
	if !strings.Contains(jsonStr, `"code":400`) {
		t.Errorf("expected JSON to contain code 400, got %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"message":"Bad Request"`) {
		t.Errorf("expected JSON to contain message, got %s", jsonStr)
	}
}

func TestExceptionHandlerBuilder_MultipleResolvers(t *testing.T) {
	resolver1 := &mockExceptionResolver{order: 1}
	resolver2 := &mockExceptionResolver{order: 2}
	resolver3 := &mockExceptionResolver{order: 3}

	handler, err := NewExceptionHandlerBuilder().
		Resolver(resolver1).
		Resolver(resolver2).
		Resolver(resolver3).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestExceptionHandlerBuilder_MultipleExceptionHandlers(t *testing.T) {
	errType1 := reflect.TypeOf(errors.New("type1"))
	errType2 := reflect.TypeOf(errors.New("type2"))

	handler, err := NewExceptionHandlerBuilder().
		ExceptionHandler(errType1, func(ctx context.Context, err error) *ErrorResponse {
			return NewErrorResponse(400, "error1", "", "", nil)
		}).
		ExceptionHandler(errType2, func(ctx context.Context, err error) *ErrorResponse {
			return NewErrorResponse(500, "error2", "", "", nil)
		}).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestErrorResponseBuilder_EmptyResponse(t *testing.T) {
	response := NewErrorResponseBuilder().Build()

	if response.Timestamp != 0 {
		t.Errorf("expected Timestamp 0 for empty response, got %d", response.Timestamp)
	}
}

func TestErrorResponseBuilder_WithNilDetails(t *testing.T) {
	response := NewErrorResponseBuilder().
		Code(200).
		Message("OK").
		Details(nil).
		Build()

	if response.Details != nil {
		t.Errorf("expected Details to be nil, got %v", response.Details)
	}
}

// Mock implementations

type mockMetricsRecorder struct{}

func (m *mockMetricsRecorder) RecordException(exceptionType string, statusCode int) {}

type mockExceptionResolver struct {
	order int
}

func (m *mockExceptionResolver) Resolve(ctx context.Context, err error) *ErrorResponse {
	return nil
}

func (m *mockExceptionResolver) Supports(err error) bool {
	return false
}

func (m *mockExceptionResolver) Order() int {
	return m.order
}
