package proxy

import (
	"reflect"
)

// ProxyFactory 代理工厂接口
type ProxyFactory interface {
	CreateProxy(target any, advisors []Advisor) any
}

// Advisor 切面顾问
type Advisor interface {
	GetPointcut() Pointcut
	GetAdvice() Advice
}

// Pointcut 切点接口
type Pointcut interface {
	Matches(method reflect.Method) bool
}

// Advice 通知接口
type Advice interface {
	Invoke(invocation MethodInvocation) (any, error)
}

// MethodInvocation 方法调用上下文
type MethodInvocation interface {
	GetMethod() reflect.Method
	GetArgs() []reflect.Value
	Proceed() (any, error)
	GetTarget() any
}

// defaultProxyFactory 默认代理工厂实现
type defaultProxyFactory struct{}

// NewProxyFactory 创建代理工厂
func NewProxyFactory() ProxyFactory {
	return &defaultProxyFactory{}
}

// CreateProxy 创建代理对象
func (f *defaultProxyFactory) CreateProxy(target any, advisors []Advisor) any {
	// TODO: 实现动态代理
	// 这里先返回原始对象，后续实现代理逻辑
	return target
}
