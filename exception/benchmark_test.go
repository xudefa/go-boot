package exception

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkDefaultExceptionHandler_Handle(b *testing.B) {
	handler := NewDefaultExceptionHandler()
	response := &mockResponseWriter{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.Handle(context.Background(), ErrNotFound, response)
	}
}

func BenchmarkDefaultExceptionHandler_HandleWithLogger(b *testing.B) {
	logger := NewDefaultLogger()
	handler := NewDefaultExceptionHandler(WithLogger(logger))
	response := &mockResponseWriter{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.Handle(context.Background(), ErrNotFound, response)
	}
}

func BenchmarkDefaultExceptionHandler_HandleWithMetrics(b *testing.B) {
	metrics := NewDefaultMetricsRecorder()
	handler := NewDefaultExceptionHandler(WithMetricsRecorder(metrics))
	response := &mockResponseWriter{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.Handle(context.Background(), ErrNotFound, response)
	}
}

func BenchmarkResolverChain_Resolve(b *testing.B) {
	chain := NewResolverChain()
	chain.AddResolver(NewBuiltinExceptionResolver())
	chain.AddResolver(NewDefaultExceptionResolver())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chain.Resolve(context.Background(), ErrNotFound)
	}
}

func BenchmarkExceptionHandlingMiddleware(b *testing.B) {
	handler := NewDefaultExceptionHandler()
	middleware := ExceptionHandlingMiddleware(handler)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	wrappedHandler := middleware(nextHandler)
	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)
	}
}

func BenchmarkDefaultLogger_Error(b *testing.B) {
	logger := NewDefaultLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Error(context.Background(), "test error",
			KeyValue{Key: "key1", Value: "value1"},
			KeyValue{Key: "key2", Value: 123},
		)
	}
}

func BenchmarkDefaultMetricsRecorder_RecordException(b *testing.B) {
	metrics := NewDefaultMetricsRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.RecordException("TestError", 500)
	}
}

func BenchmarkAccessDeniedHandlerAdapter_Handle(b *testing.B) {
	handler := NewDefaultExceptionHandler()
	adapter := NewAccessDeniedHandlerAdapter(handler)
	response := &mockResponseWriter{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.Handle(response, ErrForbidden)
	}
}

func BenchmarkNewErrorResponse(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewErrorResponse(500, "Internal Server Error", "req-123", "trace-456", nil)
	}
}

func BenchmarkBuiltinExceptionResolver_Supports(b *testing.B) {
	resolver := NewBuiltinExceptionResolver()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resolver.Supports(ErrNotFound)
	}
}

func BenchmarkBuiltinExceptionResolver_Resolve(b *testing.B) {
	resolver := NewBuiltinExceptionResolver()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resolver.Resolve(context.Background(), ErrNotFound)
	}
}

func BenchmarkDefaultExceptionHandler_UnknownError(b *testing.B) {
	handler := NewDefaultExceptionHandler()
	response := &mockResponseWriter{}
	err := errors.New("unknown error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.Handle(context.Background(), err, response)
	}
}
