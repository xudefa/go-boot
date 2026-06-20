package aop

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Interceptor 拦截器函数
//
// 在通知链执行前后提供额外的处理逻辑，采用中间件模式。
// inv 为当前调用信息，next 为下一个处理函数。
//
// 示例:
//
//	aop.WithInterceptor(func(inv aop.Invocation, next func(aop.Invocation) any) any {
//	    start := time.Now()
//	    result := next(inv)
//	    slog.Info("method called", "name", inv.Signature().Name(), "duration", time.Since(start))
//	    return result
//	})
type Interceptor func(inv Invocation, next func(Invocation) any) any

// ChainExecutor 通知链执行器接口
//
// 定义通知链的执行策略。默认实现支持 panic 恢复、自定义拦截器和 context 传播。
// 可通过实现此接口自定义执行策略（如异步执行、限流等）。
type ChainExecutor interface {
	// Execute 执行通知链
	//
	// 参数:
	//   - inv: 调用信息
	//   - aspects: 切面元数据列表
	//   - targetFunc: 目标方法调用函数
	//
	// 返回值:
	//   - any: 方法执行结果
	Execute(inv Invocation, aspects []*AspectMeta, targetFunc func(...any) any) any
}

// classifiedAdvices 按类型分类后的通知列表
//
// 将切面列表中的通知按 AdviceType 分类存储，
// 避免在每次执行时重复遍历和分类。
type classifiedAdvices struct {
	before         []Advice // Before 前置通知列表
	after          []Advice // After 后置通知列表(无论是否异常都会执行)
	afterReturning []Advice // AfterReturning 返回后通知列表(仅在正常返回时执行)
	afterThrowing  []Advice // AfterThrowing 异常通知列表(仅在 panic 时执行)
	around         []Advice // Around 环绕通知列表
}

// classifyAdvices 将切面列表中的通知按类型分类
func classifyAdvices(aspects []*AspectMeta) *classifiedAdvices {
	ca := &classifiedAdvices{}
	for _, aspect := range aspects {
		if aspect == nil || aspect.Advice == nil {
			continue
		}
		switch aspect.Advice.Type() {
		case AdviceBefore:
			ca.before = append(ca.before, aspect.Advice)
		case AdviceAfter:
			ca.after = append(ca.after, aspect.Advice)
		case AdviceAfterReturning:
			ca.afterReturning = append(ca.afterReturning, aspect.Advice)
		case AdviceAfterThrowing:
			ca.afterThrowing = append(ca.afterThrowing, aspect.Advice)
		case AdviceAround:
			ca.around = append(ca.around, aspect.Advice)
		}
	}
	return ca
}

// chainExecutorConfig 执行器配置
//
// 存储通知链执行器的配置信息。
//
// 字段说明:
//   - recoverPanic: 是否启用 panic 恢复,启用后目标方法的 panic 会被捕获并触发 AfterThrowing 通知
//   - interceptors: 自定义拦截器列表,按添加顺序嵌套执行
type chainExecutorConfig struct {
	recoverPanic bool          // 是否启用 panic 恢复
	interceptors []Interceptor // 自定义拦截器列表
}

// ChainExecutorOption 执行器选项函数
//
// 用于通过函数式选项模式配置 ChainExecutor。
type ChainExecutorOption func(*chainExecutorConfig)

// WithRecovery 启用 panic 恢复
//
// 启用后，目标方法的 panic 会被捕获，
// afterThrowing 通知会正常执行，然后重新抛出 panic。
func WithRecovery() ChainExecutorOption {
	return func(c *chainExecutorConfig) {
		c.recoverPanic = true
	}
}

// WithInterceptor 添加自定义拦截器
//
// 拦截器按添加顺序嵌套：第一个添加的拦截器在最外层。
// 拦截器可以修改调用信息、观察结果、处理 panic 等。
func WithInterceptor(i Interceptor) ChainExecutorOption {
	return func(c *chainExecutorConfig) {
		c.interceptors = append(c.interceptors, i)
	}
}

