package core

import (
	"errors"
	"testing"
)

// TestContainer_CircularDependencyBoundary 测试循环依赖检测边界情况
func TestContainer_CircularDependencyBoundary(t *testing.T) {
	t.Parallel()
	container := New()

	type ServiceA struct {
		B interface{} `inject:"serviceB"`
	}
	type ServiceB struct {
		C interface{} `inject:"serviceC"`
	}
	type ServiceC struct {
		A interface{} `inject:"serviceA"`
	}

	_ = container.Register("serviceA", Bean(&ServiceA{}))
	_ = container.Register("serviceB", Bean(&ServiceB{}))
	_ = container.Register("serviceC", Bean(&ServiceC{}))

	_, err := container.Get("serviceA")
	if err == nil {
		t.Error("Expected circular dependency error, got nil")
	}
	if !errors.Is(err, ErrCircularDep) {
		t.Errorf("Expected ErrCircularDep, got %v", err)
	}
}

// TestContainer_NilInjection 测试 nil 注入
func TestContainer_NilInjection(t *testing.T) {
	t.Parallel()
	container := New()

	type Handler struct {
		Service *mockBean `inject:"nonExistentBean"`
	}
	var h Handler

	err := container.Inject(&h)
	if err == nil {
		t.Error("Expected error when injecting non-existent bean, got nil")
	}
}

// TestContainer_InvalidScope 测试无效作用域
func TestContainer_InvalidScope(t *testing.T) {
	t.Parallel()
	container := New()

	type TestBean struct{}

	_ = container.Register("test", Bean(&TestBean{}), SetScope("invalid"))

	bean, err := container.Get("test")
	if err != nil {
		t.Errorf("Failed to get bean with invalid scope: %v", err)
	}
	if bean == nil {
		t.Error("Expected bean to be created despite invalid scope")
	}
}

// TestContainer_ConcurrentRegister 测试并发注册
func TestContainer_ConcurrentRegister(t *testing.T) {
	t.Parallel()
	container := New()

	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func(id int) {
			beanName := "bean" + string(rune('0'+id%10))
			_ = container.Register(beanName, Bean(&mockBean{Name: beanName}))
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}

// TestContainer_ParentChainLookup 测试容器链查找
func TestContainer_ParentChainLookup(t *testing.T) {
	t.Parallel()
	parent := New()
	_ = parent.Register("shared", Bean(&mockBean{Name: "shared"}))

	child := New()
	_ = child.Register("private", Bean(&mockBean{Name: "private"}))

	_, err := child.Get("shared")
	if err == nil {
		t.Error("Expected error when getting bean from non-parent container")
	}
	if !errors.Is(err, ErrBeanNotFound) {
		t.Errorf("Expected ErrBeanNotFound, got %v", err)
	}

	private, err := child.Get("private")
	if err != nil {
		t.Errorf("Failed to get private bean: %v", err)
	}
	if private.(*mockBean).Name != "private" {
		t.Error("Expected private bean from child")
	}
}
