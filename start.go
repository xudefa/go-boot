// Package src 提供应用启动器和配置接口。
//
// 该包定义通用的启动器接口，用于启动各种服务和中间件。
//
// 定义:
//
//   - Starter: 应用启动器接口
//   - StarterFunc: 启动函数类型
//   - Runners: 启动Runner列表
//
// 快速开始:
//
//	server := func() error {
//	    // 启动服务器
//	    return nil
//	}
//	server.Start()
package src

import (
	"context"
)

// Starter 是应用启动器接口。
//
// 所有需要启动的服务、中间件、插件都应实现此接口。
// 通常在 main 函数中通过容器统一管理和启动。
//
// 方法说明:
//
//   - Starter() error: 启动组件并返回错误，实现应支持优雅关闭
type Starter interface {
	Starter() error
}

// StarterFunc 是启动函数类型。
type StarterFunc func() error

// Starter 实现 Starter 接口。
func (f StarterFunc) Starter() error {
	return f()
}

// Runners 是启动Runner列表，用于按顺序启动多个组件。
type Runners []Starter

// Add 添加启动Runner。
func (r *Runners) Add(starters ...Starter) {
	*r = append(*r, starters...)
}

// StartAll 按顺序启动所有Runner。
func (r *Runners) StartAll() error {
	for _, starter := range *r {
		if err := starter.Starter(); err != nil {
			return err
		}
	}
	return nil
}

// ContextStarter 是带上下文的启动器接口。
type ContextStarter interface {
	Start(context.Context) error
}

// ContextStarterFunc 是带上下文的启动函数类型。
type ContextStarterFunc func(context.Context) error

// Start 实现 ContextStarter 接口。
func (f ContextStarterFunc) Start(ctx context.Context) error {
	return f(ctx)
}

// ContextRunners 是带上下文的启动Runner列表。
type ContextRunners []ContextStarter

// Add 添加启动Runner。
func (r *ContextRunners) Add(starters ...ContextStarter) {
	*r = append(*r, starters...)
}

// StartAll 按顺序启动所有Runner。
func (r *ContextRunners) StartAll(ctx context.Context) error {
	for _, starter := range *r {
		if err := starter.Start(ctx); err != nil {
			return err
		}
	}
	return nil
}
