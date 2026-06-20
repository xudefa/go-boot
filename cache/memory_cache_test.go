package cache

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemoryCache_Get(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache()

	_, err := c.Get(ctx, "not_exist")
	if err == nil {
		t.Error("expected error for not exist key")
	}
	if !IsNotFound(err) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	err = c.Set(ctx, "key1", "value1", 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err := c.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}
}

func TestMemoryCache_Set(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache()

	err := c.Set(ctx, "key1", "value1", time.Hour)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, _ := c.Get(ctx, "key1")
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}
}

func TestMemoryCache_Del(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache()

	_ = c.Set(ctx, "key1", "value1", 0)
	err := c.Del(ctx, "key1")
	if err != nil {
		t.Fatalf("Del failed: %v", err)
	}

	_, err = c.Get(ctx, "key1")
	if !IsNotFound(err) {
		t.Errorf("expected ErrNotFound after delete")
	}
}

func TestMemoryCache_Exists(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache()

	_ = c.Set(ctx, "key1", "value1", 0)

	exists, err := c.Exists(ctx, "key1")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("expected key1 to exist")
	}

	exists, err = c.Exists(ctx, "key2")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("expected key2 not to exist")
	}
}

func TestMemoryCache_GetMulti(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache()

	_ = c.Set(ctx, "key1", "value1", 0)
	_ = c.Set(ctx, "key2", "value2", 0)

	result, err := c.GetMulti(ctx, []string{"key1", "key2", "key3"})
	if err != nil {
		t.Fatalf("GetMulti failed: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
	if result["key1"] != "value1" {
		t.Errorf("expected value1, got %v", result["key1"])
	}
	if result["key2"] != "value2" {
		t.Errorf("expected value2, got %v", result["key2"])
	}
}

func TestMemoryCache_SetMulti(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache()

	items := map[string]any{
		"key1": "value1",
		"key2": "value2",
	}
	err := c.SetMulti(ctx, items, time.Hour)
	if err != nil {
		t.Fatalf("SetMulti failed: %v", err)
	}

	result, _ := c.GetMulti(ctx, []string{"key1", "key2"})
	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
}

func TestMemoryCache_DeleteMulti(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache()

	_ = c.Set(ctx, "key1", "value1", 0)
	_ = c.Set(ctx, "key2", "value2", 0)

	err := c.DeleteMulti(ctx, []string{"key1", "key2"})
	if err != nil {
		t.Fatalf("DeleteMulti failed: %v", err)
	}

	result, _ := c.GetMulti(ctx, []string{"key1", "key2"})
	if len(result) != 0 {
		t.Errorf("expected 0 items, got %d", len(result))
	}
}

func TestMemoryCache_Clear(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache()

	_ = c.Set(ctx, "key1", "value1", 0)
	_ = c.Set(ctx, "key2", "value2", 0)

	err := c.Clear(ctx)
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	result, _ := c.GetMulti(ctx, []string{"key1", "key2"})
	if len(result) != 0 {
		t.Errorf("expected 0 items after clear, got %d", len(result))
	}
}

func TestMemoryCache_Expiration(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache()

	_ = c.Set(ctx, "key1", "value1", time.Millisecond*50)

	val, err := c.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}

	time.Sleep(100 * time.Millisecond)

	_, err = c.Get(ctx, "key1")
	if !IsCacheMiss(err) {
		t.Errorf("expected ErrCacheMiss after expiration")
	}
}

func TestMemoryCache_GetWithGetter(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache()

	_ = c.Set(ctx, "key1", "value1", 0)

	val, err := c.GetWithGetter(ctx, "key1", func(ctx context.Context, key string) (any, error) {
		return "should not be called", nil
	})
	if err != nil {
		t.Fatalf("GetWithGetter failed: %v", err)
	}
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}
}

func TestMemoryCache_GetWithGetter_Miss(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache()

	var called bool
	val, err := c.GetWithGetter(ctx, "key1", func(ctx context.Context, key string) (any, error) {
		called = true
		return "loaded_value", nil
	})
	if err != nil {
		t.Fatalf("GetWithGetter failed: %v", err)
	}
	if !called {
		t.Error("expected getter to be called")
	}
	if val != "loaded_value" {
		t.Errorf("expected loaded_value, got %v", val)
	}

	val, _ = c.Get(ctx, "key1")
	if val != "loaded_value" {
		t.Errorf("expected cached value, got %v", val)
	}
}

func TestMemoryCache_GetWithGetter_Error(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache()

	testErr := errors.New("getter failed")
	_, err := c.GetWithGetter(ctx, "key1", func(ctx context.Context, key string) (any, error) {
		return nil, testErr
	})
	if err == nil {
		t.Fatal("expected error from getter")
	}
	if !errors.Is(err, testErr) {
		t.Errorf("expected getter error, got %v", err)
	}
}

func TestMemoryCache_GetWithGetter_NilValue(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache()

	val, err := c.GetWithGetter(ctx, "key1", func(ctx context.Context, key string) (any, error) {
		return nil, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != nil {
		t.Errorf("expected nil value, got %v", val)
	}
}

func TestMemoryCache_TTL(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache()

	_, err := c.TTL(ctx, "nonexistent")
	if !IsNotFound(err) {
		t.Errorf("expected ErrNotFound for missing key")
	}

	_ = c.Set(ctx, "permanent", "value", 0)
	ttl, err := c.TTL(ctx, "permanent")
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}
	if ttl != 0 {
		t.Errorf("expected 0 TTL for permanent key, got %v", ttl)
	}
}

func TestMemoryCache_TTL_ExpiredCleansUp(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache()

	_ = c.Set(ctx, "ephemeral", "value", time.Millisecond*20)

	time.Sleep(50 * time.Millisecond)

	_, err := c.TTL(ctx, "ephemeral")
	if !IsCacheMiss(err) {
		t.Errorf("expected ErrCacheMiss for expired key, got %v", err)
	}

	_, err = c.Get(ctx, "ephemeral")
	if !IsNotFound(err) {
		t.Errorf("expected ErrNotFound (key should have been deleted), got %v", err)
	}
}

func TestMemoryCache_Close(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache()

	_ = c.Set(ctx, "key1", "value1", 0)
	err := c.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	result, _ := c.GetMulti(ctx, []string{"key1"})
	if len(result) != 0 {
		t.Errorf("expected 0 items after close, got %d", len(result))
	}
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

func IsCacheMiss(err error) bool {
	return errors.Is(err, ErrCacheMiss)
}
