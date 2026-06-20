package environment

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/xudefa/go-boot/refresh"
)

func TestEnvironment_GetProperty(t *testing.T) {
	env := NewEnvironment()

	src1 := NewMapPropertySource("high", 200, map[string]any{"key": "value1"})
	src2 := NewMapPropertySource("low", 100, map[string]any{"key": "value2"})

	env.AddPropertySource(src2)
	env.AddPropertySource(src1)

	val := env.GetString("key", "")
	if val != "value1" {
		t.Fatalf("expected value1 (higher priority), got %s", val)
	}
}

func TestEnvironment_Profile(t *testing.T) {
	env := NewEnvironment()
	env.AddActiveProfile("dev")

	if !env.AcceptsProfile("dev") {
		t.Fatal("expected to accept dev profile")
	}
	if env.AcceptsProfile("prod") {
		t.Fatal("expected to not accept prod profile")
	}
	if !env.AcceptsProfile("!prod") {
		t.Fatal("expected to accept !prod when prod is not active")
	}
}

func TestEnvironment_MultiSourceMerge(t *testing.T) {
	env := NewEnvironment()

	low := NewMapPropertySource("low", 100, map[string]any{
		"server.port": 8080,
		"server.host": "default.com",
	})
	high := NewMapPropertySource("high", 200, map[string]any{
		"server.port": 9090,
	})

	env.AddPropertySource(low)
	env.AddPropertySource(high)

	port := env.GetInt("server.port", 0)
	if port != 9090 {
		t.Fatalf("expected 9090 from high priority, got %d", port)
	}
	host := env.GetString("server.host", "")
	if host != "default.com" {
		t.Fatalf("expected default.com from low priority fallback, got %s", host)
	}
}

func TestEnvironment_GetBool(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", 0, map[string]any{
		"enabled":  true,
		"disabled": false,
	}))

	if !env.GetBool("enabled", false) {
		t.Fatal("expected enabled to be true")
	}
	if env.GetBool("disabled", true) {
		t.Fatal("expected disabled to be false")
	}
}

func TestEnvironment_ContainsProperty(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", 0, map[string]any{
		"exists": "yes",
	}))

	if !env.ContainsProperty("exists") {
		t.Fatal("expected ContainsProperty to be true")
	}
	if env.ContainsProperty("missing") {
		t.Fatal("expected ContainsProperty to be false")
	}
}

func TestResolvePlaceholders_Basic(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.name": "myapp",
		"host":     "localhost",
		"port":     8080,
	}))

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "resolve existing key",
			input: "${app.name}",
			want:  "myapp",
		},
		{
			name:  "placeholder in text",
			input: "http://${host}:${port}",
			want:  "http://localhost:8080",
		},
		{
			name:  "no placeholder",
			input: "plain text",
			want:  "plain text",
		},
		{
			name:  "nonexistent key without default",
			input: "${nonexistent}",
			want:  "${nonexistent}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := env.ResolvePlaceholders(tt.input)
			if got != tt.want {
				t.Errorf("ResolvePlaceholders(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolvePlaceholders_DefaultValue(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"host": "localhost",
	}))

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "existing key with default",
			input: "${host:fallback}",
			want:  "localhost",
		},
		{
			name:  "nonexistent key with default",
			input: "${port:3306}",
			want:  "3306",
		},
		{
			name:  "default with nested placeholder",
			input: "${nonexistent:${host}}",
			want:  "localhost",
		},
		{
			name:  "nested default fallback chain",
			input: "${a:${b:fallback}}",
			want:  "fallback",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := env.ResolvePlaceholders(tt.input)
			if got != tt.want {
				t.Errorf("ResolvePlaceholders(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolvePlaceholders_Recursive(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"env":     "dev",
		"host":    "my-${env}.com",
		"url":     "http://${host}:${port}",
		"port":    "8080",
		"chained": "${url}",
	}))

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "single level",
			input: "${host}",
			want:  "my-dev.com",
		},
		{
			name:  "multi level",
			input: "${url}",
			want:  "http://my-dev.com:8080",
		},
		{
			name:  "chained reference",
			input: "${chained}",
			want:  "http://my-dev.com:8080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := env.ResolvePlaceholders(tt.input)
			if got != tt.want {
				t.Errorf("ResolvePlaceholders(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolvePlaceholders_CircularReference(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"a": "${b}",
		"b": "${a}",
	}))

	got := env.ResolvePlaceholders("${a}")
	if got != "${a}" {
		t.Fatalf("expected '${a}' (preserved on circular ref), got %q", got)
	}
}

func TestGetProperty_AutoResolvePlaceholders(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"host": "localhost",
		"port": 8080,
		"url":  "http://${host}:${port}",
	}))

	val, ok := env.GetProperty("url")
	if !ok {
		t.Fatal("expected url to exist")
	}
	got, ok := val.(string)
	if !ok {
		t.Fatalf("expected string, got %T", val)
	}
	if got != "http://localhost:8080" {
		t.Fatalf("url = %q, want http://localhost:8080", got)
	}
}

func TestResolvePlaceholders_Nested(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"db.host":    "localhost",
		"db.port":    "5432",
		"db.url":     "jdbc:postgresql://${db.host}:${db.port}/mydb",
		"app.db.url": "${db.url}?ssl=true",
	}))

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "nested placeholder in value",
			input: "${app.db.url}",
			want:  "jdbc:postgresql://localhost:5432/mydb?ssl=true",
		},
		{
			name:  "placeholder with colon in text",
			input: "prefix:${db.host}:suffix",
			want:  "prefix:localhost:suffix",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := env.ResolvePlaceholders(tt.input)
			if got != tt.want {
				t.Errorf("ResolvePlaceholders(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolvePlaceholders_CircularDependencyDetection(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"a": "${b}",
		"b": "${c}",
		"c": "${a}",
	}))

	got := env.ResolvePlaceholders("${a}")
	if got != "${a}" {
		t.Fatalf("expected '${a}' (preserved on circular ref), got %q", got)
	}
}

func TestResolvePlaceholders_SelfReference(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"self": "${self}",
	}))

	got := env.ResolvePlaceholders("${self}")
	if got != "${self}" {
		t.Fatalf("expected '${self}' (preserved on self-circular ref), got %q", got)
	}
}

func TestParseProfiles(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"dev", []string{"dev"}},
		{"dev,prod", []string{"dev", "prod"}},
		{" dev , prod ", []string{"dev", "prod"}},
		{"", nil},
	}
	for _, tt := range tests {
		result := ParseProfiles(tt.input)
		if len(result) != len(tt.expected) {
			t.Fatalf("ParseProfiles(%q) = %v, want %v", tt.input, result, tt.expected)
		}
		for i := range result {
			if result[i] != tt.expected[i] {
				t.Fatalf("ParseProfiles(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		}
	}
}

func TestEnvironment_ConfigChangeListener(t *testing.T) {
	env := NewEnvironment()

	var called int32
	env.AddConfigChangeListener(func(event refresh.ConfigChangeEvent) {
		atomic.StoreInt32(&called, 1)
	})

	event := refresh.NewConfigChangeEvent(
		"modify",
		[]string{"test.key"},
		map[string]any{"test.key": "old"},
		map[string]any{"test.key": "new"},
		"test",
	)

	env.notifyConfigChange(event)

	// 等待异步通知
	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt32(&called) == 0 {
		t.Error("expected config change listener to be called")
	}
}