// defaultChainExecutor 默认通知链执行器
//
// 实现了 ChainExecutor 接口，提供完整的通知链执行功能，
// 包括 panic 恢复、自定义拦截器和 context 传播。
type defaultChainExecutor struct {
	config chainExecutorConfig // 执行器配置
}

// NewChainExecutor 创建通知链执行器
//
// 默认启用 panic 恢复。可通过选项自定义行为。
//
// 示例:
//
//	executor := aop.NewChainExecutor(
//	    aop.WithInterceptor(tracingInterceptor),
//	    aop.WithInterceptor(metricsInterceptor),
//	)
//	aop.SetDefaultChainExecutor(executor)
func NewChainExecutor(opts ...ChainExecutorOption) ChainExecutor {
	config := chainExecutorConfig{
		recoverPanic: true,
	}
	for _, opt := range opts {
		opt(&config)
	}
	return &defaultChainExecutor{config: config}
}

// Execute 执行通知链
//
// 执行顺序:
//  1. Before 通知（按 Order 升序）
//  2. Around 通知链（如果存在），否则直接调用目标方法
//  3. After 通知（无论是否 panic 都会执行）
//  4. AfterThrowing 通知（仅在 panic 时执行）或 AfterReturning 通知（仅在正常返回时执行）
//
// 当启用 panic 恢复时:
//   - 目标方法的 panic 会被捕获
//   - After 通知始终执行
//   - AfterThrowing 通知在 panic 时执行
//   - 执行完 AfterThrowing 后重新抛出原始 panic
func (e *defaultChainExecutor) Execute(inv Invocation, aspects []*AspectMeta, targetFunc func(...any) any) any {
	if inv == nil || targetFunc == nil {
		return nil
	}

	// 按类型分类通知
	ca := classifyAdvices(aspects)

	// 核心执行逻辑: 执行 Before -> Around/目标方法 -> After -> AfterThrowing/AfterReturning
	coreExecute := func(invocation Invocation) any {
		// 1. 执行所有 Before 通知
		for _, advice := range ca.before {
			advice.Apply(invocation, nil)
		}

		var result any
		var panicked any

		// 2. 执行 Around 通知链或目标方法
		if e.config.recoverPanic {
			// 启用 panic 恢复: 捕获 panic 以确保 After/AfterThrowing 通知执行
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = r
					}
				}()
				if len(ca.around) > 0 {
					chain := buildAdviceChain(ca.around, targetFunc)
					result = chain(invocation)
				} else {
					result = targetFunc(invocation.Args()...)
				}
			}()
		} else {
			// 未启用 panic 恢复: panic 会直接传播
			if len(ca.around) > 0 {
				chain := buildAdviceChain(ca.around, targetFunc)
				result = chain(invocation)
			} else {
				result = targetFunc(invocation.Args()...)
			}
		}

		// 3. 执行所有 After 通知(无论是否 panic 都会执行)
		for _, advice := range ca.after {
			advice.Apply(invocation, nil)
		}

		// 4. 根据 panic 状态执行 AfterThrowing 或 AfterReturning
		if panicked != nil {
			// 将 panic 值转换为 error 传递给 AfterThrowing 通知
			var err error
			switch v := panicked.(type) {
			case error:
				err = v
			default:
				err = fmt.Errorf("%v", panicked)
			}
			for _, advice := range ca.afterThrowing {
				advice.Apply(invocation, func(...any) any { return err })
			}
			// 重新抛出原始 panic
			panic(panicked)
		}

		for _, advice := range ca.afterReturning {
			advice.Apply(invocation, func(...any) any { return result })
		}

		return result
	}

	// 从内到外包装拦截器,形成嵌套调用链
	interceptorCount := len(e.config.interceptors)
	execute := coreExecute
	for i := len(e.config.interceptors) - 1; i >= 0; i-- {
		interceptor := e.config.interceptors[i]
		prev := execute
		captured := interceptor
		execute = func(inv Invocation) any {
			return captured(inv, prev)
		}
	}

	// 执行通知链并更新统计
	var finalResult any
	func() {
		defer func() {
			if r := recover(); r != nil {
				updateStats(r, interceptorCount)
				panic(r)
			}
		}()
		finalResult = execute(inv)
	}()
	updateStats(nil, interceptorCount)
	return finalResult
}

