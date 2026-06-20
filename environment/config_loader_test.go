package environment

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoader_LoadWithCustomLocation(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "custom.json")
	configContent := `{"app": {"name": "custom-app"}}`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	loader := NewConfigLoader("application", ConfigTypeJSON, configPath, []string{})
	sources, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(sources) != 1 {
		t.Errorf("Expected 1 source, got %d", len(sources))
	}
	if sources[0].Name() != "custom-config" {
		t.Errorf("Expected name 'custom-config', got %s", sources[0].Name())
	}

	val, ok := sources[0].GetProperty("app.name")
	if !ok {
		t.Error("Expected property 'app.name' to exist")
	}
	if val != "custom-app" {
		t.Errorf("Expected value 'custom-app', got %v", val)
	}
}

func TestConfigLoader_LoadWithProfile(t *testing.T) {
	tmpDir := t.TempDir()

	baseConfig := filepath.Join(tmpDir, "application.json")
	baseContent := `{"app": {"name": "base-app", "port": 8080}}`
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
		t.Fatalf("Failed to change dir: %v", err)
	}

	loader := NewConfigLoader("application", ConfigTypeJSON, "", []string{"dev"})
	sources, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(sources) != 2 {
		t.Errorf("Expected 2 sources, got %d", len(sources))
	}

	if sources[0].Name() != "base-config" {
		t.Errorf("Expected first source name 'base-config', got %s", sources[0].Name())
	}
	val, ok := sources[0].GetProperty("app.name")
	if !ok {
		t.Error("Expected property 'app.name' to exist")
	}
	if val != "base-app" {
		t.Errorf("Expected value 'base-app', got %v", val)
	}

	if sources[1].Name() != "profile-config-dev" {
		t.Errorf("Expected second source name 'profile-config-dev', got %s", sources[1].Name())
	}
	val, ok = sources[1].GetProperty("app.port")
	if !ok {
		t.Error("Expected property 'app.port' to exist")
	}
	if val != 9090.0 { // JSON numbers are float64
		t.Errorf("Expected value 9090.0, got %v", val)
	}
}

func TestConfigLoader_FindConfigFile(t *testing.T) {
	tmpDir := t.TempDir()

	configPath := filepath.Join(tmpDir, "application.json")
	if err := os.WriteFile(configPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(originalDir)
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	loader := NewConfigLoader("application", ConfigTypeJSON, "", []string{})
	foundPath, err := loader.findConfigFile("application")
	if err != nil {
		t.Fatalf("findConfigFile() error = %v", err)
	}
	if foundPath != "application.json" {
		t.Errorf("Expected 'application.json', got %s", foundPath)
	}
}
