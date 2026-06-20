package boot

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xudefa/go-boot/environment"
)

func TestBoot_LoadConfigFiles(t *testing.T) {
	tmpDir := t.TempDir()

	baseConfig := filepath.Join(tmpDir, "application.json")
	baseContent := `{"app": {"name": "test-app", "port": 8080}}`
	if err := os.WriteFile(baseConfig, []byte(baseContent), 0644); err != nil {
		t.Fatalf("Failed to write base config: %v", err)
	}

	devConfig := filepath.Join(tmpDir, "application-dev.json")
	devContent := `{"app": {"port": 9090, "debug": true}}`
	if err := os.WriteFile(devConfig, []byte(devContent), 0644); err != nil {
		t.Fatalf("Failed to write dev config: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(originalDir)
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	boot, err := NewApplication(
		WithProfiles("dev"),
	)
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	if boot == nil {
		t.Fatal("Expected boot to not be nil")
	}

	err = boot.Start()
	if err != nil {
		t.Fatalf("boot.Start() error = %v", err)
	}
	defer func() {
		_ = boot.Stop()
	}()

	env := boot.Environment()

	name := env.GetString("app.name", "")
	if name != "test-app" {
		t.Errorf("Expected app.name 'test-app', got %s", name)
	}

	port := env.GetInt("app.port", 0)
	if port != 9090 {
		t.Errorf("Expected app.port 9090, got %d", port)
	}

	debug := env.GetBool("app.debug", false)
	if !debug {
		t.Error("Expected app.debug to be true")
	}
}

func TestBoot_CustomConfigLocation(t *testing.T) {
	tmpDir := t.TempDir()

	customConfig := filepath.Join(tmpDir, "custom.json")
	customContent := `{"app": {"name": "custom-app", "port": 8888}}`
	if err := os.WriteFile(customConfig, []byte(customContent), 0644); err != nil {
		t.Fatalf("Failed to write custom config: %v", err)
	}

	boot, err := NewApplication(
		WithConfigLocation(customConfig),
	)
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	if boot == nil {
		t.Fatal("Expected boot to not be nil")
	}

	err = boot.Start()
	if err != nil {
		t.Fatalf("boot.Start() error = %v", err)
	}
	defer func() {
		_ = boot.Stop()
	}()

	env := boot.Environment()

	name := env.GetString("app.name", "")
	if name != "custom-app" {
		t.Errorf("Expected app.name 'custom-app', got %s", name)
	}

	port := env.GetInt("app.port", 0)
	if port != 8888 {
		t.Errorf("Expected app.port 8888, got %d", port)
	}
}

func TestBoot_CustomPropertySource(t *testing.T) {
	customSource := environment.NewMapPropertySource(
		"custom",
		environment.PriorityHigh,
		map[string]any{
			"app.name": "custom-source-app",
			"app.port": 7777,
		},
	)

	boot, err := NewApplication(
		WithPropertySource(customSource),
	)
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	if boot == nil {
		t.Fatal("Expected boot to not be nil")
	}

	err = boot.Start()
	if err != nil {
		t.Fatalf("boot.Start() error = %v", err)
	}
	defer func() {
		_ = boot.Stop()
	}()

	env := boot.Environment()

	name := env.GetString("app.name", "")
	if name != "custom-source-app" {
		t.Errorf("Expected app.name 'custom-source-app', got %s", name)
	}

	port := env.GetInt("app.port", 0)
	if port != 7777 {
		t.Errorf("Expected app.port 7777, got %d", port)
	}
}
