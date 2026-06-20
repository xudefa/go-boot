package aop

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// GeneratedProxyRegistry 代码生成代理注册表
//
// 管理代码生成的代理对象，提供查找和获取功能
type GeneratedProxyRegistry struct {
	proxies map[string]reflect.Type
	mu      sync.RWMutex
}

// NewGeneratedProxyRegistry 创建代码生成代理注册表
func NewGeneratedProxyRegistry() *GeneratedProxyRegistry {
	return &GeneratedProxyRegistry{
		proxies: make(map[string]reflect.Type),
	}
}

// Register 注册代理类型
func (r *GeneratedProxyRegistry) Register(beanID string, proxyType reflect.Type) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.proxies[beanID] = proxyType
}

// Get 获取代理类型
func (r *GeneratedProxyRegistry) Get(beanID string) (reflect.Type, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.proxies[beanID]
	return t, ok
}

// Has 检查是否存在代理
func (r *GeneratedProxyRegistry) Has(beanID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.proxies[beanID]
	return ok
}

// List 列出所有注册的bean ID
func (r *GeneratedProxyRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.proxies))
	for id := range r.proxies {
		ids = append(ids, id)
	}
	return ids
}

// Clear 清空注册表
func (r *GeneratedProxyRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.proxies = make(map[string]reflect.Type)
}

// GlobalGeneratedProxyRegistry 全局代码生成代理注册表
var GlobalGeneratedProxyRegistry = NewGeneratedProxyRegistry()

// RegisterGeneratedProxy 注册代码生成的代理
func RegisterGeneratedProxy(beanID string, proxyType reflect.Type) {
	GlobalGeneratedProxyRegistry.Register(beanID, proxyType)
}

// GetGeneratedProxy 获取代码生成的代理类型
func GetGeneratedProxy(beanID string) (reflect.Type, bool) {
	return GlobalGeneratedProxyRegistry.Get(beanID)
}

// HasGeneratedProxy 检查是否存在代码生成的代理
func HasGeneratedProxy(beanID string) bool {
	return GlobalGeneratedProxyRegistry.Has(beanID)
}

// GeneratedProxyFactory 代码生成代理工厂
//
// 创建代码生成的代理对象
type GeneratedProxyFactory struct {
	registry *GeneratedProxyRegistry
}

// NewGeneratedProxyFactory 创建代码生成代理工厂
func NewGeneratedProxyFactory() *GeneratedProxyFactory {
	return &GeneratedProxyFactory{
		registry: GlobalGeneratedProxyRegistry,
	}
}

// Create 创建代理对象
func (f *GeneratedProxyFactory) Create(beanID string, target any) (any, error) {
	proxyType, ok := f.registry.Get(beanID)
	if !ok {
		return nil, fmt.Errorf("no generated proxy found for bean: %s", beanID)
	}

	proxyValue := reflect.New(proxyType.Elem())
	proxy := proxyValue.Interface()

	// 设置目标对象
	targetField := proxyValue.Elem().FieldByName("target")
	if targetField.IsValid() && targetField.CanSet() {
		targetField.Set(reflect.ValueOf(target))
	}

	return proxy, nil
}

// CreateOrFallback 创建代理或回退到运行时代理
func (f *GeneratedProxyFactory) CreateOrFallback(beanID string, target any, fallback Weaver) any {
	proxy, err := f.Create(beanID, target)
	if err != nil {
		// 回退到运行时代理
		if fallback != nil {
			return fallback.Weave(target)
		}
		return target
	}
	return proxy
}

// AspectMetadataExtractor 切面元数据提取器
//
// 从代码生成的代理中提取切面元数据
type AspectMetadataExtractor struct{}

// NewAspectMetadataExtractor 创建切面元数据提取器
func NewAspectMetadataExtractor() *AspectMetadataExtractor {
	return &AspectMetadataExtractor{}
}

// Extract 从代理类型提取切面元数据
func (e *AspectMetadataExtractor) Extract(proxyType reflect.Type) []*AspectMeta {
	if proxyType.Kind() != reflect.Struct {
		return nil
	}

	_, found := proxyType.FieldByName("aspects")
	if !found {
		return nil
	}

	// 这里需要解析代码生成的切面元数据
	// 由于代码生成的结构是固定的，我们可以直接提取
	return nil
}

// ExtractFromBeanID 从bean ID提取切面元数据
func (e *AspectMetadataExtractor) ExtractFromBeanID(beanID string) []*AspectMeta {
	proxyType, ok := GlobalGeneratedProxyRegistry.Get(beanID)
	if !ok {
		return nil
	}
	return e.Extract(proxyType)
}

// AopIntegration AOP集成器
//
// 提供代码生成和运行时AOP的统一集成
type AopIntegration struct {
	config            *AopConfig
	manager           *AopManager
	proxyFactory      *GeneratedProxyFactory
	metadataExtractor *AspectMetadataExtractor
}

