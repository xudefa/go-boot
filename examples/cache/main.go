// 示例: go-boot/cache 缓存使用指南
//
// 本示例演示 go-boot/cache 包的核心功能:
//
// 1. 使用内存缓存
// 2. 使用Redis缓存
// 3. 缓存基本操作: Get/Set/Delete/Exists
// 4. 批量操作: GetMulti/SetMulti/DeleteMulti
// 5. 使用GetWithGetter实现缓存穿透
//
// 运行方式:
//
//	cd examples/cache && go run .
//
// 注意:
//   - 内存缓存不需要额外依赖
//   - Redis缓存需要本地Redis运行在localhost:6379
package main

import (
	"context"
	"fmt"
	"github.com/xudefa/go-boot/cache"
	"log"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	fmt.Println("=== Cache Example ===")

	if err := exampleMemoryCache(); err != nil {
		return err
	}

	fmt.Println()

	if err := exampleRedisCache(); err != nil {
		fmt.Printf("Redis not available, skipping: %v\n", err)
	}

	fmt.Println("=== Cache Example ===")
	return nil
}

func exampleMemoryCache() error {
	fmt.Println("--- Memory Cache ---")

	ctx := context.Background()
	memCache := cache.NewMemoryCache()

	err := memCache.Set(ctx, "user:1", map[string]any{
		"name":  "John",
		"email": "john@example.com",
	}, time.Hour)
	if err != nil {
		return fmt.Errorf("set failed: %w", err)
	}

	val, err := memCache.Get(ctx, "user:1")
	if err != nil {
		return fmt.Errorf("get failed: %w", err)
	}
	fmt.Printf("Got user: %v\n", val)

	exists, err := memCache.Exists(ctx, "user:1")
	if err != nil {
		return fmt.Errorf("exists failed: %w", err)
	}
	fmt.Printf("Key exists: %v\n", exists)

	items := map[string]any{
		"user:2": map[string]any{"name": "Jane", "email": "jane@example.com"},
		"user:3": map[string]any{"name": "Bob", "email": "bob@example.com"},
	}
	err = memCache.SetMulti(ctx, items, time.Hour)
	if err != nil {
		return fmt.Errorf("setmulti failed: %w", err)
	}

	result, err := memCache.GetMulti(ctx, []string{"user:1", "user:2", "user:3"})
	if err != nil {
		return fmt.Errorf("getmulti failed: %w", err)
	}
	fmt.Printf("Got %d users\n", len(result))

	err = memCache.Delete(ctx, "user:1")
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	fmt.Println("Deleted user:1")

	return nil
}

func exampleRedisCache() error {
	fmt.Println("--- Redis Cache ---")

	ctx := context.Background()
	redisCache, err := cache.NewRedisCache(&cache.RedisConfig{
		Address:  "127.0.0.1",
		Port:     6379,
		Username: "",
		Password: "",
		Prefix:   "",
		Timeout:  time.Second * 5,
	})
	if err != nil {
		return fmt.Errorf("create redis cache failed: %w", err)
	}
	defer redisCache.Close()

	err = redisCache.Set(ctx, "user:1", map[string]any{
		"name":  "John",
		"email": "john@example.com",
	}, time.Hour)
	if err != nil {
		return fmt.Errorf("set failed: %w", err)
	}

	val, err := redisCache.Get(ctx, "user:1")
	if err != nil {
		return fmt.Errorf("get failed: %w", err)
	}
	fmt.Printf("Got user: %v\n", val)

	err = redisCache.Delete(ctx, "user:1")
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	fmt.Println("Deleted user:1")

	return nil
}
