package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewComponentScanner(t *testing.T) {
	scanner := NewComponentScanner(".")
	if scanner == nil {
		t.Error("NewComponentScanner() returned nil")
	}
}

func TestComponentScanner_Scan(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.go")
	content := `package test

// @Component
type UserService struct {
	Name string
}

// @Configuration
type AppConfig struct {
	Port int
}

// @Service
type UserRepo struct {
	Table string
}

// @Component("customName")
type TestComponent struct {
	Field string
}

// RegularStruct is not a component
type RegularStruct struct {
	Field string
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	container := New()
	scanner := NewComponentScanner(tmpDir)

	err := scanner.Scan(container)
	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}

	if !container.Has("userService") {
		t.Error("userService should be registered")
	}

	if !container.Has("appConfig") {
		t.Error("appConfig should be registered")
	}

	if !container.Has("userRepo") {
		t.Error("userRepo should be registered")
	}

	if !container.Has("customName") {
		t.Error("customName should be registered")
	}

	if container.Has("regularStruct") {
		t.Error("regularStruct should not be registered")
	}
}

func TestComponentScanner_Scan_NilDoc(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.go")
	content := `package test

type NoComment struct {
	Name string
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	container := New()
	scanner := NewComponentScanner(tmpDir)

	err := scanner.Scan(container)
	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}

	if container.Has("noComment") {
		t.Error("noComment should not be registered")
	}
}

func TestComponentScanner_WithCustomName(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.go")
	content := `package test

// @Component("myCustomService")
type MyService struct {
	Name string
}

// @Component
type AnotherService struct {
	Address string
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	container := New()
	scanner := NewComponentScanner(tmpDir)

	err := scanner.Scan(container)
	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}

	if !container.Has("myCustomService") {
		t.Error("myCustomService should be registered with custom name")
	}
	if !container.Has("anotherService") {
		t.Error("anotherService should be registered")
	}
}

func TestToFirstCharLower(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"UserService", "userService"},
		{"Config", "config"},
		{"A", "a"},
		{"", ""},
		{"ABC", "aBC"},
	}

	for _, tt := range tests {
		result := toFirstCharLower(tt.input)
		if result != tt.expected {
			t.Errorf("toFirstCharLower(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}
