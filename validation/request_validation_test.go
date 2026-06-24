package validation

import (
	"net/http"
	"net/url"
	"testing"
)

func TestRequestValidator_Required(t *testing.T) {
	config := ValidationConfig{
		Rules: []ValidationRule{
			{Field: "name", Type: "required"},
		},
		Source: "query",
	}

	validator, err := NewRequestValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Test missing required field
	req := &http.Request{URL: &url.URL{}}
	result := validator.Validate(req)
	if result.Valid {
		t.Error("Expected validation to fail for missing required field")
	}
	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}

	// Test present required field
	req = &http.Request{URL: &url.URL{RawQuery: "name=test"}}
	result = validator.Validate(req)
	if !result.Valid {
		t.Error("Expected validation to pass for present required field")
	}
}

func TestRequestValidator_String(t *testing.T) {
	config := ValidationConfig{
		Rules: []ValidationRule{
			{Field: "name", Type: "string", MinLength: intPtr(2), MaxLength: intPtr(10)},
		},
		Source: "query",
	}

	validator, err := NewRequestValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Test valid string
	req := &http.Request{URL: &url.URL{RawQuery: "name=test"}}
	result := validator.Validate(req)
	if !result.Valid {
		t.Error("Expected validation to pass for valid string")
	}

	// Test too short
	req = &http.Request{URL: &url.URL{RawQuery: "name=a"}}
	result = validator.Validate(req)
	if result.Valid {
		t.Error("Expected validation to fail for too short string")
	}

	// Test too long
	req = &http.Request{URL: &url.URL{RawQuery: "name=verylongstring"}}
	result = validator.Validate(req)
	if result.Valid {
		t.Error("Expected validation to fail for too long string")
	}
}

func TestRequestValidator_Email(t *testing.T) {
	config := ValidationConfig{
		Rules: []ValidationRule{
			{Field: "email", Type: "email"},
		},
		Source: "query",
	}

	validator, err := NewRequestValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Test valid email
	req := &http.Request{URL: &url.URL{RawQuery: "email=test@example.com"}}
	result := validator.Validate(req)
	if !result.Valid {
		t.Error("Expected validation to pass for valid email")
	}

	// Test invalid email
	req = &http.Request{URL: &url.URL{RawQuery: "email=invalid"}}
	result = validator.Validate(req)
	if result.Valid {
		t.Error("Expected validation to fail for invalid email")
	}
}

func TestRequestValidator_Enum(t *testing.T) {
	config := ValidationConfig{
		Rules: []ValidationRule{
			{Field: "status", Type: "enum", In: []string{"active", "inactive"}},
		},
		Source: "query",
	}

	validator, err := NewRequestValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Test valid enum value
	req := &http.Request{URL: &url.URL{RawQuery: "status=active"}}
	result := validator.Validate(req)
	if !result.Valid {
		t.Error("Expected validation to pass for valid enum value")
	}

	// Test invalid enum value
	req = &http.Request{URL: &url.URL{RawQuery: "status=unknown"}}
	result = validator.Validate(req)
	if result.Valid {
		t.Error("Expected validation to fail for invalid enum value")
	}
}

func TestRequestValidator_FailFast(t *testing.T) {
	config := ValidationConfig{
		Rules: []ValidationRule{
			{Field: "name", Type: "required"},
			{Field: "email", Type: "email"},
		},
		Source:   "query",
		FailFast: true,
	}

	validator, err := NewRequestValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Test fail fast - should stop at first error
	req := &http.Request{URL: &url.URL{}}
	result := validator.Validate(req)
	if result.Valid {
		t.Error("Expected validation to fail")
	}
	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error with fail fast, got %d", len(result.Errors))
	}
}

func TestRequestValidator_CustomMessage(t *testing.T) {
	config := ValidationConfig{
		Rules: []ValidationRule{
			{Field: "name", Type: "required", Message: "Name is required, please provide it"},
		},
		Source: "query",
	}

	validator, err := NewRequestValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	req := &http.Request{URL: &url.URL{}}
	result := validator.Validate(req)
	if result.Valid {
		t.Error("Expected validation to fail")
	}
	if result.Errors[0].Message != "Name is required, please provide it" {
		t.Errorf("Expected custom message, got: %s", result.Errors[0].Message)
	}
}

