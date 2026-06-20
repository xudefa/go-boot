package aop

import (
	"sync"
	"testing"
)

// TestProxyFactory_ConcurrentGetProxy 测试并发获取代理
func TestProxyFactory_ConcurrentGetProxy(t *testing.T) {
	t.Parallel()
	factory := NewProxyFactory(&TestUserService{})
	factory.SetAspects([]*AspectMeta{
		{
			PointCut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) any { return proceed(jp.Args()...) }),
			Order:    1,
		},
	})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			proxy := factory.GetProxy()
			if proxy == nil {
				t.Error("GetProxy should return non-nil proxy")
			}
		}()
	}

	wg.Wait()
}

// TestProxyFactory_ConcurrentSetAspects 测试并发设置切面
func TestProxyFactory_ConcurrentSetAspects(t *testing.T) {
	t.Parallel()
	factory := NewProxyFactory(&TestUserService{})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			aspects := []*AspectMeta{
				{
					PointCut: MatchByName("DoSomething"),
					Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) any { return proceed(jp.Args()...) }),
					Order:    id,
				},
			}
			factory.SetAspects(aspects)
		}(i)
	}

	wg.Wait()
}

// TestProxyFactory_ConcurrentProxyInvocation 测试并发代理调用
func TestProxyFactory_ConcurrentProxyInvocation(t *testing.T) {
	t.Parallel()
	factory := NewProxyFactory(&TestUserService{})
	factory.SetAspects([]*AspectMeta{
		{
			PointCut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) any { return proceed(jp.Args()...) }),
			Order:    1,
		},
	})

	proxy := factory.GetProxy()
	reflectiveProxy, ok := AsReflectiveProxy(proxy)
	if !ok {
		t.Fatal("Proxy should be ReflectiveAopProxy")
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := reflectiveProxy.Call("DoSomething"); err != nil {
				t.Errorf("Call DoSomething failed: %v", err)
			}
		}()
	}

	wg.Wait()
}
