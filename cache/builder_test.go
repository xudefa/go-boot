package cache

import (
	"context"
	"testing"
	"time"
)

func TestMemoryCacheBuilder_Defaults(t *testing.T) {
	builder := NewMemoryCacheBuilder()

	if builder.initialCapacity != 1024 {
		t.Errorf("expected default initialCapacity 1024, got %d", builder.initialCapacity)
	}
}

func TestMemoryCacheBuilder_ChainConfig(t *testing.T) {
	cache := NewMemoryCacheBuilder().
		InitialCapacity(2048).
		Build()

	if cache == nil {
		t.Fatal("expected non-nil cache")
	}

	if cache.data == nil {
		t.Error("expected non-nil data map")
	}
}

func TestMemoryCacheBuilder_MustBuild(t *testing.T) {
	cache := NewMemoryCacheBuilder().MustBuild()

	if cache == nil {
		t.Fatal("expected non-nil cache")
	}
}

func TestCacheHelper_Get(t *testing.T) {
	memCache := NewMemoryCacheBuilder().Build()
	helper := NewCacheHelper(memCache)

	ctx := context.Background()

	// Set a value
	err := helper.Set(ctx, "test-key", "test-value", 5*time.Minute)
	if err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	// Get the value
	val, err := helper.Get(ctx, "test-key")
	if err != nil {
		t.Fatalf("failed to get cache: %v", err)
	}

	strVal, ok := val.(string)
	if !ok {
		t.Fatalf("expected string type, got %T", val)
	}

	if strVal != "test-value" {
		t.Errorf("expected 'test-value', got %s", strVal)
	}
}

func TestCacheHelper_GetTypeMismatch(t *testing.T) {
	memCache := NewMemoryCacheBuilder().Build()
	helper := NewCacheHelper(memCache)

	ctx := context.Background()

	// Set a string value
	err := helper.Set(ctx, "test-key", "string-value", 5*time.Minute)
	if err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	// Try to get as int - should fail type assertion
	val, err := helper.Get(ctx, "test-key")
	if err != nil {
		t.Fatalf("failed to get cache: %v", err)
	}

	_, ok := val.(int)
	if ok {
		t.Error("expected type assertion to fail")
	}
}

func TestCacheHelper_GetOrSet(t *testing.T) {
	memCache := NewMemoryCacheBuilder().Build()
	helper := NewCacheHelper(memCache)

	ctx := context.Background()

	callCount := 0

	// First call should invoke the function
	val, err := helper.GetOrSet(ctx, "test-key", func() (any, error) {
		callCount++
		return "generated-value", nil
	}, 5*time.Minute)

	if err != nil {
		t.Fatalf("failed to get or set: %v", err)
	}

	strVal, ok := val.(string)
	if !ok {
		t.Fatalf("expected string type, got %T", val)
	}

	if strVal != "generated-value" {
		t.Errorf("expected 'generated-value', got %s", strVal)
	}

	if callCount != 1 {
		t.Errorf("expected callCount 1, got %d", callCount)
	}

	// Second call should use cache
	val, err = helper.GetOrSet(ctx, "test-key", func() (any, error) {
		callCount++
		return "should-not-be-called", nil
	}, 5*time.Minute)

	if err != nil {
		t.Fatalf("failed to get from cache: %v", err)
	}

	strVal, ok = val.(string)
	if !ok {
		t.Fatalf("expected string type, got %T", val)
	}

	if strVal != "generated-value" {
		t.Errorf("expected 'generated-value' from cache, got %s", strVal)
	}

	if callCount != 1 {
		t.Errorf("expected callCount still 1 (cached), got %d", callCount)
	}
}

