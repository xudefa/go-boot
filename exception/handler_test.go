package exception

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

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

func TestDefaultExceptionHandler_Handle(t *testing.T) {
	handler := NewDefaultExceptionHandler()
	response := &mockResponseWriter{}

	resp := handler.Handle(context.Background(), ErrNotFound, response)

	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	if resp.Code != 404 {
		t.Errorf("Expected code 404, got %d", resp.Code)
	}
}

func TestDefaultExceptionHandler_RegisterResolver(t *testing.T) {
	handler := NewDefaultExceptionHandler()

	resolver := &mockResolver{
		order:    50,
		supports: func(err error) bool { return errors.Is(err, ErrBadRequest) },
		resolve: func(ctx context.Context, err error) *ErrorResponse {
			return NewErrorResponse(400, "Custom bad request", "", "", nil)
		},
	}

	handler.RegisterResolver(resolver)

	response := &mockResponseWriter{}
	resp := handler.Handle(context.Background(), ErrBadRequest, response)

	if resp.Message != "Custom bad request" {
		t.Errorf("Expected message 'Custom bad request', got %s", resp.Message)
	}
}

func TestDefaultExceptionHandler_RegisterHandlerFunc(t *testing.T) {
	handler := NewDefaultExceptionHandler()

	err := errors.New("test error")
	errType := reflect.TypeOf(err)
	handler.RegisterHandlerFunc(errType,
		func(ctx context.Context, err error) *ErrorResponse {
			return NewErrorResponse(418, "I'm a teapot", "", "", nil)
		})

	response := &mockResponseWriter{}
	resp := handler.Handle(context.Background(), err, response)

	if resp.Code != 418 {
		t.Errorf("Expected code 418, got %d", resp.Code)
	}
}

func TestDefaultExceptionHandler_UnknownError(t *testing.T) {
	handler := NewDefaultExceptionHandler()
	response := &mockResponseWriter{}

	resp := handler.Handle(context.Background(), errors.New("unknown error"), response)

	if resp.Code != 500 {
		t.Errorf("Expected code 500, got %d", resp.Code)
	}
	if resp.Message != "Internal Server Error" {
		t.Errorf("Expected message 'Internal Server Error', got %s", resp.Message)
	}
}
