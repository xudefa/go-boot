// Package context 提供应用上下文，聚合 IoC 容器、环境配置、生命周期和事件系统。
//
// ApplicationContext 是 go-boot 框架的核心运行时入口，封装了：
//   - Container: 依赖注入容器
//   - Environment: 分层配置源管理
//   - Lifecycle: 应用生命周期阶段管理
//   - EventBus: 事件发布与订阅
//
// 使用示例：
//
//	ctx := context.NewApplicationContext(container, env)
//	ctx.Start()
package context

import (
	"errors"
	"log/slog"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/xudefa/go-boot/core"
	"github.com/xudefa/go-boot/environment"
	"github.com/xudefa/go-boot/event"
	"github.com/xudefa/go-boot/life"
	"github.com/xudefa/go-boot/refresh"
)

// EventPublisher 事件发布器接口
//
// 用于解耦事件发布逻辑，便于测试和替换实现。
type EventPublisher interface {
	Publish(event event.ApplicationEvent)
}

// ApplicationContext 应用上下文接口
//
// 统一封装了 IoC 容器、Environment、生命周期和事件系统。
// 是使用 go-boot 框架的核心入口，提供了应用的完整运行环境。
//
// 功能说明:
//   - Container: 获取依赖注入容器，管理 Bean 的注册和获取
//   - Environment: 获取环境配置，支持多级配置源
//   - Lifecycle: 获取生命周期管理器，控制应用状态流转
//   - EventBus: 获取事件总线，支持事件的发布与订阅
//   - EventPublisher: 获取事件发布器接口，用于解耦事件发布逻辑
//   - Start/Stop: 控制应用的启动和停止
type ApplicationContext interface {
	// Container 返回 IoC 容器实例
	Container() core.Container
	// Environment 返回环境配置实例
	Environment() *environment.Environment
	// Lifecycle 返回生命周期管理器
	Lifecycle() *life.LifecycleManager
	// EventBus 返回事件总线（完整访问，向后兼容）
	EventBus() *event.EventBus
	// EventPublisher 返回事件发布器接口（解耦事件发布）
	EventPublisher() EventPublisher
	// RefreshScopeManager 返回刷新作用域管理器
	RefreshScopeManager() *refresh.RefreshScopeManager

	// Register 在容器中注册 Bean
	Register(name string, opts ...core.BuilderOption) error
	// Get 从容器中获取指定名称的 Bean
	Get(name string) (any, error)
	// Invoke 调用函数并自动注入依赖参数
	Invoke(fn any, opts ...core.InvokeOption) error

	// Start 启动应用，发布启动事件并切换至运行阶段
	Start() error
	// Stop 停止应用，切换至停止阶段并发布停止事件
	Stop() error
	// IsRunning 检查应用是否处于运行状态
	IsRunning() bool
}

// DefaultApplicationContext 默认应用上下文实现
//
// 组合了 IoC 容器、环境配置、生命周期管理和事件总线，
// 提供 go-boot 框架的核心运行时能力。
type DefaultApplicationContext struct {
	container       core.Container               // IoC 依赖注入容器
	env             *environment.Environment     // 环境配置管理
	lifecycle       *life.LifecycleManager       // 生命周期管理器
	events          *event.EventBus              // 事件总线
	refreshScopeMgr *refresh.RefreshScopeManager // 刷新作用域管理器
}

// NewApplicationContext 创建默认应用上下文实例
func NewApplicationContext(container core.Container, env *environment.Environment, opts ...refresh.RefreshOption) *DefaultApplicationContext {
	refreshMgr := refresh.NewRefreshScopeManager(container, slog.Default(), opts...)
	return &DefaultApplicationContext{
		container:       container,
		env:             env,
		lifecycle:       life.NewLifecycleManager(),
		events:          event.NewEventBus(),
		refreshScopeMgr: refreshMgr,
	}
}

func (c *DefaultApplicationContext) Container() core.Container {
	return c.container
}

func (c *DefaultApplicationContext) Environment() *environment.Environment {
	return c.env
}

func (c *DefaultApplicationContext) Lifecycle() *life.LifecycleManager {
	return c.lifecycle
}

func (c *DefaultApplicationContext) EventBus() *event.EventBus {
	return c.events
}

func (c *DefaultApplicationContext) EventPublisher() EventPublisher {
	return c.events
}

func (c *DefaultApplicationContext) RefreshScopeManager() *refresh.RefreshScopeManager {
	return c.refreshScopeMgr
}

func (c *DefaultApplicationContext) Register(name string, opts ...core.BuilderOption) error {
	return c.container.Register(name, opts...)
}

func (c *DefaultApplicationContext) Get(name string) (any, error) {
	return c.container.Get(name)
}