func TestRequestValidator_Regex(t *testing.T) {
	config := ValidationConfig{
		Rules: []ValidationRule{
			{Field: "phone", Type: "regex", Pattern: `^\d{3}-\d{3}-\d{4}$`},
		},
		Source: "query",
	}

	validator, err := NewRequestValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Test valid phone
	req := &http.Request{URL: &url.URL{RawQuery: "phone=123-456-7890"}}
	result := validator.Validate(req)
	if !result.Valid {
		t.Error("Expected validation to pass for valid phone")
	}

	// Test invalid phone
	req = &http.Request{URL: &url.URL{RawQuery: "phone=1234567890"}}
	result = validator.Validate(req)
	if result.Valid {
		t.Error("Expected validation to fail for invalid phone")
	}
}

func TestRequestValidator_InvalidRegex(t *testing.T) {
	config := ValidationConfig{
		Rules: []ValidationRule{
			{Field: "test", Type: "regex", Pattern: `[invalid`},
		},
		Source: "query",
	}

	_, err := NewRequestValidator(config)
	if err == nil {
		t.Error("Expected error for invalid regex pattern")
	}
}

func TestRequestValidator_Number(t *testing.T) {
	minVal := 1.0
	maxVal := 100.0
	config := ValidationConfig{
		Rules: []ValidationRule{
			{Field: "age", Type: "number", Min: &minVal, Max: &maxVal},
		},
		Source: "query",
	}

	validator, err := NewRequestValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Test valid number
	req := &http.Request{URL: &url.URL{RawQuery: "age=25"}}
	result := validator.Validate(req)
	if !result.Valid {
		t.Error("Expected validation to pass for valid number")
	}

	// Test out of range
	req = &http.Request{URL: &url.URL{RawQuery: "age=200"}}
	result = validator.Validate(req)
	if result.Valid {
		t.Error("Expected validation to fail for out of range number")
	}

	// Test not a number
	req = &http.Request{URL: &url.URL{RawQuery: "age=abc"}}
	result = validator.Validate(req)
	if result.Valid {
		t.Error("Expected validation to fail for non-number")
	}
}

func TestRequestValidator_Header(t *testing.T) {
	config := ValidationConfig{
		Rules: []ValidationRule{
			{Field: "X-Api-Key", Type: "required"},
		},
		Source: "header",
	}

	validator, err := NewRequestValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Test missing header
	req := &http.Request{Header: make(http.Header)}
	result := validator.Validate(req)
	if result.Valid {
		t.Error("Expected validation to fail for missing header")
	}

	// Test present header
	req = &http.Request{Header: http.Header{"X-Api-Key": []string{"secret-key"}}}
	result = validator.Validate(req)
	if !result.Valid {
		t.Error("Expected validation to pass for present header")
	}
}

func TestValidateJSONBody(t *testing.T) {
	rules := []ValidationRule{
		{Field: "name", Type: "required"},
		{Field: "email", Type: "email"},
	}

	// Test valid JSON
	body := []byte(`{"name":"test","email":"test@example.com"}`)
	result := ValidateJSONBody(body, rules)
	if !result.Valid {
		t.Errorf("Expected validation to pass, got errors: %v", result.Errors)
	}

	// Test invalid JSON
	body = []byte(`{"name":"","email":"invalid"}`)
	result = ValidateJSONBody(body, rules)
	if result.Valid {
		t.Error("Expected validation to fail")
	}

	// Test malformed JSON
	body = []byte(`{invalid json}`)
	result = ValidateJSONBody(body, rules)
	if result.Valid {
		t.Error("Expected validation to fail for malformed JSON")
	}
}

func TestValidateHeaders(t *testing.T) {
	rules := []ValidationRule{
		{Field: "X-Request-Id", Type: "required"},
	}

	req := &http.Request{Header: http.Header{"X-Request-Id": []string{"123"}}}
	result := ValidateHeaders(req, rules)
	if !result.Valid {
		t.Error("Expected validation to pass")
	}

	req = &http.Request{Header: make(http.Header)}
	result = ValidateHeaders(req, rules)
	if result.Valid {
		t.Error("Expected validation to fail")
	}
}

func TestValidateQuery(t *testing.T) {
	rules := []ValidationRule{
		{Field: "page", Type: "required"},
	}

	req := &http.Request{URL: &url.URL{RawQuery: "page=1"}}
	result := ValidateQuery(req, rules)
	if !result.Valid {
		t.Error("Expected validation to pass")
	}

	req = &http.Request{URL: &url.URL{}}
	result = ValidateQuery(req, rules)
	if result.Valid {
		t.Error("Expected validation to fail")
	}
}

func intPtr(i int) *int {
	return &i
}
