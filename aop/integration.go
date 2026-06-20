package aop

import (
	"fmt"
	"reflect"
	"sync"
)

// AopMode AOP模式
type AopMode string

const (
	// AopModeRuntime 运行时模式
	// 使用反射和动态代理，性能较低但灵活性高
	AopModeRuntime AopMode = "runtime"

	// AopModeGenerated 代码生成模式
	// 使用代码生成的静态代理，性能高但需要编译时代码生成
	AopModeGenerated AopMode = "generated"

	// AopModeMixed 混合模式
	// 根据情况自动选择运行时或代码生成模式
	AopModeMixed AopMode = "mixed"
)

// AopConfig AOP配置
type AopConfig struct {
	Mode        AopMode
	Registry    *AopRegistry
	Weaver      Weaver
	EnableCache bool
}

// DefaultAopConfig 默认AOP配置
func DefaultAopConfig() *AopConfig {
	return &AopConfig{
		Mode:        AopModeMixed,
		Registry:    NewAopRegistry(),
		Weaver:      NewWeaver(),
		EnableCache: true,
	}
}

// AopManager AOP管理器
//
// 统一管理代码生成和运行时AOP，提供一致的接口
type AopManager struct {
	config     *AopConfig
	proxyCache sync.Map
	mu         sync.RWMutex
}

// NewAopManager 创建AOP管理器
func NewAopManager(config *AopConfig) *AopManager {
	if config == nil {
		config = DefaultAopConfig()
	}
	return &AopManager{
		config: config,
	}
}

// GetConfig 获取配置
func (m *AopManager) GetConfig() *AopConfig {
	return m.config
}

// RegisterAspect 注册切面
func (m *AopManager) RegisterAspect(aspect *AspectMeta) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config.Registry.RegisterAspect(aspect)
	if m.config.Weaver != nil {
		m.config.Weaver.AddAspects(aspect)
	}

	// 清除缓存
	m.proxyCache.Range(func(key, value any) bool {
		m.proxyCache.Delete(key)
		return true
	})
}

// RegisterAspects 批量注册切面
func (m *AopManager) RegisterAspects(aspects ...*AspectMeta) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, aspect := range aspects {
		m.config.Registry.RegisterAspect(aspect)
		if m.config.Weaver != nil {
			m.config.Weaver.AddAspects(aspect)
		}
	}

	// 清除缓存
	m.proxyCache.Range(func(key, value any) bool {
		m.proxyCache.Delete(key)
		return true
	})
}

// GetProxy 获取代理对象
func (m *AopManager) GetProxy(target any) any {
	if target == nil {
		return nil
	}

	targetType := reflect.TypeOf(target)
	cacheKey := fmt.Sprintf("%p_%v", target, targetType)

	// 检查缓存
	if m.config.EnableCache {
		if cached, ok := m.proxyCache.Load(cacheKey); ok {
			return cached
		}
	}

	var proxy any

	switch m.config.Mode {
	case AopModeRuntime:
		proxy = m.getRuntimeProxy(target)
	case AopModeGenerated:
		proxy = m.getGeneratedProxy(target)
	case AopModeMixed:
		proxy = m.getMixedProxy(target)
	default:
		proxy = m.getRuntimeProxy(target)
	}

	// 缓存结果
	if m.config.EnableCache && proxy != nil {
		m.proxyCache.Store(cacheKey, proxy)
	}

	return proxy
}

// getRuntimeProxy 获取运行时代理
func (m *AopManager) getRuntimeProxy(target any) any {
	if m.config.Weaver == nil {
		return target
	}
	return m.config.Weaver.Weave(target)
}

// getGeneratedProxy 获取代码生成的代理
func (m *AopManager) getGeneratedProxy(target any) any {
	// 这里需要与代码生成工具集成
	// 目前先返回运行时代理作为后备
	return m.getRuntimeProxy(target)
}

// getMixedProxy 获取混合模式代理
func (m *AopManager) getMixedProxy(target any) any {
	// 首先尝试获取代码生成的代理
	proxy := m.getGeneratedProxy(target)

	// 如果代码生成代理不可用，使用运行时代理
	if proxy == nil || proxy == target {
		proxy = m.getRuntimeProxy(target)
	}

	return proxy
}

// ClearCache 清除代理缓存
func (m *AopManager) ClearCache() {
	m.proxyCache.Range(func(key, value any) bool {
		m.proxyCache.Delete(key)
		return true
	})
}

// GetAspects 获取所有切面
func (m *AopManager) GetAspects() []*AspectMeta {
	return m.config.Registry.GetAspects()
}

// MatchAspectsForType 为类型匹配切面
func (m *AopManager) MatchAspectsForType(t reflect.Type) []*AspectMeta {
	return m.config.Registry.MatchAspectsForType(t)
}

// GlobalAopManager 全局AOP管理器
var GlobalAopManager = NewAopManager(nil)

// RegisterAspect 注册切面到全局管理器
func RegisterAspect(aspect *AspectMeta) {
	GlobalAopManager.RegisterAspect(aspect)
}

// RegisterAspects 批量注册切面到全局管理器
func RegisterAspects(aspects ...*AspectMeta) {
	GlobalAopManager.RegisterAspects(aspects...)
}

// GetProxy 获取代理对象（使用全局管理器）
func GetProxy(target any) any {
	return GlobalAopManager.GetProxy(target)
}

// ClearCache 清除全局代理缓存
func ClearCache() {
	GlobalAopManager.ClearCache()
}