func TestCacheHelper_GetOrSet_FunctionError(t *testing.T) {
	memCache := NewMemoryCacheBuilder().Build()
	helper := NewCacheHelper(memCache)

	ctx := context.Background()

	_, err := helper.GetOrSet(ctx, "test-key", func() (any, error) {
		return nil, nil
	}, 5*time.Minute)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCacheHelper_Invalidate(t *testing.T) {
	memCache := NewMemoryCacheBuilder().Build()
	helper := NewCacheHelper(memCache)

	ctx := context.Background()

	// Set a value
	err := helper.Set(ctx, "test-key", "test-value", 5*time.Minute)
	if err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	// Invalidate
	err = helper.Invalidate(ctx, "test-key")
	if err != nil {
		t.Fatalf("failed to invalidate: %v", err)
	}

	// Should not exist
	exists, err := helper.Exists(ctx, "test-key")
	if err != nil {
		t.Fatalf("failed to check exists: %v", err)
	}

	if exists {
		t.Error("expected key to not exist after invalidation")
	}
}

func TestCacheHelper_InvalidateAll(t *testing.T) {
	memCache := NewMemoryCacheBuilder().Build()
	helper := NewCacheHelper(memCache)

	ctx := context.Background()

	// Set multiple values
	_ = helper.Set(ctx, "key1", "value1", 5*time.Minute)
	_ = helper.Set(ctx, "key2", "value2", 5*time.Minute)
	_ = helper.Set(ctx, "key3", "value3", 5*time.Minute)

	// Invalidate all
	err := helper.InvalidateAll(ctx, "key1", "key2")
	if err != nil {
		t.Fatalf("failed to invalidate all: %v", err)
	}

	// key1 and key2 should not exist
	exists1, _ := helper.Exists(ctx, "key1")
	exists2, _ := helper.Exists(ctx, "key2")
	exists3, _ := helper.Exists(ctx, "key3")

	if exists1 {
		t.Error("expected key1 to not exist")
	}

	if exists2 {
		t.Error("expected key2 to not exist")
	}

	if !exists3 {
		t.Error("expected key3 to still exist")
	}
}

func TestCacheHelper_Clear(t *testing.T) {
	memCache := NewMemoryCacheBuilder().Build()
	helper := NewCacheHelper(memCache)

	ctx := context.Background()

	// Set multiple values
	_ = helper.Set(ctx, "key1", "value1", 5*time.Minute)
	_ = helper.Set(ctx, "key2", "value2", 5*time.Minute)

	// Clear
	err := helper.Clear(ctx)
	if err != nil {
		t.Fatalf("failed to clear: %v", err)
	}

	// All keys should not exist
	exists1, _ := helper.Exists(ctx, "key1")
	exists2, _ := helper.Exists(ctx, "key2")

	if exists1 {
		t.Error("expected key1 to not exist after clear")
	}

	if exists2 {
		t.Error("expected key2 to not exist after clear")
	}
}

func TestCacheHelper_TTL(t *testing.T) {
	memCache := NewMemoryCacheBuilder().Build()
	helper := NewCacheHelper(memCache)

	ctx := context.Background()

	// Set with TTL
	err := helper.Set(ctx, "test-key", "test-value", 10*time.Minute)
	if err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	// Get TTL
	ttl, err := helper.TTL(ctx, "test-key")
	if err != nil {
		t.Fatalf("failed to get TTL: %v", err)
	}

	if ttl <= 0 || ttl > 10*time.Minute {
		t.Errorf("expected TTL between 0 and 10m, got %v", ttl)
	}
}

func TestCacheTemplate_Key(t *testing.T) {
	memCache := NewMemoryCacheBuilder().Build()
	template := NewCacheTemplate(memCache, "myapp")

	key := template.Key("user:123")
	if key != "myapp:user:123" {
		t.Errorf("expected key 'myapp:user:123', got %s", key)
	}
}

func TestCacheTemplate_KeyNoPrefix(t *testing.T) {
	memCache := NewMemoryCacheBuilder().Build()
	template := NewCacheTemplate(memCache, "")

	key := template.Key("user:123")
	if key != "user:123" {
		t.Errorf("expected key 'user:123', got %s", key)
	}
}

func TestCacheTemplate_GetSet(t *testing.T) {
	memCache := NewMemoryCacheBuilder().Build()
	template := NewCacheTemplate(memCache, "myapp")

	ctx := context.Background()

	// Set
	err := template.Set(ctx, "test-key", "test-value", 5*time.Minute)
	if err != nil {
		t.Fatalf("failed to set: %v", err)
	}

	// Get
	val, err := template.Get(ctx, "test-key")
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}

	if val != "test-value" {
		t.Errorf("expected 'test-value', got %v", val)
	}
}

func TestCacheTemplate_Del(t *testing.T) {
	memCache := NewMemoryCacheBuilder().Build()
	template := NewCacheTemplate(memCache, "myapp")

	ctx := context.Background()

	// Set
	_ = template.Set(ctx, "test-key", "test-value", 5*time.Minute)

	// Del
	err := template.Del(ctx, "test-key")
	if err != nil {
		t.Fatalf("failed to del: %v", err)
	}

	// Should not exist
	exists, err := template.Exists(ctx, "test-key")
	if err != nil {
		t.Fatalf("failed to check exists: %v", err)
	}

	if exists {
		t.Error("expected key to not exist after deletion")
	}
}

