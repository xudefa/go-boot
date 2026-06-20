package aop

import (
	"context"
	"reflect"
)

// MethodInvocation 方法调用信息
//
// 用于代码生成的代理类，包含方法调用所需的所有信息。
// 新增字段 Proxy、Ctx 可按需设置，未设置时保持零值兼容。
//
// 字段说明:
//   - MethodName: 方法名称
//   - Func: 目标方法(函数值)
//   - Params: 方法调用参数列表
//   - Object: 目标对象(被代理的原始对象)
//   - Proxy: 代理对象本身,未设置时 This() 返回 nil
//   - Ctx: 上下文信息,未设置时 Context() 返回 context.Background()
//   - proceed: 继续执行函数,用于Around通知中控制执行流程
type MethodInvocation struct {
	MethodName string          // 方法名称
	Func       any             // 目标方法(函数值)
	Params     []any           // 方法调用参数列表
	Object     any             // 目标对象(被代理的原始对象)
	Proxy      any             // 代理对象本身,未设置时 This() 返回 nil
	Ctx        context.Context // 上下文信息,未设置时 Context() 返回 context.Background()
	proceed    ProceedFunc     // 继续执行函数,用于Around通知中控制执行流程
}

// Method 获取方法
func (m *MethodInvocation) Method() any {
	return m.Func
}

// Args 获取参数
func (m *MethodInvocation) Args() []any {
	return m.Params
}

// This 获取代理对象
func (m *MethodInvocation) This() any {
	return m.Proxy
}

// Target 获取目标对象
func (m *MethodInvocation) Target() any {
	return m.Object
}

// Signature 获取方法签名
func (m *MethodInvocation) Signature() MethodSignature {
	if m.Func == nil {
		return nil
	}
	fnValue := reflect.ValueOf(m.Func)
	fnType := fnValue.Type()
	return NewMethodSignature(m.MethodName, fnType)
}

// Context 获取上下文
//
// 如果已设置上下文,则返回该上下文;否则返回 context.Background()。
func (m *MethodInvocation) Context() context.Context {
	if m.Ctx != nil {
		return m.Ctx
	}
	return context.Background()
}

// SetContext 设置上下文
//
// 设置后,后续通知链中的 JoinPoint.Context() 将返回新的上下文。
func (m *MethodInvocation) SetContext(ctx context.Context) {
	m.Ctx = ctx
}

// Proceed 继续执行
//
// 如果已通过 SetProceed 设置了执行函数，则调用它；
// 否则通过反射调用目标方法。
func (m *MethodInvocation) Proceed(args ...any) any {
	if m.proceed != nil {
		return m.proceed(args...)
	}
	return m.callMethod(args...)
}

// SetProceed 设置继续执行函数
//
// 在Around通知中,用于设置继续执行目标方法或下一个通知的函数。
func (m *MethodInvocation) SetProceed(p ProceedFunc) {
	m.proceed = p
}

// callMethod 通过反射调用目标方法
//
// 如果未传入参数,则使用 MethodInvocation.Params 作为调用参数。
// 返回值根据结果数量处理:无返回值返回 nil,单个返回值直接返回,多个返回值返回 []any。
func (m *MethodInvocation) callMethod(args ...any) any {
	if m.Func == nil {
		return nil
	}
	methodValue := reflect.ValueOf(m.Func)
	if methodValue.Kind() != reflect.Func {
		return nil
	}

	callArgs := args
	if len(callArgs) == 0 {
		callArgs = m.Params
	}

	var argValues []reflect.Value
	for _, arg := range callArgs {
		argValues = append(argValues, reflect.ValueOf(arg))
	}

	results := methodValue.Call(argValues)
	switch len(results) {
	case 0:
		return nil
	case 1:
		return results[0].Interface()
	default:
		resultSlice := make([]any, len(results))
		for i, r := range results {
			resultSlice[i] = r.Interface()
		}
		return resultSlice
	}
}

// ExecuteChain 执行通知链
//
// 为代码生成的代理类提供通知链执行功能。
// 按照切面的 Order 排序，通过默认 ChainExecutor 执行通知链。
//
// 参数:
//   - jp: 方法调用信息
//   - aspects: 切面元数据列表（指针类型）
//
// 返回值:
//   - any: 方法执行结果
func ExecuteChain(jp *MethodInvocation, aspects []*AspectMeta) any {
	if jp == nil || jp.Func == nil {
		return nil
	}

	SortAspectsByOrder(aspects)

	targetFunc := func(args ...any) any {
		return jp.callMethod(args...)
	}

	return getDefaultExecutor().Execute(jp, aspects, targetFunc)
}
