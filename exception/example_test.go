package exception_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/xudefa/go-boot/exception"
)

func ExampleDefaultExceptionHandler() {
	handler := exception.NewDefaultExceptionHandler()

	response := &mockResponseWriter{}
	resp := handler.Handle(context.Background(), exception.ErrNotFound, response)

	fmt.Printf("Code: %d, Message: %s\n", resp.Code, resp.Message)

}

func ExampleDefaultExceptionHandler_withOptions() {
	logger := exception.NewDefaultLogger()
	metrics := exception.NewDefaultMetricsRecorder()

	handler := exception.NewDefaultExceptionHandler(
		exception.WithLogger(logger),
		exception.WithMetricsRecorder(metrics),
	)

	response := &mockResponseWriter{}
	resp := handler.Handle(context.Background(), exception.ErrBadRequest, response)

	fmt.Printf("Code: %d, Message: %s\n", resp.Code, resp.Message)

}

func ExampleExceptionHandlingMiddleware() {
	handler := exception.NewDefaultExceptionHandler()
	middleware := exception.ExceptionHandlingMiddleware(handler)

	httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/panic" {
			panic("something went wrong")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	wrappedHandler := middleware(httpHandler)

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	fmt.Printf("Status: %d\n", w.Code)

}

func ExampleAccessDeniedHandlerAdapter() {
	handler := exception.NewDefaultExceptionHandler()
	adapter := exception.NewAccessDeniedHandlerAdapter(handler)

	response := &mockResponseWriter{}
	adapter.Handle(response, exception.ErrForbidden)

	fmt.Printf("Status: %d\n", response.statusCode)

}

func ExampleAuthenticationEntryPointAdapter() {
	handler := exception.NewDefaultExceptionHandler()
	adapter := exception.NewAuthenticationEntryPointAdapter(handler)

	response := &mockResponseWriter{}
	adapter.Commence(response, exception.ErrUnauthorized)

	fmt.Printf("Status: %d\n", response.statusCode)

}

func ExampleDefaultLogger() {
	logger := exception.NewDefaultLogger()

	logger.Error(context.Background(), "test error",
		exception.KeyValue{Key: "key1", Value: "value1"},
		exception.KeyValue{Key: "key2", Value: 123},
	)

}

func ExampleDefaultMetricsRecorder() {
	metrics := exception.NewDefaultMetricsRecorder()

	metrics.RecordException("TestError", 500)
	metrics.RecordException("TestError", 500)

	count := metrics.(*exception.DefaultMetricsRecorder).GetCount("TestError", 500)
	fmt.Printf("Count: %d\n", count)

}

type mockResponseWriter struct {
	statusCode int
	headers    map[string]string
	body       []byte
}

func (m *mockResponseWriter) SetStatusCode(code int) {
	m.statusCode = code
}

func (m *mockResponseWriter) SetHeader(key, value string) {
	if m.headers == nil {
		m.headers = make(map[string]string)
	}
	m.headers[key] = value
}

func (m *mockResponseWriter) Write(data []byte) error {
	m.body = data
	return nil
}
