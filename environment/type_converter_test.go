package environment

import (
	"reflect"
	"testing"
)

func TestTypeConverter_ConvertTo_Int(t *testing.T) {
	c := NewTypeConverter()

	tests := []struct {
		input    any
		expected int
	}{
		{42, 42},
		{float64(42.5), 42},
		{"123", 123},
		{int64(456), 456},
	}

	for _, tt := range tests {
		result, err := c.ConvertTo(tt.input, reflect.TypeOf(int(0)))
		if err != nil {
			t.Errorf("ConvertTo(%v, int) error: %v", tt.input, err)
			continue
		}
		if result.Int() != int64(tt.expected) {
			t.Errorf("ConvertTo(%v, int) = %d, want %d", tt.input, result.Int(), tt.expected)
		}
	}
}

func TestTypeConverter_ConvertTo_Bool(t *testing.T) {
	c := NewTypeConverter()

	tests := []struct {
		input    any
		expected bool
	}{
		{true, true},
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		result, err := c.ConvertTo(tt.input, reflect.TypeOf(false))
		if err != nil {
			t.Errorf("ConvertTo(%v, bool) error: %v", tt.input, err)
			continue
		}
		if result.Bool() != tt.expected {
			t.Errorf("ConvertTo(%v, bool) = %v, want %v", tt.input, result.Bool(), tt.expected)
		}
	}
}

func TestTypeConverter_ConvertTo_String(t *testing.T) {
	c := NewTypeConverter()

	tests := []struct {
		input    any
		expected string
	}{
		{42, "42"},
		{3.14, "3.14"},
		{true, "true"},
		{"hello", "hello"},
	}

	for _, tt := range tests {
		result, err := c.ConvertTo(tt.input, reflect.TypeOf(""))
		if err != nil {
			t.Errorf("ConvertTo(%v, string) error: %v", tt.input, err)
			continue
		}
		if result.String() != tt.expected {
			t.Errorf("ConvertTo(%v, string) = %q, want %q", tt.input, result.String(), tt.expected)
		}
	}
}
