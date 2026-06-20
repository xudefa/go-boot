package boot

import (
	"testing"

	"github.com/xudefa/go-boot/condition"
	"github.com/xudefa/go-boot/life"
)

func TestBoot_Stop_DuringPhaseContextRefreshed(t *testing.T) {
	t.Parallel()

	s := newMockStarter("test")
	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	// 手动设置 starter 绕过 Start() 完整流程
	boot.starters = []Starter{s}

	// 模拟 PhaseInitializing → PhaseConfiguring → PhaseContextRefreshed
	if err := boot.ctx.Lifecycle().SetPhase(life.PhaseConfiguring); err != nil {
		t.Fatalf("SetPhase(Configuring) error = %v", err)
	}
	if err := boot.ctx.Lifecycle().SetPhase(life.PhaseContextRefreshed); err != nil {
		t.Fatalf("SetPhase(ContextRefreshed) error = %v", err)
	}

	// 在 PhaseContextRefreshed 时调用 Stop()
	if err := boot.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	// BUG 1 的验证：Stop() 不应调用 starter.Stop()
	// 因为 starter 只执行了 Configure(), 未执行 Start()
	if s.stopped {
		t.Error("BUG: starter.Stop() was called but starter was never started (phase was ContextRefreshed)")
	}
	if boot.ctx.Lifecycle().GetPhase() != life.PhaseStopped {
		t.Errorf("expected PhaseStopped, got %v", boot.ctx.Lifecycle().GetPhase())
	}
}

func TestBoot_Stop_DuringPhaseConfiguring(t *testing.T) {
	t.Parallel()

	s := newMockStarter("test")
	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	boot.starters = []Starter{s}

	if err := boot.ctx.Lifecycle().SetPhase(life.PhaseConfiguring); err != nil {
		t.Fatalf("SetPhase(Configuring) error = %v", err)
	}

	if err := boot.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	// PhaseConfiguring 时 starter 尚未 Configure，不应调用 Stop
	if s.stopped {
		t.Error("starter.Stop() should not be called when phase was Configuring")
	}
}

func TestBoot_Stop_DoubleCall(t *testing.T) {
	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	// 正常启动
	if err := boot.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if !boot.IsRunning() {
		t.Fatal("expected running after Start()")
	}

	// 第一次 Stop
	if err := boot.Stop(); err != nil {
		t.Fatalf("first Stop() error = %v", err)
	}
	if boot.IsRunning() {
		t.Fatal("should not be running after Stop()")
	}

	// 第二次 Stop — 应安全返回 nil
	if err := boot.Stop(); err != nil {
		t.Fatalf("second Stop() should return nil, got: %v", err)
	}
}

func TestBoot_Stop_PhaseStoppingGuard(t *testing.T) {
	t.Parallel()

	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	if err := boot.ctx.Lifecycle().SetPhase(life.PhaseRunning); err != nil {
		t.Fatalf("SetPhase(Running) error = %v", err)
	}
	if err := boot.ctx.Lifecycle().SetPhase(life.PhaseStopping); err != nil {
		t.Fatalf("SetPhase(Stopping) error = %v", err)
	}

	// 在 PhaseStopping 调用 Stop() — 应直接返回 nil
	if err := boot.Stop(); err != nil {
		t.Fatalf("Stop() during PhaseStopping should return nil, got: %v", err)
	}
}

func TestBoot_Stop_OnlyStopsStartedStarters(t *testing.T) {
	t.Parallel()

	s := newMockStarter("test")
	// 通过全局注册表注册，Start() 会从全局注册表加载
	orig := globalStarterRegistry
	registry := NewStarterRegistry()
	globalStarterRegistry = registry
	t.Cleanup(func() { globalStarterRegistry = orig })
	registry.Add(s)

	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	// 正常启动 — starter 会经历 Configure → Start
	if err := boot.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if !s.started {
		t.Fatal("starter should be started after Start()")
	}

	if err := boot.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if !s.stopped {
		t.Error("starter.Stop() should be called after proper start")
	}
}

func TestBoot_Stop_WithConditionalStarters(t *testing.T) {
	enabled := newMockStarter("enabled")
	disabled := newMockStarterWithCondition("disabled", condition.OnProperty("never.match"))

	orig := globalStarterRegistry
	registry := NewStarterRegistry()
	globalStarterRegistry = registry
	defer func() { globalStarterRegistry = orig }()
	registry.Add(enabled)
	registry.Add(disabled)

	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	if err := boot.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := boot.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if !enabled.stopped {
		t.Error("enabled starter should be stopped")
	}
	if disabled.stopped {
		t.Error("disabled starter should NOT be stopped when condition does not match")
	}
}

// mockStarterWithCondition 支持条件的 mock starter
type mockStarterWithCondition struct {
	*mockStarter
	cond condition.Condition
}

func newMockStarterWithCondition(name string, cond condition.Condition) *mockStarterWithCondition {
	return &mockStarterWithCondition{
		mockStarter: newMockStarter(name),
		cond:        cond,
	}
}

func (m *mockStarterWithCondition) GetCondition() condition.Condition {
	return m.cond
}
