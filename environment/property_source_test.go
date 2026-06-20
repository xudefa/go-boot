package environment

import (
	"testing"
)

func TestMapPropertySource_GetProperty(t *testing.T) {
	src := NewMapPropertySource("test", 0, map[string]any{
		"server.port": 8080,
		"server.host": "localhost",
	})

	val, ok := src.GetProperty("server.port")
	if !ok {
		t.Fatal("expected property to exist")
	}
	v, ok := val.(int)
	if !ok {
		t.Fatalf("expected int, got %T", val)
	}
	if v != 8080 {
		t.Fatalf("expected 8080, got %v", val)
	}

	_, ok = src.GetProperty("nonexistent")
	if ok {
		t.Fatal("expected property to not exist")
	}
}

func TestMapPropertySource_Priority(t *testing.T) {
	src1 := NewMapPropertySource("low", 100, nil)
	src2 := NewMapPropertySource("high", 200, nil)

	if src1.Priority() >= src2.Priority() {
		t.Fatal("expected src1 to have lower priority than src2")
	}
}

func TestArgsPropertySource(t *testing.T) {
	src := NewArgsPropertySource("args", []string{
		"--server.port=9090",
		"--server.host=example.com",
		"--some-flag",
	})

	val, ok := src.GetProperty("server.port")
	if !ok {
		t.Fatal("expected server.port to exist")
	}
	v, ok := val.(string)
	if !ok {
		t.Fatalf("expected string, got %T", val)
	}
	if v != "9090" {
		t.Fatalf("expected 9090, got %v", val)
	}

	val, ok = src.GetProperty("server.host")
	if !ok {
		t.Fatal("expected server.host to exist")
	}
	s, ok := val.(string)
	if !ok {
		t.Fatalf("expected string, got %T", val)
	}
	if s != "example.com" {
		t.Fatalf("expected example.com, got %v", val)
	}
}

func TestToEnvKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"server.port", "SERVER_PORT"},
		{"server.host", "SERVER_HOST"},
		{"app.name", "APP_NAME"},
		{"simple", "SIMPLE"},
		{"nested.key.path", "NESTED_KEY_PATH"},
	}
	for _, tt := range tests {
		result := toEnvKey(tt.input)
		if result != tt.expected {
			t.Errorf("toEnvKey(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestEnvPropertySource_GetProperty(t *testing.T) {
	origLookupEnv := lookupEnv
	defer func() { lookupEnv = origLookupEnv }()

	lookupEnv = func(key string) (string, bool) {
		if key == "GO_BOOT_SERVER_PORT" {
			return "9090", true
		}
		return "", false
	}

	src := NewEnvPropertySource("env", "GO_BOOT")
	val, ok := src.GetProperty("server.port")
	if !ok {
		t.Fatal("expected server.port to exist")
	}
	v, ok := val.(string)
	if !ok {
		t.Fatalf("expected string, got %T", val)
	}
	if v != "9090" {
		t.Fatalf("expected 9090, got %v", val)
	}
}

func TestNewDefaultPropertySource(t *testing.T) {
	src := NewDefaultPropertySource("defaults", map[string]any{
		"server.port": 8080,
		"server.host": "localhost",
	})

	if src.Name() != "defaults" {
		t.Fatalf("Name() = %s, want defaults", src.Name())
	}
	if src.Priority() != PriorityLowest {
		t.Fatalf("Priority() = %d, want %d", src.Priority(), PriorityLowest)
	}

	val, ok := src.GetProperty("server.port")
	if !ok {
		t.Fatal("expected server.port to exist")
	}
	v, ok := val.(int)
	if !ok || v != 8080 {
		t.Fatalf("server.port = %v, want 8080", val)
	}
}

func TestDefaultPropertySource_OverriddenByOtherSource(t *testing.T) {
	env := NewEnvironment()

	defaults := NewDefaultPropertySource("defaults", map[string]any{
		"server.port": 8080,
		"server.host": "default.com",
	})
	normal := NewMapPropertySource("normal", PriorityNormal, map[string]any{
		"server.port": 9090,
	})

	env.AddPropertySource(defaults)
	env.AddPropertySource(normal)

	port := env.GetInt("server.port", 0)
	if port != 9090 {
		t.Fatalf("expected 9090 (normal priority), got %d", port)
	}
	host := env.GetString("server.host", "")
	if host != "default.com" {
		t.Fatalf("expected default.com (fallback from defaults), got %s", host)
	}
}

func TestEnvPropertySource_EmptyPrefix(t *testing.T) {
	origLookupEnv := lookupEnv
	defer func() { lookupEnv = origLookupEnv }()

	lookupEnv = func(key string) (string, bool) {
		if key == "SERVER_PORT" {
			return "8080", true
		}
		return "", false
	}

	src := NewEnvPropertySource("env", "")
	val, ok := src.GetProperty("server.port")
	if !ok {
		t.Fatal("expected server.port to exist with empty prefix")
	}
	v, ok := val.(string)
	if !ok {
		t.Fatalf("expected string, got %T", val)
	}
	if v != "8080" {
		t.Fatalf("expected 8080, got %v", val)
	}
}