func (c *DefaultApplicationContext) Invoke(fn any, opts ...core.InvokeOption) error {
	_, err := c.container.Invoke(fn, opts...)
	return err
}

// Start 启动应用：PhaseReady → PhaseRunning
func (c *DefaultApplicationContext) Start() error {
	c.events.Publish(&event.BaseEvent{EventType: event.EventApplicationStarted})

	if err := c.lifecycle.SetPhase(life.PhaseRunning); err != nil {
		return err
	}

	c.events.Publish(&event.BaseEvent{EventType: event.EventApplicationReady})
	return nil
}

// Stop 停止应用：PhaseRunning → PhaseStopping → PhaseStopped
func (c *DefaultApplicationContext) Stop() error {
	if err := c.lifecycle.SetPhase(life.PhaseStopping); err != nil {
		return err
	}

	if err := c.lifecycle.SetPhase(life.PhaseStopped); err != nil {
		return err
	}

	c.events.Publish(&event.BaseEvent{EventType: event.EventApplicationStopped})
	return nil
}

// IsRunning 检查应用是否运行中
func (c *DefaultApplicationContext) IsRunning() bool {
	return c.lifecycle.GetPhase() == life.PhaseRunning
}

func (c *DefaultApplicationContext) GetBean(beanID string) (any, bool) {
	val, err := c.container.Get(beanID)
	return val, err == nil
}

func (c *DefaultApplicationContext) HasProperty(key string) bool {
	_, ok := c.env.GetProperty(key)
	return ok
}

func (c *DefaultApplicationContext) GetProperty(key string) (any, bool) {
	return c.env.GetProperty(key)
}

func (c *DefaultApplicationContext) ClassLoader() interface{ HasClass(name string) bool } {
	return globalClassLoader
}

// globalClassLoader 全局共享 ClassLoader 实例，缓存构建信息以提升性能
var globalClassLoader = &buildInfoClassLoader{}

// buildInfoClassLoader 使用 runtime/debug.ReadBuildInfo() 检查模块是否在编译依赖中。
// 通过提取类名中的模块路径前缀，与构建信息中列出的模块路径进行匹配。
// 如果 ReadBuildInfo 不可用（如 go run 临时二进制），回退为 false。
//
// 优化：使用 sync.Once 延迟初始化并缓存依赖列表，避免每次调用都读取构建信息。
type buildInfoClassLoader struct {
	once     sync.Once
	deps     []string // 缓存的依赖路径列表
	mainPath string
	err      error
}

func (b *buildInfoClassLoader) HasClass(name string) bool {
	b.once.Do(b.init)
	if b.err != nil {
		return false
	}

	pkgPath := extractPkgPath(name)
	if pkgPath == "" {
		return false
	}

	// 先检查主模块
	if b.mainPath != "" && pathMatches(b.mainPath, pkgPath) {
		return true
	}

	// 再检查缓存的依赖列表
	for _, dep := range b.deps {
		if pathMatches(dep, pkgPath) {
			return true
		}
	}
	return false
}

func (b *buildInfoClassLoader) init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		slog.Warn("build info not available; OnClass/OnMissingClass conditions will not work")
		b.err = errors.New("build info not available")
		return
	}

	b.mainPath = info.Main.Path
	b.deps = make([]string, 0, len(info.Deps))
	for _, dep := range info.Deps {
		if dep != nil {
			b.deps = append(b.deps, dep.Path)
		}
	}
}

// extractPkgPath 从完整类型名中提取包路径。
// "gin.Engine" → "gin"
// "github.com/gin-gonic/gin.Engine" → "github.com/gin-gonic/gin"
func extractPkgPath(name string) string {
	if idx := strings.LastIndex(name, "."); idx > 0 {
		return name[:idx]
	}
	return name
}

// pathMatches 检查模块路径是否匹配包路径。
//
// 匹配策略：
//  1. 完全匹配（modulePath == pkgPath）
//  2. pkgPath 以 modulePath/ 开头（子包匹配）
//  3. 当 pkgPath 包含 "." 时（如 "gin.Engine" → pkgPath="gin"），
//     modulePath 至少包含一个 "/" 且以 "/"+pkgPath 结尾，
//     避免单段包名导致的误匹配（如 "gin" 误匹 "xxx/gin"）
func pathMatches(modulePath, pkgPath string) bool {
	if modulePath == pkgPath {
		return true
	}
	if strings.HasPrefix(pkgPath, modulePath+"/") {
		return true
	}
	if !strings.Contains(pkgPath, ".") && strings.Count(modulePath, "/") > 0 && strings.HasSuffix(modulePath, "/"+pkgPath) {
		return true
	}
	return false
}