func TestCacheTemplate_GetOrSet(t *testing.T) {
	memCache := NewMemoryCacheBuilder().Build()
	template := NewCacheTemplate(memCache, "myapp")

	ctx := context.Background()

	callCount := 0

	// First call
	val, err := template.GetOrSet(ctx, "test-key", func() (any, error) {
		callCount++
		return "generated", nil
	}, 5*time.Minute)

	if err != nil {
		t.Fatalf("failed to get or set: %v", err)
	}

	if val != "generated" {
		t.Errorf("expected 'generated', got %v", val)
	}

	if callCount != 1 {
		t.Errorf("expected callCount 1, got %d", callCount)
	}

	// Second call should use cache
	val, err = template.GetOrSet(ctx, "test-key", func() (any, error) {
		callCount++
		return "should-not-be-called", nil
	}, 5*time.Minute)

	if err != nil {
		t.Fatalf("failed to get from cache: %v", err)
	}

	if val != "generated" {
		t.Errorf("expected 'generated' from cache, got %v", val)
	}

	if callCount != 1 {
		t.Errorf("expected callCount still 1, got %d", callCount)
	}
}

func TestCacheConfig_Default(t *testing.T) {
	config := DefaultCacheConfig()

	if !config.Enabled {
		t.Error("expected Enabled to be true")
	}

	if config.DefaultTTL != 30*time.Minute {
		t.Errorf("expected DefaultTTL 30m, got %v", config.DefaultTTL)
	}

	if config.MaxSize != 10000 {
		t.Errorf("expected MaxSize 10000, got %d", config.MaxSize)
	}

	if config.StatsEnabled {
		t.Error("expected StatsEnabled to be false")
	}
}

func TestCacheConfig_ApplyOptions(t *testing.T) {
	config := DefaultCacheConfig()

	config.ApplyOptions([]CacheOption{
		WithCacheEnabled(false),
		WithDefaultTTL(1 * time.Hour),
		WithMaxSize(5000),
		WithKeyPrefix("test"),
		WithStatsEnabled(true),
	})

	if config.Enabled {
		t.Error("expected Enabled to be false")
	}

	if config.DefaultTTL != 1*time.Hour {
		t.Errorf("expected DefaultTTL 1h, got %v", config.DefaultTTL)
	}

	if config.MaxSize != 5000 {
		t.Errorf("expected MaxSize 5000, got %d", config.MaxSize)
	}

	if config.KeyPrefix != "test" {
		t.Errorf("expected KeyPrefix 'test', got %s", config.KeyPrefix)
	}

	if !config.StatsEnabled {
		t.Error("expected StatsEnabled to be true")
	}
}

func TestCacheHelper_GetNotFound(t *testing.T) {
	memCache := NewMemoryCacheBuilder().Build()
	helper := NewCacheHelper(memCache)

	ctx := context.Background()

	_, err := helper.Get(ctx, "non-existent-key")
	if err == nil {
		t.Error("expected error for non-existent key")
	}
}

func TestCacheTemplate_TTL(t *testing.T) {
	memCache := NewMemoryCacheBuilder().Build()
	template := NewCacheTemplate(memCache, "myapp")

	ctx := context.Background()

	// Set with TTL
	_ = template.Set(ctx, "test-key", "test-value", 10*time.Minute)

	// Get TTL
	ttl, err := template.TTL(ctx, "test-key")
	if err != nil {
		t.Fatalf("failed to get TTL: %v", err)
	}

	if ttl <= 0 || ttl > 10*time.Minute {
		t.Errorf("expected TTL between 0 and 10m, got %v", ttl)
	}
}

func TestCacheHelper_GetOrSetWithNilResult(t *testing.T) {
	memCache := NewMemoryCacheBuilder().Build()
	helper := NewCacheHelper(memCache)

	ctx := context.Background()

	// Test with nil result
	val, err := helper.GetOrSet(ctx, "nil-key", func() (any, error) {
		return nil, nil
	}, 5*time.Minute)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val != nil {
		t.Errorf("expected nil value, got %v", val)
	}
}
