package actuator

import (
	"testing"
)

func TestSanitizer_Sanitize_KeyWords(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		key      string
		value    any
		expected any
	}{
		{"password", "secret123", redactedValue},
		{"db.password", "mypass", redactedValue},
		{"api_key", "abc123", redactedValue},
		{"client_secret", "xyz", redactedValue},
		{"normal_field", "normal_value", "normal_value"},
		{"username", "admin", "admin"},
	}

	for _, tt := range tests {
		result := s.Sanitize(tt.key, tt.value)
		if result != tt.expected {
			t.Errorf("Sanitize(%q, %v) = %v, want %v", tt.key, tt.value, result, tt.expected)
		}
	}
}

func TestSanitizer_Sanitize_TokenFormats(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		key      string
		value    string
		expected any
	}{
		{"token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.abc123", redactedValue},
		{"key", "-----BEGIN RSA PRIVATE KEY-----test", redactedValue},
		{"value", "normal_string", "normal_string"},
	}

	for _, tt := range tests {
		result := s.Sanitize(tt.key, tt.value)
		if result != tt.expected {
			t.Errorf("Sanitize(%q, %q) = %v, want %v", tt.key, tt.value, result, tt.expected)
		}
	}
}

func TestSanitizer_AddStrategy(t *testing.T) {
	s := NewSanitizer()

	// 添加自定义策略
	s.AddStrategy(&customStrategy{})

	result := s.Sanitize("custom_key", "custom_value")
	if result != redactedValue {
		t.Errorf("Expected redacted value for custom strategy, got %v", result)
	}
}

type customStrategy struct{}

func (c *customStrategy) IsSensitive(key string, value any) bool {
	return key == "custom_key"
}
