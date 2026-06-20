package core

import (
	"reflect"
	"testing"
)

// mockBean 用于测试的模拟 Bean 结构
type mockBean struct {
	Name string
}

// BeanPostProcessorFunc 用于测试的 Bean 后置处理器函数类型
type BeanPostProcessorFunc func(bean any, beanID string) (any, error)

// ProcessBean 实现 BeanPostProcessor 接口
func (f BeanPostProcessorFunc) ProcessBean(bean any, beanID string) (any, error) {
	return f(bean, beanID)
}

// PostProcess 实现 BeanPostProcessor 接口
func (f BeanPostProcessorFunc) PostProcess(bean any, beanID string) (any, error) {
	return f(bean, beanID)
}

// TestBean 测试 Bean 构建器选项
// 验证：Bean 函数正确设置 Instance 和 ConcreteType
func TestBean(t *testing.T) {
	t.Parallel()
	bean := &mockBean{Name: "test"}
	opt := Bean(bean)

	def := &BeanDefinition{}
	if err := opt(def); err != nil {
		t.Errorf("Bean() error = %v", err)
	}

	if def.Instance != bean {
		t.Error("Bean() should set Instance")
	}

	if def.ConcreteType != reflect.TypeFor[*mockBean]() {
		t.Error("Bean() should set ConcreteType")
	}
}

// TestFactory 测试 Factory 构建器选项
// 验证：Factory 函数正确设置 Factory 和 ConcreteType
func TestFactory(t *testing.T) {
	t.Parallel()
	factoryFn := func(c Container) (any, error) {
		return &mockBean{Name: "factory"}, nil
	}
	tp := reflect.TypeFor[mockBean]()

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

// TestType 测试 Type 构建器选项
// 验证：Type 函数正确设置 ConcreteType
func TestType(t *testing.T) {
	t.Parallel()
	tp := reflect.TypeFor[mockBean]()

	opt := Type(tp)

	def := &BeanDefinition{}
	if err := opt(def); err != nil {
		t.Errorf("Type() error = %v", err)
	}

	if def.ConcreteType != tp {
		t.Error("Type() should set ConcreteType")
	}
}

// TestSetScope 测试 SetScope 构建器选项
// 验证：SetScope 函数正确设置 Scope
func TestSetScope(t *testing.T) {
	t.Parallel()
	opt := SetScope(PrototypeScope)

	def := &BeanDefinition{}
	if err := opt(def); err != nil {
		t.Errorf("SetScope() error = %v", err)
	}

	if def.Scope != PrototypeScope {
		t.Errorf("SetScope() = %v, want %v", def.Scope, PrototypeScope)
	}
}

// TestSingleton 测试 Singleton 构建器选项
// 验证：Singleton 函数正确设置 Scope 为 SingletonScope
func TestSingleton(t *testing.T) {
	t.Parallel()
	opt := Singleton()

	def := &BeanDefinition{}
	if err := opt(def); err != nil {
		t.Errorf("Singleton() error = %v", err)
	}

	if def.Scope != SingletonScope {
		t.Errorf("Singleton() = %v, want %v", def.Scope, SingletonScope)
	}
}

// TestPrototype 测试 Prototype 构建器选项
// 验证：Prototype 函数正确设置 Scope 为 PrototypeScope
func TestPrototype(t *testing.T) {
	t.Parallel()
	opt := Prototype()

	def := &BeanDefinition{}
	if err := opt(def); err != nil {
		t.Errorf("Prototype() error = %v", err)
	}

	if def.Scope != PrototypeScope {
		t.Errorf("Prototype() = %v, want %v", def.Scope, PrototypeScope)
	}
}

// TestFields 测试 Fields 构建器选项
// 验证：Fields 函数正确添加多个字段注入配置
func TestFields(t *testing.T) {
	t.Parallel()
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

// TestField 测试 Field 构建器选项
// 验证：Field 函数正确创建值注入配置（IsRef 为 false）
func TestField(t *testing.T) {
	t.Parallel()
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

// TestBuilderRef 测试 Ref 构建器选项
// 验证：Ref 函数正确创建引用注入配置（IsRef 为 true）
func TestBuilderRef(t *testing.T) {
	t.Parallel()
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

// TestBuilderRef_WithFieldName 测试带字段名的 Ref 构建器选项
// 验证：Ref 函数第二个参数作为字段名，Name 使用该值而非 beanID
func TestBuilderRef_WithFieldName(t *testing.T) {
	t.Parallel()
	opt := Ref("myService", "Service")

	def := &BeanDefinition{}
	if err := opt(def); err != nil {
		t.Errorf("Ref() error = %v", err)
	}

	if def.Fields[0].Name != "Service" {
		t.Errorf("Ref() name = %v, want Service", def.Fields[0].Name)
	}
}

// TestBuilderDependsOn 测试 DependsOn 构建器选项
// 验证：DependsOn 函数正确设置依赖列表
func TestBuilderDependsOn(t *testing.T) {
	t.Parallel()
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

// TestInit 测试 Init 构建器选项
// 验证：Init 函数正确设置初始化函数，并在调用时正确执行
func TestInit(t *testing.T) {
	t.Parallel()
	initialized := false
	initFn := func(i any) error {
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
		t.Fatalf("Init() error = %v", err)
	}
	if !initialized {
		t.Error("Init function should have been called")
	}
}

// TestBuilderCondition 测试 Condition 构建器选项
// 验证：Condition 函数正确设置条件函数
func TestBuilderCondition(t *testing.T) {
	t.Parallel()
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

// TestBuilderPostProcessor 测试 PostProcessor 构建器选项
// 验证：PostProcessor 函数正确添加后置处理器
func TestBuilderPostProcessor(t *testing.T) {
	t.Parallel()
	processor := BeanPostProcessorFunc(func(bean any, beanID string) (any, error) {
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

// TestMultipleOptions 测试多个构建器选项组合使用
// 验证：可以同时使用 Bean、Singleton、Field 等多个选项
func TestMultipleOptions(t *testing.T) {
	t.Parallel()
	container := New()

	err := container.Register("test", Bean(&mockBean{}), Singleton(), Field("Name", "test"))
	if err != nil {
		t.Errorf("Register() with multiple options error = %v", err)
	}

	bean, err := container.Get("test")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if bean.(*mockBean).Name != "test" {
		t.Errorf("Bean() = %v, want test", bean.(*mockBean).Name)
	}
}
