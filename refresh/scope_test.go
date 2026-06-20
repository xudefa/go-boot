package refresh

import (
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/xudefa/go-boot/event"
)

type mockBeanCreator struct{}

func (m *mockBeanCreator) CreateBean(beanID string) (any, error) {
	return nil, errors.New("mock bean creator")
}

func TestConfigChangeEvent(t *testing.T) {
	event := ConfigChangeEvent{
		EventType: "modify",
		Keys:      []string{"server.port", "server.host"},
		OldValues: map[string]any{"server.port": 8080},
		NewValues: map[string]any{"server.port": 9090},
		Source:    "viper",
		timestamp: time.Now(),
	}

	if event.EventType != "modify" {
		t.Errorf("expected EventType 'modify', got '%s'", event.EventType)
	}

	if len(event.Keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(event.Keys))
	}
}

func TestBeanRefreshedEvent(t *testing.T) {
	event := BeanRefreshedEvent{
		BeanID:      "userService",
		OldVersion:  1,
		NewVersion:  2,
		RefreshTime: time.Now(),
		Success:     true,
	}

	if event.BeanID != "userService" {
		t.Errorf("expected BeanID 'userService', got '%s'", event.BeanID)
	}

	if !event.Success {
		t.Error("expected Success to be true")
	}
}

func TestRefreshProxy_GetTarget(t *testing.T) {
	logger := slog.Default()
	creator := &mockBeanCreator{}
	manager := NewRefreshScopeManager(creator, logger)

	originalBean := &struct{ Value int }{Value: 1}
	proxy := NewRefreshProxy("testBean", originalBean, manager, logger)

	// 初始调用
	target1 := proxy.GetTarget()
	if target1 != originalBean {
		t.Error("expected original bean on first call")
	}

	// 标记需要刷新
	proxy.MarkForRefresh()

	// 再次调用，应该返回旧实例（因为 createBean 返回错误）
	target2 := proxy.GetTarget()
	if target2 != originalBean {
		t.Error("expected original bean when refresh fails")
	}
}

func TestRefreshProxy_MarkForRefresh(t *testing.T) {
	logger := slog.Default()
	creator := &mockBeanCreator{}
	manager := NewRefreshScopeManager(creator, logger)

	proxy := NewRefreshProxy("testBean", &struct{}{}, manager, logger)

	// 初始状态不需要刷新
	if proxy.needsRefresh.Load() {
		t.Error("expected needsRefresh to be false initially")
	}

	// 标记需要刷新
	proxy.MarkForRefresh()

	// 验证标记已设置
	if !proxy.needsRefresh.Load() {
		t.Error("expected needsRefresh to be true after marking")
	}
}

func TestRefreshScopeManager_MarkBeanForRefresh(t *testing.T) {
	logger := slog.Default()
	creator := &mockBeanCreator{}
	manager := NewRefreshScopeManager(creator, logger)

	manager.MarkBeanForRefresh("testBean")

	manager.mu.RLock()
	flagged := manager.refreshFlags["testBean"]
	manager.mu.RUnlock()

	if !flagged {
		t.Error("expected bean to be marked for refresh")
	}
}

func TestRefreshScopeManager_RegisterRefreshableBean(t *testing.T) {
	logger := slog.Default()
	creator := &mockBeanCreator{}
	manager := NewRefreshScopeManager(creator, logger)

	bean := &mockRefreshableBean{}
	manager.RegisterRefreshableBean("testBean", bean)

	manager.mu.RLock()
	registered := manager.refreshableBeans["testBean"]
	manager.mu.RUnlock()

	if registered != bean {
		t.Error("expected bean to be registered")
	}
}

func TestRefreshScopeManager_IncrementBeanVersion(t *testing.T) {
	logger := slog.Default()
	creator := &mockBeanCreator{}
	manager := NewRefreshScopeManager(creator, logger)

	v1 := manager.incrementBeanVersion("testBean")
	v2 := manager.incrementBeanVersion("testBean")

	if v2 != v1+1 {
		t.Errorf("expected version %d, got %d", v1+1, v2)
	}
}

// mockRefreshableBean 测试用的可刷新 Bean
type mockRefreshableBean struct{}

func (m *mockRefreshableBean) OnConfigChange(event ConfigChangeEvent) error {
	return nil
}

func TestEventRouter_RegisterBean(t *testing.T) {
	eventBus := event.NewEventBus()
	router := NewEventRouter(eventBus)

	router.RegisterBean("userService", []string{"server.port", "server.host"})

	router.mu.RLock()
	keys := router.beanConfigMap["userService"]
	router.mu.RUnlock()

	if len(keys) != 2 {
		t.Errorf("expected 2 config keys, got %d", len(keys))
	}
}

func TestEventRouter_FindAffectedBeans(t *testing.T) {
	eventBus := event.NewEventBus()
	router := NewEventRouter(eventBus)

	router.RegisterBean("userService", []string{"server.port", "server.host"})
	router.RegisterBean("dbService", []string{"db.host", "db.port"})

	affected := router.findAffectedBeans([]string{"server.port"})

	if len(affected) != 1 || affected[0] != "userService" {
		t.Errorf("expected userService to be affected, got %v", affected)
	}
}
