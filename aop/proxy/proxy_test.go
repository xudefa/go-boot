package proxy

import (
	"testing"
)

type TestTarget struct{}

func (t *TestTarget) DoSomething() string {
	return "done"
}

func TestProxyFactory_CreateProxy(t *testing.T) {
	factory := NewProxyFactory()
	target := &TestTarget{}

	proxy := factory.CreateProxy(target, nil)

	if proxy == nil {
		t.Fatal("expected non-nil proxy")
	}

	// 验证代理可以转换为原始类型
	_, ok := proxy.(*TestTarget)
	if !ok {
		t.Error("expected proxy to be convertible to *TestTarget")
	}
}

func TestProxyFactory_CreateProxy_WithAdvisors(t *testing.T) {
	factory := NewProxyFactory()
	target := &TestTarget{}

	// 创建 mock advisor
	advisors := []Advisor{}

	proxy := factory.CreateProxy(target, advisors)

	if proxy == nil {
		t.Fatal("expected non-nil proxy")
	}
}
