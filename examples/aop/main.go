// 示例: go-boot/aop 面向切面编程使用指南
//
// 本示例演示 go-boot/aop 包的核心功能:
//
// 1. 创建织入器(Weaver)
// 2. 注册切面元数据(AspectMeta)
// 3. 使用不同的切点匹配方式
// 4. 执行顺序控制
//
// 运行方式:
//
//	cd examples/aop && go run .
package main

import (
	"fmt"
	"github.com/xudefa/go-boot/aop"
	"log"
	"reflect"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	fmt.Println("=== AOP Example ===")

	basicExample()
	aroundExample()
	orderExample()

	fmt.Println("=== AOP Example ===")
	return nil
}

// basicExample 演示基本的 Around 通知
// 使用 LoggingAspect 拦截以 "Do" 开头的方法调用
// 在方法执行前后打印日志
func basicExample() {
	fmt.Println("--- Basic Around Example ---")

	aspectMeta := &aop.AspectMeta{
		Instance: &LoggingAspect{},
		Pointcut: aop.MatchByNamePrefix("Do"),
		Advice: aop.Around(func(jp aop.JoinPoint, proceed aop.ProceedFunc) interface{} {
			result := proceed(jp.Args()...)
			return result
		}),
		Order: 0,
	}

	weaver := aop.NewWeaver()
	weaver.AddAspects(aspectMeta)

	service := &UserService{}
	weaved := weaver.Weave(service)

	s := weaved.(*UserService)
	s.DoSomething()
	s.DoAnother()
	fmt.Println("--- Basic Around Example ---")
}

// aroundExample 演示 Around 通知处理返回值
// 使用 TimingAspect 拦截所有方法，测量执行时间
func aroundExample() {
	fmt.Println("--- Around with Return Value ---")

	aspectMeta := &aop.AspectMeta{
		Instance: &TimingAspect{},
		Pointcut: aop.MatchAll(),
		Advice: aop.Around(func(jp aop.JoinPoint, proceed aop.ProceedFunc) interface{} {
			result := proceed(jp.Args()...)
			return result
		}),
		Order: 0,
	}

	weaver := aop.NewWeaver()
	weaver.AddAspects(aspectMeta)

	service := &Calculator{}
	weaved := weaver.Weave(service).(*Calculator)
	result := weaved.Add(1, 2)
	fmt.Printf("Final result: %d\n", result)
	fmt.Println("--- Around with Return Value ---")
}

// orderExample 演示多个切面的执行顺序
// Order 值越小越先执行
// 第一个切面(Order=1)在方法执行前先执行
// 第二个切面(Order=2)在方法执行后执行
func orderExample() {
	fmt.Println("--- Order Example ---")

	aspects := []*aop.AspectMeta{
		{
			Instance: &LoggingAspect{},
			Pointcut: aop.MatchByName("DoSomething"),
			Advice: aop.Around(func(jp aop.JoinPoint, proceed aop.ProceedFunc) interface{} {
				result := proceed(jp.Args()...)
				return result
			}),
			Order: 2,
		},
		{
			Instance: &LoggingAspect{},
			Pointcut: aop.MatchByName("DoSomething"),
			Advice: aop.Around(func(jp aop.JoinPoint, proceed aop.ProceedFunc) interface{} {
				fmt.Println("[Order 1] First aspect - Before")
				result := proceed(jp.Args()...)
				fmt.Println("[Order 1] First aspect - After")
				return result
			}),
			Order: 1,
		},
	}

	weaver := aop.NewWeaver()
	weaver.AddAspects(aspects...)

	service := &UserService{}
	weaved := weaver.Weave(service).(*UserService)
	weaved.DoSomething()
	fmt.Println("--- Order Example ---")
}

// pointcutExample 演示不同的切点匹配方式
// 支持按方法名、前缀、正则表达式、接口类型匹配
func pointcutExample() {
	fmt.Println("--- Pointcut Matching Example ---")

	aspects := []*aop.AspectMeta{
		{
			Instance: &LoggingAspect{},
			Pointcut: aop.MatchByName("DoSomething"),
			Advice: aop.Around(func(jp aop.JoinPoint, proceed aop.ProceedFunc) interface{} {
				fmt.Printf("[MatchByName] %s\n", jp.Signature().Name())
				return proceed(jp.Args()...)
			}),
			Order: 0,
		},
		{
			Instance: &LoggingAspect{},
			Pointcut: aop.MatchByRegex("Do.*"),
			Advice: aop.Around(func(jp aop.JoinPoint, proceed aop.ProceedFunc) interface{} {
				fmt.Printf("[MatchByRegex] %s\n", jp.Signature().Name())
				return proceed(jp.Args()...)
			}),
			Order: 1,
		},
		{
			Instance: &LoggingAspect{},
			Pointcut: aop.MatchInterface((*ServiceInterface)(nil)),
			Advice: aop.Around(func(jp aop.JoinPoint, proceed aop.ProceedFunc) interface{} {
				fmt.Printf("[MatchInterface] %s\n", jp.Signature().Name())
				return proceed(jp.Args()...)
			}),
			Order: 2,
		},
	}

	weaver := aop.NewWeaver()
	weaver.AddAspects(aspects...)

	impl := &ServiceImpl{}
	weaved := weaver.Weave(impl)
	s := weaved.(*ServiceImpl)
	s.DoSomething()
	s.DoAnother()
	fmt.Println("--- Pointcut Matching Example ---")
}

// UserService 用户服务示例
// 包含 DoSomething 和 DoAnother 两个方法
type UserService struct{}

// DoSomething 执行用户相关操作
func (s *UserService) DoSomething() {
	fmt.Println("  -> UserService.DoSomething executed")
}

// DoAnother 执行另一个用户操作
func (s *UserService) DoAnother() {
	fmt.Println("  -> UserService.DoAnother executed")
}

// ServiceInterface 服务接口定义
// 定义了 DoSomething 和 DoAnother 方法
type ServiceInterface interface {
	DoSomething()
	DoAnother()
}

// ServiceImpl ServiceInterface 的实现
type ServiceImpl struct{}

func (s *ServiceImpl) DoSomething() {
	fmt.Println("  -> ServiceImpl.DoSomething executed")
}

func (s *ServiceImpl) DoAnother() {
	fmt.Println("  -> ServiceImpl.DoAnother executed")
}

// Calculator 计算器示例
// 提供基本的数学运算
type Calculator struct{}

// Add 返回两个数的和
func (c *Calculator) Add(a, b int) int {
	fmt.Printf("  -> Calculate: %d + %d = %d\n", a, b, a+b)
	return a + b
}

// LoggingAspect 日志切面
// 用于在方法执行前后记录日志
type LoggingAspect struct{}

// AspectName 返回切面名称
func (a *LoggingAspect) AspectName() string {
	return "logging"
}

// TimingAspect 时间切面
// 用于测量方法执行时间
type TimingAspect struct{}

// AspectName 返回切面名称
func (a *TimingAspect) AspectName() string {
	return "timing"
}

// 用于编译时接口检查
var _ = reflect.TypeOf((*ServiceInterface)(nil)).Elem()
