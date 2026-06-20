package environment

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewJSONPropertySource(t *testing.T) {
	// 创建临时测试文件
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")

	testData := `{
		"server": {
			"host": "localhost",
			"port": 8080
		},
		"app": {
			"name": "test-app",
			"version": "1.0.0"
		}
	}`

	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 测试正常加载
	source, err := NewJSONPropertySource("test", testFile)
	if err != nil {
		t.Fatalf("Failed to create JSONPropertySource: %v", err)
	}

	if source.Name() != "test" {
		t.Errorf("Expected name 'test', got '%s'", source.Name())
	}

	if source.Priority() != PriorityLowest {
		t.Errorf("Expected priority PriorityLowest, got %v", source.Priority())
	}

	// 测试获取属性
	val, ok := source.GetProperty("server.host")
	if !ok {
		t.Error("Expected to find 'server.host'")
	}
	if val != "localhost" {
		t.Errorf("Expected 'localhost', got '%v'", val)
	}

	val, ok = source.GetProperty("server.port")
	if !ok {
		t.Error("Expected to find 'server.port'")
	}
	if val != float64(8080) {
		t.Errorf("Expected 8080, got '%v'", val)
	}

	// 测试不存在的键
	_, ok = source.GetProperty("nonexistent.key")
	if ok {
		t.Error("Expected not to find 'nonexistent.key'")
	}

	// 测试 Contains
	if !source.Contains("app.name") {
		t.Error("Expected Contains('app.name') to return true")
	}
	if source.Contains("nonexistent") {
		t.Error("Expected Contains('nonexistent') to return false")
	}

	// 测试 Keys
	keys := source.Keys()
	if len(keys) != 4 {
		t.Errorf("Expected 4 keys, got %d", len(keys))
	}
}

func TestNewJSONPropertySourceOrDefault(t *testing.T) {
	// 测试不存在的文件
	source := NewJSONPropertySourceOrDefault("test", "/nonexistent/file.json")
	if source == nil {
		t.Error("Expected source to be created even for nonexistent file")
	}

	_, ok := source.GetProperty("any.key")
	if ok {
		t.Error("Expected no properties from nonexistent file")
	}
}

func TestJSONPropertySourceNestedKeys(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "nested.json")

	testData := `{
		"level1": {
			"level2": {
				"level3": {
					"value": "deep"
				}
			}
		}
	}`

	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	source, err := NewJSONPropertySource("nested", testFile)
	if err != nil {
		t.Fatalf("Failed to create JSONPropertySource: %v", err)
	}

	val, ok := source.GetProperty("level1.level2.level3.value")
	if !ok {
		t.Error("Expected to find deep nested key")
	}
	if val != "deep" {
		t.Errorf("Expected 'deep', got '%v'", val)
	}
}

func TestJSONPropertySourceInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.json")

	if err := os.WriteFile(testFile, []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := NewJSONPropertySource("invalid", testFile)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestFindApplicationConfigFile(t *testing.T) {
	// 创建临时目录结构
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	if err := os.Mkdir(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// 创建应用配置文件
	applicationConfig := filepath.Join(tmpDir, "application.json")
	if err := os.WriteFile(applicationConfig, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create application config file: %v", err)
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

	// 测试查找
	found := FindApplicationConfigFile()
	if found == "" {
		t.Error("Expected to find application config file")
	}

	// 测试在 config 目录中查找
	if err := os.Remove(applicationConfig); err != nil {
		t.Fatalf("Failed to remove application config: %v", err)
	}

	configFile := filepath.Join(configDir, "application.json")
	if err := os.WriteFile(configFile, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	found = FindApplicationConfigFile()
	if found == "" {
		t.Error("Expected to find application config file in config directory")
	}
}

func TestFlattenKeys(t *testing.T) {
	data := map[string]any{
		"simple": "value",
		"nested": map[string]any{
			"key1": "val1",
			"key2": "val2",
		},
		"deep": map[string]any{
			"level": map[string]any{
				"final": "result",
			},
		},
	}

	keys := flattenKeys(data, "")

	expectedKeys := []string{"simple", "nested.key1", "nested.key2", "deep.level.final"}
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(keys))
	}

	keyMap := make(map[string]bool)
	for _, key := range keys {
		keyMap[key] = true
	}

	for _, expected := range expectedKeys {
		if !keyMap[expected] {
			t.Errorf("Expected key '%s' not found", expected)
		}
	}
}