// NewAopIntegration 创建AOP集成器
func NewAopIntegration(config *AopConfig) *AopIntegration {
	if config == nil {
		config = DefaultAopConfig()
	}

	return &AopIntegration{
		config:            config,
		manager:           GlobalAopManager,
		proxyFactory:      NewGeneratedProxyFactory(),
		metadataExtractor: NewAspectMetadataExtractor(),
	}
}

// GetManager 获取AOP管理器
func (i *AopIntegration) GetManager() *AopManager {
	return i.manager
}

// GetProxyFactory 获取代理工厂
func (i *AopIntegration) GetProxyFactory() *GeneratedProxyFactory {
	return i.proxyFactory
}

// GetMetadataExtractor 获取元数据提取器
func (i *AopIntegration) GetMetadataExtractor() *AspectMetadataExtractor {
	return i.metadataExtractor
}

// CreateProxy 创建代理对象
func (i *AopIntegration) CreateProxy(beanID string, target any) any {
	switch i.manager.config.Mode {
	case AopModeGenerated:
		return i.proxyFactory.CreateOrFallback(beanID, target, i.manager.config.Weaver)
	case AopModeRuntime:
		return i.manager.config.Weaver.Weave(target)
	case AopModeMixed:
		// 优先使用代码生成的代理
		if HasGeneratedProxy(beanID) {
			proxy, err := i.proxyFactory.Create(beanID, target)
			if err == nil {
				return proxy
			}
		}
		// 回退到运行时代理
		return i.manager.config.Weaver.Weave(target)
	default:
		return target
	}
}

// RegisterAspect 注册切面
func (i *AopIntegration) RegisterAspect(aspect *AspectMeta) {
	i.manager.RegisterAspect(aspect)
}

// RegisterAspects 批量注册切面
func (i *AopIntegration) RegisterAspects(aspects ...*AspectMeta) {
	i.manager.RegisterAspects(aspects...)
}

// GetAspects 获取所有切面
func (i *AopIntegration) GetAspects() []*AspectMeta {
	return i.manager.GetAspects()
}

// GlobalAopIntegration 全局AOP集成器
var GlobalAopIntegration = NewAopIntegration(nil)

// CreateProxy 创建代理对象（使用全局集成器）
func CreateProxy(beanID string, target any) any {
	return GlobalAopIntegration.CreateProxy(beanID, target)
}

// RegisterAspectToGlobal 注册切面到全局集成器
func RegisterAspectToGlobal(aspect *AspectMeta) {
	GlobalAopIntegration.RegisterAspect(aspect)
}

// GetGlobalAspects 获取全局切面
func GetGlobalAspects() []*AspectMeta {
	return GlobalAopIntegration.GetAspects()
}

// AutoRegister 自动注册切面
//
// 从代码生成的代理中自动提取并注册切面
func AutoRegister(beanID string) error {
	aspects := GlobalAopIntegration.GetMetadataExtractor().ExtractFromBeanID(beanID)
	if len(aspects) == 0 {
		return fmt.Errorf("no aspects found for bean: %s", beanID)
	}

	GlobalAopIntegration.RegisterAspects(aspects...)
	return nil
}

// AutoRegisterAll 自动注册所有切面
//
// 从所有代码生成的代理中自动提取并注册切面
func AutoRegisterAll() error {
	beanIDs := GlobalGeneratedProxyRegistry.List()
	for _, beanID := range beanIDs {
		if err := AutoRegister(beanID); err != nil {
			// 继续处理其他bean，不中断
			continue
		}
	}
	return nil
}

// BuildTagChecker 构建标签检查器
//
// 检查当前构建是否包含特定标签
type BuildTagChecker struct{}

// NewBuildTagChecker 创建构建标签检查器
func NewBuildTagChecker() *BuildTagChecker {
	return &BuildTagChecker{}
}

// HasTag 检查是否有指定标签
func (c *BuildTagChecker) HasTag(tag string) bool {
	// 这里需要实际的构建标签检查逻辑
	// 目前先返回false，实际使用时需要实现
	return false
}

// IsGeneratedMode 检查是否为代码生成模式
func (c *BuildTagChecker) IsGeneratedMode() bool {
	return c.HasTag("goaop")
}

// IsRuntimeMode 检查是否为运行时模式
func (c *BuildTagChecker) IsRuntimeMode() bool {
	return !c.IsGeneratedMode()
}

// GetOptimalMode 获取最优模式
func (c *BuildTagChecker) GetOptimalMode() AopMode {
	if c.IsGeneratedMode() {
		return AopModeGenerated
	}
	return AopModeRuntime
}

// GlobalBuildTagChecker 全局构建标签检查器
var GlobalBuildTagChecker = NewBuildTagChecker()

// DetectOptimalMode 检测最优AOP模式
func DetectOptimalMode() AopMode {
	return GlobalBuildTagChecker.GetOptimalMode()
}

// ConfigureAopManager 配置AOP管理器
//
// 根据构建标签自动配置最优的AOP模式
func ConfigureAopManager() *AopConfig {
	config := DefaultAopConfig()
	config.Mode = DetectOptimalMode()
	return config
}

