package life

import (
	"errors"
	"testing"
)

func TestLifecycleBuilder_Defaults(t *testing.T) {
	builder := NewLifecycleBuilder()

	if builder.initialPhase != PhaseInitializing {
		t.Errorf("expected default initialPhase PhaseInitializing, got %v", builder.initialPhase)
	}

	if builder.listeners == nil {
		t.Error("expected non-nil listeners")
	}
}

func TestLifecycleBuilder_ChainConfig(t *testing.T) {
	recorder := NewPhaseTransitionRecorder()

	mgr := NewLifecycleBuilder().
		InitialPhase(PhaseInitializing).
		Listener(recorder).
		OnError(func(oldPhase, newPhase ApplicationPhase, err error) {
			// error handler for testing
		}).
		Build()

	if mgr == nil {
		t.Fatal("expected non-nil manager")
	}

	if mgr.GetPhase() != PhaseInitializing {
		t.Errorf("expected phase PhaseInitializing, got %v", mgr.GetPhase())
	}
}

func TestLifecycleBuilder_MultipleListeners(t *testing.T) {
	recorder1 := NewPhaseTransitionRecorder()
	recorder2 := NewPhaseTransitionRecorder()

	mgr := NewLifecycleBuilder().
		Listener(recorder1).
		Listener(recorder2).
		Build()

	err := mgr.SetPhase(PhaseConfiguring)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if recorder1.Count() != 1 {
		t.Errorf("expected recorder1 to have 1 transition, got %d", recorder1.Count())
	}

	if recorder2.Count() != 1 {
		t.Errorf("expected recorder2 to have 1 transition, got %d", recorder2.Count())
	}
}

func TestLifecycleBuilder_WithErrorHandler(t *testing.T) {
	var capturedErr error
	var capturedOldPhase, capturedNewPhase ApplicationPhase

	mgr := NewLifecycleBuilder().
		OnError(func(oldPhase, newPhase ApplicationPhase, err error) {
			capturedOldPhase = oldPhase
			capturedNewPhase = newPhase
			capturedErr = err
		}).
		Build()

	// 添加一个会返回错误的监听器
	errorListener := &errorPhaseListener{err: errors.New("test error")}
	mgr.AddListener(errorListener)

	_ = mgr.SetPhase(PhaseConfiguring)

	if capturedErr == nil {
		t.Fatal("expected error to be captured")
	}

	if capturedErr.Error() != "test error" {
		t.Errorf("expected error 'test error', got %v", capturedErr)
	}

	if capturedOldPhase != PhaseInitializing {
		t.Errorf("expected oldPhase PhaseInitializing, got %v", capturedOldPhase)
	}

	if capturedNewPhase != PhaseConfiguring {
		t.Errorf("expected newPhase PhaseConfiguring, got %v", capturedNewPhase)
	}
}

type errorPhaseListener struct {
	err error
}

func (l *errorPhaseListener) OnPhaseChange(oldPhase, newPhase ApplicationPhase) error {
	return l.err
}

func TestPhaseChecker_IsInitializing(t *testing.T) {
	mgr := NewLifecycleManager()
	checker := NewPhaseChecker(mgr)

	if !checker.IsInitializing() {
		t.Error("expected IsInitializing to be true")
	}

	if checker.IsConfiguring() {
		t.Error("expected IsConfiguring to be false")
	}
}

func TestPhaseChecker_IsConfiguring(t *testing.T) {
	mgr := NewLifecycleManager()
	_ = mgr.SetPhase(PhaseConfiguring)
	checker := NewPhaseChecker(mgr)

	if !checker.IsConfiguring() {
		t.Error("expected IsConfiguring to be true")
	}

	if checker.IsInitializing() {
		t.Error("expected IsInitializing to be false")
	}
}

func TestPhaseChecker_IsReady(t *testing.T) {
	mgr := NewLifecycleManager()
	_ = mgr.SetPhase(PhaseConfiguring)
	_ = mgr.SetPhase(PhaseContextRefreshed)
	_ = mgr.SetPhase(PhaseReady)
	checker := NewPhaseChecker(mgr)

	if !checker.IsReady() {
		t.Error("expected IsReady to be true")
	}
}

