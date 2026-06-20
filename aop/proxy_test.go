package aop

import (
	"reflect"
	"testing"
)

func TestNewProxyFactory(t *testing.T) {
	t.Parallel()
	factory := NewProxyFactory(&TestUserService{})
	if factory == nil {
		t.Error("NewProxyFactory should return non-nil factory")
	}
}

func TestProxyFactory_SetAspects(t *testing.T) {
	t.Parallel()
	factory := NewProxyFactory(&TestUserService{})
	aspects := []*AspectMeta{
		{
			PointCut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) any { return proceed(jp.Args()...) }),
			Order:    1,
		},
	}
	factory.SetAspects(aspects)

	if len(factory.aspects) != 1 {
		t.Errorf("SetAspects = %d, want 1", len(factory.aspects))
	}
}

func TestProxyFactory_GetProxy_WithoutAspects(t *testing.T) {
	t.Parallel()
	factory := NewProxyFactory(&TestUserService{})
	proxy := factory.GetProxy()

	if proxy != nil {
		_, ok := proxy.(*TestUserService)
		if !ok {
			t.Error("GetProxy should return target when no aspects")
		}
	}
}

func TestProxyFactory_GetProxy_WithAspects(t *testing.T) {
	t.Parallel()
	factory := NewProxyFactory(&TestUserService{})
	factory.SetAspects([]*AspectMeta{
		{
			PointCut: MatchByName("DoSomething"),
			Advice:   Before(func(jp JoinPoint) {}),
			Order:    1,
		},
	})

	proxy := factory.GetProxy()
	if proxy == nil {
		t.Error("GetProxy should return proxy when aspects exist")
	}
}

func TestProxyFactory_GetProxy_Interface(t *testing.T) {
	t.Parallel()
	var aInterface TestServiceInterface = &TestServiceImpl{}
	factory := NewProxyFactory(aInterface)
	factory.SetAspects([]*AspectMeta{
		{
			PointCut: MatchInterface((*TestServiceInterface)(nil)),
			Advice:   Before(func(jp JoinPoint) {}),
			Order:    1,
		},
	})

	proxy := factory.GetProxy()
	if proxy == nil {
		t.Error("GetProxy should return proxy for interface")
	}
}

func TestProxyFactory_filterAspects(t *testing.T) {
	t.Parallel()
	factory := NewProxyFactory(&TestUserService{})
	factory.SetAspects([]*AspectMeta{
		{
			PointCut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) any { return nil }),
			Order:    1,
		},
		{
			PointCut: MatchByName("DoAnother"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) any { return nil }),
			Order:    2,
		},
	})

	m := reflect.Method{Name: "DoSomething"}
	matched := factory.filterAspects(m)

	if len(matched) != 1 {
		t.Errorf("filterAspects matched %d, want 1", len(matched))
	}
}

func TestProxyFactory_filterAspects_SortedByOrder(t *testing.T) {
	t.Parallel()
	factory := NewProxyFactory(&TestUserService{})
	factory.SetAspects([]*AspectMeta{
		{
			PointCut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) any { return nil }),
			Order:    10,
		},
		{
			PointCut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) any { return nil }),
			Order:    1,
		},
		{
			PointCut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) any { return nil }),
			Order:    5,
		},
	})

	m := reflect.Method{Name: "DoSomething"}
	matched := factory.filterAspects(m)

	if len(matched) != 3 {
		t.Fatalf("filterAspects matched %d, want 3", len(matched))
	}

	if matched[0].Order != 1 || matched[1].Order != 5 || matched[2].Order != 10 {
		t.Error("Aspects should be sorted by Order")
	}
}

func TestProxyFactory_buildAdviceChain(t *testing.T) {
	t.Parallel()
	callOrder := []int{}

	advices := []Advice{
		Around(func(jp JoinPoint, proceed ProceedFunc) any {
			callOrder = append(callOrder, 1)
			return proceed(jp.Args()...)
		}),
		Around(func(jp JoinPoint, proceed ProceedFunc) any {
			callOrder = append(callOrder, 2)
			return proceed(jp.Args()...)
		}),
	}

	targetFunc := func(args ...any) any {
		callOrder = append(callOrder, 0)
		return "result"
	}

	chain := buildAdviceChain(advices, targetFunc)
	inv := &invocation{args: nil, target: nil}

	result := chain(inv)

	if result != "result" {
		t.Errorf("chain result = %v, want result", result)
	}

	if len(callOrder) != 3 {
		t.Errorf("callOrder = %v, want [1, 2, 0]", callOrder)
	}
}

func TestProxyFactory_filterAspects_NoMatch(t *testing.T) {
	t.Parallel()
	factory := NewProxyFactory(&TestUserService{})
	factory.SetAspects([]*AspectMeta{
		{
			PointCut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) any { return nil }),
			Order:    1,
		},
	})

	m := reflect.Method{Name: "DoAnother"}
	matched := factory.filterAspects(m)

	if len(matched) != 0 {
		t.Errorf("filterAspects matched %d, want 0", len(matched))
	}
}

func TestProxyFactory_SetAspects_ClearsMethodCache(t *testing.T) {
	factory := NewProxyFactory(&TestUserService{})
	factory.SetAspects([]*AspectMeta{
		{
			PointCut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) any { return nil }),
			Order:    1,
		},
	})

	m := reflect.Method{Name: "DoSomething"}
	matched := factory.filterAspects(m)
	if len(matched) != 1 {
		t.Fatalf("expected 1 matched aspect, got %d", len(matched))
	}

	// 替换 aspects，模拟 SetAspects 清除缓存
	factory.SetAspects([]*AspectMeta{
		{
			PointCut: MatchByName("DoAnother"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) any { return nil }),
			Order:    1,
		},
	})

	// 即使 DoSomething 之前被缓存了，SetAspects 清除了缓存，
	// 重新 filter 应该匹配不到（aspect 已改为 DoAnother）
	matched = factory.filterAspects(m)
	if len(matched) != 0 {
		t.Errorf("after SetAspects with different aspects, expected 0 matched, got %d", len(matched))
	}
}

func TestProxyFactory_NonStructuralType(t *testing.T) {
	t.Parallel()
	factory := NewProxyFactory("string")
	proxy := factory.GetProxy()

	if proxy != "string" {
		t.Error("GetProxy should return original for non-struct type")
	}
}
