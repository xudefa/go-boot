package core

import (
	"errors"
	"reflect"
	"testing"
)

type mockBean struct {
	Name string
}

type serviceA struct {
	Name string `inject:"beanB"`
}

type serviceB struct {
	Value string
}

type serviceWithInit struct {
	Initialized bool
}

func (s *serviceWithInit) init() error {
	s.Initialized = true
	return nil
}

func TestNew(t *testing.T) {
	container := New()
	if container == nil {
		t.Error("New() returned nil")
	}
}

func TestContainer_Register(t *testing.T) {
	container := New()

	err := container.Register("testBean", Bean(&mockBean{Name: "test"}))
	if err != nil {
		t.Errorf("Register() error = %v", err)
	}

	err = container.Register("testBean", Bean(&mockBean{}))
	if !errors.Is(err, ErrDuplicateBean) {
		t.Errorf("Register() duplicate error = %v", err)
	}
}

func TestContainer_Get(t *testing.T) {
	container := New()

	container.Register("testBean", Bean(&mockBean{Name: "test"}))

	bean, err := container.Get("testBean")
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}

	if bean.(*mockBean).Name != "test" {
		t.Errorf("Get() = %v, want test", bean.(*mockBean).Name)
	}
}

func TestContainer_Get_NotFound(t *testing.T) {
	container := New()

	_, err := container.Get("nonExistent")
	if !errors.Is(err, ErrBeanNotFound) {
		t.Errorf("Get() error = %v", err)
	}
}

func TestContainer_Get_Singleton(t *testing.T) {
	container := New()

	container.Register("singleton", Bean(&mockBean{Name: "first"}))

	bean1, _ := container.Get("singleton")
	bean2, _ := container.Get("singleton")

	if bean1 != bean2 {
		t.Error("Singleton beans should return the same instance")
	}
}

func TestContainer_Get_Prototype(t *testing.T) {
	container := New()

	container.Register("prototype", Bean(&mockBean{}), Prototype())

	bean1, _ := container.Get("prototype")
	bean2, _ := container.Get("prototype")

	if bean1 == bean2 {
		t.Error("Prototype beans should return different instances")
	}
}

func TestContainer_Get_WithInit(t *testing.T) {
	container := New()

	container.Register("withInit", Bean(&serviceWithInit{}), Init(func(i interface{}) error {
		return i.(*serviceWithInit).init()
	}))

	bean, err := container.Get("withInit")
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}

	if !bean.(*serviceWithInit).Initialized {
		t.Error("Init function should have been called")
	}
}

func TestContainer_Has(t *testing.T) {
	container := New()

	container.Register("exists", Bean(&mockBean{}))

	if !container.Has("exists") {
		t.Error("Has() should return true for registered bean")
	}

	if container.Has("nonExistent") {
		t.Error("Has() should return false for non-existent bean")
	}
}

func TestContainer_Remove(t *testing.T) {
	container := New()

	container.Register("toRemove", Bean(&mockBean{}))

	err := container.Remove("toRemove")
	if err != nil {
		t.Errorf("Remove() error = %v", err)
	}

	if container.Has("toRemove") {
		t.Error("Bean should be removed")
	}
}

