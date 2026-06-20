package refresh

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/xudefa/go-boot/core"
)

// RefreshScopeManagerBuilder 刷新作用域管理器构建器，支持链式配置
type RefreshScopeManagerBuilder struct {
	beanCreator      core.BeanCreator
	logger           *slog.Logger
	opts             []RefreshOption
	refreshableBeans map[string]RefreshableBean
}

// NewRefreshScopeManagerBuilder 创建刷新作用域管理器构建器
func NewRefreshScopeManagerBuilder() *RefreshScopeManagerBuilder {
	return &RefreshScopeManagerBuilder{
		refreshableBeans: make(map[string]RefreshableBean),
	}
}

// BeanCreator 设置Bean创建器
func (b *RefreshScopeManagerBuilder) BeanCreator(creator core.BeanCreator) *RefreshScopeManagerBuilder {
	b.beanCreator = creator
	return b
}

// Logger 设置日志记录器
func (b *RefreshScopeManagerBuilder) Logger(logger *slog.Logger) *RefreshScopeManagerBuilder {
	b.logger = logger
	return b
}

// Enabled 设置是否启用刷新功能
func (b *RefreshScopeManagerBuilder) Enabled(enabled bool) *RefreshScopeManagerBuilder {
	b.opts = append(b.opts, WithRefreshEnabled(enabled))
	return b
}

// RefreshDelay 设置刷新延迟
func (b *RefreshScopeManagerBuilder) RefreshDelay(delay time.Duration) *RefreshScopeManagerBuilder {
	b.opts = append(b.opts, WithRefreshDelay(delay))
	return b
}

// MaxRefreshAttempts 设置最大刷新尝试次数
func (b *RefreshScopeManagerBuilder) MaxRefreshAttempts(attempts int) *RefreshScopeManagerBuilder {
	b.opts = append(b.opts, WithMaxRefreshAttempts(attempts))
	return b
}

// RefreshableBean 注册可刷新Bean
func (b *RefreshScopeManagerBuilder) RefreshableBean(beanID string, bean RefreshableBean) *RefreshScopeManagerBuilder {
	b.refreshableBeans[beanID] = bean
	return b
}

// Build 构建刷新作用域管理器
func (b *RefreshScopeManagerBuilder) Build() (*RefreshScopeManager, error) {
	if b.beanCreator == nil {
		return nil, fmt.Errorf("beanCreator is required")
	}

	manager := NewRefreshScopeManager(b.beanCreator, b.logger, b.opts...)

	// 注册可刷新Bean
	for beanID, bean := range b.refreshableBeans {
		manager.RegisterRefreshableBean(beanID, bean)
	}

	return manager, nil
}

// MustBuild 构建刷新作用域管理器，失败则panic
func (b *RefreshScopeManagerBuilder) MustBuild() *RefreshScopeManager {
	manager, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build refresh scope manager: %v", err))
	}
	return manager
}

// RefreshProxyBuilder 刷新代理构建器
type RefreshProxyBuilder struct {
	beanID  string
	target  any
	manager *RefreshScopeManager
	logger  *slog.Logger
}

// NewRefreshProxyBuilder 创建刷新代理构建器
func NewRefreshProxyBuilder() *RefreshProxyBuilder {
	return &RefreshProxyBuilder{}
}

// BeanID 设置Bean标识
func (b *RefreshProxyBuilder) BeanID(beanID string) *RefreshProxyBuilder {
	b.beanID = beanID
	return b
}

// Target 设置目标Bean实例
func (b *RefreshProxyBuilder) Target(target any) *RefreshProxyBuilder {
	b.target = target
	return b
}

// Manager 设置刷新作用域管理器
func (b *RefreshProxyBuilder) Manager(manager *RefreshScopeManager) *RefreshProxyBuilder {
	b.manager = manager
	return b
}

// Logger 设置日志记录器
func (b *RefreshProxyBuilder) Logger(logger *slog.Logger) *RefreshProxyBuilder {
	b.logger = logger
	return b
}

// Build 构建刷新代理
func (b *RefreshProxyBuilder) Build() (*RefreshProxy, error) {
	if b.beanID == "" {
		return nil, fmt.Errorf("beanID is required")
	}
	if b.target == nil {
		return nil, fmt.Errorf("target is required")
	}
	if b.manager == nil {
		return nil, fmt.Errorf("manager is required")
	}

	if b.logger == nil {
		b.logger = slog.Default()
	}

	return NewRefreshProxy(b.beanID, b.target, b.manager, b.logger), nil
}

// MustBuild 构建刷新代理，失败则panic
func (b *RefreshProxyBuilder) MustBuild() *RefreshProxy {
	proxy, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build refresh proxy: %v", err))
	}
	return proxy
}

// ConfigChangeEventBuilder 配置变更事件构建器
type ConfigChangeEventBuilder struct {
	event ConfigChangeEvent
}

// NewConfigChangeEventBuilder 创建配置变更事件构建器
func NewConfigChangeEventBuilder() *ConfigChangeEventBuilder {
	return &ConfigChangeEventBuilder{
		event: ConfigChangeEvent{
			OldValues: make(map[string]any),
			NewValues: make(map[string]any),
			Metadata:  make(map[string]string),
		},
	}
}

// EventType 设置事件类型
func (b *ConfigChangeEventBuilder) EventType(eventType string) *ConfigChangeEventBuilder {
	b.event.EventType = eventType
	return b
}

// Keys 设置变更的配置键列表
func (b *ConfigChangeEventBuilder) Keys(keys []string) *ConfigChangeEventBuilder {
	b.event.Keys = keys
	return b
}

// OldValues 设置旧值
func (b *ConfigChangeEventBuilder) OldValues(values map[string]any) *ConfigChangeEventBuilder {
	b.event.OldValues = values
	return b
}

// NewValues 设置新值
func (b *ConfigChangeEventBuilder) NewValues(values map[string]any) *ConfigChangeEventBuilder {
	b.event.NewValues = values
	return b
}

// Source 设置配置来源
func (b *ConfigChangeEventBuilder) Source(source string) *ConfigChangeEventBuilder {
	b.event.Source = source
	return b
}

// Metadata 设置元数据
func (b *ConfigChangeEventBuilder) Metadata(metadata map[string]string) *ConfigChangeEventBuilder {
	b.event.Metadata = metadata
	return b
}

// Build 构建配置变更事件
func (b *ConfigChangeEventBuilder) Build() ConfigChangeEvent {
	return b.event
}
