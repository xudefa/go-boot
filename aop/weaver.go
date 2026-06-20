package aop

import (
	"reflect"
	"sync"
)

// Weaver 织入器接口
//
// 负责将切面织入目标对象,生成代理对象.
// 类似于Spring中的AopProxyFactory.
//
// # 工作流程
//
// 1. 创建织入器: NewWeaver()
// 2. 添加切面: AddAspects(aspect1, aspect2, ...)
// 3. 织入目标: Weave(target) -> 返回代理对象
//
// # 织入规则
//
//   - 如果目标对象没有任何匹配的切面,返回原对象(不进行代理)
//   - 如果目标对象有匹配的切面,创建代理对象
//   - 代理对象的方法调用会触发匹配的通知
type Weaver interface {
	// Weave 织入目标对象
	//
	// 将已注册的切面织入目标对象,返回代理对象.
	//
	// 参数:
	//   - target: 目标对象,可以是结构体指针或接口实现
	//
	// 返回值:
	//   - interface{}: 代理对象(如果匹配到切面)或原对象
	//
	// 注意:
	//   - 返回的对象类型可能和原对象不同(代理类型)
	//   - 使用代理对象调用方法时,匹配的通知会自动执行
	Weave(target any) any

	// AddAspects 添加切面
	//
	// 添加一个或多个切面到织入器.
	// 添加后,这些切面会应用于后续 Weave 调用.
	//
	// 参数:
	//   - aspects: 一个或多个切面元数据
	AddAspects(aspects ...*AspectMeta)
}

// weaver 织入器实现
type weaver struct {
	aspects []*AspectMeta
	factory *ProxyFactory
	mu      sync.RWMutex
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
	w.mu.Lock()
	defer w.mu.Unlock()
	w.aspects = append(w.aspects, aspects...)
}

// Weave 织入目标对象
func (w *weaver) Weave(target any) any {
	if target == nil {
		return nil
	}

	w.mu.RLock()
	if len(w.aspects) == 0 {
		w.mu.RUnlock()
		return target
	}
	aspects := make([]*AspectMeta, len(w.aspects))
	copy(aspects, w.aspects)
	w.mu.RUnlock()

	factory := NewProxyFactory(target)
	factory.SetAspects(aspects)
	return factory.GetProxy()
}

// AopRegistry AOP注册表
//
// 管理所有切面和织入器的注册中心.
// 用于在IOC容器中集成AOP功能.
type AopRegistry struct {
	aspects []*AspectMeta
	weavers map[string]Weaver
	mu      sync.RWMutex
}

// NewAopRegistry 创建AOP注册表
//
// 返回值:
//   - *AopRegistry: 注册表实例
func NewAopRegistry() *AopRegistry {
	return &AopRegistry{
		aspects: make([]*AspectMeta, 0),
		weavers: make(map[string]Weaver),
	}
}

// RegisterAspect 注册切面
func (r *AopRegistry) RegisterAspect(aspect *AspectMeta) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.aspects = append(r.aspects, aspect)
}

// GetAspects 获取所有切面
func (r *AopRegistry) GetAspects() []*AspectMeta {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*AspectMeta, len(r.aspects))
	copy(result, r.aspects)
	return result
}

// RegisterWeaver 注册织入器
func (r *AopRegistry) RegisterWeaver(beanID string, weaver Weaver) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.weavers[beanID] = weaver
}

// GetWeaver 获取织入器
func (r *AopRegistry) GetWeaver(beanID string) (Weaver, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	w, ok := r.weavers[beanID]
	return w, ok
}

// WeaveIfNeeded 按需织入
//
// 如果指定beanID有对应的织入器,则织入目标对象.
func (r *AopRegistry) WeaveIfNeeded(beanID string, target any) any {
	r.mu.RLock()
	weaver, ok := r.weavers[beanID]
	r.mu.RUnlock()
	if !ok {
		return target
	}
	return weaver.Weave(target)
}

// MatchAspectsForType 为类型匹配切面
//
// 根据类型匹配所有适用的切面,并按Order排序.
func (r *AopRegistry) MatchAspectsForType(t reflect.Type) []*AspectMeta {
	r.mu.RLock()
	aspects := make([]*AspectMeta, len(r.aspects))
	copy(aspects, r.aspects)
	r.mu.RUnlock()

	var matched []*AspectMeta
	for _, a := range aspects {
		if a.PointCut == nil {
			continue
		}
		for i := 0; i < t.NumMethod(); i++ {
			m := t.Method(i)
			if a.PointCut.MatchMethod(m) {
				matched = append(matched, a)
				break
			}
		}
	}
	SortAspectsByOrder(matched)
	return matched
}