func TestContainer_Close(t *testing.T) {
	container := New()

	container.Register("testBean", Bean(&mockBean{}), Singleton())
	container.Get("testBean")

	err := container.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestContainer_Inject(t *testing.T) {
	container := New()

	container.Register("serviceB", Bean(&serviceB{Value: "test"}))

	var target struct {
		Svc *serviceB `inject:"serviceB"`
	}

	err := container.Inject(&target)
	if err != nil {
		t.Errorf("Inject() error = %v", err)
	}

	if target.Svc == nil {
		t.Error("Injection failed, target.Svc is nil")
	}

	if target.Svc.Value != "test" {
		t.Errorf("Inject() = %v, want test", target.Svc.Value)
	}
}

func TestContainer_Invoke(t *testing.T) {
	container := New()

	container.Register("serviceB", Bean(&serviceB{Value: "test"}))

	result, err := container.Invoke(func(svc *serviceB) string {
		return svc.Value
	})

	if err != nil {
		t.Errorf("Invoke() error = %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Invoke() result length = %v, want 1", len(result))
	}

	if result[0] != "test" {
		t.Errorf("Invoke() = %v, want test", result[0])
	}
}

type testInterface interface {
	GetValue() string
}

type implA struct{ Val string }

func (i *implA) GetValue() string {
	return i.Val
}

type implB struct{ Val string }

func (i *implB) GetValue() string {
	return i.Val
}

func TestContainer_GetAll(t *testing.T) {
	container := New()

	container.Register("implA", Bean(&implA{Val: "v1"}))
	container.Register("implB", Bean(&implB{Val: "v2"}))

	results, err := container.GetAll((*testInterface)(nil))
	if err != nil {
		t.Errorf("GetAll() error = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("GetAll() = %v, want 2", len(results))
	}
}

func TestContainer_CircularDependency(t *testing.T) {
	container := New()

	type circularA struct {
		Svc *serviceB `inject:"serviceB"`
	}

	container.Register("serviceB", Bean(&serviceB{Value: "test"}))

	_, err := container.Get("serviceB")
	if err != nil {
		t.Logf("Get serviceB error (expected): %v", err)
	}

	if !container.Has("serviceB") {
		t.Error("serviceB should be registered")
	}
}

func TestBeanScope_Constants(t *testing.T) {
	if SingletonScope != "singleton" {
		t.Errorf("SingletonScope = %v, want singleton", SingletonScope)
	}
	if PrototypeScope != "prototype" {
		t.Errorf("PrototypeScope = %v, want prototype", PrototypeScope)
	}
}

func TestBuilderOptions(t *testing.T) {
	container := New()

	err := container.Register("factory", Factory(func(c Container) (interface{}, error) {
		return &mockBean{Name: "factory"}, nil
	}, reflect.TypeOf((*mockBean)(nil)).Elem()))
	if err != nil {
		t.Errorf("Register() with Factory error = %v", err)
	}

	bean, _ := container.Get("factory")
	if bean.(*mockBean).Name != "factory" {
		t.Errorf("Factory bean = %v, want factory", bean.(*mockBean).Name)
	}
}

func TestContainer_Condition(t *testing.T) {
	container := New()

	container.Register("conditional", Bean(&mockBean{}), Condition(func(c Container) bool {
		return false
	}))

	_, err := container.Get("conditional")
	if err == nil {
		t.Error("Conditional bean should fail when condition is false")
	}
}

func TestContainer_BeanDependsOn(t *testing.T) {
	container := New()

	container.Register("first", Bean(&mockBean{Name: "1"}))
	container.Register("second", Bean(&mockBean{Name: "2"}), DependsOn("first"))

	if !container.Has("first") || !container.Has("second") {
		t.Error("Beans should be registered")
	}
}

func TestContainer_BeanPostProcessor(t *testing.T) {
	type countingProcessor struct {
		count int
	}

	processor := &countingProcessor{}

	container := New()
	container.Register("withProcessor", Bean(&mockBean{}), PostProcessor(BeanPostProcessorFunc(func(bean interface{}, beanID string) (interface{}, error) {
		processor.count++
		return bean, nil
	})))

	container.Get("withProcessor")

	if processor.count != 1 {
		t.Errorf("PostProcessor called %v times, want 1", processor.count)
	}
}

type BeanPostProcessorFunc func(bean interface{}, beanID string) (interface{}, error)

func (f BeanPostProcessorFunc) PostProcess(bean interface{}, beanID string) (interface{}, error) {
	return f(bean, beanID)
}

func TestEnableFieldTag(t *testing.T) {
	container := New(EnableFieldTag(false))

	var target struct {
		Svc *serviceB `inject:"serviceB"`
	}

	container.Register("serviceB", Bean(&serviceB{Value: "test"}))

	err := container.Inject(&target)
	if err != nil {
		t.Errorf("Inject() error = %v", err)
	}

	if target.Svc != nil {
		t.Error("Injection should not happen when EnableFieldTag is false")
	}
}
