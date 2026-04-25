package core

import (
	"reflect"
	"testing"
)

func TestBean(t *testing.T) {
	bean := &mockBean{Name: "test"}
	opt := Bean(bean)

	def := &BeanDefinition{}
	if err := opt(def); err != nil {
		t.Errorf("Bean() error = %v", err)
	}

	if def.Instance != bean {
		t.Error("Bean() should set Instance")
	}

	if def.ConcreteType != reflect.TypeOf(bean) {
		t.Error("Bean() should set ConcreteType")
	}
}

func TestFactory(t *testing.T) {
	factoryFn := func(c Container) (interface{}, error) {
		return &mockBean{Name: "factory"}, nil
	}
	tp := reflect.TypeOf((*mockBean)(nil)).Elem()

	opt := Factory(factoryFn, tp)

	def := &BeanDefinition{}
	if err := opt(def); err != nil {
		t.Errorf("Factory() error = %v", err)
	}

	if def.Factory == nil {
		t.Error("Factory() should set Factory")
	}

	if def.ConcreteType != tp {
		t.Error("Factory() should set ConcreteType")
	}
}

func TestType(t *testing.T) {
	tp := reflect.TypeOf((*mockBean)(nil)).Elem()

	opt := Type(tp)

	def := &BeanDefinition{}
	if err := opt(def); err != nil {
		t.Errorf("Type() error = %v", err)
	}

	if def.ConcreteType != tp {
		t.Error("Type() should set ConcreteType")
	}
}

func TestSetScope(t *testing.T) {
	opt := SetScope(PrototypeScope)

	def := &BeanDefinition{}
	if err := opt(def); err != nil {
		t.Errorf("SetScope() error = %v", err)
	}

	if def.Scope != PrototypeScope {
		t.Errorf("SetScope() = %v, want %v", def.Scope, PrototypeScope)
	}
}

func TestSingleton(t *testing.T) {
	opt := Singleton()

	def := &BeanDefinition{}
	if err := opt(def); err != nil {
		t.Errorf("Singleton() error = %v", err)
	}

	if def.Scope != SingletonScope {
		t.Errorf("Singleton() = %v, want %v", def.Scope, SingletonScope)
	}
}

func TestPrototype(t *testing.T) {
	opt := Prototype()

	def := &BeanDefinition{}
	if err := opt(def); err != nil {
		t.Errorf("Prototype() error = %v", err)
	}

	if def.Scope != PrototypeScope {
		t.Errorf("Prototype() = %v, want %v", def.Scope, PrototypeScope)
	}
}

func TestFields(t *testing.T) {
	fields := []FieldInjection{
		{Name: "Name", Value: "test", IsRef: false},
		{Name: "Service", Value: "svcRef", IsRef: true},
	}

	opt := Fields(fields...)

	def := &BeanDefinition{}
	if err := opt(def); err != nil {
		t.Errorf("Fields() error = %v", err)
	}

	if len(def.Fields) != 2 {
		t.Errorf("Fields() length = %v, want 2", len(def.Fields))
	}
}

func TestField(t *testing.T) {
	opt := Field("Name", "test")

	def := &BeanDefinition{}
	if err := opt(def); err != nil {
		t.Errorf("Field() error = %v", err)
	}

	if len(def.Fields) != 1 {
		t.Errorf("Field() length = %v, want 1", len(def.Fields))
	}

	if def.Fields[0].Name != "Name" {
		t.Errorf("Field() name = %v, want Name", def.Fields[0].Name)
	}

	if def.Fields[0].Value != "test" {
		t.Errorf("Field() value = %v, want test", def.Fields[0].Value)
	}

	if def.Fields[0].IsRef != false {
		t.Error("Field() IsRef should be false")
	}
}

func TestBuilderRef(t *testing.T) {
	opt := Ref("myService")

	def := &BeanDefinition{}
	if err := opt(def); err != nil {
		t.Errorf("Ref() error = %v", err)
	}

	if len(def.Fields) != 1 {
		t.Errorf("Ref() length = %v, want 1", len(def.Fields))
	}

	if def.Fields[0].Name != "myService" {
		t.Errorf("Ref() name = %v, want myService", def.Fields[0].Name)
	}

	if def.Fields[0].Value != "myService" {
		t.Errorf("Ref() value = %v, want myService", def.Fields[0].Value)
	}

	if !def.Fields[0].IsRef {
		t.Error("Ref() IsRef should be true")
	}
}

func TestBuilderRef_WithFieldName(t *testing.T) {
	opt := Ref("myService", "Service")

	def := &BeanDefinition{}
	if err := opt(def); err != nil {
		t.Errorf("Ref() error = %v", err)
	}

	if def.Fields[0].Name != "Service" {
		t.Errorf("Ref() name = %v, want Service", def.Fields[0].Name)
	}
}

func TestBuilderDependsOn(t *testing.T) {
	opt := DependsOn("db", "logger")

	def := &BeanDefinition{}
	if err := opt(def); err != nil {
		t.Errorf("DependsOn() error = %v", err)
	}

	if len(def.DependsOn) != 2 {
		t.Errorf("DependsOn() length = %v, want 2", len(def.DependsOn))
	}

	if def.DependsOn[0] != "db" {
		t.Errorf("DependsOn()[0] = %v, want db", def.DependsOn[0])
	}

	if def.DependsOn[1] != "logger" {
		t.Errorf("DependsOn()[1] = %v, want logger", def.DependsOn[1])
	}
}

func TestInit(t *testing.T) {
	initialized := false
	initFn := func(i interface{}) error {
		initialized = true
		return nil
	}

	opt := Init(initFn)

	def := &BeanDefinition{}
	if err := opt(def); err != nil {
		t.Errorf("Init() error = %v", err)
	}

	if def.Init == nil {
		t.Error("Init() should set Init")
	}

	err := def.Init(&mockBean{})
	if err != nil {
		return
	}
	if !initialized {
		t.Error("Init function should have been called")
	}
}

func TestBuilderCondition(t *testing.T) {
	conditionFn := func(c Container) bool {
		return true
	}

	opt := Condition(conditionFn)

	def := &BeanDefinition{}
	if err := opt(def); err != nil {
		t.Errorf("Condition() error = %v", err)
	}

	if def.Condition == nil {
		t.Error("Condition() should set Condition")
	}
}

func TestBuilderPostProcessor(t *testing.T) {
	processor := BeanPostProcessorFunc(func(bean interface{}, beanID string) (interface{}, error) {
		return bean, nil
	})

	opt := PostProcessor(processor)

	def := &BeanDefinition{}
	if err := opt(def); err != nil {
		t.Errorf("PostProcessor() error = %v", err)
	}

	if len(def.PostProcessors) != 1 {
		t.Errorf("PostProcessor() length = %v, want 1", len(def.PostProcessors))
	}
}

func TestMultipleOptions(t *testing.T) {
	container := New()

	err := container.Register("test", Bean(&mockBean{}), Singleton(), Field("Name", "test"))
	if err != nil {
		t.Errorf("Register() with multiple options error = %v", err)
	}

	bean, _ := container.Get("test")
	if bean.(*mockBean).Name != "test" {
		t.Errorf("Bean() = %v, want test", bean.(*mockBean).Name)
	}
}
