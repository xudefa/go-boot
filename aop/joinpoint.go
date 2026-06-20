package aop

import (
	"context"
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
	Method() any
	// Args 获取方法调用时的参数
	Args() []any
	// Signature 获取方法签名
	Signature() MethodSignature
	// This 获取代理对象
	This() any
	// Target 获取目标对象
	Target() any
	// Context 获取调用上下文
	Context() context.Context
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

// methodSignature 方法签名内部实现
type methodSignature struct {
	name          string       // 方法名
	declaringType reflect.Type // 方法声明的类型
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
type ProceedFunc func(args ...any) any

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

	Proceed(args ...any) any
	// SetContext 设置调用上下文
	SetContext(ctx context.Context)
}

// invocation 调用信息内部实现
//
// 存储方法调用的完整上下文，包括方法本身、参数、代理对象、目标对象和继续执行函数。
type invocation struct {
	method  any             // 被调用的方法
	args    []any           // 调用参数列表
	this    any             // 代理对象
	target  any             // 原始目标对象
	sig     MethodSignature // 方法签名
	proceed ProceedFunc     // 继续执行函数（用于通知链）
	ctx     context.Context // 调用上下文
}

func (i *invocation) Method() any {
	return i.method
}

func (i *invocation) Args() []any {
	return i.args
}

func (i *invocation) This() any {
	return i.this
}

func (i *invocation) Target() any {
	return i.target
}

func (i *invocation) Signature() MethodSignature {
	return i.sig
}

func (i *invocation) Proceed(args ...any) any {
	if i.proceed != nil {
		return i.proceed(args...)
	}
	return nil
}

func (i *invocation) SetProceed(p ProceedFunc) {
	i.proceed = p
}

func (i *invocation) Context() context.Context {
	if i.ctx != nil {
		return i.ctx
	}
	return context.Background()
}

func (i *invocation) SetContext(ctx context.Context) {
	i.ctx = ctx
}
