package exception

import (
	"context"
	"sort"
)

// ResolverChain 解析器链
//
// ResolverChain 管理多个 ExceptionResolver，并按优先级顺序执行它们。
// 当处理异常时，会遍历解析器链，找到第一个支持该异常的解析器进行处理。
type ResolverChain struct {
	resolvers []ExceptionResolver
}

// NewResolverChain 创建新的解析器链
//
// 创建一个空的解析器链，可以通过 AddResolver 方法添加解析器。
func NewResolverChain() *ResolverChain {
	return &ResolverChain{
		resolvers: make([]ExceptionResolver, 0),
	}
}

// AddResolver 添加解析器（自动按 Order 排序）
//
// 将解析器添加到链中，并自动按照 Order 值进行排序。
// Order 值越小，优先级越高，会先被调用。
func (c *ResolverChain) AddResolver(resolver ExceptionResolver) {
	c.resolvers = append(c.resolvers, resolver)
	sort.Slice(c.resolvers, func(i, j int) bool {
		return c.resolvers[i].Order() < c.resolvers[j].Order()
	})
}

// GetResolvers 获取所有解析器
//
// 返回当前解析器链中的所有解析器，按优先级排序。
func (c *ResolverChain) GetResolvers() []ExceptionResolver {
	return c.resolvers
}

// Resolve 查找第一个支持的解析器并解析异常
//
// 遍历解析器链，找到第一个 Supports 返回 true 的解析器，
// 并调用其 Resolve 方法处理异常。如果没有找到支持的解析器，返回 nil。
func (c *ResolverChain) Resolve(ctx context.Context, err error) *ErrorResponse {
	for _, resolver := range c.resolvers {
		if resolver.Supports(err) {
			return resolver.Resolve(ctx, err)
		}
	}
	return nil
}
