package circuit

import (
	"testing"
	"time"
)

func TestNewDefaultBreaker(t *testing.T) {
	b := NewDefaultBreaker()
	if b == nil {
		t.Fatal("expected breaker to be created")
	}
	if b.state != StateClosed {
		t.Errorf("expected initial state to be closed, got %v", b.state)
	}
}

func TestBreakerAllow(t *testing.T) {
	b := NewDefaultBreaker()
	if err := b.Allow(); err != nil {
		t.Errorf("expected no error in closed state, got %v", err)
	}
}

func TestBreakerOpenAfterFailures(t *testing.T) {
	b := NewDefaultBreaker(
		WithErrorThreshold(0.5),
		WithMaxRequests(5),
	)

	// 记录足够的失败请求以触发熔断
	for i := 0; i < 10; i++ {
		_ = b.Allow()
		b.RecordFailure()
	}

	if b.State() != StateOpen {
		t.Errorf("expected state to be open, got %v", b.State())
	}

	if err := b.Allow(); err != ErrCircuitOpen {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestBreakerHalfOpenAfterWait(t *testing.T) {
	b := NewDefaultBreaker(
		WithErrorThreshold(0.5),
		WithWaitDuration(50*time.Millisecond),
	)

	// 触发熔断
	for i := 0; i < 10; i++ {
		_ = b.Allow()
		b.RecordFailure()
	}

	if b.State() != StateOpen {
		t.Fatalf("expected state to be open, got %v", b.State())
	}

	// 等待恢复时间
	time.Sleep(60 * time.Millisecond)

	if err := b.Allow(); err != nil {
		t.Errorf("expected no error in half-open state, got %v", err)
	}

	if b.State() != StateHalfOpen {
		t.Errorf("expected state to be half-open, got %v", b.State())
	}
}

func TestBreakerRecovery(t *testing.T) {
	b := NewDefaultBreaker(
		WithErrorThreshold(0.5),
		WithWaitDuration(50*time.Millisecond),
		WithMaxRequests(3),
	)

	// 触发熔断
	for i := 0; i < 10; i++ {
		_ = b.Allow()
		b.RecordFailure()
	}

	// 等待恢复
	time.Sleep(60 * time.Millisecond)

	// 成功请求恢复
	for i := 0; i < 3; i++ {
		_ = b.Allow()
		b.RecordSuccess()
	}

	if b.State() != StateClosed {
		t.Errorf("expected state to be closed after recovery, got %v", b.State())
	}
}

func TestBreakerStateString(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
		{State(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.expected {
			t.Errorf("State(%d).String() = %q, want %q", tt.state, got, tt.expected)
		}
	}
}

func TestBreakerOptions(t *testing.T) {
	b := NewDefaultBreaker(
		WithMaxRequests(20),
		WithErrorThreshold(0.3),
		WithWaitDuration(60*time.Second),
	)

	if b.maxRequests != 20 {
		t.Errorf("expected maxRequests=20, got %d", b.maxRequests)
	}
	if b.errorThreshold != 0.3 {
		t.Errorf("expected errorThreshold=0.3, got %f", b.errorThreshold)
	}
	if b.waitDuration != 60*time.Second {
		t.Errorf("expected waitDuration=60s, got %v", b.waitDuration)
	}
}
