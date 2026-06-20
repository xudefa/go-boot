package boot

import (
	"testing"

	"github.com/xudefa/go-boot/condition"
)

type mockAutoConfig struct {
	configured bool
}

func (m *mockAutoConfig) Configure(ctx ApplicationContext) error {
	m.configured = true
	return nil
}

type mockConditionContext struct{}

func (m *mockConditionContext) Environment() interface{ GetProperty(key string) (any, bool) } {
	return envGetter{func(key string) (any, bool) { return nil, false }}
}

func (m *mockConditionContext) Container() interface{ Has(id string) bool } {
	return containerChecker{func(id string) bool { return false }}
}

func (m *mockConditionContext) ClassLoader() interface{ HasClass(name string) bool } {
	return classChecker{func(name string) bool { return false }}
}

func (m *mockConditionContext) GetBean(beanID string) (any, bool) {
	return nil, false
}

func (m *mockConditionContext) HasProperty(key string) bool {
	return false
}

func (m *mockConditionContext) GetProperty(key string) (any, bool) {
	return nil, false
}

type envGetter struct{ fn func(string) (any, bool) }

func (e envGetter) GetProperty(key string) (any, bool) { return e.fn(key) }

type containerChecker struct{ fn func(string) bool }

func (c containerChecker) Has(id string) bool { return c.fn(id) }

type classChecker struct{ fn func(string) bool }

func (c classChecker) HasClass(name string) bool { return c.fn(name) }

func TestAutoConfigRegistry_Register(t *testing.T) {
	t.Parallel()
	registry := NewAutoConfigRegistry()
	cfg := &mockAutoConfig{}

	registry.Add(AutoConfigEntry{
		Config:     cfg,
		Conditions: nil,
		Order:      0,
	})

	if len(registry.GetAll()) != 1 {
		t.Fatal("expected 1 auto config to be registered")
	}
}

func TestAutoConfigRegistry_Matching(t *testing.T) {
	t.Parallel()
	registry := NewAutoConfigRegistry()

	alwaysMatch := &mockAutoConfig{}
	registry.Add(AutoConfigEntry{
		Config:     alwaysMatch,
		Conditions: nil,
		Order:      0,
	})

	ctx := &mockConditionContext{}
	matched := registry.GetMatching(ctx)
	if len(matched) != 1 {
		t.Fatalf("expected 1 matching config, got %d", len(matched))
	}
}

func TestAutoConfigRegistry_ConditionFiltering(t *testing.T) {
	t.Parallel()
	registry := NewAutoConfigRegistry()

	registry.Add(AutoConfigEntry{
		Config:     &mockAutoConfig{},
		Conditions: []condition.Condition{condition.OnProperty("should.match")},
		Order:      0,
	})

	ctx := &mockConditionContext{}
	matched := registry.GetMatching(ctx)
	if len(matched) != 0 {
		t.Fatal("expected 0 matching configs when condition fails")
	}
}

func TestGlobalRegistry(t *testing.T) {
	cfg := &mockAutoConfig{}
	RegisterAutoConfig(cfg)

	if GlobalRegistry() == nil {
		t.Fatal("expected global registry to exist")
	}
}
