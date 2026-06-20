package exception

import (
	"context"
	"errors"
	"testing"
)

func TestDefaultExceptionResolver_Supports(t *testing.T) {
	resolver := NewDefaultExceptionResolver()

	if !resolver.Supports(errors.New("any error")) {
		t.Error("Default resolver should support all errors")
	}
}

func TestDefaultExceptionResolver_Resolve(t *testing.T) {
	resolver := NewDefaultExceptionResolver()
	resp := resolver.Resolve(context.Background(), errors.New("any error"))

	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	if resp.Code != 500 {
		t.Errorf("Expected code 500, got %d", resp.Code)
	}
	if resp.Message != "Internal Server Error" {
		t.Errorf("Expected message 'Internal Server Error', got %s", resp.Message)
	}
}

func TestBuiltinExceptionResolver_Supports(t *testing.T) {
	resolver := NewBuiltinExceptionResolver()

	testCases := []struct {
		err      error
		expected bool
	}{
		{ErrNotFound, true},
		{ErrBadRequest, true},
		{ErrUnauthorized, true},
		{ErrForbidden, true},
		{ErrConflict, true},
		{ErrInternalServer, true},
		{errors.New("other"), false},
	}

	for _, tc := range testCases {
		if got := resolver.Supports(tc.err); got != tc.expected {
			t.Errorf("Supports(%v) = %v, expected %v", tc.err, got, tc.expected)
		}
	}
}

func TestBuiltinExceptionResolver_Resolve(t *testing.T) {
	resolver := NewBuiltinExceptionResolver()

	testCases := []struct {
		err     error
		code    int
		message string
	}{
		{ErrNotFound, 404, "Resource not found"},
		{ErrBadRequest, 400, "Bad request"},
		{ErrUnauthorized, 401, "Unauthorized"},
		{ErrForbidden, 403, "Forbidden"},
		{ErrConflict, 409, "Conflict"},
		{ErrInternalServer, 500, "Internal server error"},
	}

	for _, tc := range testCases {
		resp := resolver.Resolve(context.Background(), tc.err)
		if resp.Code != tc.code {
			t.Errorf("Resolve(%v).Code = %d, expected %d", tc.err, resp.Code, tc.code)
		}
		if resp.Message != tc.message {
			t.Errorf("Resolve(%v).Message = %s, expected %s", tc.err, resp.Message, tc.message)
		}
	}
}
