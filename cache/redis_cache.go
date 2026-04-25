package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	prefix string
}

func NewRedisCache(prefix string, expireAt time.Duration, client *redis.Client) (*RedisCache, error) {
	ctx, cancel := context.WithTimeout(context.Background(), expireAt)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}
	return &RedisCache{
		client: client,
		prefix: prefix,
	}, nil
}

func (c *RedisCache) fullKey(key string) string {
	return c.prefix + key
}

func (c *RedisCache) Get(ctx context.Context, key string) (any, error) {
	val, err := c.client.Get(ctx, c.fullKey(key)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var result any
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if ttl > 0 {
		return c.client.Set(ctx, c.fullKey(key), data, ttl).Err()
	}
	return c.client.Set(ctx, c.fullKey(key), data, 0).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, c.fullKey(key)).Err()
}

func (c *RedisCache) Exists(ctx context.Context, keys ...string) (bool, error) {
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = c.fullKey(key)
	}
	n, err := c.client.Exists(ctx, fullKeys...).Result()
	if err != nil {
		return false, err
	}
	return n == int64(len(keys)), nil
}

func (c *RedisCache) GetMulti(ctx context.Context, keys []string) (map[string]any, error) {
	if len(keys) == 0 {
		return make(map[string]any), nil
	}
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = c.fullKey(key)
	}
	values, err := c.client.MGet(ctx, fullKeys...).Result()
	if err != nil {
		return nil, err
	}
	result := make(map[string]any)
	for i, val := range values {
		if val == nil {
			continue
		}
		str, ok := val.(string)
		if !ok {
			continue
		}
		var item any
		if err := json.Unmarshal([]byte(str), &item); err != nil {
			continue
		}
		result[keys[i]] = item
	}
	return result, nil
}

func (c *RedisCache) SetMulti(ctx context.Context, items map[string]any, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}
	pipe := c.client.Pipeline()
	for key, value := range items {
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		if ttl > 0 {
			pipe.Set(ctx, c.fullKey(key), data, ttl)
		} else {
			pipe.Set(ctx, c.fullKey(key), data, 0)
		}
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (c *RedisCache) DeleteMulti(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = c.fullKey(key)
	}
	return c.client.Del(ctx, fullKeys...).Err()
}

func (c *RedisCache) Clear(ctx context.Context) error {
	keys, err := c.client.Keys(ctx, c.prefix+"*").Result()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

func (c *RedisCache) GetWithGetter(ctx context.Context, key string, fn Getter) (any, error) {
	val, err := c.Get(ctx, key)
	if err == nil {
		return val, nil
	}
	if !errors.Is(err, ErrNotFound) && !errors.Is(err, ErrCacheMiss) {
		return nil, err
	}

	val, err = fn(ctx, key)
	if err != nil {
		return nil, err
	}
	if val != nil {
		_ = c.Set(ctx, key, val, 0)
	}
	return val, nil
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

var _ Cache = (*RedisCache)(nil)
