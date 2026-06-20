package aop

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/xudefa/go-boot/core"
)

// AopBeanPostProcessor AOP Bean后置处理器
//
// 在Bean创建后自动应用AOP代理
type AopBeanPostProcessor struct {
	integration *AopIntegration
	enabled     bool
	mu          sync.RWMutex
}

// NewAopBeanPostProcessor 创建AOP Bean后置处理器
func NewAopBeanPostProcessor(integration *AopIntegration) *AopBeanPostProcessor {
	if integration == nil {
		integration = GlobalAopIntegration
	}
	return &AopBeanPostProcessor{
		integration: integration,
		enabled:     true,
	}
}

// ProcessBean 处理Bean
func (p *AopBeanPostProcessor) ProcessBean(beanID string, bean any) any {
	p.mu.RLock()
	enabled := p.enabled
	integration := p.integration
	p.mu.RUnlock()

	if !enabled || bean == nil {
		return bean
	}

	// 检查是否需要AOP代理
	if p.needsProxy(bean) {
		proxy := integration.CreateProxy(beanID, bean)
		if proxy != nil && proxy != bean {
			return proxy
		}
	}

	return bean
}

// needsProxy 检查是否需要代理
func (p *AopBeanPostProcessor) needsProxy(bean any) bool {
	if bean == nil {
		return false
	}

	// 检查是否有匹配的切面
	beanType := reflect.TypeOf(bean)
	aspects := p.integration.GetManager().MatchAspectsForType(beanType)
	return len(aspects) > 0
}

// Enable 启用处理器
func (p *AopBeanPostProcessor) Enable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = true
}

// Disable 禁用处理器
func (p *AopBeanPostProcessor) Disable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = false
}

// IsEnabled 检查是否启用
func (p *AopBeanPostProcessor) IsEnabled() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.enabled
}

// GlobalAopBeanPostProcessor 全局AOP Bean后置处理器
var GlobalAopBeanPostProcessor = NewAopBeanPostProcessor(nil)

// AopBeanDefinition AOP Bean定义
//
// 扩展标准Bean定义，添加AOP相关配置
type AopBeanDefinition struct {
	*core.BeanDefinition
	BeanID     string
	EnableAop  bool
	ProxyMode  AopMode
	TargetType reflect.Type
	ProxyType  reflect.Type
	Aspects    []*AspectMeta
}

// NewAopBeanDefinition 创建AOP Bean定义
func NewAopBeanDefinition(beanID string, beanType reflect.Type) *AopBeanDefinition {
	return &AopBeanDefinition{
		BeanDefinition: &core.BeanDefinition{
			ConcreteType: beanType,
		},
		EnableAop:  true,
		ProxyMode:  AopModeMixed,
		TargetType: beanType,
	}
}

// WithAopEnabled 设置启用AOP
func (d *AopBeanDefinition) WithAopEnabled(enabled bool) *AopBeanDefinition {
	d.EnableAop = enabled
	return d
}

// WithProxyMode 设置代理模式
func (d *AopBeanDefinition) WithProxyMode(mode AopMode) *AopBeanDefinition {
	d.ProxyMode = mode
	return d
}

// WithAspects 设置切面
func (d *AopBeanDefinition) WithAspects(aspects ...*AspectMeta) *AopBeanDefinition {
	d.Aspects = append(d.Aspects, aspects...)
	return d
}

// WithProxyType 设置代理类型
func (d *AopBeanDefinition) WithProxyType(proxyType reflect.Type) *AopBeanDefinition {
	d.ProxyType = proxyType
	return d
}

// AopBeanFactory AOP Bean工厂
//
// 创建AOP代理Bean
type AopBeanFactory struct {
	integration *AopIntegration
	processor   *AopBeanPostProcessor
}

// NewAopBeanFactory 创建AOP Bean工厂
func NewAopBeanFactory(integration *AopIntegration) *AopBeanFactory {
	if integration == nil {
		integration = GlobalAopIntegration
	}
	return &AopBeanFactory{
		integration: integration,
		processor:   NewAopBeanPostProcessor(integration),
	}
}

