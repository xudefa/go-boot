package config

import (
	"errors"
	"testing"
)

func TestConfigBuilder_Defaults(t *testing.T) {
	builder := NewConfigBuilder()

	if builder.configName != "application" {
		t.Errorf("expected default configName 'application', got %s", builder.configName)
	}

	if len(builder.configPaths) != 2 {
		t.Errorf("expected 2 default paths, got %d", len(builder.configPaths))
	}

	if builder.configType != "json" {
		t.Errorf("expected default configType 'json', got %s", builder.configType)
	}
}

func TestConfigBuilder_ChainConfig(t *testing.T) {
	model, err := NewConfigBuilder().
		Name("myapp").
		Paths("/etc/app", "./config").
		Type("yaml").
		Environment("prod").
		EnvPrefix("MYAPP").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if model.ConfigName != "myapp" {
		t.Errorf("expected ConfigName 'myapp', got %s", model.ConfigName)
	}

	if len(model.ConfigPaths) != 2 || model.ConfigPaths[0] != "/etc/app" {
		t.Errorf("expected ConfigPaths ['/etc/app', './config'], got %v", model.ConfigPaths)
	}

	if model.ConfigType != "yaml" {
		t.Errorf("expected ConfigType 'yaml', got %s", model.ConfigType)
	}

	if model.Env != "prod" {
		t.Errorf("expected Env 'prod', got %s", model.Env)
	}

	if model.OptionName != "MYAPP" {
		t.Errorf("expected OptionName 'MYAPP', got %s", model.OptionName)
	}
}

func TestConfigBuilder_WithFile(t *testing.T) {
	model, err := NewConfigBuilder().
		File("/etc/app/config.json").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if model.ConfigFile != "/etc/app/config.json" {
		t.Errorf("expected ConfigFile '/etc/app/config.json', got %s", model.ConfigFile)
	}
}

func TestConfigBuilder_BuildAndLoad_NoLoader(t *testing.T) {
	model, err := NewConfigBuilder().
		Name("test").
		BuildAndLoad()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if model == nil {
		t.Fatal("expected non-nil model")
	}
}

func TestConfigBuilder_BuildAndLoad_WithValidator(t *testing.T) {
	validator := NewValidationRuleBuilder().
		Required("name", "port").
		Min("port", 1).
		Max("port", 65535).
		Build()

	// 使用 mock loader 提供所需的字段
	mockLoader := &mockLoader{
		data: map[string]any{
			"name": "test-app",
			"port": 8080,
		},
	}

	model, err := NewConfigBuilder().
		Name("test").
		Validator(validator).
		Loader(mockLoader).
		BuildAndLoad()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if model == nil {
		t.Fatal("expected non-nil model")
	}
}

func TestValidationRuleBuilder_Required(t *testing.T) {
	validator := NewValidationRuleBuilder().
		Required("name", "email").
		Build()

	data := map[string]any{
		"name": "test",
	}

	err := validator.Validate(data)
	if err == nil {
		t.Error("expected validation error for missing 'email'")
	}

	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("expected ValidationError, got %T", err)
	}

	if validationErr.Field != "email" {
		t.Errorf("expected field 'email', got %s", validationErr.Field)
	}
}

