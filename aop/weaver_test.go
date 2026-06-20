package aop

import (
	"reflect"
	"testing"
)

func TestNewWeaver(t *testing.T) {
	t.Parallel()
	weaver := NewWeaver()
	if weaver == nil {
		t.Error("NewWeaver should return non-nil weaver")
	}
}

func TestWeaverWeaveWithoutAspects(t *testing.T) {
	t.Parallel()
	weaver := NewWeaver()
	service := &TestUserService{}

	result := weaver.Weave(service)
	if result != service {
		t.Error("Weave should return original target when no aspects registered")
	}
}

func TestWeaverWithBasicAspect(t *testing.T) {
	t.Parallel()
	weaver := NewWeaver()
	weaver.AddAspects(&AspectMeta{
		Instance: &TestAspect{},
		PointCut: MatchByNamePrefix("Do"),
		Advice: Around(func(jp JoinPoint, proceed ProceedFunc) any {
			return proceed(jp.Args()...)
		}),
		Order: 0,
	})

	service := &TestUserService{}
	weaved := weaver.Weave(service)
	proxy, ok := AsReflectiveProxy(weaved)
	if !ok {
		t.Fatal("expected ReflectiveAopProxy")
	}
	if _, err := proxy.Call("DoSomething"); err != nil {
		t.Errorf("DoSomething failed: %v", err)
	}
	if _, err := proxy.Call("DoAnother"); err != nil {
		t.Errorf("DoAnother failed: %v", err)
	}
}

func TestWeaverWithMultipleAspects(t *testing.T) {
	t.Parallel()
	weaver := NewWeaver()
	weaver.AddAspects(
		&AspectMeta{
			PointCut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) any { return proceed(jp.Args()...) }),
			Order:    2,
		},
		&AspectMeta{
			PointCut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) any { return proceed(jp.Args()...) }),
			Order:    1,
		},
	)

	service := &TestUserService{}
	weaved := weaver.Weave(service)
	proxy, ok := AsReflectiveProxy(weaved)
	if !ok {
		t.Fatal("expected ReflectiveAopProxy")
	}
	if _, err := proxy.Call("DoSomething"); err != nil {
		t.Errorf("DoSomething failed: %v", err)
	}
}

func TestWeaverWithOrderSorting(t *testing.T) {
	t.Parallel()
	weaver := NewWeaver()
	weaver.AddAspects(
		&AspectMeta{
			PointCut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) any { return proceed(jp.Args()...) }),
			Order:    10,
		},
		&AspectMeta{
			PointCut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) any { return proceed(jp.Args()...) }),
			Order:    5,
		},
		&AspectMeta{
			PointCut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) any { return proceed(jp.Args()...) }),
			Order:    1,
		},
	)

	service := &TestUserService{}
	weaved := weaver.Weave(service)
	proxy, ok := AsReflectiveProxy(weaved)
	if !ok {
		t.Fatal("expected ReflectiveAopProxy")
	}
	if _, err := proxy.Call("DoSomething"); err != nil {
		t.Errorf("DoSomething failed: %v", err)
	}
}

func TestWeaverWithInterfaceTarget(t *testing.T) {
	t.Parallel()
	weaver := NewWeaver()
	weaver.AddAspects(&AspectMeta{
		PointCut: MatchInterface((*TestServiceInterface)(nil)),
		Advice: Around(func(jp JoinPoint, proceed ProceedFunc) any {
			return proceed(jp.Args()...)
		}),
		Order: 0,
	})

	impl := &TestServiceImpl{}
	weaved := weaver.Weave(impl)
	proxy, ok := AsReflectiveProxy(weaved)
	if !ok {
		t.Fatal("expected ReflectiveAopProxy")
	}
	if _, err := proxy.Call("DoSomething"); err != nil {
		t.Errorf("DoSomething failed: %v", err)
	}
}

func TestWeaverWeaveNil(t *testing.T) {
	t.Parallel()
	weaver := NewWeaver()
	result := weaver.Weave(nil)
	if result != nil {
		t.Error("Weave should return nil for nil target")
	}
}

func TestAopRegistry(t *testing.T) {
	t.Parallel()
	registry := NewAopRegistry()
	if registry == nil {
		t.Error("NewAopRegistry should return non-nil registry")
	}

	aspect := &AspectMeta{
		PointCut: MatchAll(),
		Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) any { return proceed(jp.Args()...) }),
	}
	registry.RegisterAspect(aspect)

	aspects := registry.GetAspects()
	if len(aspects) != 1 {
		t.Errorf("expected 1 aspect, got %d", len(aspects))
	}
}

func TestAopRegistryMatchAspectsForType(t *testing.T) {
	t.Parallel()
	registry := NewAopRegistry()
	registry.RegisterAspect(&AspectMeta{
		PointCut: MatchByName("DoSomething"),
		Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) any { return proceed(jp.Args()...) }),
	})

	matched := registry.MatchAspectsForType(reflect.TypeFor[*TestUserService]())
	if len(matched) == 0 {
		t.Error("expected matched aspects")
	}
}

func TestAopRegistryWeaveIfNeeded(t *testing.T) {
	t.Parallel()
	registry := NewAopRegistry()
	testWeaver := NewWeaver()
	registry.RegisterWeaver("test", testWeaver)

	weaver, ok := registry.GetWeaver("test")
	if !ok || weaver == nil {
		t.Error("expected weaver")
	}

	result := registry.WeaveIfNeeded("test", &TestUserService{})
	if result == nil {
		t.Error("expected weaved result")
	}
}

type TestAspect struct{}

func (a *TestAspect) AspectName() string {
	return "test"
}

type TestUserService struct{}

func (s *TestUserService) DoSomething() {}

func (s *TestUserService) DoAnother() {}

type TestServiceInterface interface {
	DoSomething()
	DoAnother()
}

type TestServiceImpl struct{}

func (s *TestServiceImpl) DoSomething() {}

func (s *TestServiceImpl) DoAnother() {}
