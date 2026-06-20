// Package aop 提供了一个面向切面编程(AOP)框架,灵感来自Spring AOP.
//
// # 核心概念
//
//   - Advice(通知): 在特定连接点执行的增强逻辑,包括Before、After、Around等类型
//   - PointCut(切点): 定义哪些方法需要被拦截的匹配规则
//   - Advisor(顾问): 切点的组合,包含一个PointCut和一个Advice
//   - Aspect(切面): 切面的元数据,包含切点、通知和执行顺序
//   - Weaver(织入器): 将切面织入目标对象,生成代理对象
//
// # 通知类型
//
//   - AdviceBefore: 前置通知,在目标方法执行前调用
//   - AdviceAfter: 后置通知,在目标方法执行后调用(无论是否异常)
//   - AdviceAfterReturning: 返回通知,在目标方法正常返回后调用
//   - AdviceAfterThrowing: 异常通知,在目标方法抛出异常后调用
//   - AdviceAround: 环绕通知,包裹目标方法,可以控制方法是否执行
//
// # 快速开始
//
//	// 创建切点 - 匹配所有名为DoSomething的方法
//	pointCut := aop.MatchByName("DoSomething")
//
//	// 创建前置通知
//	beforeAdvice := aop.Before(func(jp aop.JoinPoint) {
//	    fmt.Println("方法执行前:", jp.Signature().Name())
//	})
//
//	// 创建切面元数据
//	aspect := &aop.AspectMeta{
//	    PointCut: pointCut,
//	    Advice:   beforeAdvice,
//	    Order:    1,
//	}
//
//	// 创建织入器并添加切面
//	weaver := aop.NewWeaver()
//	weaver.AddAspects(aspect)
//
//	// 织入目标对象
//	target := &UserService{}
//	proxy := weaver.Weave(target)
//
//	// 使用代理对象,通知会自动执行
//	proxy.DoSomething() // 会先打印"方法执行前: DoSomething"
//
// # 切点匹配
//
// 包提供了多种切点匹配方式:
//   - MatchAll(): 匹配所有方法
//   - MatchByName(name): 按方法名精确匹配
//   - MatchByNamePrefix(prefix): 按方法名前缀匹配
//   - MatchByRegex(pattern): 按正则表达式匹配
//   - MatchClass(matcher): 自定义类匹配
//   - MatchMethod(matcher): 自定义方法匹配
//   - MatchInterface(iface): 匹配实现指定接口的类型
//
// # 相关包
//
//   - aop.Advice: 通知接口和工厂函数
//   - aop.PointCut: 切点接口和匹配函数
//   - aop.Weaver: 织入器接口和实现
//   - aop.ProxyFactory: 代理工厂,创建AOP代理对象
package aop

// AdviceType 通知类型
//
// 定义AOP中不同的通知类型
type AdviceType string

const (
	// AdviceBefore 前置通知
	// 在目标方法执行之前调用
	AdviceBefore AdviceType = "before"
	// AdviceAfter 后置通知
	// 在目标方法执行之后调用,无论是否抛出异常
	AdviceAfter AdviceType = "after"
	// AdviceAfterReturning 返回通知
	// 在目标方法正常返回后调用
	AdviceAfterReturning AdviceType = "after_returning"
	// AdviceAfterThrowing 异常通知
	// 在目标方法抛出异常后调用
	AdviceAfterThrowing AdviceType = "after_throwing"
	// AdviceAround 环绕通知
	// 包裹目标方法,可以决定是否执行以及如何执行
	AdviceAround AdviceType = "around"
)

// Advice 通知接口
//
// 定义通知的核心行为.
// 通知是对目标方法的增强逻辑.
type Advice interface {
	// Type 返回通知类型
	Type() AdviceType
	// Apply 应用通知
	//
	// 参数:
	//   - jp: 连接点,包含方法调用信息
	//   - proceed: 继续执行函数,在Around通知中使用
	//
	// 返回值:
	//   - interface{}: 通知的返回值,通常用于Around通知的返回值
	Apply(jp JoinPoint, proceed ProceedFunc) any
}

// advice 通知实现
type advice struct {
	adviceType AdviceType
	fn         func(JoinPoint, ProceedFunc) any
}

