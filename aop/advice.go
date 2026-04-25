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
	Apply(jp JoinPoint, proceed ProceedFunc) interface{}
}

// advice 通知实现
type advice struct {
	adviceType AdviceType
	fn         func(JoinPoint, ProceedFunc) interface{}
}

func (a *advice) Type() AdviceType {
	return a.adviceType
}

func (a *advice) Apply(jp JoinPoint, proceed ProceedFunc) interface{} {
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
		fn: func(jp JoinPoint, _ ProceedFunc) interface{} {
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
		fn: func(jp JoinPoint, _ ProceedFunc) interface{} {
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
func AfterReturning(fn func(JoinPoint, interface{})) Advice {
	return &advice{
		adviceType: AdviceAfterReturning,
		fn: func(jp JoinPoint, proceed ProceedFunc) interface{} {
			var result interface{}
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
		fn: func(jp JoinPoint, proceed ProceedFunc) interface{} {
			var err error
			if proceed != nil {
				result := proceed()
				if result != nil {
					if e, ok := result.(error); ok {
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
func Around(fn func(JoinPoint, ProceedFunc) interface{}) Advice {
	return &advice{
		adviceType: AdviceAround,
		fn:         fn,
	}
}
