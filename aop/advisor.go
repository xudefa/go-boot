package aop

import (
	"reflect"
)

// Advisor 顾问接口
//
// 顾问是AOP中的基本单元,包含一个切点和一个通知.
// 类似于Spring中的Advisor概念.
type Advisor interface {
	// GetPointcut 获取切点
	GetPointcut() Pointcut
	// GetAdvice 获取通知
	GetAdvice() Advice
	// Order 获取执行顺序
	Order() int
}

// advisor 顾问实现
type advisor struct {
	pointcut Pointcut
	advice   Advice
	order    int
}

func (a *advisor) GetPointcut() Pointcut {
	return a.pointcut
}

func (a *advisor) GetAdvice() Advice {
	return a.advice
}

func (a *advisor) Order() int {
	return a.order
}

// NewAdvisor 创建顾问
//
// 参数:
//   - pointcut: 切点
//   - advice: 通知
//   - order: 可选的执行顺序,默认0
//
// 返回值:
//   - Advisor: 顾问实例
//
// 示例:
//
//	advisor := aop.NewAdvisor(
//	    aop.MatchByName("DoSomething"),
//	    aop.Before(func(jp aop.JoinPoint) { fmt.Println("before") }),
//	    1, // order
//	)
func NewAdvisor(pointcut Pointcut, advice Advice, order ...int) Advisor {
	o := 0
	if len(order) > 0 {
		o = order[0]
	}
	return &advisor{
		pointcut: pointcut,
		advice:   advice,
		order:    o,
	}
}

// AspectMetadata 切面元数据
//
// 存储切面的类型信息
//
// 字段说明:
//   - Type: 切面类型
//   - AspectClass: 切面类类型
type AspectMetadata struct {
	Type        reflect.Type
	AspectClass reflect.Type
}

// aspectMetadata 切面元数据实现
type aspectMetadata struct {
	aspectClass reflect.Type
}

func (m *aspectMetadata) GetType() reflect.Type {
	return m.aspectClass
}

func (m *aspectMetadata) GetAspectClass() reflect.Type {
	return m.aspectClass
}

func newAspectMetadata(t reflect.Type) *aspectMetadata {
	return &aspectMetadata{
		aspectClass: t,
	}
}
