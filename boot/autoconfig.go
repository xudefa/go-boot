// Package boot 提供应用启动器、自动配置注册、横幅打印和失败分析功能。
//
// 核心机制：
//   - AutoConfiguration: 各模块通过 RegisterAutoConfig 注册自动配置
//   - Starter: 启动器管理组件的启动和停止生命周期
//   - Banner: 启动横幅显示
//   - FailureAnalyzer: 启动失败时的友好错误提示
//
// 推荐入口：
//
//	ctx, err := boot.NewApplication(boot.WithAppName("my-app"))
//	if err != nil { log.Fatal(err) }
//	ctx.Start()
package boot

import (
	"sort"
	"sync"

	"github.com/xudefa/go-boot/condition"
	"github.com/xudefa/go-boot/core"
	"github.com/xudefa/go-boot/environment"
	"github.com/xudefa/go-boot/event"
)

// ApplicationContext 自动配置看到的上下文（结构型接口，避免包循环依赖）
//
// 实际由 context.DefaultApplicationContext 实现。
type ApplicationContext interface {
	Container() core.Container
	Environment() *environment.Environment
	Register(name string, opts ...core.BuilderOption) error
	Get(name string) (any, error)
	EventBus() interface {
		Publish(event event.ApplicationEvent)
	}
}

// AutoConfiguration 自动配置接口
//
// 参考 Spring Boot 的 @Configuration + @Bean 模式。
// 每个模块实现此接口，通过 RegisterAutoConfig 注册。
type AutoConfiguration interface {
	Configure(ctx ApplicationContext) error
}

// AutoConfigEntry 自动配置条目
type AutoConfigEntry struct {
	Config       AutoConfiguration
	Conditions   []condition.Condition
	Order        int
	Dependencies []string
}

// AutoConfigurationOption 自动配置选项
type AutoConfigurationOption func(entry *AutoConfigEntry)

// WithOrder 设置执行顺序，值越小优先级越高
func WithOrder(order int) AutoConfigurationOption {
	return func(entry *AutoConfigEntry) {
		entry.Order = order
	}
}

// WithDependsOn 设置依赖的配置名称
func WithDependsOn(deps ...string) AutoConfigurationOption {
	return func(entry *AutoConfigEntry) {
		entry.Dependencies = append(entry.Dependencies, deps...)
	}
}

// WithConditions 设置条件
func WithConditions(conds ...condition.Condition) AutoConfigurationOption {
	return func(entry *AutoConfigEntry) {
		entry.Conditions = append(entry.Conditions, conds...)
	}
}

// AutoConfigRegistry 自动配置注册表
type AutoConfigRegistry struct {
	mu      sync.RWMutex
	entries []AutoConfigEntry
}

var globalRegistry = NewAutoConfigRegistry()

// NewAutoConfigRegistry 创建注册表
func NewAutoConfigRegistry() *AutoConfigRegistry {
	return &AutoConfigRegistry{}
}

// RegisterAutoConfig 注册自动配置到全局注册表
//
// 在模块的 init() 中调用：
//
//	func init() {
//	    boot.RegisterAutoConfig(&CircuitAutoConfiguration{},
//	        condition.OnProperty("circuit.enabled", "true"),
//	    )
//	}
func RegisterAutoConfig(config AutoConfiguration, conditions ...condition.Condition) {
	globalRegistry.Add(AutoConfigEntry{
		Config:     config,
		Conditions: conditions,
		Order:      len(globalRegistry.GetAll()),
	})
}

// RegisterAutoConfigWith 注册自动配置到全局注册表（支持选项）
func RegisterAutoConfigWith(config AutoConfiguration, opts ...AutoConfigurationOption) {
	entry := AutoConfigEntry{
		Config: config,
		Order:  len(globalRegistry.GetAll()),
	}
	for _, opt := range opts {
		opt(&entry)
	}
	globalRegistry.Add(entry)
}

// Add 添加自动配置条目
func (r *AutoConfigRegistry) Add(entry AutoConfigEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, entry)
}

// GetAll 获取所有注册的自动配置
func (r *AutoConfigRegistry) GetAll() []AutoConfigEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]AutoConfigEntry, len(r.entries))
	copy(result, r.entries)
	return result
}

// GetMatching 获取匹配条件的自动配置（按 Order 排序）
func (r *AutoConfigRegistry) GetMatching(ctx condition.ConditionContext) []AutoConfigEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	matched := make([]AutoConfigEntry, 0, len(r.entries))
	for _, entry := range r.entries {
		if r.matchesAll(ctx, entry.Conditions) {
			matched = append(matched, entry)
		}
	}

	sort.Slice(matched, func(i, j int) bool {
		return matched[i].Order < matched[j].Order
	})
	return matched
}

func (r *AutoConfigRegistry) matchesAll(ctx condition.ConditionContext, conditions []condition.Condition) bool {
	if len(conditions) == 0 {
		return true
	}
	all := condition.All(conditions...)
	return all.Matches(ctx)
}

// GlobalRegistry 返回全局注册表
func GlobalRegistry() *AutoConfigRegistry {
	return globalRegistry
}
