package errors

import (
	"errors"
	"testing"
)

func TestErrorBuilder_Build_WithCause(t *testing.T) {
	cause := errors.New("original error")
	err := NewError("TEST_001").
		Message("test error").
		Cause(cause).
		Context("key", "value").
		Build()

	structuredErr := err.(*StructuredError)

	if structuredErr.Code != "TEST_001" {
		t.Errorf("expected code TEST_001, got %s", structuredErr.Code)
	}
	if structuredErr.Message != "test error" {
		t.Errorf("expected message 'test error', got %s", structuredErr.Message)
	}
	if !errors.Is(structuredErr.Cause, cause) {
		t.Error("expected cause to match original error")
	}
	if structuredErr.Context["key"] != "value" {
		t.Error("expected context key to be 'value'")
	}
}

func TestErrorBuilder_Build_WithoutCause(t *testing.T) {
	err := NewError("TEST_002").
		Message("simple error").
		Build()

	structuredErr := err.(*StructuredError)

	if structuredErr.Cause != nil {
		t.Error("expected nil cause")
	}
	if structuredErr.Error() != "[TEST_002] simple error" {
		t.Errorf("unexpected error string: %s", structuredErr.Error())
	}
}

func TestStructuredError_Unwrap(t *testing.T) {
	cause := errors.New("cause")
	err := NewError("TEST_003").Cause(cause).Build()

	unwrapped := errors.Unwrap(err)
	if unwrapped != cause {
		t.Error("expected unwrap to return cause")
	}
}

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrBeanNotFound", ErrBeanNotFound},
		{"ErrCircularDependency", ErrCircularDependency},
		{"ErrInvalidConfiguration", ErrInvalidConfiguration},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Error("expected non-nil sentinel error")
			}
			if tt.err.Error() == "" {
				t.Error("expected non-empty error message")
			}
		})
	}
}
