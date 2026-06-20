package environment

import (
	"os"
	"testing"
)

func TestEnvironmentBuilder_Defaults(t *testing.T) {
	builder := NewEnvironmentBuilder()

	if builder.profiles == nil {
		t.Error("expected non-nil profiles")
	}

	if builder.propertySources == nil {
		t.Error("expected non-nil propertySources")
	}
}

func TestEnvironmentBuilder_WithProfile(t *testing.T) {
	env := NewEnvironmentBuilder().
		WithProfile("dev").
		Build()

	if env == nil {
		t.Fatal("expected non-nil env")
	}

	profiles := env.GetActiveProfiles()
	if len(profiles) < 1 {
		t.Error("expected at least 1 profile")
	}

	found := false
	for _, p := range profiles {
		if p == "dev" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected 'dev' profile to be active")
	}
}

func TestEnvironmentBuilder_WithProfiles(t *testing.T) {
	env := NewEnvironmentBuilder().
		WithProfiles("dev", "test").
		Build()

	if env == nil {
		t.Fatal("expected non-nil env")
	}

	profiles := env.GetActiveProfiles()

	devFound := false
	testFound := false
	for _, p := range profiles {
		if p == "dev" {
			devFound = true
		}
		if p == "test" {
			testFound = true
		}
	}

	if !devFound {
		t.Error("expected 'dev' profile to be active")
	}

	if !testFound {
		t.Error("expected 'test' profile to be active")
	}
}

func TestEnvironmentBuilder_WithPropertySource(t *testing.T) {
	source := NewMapPropertySource("test-source", PriorityNormal, map[string]any{
		"test.key": "test-value",
	})

	env := NewEnvironmentBuilder().
		WithPropertySource(source).
		Build()

	if env == nil {
		t.Fatal("expected non-nil env")
	}

	val, ok := env.GetProperty("test.key")
	if !ok {
		t.Fatal("expected property to exist")
	}

	if val != "test-value" {
		t.Errorf("expected 'test-value', got %v", val)
	}
}

func TestEnvironmentBuilder_WithPropertySourceFirst(t *testing.T) {
	source1 := NewMapPropertySource("source1", PriorityLow, map[string]any{
		"test.key": "value1",
	})

	source2 := NewMapPropertySource("source2", PriorityHigh, map[string]any{
		"test.key": "value2",
	})

	env := NewEnvironmentBuilder().
		WithPropertySource(source1).
		WithPropertySourceFirst(source2).
		Build()

	if env == nil {
		t.Fatal("expected non-nil env")
	}

	val, ok := env.GetProperty("test.key")
	if !ok {
		t.Fatal("expected property to exist")
	}

	// source2 has higher priority (PriorityHigh > PriorityLow)
	if val != "value2" {
		t.Errorf("expected 'value2' (higher priority), got %v", val)
	}
}

func TestEnvironmentBuilder_WithEnvPrefix(t *testing.T) {
	// Set an environment variable
	_ = os.Setenv("TEST_BUILDER_KEY", "test-env-value")
	defer func() { _ = os.Unsetenv("TEST_BUILDER_KEY") }()

	env := NewEnvironmentBuilder().
		WithEnvPrefix("TEST_BUILDER").
		Build()

	if env == nil {
		t.Fatal("expected non-nil env")
	}

	// EnvPropertySource converts "key" to "TEST_BUILDER_KEY"
	val, ok := env.GetProperty("key")
	if !ok {
		t.Fatal("expected property to exist")
	}

	if val != "test-env-value" {
		t.Errorf("expected 'test-env-value', got %v", val)
	}
}

func TestEnvironmentBuilder_WithArgs(t *testing.T) {
	env := NewEnvironmentBuilder().
		WithArgs("--app.name=test-app", "--app.port=8080").
		Build()

	if env == nil {
		t.Fatal("expected non-nil env")
	}

	val, ok := env.GetProperty("app.name")
	if !ok {
		t.Fatal("expected property to exist")
	}

	if val != "test-app" {
		t.Errorf("expected 'test-app', got %v", val)
	}
}

func TestEnvironmentBuilder_ChainConfig(t *testing.T) {
	_ = os.Setenv("CHAIN_TEST_KEY", "chain-value")
	defer func() { _ = os.Unsetenv("CHAIN_TEST_KEY") }()

	env := NewEnvironmentBuilder().
		WithProfiles("dev", "test").
		WithEnvPrefix("CHAIN_TEST").
		WithArgs("--custom.key=custom-value").
		Build()

	if env == nil {
		t.Fatal("expected non-nil env")
	}

	// Check profiles
	profiles := env.GetActiveProfiles()
	if len(profiles) < 2 {
		t.Errorf("expected at least 2 profiles, got %d", len(profiles))
	}

	// Check env variable - EnvPropertySource converts "key" to "CHAIN_TEST_KEY"
	val1, ok := env.GetProperty("key")
	if !ok || val1 != "chain-value" {
		t.Errorf("expected 'chain-value', got %v", val1)
	}

	// Check args
	val2, ok := env.GetProperty("custom.key")
	if !ok || val2 != "custom-value" {
		t.Errorf("expected 'custom-value', got %v", val2)
	}
}