func TestValidationRuleBuilder_MinMax(t *testing.T) {
	validator := NewValidationRuleBuilder().
		Min("port", 1).
		Max("port", 65535).
		Build()

	data := map[string]any{
		"port": 8080,
	}

	err := validator.Validate(data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	data["port"] = 0
	err = validator.Validate(data)
	if err == nil {
		t.Error("expected validation error for port below minimum")
	}

	data["port"] = 70000
	err = validator.Validate(data)
	if err == nil {
		t.Error("expected validation error for port above maximum")
	}
}

func TestValidationRuleBuilder_Regex(t *testing.T) {
	validator := NewValidationRuleBuilder().
		Regex("email", `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).
		Build()

	data := map[string]any{
		"email": "test@example.com",
	}

	err := validator.Validate(data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	data["email"] = "invalid-email"
	err = validator.Validate(data)
	if err == nil {
		t.Error("expected validation error for invalid email")
	}
}

func TestValidationRuleBuilder_Enum(t *testing.T) {
	validator := NewValidationRuleBuilder().
		Enum("status", "active", "inactive").
		Build()

	data := map[string]any{
		"status": "active",
	}

	err := validator.Validate(data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	data["status"] = "unknown"
	err = validator.Validate(data)
	if err == nil {
		t.Error("expected validation error for invalid enum value")
	}
}

func TestValidationRuleBuilder_Custom(t *testing.T) {
	validator := NewValidationRuleBuilder().
		Custom("age", func(val any) error {
			age, ok := val.(int)
			if !ok {
				return errors.New("age must be integer")
			}
			if age < 0 || age > 150 {
				return errors.New("age must be between 0 and 150")
			}
			return nil
		}).
		Build()

	data := map[string]any{
		"age": 25,
	}

	err := validator.Validate(data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	data["age"] = 200
	err = validator.Validate(data)
	if err == nil {
		t.Error("expected validation error for invalid age")
	}
}

func TestConfigWatcher_OnChange(t *testing.T) {
	watcher := NewConfigWatcher()

	called := false
	watcher.OnChange("test", func(event WatchEvent) {
		called = true
	})

	watcher.Notify(WatchEvent{
		Type:  EventModify,
		Key:   "test.key",
		Value: "new",
	})

	if !called {
		t.Error("expected callback to be called")
	}
}

func TestConfigWatcher_Remove(t *testing.T) {
	watcher := NewConfigWatcher()

	callCount := 0
	watcher.OnChange("test", func(event WatchEvent) {
		callCount++
	})

	watcher.Notify(WatchEvent{Type: EventModify})
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}

	watcher.Remove("test")
	watcher.Notify(WatchEvent{Type: EventModify})
	if callCount != 1 {
		t.Errorf("expected still 1 call after remove, got %d", callCount)
	}
}

func TestConfigWatcher_Close(t *testing.T) {
	watcher := NewConfigWatcher()

	watcher.OnChange("test", func(event WatchEvent) {
		// callback
	})

	watcher.Close()

	// 关闭后不应再触发回调
	watcher.Notify(WatchEvent{Type: EventModify})
}

func TestConfigWatcher_Manager(t *testing.T) {
	watcher := NewConfigWatcher()

	manager := watcher.Manager()
	if manager == nil {
		t.Error("expected non-nil manager")
	}
}

func TestMergeMaps(t *testing.T) {
	dst := map[string]any{
		"a": 1,
		"b": map[string]any{
			"c": 2,
			"d": 3,
		},
	}

	src := map[string]any{
		"b": map[string]any{
			"d": 4,
			"e": 5,
		},
		"f": 6,
	}

	mergeMaps(dst, src)

	if dst["a"] != 1 {
		t.Errorf("expected a=1, got %v", dst["a"])
	}

	bMap := dst["b"].(map[string]any)
	if bMap["c"] != 2 {
		t.Errorf("expected b.c=2, got %v", bMap["c"])
	}
	if bMap["d"] != 4 {
		t.Errorf("expected b.d=4 (overwritten), got %v", bMap["d"])
	}
	if bMap["e"] != 5 {
		t.Errorf("expected b.e=5, got %v", bMap["e"])
	}
	if dst["f"] != 6 {
		t.Errorf("expected f=6, got %v", dst["f"])
	}
}

func TestValidationRuleBuilder_ChainAll(t *testing.T) {
	validator := NewValidationRuleBuilder().
		Required("name", "email", "age").
		Min("age", 0).
		Max("age", 150).
		Regex("email", `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).
		Enum("role", "admin", "user").
		Custom("name", func(val any) error {
			name, ok := val.(string)
			if !ok || len(name) < 2 {
				return errors.New("name must be at least 2 characters")
			}
			return nil
		}).
		Build()

	validData := map[string]any{
		"name":  "John",
		"email": "john@example.com",
		"age":   30,
		"role":  "admin",
	}

	err := validator.Validate(validData)
	if err != nil {
		t.Errorf("unexpected error for valid data: %v", err)
	}

	invalidData := map[string]any{
		"name":  "J",
		"email": "invalid",
		"age":   200,
		"role":  "superadmin",
	}

	err = validator.Validate(invalidData)
	if err == nil {
		t.Error("expected validation error for invalid data")
	}
}

// mockLoader 模拟配置加载器
type mockLoader struct {
	data map[string]any
}

func (m *mockLoader) Load(opts ...LoaderOption) (Config, error) {
	cfg := NewConfig()
	for k, v := range m.data {
		cfg.Set(k, v)
	}
	return cfg, nil
}

func (m *mockLoader) Priority() int {
	return 100
}

func (m *mockLoader) Name() string {
	return "mock"
}

func (m *mockLoader) SupportsWatch() bool {
	return false
}