func TestPhaseChecker_IsRunning(t *testing.T) {
	mgr := NewLifecycleManager()
	_ = mgr.SetPhase(PhaseConfiguring)
	_ = mgr.SetPhase(PhaseContextRefreshed)
	_ = mgr.SetPhase(PhaseReady)
	_ = mgr.SetPhase(PhaseRunning)
	checker := NewPhaseChecker(mgr)

	if !checker.IsRunning() {
		t.Error("expected IsRunning to be true")
	}
}

func TestPhaseChecker_IsStopping(t *testing.T) {
	mgr := NewLifecycleManager()
	_ = mgr.SetPhase(PhaseConfiguring)
	_ = mgr.SetPhase(PhaseContextRefreshed)
	_ = mgr.SetPhase(PhaseReady)
	_ = mgr.SetPhase(PhaseRunning)
	_ = mgr.SetPhase(PhaseStopping)
	checker := NewPhaseChecker(mgr)

	if !checker.IsStopping() {
		t.Error("expected IsStopping to be true")
	}
}

func TestPhaseChecker_IsStopped(t *testing.T) {
	mgr := NewLifecycleManager()
	_ = mgr.SetPhase(PhaseConfiguring)
	_ = mgr.SetPhase(PhaseContextRefreshed)
	_ = mgr.SetPhase(PhaseReady)
	_ = mgr.SetPhase(PhaseRunning)
	_ = mgr.SetPhase(PhaseStopping)
	_ = mgr.SetPhase(PhaseStopped)
	checker := NewPhaseChecker(mgr)

	if !checker.IsStopped() {
		t.Error("expected IsStopped to be true")
	}
}

func TestPhaseChecker_IsBefore(t *testing.T) {
	mgr := NewLifecycleManager()
	checker := NewPhaseChecker(mgr)

	if !checker.IsBefore(PhaseConfiguring) {
		t.Error("expected IsBefore(PhaseConfiguring) to be true")
	}

	if checker.IsBefore(PhaseInitializing) {
		t.Error("expected IsBefore(PhaseInitializing) to be false")
	}
}

func TestPhaseChecker_IsAfter(t *testing.T) {
	mgr := NewLifecycleManager()
	_ = mgr.SetPhase(PhaseRunning)
	checker := NewPhaseChecker(mgr)

	if !checker.IsAfter(PhaseConfiguring) {
		t.Error("expected IsAfter(PhaseConfiguring) to be true")
	}

	if checker.IsAfter(PhaseRunning) {
		t.Error("expected IsAfter(PhaseRunning) to be false")
	}
}

func TestPhaseChecker_IsAtOrBefore(t *testing.T) {
	mgr := NewLifecycleManager()
	checker := NewPhaseChecker(mgr)

	if !checker.IsAtOrBefore(PhaseInitializing) {
		t.Error("expected IsAtOrBefore(PhaseInitializing) to be true")
	}

	if !checker.IsAtOrBefore(PhaseConfiguring) {
		t.Error("expected IsAtOrBefore(PhaseConfiguring) to be true")
	}

	if checker.IsAtOrBefore(PhaseInitializing) {
		// 应该为true，因为当前就是PhaseInitializing
	} else {
		t.Error("expected IsAtOrBefore(PhaseInitializing) to be true")
	}
}

func TestPhaseChecker_IsAtOrAfter(t *testing.T) {
	mgr := NewLifecycleManager()
	_ = mgr.SetPhase(PhaseRunning)
	checker := NewPhaseChecker(mgr)

	if !checker.IsAtOrAfter(PhaseRunning) {
		t.Error("expected IsAtOrAfter(PhaseRunning) to be true")
	}

	if !checker.IsAtOrAfter(PhaseConfiguring) {
		t.Error("expected IsAtOrAfter(PhaseConfiguring) to be true")
	}

	if checker.IsAtOrAfter(PhaseStopping) {
		t.Error("expected IsAtOrAfter(PhaseStopping) to be false")
	}
}

