package exception

import (
	"testing"
)

func TestAccessDeniedHandlerAdapter_Handle(t *testing.T) {
	handler := NewDefaultExceptionHandler()
	adapter := NewAccessDeniedHandlerAdapter(handler)

	response := &mockResponseWriter{}
	err := ErrForbidden

	adapter.Handle(response, err)

	if response.statusCode != 403 {
		t.Errorf("Expected status 403, got %d", response.statusCode)
	}
}

func TestAuthenticationEntryPointAdapter_Commence(t *testing.T) {
	handler := NewDefaultExceptionHandler()
	adapter := NewAuthenticationEntryPointAdapter(handler)

	response := &mockResponseWriter{}
	err := ErrUnauthorized

	adapter.Commence(response, err)

	if response.statusCode != 401 {
		t.Errorf("Expected status 401, got %d", response.statusCode)
	}
}

func TestAccessDeniedHandlerAdapter_NilError(t *testing.T) {
	handler := NewDefaultExceptionHandler()
	adapter := NewAccessDeniedHandlerAdapter(handler)

	response := &mockResponseWriter{}
	adapter.Handle(response, nil)

	if response.statusCode != 0 {
		t.Errorf("Expected no status change, got %d", response.statusCode)
	}
}

func TestAuthenticationEntryPointAdapter_NilError(t *testing.T) {
	handler := NewDefaultExceptionHandler()
	adapter := NewAuthenticationEntryPointAdapter(handler)

	response := &mockResponseWriter{}
	adapter.Commence(response, nil)

	if response.statusCode != 0 {
		t.Errorf("Expected no status change, got %d", response.statusCode)
	}
}
