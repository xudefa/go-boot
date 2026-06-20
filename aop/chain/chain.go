package chain

import (
	"reflect"

	"github.com/xudefa/go-boot/aop/advice"
	"github.com/xudefa/go-boot/aop/pointcut"
)

// Advisor 顾问
type Advisor struct {
	Pointcut pointcut.Pointcut
	Advice   advice.Advice
}

// Matches 检查方法是否匹配
func (a *Advisor) Matches(method reflect.Method) bool {
	return a.Pointcut.Matches(method)
}

// MethodInvocation 方法调用实现
type MethodInvocation struct {
	Target any
	Method reflect.Method
	Args   []reflect.Value
	Chain  *AdviceChain
	Index  int
}

// GetMethod 实现 advice.MethodInvocation.GetMethod
func (m *MethodInvocation) GetMethod() reflect.Method {
	return m.Method
}

// GetArgs 实现 advice.MethodInvocation.GetArgs
func (m *MethodInvocation) GetArgs() []reflect.Value {
	return m.Args
}

// GetTarget 实现 advice.MethodInvocation.GetTarget
func (m *MethodInvocation) GetTarget() any {
	return m.Target
}

// Proceed 实现 advice.MethodInvocation.Proceed
func (m *MethodInvocation) Proceed() (any, error) {
	if m.Index >= len(m.Chain.advisors) {
		// 执行目标方法
		results := m.Method.Func.Call(m.Args)
		if len(results) == 0 {
			return nil, nil
		}
		if len(results) == 1 {
			return results[0].Interface(), nil
		}
		lastResult := results[len(results)-1]
		if lastResult.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			if !lastResult.IsNil() {
				return nil, lastResult.Interface().(error)
			}
		}
		return results[0].Interface(), nil
	}

	advisor := m.Chain.advisors[m.Index]
	m.Index++

	return advisor.Advice.Invoke(m)
}

// AdviceChain 通知链
type AdviceChain struct {
	advisors []Advisor
}

// NewAdviceChain 创建通知链
func NewAdviceChain(advisors []Advisor) *AdviceChain {
	return &AdviceChain{
		advisors: advisors,
	}
}

// CreateInvocation 创建方法调用
func (c *AdviceChain) CreateInvocation(target any, method reflect.Method, args []reflect.Value) *MethodInvocation {
	return &MethodInvocation{
		Target: target,
		Method: method,
		Args:   args,
		Chain:  c,
		Index:  0,
	}
}
