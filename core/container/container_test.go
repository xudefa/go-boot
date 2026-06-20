package container

import (
	"reflect"
	"testing"
)

type TestService struct {
	Name string
}

func TestContainer_RegisterAndGet(t *testing.T) {
	c := New()

	def := BeanDefinition{
		Type:  reflect.TypeOf((*TestService)(nil)).Elem(),
		Scope: Singleton,
		Factory: func() (any, error) {
			return &TestService{Name: "test"}, nil
		},
	}

	err := c.Register("testService", def)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bean := c.Get("testService")
	svc := bean.(*TestService)

	if svc.Name != "test" {
		t.Errorf("expected name 'test', got %s", svc.Name)
	}
}

func TestContainer_GetT(t *testing.T) {
	c := New()

	def := BeanDefinition{
		Type:  reflect.TypeOf((*TestService)(nil)).Elem(),
		Scope: Singleton,
		Factory: func() (any, error) {
			return &TestService{Name: "typed"}, nil
		},
	}

	if err := c.Register("typedService", def); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	svc, err := GetT[TestService](c, "typedService")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if svc.Name != "typed" {
		t.Errorf("expected name 'typed', got %s", svc.Name)
	}
}

func TestContainer_Contains(t *testing.T) {
	c := New()

	if c.Contains("nonexistent") {
		t.Error("expected Contains to return false for nonexistent bean")
	}

	def := BeanDefinition{
		Type:  reflect.TypeOf((*TestService)(nil)).Elem(),
		Scope: Singleton,
	}
	if err := c.Register("existing", def); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !c.Contains("existing") {
		t.Error("expected Contains to return true for existing bean")
	}
}

func TestContainer_DuplicateRegistration(t *testing.T) {
	c := New()

	def := BeanDefinition{
		Type:  reflect.TypeOf((*TestService)(nil)).Elem(),
		Scope: Singleton,
	}

	var err error
	err = c.Register("service", def)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = c.Register("service", def)

	if err == nil {
		t.Error("expected error for duplicate registration")
	}
}

func TestContainer_SingletonScope(t *testing.T) {
	c := New()

	callCount := 0
	def := BeanDefinition{
		Type:  reflect.TypeOf((*TestService)(nil)).Elem(),
		Scope: Singleton,
		Factory: func() (any, error) {
			callCount++
			return &TestService{Name: "singleton"}, nil
		},
	}

	if err := c.Register("singleton", def); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 获取两次，应该只调用一次 Factory
	c.Get("singleton")
	c.Get("singleton")

	if callCount != 1 {
		t.Errorf("expected factory called once, got %d", callCount)
	}
}

func TestContainer_PrototypeScope(t *testing.T) {
	c := New()

	callCount := 0
	def := BeanDefinition{
		Type:  reflect.TypeOf((*TestService)(nil)).Elem(),
		Scope: Prototype,
		Factory: func() (any, error) {
			callCount++
			return &TestService{Name: "prototype"}, nil
		},
	}

	if err := c.Register("prototype", def); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 获取两次，应该调用两次 Factory
	c.Get("prototype")
	c.Get("prototype")

	if callCount != 2 {
		t.Errorf("expected factory called twice, got %d", callCount)
	}
}

func TestContainer_StartStop(t *testing.T) {
	c := New()

	initCalled := false
	destroyCalled := false

	def := BeanDefinition{
		Type:  reflect.TypeOf((*TestService)(nil)).Elem(),
		Scope: Singleton,
		Factory: func() (any, error) {
			return &TestService{Name: "lifecycle"}, nil
		},
		InitFunc: func(bean any) error {
			initCalled = true
			return nil
		},
		DestroyFunc: func(bean any) {
			destroyCalled = true
		},
	}

	if err := c.Register("lifecycle", def); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := c.Start(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := c.Stop(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !initCalled {
		t.Error("expected init callback to be called")
	}
	if !destroyCalled {
		t.Error("expected destroy callback to be called")
	}
}

func TestContainer_ListDefinitions(t *testing.T) {
	c := New()

	if err := c.Register("service1", BeanDefinition{Type: reflect.TypeOf((*TestService)(nil)).Elem()}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := c.Register("service2", BeanDefinition{Type: reflect.TypeOf((*TestService)(nil)).Elem()}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ids := c.ListDefinitions()

	if len(ids) != 2 {
		t.Errorf("expected 2 definitions, got %d", len(ids))
	}
}
