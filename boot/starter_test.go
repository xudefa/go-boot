package boot

import (
	"testing"

	"github.com/xudefa/go-boot/condition"
)

type mockStarter struct {
	name         string
	dependencies []string
	configured   bool
	started      bool
	stopped      bool
}

func newMockStarter(name string, deps ...string) *mockStarter {
	return &mockStarter{name: name, dependencies: deps}
}

func (m *mockStarter) Name() string                           { return m.name }
func (m *mockStarter) Dependencies() []string                 { return m.dependencies }
func (m *mockStarter) Configure(ctx ApplicationContext) error { m.configured = true; return nil }
func (m *mockStarter) Start(ctx ApplicationContext) error     { m.started = true; return nil }
func (m *mockStarter) Stop(ctx ApplicationContext) error      { m.stopped = true; return nil }
func (m *mockStarter) GetCondition() condition.Condition      { return nil }

func TestStarterRegistry_Add(t *testing.T) {
	t.Parallel()
	registry := NewStarterRegistry()
	s := newMockStarter("test")

	registry.Add(s)
	if len(registry.GetAll()) != 1 {
		t.Fatal("expected 1 starter in registry")
	}
}

func TestStarterRegistry_GetAll(t *testing.T) {
	t.Parallel()
	registry := NewStarterRegistry()
	registry.Add(newMockStarter("s1"))
	registry.Add(newMockStarter("s2"))

	all := registry.GetAll()
	if len(all) != 2 {
		t.Fatalf("expected 2 starters, got %d", len(all))
	}
}

func TestStarterRegistry_GetOrdered(t *testing.T) {
	t.Parallel()
	registry := NewStarterRegistry()
	registry.Add(newMockStarter("s1", "s2"))
	registry.Add(newMockStarter("s2"))

	ordered := registry.GetOrdered()
	if len(ordered) != 2 {
		t.Fatalf("expected 2 ordered starters, got %d", len(ordered))
	}
}

func TestStarterRegistry_Empty(t *testing.T) {
	t.Parallel()
	registry := NewStarterRegistry()

	if len(registry.GetAll()) != 0 {
		t.Error("expected empty registry")
	}
	if len(registry.GetOrdered()) != 0 {
		t.Error("expected empty ordered starters")
	}
}

func TestRegisterStarterGlobal(t *testing.T) {
	orig := globalStarterRegistry
	registry := NewStarterRegistry()
	globalStarterRegistry = registry
	defer func() { globalStarterRegistry = orig }()

	RegisterStarter(newMockStarter("global"))
	if len(registry.GetAll()) != 1 {
		t.Fatal("expected 1 starter registered globally")
	}
}

func TestGlobalStarterRegistry(t *testing.T) {
	registry := GlobalStarterRegistry()
	if registry == nil {
		t.Fatal("expected non-nil global registry")
	}
}
