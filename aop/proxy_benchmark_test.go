package aop

import (
	"testing"
)

// BenchmarkProxyFactory_GetProxy 测试获取代理性能
func BenchmarkProxyFactory_GetProxy(b *testing.B) {
	factory := NewProxyFactory(&TestUserService{})
	factory.SetAspects([]*AspectMeta{
		{
			PointCut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) any { return proceed(jp.Args()...) }),
			Order:    1,
		},
	})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		factory.GetProxy()
	}
}

// BenchmarkProxyFactory_ProxyInvocation 测试代理调用性能
func BenchmarkProxyFactory_ProxyInvocation(b *testing.B) {
	factory := NewProxyFactory(&TestUserService{})
	factory.SetAspects([]*AspectMeta{
		{
			PointCut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) any { return proceed(jp.Args()...) }),
			Order:    1,
		},
	})

	proxy := factory.GetProxy()
	reflective, ok := AsReflectiveProxy(proxy)
	if !ok {
		b.Fatal("expected ReflectiveAopProxy")
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		reflective.MustCall("DoSomething")
	}
}

// BenchmarkProxyFactory_ProxyInvocation_MultipleAspects 测试多个切面的代理调用性能
func BenchmarkProxyFactory_ProxyInvocation_MultipleAspects(b *testing.B) {
	factory := NewProxyFactory(&TestUserService{})
	factory.SetAspects([]*AspectMeta{
		{
			PointCut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) any { return proceed(jp.Args()...) }),
			Order:    1,
		},
		{
			PointCut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) any { return proceed(jp.Args()...) }),
			Order:    2,
		},
		{
			PointCut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) any { return proceed(jp.Args()...) }),
			Order:    3,
		},
	})

	proxy := factory.GetProxy()
	reflective, ok := AsReflectiveProxy(proxy)
	if !ok {
		b.Fatal("expected ReflectiveAopProxy")
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		reflective.MustCall("DoSomething")
	}
}
