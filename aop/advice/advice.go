package advice

import (
	"reflect"
)

// Advice 基础通知接口
type Advice interface {
	Invoke(invocation MethodInvocation) (any, error)
}

// BeforeAdvice 前置通知
type BeforeAdvice interface {
	Before(method reflect.Method, args []reflect.Value) error
}

// AfterAdvice 后置通知
type AfterAdvice interface {
	After(method reflect.Method, result any, err error) error
}

// AroundAdvice 环绕通知
type AroundAdvice interface {
	Invoke(invocation MethodInvocation) (any, error)
}

// AfterReturningAdvice 返回后通知
type AfterReturningAdvice interface {
	AfterReturning(method reflect.Method, result any) error
}

// AfterThrowingAdvice 异常后通知
type AfterThrowingAdvice interface {
	AfterThrowing(method reflect.Method, err error) error
}

// MethodInvocation 方法调用接口
type MethodInvocation interface {
	GetMethod() reflect.Method
	GetArgs() []reflect.Value
	Proceed() (any, error)
	GetTarget() any
}

// SimpleMethodInvocation 简单方法调用实现
type SimpleMethodInvocation struct {
	Target any
	Method reflect.Method
	Args   []reflect.Value
}

// GetMethod 实现 MethodInvocation.GetMethod
func (m *SimpleMethodInvocation) GetMethod() reflect.Method {
	return m.Method
}

// GetArgs 实现 MethodInvocation.GetArgs
func (m *SimpleMethodInvocation) GetArgs() []reflect.Value {
	return m.Args
}

// Proceed 实现 MethodInvocation.Proceed
func (m *SimpleMethodInvocation) Proceed() (any, error) {
	results := m.Method.Func.Call(m.Args)
	if len(results) == 0 {
		return nil, nil
	}
	if len(results) == 1 {
		return results[0].Interface(), nil
	}
	// 最后一个返回值可能是 error
	lastResult := results[len(results)-1]
	if lastResult.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		if !lastResult.IsNil() {
			return nil, lastResult.Interface().(error)
		}
	}
	return results[0].Interface(), nil
}

// GetTarget 实现 MethodInvocation.GetTarget
func (m *SimpleMethodInvocation) GetTarget() any {
	return m.Target
}