// updateStats 更新通知链统计信息
func updateStats(panicked any, interceptorCount int) {
	GlobalChainStats.TotalExecutions.Add(1)
	if panicked != nil {
		GlobalChainStats.TotalPanics.Add(1)
	}
	if interceptorCount > 0 {
		GlobalChainStats.TotalInterceptors.Add(int64(interceptorCount))
	}
}

// buildAdviceChain 构建环绕通知链
//
// 将多个 Around 通知串联成一个调用链。
func buildAdviceChain(advices []Advice, targetFunc func(...any) any) func(Invocation) any {
	return func(inv Invocation) any {
		return executeAdviceChain(0, advices, inv, targetFunc)
	}
}

// executeAdviceChain 递归执行环绕通知链
//
// 如果 proceed 传递了自定义参数，则使用新参数替代原始参数。
// 新创建的 invocation 会保留原始的上下文信息。
func executeAdviceChain(idx int, advices []Advice, inv Invocation, targetFunc func(...any) any) any {
	if idx >= len(advices) {
		return targetFunc(inv.Args()...)
	}

	currentIdx := idx

	proceed := func(args ...any) any {
		if len(args) > 0 {
			// 使用自定义参数创建新的 invocation,保留原始上下文
			newInv := &invocation{
				method: inv.Method(),
				this:   inv.This(),
				target: inv.Target(),
				sig:    inv.Signature(),
				args:   args,
				ctx:    inv.Context(),
			}
			return executeAdviceChain(currentIdx+1, advices, newInv, targetFunc)
		}
		return executeAdviceChain(currentIdx+1, advices, inv, targetFunc)
	}

	return advices[idx].Apply(inv, proceed)
}

// defaultExecutor 全局默认通知链执行器
var defaultExecutor ChainExecutor = NewChainExecutor()
var defaultExecutorMu sync.RWMutex

// DefaultChainExecutor 获取默认通知链执行器
func DefaultChainExecutor() ChainExecutor {
	defaultExecutorMu.RLock()
	defer defaultExecutorMu.RUnlock()
	return defaultExecutor
}

// SetDefaultChainExecutor 设置默认通知链执行器
//
// 传入 nil 会被忽略。设置后，所有使用默认执行器的代码（包括 ExecuteChain 和 ReflectiveAopProxy）
// 都会使用新的执行器。并发安全。
func SetDefaultChainExecutor(executor ChainExecutor) {
	if executor == nil {
		return
	}
	defaultExecutorMu.Lock()
	defer defaultExecutorMu.Unlock()
	defaultExecutor = executor
}

// getDefaultExecutor 内部获取默认执行器（无锁，仅包内使用）
func getDefaultExecutor() ChainExecutor {
	defaultExecutorMu.RLock()
	defer defaultExecutorMu.RUnlock()
	return defaultExecutor
}

// ChainStats 通知链统计信息
//
// 用于收集通知链执行的运行时统计，所有字段均为原子操作，并发安全。
//
// 字段说明:
//   - TotalExecutions: 总执行次数
//   - TotalPanics: 总 panic 次数
//   - TotalInterceptors: 总拦截器调用次数
type ChainStats struct {
	TotalExecutions   atomic.Int64 // 总执行次数
	TotalPanics       atomic.Int64 // 总 panic 次数
	TotalInterceptors atomic.Int64 // 总拦截器调用次数
}

// GlobalChainStats 全局通知链统计
var GlobalChainStats ChainStats