func (a *advice) Type() AdviceType {
	return a.adviceType
}

func (a *advice) Apply(jp JoinPoint, proceed ProceedFunc) any {
	if a.fn != nil {
		return a.fn(jp, proceed)
	}
	return nil
}

// Before 创建前置通知
//
// 在目标方法执行之前执行增强逻辑.
//
// 参数:
//   - fn: 前置通知函数,接收JoinPoint参数
//
// 返回值:
//   - Advice: 前置通知实例
//
// 示例:
//
//	aop.Before(func(jp aop.JoinPoint) {
//	    fmt.Println("方法执行前:", jp.Signature().Name())
//	})
func Before(fn func(JoinPoint)) Advice {
	return &advice{
		adviceType: AdviceBefore,
		fn: func(jp JoinPoint, _ ProceedFunc) any {
			fn(jp)
			return nil
		},
	}
}

// After 创建后置通知
//
// 在目标方法执行之后执行增强逻辑,无论方法是否抛出异常.
//
// 参数:
//   - fn: 后置通知函数,接收JoinPoint参数
//
// 返回值:
//   - Advice: 后置通知实例
//
// 示例:
//
//	aop.After(func(jp aop.JoinPoint) {
//	    fmt.Println("方法执行后:", jp.Signature().Name())
//	})
func After(fn func(JoinPoint)) Advice {
	return &advice{
		adviceType: AdviceAfter,
		fn: func(jp JoinPoint, _ ProceedFunc) any {
			fn(jp)
			return nil
		},
	}
}

// AfterReturning 创建返回通知
//
// 在目标方法正常返回后执行增强逻辑.
//
// 参数:
//   - fn: 返回通知函数,接收JoinPoint和返回值
//
// 返回值:
//   - Advice: 返回通知实例
//
// 示例:
//
//	aop.AfterReturning(func(jp aop.JoinPoint, result interface{}) {
//	    fmt.Println("方法返回:", result)
//	})
func AfterReturning(fn func(JoinPoint, any)) Advice {
	return &advice{
		adviceType: AdviceAfterReturning,
		fn: func(jp JoinPoint, proceed ProceedFunc) any {
			var result any
			if proceed != nil {
				result = proceed()
			}
			fn(jp, result)
			return result
		},
	}
}

// AfterThrowing 创建异常通知
//
// 在目标方法抛出异常后执行增强逻辑.
//
// 参数:
//   - fn: 异常通知函数,接收JoinPoint和错误
//
// 返回值:
//   - Advice: 异常通知实例
//
// 示例:
//
//	aop.AfterThrowing(func(jp aop.JoinPoint, err error) {
//	    fmt.Println("方法异常:", err)
//	})
func AfterThrowing(fn func(JoinPoint, error)) Advice {
	return &advice{
		adviceType: AdviceAfterThrowing,
		fn: func(jp JoinPoint, proceed ProceedFunc) any {
			var err error
			if proceed != nil {
				result := proceed()
				if result != nil {
					// 按返回值位置查找 error（通常 error 是最后一个返回值）
					if multiResult, ok := result.([]any); ok && len(multiResult) > 0 {
						// 从后往前查找，优先取最后一个 error
						for i := len(multiResult) - 1; i >= 0; i-- {
							if e, ok := multiResult[i].(error); ok {
								err = e
								break
							}
						}
					} else if e, ok := result.(error); ok {
						err = e
					}
				}
			}
			fn(jp, err)
			return nil
		},
	}
}

// Around 创建环绕通知
//
// 包裹目标方法,可以完全控制方法的执行.
//
// 参数:
//   - fn: 环绕通知函数,接收JoinPoint和ProceedFunc
//     可以通过ProceedFunc调用目标方法
//
// 返回值:
//   - Advice: 环绕通知实例
//
// 示例:
//
//	aop.Around(func(jp aop.JoinPoint, proceed aop.ProceedFunc) interface{} {
//	    fmt.Println("方法执行前:", jp.Signature().Name())
//	    result := proceed(jp.Args()...)
//	    fmt.Println("方法执行后:", result)
//	    return result
//	})
func Around(fn func(JoinPoint, ProceedFunc) any) Advice {
	return &advice{
		adviceType: AdviceAround,
		fn:         fn,
	}
}
