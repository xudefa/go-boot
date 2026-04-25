package aop

import (
	"reflect"
)

// Aspect 切面接口
//
// 定义切面的核心方法,返回切点信息.
// 实现此接口的类可以作为切面注册到AOP容器中.
type Aspect interface {
	// Pointcut 返回切点,定义哪些方法需要被拦截
	Pointcut() Pointcut
}

// AspectFunc 切面函数类型
//
// 作为简化方式,可以用函数代替结构体实现切面
type AspectFunc func(JoinPoint) interface{}

// aspect 切面实现
//
// 内部实现结构,用于存储切面元数据
type aspect struct {
	pc       Pointcut
	advice   Advice
	method   interface{}
	numIn    int
	numOut   int
	argTypes []reflect.Type
	retType  reflect.Type
}

func (a *aspect) Pointcut() Pointcut {
	return a.pc
}

func (a *aspect) Advice() Advice {
	return a.advice
}

func (a *aspect) Method() interface{} {
	return a.method
}

func (a *aspect) NumIn() int {
	return a.numIn
}

func (a *aspect) NumOut() int {
	return a.numOut
}

func (a *aspect) ArgTypes() []reflect.Type {
	return a.argTypes
}

func (a *aspect) RetType() reflect.Type {
	return a.retType
}

// AspectMeta 切面元数据
//
// 用于存储切面的实例、切点、通知和执行顺序.
// 字段说明:
//   - Instance: 切面实例
//   - Pointcut: 切点定义,匹配目标方法
//   - Advice: 通知,包含增强逻辑
//   - Order: 执行顺序,数字越小越先执行
type AspectMeta struct {
	Instance interface{}
	Pointcut Pointcut
	Advice   Advice
	Order    int
}

// aspectRegistry 切面注册表
//
// 内部结构,管理所有注册的切面
type aspectRegistry struct {
	aspects []*AspectMeta
}

func newAspectRegistry() *aspectRegistry {
	return &aspectRegistry{
		aspects: make([]*AspectMeta, 0),
	}
}

// Register 注册切面元数据
func (r *aspectRegistry) Register(aspectMeta *AspectMeta) {
	r.aspects = append(r.aspects, aspectMeta)
}

// GetAspects 获取所有已注册的切面
func (r *aspectRegistry) GetAspects() []*AspectMeta {
	return r.aspects
}

// MatchAspects 根据方法匹配切面
//
// 根据给定的方法,返回所有匹配的切面,并按Order排序
func (r *aspectRegistry) MatchAspects(method reflect.Method) []*AspectMeta {
	var matched []*AspectMeta
	for _, a := range r.aspects {
		if a.Pointcut.MatchMethod(method) {
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
