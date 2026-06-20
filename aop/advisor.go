package aop

// Advisor 顾问接口
//
// 顾问是AOP中的基本单元,包含一个切点和一个通知.
// 类似于Spring中的Advisor概念.
type Advisor interface {
	// GetPointCut 获取切点
	GetPointCut() PointCut
	// GetAdvice 获取通知
	GetAdvice() Advice
	// Order 获取执行顺序
	Order() int
}

// advisor 顾问内部实现
//
// 存储切点、通知和执行顺序信息
type advisor struct {
	pointCut PointCut // 切点定义
	advice   Advice   // 通知实例
	order    int      // 执行顺序，值越小优先级越高
}

func (a *advisor) GetPointCut() PointCut {
	return a.pointCut
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
//   - pointCut: 切点
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
func NewAdvisor(pointCut PointCut, advice Advice, order ...int) Advisor {
	o := 0
	if len(order) > 0 {
		o = order[0]
	}
	return &advisor{
		pointCut: pointCut,
		advice:   advice,
		order:    o,
	}
}