// CreateBean 创建Bean
func (f *AopBeanFactory) CreateBean(beanID string, beanDef *AopBeanDefinition, target any) (any, error) {
	if beanDef == nil || !beanDef.EnableAop {
		return target, nil
	}

	// 注册切面
	if len(beanDef.Aspects) > 0 {
		f.integration.RegisterAspects(beanDef.Aspects...)
	}

	// 创建代理
	proxy := f.integration.CreateProxy(beanID, target)
	if proxy == nil {
		return target, nil
	}

	return proxy, nil
}

// RegisterBean 注册Bean到容器
func (f *AopBeanFactory) RegisterBean(container core.Container, beanID string, beanDef *AopBeanDefinition, target any) error {
	proxy, err := f.CreateBean(beanID, beanDef, target)
	if err != nil {
		return err
	}

	// 注册到容器
	return container.Register(beanID, core.Factory(func(c core.Container) (any, error) {
		return proxy, nil
	}, beanDef.TargetType), core.Singleton())
}

// GetProcessor 获取后置处理器
func (f *AopBeanFactory) GetProcessor() *AopBeanPostProcessor {
	return f.processor
}

// GlobalAopBeanFactory 全局AOP Bean工厂
var GlobalAopBeanFactory = NewAopBeanFactory(nil)

// AopContainer AOP容器
//
// 集成AOP功能的IoC容器
type AopContainer struct {
	core.Container
	integration *AopIntegration
	factory     *AopBeanFactory
	processor   *AopBeanPostProcessor
}

// NewAopContainer 创建AOP容器
func NewAopContainer(baseContainer core.Container) *AopContainer {
	if baseContainer == nil {
		baseContainer = core.New()
	}

	// 使用全局集成器，确保切面共享
	integration := GlobalAopIntegration
	factory := NewAopBeanFactory(integration)
	processor := factory.GetProcessor()

	return &AopContainer{
		Container:   baseContainer,
		integration: integration,
		factory:     factory,
		processor:   processor,
	}
}

// RegisterAopBean 注册AOP Bean
func (c *AopContainer) RegisterAopBean(beanDef *AopBeanDefinition, target any) error {
	return c.factory.RegisterBean(c.Container, beanDef.BeanID, beanDef, target)
}

// RegisterAopBeanWithID 注册AOP Bean（指定ID）
func (c *AopContainer) RegisterAopBeanWithID(beanID string, beanType reflect.Type, target any) error {
	beanDef := NewAopBeanDefinition(beanID, beanType)
	beanDef.BeanID = beanID
	return c.RegisterAopBean(beanDef, target)
}

// RegisterAopBeanWithAspects 注册AOP Bean（带切面）
func (c *AopContainer) RegisterAopBeanWithAspects(beanID string, beanType reflect.Type, target any, aspects ...*AspectMeta) error {
	beanDef := NewAopBeanDefinition(beanID, beanType)
	beanDef.WithAspects(aspects...)
	return c.RegisterAopBean(beanDef, target)
}

// GetAopProxy 获取AOP代理
func (c *AopContainer) GetAopProxy(beanID string) (any, error) {
	bean, err := c.Get(beanID)
	if err != nil {
		return nil, err
	}

	// 应用后置处理器
	return c.processor.ProcessBean(beanID, bean), nil
}

// RegisterAspect 注册切面
func (c *AopContainer) RegisterAspect(aspect *AspectMeta) {
	c.integration.RegisterAspect(aspect)
}

// RegisterAspects 批量注册切面
func (c *AopContainer) RegisterAspects(aspects ...*AspectMeta) {
	c.integration.RegisterAspects(aspects...)
}

// GetAspects 获取所有切面
func (c *AopContainer) GetAspects() []*AspectMeta {
	return c.integration.GetAspects()
}

// GetIntegration 获取AOP集成器
func (c *AopContainer) GetIntegration() *AopIntegration {
	return c.integration
}

// GetFactory 获取Bean工厂
func (c *AopContainer) GetFactory() *AopBeanFactory {
	return c.factory
}

// GetProcessor 获取后置处理器
func (c *AopContainer) GetProcessor() *AopBeanPostProcessor {
	return c.processor
}

// EnableAop 启用AOP
func (c *AopContainer) EnableAop() {
	c.processor.Enable()
}

// DisableAop 禁用AOP
func (c *AopContainer) DisableAop() {
	c.processor.Disable()
}

// IsAopEnabled 检查AOP是否启用
func (c *AopContainer) IsAopEnabled() bool {
	return c.processor.IsEnabled()
}

