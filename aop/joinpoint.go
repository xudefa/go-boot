package aop

import (
	"reflect"
)

// JoinPoint 连接点
//
// AOP核心概念,代表程序执行的某个位置.
// 在AOP中,连接点通常指方法调用.
//
// 方法说明:
//   - Method: 获取被拦截的方法
//   - Args: 获取方法调用参数
//   - Signature: 获取方法签名信息
//   - This: 获取代理对象本身
//   - Target: 获取目标对象(被代理的原始对象)
type JoinPoint interface {
	// Method 获取被拦截的方法
	Method() interface{}
	// Args 获取方法调用时的参数
	Args() []interface{}
	// Signature 获取方法签名
	Signature() MethodSignature
	// This 获取代理对象
	This() interface{}
	// Target 获取目标对象
	Target() interface{}
}

// MethodSignature 方法签名
//
// 描述方法的元数据信息,包括方法名和声明类型
type MethodSignature interface {
	// Name 获取方法名
	Name() string
	// DeclaringType 获取方法声明的类型
	DeclaringType() reflect.Type
}

// methodSignature 方法签名实现
type methodSignature struct {
	name          string
	declaringType reflect.Type
}

func (m *methodSignature) Name() string {
	return m.name
}

func (m *methodSignature) DeclaringType() reflect.Type {
	return m.declaringType
}

// NewMethodSignature 创建方法签名
//
// 参数:
//   - name: 方法名
//   - t: 方法声明的类型
//
// 返回值:
//   - MethodSignature: 方法签名实例
func NewMethodSignature(name string, t reflect.Type) MethodSignature {
	return &methodSignature{
		name:          name,
		declaringType: t,
	}
}

// ProceedFunc 继续执行函数
//
// 在Around通知中,调用此函数可以继续执行目标方法或下一个通知.
// 参数为传递给目标方法的参数,返回值为目标方法的返回值.
type ProceedFunc func(args ...interface{}) interface{}

// Invocation 调用信息
//
// 继承自JoinPoint,并添加了Proceed方法.
// 用于在Around通知中控制方法的执行流程.
type Invocation interface {
	JoinPoint
	// Proceed 继续执行
	//
	// 调用此方法可以执行目标方法或通知链中的下一个通知.
	// 可以传递自定义参数,这些参数会传递给下游的调用.
	Proceed(args ...interface{}) interface{}
}

// invocation 调用信息实现
type invocation struct {
	method  interface{}
	args    []interface{}
	this    interface{}
	target  interface{}
	sig     MethodSignature
	proceed ProceedFunc
}

func (i *invocation) Method() interface{} {
	return i.method
}

func (i *invocation) Args() []interface{} {
	return i.args
}

func (i *invocation) This() interface{} {
	return i.this
}

func (i *invocation) Target() interface{} {
	return i.target
}

func (i *invocation) Signature() MethodSignature {
	return i.sig
}

func (i *invocation) Proceed(args ...interface{}) interface{} {
	if i.proceed != nil {
		return i.proceed(args...)
	}
	return nil
}

// newInvocation 创建调用信息实例
func newInvocation(method interface{}, args []interface{}, target interface{}, sig MethodSignature, proceed ProceedFunc) Invocation {
	return &invocation{
		method:  method,
		args:    args,
		this:    target,
		target:  target,
		sig:     sig,
		proceed: proceed,
	}
}
