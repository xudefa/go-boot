package aop

import (
	"reflect"
)

// ProxyFactory 代理工厂
//
// 负责创建AOP代理对象.
// 根据目标对象的类型(接口或结构体),创建相应的代理.
type ProxyFactory struct {
	target      interface{}   // 目标对象
	aspects     []*AspectMeta // 切面元数据列表
	proxyType   reflect.Type  // 代理类型
	isInterface bool          // 是否为接口类型
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
func NewProxyFactory(target interface{}) *ProxyFactory {
	t := reflect.TypeOf(target)
	if t.Kind() == reflect.Ptr {
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
	p.aspects = aspects
}

// GetProxy 获取代理对象
//
// 根据目标对象的类型,创建并返回代理对象.
// 如果没有匹配的切面,则返回原对象.
//
// 返回值:
//   - interface{}: 代理对象或原对象
func (p *ProxyFactory) GetProxy() interface{} {
	targetVal := reflect.ValueOf(p.target)
	targetType := reflect.TypeOf(p.target)
	if targetType.Kind() == reflect.Ptr {
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
func (p *ProxyFactory) createInterfaceProxy() interface{} {
	targetType := reflect.TypeOf(p.target)
	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
	}

	var impl reflect.Value
	if reflect.TypeOf(p.target).Kind() == reflect.Ptr {
		impl = reflect.ValueOf(p.target).Elem()
	} else {
		impl = reflect.New(reflect.TypeOf(p.target).Elem()).Elem()
	}

	proxy := reflect.New(targetType).Interface()
	proxyVal := reflect.ValueOf(proxy).Elem()

	for i := 0; i < targetType.NumMethod(); i++ {
		method := targetType.Method(i)
		proxyMethod := p.wrapMethod(method, impl)
		proxyVal.MethodByName(method.Name).Set(proxyMethod)
	}

	return proxy
}

// createStructProxy 创建结构体代理
func (p *ProxyFactory) createStructProxy(targetVal reflect.Value, targetType reflect.Type) interface{} {
	proxyVal := reflect.New(targetType)

	if targetVal.Kind() == reflect.Ptr {
		proxyVal.Elem().Set(targetVal.Elem())
	} else {
		proxyVal.Elem().Set(targetVal)
	}

	for i := 0; i < targetType.NumMethod(); i++ {
		method := targetType.Method(i)
		if method.PkgPath != "" {
			continue
		}

		wrappedMethod := p.wrapMethod(method, proxyVal.Elem())
		proxyVal.MethodByName(method.Name).Set(wrappedMethod)
	}

	return proxyVal.Interface()
}

// wrapMethod 包装方法
//
// 为方法创建代理逻辑,包含通知的执行流程.
func (p *ProxyFactory) wrapMethod(method reflect.Method, targetVal reflect.Value) reflect.Value {
	matchedAspects := p.filterAspects(method)
	if len(matchedAspects) == 0 {
		return targetVal.MethodByName(method.Name)
	}

	var beforeAdvices []Advice
	var aroundAdvices []Advice
	var afterAdvices []Advice

	for _, aspect := range matchedAspects {
		switch aspect.Advice.Type() {
		case AdviceBefore:
			beforeAdvices = append(beforeAdvices, aspect.Advice)
		case AdviceAround:
			aroundAdvices = append(aroundAdvices, aspect.Advice)
		case AdviceAfterReturning, AdviceAfterThrowing, AdviceAfter:
			afterAdvices = append(afterAdvices, aspect.Advice)
		}
	}

	targetFunc := func(args ...interface{}) interface{} {
		if method.Type.NumOut() == 0 {
			targetVal.MethodByName(method.Name).Call(toValues(args))
			return nil
		}
		return targetVal.MethodByName(method.Name).Call(toValues(args))[0].Interface()
	}

	return reflect.MakeFunc(method.Type, func(args []reflect.Value) []reflect.Value {
		sig := NewMethodSignature(method.Name, method.Type)
		argInterfaces := toInterfaces(args)

		var result interface{}

		if len(aroundAdvices) > 0 {
			chain := buildAdviceChain(aroundAdvices, targetFunc)
			inv := newInvocation(method.Func, argInterfaces, targetVal.Interface(), sig, nil)
			result = chain(inv)
		} else {
			inv := newInvocation(method.Func, argInterfaces, targetVal.Interface(), sig, targetFunc)

			for _, advice := range beforeAdvices {
				advice.Apply(inv, nil)
			}

			result = targetFunc(argInterfaces...)

			for _, advice := range afterAdvices {
				advice.Apply(inv, nil)
			}
		}

		if method.Type.NumOut() == 0 {
			return nil
		}
		if result == nil {
			return []reflect.Value{reflect.Zero(method.Type.Out(0))}
		}
		return []reflect.Value{reflect.ValueOf(result)}
	})
}

// buildAdviceChain 构建通知链
//
// 将多个Around通知串联成一个调用链.
func buildAdviceChain(advices []Advice, targetFunc func(...interface{}) interface{}) func(inv Invocation) interface{} {
	return func(inv Invocation) interface{} {
		return executeAdviceChain(0, advices, inv, targetFunc)
	}
}

// executeAdviceChain 执行通知链
//
// 递归执行通知链中的每个通知.
func executeAdviceChain(idx int, advices []Advice, inv Invocation, targetFunc func(...interface{}) interface{}) interface{} {
	if idx >= len(advices) {
		return targetFunc(inv.Args()...)
	}

	currentIdx := idx

	proceed := func(args ...interface{}) interface{} {
		return executeAdviceChain(currentIdx+1, advices, inv, targetFunc)
	}

	return advices[idx].Apply(inv, proceed)
}

// filterAspects 过滤匹配的切面
//
// 根据方法匹配切面,并按Order排序.
func (p *ProxyFactory) filterAspects(method reflect.Method) []*AspectMeta {
	var matched []*AspectMeta
	for _, a := range p.aspects {
		if a.Pointcut != nil && a.Pointcut.MatchMethod(method) {
			matched = append(matched, a)
		}
	}
	for i := 0; i < len(matched)-1; i++ {
		for j := 0; j < len(matched)-1-i; j++ {
			if matched[j].Order > matched[j+1].Order {
				matched[j], matched[j+1] = matched[j+1], matched[j]
			}
		}
	}
	return matched
}

func toInterfaces(vals []reflect.Value) []interface{} {
	result := make([]interface{}, len(vals))
	for i, v := range vals {
		result[i] = v.Interface()
	}
	return result
}

func toValues(args []interface{}) []reflect.Value {
	result := make([]reflect.Value, len(args))
	for i, a := range args {
		if a == nil {
			result[i] = reflect.Value{}
		} else {
			result[i] = reflect.ValueOf(a)
		}
	}
	return result
}