// AopContainerBuilder AOP容器构建器
//
// 提供流式API构建AOP容器
type AopContainerBuilder struct {
	baseContainer core.Container
	config        *AopConfig
	aspects       []*AspectMeta
	beans         []*AopBeanDefinition
}

// NewAopContainerBuilder 创建AOP容器构建器
func NewAopContainerBuilder() *AopContainerBuilder {
	return &AopContainerBuilder{
		config:  DefaultAopConfig(),
		aspects: make([]*AspectMeta, 0),
		beans:   make([]*AopBeanDefinition, 0),
	}
}

// WithBaseContainer 设置基础容器
func (b *AopContainerBuilder) WithBaseContainer(container core.Container) *AopContainerBuilder {
	b.baseContainer = container
	return b
}

// WithConfig 设置配置
func (b *AopContainerBuilder) WithConfig(config *AopConfig) *AopContainerBuilder {
	b.config = config
	return b
}

// WithAopMode 设置AOP模式
func (b *AopContainerBuilder) WithAopMode(mode AopMode) *AopContainerBuilder {
	b.config.Mode = mode
	return b
}

// WithAspect 添加切面
func (b *AopContainerBuilder) WithAspect(aspect *AspectMeta) *AopContainerBuilder {
	b.aspects = append(b.aspects, aspect)
	return b
}

// WithAspects 批量添加切面
func (b *AopContainerBuilder) WithAspects(aspects ...*AspectMeta) *AopContainerBuilder {
	b.aspects = append(b.aspects, aspects...)
	return b
}

// WithBean 添加Bean
func (b *AopContainerBuilder) WithBean(beanDef *AopBeanDefinition) *AopContainerBuilder {
	b.beans = append(b.beans, beanDef)
	return b
}

// WithBeanWithID 添加Bean（指定ID）
func (b *AopContainerBuilder) WithBeanWithID(beanID string, beanType reflect.Type, target any) *AopContainerBuilder {
	beanDef := NewAopBeanDefinition(beanID, beanType)
	b.beans = append(b.beans, beanDef)
	return b
}

// WithBeanWithAspects 添加Bean（带切面）
func (b *AopContainerBuilder) WithBeanWithAspects(beanID string, beanType reflect.Type, target any, aspects ...*AspectMeta) *AopContainerBuilder {
	beanDef := NewAopBeanDefinition(beanID, beanType)
	beanDef.WithAspects(aspects...)
	b.beans = append(b.beans, beanDef)
	return b
}

// Build 构建AOP容器
func (b *AopContainerBuilder) Build() (*AopContainer, error) {
	container := NewAopContainer(b.baseContainer)
	container.integration.manager.config = b.config

	// 注册切面
	container.RegisterAspects(b.aspects...)

	// 注册Bean
	for _, beanDef := range b.beans {
		// 从基础容器获取目标对象，如果基础容器不存在或对象不存在，则创建默认实例
		var target any
		if b.baseContainer != nil {
			if obj, err := b.baseContainer.Get(beanDef.BeanID); err == nil {
				target = obj
			}
		}

		// 如果仍未获取到目标对象，且定义中有具体类型，则创建零值
		if target == nil && beanDef.TargetType != nil {
			target = reflect.New(beanDef.TargetType).Interface()
		}

		// 注册AOP Bean
		if err := container.RegisterAopBean(beanDef, target); err != nil {
			return nil, fmt.Errorf("failed to register AOP bean %s: %w", beanDef.BeanID, err)
		}
	}

	return container, nil
}

// BuildOrPanic 构建AOP容器（失败时panic）
func (b *AopContainerBuilder) BuildOrPanic() *AopContainer {
	container, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build AOP container: %v", err))
	}
	return container
}

// CreateAopContainer 创建AOP容器的便捷函数
func CreateAopContainer() *AopContainer {
	return NewAopContainer(nil)
}

// CreateAopContainerWithConfig 创建AOP容器（指定配置）
func CreateAopContainerWithConfig(config *AopConfig) *AopContainer {
	container := NewAopContainer(nil)
	container.integration.manager.config = config
	return container
}

// RegisterAopBeanToGlobal 注册AOP Bean到全局容器
func RegisterAopBeanToGlobal(beanID string, beanType reflect.Type, target any) error {
	container := CreateAopContainer()
	return container.RegisterAopBeanWithID(beanID, beanType, target)
}

