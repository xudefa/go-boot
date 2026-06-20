package exception

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type CustomError struct {
	Message string
}

func (e *CustomError) Error() string {
	return e.Message
}

func TestIntegration_FullExceptionHandling(t *testing.T) {
	logger := NewDefaultLogger()
	metrics := NewDefaultMetricsRecorder()
	handler := NewDefaultExceptionHandler(
		WithLogger(logger),
		WithMetricsRecorder(metrics),
	)

	response := &mockResponseWriter{}
	resp := handler.Handle(context.Background(), ErrNotFound, response)

	if resp.Code != 404 {
		t.Errorf("Expected code 404, got %d", resp.Code)
	}

	count := metrics.(*DefaultMetricsRecorder).GetCount("*errors.errorString", 404)
	if count == 0 {
		t.Error("Expected metrics to be recorded")
	}
}

func TestIntegration_HTTPMiddleware(t *testing.T) {
	handler := NewDefaultExceptionHandler()
	middleware := ExceptionHandlingMiddleware(handler)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware(nextHandler).ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Code != 500 {
		t.Errorf("Expected code 500, got %d", resp.Code)
	}
}

func TestIntegration_CustomResolver(t *testing.T) {
	handler := NewDefaultExceptionHandler()

	customResolver := &mockResolver{
		order: 50,
		supports: func(err error) bool {
			_, ok := err.(*CustomError)
			return ok
		},
		resolve: func(ctx context.Context, err error) *ErrorResponse {
			return NewErrorResponse(418, "Custom error handled", "", "", nil)
		},
	}

	handler.RegisterResolver(customResolver)

	response := &mockResponseWriter{}
	resp := handler.Handle(context.Background(), &CustomError{Message: "test"}, response)

	if resp.Code != 418 {
		t.Errorf("Expected code 418, got %d", resp.Code)
	}
	if resp.Message != "Custom error handled" {
		t.Errorf("Expected message 'Custom error handled', got %s", resp.Message)
	}
}

func TestIntegration_SecurityAdapter(t *testing.T) {
	handler := NewDefaultExceptionHandler()
	adapter := NewAccessDeniedHandlerAdapter(handler)

	response := &mockResponseWriter{}
	adapter.Handle(response, ErrForbidden)

	if response.statusCode != 403 {
		t.Errorf("Expected status 403, got %d", response.statusCode)
	}

	if response.headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type header, got %v", response.headers)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(response.body, &resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Code != 403 {
		t.Errorf("Expected code 403, got %d", resp.Code)
	}
}

func TestIntegration_HandlerFuncRegistration(t *testing.T) {
	handler := NewDefaultExceptionHandler()

	err := errors.New("specific error")
	errType := reflect.TypeOf(err)

	handler.RegisterHandlerFunc(errType,
		func(ctx context.Context, err error) *ErrorResponse {
			return NewErrorResponse(422, "Unprocessable Entity", "", "", nil)
		})

	response := &mockResponseWriter{}
	resp := handler.Handle(context.Background(), err, response)

	if resp.Code != 422 {
		t.Errorf("Expected code 422, got %d", resp.Code)
	}
}

func TestIntegration_ResolverChainPriority(t *testing.T) {
	handler := NewDefaultExceptionHandler()

	resolver1 := &mockResolver{
		order:    10,
		supports: func(err error) bool { return errors.Is(err, ErrNotFound) },
		resolve: func(ctx context.Context, err error) *ErrorResponse {
			return NewErrorResponse(404, "Priority 10", "", "", nil)
		},
	}

	resolver2 := &mockResolver{
		order:    5,
		supports: func(err error) bool { return errors.Is(err, ErrNotFound) },
		resolve: func(ctx context.Context, err error) *ErrorResponse {
			return NewErrorResponse(404, "Priority 5", "", "", nil)
		},
	}

	handler.RegisterResolver(resolver1)
	handler.RegisterResolver(resolver2)

	response := &mockResponseWriter{}
	resp := handler.Handle(context.Background(), ErrNotFound, response)

	if resp.Message != "Priority 5" {
		t.Errorf("Expected message 'Priority 5', got %s", resp.Message)
	}
}