func TestEnvironmentBuilder_MustBuild(t *testing.T) {
	env := NewEnvironmentBuilder().MustBuild()

	if env == nil {
		t.Fatal("expected non-nil env")
	}
}

func TestEnvironmentHelper_GetString(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.name": "test-app",
	}))

	helper := NewEnvironmentHelper(env)

	val := helper.GetString("app.name", "default")
	if val != "test-app" {
		t.Errorf("expected 'test-app', got %s", val)
	}

	val = helper.GetString("non.existent", "default")
	if val != "default" {
		t.Errorf("expected 'default', got %s", val)
	}
}

func TestEnvironmentHelper_WithPrefix(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"server.host": "localhost",
		"server.port": 8080,
	}))

	helper := NewEnvironmentHelper(env).WithPrefix("server")

	host := helper.GetString("host", "default")
	if host != "localhost" {
		t.Errorf("expected 'localhost', got %s", host)
	}

	port := helper.GetInt("port", 0)
	if port != 8080 {
		t.Errorf("expected 8080, got %d", port)
	}
}

func TestEnvironmentHelper_GetInt(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.port": 9090,
	}))

	helper := NewEnvironmentHelper(env)

	val := helper.GetInt("app.port", 0)
	if val != 9090 {
		t.Errorf("expected 9090, got %d", val)
	}

	val = helper.GetInt("non.existent", 1234)
	if val != 1234 {
		t.Errorf("expected 1234, got %d", val)
	}
}

func TestEnvironmentHelper_GetBool(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.debug": true,
	}))

	helper := NewEnvironmentHelper(env)

	val := helper.GetBool("app.debug", false)
	if !val {
		t.Error("expected true")
	}

	val = helper.GetBool("non.existent", true)
	if !val {
		t.Error("expected true (default)")
	}
}

func TestEnvironmentHelper_GetFloat64(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.ratio": 3.14,
	}))

	helper := NewEnvironmentHelper(env)

	val := helper.GetFloat64("app.ratio", 0.0)
	if val != 3.14 {
		t.Errorf("expected 3.14, got %f", val)
	}

	val = helper.GetFloat64("non.existent", 1.5)
	if val != 1.5 {
		t.Errorf("expected 1.5, got %f", val)
	}
}

func TestEnvironmentHelper_ContainsProperty(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.name": "test",
	}))

	helper := NewEnvironmentHelper(env)

	if !helper.ContainsProperty("app.name") {
		t.Error("expected property to exist")
	}

	if helper.ContainsProperty("non.existent") {
		t.Error("expected property to not exist")
	}
}

func TestEnvironmentHelper_IsDev(t *testing.T) {
	env := NewEnvironment()
	env.AddActiveProfile("dev")

	helper := NewEnvironmentHelper(env)

	if !helper.IsDev() {
		t.Error("expected IsDev to return true")
	}

	if helper.IsProd() {
		t.Error("expected IsProd to return false")
	}
}

func TestEnvironmentHelper_IsProd(t *testing.T) {
	env := NewEnvironment()
	env.AddActiveProfile("prod")

	helper := NewEnvironmentHelper(env)

	if !helper.IsProd() {
		t.Error("expected IsProd to return true")
	}

	if helper.IsDev() {
		t.Error("expected IsDev to return false")
	}
}

func TestEnvironmentHelper_GetActiveProfile(t *testing.T) {
	env := NewEnvironment()
	env.AddActiveProfile("test")

	helper := NewEnvironmentHelper(env)

	profile := helper.GetActiveProfile()
	if profile != "test" {
		t.Errorf("expected 'test', got %s", profile)
	}
}

func TestEnvironmentHelper_GetActiveProfile_Default(t *testing.T) {
	env := NewEnvironment()

	helper := NewEnvironmentHelper(env)

	profile := helper.GetActiveProfile()
	if profile != "default" {
		t.Errorf("expected 'default', got %s", profile)
	}
}

func TestEnvironmentTemplate_GetDatabaseURL(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"database.url": "mysql://localhost:3306/testdb",
	}))

	template := NewEnvironmentTemplate(env)

	val := template.GetDatabaseURL("default-url")
	if val != "mysql://localhost:3306/testdb" {
		t.Errorf("expected 'mysql://localhost:3306/testdb', got %s", val)
	}

	val = template.GetDatabaseURL("default-url")
	if val != "mysql://localhost:3306/testdb" {
		t.Errorf("expected 'mysql://localhost:3306/testdb', got %s", val)
	}
}

func TestEnvironmentTemplate_GetServerPort(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"server.port": 9090,
	}))

	template := NewEnvironmentTemplate(env)

	val := template.GetServerPort(8080)
	if val != 9090 {
		t.Errorf("expected 9090, got %d", val)
	}
}

func TestEnvironmentTemplate_IsDebugMode(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"debug": true,
	}))

	template := NewEnvironmentTemplate(env)

	if !template.IsDebugMode() {
		t.Error("expected debug mode to be true")
	}
}