// InitializeAop 初始化AOP
//
// 自动配置并初始化AOP系统
func InitializeAop() {
	config := ConfigureAopManager()
	GlobalAopIntegration = NewAopIntegration(config)

	// 如果是代码生成模式，自动注册切面
	if config.Mode == AopModeGenerated || config.Mode == AopModeMixed {
		if err := AutoRegisterAll(); err != nil {
			fmt.Printf("[go-boot] failed to auto register aspects: %v\n", err)
		}
	}
}

// GetProxyWithAutoMode 使用自动模式获取代理
func GetProxyWithAutoMode(beanID string, target any) any {
	return GlobalAopIntegration.CreateProxy(beanID, target)
}

// AspectBuilder 切面构建器
//
// 提供流式API构建切面
type AspectBuilder struct {
	pointCut PointCut
	advice   Advice
	order    int
	instance any
}

// NewAspectBuilder 创建切面构建器
func NewAspectBuilder() *AspectBuilder {
	return &AspectBuilder{
		order: 0,
	}
}

// PointCut 设置切点
func (b *AspectBuilder) PointCut(pointCut PointCut) *AspectBuilder {
	b.pointCut = pointCut
	return b
}

// Advice 设置通知
func (b *AspectBuilder) Advice(advice Advice) *AspectBuilder {
	b.advice = advice
	return b
}

// Order 设置执行顺序
func (b *AspectBuilder) Order(order int) *AspectBuilder {
	b.order = order
	return b
}

// Instance 设置切面实例
func (b *AspectBuilder) Instance(instance any) *AspectBuilder {
	b.instance = instance
	return b
}

// Before 设置前置通知
func (b *AspectBuilder) Before(fn func(JoinPoint)) *AspectBuilder {
	b.advice = Before(fn)
	return b
}

// After 设置后置通知
func (b *AspectBuilder) After(fn func(JoinPoint)) *AspectBuilder {
	b.advice = After(fn)
	return b
}

// Around 设置环绕通知
func (b *AspectBuilder) Around(fn func(JoinPoint, ProceedFunc) any) *AspectBuilder {
	b.advice = Around(fn)
	return b
}

// MatchByName 设置按名称匹配的切点
func (b *AspectBuilder) MatchByName(name string) *AspectBuilder {
	b.pointCut = MatchByName(name)
	return b
}

// MatchByRegex 设置按正则匹配的切点
func (b *AspectBuilder) MatchByRegex(pattern string) *AspectBuilder {
	b.pointCut = MatchByRegex(pattern)
	return b
}

// MatchInterface 设置按接口匹配的切点
func (b *AspectBuilder) MatchInterface(iface any) *AspectBuilder {
	b.pointCut = MatchInterface(iface)
	return b
}

// MatchAll 设置匹配所有
func (b *AspectBuilder) MatchAll() *AspectBuilder {
	b.pointCut = MatchAll()
	return b
}

// Build 构建切面元数据
func (b *AspectBuilder) Build() *AspectMeta {
	return &AspectMeta{
		PointCut: b.pointCut,
		Advice:   b.advice,
		Order:    b.order,
		Instance: b.instance,
	}
}

// BuildAndRegister 构建并注册切面
func (b *AspectBuilder) BuildAndRegister() *AspectMeta {
	aspect := b.Build()
	RegisterAspectToGlobal(aspect)
	return aspect
}

// CreateAspect 创建切面的便捷函数
func CreateAspect(pointCut PointCut, advice Advice, order int) *AspectMeta {
	return &AspectMeta{
		PointCut: pointCut,
		Advice:   advice,
		Order:    order,
	}
}

// CreateBeforeAspect 创建前置切面的便捷函数
func CreateBeforeAspect(methodName string, fn func(JoinPoint), order int) *AspectMeta {
	return CreateAspect(MatchByName(methodName), Before(fn), order)
}

// CreateAfterAspect 创建后置切面的便捷函数
func CreateAfterAspect(methodName string, fn func(JoinPoint), order int) *AspectMeta {
	return CreateAspect(MatchByName(methodName), After(fn), order)
}

// CreateAroundAspect 创建环绕切面的便捷函数
func CreateAroundAspect(methodName string, fn func(JoinPoint, ProceedFunc) any, order int) *AspectMeta {
	return CreateAspect(MatchByName(methodName), Around(fn), order)
}

// ParseAspectTarget 解析切面目标
//
// 解析类似 "UserService.GetUser" 的目标字符串
func ParseAspectTarget(target string) (structName, methodName string, err error) {
	parts := strings.Split(target, ".")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid target format: %s, expected Struct.Method", target)
	}
	return parts[0], parts[1], nil
}

// CreateAspectFromTarget 从目标字符串创建切面
func CreateAspectFromTarget(target string, advice Advice, order int) (*AspectMeta, error) {
	structName, methodName, err := ParseAspectTarget(target)
	if err != nil {
		return nil, err
	}

	pointCut := MatchByName(methodName)
	return &AspectMeta{
		PointCut: pointCut,
		Advice:   advice,
		Order:    order,
		Instance: structName,
	}, nil
}
