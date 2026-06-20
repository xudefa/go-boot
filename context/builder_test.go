package context

import (
	"testing"

	"github.com/xudefa/go-boot/core"
	"github.com/xudefa/go-boot/environment"
	"github.com/xudefa/go-boot/event"
	"github.com/xudefa/go-boot/life"
	"github.com/xudefa/go-boot/refresh"
)

func TestNewApplicationContextBuilder(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	if builder == nil {
		t.Fatal("builder should not be nil")
	}
	if builder.eventListeners == nil {
		t.Error("eventListeners should be initialized")
	}
	if builder.beans == nil {
		t.Error("beans should be initialized")
	}
}

func TestBuilder_ChainMethods(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	container := core.New()
	env := environment.NewEnvironment()
	lifecycle := life.NewLifecycleManager()
	eventBus := event.NewEventBus()
	refreshMgr := refresh.NewRefreshScopeManager(container, nil)

	result := builder.
		Container(container).
		Environment(env).
		Lifecycle(lifecycle).
		EventBus(eventBus).
		RefreshScopeManager(refreshMgr)

	if result != builder {
		t.Error("chain methods should return the same builder")
	}
	if builder.container != container {
		t.Error("container not set correctly")
	}
	if builder.env != env {
		t.Error("environment not set correctly")
	}
	if builder.lifecycle != lifecycle {
		t.Error("lifecycle not set correctly")
	}
	if builder.eventBus != eventBus {
		t.Error("eventBus not set correctly")
	}
	if builder.refreshScopeMgr != refreshMgr {
		t.Error("refreshScopeMgr not set correctly")
	}
}

func TestBuilder_WithRefreshOption(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	opt1 := func(o *refresh.RefreshConfig) {}
	opt2 := func(o *refresh.RefreshConfig) {}

	result := builder.WithRefreshOption(opt1, opt2)

	if result != builder {
		t.Error("WithRefreshOption should return the same builder")
	}
	if len(builder.refreshOpts) != 2 {
		t.Errorf("expected 2 refresh options, got %d", len(builder.refreshOpts))
	}
}

func TestBuilder_WithPhaseListener(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	listener := &testPhaseListener{}
	result := builder.WithPhaseListener(listener)

	if result != builder {
		t.Error("WithPhaseListener should return the same builder")
	}
	if len(builder.phaseListeners) != 1 {
		t.Errorf("expected 1 phase listener, got %d", len(builder.phaseListeners))
	}
}

func TestBuilder_WithEventListener(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	builder.WithEventListener("testEvent", func(e event.ApplicationEvent) {})

	if len(builder.eventListeners["testEvent"]) != 1 {
		t.Errorf("expected 1 event listener, got %d", len(builder.eventListeners["testEvent"]))
	}
}

func TestBuilder_OnApplicationStarted(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	listener := func(e event.ApplicationEvent) {}
	result := builder.OnApplicationStarted(listener)

	if result != builder {
		t.Error("OnApplicationStarted should return the same builder")
	}
	if len(builder.eventListeners[event.EventApplicationStarted]) != 1 {
		t.Errorf("expected 1 listener for ApplicationStarted, got %d", len(builder.eventListeners[event.EventApplicationStarted]))
	}
}

func TestBuilder_OnApplicationReady(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	listener := func(e event.ApplicationEvent) {}
	result := builder.OnApplicationReady(listener)

	if result != builder {
		t.Error("OnApplicationReady should return the same builder")
	}
	if len(builder.eventListeners[event.EventApplicationReady]) != 1 {
		t.Errorf("expected 1 listener for ApplicationReady, got %d", len(builder.eventListeners[event.EventApplicationReady]))
	}
}

func TestBuilder_OnApplicationStopped(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	listener := func(e event.ApplicationEvent) {}
	result := builder.OnApplicationStopped(listener)

	if result != builder {
		t.Error("OnApplicationStopped should return the same builder")
	}
	if len(builder.eventListeners[event.EventApplicationStopped]) != 1 {
		t.Errorf("expected 1 listener for ApplicationStopped, got %d", len(builder.eventListeners[event.EventApplicationStopped]))
	}
}

func TestBuilder_Bean(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	result := builder.Bean("testBean", core.Singleton())

	if result != builder {
		t.Error("Bean should return the same builder")
	}
	if len(builder.beans["testBean"]) != 1 {
		t.Errorf("expected 1 bean option, got %d", len(builder.beans["testBean"]))
	}
}

