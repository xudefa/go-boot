package cache

import (
	"context"
	"testing"
	"time"
)

func TestLRUCache_Basic(t *testing.T) {
	cache := NewLRUCache(10)
	ctx := context.Background()

	err := cache.Set(ctx, "key1", "value1", time.Minute)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err := cache.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}
}

func TestLRUCache_NotFound(t *testing.T) {
	cache := NewLRUCache(10)
	ctx := context.Background()

	_, err := cache.Get(ctx, "nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestLRUCache_TTL(t *testing.T) {
	cache := NewLRUCache(10)
	ctx := context.Background()

	err := cache.Set(ctx, "key1", "value1", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 立即获取应该成功
	val, err := cache.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	_, err = cache.Get(ctx, "key1")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after TTL, got %v", err)
	}
}

func TestLRUCache_Eviction(t *testing.T) {
	cache := NewLRUCache(3)
	ctx := context.Background()

	// 添加 3 个项
	_ = cache.Set(ctx, "key1", "value1", time.Minute)
	_ = cache.Set(ctx, "key2", "value2", time.Minute)
	_ = cache.Set(ctx, "key3", "value3", time.Minute)

	// 添加第 4 个项，应该淘汰最久未使用的 key1
	_ = cache.Set(ctx, "key4", "value4", time.Minute)

	_, err := cache.Get(ctx, "key1")
	if err != ErrNotFound {
		t.Errorf("expected key1 to be evicted")
	}

	// key2, key3, key4 应该还在
	for _, key := range []string{"key2", "key3", "key4"} {
		_, err := cache.Get(ctx, key)
		if err != nil {
			t.Errorf("expected %s to exist", key)
		}
	}
}

func TestLRUCache_LRUOrder(t *testing.T) {
	cache := NewLRUCache(3)
	ctx := context.Background()

	_ = cache.Set(ctx, "key1", "value1", time.Minute)
	_ = cache.Set(ctx, "key2", "value2", time.Minute)
	_ = cache.Set(ctx, "key3", "value3", time.Minute)

	// 访问 key1，使其变为最近使用
	_, _ = cache.Get(ctx, "key1")

	// 添加新项，应该淘汰 key2（最久未使用）
	_ = cache.Set(ctx, "key4", "value4", time.Minute)

	_, err := cache.Get(ctx, "key2")
	if err != ErrNotFound {
		t.Errorf("expected key2 to be evicted, key1 should be kept")
	}
}

func TestLRUCache_Del(t *testing.T) {
	cache := NewLRUCache(10)
	ctx := context.Background()

	_ = cache.Set(ctx, "key1", "value1", time.Minute)
	_ = cache.Set(ctx, "key2", "value2", time.Minute)

	err := cache.Del(ctx, "key1")
	if err != nil {
		t.Fatalf("Del failed: %v", err)
	}

	_, err = cache.Get(ctx, "key1")
	if err != ErrNotFound {
		t.Errorf("expected key1 to be deleted")
	}

	// key2 应该还在
	val, err := cache.Get(ctx, "key2")
	if err != nil {
		t.Fatalf("Get key2 failed: %v", err)
	}
	if val != "value2" {
		t.Errorf("expected value2, got %v", val)
	}
}

func TestLRUCache_Exists(t *testing.T) {
	cache := NewLRUCache(10)
	ctx := context.Background()

	_ = cache.Set(ctx, "key1", "value1", 100*time.Millisecond)

	exists, err := cache.Exists(ctx, "key1")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Errorf("expected key1 to exist")
	}

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	exists, err = cache.Exists(ctx, "key1")
	if err != nil {
		t.Fatalf("Exists failed after TTL: %v", err)
	}
	if exists {
		t.Errorf("expected key1 to be expired")
	}
}

func TestLRUCache_TTLRemaining(t *testing.T) {
	cache := NewLRUCache(10)
	ctx := context.Background()

	_ = cache.Set(ctx, "key1", "value1", time.Second)

	remaining, err := cache.TTL(ctx, "key1")
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}
	if remaining > time.Second || remaining <= 0 {
		t.Errorf("expected TTL around 1s, got %v", remaining)
	}
}

func TestLRUCache_Clear(t *testing.T) {
	cache := NewLRUCache(10)
	ctx := context.Background()

	_ = cache.Set(ctx, "key1", "value1", time.Minute)
	_ = cache.Set(ctx, "key2", "value2", time.Minute)

	cache.Clear()

	if cache.Len() != 0 {
		t.Errorf("expected cache to be empty after clear")
	}
}

func TestLRUCache_EvictCallback(t *testing.T) {
	evicted := make(map[string]any)
	cache := NewLRUCache(2, WithEvictCallback(func(key string, value any) {
		evicted[key] = value
	}))
	ctx := context.Background()

	_ = cache.Set(ctx, "key1", "value1", time.Minute)
	_ = cache.Set(ctx, "key2", "value2", time.Minute)
	_ = cache.Set(ctx, "key3", "value3", time.Minute) // 应该淘汰 key1

	if val, ok := evicted["key1"]; !ok || val != "value1" {
		t.Errorf("expected key1 to be evicted with value1")
	}
}

func TestLRUCache_ConcurrentAccess(t *testing.T) {
	cache := NewLRUCache(100)
	ctx := context.Background()

	// 并发写入
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 10; j++ {
				key := string(rune('A'+n)) + string(rune('0'+j))
				_ = cache.Set(ctx, key, j, time.Minute)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证没有 panic
	if cache.Len() == 0 {
		t.Error("expected cache to have items")
	}
}
