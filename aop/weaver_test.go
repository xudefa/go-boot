package aop

import (
	"reflect"
	"testing"
)

func TestNewWeaver(t *testing.T) {
	weaver := NewWeaver()
	if weaver == nil {
		t.Error("NewWeaver should return non-nil weaver")
	}
}

func TestWeaverWeaveWithoutAspects(t *testing.T) {
	weaver := NewWeaver()
	service := &TestUserService{}

	result := weaver.Weave(service)
	if result != service {
		t.Error("Weave should return original target when no aspects registered")
	}
}

func TestWeaverWithBasicAspect(t *testing.T) {
	weaver := NewWeaver()
	weaver.AddAspects(&AspectMeta{
		Instance: &TestAspect{},
		Pointcut: MatchByNamePrefix("Do"),
		Advice: Around(func(jp JoinPoint, proceed ProceedFunc) interface{} {
			return proceed(jp.Args()...)
		}),
		Order: 0,
	})

	service := &TestUserService{}
	weaved := weaver.Weave(service).(*TestUserService)

	weaved.DoSomething()
	weaved.DoAnother()
}

func TestWeaverWithMultipleAspects(t *testing.T) {
	weaver := NewWeaver()
	weaver.AddAspects(
		&AspectMeta{
			Pointcut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) interface{} { return proceed(jp.Args()...) }),
			Order:    2,
		},
		&AspectMeta{
			Pointcut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) interface{} { return proceed(jp.Args()...) }),
			Order:    1,
		},
	)

	service := &TestUserService{}
	weaved := weaver.Weave(service).(*TestUserService)
	weaved.DoSomething()
}

func TestWeaverWithOrderSorting(t *testing.T) {
	weaver := NewWeaver()
	weaver.AddAspects(
		&AspectMeta{
			Pointcut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) interface{} { return proceed(jp.Args()...) }),
			Order:    10,
		},
		&AspectMeta{
			Pointcut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) interface{} { return proceed(jp.Args()...) }),
			Order:    5,
		},
		&AspectMeta{
			Pointcut: MatchByName("DoSomething"),
			Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) interface{} { return proceed(jp.Args()...) }),
			Order:    1,
		},
	)

	service := &TestUserService{}
	weaved := weaver.Weave(service).(*TestUserService)
	weaved.DoSomething()
}

func TestWeaverWithInterfaceTarget(t *testing.T) {
	weaver := NewWeaver()
	weaver.AddAspects(&AspectMeta{
		Pointcut: MatchInterface((*TestServiceInterface)(nil)),
		Advice: Around(func(jp JoinPoint, proceed ProceedFunc) interface{} {
			return proceed(jp.Args()...)
		}),
		Order: 0,
	})

	impl := &TestServiceImpl{}
	weaved := weaver.Weave(impl).(*TestServiceImpl)
	weaved.DoSomething()
}

func TestWeaverWeaveNil(t *testing.T) {
	weaver := NewWeaver()
	result := weaver.Weave(nil)
	if result != nil {
		t.Error("Weave should return nil for nil target")
	}
}

func TestAopRegistry(t *testing.T) {
	registry := NewAopRegistry()
	if registry == nil {
		t.Error("NewAopRegistry should return non-nil registry")
	}

	aspect := &AspectMeta{
		Pointcut: MatchAll(),
		Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) interface{} { return proceed(jp.Args()...) }),
	}
	registry.RegisterAspect(aspect)

	aspects := registry.GetAspects()
	if len(aspects) != 1 {
		t.Errorf("expected 1 aspect, got %d", len(aspects))
	}
}

func TestAopRegistryMatchAspectsForType(t *testing.T) {
	registry := NewAopRegistry()
	registry.RegisterAspect(&AspectMeta{
		Pointcut: MatchByName("DoSomething"),
		Advice:   Around(func(jp JoinPoint, proceed ProceedFunc) interface{} { return proceed(jp.Args()...) }),
	})

	matched := registry.MatchAspectsForType(reflect.TypeOf(&TestUserService{}))
	if len(matched) == 0 {
		t.Error("expected matched aspects")
	}
}

func TestAopRegistryWeaveIfNeeded(t *testing.T) {
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

func TestAopRegistryMarkWeaved(t *testing.T) {
	registry := NewAopRegistry()

	if registry.IsWeaved("test") {
		t.Error("should not be marked initially")
	}

	registry.MarkWeaved("test")

	if !registry.IsWeaved("test") {
		t.Error("should be marked after calling MarkWeaved")
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
