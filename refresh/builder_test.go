package refresh

import (
	"log/slog"
	"testing"
)

func TestRefreshScopeManagerBuilder_Defaults(t *testing.T) {
	builder := NewRefreshScopeManagerBuilder()

	if builder.refreshableBeans == nil {
		t.Error("expected non-nil refreshableBeans")
	}
}

func TestRefreshScopeManagerBuilder_ChainConfig(t *testing.T) {
	creator := &mockBeanCreator{}

	manager, err := NewRefreshScopeManagerBuilder().
		BeanCreator(creator).
		Enabled(true).
		MaxRefreshAttempts(5).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if manager == nil {
		t.Fatal("expected non-nil manager")
	}
}

func TestRefreshScopeManagerBuilder_MissingBeanCreator(t *testing.T) {
	_, err := NewRefreshScopeManagerBuilder().Build()
	if err == nil {
		t.Error("expected error for missing beanCreator")
	}
}

func TestRefreshScopeManagerBuilder_WithLogger(t *testing.T) {
	creator := &mockBeanCreator{}
	logger := slog.Default()

	manager, err := NewRefreshScopeManagerBuilder().
		BeanCreator(creator).
		Logger(logger).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if manager == nil {
		t.Fatal("expected non-nil manager")
	}
}

func TestRefreshScopeManagerBuilder_WithRefreshableBeans(t *testing.T) {
	creator := &mockBeanCreator{}
	bean := &mockRefreshableBean{}

	manager, err := NewRefreshScopeManagerBuilder().
		BeanCreator(creator).
		RefreshableBean("testBean", bean).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if manager == nil {
		t.Fatal("expected non-nil manager")
	}
}

func TestRefreshScopeManagerBuilder_MustBuild(t *testing.T) {
	creator := &mockBeanCreator{}

	manager := NewRefreshScopeManagerBuilder().
		BeanCreator(creator).
		MustBuild()

	if manager == nil {
		t.Fatal("expected non-nil manager")
	}
}

func TestRefreshScopeManagerBuilder_MustBuild_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	NewRefreshScopeManagerBuilder().MustBuild()
}

func TestRefreshProxyBuilder_Defaults(t *testing.T) {
	builder := NewRefreshProxyBuilder()

	if builder == nil {
		t.Fatal("expected non-nil builder")
	}
}

func TestRefreshProxyBuilder_ChainConfig(t *testing.T) {
	manager := &RefreshScopeManager{}

	proxy, err := NewRefreshProxyBuilder().
		BeanID("testBean").
		Target("testTarget").
		Manager(manager).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if proxy == nil {
		t.Fatal("expected non-nil proxy")
	}
}

func TestRefreshProxyBuilder_MissingBeanID(t *testing.T) {
	manager := &RefreshScopeManager{}

	_, err := NewRefreshProxyBuilder().
		Target("testTarget").
		Manager(manager).
		Build()

	if err == nil {
		t.Error("expected error for missing beanID")
	}
}

func TestRefreshProxyBuilder_MissingTarget(t *testing.T) {
	manager := &RefreshScopeManager{}

	_, err := NewRefreshProxyBuilder().
		BeanID("testBean").
		Manager(manager).
		Build()

	if err == nil {
		t.Error("expected error for missing target")
	}
}

func TestRefreshProxyBuilder_MissingManager(t *testing.T) {
	_, err := NewRefreshProxyBuilder().
		BeanID("testBean").
		Target("testTarget").
		Build()

	if err == nil {
		t.Error("expected error for missing manager")
	}
}

func TestRefreshProxyBuilder_WithLogger(t *testing.T) {
	manager := &RefreshScopeManager{}
	logger := slog.Default()

	proxy, err := NewRefreshProxyBuilder().
		BeanID("testBean").
		Target("testTarget").
		Manager(manager).
		Logger(logger).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if proxy == nil {
		t.Fatal("expected non-nil proxy")
	}
}

func TestRefreshProxyBuilder_MustBuild(t *testing.T) {
	manager := &RefreshScopeManager{}

	proxy := NewRefreshProxyBuilder().
		BeanID("testBean").
		Target("testTarget").
		Manager(manager).
		MustBuild()

	if proxy == nil {
		t.Fatal("expected non-nil proxy")
	}
}

func TestRefreshProxyBuilder_MustBuild_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	NewRefreshProxyBuilder().MustBuild()
}

func TestConfigChangeEventBuilder_BasicEvent(t *testing.T) {
	event := NewConfigChangeEventBuilder().
		EventType("modify").
		Source("nacos").
		Keys([]string{"db.host", "db.port"}).
		OldValues(map[string]any{"db.host": "localhost"}).
		NewValues(map[string]any{"db.host": "192.168.1.100"}).
		Metadata(map[string]string{"env": "production"}).
		Build()

	if event.EventType != "modify" {
		t.Errorf("expected EventType 'modify', got %s", event.EventType)
	}

	if event.Source != "nacos" {
		t.Errorf("expected Source 'nacos', got %s", event.Source)
	}

	if len(event.Keys) != 2 {
		t.Errorf("expected 2 Keys, got %d", len(event.Keys))
	}

	if event.OldValues["db.host"] != "localhost" {
		t.Errorf("expected OldValues['db.host'] 'localhost', got %v", event.OldValues["db.host"])
	}

	if event.NewValues["db.host"] != "192.168.1.100" {
		t.Errorf("expected NewValues['db.host'] '192.168.1.100', got %v", event.NewValues["db.host"])
	}

	if event.Metadata["env"] != "production" {
		t.Errorf("expected Metadata['env'] 'production', got %s", event.Metadata["env"])
	}
}

func TestConfigChangeEventBuilder_EmptyEvent(t *testing.T) {
	event := NewConfigChangeEventBuilder().Build()

	if event.Source != "" {
		t.Errorf("expected empty Source, got %s", event.Source)
	}

	if event.Keys != nil {
		t.Errorf("expected nil Keys, got %v", event.Keys)
	}

	if event.OldValues == nil {
		t.Error("expected non-nil OldValues")
	}

	if event.NewValues == nil {
		t.Error("expected non-nil NewValues")
	}
}

func TestRefreshScopeManagerBuilder_MultipleRefreshableBeans(t *testing.T) {
	creator := &mockBeanCreator{}
	bean1 := &mockRefreshableBean{}
	bean2 := &mockRefreshableBean{}
	bean3 := &mockRefreshableBean{}

	manager, err := NewRefreshScopeManagerBuilder().
		BeanCreator(creator).
		RefreshableBean("bean1", bean1).
		RefreshableBean("bean2", bean2).
		RefreshableBean("bean3", bean3).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if manager == nil {
		t.Fatal("expected non-nil manager")
	}
}

func TestRefreshScopeManagerBuilder_AllOptions(t *testing.T) {
	creator := &mockBeanCreator{}
	logger := slog.Default()

	manager, err := NewRefreshScopeManagerBuilder().
		BeanCreator(creator).
		Logger(logger).
		Enabled(true).
		RefreshDelay(200).
		MaxRefreshAttempts(10).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if manager == nil {
		t.Fatal("expected non-nil manager")
	}
}

func TestRefreshProxyBuilder_DefaultLogger(t *testing.T) {
	manager := &RefreshScopeManager{}

	proxy, err := NewRefreshProxyBuilder().
		BeanID("testBean").
		Target("testTarget").
		Manager(manager).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if proxy == nil {
		t.Fatal("expected non-nil proxy")
	}

	// Logger should default to slog.Default()
	if proxy.logger == nil {
		t.Error("expected non-nil logger")
	}
}
