package environment

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewEnvironmentWithApplicationConfig(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	// 创建应用配置文件
	applicationConfig := `{
		"server": {
			"host": "0.0.0.0",
			"port": 9090
		},
		"app": {
			"name": "test-app"
		}
	}`

	configFile := filepath.Join(tmpDir, "application.json")
	if err := os.WriteFile(configFile, []byte(applicationConfig), 0644); err != nil {
		t.Fatalf("Failed to create application config: %v", err)
	}

	// 保存当前工作目录
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	// 切换到临时目录
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// 创建 Environment
	env := NewEnvironment()

	// 验证配置源数量
	sources := env.GetPropertySources()
	if len(sources) < 3 {
		t.Errorf("Expected at least 3 sources, got %d", len(sources))
	}

	// 验证应用配置是否被加载
	val, ok := env.GetProperty("server.host")
	if !ok {
		t.Error("Expected to find 'server.host' from application config")
	}
	if val != "0.0.0.0" {
		t.Errorf("Expected '0.0.0.0', got '%v'", val)
	}

	val, ok = env.GetProperty("server.port")
	if !ok {
		t.Error("Expected to find 'server.port' from application config")
	}
	// JSON 中的数字会被解析为 float64
	if val != float64(9090) {
		t.Errorf("Expected 9090, got '%v'", val)
	}

	val, ok = env.GetProperty("app.name")
	if !ok {
		t.Error("Expected to find 'app.name' from application config")
	}
	if val != "test-app" {
		t.Errorf("Expected 'test-app', got '%v'", val)
	}
}

func TestNewEnvironmentWithoutApplicationConfig(t *testing.T) {
	// 创建临时空目录
	tmpDir := t.TempDir()

	// 保存当前工作目录
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	// 切换到临时目录
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// 创建 Environment（没有应用配置文件）
	env := NewEnvironment()

	// 验证配置源数量（应该只有 args 和 env）
	sources := env.GetPropertySources()
	if len(sources) != 2 {
		t.Errorf("Expected 2 sources (args and env), got %d", len(sources))
	}

	// 验证没有应用配置的值
	_, ok := env.GetProperty("server.host")
	if ok {
		t.Error("Expected not to find 'server.host' without application config")
	}
}

func TestNewEnvironmentPriority(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	// 创建应用配置文件
	applicationConfig := `{
		"server": {
			"port": 8080
		}
	}`

	configFile := filepath.Join(tmpDir, "application.json")
	if err := os.WriteFile(configFile, []byte(applicationConfig), 0644); err != nil {
		t.Fatalf("Failed to create application config: %v", err)
	}

	// 保存当前工作目录
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	// 切换到临时目录
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// 设置环境变量（应该覆盖应用配置）
	_ = os.Setenv("GO_BOOT_SERVER_PORT", "9090")
	defer func() {
		_ = os.Unsetenv("GO_BOOT_SERVER_PORT")
	}()

	// 创建 Environment
	env := NewEnvironment()

	// 验证环境变量优先级更高
	val, ok := env.GetProperty("server.port")
	if !ok {
		t.Error("Expected to find 'server.port'")
	}
	// 环境变量值是字符串，JSON 中的数字会被解析为 float64
	// 环境变量应该覆盖 JSON 配置
	expectedVal := "9090"
	if val != expectedVal {
		t.Errorf("Expected '%s' from env var, got '%v' (type: %T)", expectedVal, val, val)
	}
}
