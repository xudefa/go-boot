package life

import (
	"testing"
)

type testListener struct {
	oldPhase ApplicationPhase
	newPhase ApplicationPhase
	called   bool
	fail     bool
}

func (l *testListener) OnPhaseChange(oldPhase, newPhase ApplicationPhase) error {
	l.called = true
	l.oldPhase = oldPhase
	l.newPhase = newPhase
	if l.fail {
		return testError{"listener failed"}
	}
	return nil
}

type testError struct {
	msg string
}

func (e testError) Error() string {
	return e.msg
}

func TestNewLifecycleManager(t *testing.T) {
	m := NewLifecycleManager()
	if m == nil {
		t.Fatal("NewLifecycleManager() returned nil")
	}
	if m.GetPhase() != PhaseInitializing {
		t.Fatalf("expected PhaseInitializing, got %v", m.GetPhase())
	}
}

func TestSetPhase(t *testing.T) {
	m := NewLifecycleManager()

	err := m.SetPhase(PhaseConfiguring)
	if err != nil {
		t.Fatalf("SetPhase() error = %v", err)
	}
	if m.GetPhase() != PhaseConfiguring {
		t.Fatalf("expected PhaseConfiguring, got %v", m.GetPhase())
	}
}

func TestAddListener(t *testing.T) {
	m := NewLifecycleManager()
	listener := &testListener{}
	m.AddListener(listener)

	err := m.SetPhase(PhaseConfiguring)
	if err != nil {
		t.Fatalf("SetPhase() error = %v", err)
	}

	if !listener.called {
		t.Fatal("listener should have been called")
	}
	if listener.oldPhase != PhaseInitializing {
		t.Fatalf("expected oldPhase PhaseInitializing, got %v", listener.oldPhase)
	}
	if listener.newPhase != PhaseConfiguring {
		t.Fatalf("expected newPhase PhaseConfiguring, got %v", listener.newPhase)
	}
}

func TestSetPhase_MultipleListeners(t *testing.T) {
	m := NewLifecycleManager()
	l1 := &testListener{}
	l2 := &testListener{}
	m.AddListener(l1)
	m.AddListener(l2)

	err := m.SetPhase(PhaseConfiguring)
	if err != nil {
		t.Fatalf("SetPhase() error = %v", err)
	}

	if !l1.called {
		t.Fatal("listener 1 should have been called")
	}
	if !l2.called {
		t.Fatal("listener 2 should have been called")
	}
}

func TestSetPhase_InvalidTransition(t *testing.T) {
	m := NewLifecycleManager()

	_ = m.SetPhase(PhaseRunning)

	// 逆向转换是非法的
	err := m.SetPhase(PhaseInitializing)
	if err == nil {
		t.Fatal("expected error for backward transition")
	}

	err = m.SetPhase(PhaseReady)
	if err == nil {
		t.Fatal("expected error for backward transition")
	}
}

func TestSetPhase_ListenerError(t *testing.T) {
	m := NewLifecycleManager()
	m.AddListener(&testListener{fail: true})

	err := m.SetPhase(PhaseConfiguring)
	if err == nil {
		t.Fatal("expected error from failing listener")
	}
}

func TestPhaseString(t *testing.T) {
	tests := []struct {
		phase ApplicationPhase
		want  string
	}{
		{PhaseInitializing, "INITIALIZING"},
		{PhaseConfiguring, "CONFIGURING"},
		{PhaseContextRefreshed, "CONTEXT_REFRESHED"},
		{PhaseReady, "READY"},
		{PhaseRunning, "RUNNING"},
		{PhaseStopping, "STOPPING"},
		{PhaseStopped, "STOPPED"},
		{ApplicationPhase(99), "UNKNOWN(99)"},
	}

	for _, tt := range tests {
		got := tt.phase.String()
		if got != tt.want {
			t.Errorf("ApplicationPhase(%d).String() = %q, want %q", tt.phase, got, tt.want)
		}
	}
}
