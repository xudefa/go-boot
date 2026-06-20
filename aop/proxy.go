// Package aop 的代理工厂实现.
//
// ProxyFactory 负责根据目标对象创建AOP代理.
// 支持两种代理方式:
//   - 接口代理: 目标对象实现接口时,创建接口代理
//   - 结构体代理: 目标对象是结构体时,创建结构体代理
//
// # 代理创建流程
//
// 1. 调用 NewProxyFactory 创建工厂
// 2. 使用 SetAspects 设置切面列表
// 3. 调用 GetProxy 获取代理对象
//
// # 通知执行顺序
//
// 当方法被调用时,通知按以下顺序执行:
//  1. 所有 Before 通知(按Order升序)
//  2. Around 通知链(如果有的话)
//  3. 目标方法
//  4. 所有 After/AfterReturning/AfterThrowing 通知(按Order升序)
//
// # 缓存机制
//
// ProxyFactory 会缓存方法匹配的切面列表,避免每次调用都重新匹配和排序.
package aop

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// ProxyFactory 代理工厂
//
// 负责创建AOP代理对象.
// 根据目标对象的类型(接口或结构体),创建相应的代理.
//
// 字段说明:
//   - target: 目标对象,可以是实例指针或值
//   - aspects: 切面元数据列表,包含切点、通知和顺序
//   - proxyType: 代理对象的类型
//   - isInterface: 标记目标是否为接口类型
//   - methodCache: 缓存方法名到切面列表的映射,提升性能
//   - cacheMu: 保护methodCache的读写锁
//   - aspectsMu: 保护aspects切片的互斥锁
type ProxyFactory struct {
	target      any                      // 目标对象
	aspects     []*AspectMeta            // 切面元数据列表
	proxyType   reflect.Type             // 代理类型
	isInterface bool                     // 是否为接口类型
	methodCache map[string][]*AspectMeta // 缓存方法对应的切面列表
	cacheMu     sync.RWMutex             // 保护methodCache的互斥锁
	aspectsMu   sync.RWMutex             // 保护aspects切片的读写锁
	executor    ChainExecutor            // 通知链执行器
}