func TestPhaseTransitionRecorder_RecordTransitions(t *testing.T) {
	recorder := NewPhaseTransitionRecorder()

	_ = recorder.OnPhaseChange(PhaseInitializing, PhaseConfiguring)
	_ = recorder.OnPhaseChange(PhaseConfiguring, PhaseContextRefreshed)
	_ = recorder.OnPhaseChange(PhaseContextRefreshed, PhaseReady)

	if recorder.Count() != 3 {
		t.Errorf("expected 3 transitions, got %d", recorder.Count())
	}

	transitions := recorder.GetTransitions()
	if len(transitions) != 3 {
		t.Errorf("expected 3 transitions, got %d", len(transitions))
	}

	if transitions[0].OldPhase != PhaseInitializing {
		t.Errorf("expected first transition OldPhase PhaseInitializing, got %v", transitions[0].OldPhase)
	}

	if transitions[0].NewPhase != PhaseConfiguring {
		t.Errorf("expected first transition NewPhase PhaseConfiguring, got %v", transitions[0].NewPhase)
	}
}

func TestPhaseTransitionRecorder_GetLastTransition(t *testing.T) {
	recorder := NewPhaseTransitionRecorder()

	// 空记录
	_, ok := recorder.GetLastTransition()
	if ok {
		t.Error("expected no last transition for empty recorder")
	}

	_ = recorder.OnPhaseChange(PhaseInitializing, PhaseConfiguring)
	_ = recorder.OnPhaseChange(PhaseConfiguring, PhaseContextRefreshed)

	last, ok := recorder.GetLastTransition()
	if !ok {
		t.Fatal("expected last transition")
	}

	if last.OldPhase != PhaseConfiguring {
		t.Errorf("expected last OldPhase PhaseConfiguring, got %v", last.OldPhase)
	}

	if last.NewPhase != PhaseContextRefreshed {
		t.Errorf("expected last NewPhase PhaseContextRefreshed, got %v", last.NewPhase)
	}
}

func TestPhaseTransitionRecorder_ConcurrentSafe(t *testing.T) {
	recorder := NewPhaseTransitionRecorder()

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = recorder.OnPhaseChange(PhaseInitializing, PhaseConfiguring)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if recorder.Count() != 1000 {
		t.Errorf("expected 1000 transitions, got %d", recorder.Count())
	}
}

func TestPhaseTransition_String(t *testing.T) {
	transition := PhaseTransition{
		OldPhase: PhaseInitializing,
		NewPhase: PhaseConfiguring,
	}

	expected := "INITIALIZING -> CONFIGURING"
	if transition.String() != expected {
		t.Errorf("expected %s, got %s", expected, transition.String())
	}
}

func TestLifecycleManager_SetErrorHandler(t *testing.T) {
	mgr := NewLifecycleManager()

	handlerCalled := false
	mgr.SetErrorHandler(func(oldPhase, newPhase ApplicationPhase, err error) {
		handlerCalled = true
	})

	// 添加会返回错误的监听器
	mgr.AddListener(&errorPhaseListener{err: errors.New("test error")})

	_ = mgr.SetPhase(PhaseConfiguring)

	if !handlerCalled {
		t.Error("expected error handler to be called")
	}
}

func TestLifecycleManager_FullLifecycle(t *testing.T) {
	recorder := NewPhaseTransitionRecorder()
	mgr := NewLifecycleBuilder().
		Listener(recorder).
		Build()

	phases := []ApplicationPhase{
		PhaseConfiguring,
		PhaseContextRefreshed,
		PhaseReady,
		PhaseRunning,
		PhaseStopping,
		PhaseStopped,
	}

	for _, phase := range phases {
		err := mgr.SetPhase(phase)
		if err != nil {
			t.Fatalf("failed to set phase %v: %v", phase, err)
		}
	}

	if recorder.Count() != 6 {
		t.Errorf("expected 6 transitions, got %d", recorder.Count())
	}

	if mgr.GetPhase() != PhaseStopped {
		t.Errorf("expected final phase PhaseStopped, got %v", mgr.GetPhase())
	}
}

func TestLifecycleManager_InvalidTransition(t *testing.T) {
	mgr := NewLifecycleManager()

	_ = mgr.SetPhase(PhaseConfiguring)
	_ = mgr.SetPhase(PhaseContextRefreshed)

	// 尝试逆向转换
	err := mgr.SetPhase(PhaseInitializing)
	if err == nil {
		t.Error("expected error for invalid transition")
	}
}