func TestBuilder_Bean_MultipleOptions(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	builder.Bean("testBean", core.Singleton(), core.DependsOn("dep1", "dep2"))

	if len(builder.beans["testBean"]) != 2 {
		t.Errorf("expected 2 bean options, got %d", len(builder.beans["testBean"]))
	}
}

func TestBuilder_Build_WithDefaults(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	ctx, err := builder.Build()

	if err != nil {
		t.Fatalf("Build should succeed with defaults: %v", err)
	}
	if ctx == nil {
		t.Fatal("context should not be nil")
	}
	if ctx.Container() == nil {
		t.Error("container should be created")
	}
	if ctx.Environment() == nil {
		t.Error("environment should be created")
	}
	if ctx.EventBus() == nil {
		t.Error("eventBus should be created")
	}
	if ctx.Lifecycle() == nil {
		t.Error("lifecycle should be created")
	}
}

func TestBuilder_Build_WithCustomComponents(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	container := core.New()
	env := environment.NewEnvironment()
	lifecycle := life.NewLifecycleManager()
	eventBus := event.NewEventBus()
	refreshMgr := refresh.NewRefreshScopeManager(container, nil)

	builder.
		Container(container).
		Environment(env).
		Lifecycle(lifecycle).
		EventBus(eventBus).
		RefreshScopeManager(refreshMgr)

	ctx, err := builder.Build()

	if err != nil {
		t.Fatalf("Build should succeed: %v", err)
	}
	if ctx.Container() != container {
		t.Error("custom container not used")
	}
	if ctx.Environment() != env {
		t.Error("custom environment not used")
	}
	if ctx.Lifecycle() != lifecycle {
		t.Error("custom lifecycle not used")
	}
	if ctx.EventBus() != eventBus {
		t.Error("custom eventBus not used")
	}
	if ctx.RefreshScopeManager() != refreshMgr {
		t.Error("custom refreshScopeMgr not used")
	}
}

func TestBuilder_Build_WithPhaseListeners(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	lifecycle := life.NewLifecycleManager()
	listener := &testPhaseListener{}

	builder.
		Lifecycle(lifecycle).
		WithPhaseListener(listener)

	ctx, err := builder.Build()

	if err != nil {
		t.Fatalf("Build should succeed: %v", err)
	}

	if ctx.Lifecycle() == nil {
		t.Fatal("lifecycle should be set")
	}
}

func TestBuilder_Build_WithEventListeners(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	var startedCalled, readyCalled, stoppedCalled bool

	builder.
		OnApplicationStarted(func(e event.ApplicationEvent) {
			startedCalled = true
		}).
		OnApplicationReady(func(e event.ApplicationEvent) {
			readyCalled = true
		}).
		OnApplicationStopped(func(e event.ApplicationEvent) {
			stoppedCalled = true
		})

	ctx, err := builder.Build()

	if err != nil {
		t.Fatalf("Build should succeed: %v", err)
	}

	if ctx.EventBus() == nil {
		t.Fatal("eventBus should be set")
	}

	ctx.EventBus().Publish(&event.BaseEvent{EventType: event.EventApplicationStarted})
	if !startedCalled {
		t.Error("ApplicationStarted listener should be called")
	}

	ctx.EventBus().Publish(&event.BaseEvent{EventType: event.EventApplicationReady})
	if !readyCalled {
		t.Error("ApplicationReady listener should be called")
	}

	ctx.EventBus().Publish(&event.BaseEvent{EventType: event.EventApplicationStopped})
	if !stoppedCalled {
		t.Error("ApplicationStopped listener should be called")
	}
}

func TestBuilder_Build_WithBeanRegistration(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	type TestService struct {
		Name string
	}

	service := &TestService{Name: "test"}
	builder.Bean("testService", core.Bean(service))

	ctx, err := builder.Build()

	if err != nil {
		t.Fatalf("Build should succeed: %v", err)
	}

	bean, err := ctx.Get("testService")
	if err != nil {
		t.Fatalf("Get bean should succeed: %v", err)
	}

	ts, ok := bean.(*TestService)
	if !ok {
		t.Fatal("bean should be *TestService")
	}
	if ts.Name != "test" {
		t.Errorf("expected Name='test', got %s", ts.Name)
	}
}

