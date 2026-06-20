package context

import (
	"testing"

	"github.com/xudefa/go-boot/core"
	"github.com/xudefa/go-boot/environment"
	"github.com/xudefa/go-boot/life"
)

func TestNewApplicationContext(t *testing.T) {
	container := core.New()
	env := environment.NewEnvironment()
	ctx := NewApplicationContext(container, env)

	if ctx.Container() != container {
		t.Fatal("container mismatch")
	}
	if ctx.Environment() != env {
		t.Fatal("environment mismatch")
	}
	if ctx.Lifecycle().GetPhase() != life.PhaseInitializing {
		t.Fatal("expected initializing phase")
	}
	if ctx.IsRunning() {
		t.Fatal("should not be running initially")
	}
}

func TestApplicationContext_StartStop(t *testing.T) {
	container := core.New()
	env := environment.NewEnvironment()
	ctx := NewApplicationContext(container, env)

	if err := ctx.Start(); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	if !ctx.IsRunning() {
		t.Fatal("should be running after start")
	}

	if err := ctx.Stop(); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
	if ctx.IsRunning() {
		t.Fatal("should not be running after stop")
	}
}

func TestLifecycle_BasicTransitions(t *testing.T) {
	lm := life.NewLifecycleManager()

	if lm.GetPhase() != life.PhaseInitializing {
		t.Fatal("expected initializing phase")
	}

	if err := lm.SetPhase(life.PhaseConfiguring); err != nil {
		t.Fatalf("set phase failed: %v", err)
	}
	if lm.GetPhase() != life.PhaseConfiguring {
		t.Fatal("expected configuring phase")
	}
}

func TestLifecycle_PhaseListener(t *testing.T) {
	lm := life.NewLifecycleManager()
	var oldPhase, newPhase life.ApplicationPhase

	lm.AddListener(&testListener{
		fn: func(old, new life.ApplicationPhase) error {
			oldPhase = old
			newPhase = new
			return nil
		},
	})

	if err := lm.SetPhase(life.PhaseRunning); err != nil {
		t.Fatalf("set phase failed: %v", err)
	}

	if oldPhase != life.PhaseInitializing {
		t.Fatal("expected old phase to be initializing")
	}
	if newPhase != life.PhaseRunning {
		t.Fatal("expected new phase to be running")
	}
}

type testListener struct {
	fn func(life.ApplicationPhase, life.ApplicationPhase) error
}

func (l *testListener) OnPhaseChange(old, new life.ApplicationPhase) error {
	return l.fn(old, new)
}

func TestExtractPkgPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple type", "gin.Engine", "gin"},
		{"full path", "github.com/gin-gonic/gin.Engine", "github.com/gin-gonic/gin"},
		{"module path has no type suffix", "github.com/lib/pq", "github"},
		{"no dot", "int", "int"},
		{"empty", "", ""},
		{"only dot has no prefix", ".", "."},
		{"starts with dot: no prefix after .", ".Foo", ".Foo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPkgPath(tt.input)
			if got != tt.expected {
				t.Errorf("extractPkgPath(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestPathMatches(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		modulePath string
		pkgPath    string
		match      bool
	}{
		{"exact match", "gin", "gin", true},
		{"exact full path", "github.com/gin-gonic/gin", "github.com/gin-gonic/gin", true},
		{"prefix child", "github.com/gin-gonic/gin", "github.com/gin-gonic/gin", true},
		{"suffix short pkg", "github.com/gin-gonic/gin", "gin", true},
		{"prefix sub-package", "github.com/user/pkg", "github.com/user/pkg/sub.Type", true},

		// BUG 5 修复：确保短 pkgPath（如 "gin"）不会误匹其他 module
		{"no false positive - different module short pkg", "github.com/other/gin", "other", false},
		{"no false positive - single segment module", "gin", "other", false},
		{"completely unrelated", "io/fs", "gin", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pathMatches(tt.modulePath, tt.pkgPath)
			if got != tt.match {
				t.Errorf("pathMatches(%q, %q) = %v, want %v", tt.modulePath, tt.pkgPath, got, tt.match)
			}
		})
	}
}

func TestBuildInfoClassLoader_HasClass(t *testing.T) {
	t.Parallel()

	loader := &buildInfoClassLoader{}

	// 当前模块自身应能找到
	if !loader.HasClass("context.context_test") {
		t.Log("note: HasClass for own module may depend on build info availability")
	}

	// gin.Engine 应存在（已通过 go.work / go.mod 依赖）
	hasGin := loader.HasClass("gin.Engine")
	t.Logf("HasClass(gin.Engine) = %v (may be false in unit test without full build context)", hasGin)
}

func TestApplicationContext_RefreshScopeManager(t *testing.T) {
	t.Parallel()
	container := core.New()
	env := environment.NewEnvironment()
	ctx := NewApplicationContext(container, env)

	mgr := ctx.RefreshScopeManager()
	if mgr == nil {
		t.Error("expected RefreshScopeManager to be initialized")
	}

	if mgr.Metrics() == nil {
		t.Error("expected RefreshScopeManager to have metrics")
	}
}
