package aop

// BeforeDecorator 创建前置装饰器，在调用原函数前执行 before 钩子。
//
// 参数:
//   - f: 被装饰的函数，签名为 func(args ...any) []any
//   - before: 前置钩子函数，接收 f 的所有参数
//
// 返回值:
//   - func(args ...any) []any: 装饰后的函数
func BeforeDecorator(f func(args ...any) []any, before func(args ...any)) func(args ...any) []any {
	return func(args ...any) []any {
		before(args...)
		return f(args...)
	}
}

// AfterDecorator 创建后置装饰器，在调用原函数后执行 after 钩子。
//
// 参数:
//   - f: 被装饰的函数
//   - after: 后置钩子函数，接收 f 的所有参数和所有返回值
//
// 返回值:
//   - func(args ...any) []any: 装饰后的函数
func AfterDecorator(f func(args ...any) []any, after func(results []any, args ...any)) func(args ...any) []any {
	return func(args ...any) []any {
		results := f(args...)
		after(results, args...)
		return results
	}
}

// AroundFunc 环绕函数类型，完全控制原函数的执行。
//
// 参数:
//   - originalFunc: 被装饰的原始函数，接受 []any 参数并返回 []any 结果
//   - args: 原始函数的参数
//
// 返回值:
//   - []any: 最终的结果，由 around 函数决定是否调用 originalFunc 以及如何处理结果
type AroundFunc func(originalFunc func(args ...any) []any, args ...any) []any

// AroundDecorator 创建环绕装饰器，由 around 函数完全控制原函数的执行流程。
//
// 参数:
//   - f: 被装饰的函数
//   - around: 环绕函数，接收原始函数 f 和原始参数，决定是否调用 f 以及在调用前后插入逻辑
//
// 返回值:
//   - func(args ...any) []any: 装饰后的函数
func AroundDecorator(f func(args ...any) []any, around AroundFunc) func(args ...any) []any {
	return func(args ...any) []any {
		return around(f, args...)
	}
}