// NewProxyFactory 创建代理工厂
//
// 参数:
//   - target: 目标对象,可以是实例指针或值
//
// 返回值:
//   - *ProxyFactory: 代理工厂实例
//
// 示例:
//
//	factory := aop.NewProxyFactory(&UserService{})
func NewProxyFactory(target any) *ProxyFactory {
	t := reflect.TypeOf(target)
	if t == nil {
		return &ProxyFactory{
			target:      target,
			proxyType:   nil,
			isInterface: false,
		}
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return &ProxyFactory{
		target:      target,
		proxyType:   t,
		isInterface: t.Kind() == reflect.Interface,
	}
}

// SetAspects 设置切面
//
// 参数:
//   - aspects: 切面元数据列表
func (p *ProxyFactory) SetAspects(aspects []*AspectMeta) {
	p.aspectsMu.Lock()
	defer p.aspectsMu.Unlock()
	p.aspects = make([]*AspectMeta, len(aspects))
	copy(p.aspects, aspects)
	// 清除方法缓存，避免旧切面的缓存结果被新切面污染
	p.cacheMu.Lock()
	p.methodCache = nil
	p.cacheMu.Unlock()
}

// SetExecutor 设置通知链执行器
//
// 设置后，由此工厂创建的 ReflectiveAopProxy 将使用指定的执行器。
// 传入 nil 表示使用全局默认执行器。
func (p *ProxyFactory) SetExecutor(executor ChainExecutor) {
	p.executor = executor
}

// GetProxy 获取代理对象
//
// 根据目标对象的类型,创建并返回代理对象.
// 如果没有匹配的切面,则返回原对象.
//
// 返回值:
//   - interface{}: 代理对象或原对象
func (p *ProxyFactory) GetProxy() any {
	if p.proxyType == nil {
		return p.target
	}

	targetVal := reflect.ValueOf(p.target)
	targetType := reflect.TypeOf(p.target)
	if targetType.Kind() == reflect.Pointer {
		targetType = targetType.Elem()
	}

	if p.isInterface {
		return p.createInterfaceProxy()
	}

	if targetType.Kind() == reflect.Struct {
		return p.createStructProxy(targetVal, targetType)
	}

	return p.target
}

// createInterfaceProxy 创建接口代理
func (p *ProxyFactory) createInterfaceProxy() any {
	if p.target == nil {
		return nil
	}

	p.aspectsMu.RLock()
	aspectsSnapshot := make([]*AspectMeta, len(p.aspects))
	copy(aspectsSnapshot, p.aspects)
	p.aspectsMu.RUnlock()

	iface := reflect.TypeOf(p.target)

	// 检查是否有匹配的切面
	hasMatched := false
	for i := range iface.NumMethod() {
		method := iface.Method(i)
		if len(p.filterAspects(method)) > 0 {
			hasMatched = true
			break
		}
	}

	if !hasMatched {
		return p.target
	}

	return NewInterfaceProxyWrapper(p.target, aspectsSnapshot, iface)
}

// ReflectiveAopProxy 基于反射的 AOP 代理
//
// 由于 Go 运行时无法动态替换结构体方法，ReflectiveAopProxy 通过反射
// 拦截方法调用并执行通知链。用户通过 Call/CallContext 方法调用目标方法，
// AOP 通知会自动执行。
//
// 使用方式:
//
//	proxy := weaver.Weave(target).(*aop.ReflectiveAopProxy)
//	result, err := proxy.Call("MethodName", arg1, arg2)
//	result, err := proxy.CallContext(ctx, "MethodName", arg1, arg2)
type ReflectiveAopProxy struct {
	target      any
	targetType  reflect.Type
	aspects     []*AspectMeta
	methodCache map[string][]*AspectMeta
	cacheMu     sync.RWMutex
	executor    ChainExecutor
}

// Target 返回原始目标对象
func (p *ReflectiveAopProxy) Target() any {
	return p.target
}

// getExecutor 获取执行器，优先使用自定义执行器
func (p *ReflectiveAopProxy) getExecutor() ChainExecutor {
	if p.executor != nil {
		return p.executor
	}
	return getDefaultExecutor()
}

// Call 通过反射调用目标方法并执行通知链
func (p *ReflectiveAopProxy) Call(methodName string, args ...any) (any, error) {
	return p.CallContext(context.Background(), methodName, args...)
}

// CallContext 通过反射调用目标方法并执行通知链（带 context）
//
// 参数:
//   - ctx: 上下文，可通过 JoinPoint.Context() 在通知中获取
//   - methodName: 方法名
//   - args: 方法参数
//
// 返回值:
//   - any: 方法返回值（多返回值时返回 []any）
//   - error: 调用错误（仅表示方法查找失败，不包含目标方法的 panic）
func (p *ReflectiveAopProxy) CallContext(ctx context.Context, methodName string, args ...any) (any, error) {
	method, ok := p.targetType.MethodByName(methodName)
	if !ok {
		return nil, fmt.Errorf("method %s not found on %s", methodName, p.targetType)
	}

	matchedAspects := p.getMatchedAspects(methodName, method)

	targetFunc := func(callArgs ...any) any {
		val := reflect.ValueOf(p.target)
		in := make([]reflect.Value, 0, len(callArgs)+1)
		in = append(in, val)
		for _, a := range callArgs {
			in = append(in, reflect.ValueOf(a))
		}
		results := method.Func.Call(in)
		switch len(results) {
		case 0:
			return nil
		case 1:
			return results[0].Interface()
		default:
			ret := make([]any, len(results))
			for i, r := range results {
				ret[i] = r.Interface()
			}
			return ret
		}
	}

	inv := &invocation{
		method: method.Func.Interface(),
		args:   args,
		this:   p.target,
		target: p.target,
		sig:    NewMethodSignature(methodName, p.targetType),
		ctx:    ctx,
	}

	return p.getExecutor().Execute(inv, matchedAspects, targetFunc), nil
}

// MustCall 调用目标方法，panic on error
func (p *ReflectiveAopProxy) MustCall(methodName string, args ...any) any {
	result, err := p.Call(methodName, args...)
	if err != nil {
		panic(err)
	}
	return result
}

// SetExecutor 设置通知链执行器
func (p *ReflectiveAopProxy) SetExecutor(executor ChainExecutor) {
	p.executor = executor
}

// getMatchedAspects 获取方法匹配的切面
func (p *ReflectiveAopProxy) getMatchedAspects(methodName string, method reflect.Method) []*AspectMeta {
	p.cacheMu.RLock()
	if p.methodCache != nil {
		if cached, ok := p.methodCache[methodName]; ok {
			p.cacheMu.RUnlock()
			return cached
		}
	}
	p.cacheMu.RUnlock()

	var matched []*AspectMeta
	for _, a := range p.aspects {
		if a != nil && a.PointCut != nil && a.PointCut.MatchMethod(method) {
			matched = append(matched, a)
		}
	}
	SortAspectsByOrder(matched)

	p.cacheMu.Lock()
	if p.methodCache == nil {
		p.methodCache = make(map[string][]*AspectMeta)
	}
	p.methodCache[methodName] = matched
	p.cacheMu.Unlock()

	return matched
}

// IsReflectiveProxy 检查对象是否为 ReflectiveAopProxy
func IsReflectiveProxy(obj any) bool {
	_, ok := obj.(*ReflectiveAopProxy)
	return ok
}

// AsReflectiveProxy 将对象转换为 ReflectiveAopProxy
func AsReflectiveProxy(obj any) (*ReflectiveAopProxy, bool) {
	proxy, ok := obj.(*ReflectiveAopProxy)
	return proxy, ok
}

// createStructProxy 创建结构体代理
func (p *ProxyFactory) createStructProxy(targetVal reflect.Value, targetType reflect.Type) any {
	proxyVal := reflect.New(targetType)

	if targetVal.Kind() == reflect.Pointer {
		proxyVal.Elem().Set(targetVal.Elem())
	} else {
		proxyVal.Elem().Set(targetVal)
	}

	ptrType := reflect.PointerTo(targetType)

	hasMatched := false
	for i := range ptrType.NumMethod() {
		method := ptrType.Method(i)
		if method.PkgPath != "" {
			continue
		}
		if len(p.filterAspects(method)) > 0 {
			hasMatched = true
			break
		}
	}

	if !hasMatched {
		return proxyVal.Interface()
	}

	p.aspectsMu.RLock()
	aspectsSnapshot := make([]*AspectMeta, len(p.aspects))
	copy(aspectsSnapshot, p.aspects)
	p.aspectsMu.RUnlock()

	return &ReflectiveAopProxy{
		target:      p.target,
		targetType:  ptrType,
		aspects:     aspectsSnapshot,
		methodCache: make(map[string][]*AspectMeta),
		executor:    p.executor,
	}
}

// filterAspects 过滤匹配的切面
//
// 根据方法匹配切面,并按Order排序.
// 使用缓存避免每次调用都重新匹配和排序.
func (p *ProxyFactory) filterAspects(method reflect.Method) []*AspectMeta {
	// 检查缓存
	p.cacheMu.RLock()
	if p.methodCache != nil {
		if cached, ok := p.methodCache[method.Name]; ok {
			p.cacheMu.RUnlock()
			return cached
		}
	}
	p.cacheMu.RUnlock()

	// 在锁内快照 aspects，避免 SetAspects 和 filterAspects 的 TOCTOU 竞态
	p.aspectsMu.RLock()
	aspectsSnapshot := make([]*AspectMeta, len(p.aspects))
	copy(aspectsSnapshot, p.aspects)
	p.aspectsMu.RUnlock()

	var matched []*AspectMeta
	for _, a := range aspectsSnapshot {
		if a != nil && a.PointCut != nil && a.PointCut.MatchMethod(method) {
			matched = append(matched, a)
		}
	}

	// 按Order排序
	SortAspectsByOrder(matched)

	// 存入缓存
	p.cacheMu.Lock()
	if p.methodCache == nil {
		p.methodCache = make(map[string][]*AspectMeta)
	}
	p.methodCache[method.Name] = matched
	p.cacheMu.Unlock()

	return matched
}
