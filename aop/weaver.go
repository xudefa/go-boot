package aop

import (
	"reflect"
)

// Weaver 织入器接口
//
// 负责将切面织入目标对象,生成代理对象.
// 类似于Spring中的AopProxyFactory.
type Weaver interface {
	// Weave 织入目标对象
	//
	// 将已注册的切面织入目标对象,返回代理对象.
	// 参数:
	//   - target: 目标对象
	// 返回值:
	//   - interface{}: 代理对象,如果无匹配的切面则返回原对象
	Weave(target interface{}) interface{}
	// AddAspects 添加切面
	//
	// 添加一个或多个切面到织入器.
	AddAspects(aspects ...*AspectMeta)
}

// weaver 织入器实现
type weaver struct {
	aspects []*AspectMeta
	factory *ProxyFactory
}

// NewWeaver 创建织入器
//
// 返回值:
//   - Weaver: 织入器实例
//
// 示例:
//
//	weaver := aop.NewWeaver()
//	weaver.AddAspects(aspectMeta)
//	proxy := weaver.Weave(&UserService{})
func NewWeaver() Weaver {
	return &weaver{
		aspects: make([]*AspectMeta, 0),
		factory: nil,
	}
}

// AddAspects 添加切面
func (w *weaver) AddAspects(aspects ...*AspectMeta) {
	w.aspects = append(w.aspects, aspects...)
}

// Weave 织入目标对象
func (w *weaver) Weave(target interface{}) interface{} {
	if target == nil {
		return nil
	}

	if len(w.aspects) == 0 {
		return target
	}

	factory := NewProxyFactory(target)
	factory.SetAspects(w.aspects)
	return factory.GetProxy()
}

// AopRegistry AOP注册表
//
// 管理所有切面和织入器的注册中心.
// 用于在IOC容器中集成AOP功能.
type AopRegistry struct {
	aspects   []*AspectMeta
	weavers   map[string]Weaver
	weaversMu map[string]bool
}

// NewAopRegistry 创建AOP注册表
//
// 返回值:
//   - *AopRegistry: 注册表实例
func NewAopRegistry() *AopRegistry {
	return &AopRegistry{
		aspects:   make([]*AspectMeta, 0),
		weavers:   make(map[string]Weaver),
		weaversMu: make(map[string]bool),
	}
}

// RegisterAspect 注册切面
func (r *AopRegistry) RegisterAspect(aspect *AspectMeta) {
	r.aspects = append(r.aspects, aspect)
}

// GetAspects 获取所有切面
func (r *AopRegistry) GetAspects() []*AspectMeta {
	return r.aspects
}

// RegisterWeaver 注册织入器
func (r *AopRegistry) RegisterWeaver(beanID string, weaver Weaver) {
	r.weavers[beanID] = weaver
}

// GetWeaver 获取织入器
func (r *AopRegistry) GetWeaver(beanID string) (Weaver, bool) {
	w, ok := r.weavers[beanID]
	return w, ok
}

// WeaveIfNeeded 按需织入
//
// 如果指定beanID有对应的织入器,则织入目标对象.
func (r *AopRegistry) WeaveIfNeeded(beanID string, target interface{}) interface{} {
	weaver, ok := r.weavers[beanID]
	if !ok {
		return target
	}
	return weaver.Weave(target)
}

// IsWeaved 检查是否已织入
func (r *AopRegistry) IsWeaved(beanID string) bool {
	return r.weaversMu[beanID]
}

// MarkWeaved 标记已织入
func (r *AopRegistry) MarkWeaved(beanID string) {
	r.weaversMu[beanID] = true
}

// MatchAspectsForType 为类型匹配切面
//
// 根据类型匹配所有适用的切面.
func (r *AopRegistry) MatchAspectsForType(t reflect.Type) []*AspectMeta {
	var matched []*AspectMeta
	for _, a := range r.aspects {
		if a.Pointcut == nil {
			continue
		}
		for i := 0; i < t.NumMethod(); i++ {
			m := t.Method(i)
			if a.Pointcut.MatchMethod(m) {
				matched = append(matched, a)
				break
			}
		}
	}
	return matched
}