func TestBuilder_Build_BeanRegistrationError(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	container := core.New()
	if err := container.Register("existingBean", core.Bean(&struct{}{})); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	builder.Container(container)
	builder.Bean("existingBean", core.Bean(&struct{}{}))

	_, err := builder.Build()

	if err == nil {
		t.Fatal("Build should fail with duplicate bean")
	}
}

func TestBuilder_MustBuild_Success(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()

	ctx := builder.MustBuild()

	if ctx == nil {
		t.Fatal("MustBuild should return valid context")
	}
}

func TestBuilder_MustBuild_Panic(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustBuild should panic on error")
		}
	}()

	builder := NewApplicationContextBuilder()
	container := core.New()
	if err := container.Register("testBean", core.Bean(&struct{}{})); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	builder.Container(container)
	builder.Bean("testBean", core.Bean(&struct{}{}))

	builder.MustBuild()
}

func TestCreateApplicationContext(t *testing.T) {
	t.Parallel()
	container := core.New()
	env := environment.NewEnvironment()

	ctx, err := CreateApplicationContext(
		WithContainer(container),
		WithEnvironment(env),
	)

	if err != nil {
		t.Fatalf("CreateApplicationContext should succeed: %v", err)
	}
	if ctx.Container() != container {
		t.Error("container not set correctly")
	}
	if ctx.Environment() != env {
		t.Error("environment not set correctly")
	}
}

func TestWithContainer(t *testing.T) {
	t.Parallel()
	container := core.New()
	opt := WithContainer(container)

	builder := NewApplicationContextBuilder()
	opt(builder)

	if builder.container != container {
		t.Error("WithContainer should set container")
	}
}

func TestWithEnvironment(t *testing.T) {
	t.Parallel()
	env := environment.NewEnvironment()
	opt := WithEnvironment(env)

	builder := NewApplicationContextBuilder()
	opt(builder)

	if builder.env != env {
		t.Error("WithEnvironment should set environment")
	}
}

func TestWithBean(t *testing.T) {
	t.Parallel()
	opt := WithBean("testBean", core.Singleton())

	builder := NewApplicationContextBuilder()
	opt(builder)

	if len(builder.beans["testBean"]) != 1 {
		t.Errorf("WithBean should add bean option, got %d", len(builder.beans["testBean"]))
	}
}

func TestWithProfile(t *testing.T) {
	t.Parallel()
	opt := WithProfile("dev")

	builder := NewApplicationContextBuilder()
	opt(builder)

	if builder == nil {
		t.Error("WithProfile should not panic")
	}
}

func TestApplicationContextHelper_GetBean(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	type TestService struct {
		Value string
	}

	if err := ctx.Register("testService", core.Bean(&TestService{Value: "hello"})); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	bean, err := helper.GetBean("testService")
	if err != nil {
		t.Fatalf("GetBean should succeed: %v", err)
	}

	ts, ok := bean.(*TestService)
	if !ok {
		t.Fatal("bean should be *TestService")
	}
	if ts.Value != "hello" {
		t.Errorf("expected Value='hello', got %s", ts.Value)
	}
}

func TestApplicationContextHelper_GetBeanOrDefault(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	type TestService struct {
		Value string
	}

	defaultVal := &TestService{Value: "default"}

	if err := ctx.Register("testService", core.Bean(&TestService{Value: "actual"})); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	bean := helper.GetBeanOrDefault("testService", defaultVal)
	ts, ok := bean.(*TestService)
	if !ok {
		t.Fatal("bean should be *TestService")
	}
	if ts.Value != "actual" {
		t.Errorf("expected Value='actual', got %s", ts.Value)
	}

	bean = helper.GetBeanOrDefault("nonExistent", defaultVal)
	ts, ok = bean.(*TestService)
	if !ok {
		t.Fatal("default bean should be *TestService")
	}
	if ts.Value != "default" {
		t.Errorf("expected Value='default', got %s", ts.Value)
	}
}

func TestApplicationContextHelper_HasBean(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	if err := ctx.Register("testBean", core.Bean(&struct{}{})); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if !helper.HasBean("testBean") {
		t.Error("HasBean should return true for existing bean")
	}
	if helper.HasBean("nonExistent") {
		t.Error("HasBean should return false for non-existent bean")
	}
}

