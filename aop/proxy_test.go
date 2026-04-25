package aop

import (
	"reflect"
	"testing"
)

func TestNewProxyFactory(t *testing.T) {
	factory := NewProxyFactory(&TestUserService{})
	if factory == nil {
		t.Error("NewProxyFactory should return non-nil factory")
	}
}

func TestProxyFactory_SetAspects(t *testing.T) {
	factory := NewProxyFactory(&TestUserService{})
	aspects := []*AspectMeta{
		{
			Pointcut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) interface{} { return proceed(jp.Args()...) }),
			Order:    1,
		},
	}
	factory.SetAspects(aspects)

	if len(factory.aspects) != 1 {
		t.Errorf("SetAspects = %d, want 1", len(factory.aspects))
	}
}

func TestProxyFactory_GetProxy_WithoutAspects(t *testing.T) {
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
	factory := NewProxyFactory(&TestUserService{})
	factory.SetAspects([]*AspectMeta{
		{
			Pointcut: MatchByName("DoSomething"),
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
	var iface TestServiceInterface = &TestServiceImpl{}
	factory := NewProxyFactory(iface)
	factory.SetAspects([]*AspectMeta{
		{
			Pointcut: MatchInterface((*TestServiceInterface)(nil)),
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
	factory := NewProxyFactory(&TestUserService{})
	factory.SetAspects([]*AspectMeta{
		{
			Pointcut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) interface{} { return nil }),
			Order:    1,
		},
		{
			Pointcut: MatchByName("DoAnother"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) interface{} { return nil }),
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
	factory := NewProxyFactory(&TestUserService{})
	factory.SetAspects([]*AspectMeta{
		{
			Pointcut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) interface{} { return nil }),
			Order:    10,
		},
		{
			Pointcut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) interface{} { return nil }),
			Order:    1,
		},
		{
			Pointcut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) interface{} { return nil }),
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
	callOrder := []int{}

	advices := []Advice{
		Around(func(jp JoinPoint, proceed ProceedFunc) interface{} {
			callOrder = append(callOrder, 1)
			return proceed(jp.Args()...)
		}),
		Around(func(jp JoinPoint, proceed ProceedFunc) interface{} {
			callOrder = append(callOrder, 2)
			return proceed(jp.Args()...)
		}),
	}

	targetFunc := func(args ...interface{}) interface{} {
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
	factory := NewProxyFactory(&TestUserService{})
	factory.SetAspects([]*AspectMeta{
		{
			Pointcut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) interface{} { return nil }),
			Order:    1,
		},
	})

	m := reflect.Method{Name: "DoAnother"}
	matched := factory.filterAspects(m)

	if len(matched) != 0 {
		t.Errorf("filterAspects matched %d, want 0", len(matched))
	}
}

func TestProxyFactory_NonStructuralType(t *testing.T) {
	factory := NewProxyFactory("string")
	proxy := factory.GetProxy()

	if proxy != "string" {
		t.Error("GetProxy should return original for non-struct type")
	}
}