// RegisterAspectToGlobalContainer 注册切面到全局容器
func RegisterAspectToGlobalContainer(aspect *AspectMeta) {
	GlobalAopIntegration.RegisterAspect(aspect)
}

// GetAopBeanFromGlobal 从全局容器获取AOP Bean
func GetAopBeanFromGlobal(beanID string) (any, error) {
	container := CreateAopContainer()
	return container.GetAopProxy(beanID)
}

// AopBeanScanner AOP Bean扫描器
//
// 扫描并注册带有AOP注解的Bean
type AopBeanScanner struct {
	container *AopContainer
	basePath  string
	enabled   bool
	mu        sync.RWMutex
}

// NewAopBeanScanner 创建AOP Bean扫描器
func NewAopBeanScanner(container *AopContainer) *AopBeanScanner {
	if container == nil {
		container = CreateAopContainer()
	}
	return &AopBeanScanner{
		container: container,
		enabled:   true,
	}
}

// Scan 扫描指定路径
func (s *AopBeanScanner) Scan(basePath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.basePath = basePath

	// 这里需要实现实际的扫描逻辑
	// 扫描带有 @AopProxy 注解的类型并注册
	return nil
}

// Enable 启用扫描器
func (s *AopBeanScanner) Enable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = true
}

// Disable 禁用扫描器
func (s *AopBeanScanner) Disable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = false
}

// IsEnabled 检查是否启用
func (s *AopBeanScanner) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// GetContainer 获取容器
func (s *AopBeanScanner) GetContainer() *AopContainer {
	return s.container
}

// GlobalAopBeanScanner 全局AOP Bean扫描器
var GlobalAopBeanScanner = NewAopBeanScanner(nil)

// ScanAopBeans 扫描AOP Bean
func ScanAopBeans(basePath string) error {
	return GlobalAopBeanScanner.Scan(basePath)
}

// AutoScan 自动扫描
func AutoScan() error {
	return GlobalAopBeanScanner.Scan(".")
}

// AopMetrics AOP指标
//
// 收集AOP相关的性能指标
type AopMetrics struct {
	TotalProxies       int64
	GeneratedProxies   int64
	RuntimeProxies     int64
	TotalAspects       int64
	TotalInterceptions int64
	AverageLatency     float64
	mu                 sync.RWMutex
}

// NewAopMetrics 创建AOP指标
func NewAopMetrics() *AopMetrics {
	return &AopMetrics{}
}

// RecordProxyCreated 记录代理创建
func (m *AopMetrics) RecordProxyCreated(isGenerated bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalProxies++
	if isGenerated {
		m.GeneratedProxies++
	} else {
		m.RuntimeProxies++
	}
}

// RecordAspectRegistered 记录切面注册
func (m *AopMetrics) RecordAspectRegistered() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalAspects++
}

// RecordInterception 记录拦截
func (m *AopMetrics) RecordInterception(latency float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalInterceptions++

	// 计算平均延迟
	if m.TotalInterceptions == 1 {
		m.AverageLatency = latency
	} else {
		m.AverageLatency = (m.AverageLatency*float64(m.TotalInterceptions-1) + latency) / float64(m.TotalInterceptions)
	}
}

// GetMetrics 获取指标
func (m *AopMetrics) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"total_proxies":       m.TotalProxies,
		"generated_proxies":   m.GeneratedProxies,
		"runtime_proxies":     m.RuntimeProxies,
		"total_aspects":       m.TotalAspects,
		"total_interceptions": m.TotalInterceptions,
		"average_latency":     m.AverageLatency,
	}
}

// Reset 重置指标
func (m *AopMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalProxies = 0
	m.GeneratedProxies = 0
	m.RuntimeProxies = 0
	m.TotalAspects = 0
	m.TotalInterceptions = 0
	m.AverageLatency = 0
}

// GlobalAopMetrics 全局AOP指标
var GlobalAopMetrics = NewAopMetrics()

// GetGlobalAopMetrics 获取全局AOP指标
func GetGlobalAopMetrics() map[string]interface{} {
	return GlobalAopMetrics.GetMetrics()
}

// ResetGlobalAopMetrics 重置全局AOP指标
func ResetGlobalAopMetrics() {
	GlobalAopMetrics.Reset()
}
