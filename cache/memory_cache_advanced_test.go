package cache

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestMemoryCache_ConcurrentReadWrite 测试并发读写
func TestMemoryCache_ConcurrentReadWrite(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache()
	var wg sync.WaitGroup
	var readCount int32
	var writeCount int32

	for i := 0; i < 100; i++ {
		wg.Add(2)

		go func(id int) {
			defer wg.Done()
			key := "key" + string(rune('0'+id%10))
			_ = c.Set(ctx, key, "value"+string(rune('0'+id%10)), time.Hour)
			atomic.AddInt32(&writeCount, 1)
		}(i)

		go func(id int) {
			defer wg.Done()
			key := "key" + string(rune('0'+id%10))
			_, err := c.Get(ctx, key)
			if err == nil {
				atomic.AddInt32(&readCount, 1)
			}
		}(i)
	}

	wg.Wait()
	if atomic.LoadInt32(&writeCount) != 100 {
		t.Errorf("Expected 100 writes, got %d", writeCount)
	}
}

// TestMemoryCache_TTLExpiration 测试 TTL 过期
func TestMemoryCache_TTLExpiration(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache()

	err := c.Set(ctx, "expire_key", "expire_value", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err := c.Get(ctx, "expire_key")
	if err != nil {
		t.Fatalf("Get failed before expiration: %v", err)
	}
	if val != "expire_value" {
		t.Errorf("Expected 'expire_value', got %v", val)
	}

	time.Sleep(150 * time.Millisecond)

	_, err = c.Get(ctx, "expire_key")
	if !IsCacheMiss(err) {
		t.Errorf("Expected ErrCacheMiss after expiration, got %v", err)
	}
}

// TestMemoryCache_CachePenetration 测试缓存穿透
func TestMemoryCache_CachePenetration(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache()
	var wg sync.WaitGroup
	var missCount int32

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := "non_existent_key" + string(rune('0'+id%10))
			_, err := c.Get(ctx, key)
			if IsNotFound(err) || IsCacheMiss(err) {
				atomic.AddInt32(&missCount, 1)
			}
		}(i)
	}

	wg.Wait()
	if atomic.LoadInt32(&missCount) != 100 {
		t.Errorf("Expected 100 cache misses, got %d", missCount)
	}
}

// TestMemoryCache_CacheAvalanche 测试缓存雪崩
func TestMemoryCache_CacheAvalanche(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache()

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("avalanche_key%d", i)
		err := c.Set(ctx, key, fmt.Sprintf("value%d", i), 100*time.Millisecond)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}
	}

	time.Sleep(150 * time.Millisecond)

	var wg sync.WaitGroup
	var missCount int32
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("avalanche_key%d", id)
			_, err := c.Get(ctx, key)
			if IsCacheMiss(err) {
				atomic.AddInt32(&missCount, 1)
			}
		}(i)
	}

	wg.Wait()
	if atomic.LoadInt32(&missCount) != 100 {
		t.Errorf("Expected 100 cache misses after avalanche, got %d", missCount)
	}
}
