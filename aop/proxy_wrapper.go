package aop

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// InterfaceProxyWrapper 接口代理包装器
//
// 通过反射转发所有接口方法调用，支持 AOP 切面织入。
// 由于 Go 运行时无法动态替换接口方法，此包装器提供显式的 Invoke/InvokeContext 方法。
//
// 使用方式:
//
//	wrapper := aop.NewInterfaceProxyWrapper(target, advisors, iface)
//	result, err := wrapper.InvokeContext(ctx, "MethodName", arg1, arg2)
//
// 设计模式: Proxy
type InterfaceProxyWrapper struct {
	target      any
	advisors    []*AspectMeta
	iface       reflect.Type
	methodCache map[string]reflect.Method
	cacheMu     sync.RWMutex
	executor    ChainExecutor
}

// NewInterfaceProxyWrapper 创建接口代理包装器
func NewInterfaceProxyWrapper(target any, advisors []*AspectMeta, iface reflect.Type) *InterfaceProxyWrapper {
	return &InterfaceProxyWrapper{
		target:      target,
		advisors:    advisors,
		iface:       iface,
		methodCache: make(map[string]reflect.Method),
	}
}

// Invoke 调用接口方法
func (w *InterfaceProxyWrapper) Invoke(methodName string, args ...any) (any, error) {
	return w.InvokeContext(context.Background(), methodName, args...)
}

// InvokeContext 带上下文的方法调用
func (w *InterfaceProxyWrapper) InvokeContext(ctx context.Context, methodName string, args ...any) (any, error) {
	method, err := w.getMethod(methodName)
	if err != nil {
		return nil, err
	}

	targetFunc := func(callArgs ...any) any {
		val := reflect.ValueOf(w.target)
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
		this:   w.target,
		target: w.target,
		sig:    NewMethodSignature(methodName, w.iface),
		ctx:    ctx,
	}

	return w.getExecutor().Execute(inv, w.advisors, targetFunc), nil
}

// GetTarget 获取原始目标对象
func (w *InterfaceProxyWrapper) GetTarget() any {
	return w.target
}

// GetAdvisors 获取切面列表
func (w *InterfaceProxyWrapper) GetAdvisors() []*AspectMeta {
	return w.advisors
}

// SetExecutor 设置通知链执行器
func (w *InterfaceProxyWrapper) SetExecutor(executor ChainExecutor) {
	w.executor = executor
}

// getExecutor 获取执行器，优先使用自定义执行器
func (w *InterfaceProxyWrapper) getExecutor() ChainExecutor {
	if w.executor != nil {
		return w.executor
	}
	return getDefaultExecutor()
}

// getMethod 获取接口方法（带缓存）
func (w *InterfaceProxyWrapper) getMethod(methodName string) (reflect.Method, error) {
	w.cacheMu.RLock()
	if method, ok := w.methodCache[methodName]; ok {
		w.cacheMu.RUnlock()
		return method, nil
	}
	w.cacheMu.RUnlock()

	method, ok := w.iface.MethodByName(methodName)
	if !ok {
		return reflect.Method{}, fmt.Errorf("method %s not found on interface %s", methodName, w.iface.Name())
	}

	w.cacheMu.Lock()
	w.methodCache[methodName] = method
	w.cacheMu.Unlock()

	return method, nil
}