func TestEnvironmentTemplate_IsVerbose(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"verbose": true,
	}))

	template := NewEnvironmentTemplate(env)

	if !template.IsVerbose() {
		t.Error("expected verbose to be true")
	}
}

func TestEnvironmentConfig_Default(t *testing.T) {
	config := DefaultEnvironmentConfig()

	if config.DefaultProfile != "default" {
		t.Errorf("expected DefaultProfile 'default', got %s", config.DefaultProfile)
	}

	if !config.AutoDetectProfiles {
		t.Error("expected AutoDetectProfiles to be true")
	}
}

func TestEnvironmentConfig_ApplyOptions(t *testing.T) {
	config := DefaultEnvironmentConfig()

	config.ApplyOptions([]EnvironmentOption{
		WithProfiles("dev", "test"),
		WithDefaultProfile("custom"),
		WithAutoDetectProfiles(false),
	})

	if len(config.Profiles) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(config.Profiles))
	}

	if config.DefaultProfile != "custom" {
		t.Errorf("expected DefaultProfile 'custom', got %s", config.DefaultProfile)
	}

	if config.AutoDetectProfiles {
		t.Error("expected AutoDetectProfiles to be false")
	}
}

func TestCreateEnvironment(t *testing.T) {
	source := NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.name": "test-app",
	})

	env := CreateEnvironment(
		WithProfiles("dev"),
		WithPropertySources(source),
		WithDefaultProfile("custom"),
	)

	if env == nil {
		t.Fatal("expected non-nil env")
	}

	// Check property
	val, ok := env.GetProperty("app.name")
	if !ok || val != "test-app" {
		t.Errorf("expected 'test-app', got %v", val)
	}

	// Check profiles
	profiles := env.GetActiveProfiles()

	customFound := false
	devFound := false
	for _, p := range profiles {
		if p == "custom" {
			customFound = true
		}
		if p == "dev" {
			devFound = true
		}
	}

	if !customFound {
		t.Error("expected 'custom' profile to be active")
	}

	if !devFound {
		t.Error("expected 'dev' profile to be active")
	}
}

func TestEnvironmentHelper_GetRequiredProperty(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.name": "test-app",
	}))

	helper := NewEnvironmentHelper(env)

	val, err := helper.GetRequiredProperty("app.name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val != "test-app" {
		t.Errorf("expected 'test-app', got %v", val)
	}

	_, err = helper.GetRequiredProperty("non.existent")
	if err == nil {
		t.Error("expected error for non-existent property")
	}
}

func TestEnvironmentHelper_IsTest(t *testing.T) {
	env := NewEnvironment()
	env.AddActiveProfile("test")

	helper := NewEnvironmentHelper(env)

	if !helper.IsTest() {
		t.Error("expected IsTest to return true")
	}
}

func TestEnvironmentTemplate_GetDatabaseHost(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"database.host": "db.example.com",
	}))

	template := NewEnvironmentTemplate(env)

	val := template.GetDatabaseHost("default-host")
	if val != "db.example.com" {
		t.Errorf("expected 'db.example.com', got %s", val)
	}
}

func TestEnvironmentTemplate_GetDatabasePort(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"database.port": 5432,
	}))

	template := NewEnvironmentTemplate(env)

	val := template.GetDatabasePort(3306)
	if val != 5432 {
		t.Errorf("expected 5432, got %d", val)
	}
}

func TestEnvironmentTemplate_GetDatabaseName(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"database.name": "mydb",
	}))

	template := NewEnvironmentTemplate(env)

	val := template.GetDatabaseName("default-db")
	if val != "mydb" {
		t.Errorf("expected 'mydb', got %s", val)
	}
}

func TestEnvironmentTemplate_GetServerHost(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"server.host": "0.0.0.0",
	}))

	template := NewEnvironmentTemplate(env)

	val := template.GetServerHost("localhost")
	if val != "0.0.0.0" {
		t.Errorf("expected '0.0.0.0', got %s", val)
	}
}

func TestEnvironmentTemplate_GetLogLevel(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"log.level": "DEBUG",
	}))

	template := NewEnvironmentTemplate(env)

	val := template.GetLogLevel("INFO")
	if val != "DEBUG" {
		t.Errorf("expected 'DEBUG', got %s", val)
	}
}

func TestEnvironmentTemplate_GetRedisHost(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"redis.host": "redis.example.com",
	}))

	template := NewEnvironmentTemplate(env)

	val := template.GetRedisHost("localhost")
	if val != "redis.example.com" {
		t.Errorf("expected 'redis.example.com', got %s", val)
	}
}

func TestEnvironmentTemplate_GetRedisPort(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"redis.port": 6380,
	}))

	template := NewEnvironmentTemplate(env)

	val := template.GetRedisPort(6379)
	if val != 6380 {
		t.Errorf("expected 6380, got %d", val)
	}
}

func TestEnvironmentTemplate_GetRedisPassword(t *testing.T) {
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"redis.password": "secret",
	}))

	template := NewEnvironmentTemplate(env)

	val := template.GetRedisPassword("")
	if val != "secret" {
		t.Errorf("expected 'secret', got %s", val)
	}
}
