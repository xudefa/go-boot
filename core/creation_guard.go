package core

import (
	"fmt"
	"sync"
)

// BeanCreationGuard 管理 Bean 创建过程中的并发安全
//
// 使用细粒度锁和等待队列保证并发安全：
// - 第一个请求者触发创建
// - 创建期间其他请求检测到循环依赖
// - 创建完成后缓存结果
//
// 设计模式: Guarded Suspension
type BeanCreationGuard struct {
	container Container
	mu        sync.Mutex
	creating  map[string]bool           // 正在创建的 Bean 标记
	results   map[string]creationResult // 创建结果缓存
}

// creationResult Bean 创建结果
type creationResult struct {
	instance any
	err      error
}

// NewBeanCreationGuard 创建 Bean 创建守卫
func NewBeanCreationGuard(container Container) *BeanCreationGuard {
	return &BeanCreationGuard{
		container: container,
		creating:  make(map[string]bool),
		results:   make(map[string]creationResult),
	}
}

// GetOrCompute 获取或计算 Bean 实例
//
// 如果 Bean 已创建完成，直接返回缓存结果。
// 如果 Bean 正在创建中，返回 ErrCircularDep（循环依赖检测）。
// 否则执行 factory 创建 Bean 并缓存结果。
func (g *BeanCreationGuard) GetOrCompute(name string, factory func() (any, error)) (any, error) {
	g.mu.Lock()

	// 检查是否已有结果缓存
	if res, ok := g.results[name]; ok {
		g.mu.Unlock()
		return res.instance, res.err
	}

	// 检查是否正在创建中（循环依赖检测）
	if g.creating[name] {
		g.mu.Unlock()
		return nil, fmt.Errorf("%w: %s", ErrCircularDep, name)
	}

	// 标记为正在创建
	g.creating[name] = true
	g.mu.Unlock()

	// 执行创建
	instance, err := factory()

	// 缓存结果
	g.mu.Lock()
	g.results[name] = creationResult{instance: instance, err: err}
	delete(g.creating, name)
	g.mu.Unlock()

	return instance, err
}

// IsCreating 检查 Bean 是否正在创建中
func (g *BeanCreationGuard) IsCreating(name string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.creating[name]
}

// Clear 清理指定 Bean 的创建状态
func (g *BeanCreationGuard) Clear(name string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.creating, name)
	delete(g.results, name)
}