func TestApplicationContextHelper_GetProperty(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	ctx.Environment().AddPropertySource(environment.NewMapPropertySource("test", environment.PriorityHighest, map[string]any{
		"app.name": "test-app",
	}))

	value := helper.GetProperty("app.name", "default")
	if value != "test-app" {
		t.Errorf("expected 'test-app', got %s", value)
	}

	value = helper.GetProperty("non.existent", "default")
	if value != "default" {
		t.Errorf("expected 'default', got %s", value)
	}
}

func TestApplicationContextHelper_GetIntProperty(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	ctx.Environment().AddPropertySource(environment.NewMapPropertySource("test", environment.PriorityHighest, map[string]any{
		"app.port": "8080",
	}))

	value := helper.GetIntProperty("app.port", 3000)
	if value != 8080 {
		t.Errorf("expected 8080, got %d", value)
	}

	value = helper.GetIntProperty("non.existent", 3000)
	if value != 3000 {
		t.Errorf("expected 3000, got %d", value)
	}
}

func TestApplicationContextHelper_GetBoolProperty(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	ctx.Environment().AddPropertySource(environment.NewMapPropertySource("test", environment.PriorityHighest, map[string]any{
		"app.debug": "true",
	}))

	value := helper.GetBoolProperty("app.debug", false)
	if !value {
		t.Error("expected true, got false")
	}

	value = helper.GetBoolProperty("non.existent", true)
	if !value {
		t.Error("expected true, got false")
	}
}

func TestApplicationContextHelper_IsRunning(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	if helper.IsRunning() {
		t.Error("should not be running initially")
	}

	if err := ctx.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if !helper.IsRunning() {
		t.Error("should be running after start")
	}
}

func TestApplicationContextHelper_GetPhase(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	phase := helper.GetPhase()
	if phase != life.PhaseInitializing {
		t.Errorf("expected PhaseInitializing, got %v", phase)
	}
}

func TestApplicationContextHelper_GetActiveProfiles(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	profiles := helper.GetActiveProfiles()
	if profiles == nil {
		t.Error("GetActiveProfiles should not return nil")
	}
}

func TestApplicationContextHelper_IsDev(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	if helper.IsDev() {
		t.Error("should not be dev by default")
	}
}

func TestApplicationContextHelper_IsProd(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	if helper.IsProd() {
		t.Error("should not be prod by default")
	}
}

func TestApplicationContextHelper_PublishEvent(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	var called bool
	ctx.EventBus().Subscribe("customEvent", func(e event.ApplicationEvent) {
		called = true
	})

	helper.PublishEvent("customEvent")

	if !called {
		t.Error("event listener should be called")
	}
}

func TestApplicationContextHelper_PublishStarted(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	var called bool
	ctx.EventBus().Subscribe(event.EventApplicationStarted, func(e event.ApplicationEvent) {
		called = true
	})

	helper.PublishStarted()

	if !called {
		t.Error("ApplicationStarted listener should be called")
	}
}

func TestApplicationContextHelper_PublishReady(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	var called bool
	ctx.EventBus().Subscribe(event.EventApplicationReady, func(e event.ApplicationEvent) {
		called = true
	})

	helper.PublishReady()

	if !called {
		t.Error("ApplicationReady listener should be called")
	}
}

func TestApplicationContextHelper_PublishStopped(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	helper := NewApplicationContextHelper(ctx)

	var called bool
	ctx.EventBus().Subscribe(event.EventApplicationStopped, func(e event.ApplicationEvent) {
		called = true
	})

	helper.PublishStopped()

	if !called {
		t.Error("ApplicationStopped listener should be called")
	}
}

func TestApplicationRunner_Run(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	runner := NewApplicationRunner(ctx)

	if runner.Context() != ctx {
		t.Error("Context should return the same context")
	}
}

func TestApplicationRunner_Stop(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	runner := NewApplicationRunner(ctx)

	if err := runner.Stop(); err != nil {
		t.Fatalf("Stop should succeed: %v", err)
	}
}

func TestApplicationRunner_Stop_NotStarted(t *testing.T) {
	t.Parallel()
	ctx := NewApplicationContextBuilder().MustBuild()
	runner := NewApplicationRunner(ctx)

	err := runner.Stop()

	if err != nil {
		t.Logf("Stop returned error (expected behavior depends on implementation): %v", err)
	}
}

type testPhaseListener struct {
	onPhase func(old, new life.ApplicationPhase) error
}

func (l *testPhaseListener) OnPhaseChange(old, new life.ApplicationPhase) error {
	if l.onPhase != nil {
		return l.onPhase(old, new)
	}
	return nil
}
