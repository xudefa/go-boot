// Package schedule 定时任务调度框架的自动配置测试
package schedule

import (
	"testing"

	"github.com/xudefa/go-boot/constants"
	"github.com/xudefa/go-boot/core"
	"github.com/xudefa/go-boot/environment"
	"github.com/xudefa/go-boot/event"
)

// MockApplicationContext 实现应用上下文接口用于测试
type MockApplicationContext struct {
	container  *MockContainer
	properties map[string]string
}

func (m *MockApplicationContext) Container() core.Container {
	return m.container
}

func (m *MockApplicationContext) Environment() *environment.Environment {
	// Create a real Environment for testing
	env := environment.NewEnvironment()
	for k, v := range m.properties {
		env.AddPropertySource(environment.NewMapPropertySource("test", environment.PriorityNormal, map[string]any{k: v}))
	}
	return env
}

func (m *MockApplicationContext) Get(id string) (interface{}, error) {
	return m.container.Get(id)
}

func (m *MockApplicationContext) Register(beanID string, builders ...core.BuilderOption) error {
	return m.container.Register(beanID, builders...)
}

func (m *MockApplicationContext) EventBus() interface {
	Publish(event event.ApplicationEvent)
} {
	// Return a mock event bus that does nothing
	return &mockEventBus{}
}

// mockEventBus is a mock implementation of the event bus
type mockEventBus struct{}

func (m *mockEventBus) Publish(event event.ApplicationEvent) {
	// Do nothing for testing
}

func TestScheduleAutoConfiguration_Configure(t *testing.T) {
	config := &ScheduleAutoConfiguration{}

	container := &MockContainer{}
	properties := map[string]string{
		constants.SchedulePoolSize:        "5",
		constants.ScheduleScanAnnotations: "false",
	}
	ctx := &MockApplicationContext{
		container:  container,
		properties: properties,
	}

	err := config.Configure(ctx)
	if err != nil {
		t.Fatalf("Configure failed: %v", err)
	}

	// Check if scheduler was registered
	scheduler, err := ctx.Get(constants.ScheduleSchedulerBeanID)
	if err != nil {
		t.Fatalf("Scheduler not registered: %v", err)
	}

	if _, ok := scheduler.(Scheduler); !ok {
		t.Error("Registered object is not a Scheduler")
	}
}

func TestScheduleStarter_Lifecycle(t *testing.T) {
	starter := &ScheduleStarter{}

	container := &MockContainer{}
	properties := map[string]string{
		constants.ScheduleEnabled: "true",
	}
	ctx := &MockApplicationContext{
		container:  container,
		properties: properties,
	}

	// Create and register a scheduler
	scheduler := NewScheduler()
	err := ctx.Register(constants.ScheduleSchedulerBeanID, core.Bean(scheduler))
	if err != nil {
		t.Fatalf("Failed to register scheduler: %v", err)
	}

	// Test configure
	err = starter.Configure(ctx)
	if err != nil {
		t.Errorf("Configure failed: %v", err)
	}

	// Test start
	err = starter.Start(ctx)
	if err != nil {
		t.Errorf("Start failed: %v", err)
	}

	// Test stop
	err = starter.Stop(ctx)
	if err != nil {
		t.Errorf("Stop failed: %v", err)
	}
}

func TestScheduleStarter_Dependencies(t *testing.T) {
	starter := &ScheduleStarter{}
	deps := starter.Dependencies()

	if len(deps) > 0 {
		t.Errorf("Expected no dependencies, got %v", deps)
	}
}

func TestScheduleStarter_Name(t *testing.T) {
	starter := &ScheduleStarter{}
	name := starter.Name()

	if name != "ScheduleStarter" {
		t.Errorf("Expected name 'ScheduleStarter', got '%s'", name)
	}
}

func TestScheduleStarter_GetCondition(t *testing.T) {
	starter := &ScheduleStarter{}
	condition := starter.GetCondition()

	// We can't easily test the condition itself, but we can ensure it's not nil
	if condition == nil {
		t.Error("Expected non-nil condition")
	}
}

func TestResolveScheduler_Found(t *testing.T) {
	container := &MockContainer{}
	properties := map[string]string{}
	ctx := &MockApplicationContext{
		container:  container,
		properties: properties,
	}

	scheduler := NewScheduler()
	err := ctx.Register(constants.ScheduleSchedulerBeanID, core.Bean(scheduler))
	if err != nil {
		t.Fatalf("Failed to register scheduler: %v", err)
	}

	result, ok := resolveScheduler(ctx)
	if !ok {
		t.Error("Expected scheduler to be found")
	}

	if result != scheduler {
		t.Error("Returned scheduler is not the same as registered one")
	}
}

func TestResolveScheduler_NotFound(t *testing.T) {
	container := &MockContainer{}
	properties := map[string]string{}
	ctx := &MockApplicationContext{
		container:  container,
		properties: properties,
	}

	_, ok := resolveScheduler(ctx)
	if ok {
		t.Error("Expected scheduler not to be found")
	}
}
