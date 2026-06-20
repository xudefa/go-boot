// Package cache 提供缓存值获取函数类型。
//
// Getter 函数用于缓存旁路模式（Cache-Aside Pattern），
// 在缓存未命中时从数据源加载值。
package cache

import "context"

// Getter 缓存旁路模式的值加载函数
//
// 在缓存未命中时调用，从数据源加载值。
// 返回 nil 值不会被缓存。
type Getter func(ctx context.Context, key string) (any, error)
