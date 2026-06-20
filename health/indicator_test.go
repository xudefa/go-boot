package health

import (
	"context"
	"fmt"
	"testing"
)

type testIndicator struct{}

func (t *testIndicator) Name() string {
	return "test"
}

func (t *testIndicator) Health(ctx context.Context) Health {
	return Health{
		Status: StatusUp,
		Details: map[string]any{
			"version": "1.0.0",
		},
	}
}

type degradedIndicator struct{}

func (d *degradedIndicator) Name() string {
	return "cache"
}

func (d *degradedIndicator) Health(ctx context.Context) Health {
	return Health{
		Status: StatusDegraded,
		Details: map[string]any{
			"latency": "high",
		},
	}
}

type downIndicator struct{}

func (d *downIndicator) Name() string {
	return "db"
}

func (d *downIndicator) Health(ctx context.Context) Health {
	return Health{
		Status: StatusDown,
		Details: map[string]any{
			"error": "connection refused",
		},
	}
}

type outageIndicator struct{}

func (o *outageIndicator) Name() string {
	return "redis"
}

func (o *outageIndicator) Health(ctx context.Context) Health {
	return Health{
		Status: StatusOutage,
		Details: map[string]any{
			"error": "service unavailable",
		},
	}
}

type unknownIndicator struct{}

func (u *unknownIndicator) Name() string {
	return "external"
}

func (u *unknownIndicator) Health(ctx context.Context) Health {
	return Health{
		Status: StatusUnknown,
		Details: map[string]any{
			"error": "no check function",
		},
	}
}

func TestHealthIndicator(t *testing.T) {
	indicator := &testIndicator{}
	h := indicator.Health(context.Background())

	if h.Status != StatusUp {
		t.Fatalf("expected StatusUp, got %v", h.Status)
	}
	if h.Details["version"] != "1.0.0" {
		t.Fatalf("expected version 1.0.0, got %v", h.Details["version"])
	}
}

func TestAggregateAllUp(t *testing.T) {
	aggregator := NewAggregator()
	aggregator.AddIndicator(&testIndicator{})
	aggregator.AddIndicator(&testIndicator{})

	overall := aggregator.Aggregate(context.Background())
	if overall.Status != StatusUp {
		t.Fatalf("expected StatusUp overall, got %v", overall.Status)
	}
}

func TestAggregateOneDown(t *testing.T) {
	aggregator := NewAggregator()
	aggregator.AddIndicator(&testIndicator{})
	aggregator.AddIndicator(&downIndicator{})

	overall := aggregator.Aggregate(context.Background())
	if overall.Status != StatusDown {
		t.Fatalf("expected StatusDown overall, got %v", overall.Status)
	}
}

func TestAggregator_IndicatorsCopy(t *testing.T) {
	aggregator := NewAggregator()
	aggregator.AddIndicator(&testIndicator{})

	indicators := aggregator.Indicators()
	_ = indicators

	if len(aggregator.Indicators()) != 1 {
		t.Fatal("expected 1 indicator")
	}
}

func TestAggregateDegraded(t *testing.T) {
	aggregator := NewAggregator()
	aggregator.AddIndicator(&testIndicator{})
	aggregator.AddIndicator(&degradedIndicator{})

	overall := aggregator.Aggregate(context.Background())
	if overall.Status != StatusDegraded {
		t.Fatalf("expected StatusDegraded overall, got %v", overall.Status)
	}
}

func TestAggregator_SingleIndicator(t *testing.T) {
	aggregator := NewAggregator()
	aggregator.AddIndicator(&downIndicator{})

	overall := aggregator.Aggregate(context.Background())
	if overall.Status != StatusDown {
		t.Fatalf("expected StatusDown overall, got %v", overall.Status)
	}
}

func TestAggregator_Empty(t *testing.T) {
	aggregator := NewAggregator()

	overall := aggregator.Aggregate(context.Background())
	if overall.Status != StatusUp {
		t.Fatalf("expected StatusUp for empty aggregator, got %v", overall.Status)
	}
}

func TestAggregateOutage(t *testing.T) {
	aggregator := NewAggregator()
	aggregator.AddIndicator(&testIndicator{})
	aggregator.AddIndicator(&outageIndicator{})

	overall := aggregator.Aggregate(context.Background())
	if overall.Status != StatusOutage {
		t.Fatalf("expected StatusOutage overall, got %v", overall.Status)
	}
}

func TestAggregateOutagePriority(t *testing.T) {
	aggregator := NewAggregator()
	aggregator.AddIndicator(&downIndicator{})
	aggregator.AddIndicator(&outageIndicator{})

	overall := aggregator.Aggregate(context.Background())
	if overall.Status != StatusOutage {
		t.Fatalf("expected StatusOutage (outage > down), got %v", overall.Status)
	}
}

func TestAggregateUnknownIsUp(t *testing.T) {
	aggregator := NewAggregator()
	aggregator.AddIndicator(&unknownIndicator{})

	overall := aggregator.Aggregate(context.Background())
	if overall.Status != StatusUp {
		t.Fatalf("expected StatusUp (unknown ignored), got %v", overall.Status)
	}
}

func TestAggregateMixedWithUnknown(t *testing.T) {
	aggregator := NewAggregator()
	aggregator.AddIndicator(&testIndicator{})
	aggregator.AddIndicator(&unknownIndicator{})
	aggregator.AddIndicator(&downIndicator{})

	overall := aggregator.Aggregate(context.Background())
	if overall.Status != StatusDown {
		t.Fatalf("expected StatusDown, got %v", overall.Status)
	}
}

func TestHealth_Timestamp(t *testing.T) {
	aggregator := NewAggregator()
	aggregator.AddIndicator(&testIndicator{})

	overall := aggregator.Aggregate(context.Background())
	if overall.Timestamp.IsZero() {
		t.Fatal("expected non-zero timestamp")
	}
}

func TestAggregator_ConcurrencySafe(t *testing.T) {
	aggregator := NewAggregator()
	done := make(chan struct{})

	go func() {
		for range 100 {
			aggregator.AddIndicator(&testIndicator{})
		}
		done <- struct{}{}
	}()

	go func() {
		for range 100 {
			aggregator.Indicators()
		}
		done <- struct{}{}
	}()

	go func() {
		for range 100 {
			aggregator.Aggregate(context.Background())
		}
		done <- struct{}{}
	}()

	for range 3 {
		<-done
	}

	if len(aggregator.Indicators()) != 100 {
		t.Fatalf("expected 100 indicators, got %d", len(aggregator.Indicators()))
	}
}

func TestHealth_String(t *testing.T) {
	tests := []struct {
		status Status
		want   string
	}{
		{StatusUp, "UP"},
		{StatusDown, "DOWN"},
		{StatusDegraded, "DEGRADED"},
		{StatusOutage, "OUTAGE"},
		{StatusUnknown, "UNKNOWN"},
		{Status(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		got := tt.status.String()
		if got != tt.want {
			t.Errorf("Status(%d).String() = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestAggregator_MixedStatus(t *testing.T) {
	aggregator := NewAggregator()
	aggregator.AddIndicator(&testIndicator{})     // UP
	aggregator.AddIndicator(&degradedIndicator{}) // DEGRADED
	aggregator.AddIndicator(&downIndicator{})     // DOWN

	overall := aggregator.Aggregate(context.Background())
	if overall.Status != StatusDown {
		t.Fatalf("expected StatusDown for mixed status, got %v", overall.Status)
	}
}

func TestAggregator_OutagePriority(t *testing.T) {
	aggregator := NewAggregator()
	aggregator.AddIndicator(&testIndicator{})     // UP
	aggregator.AddIndicator(&degradedIndicator{}) // DEGRADED
	aggregator.AddIndicator(&downIndicator{})     // DOWN
	aggregator.AddIndicator(&outageIndicator{})   // OUTAGE

	overall := aggregator.Aggregate(context.Background())
	if overall.Status != StatusOutage {
		t.Fatalf("expected StatusOutage (highest priority), got %v", overall.Status)
	}
}

func TestAggregator_AddMultipleIndicators(t *testing.T) {
	aggregator := NewAggregator()

	// 添加多个相同的指示器
	for i := 0; i < 5; i++ {
		aggregator.AddIndicator(&testIndicator{})
	}

	indicators := aggregator.Indicators()
	if len(indicators) != 5 {
		t.Fatalf("expected 5 indicators, got %d", len(indicators))
	}

	overall := aggregator.Aggregate(context.Background())
	if overall.Status != StatusUp {
		t.Fatalf("expected StatusUp, got %v", overall.Status)
	}
}

type errorIndicator struct{}

func (e *errorIndicator) Name() string {
	return "error-indicator"
}

func (e *errorIndicator) Health(ctx context.Context) Health {
	return Health{
		Status: StatusDown,
		Error:  fmt.Errorf("connection failed"),
		Details: map[string]any{
			"host": "localhost",
		},
	}
}

func TestHealth_ErrorHandling(t *testing.T) {
	aggregator := NewAggregator()
	aggregator.AddIndicator(&errorIndicator{})

	overall := aggregator.Aggregate(context.Background())
	if overall.Status != StatusDown {
		t.Fatalf("expected StatusDown, got %v", overall.Status)
	}

	// 检查错误信息是否正确地包含在详情中
	details, ok := overall.Details["error-indicator"].(map[string]any)
	if !ok {
		t.Fatal("expected indicator details in overall details")
	}

	errorMsg, ok := details["error"].(string)
	if !ok || errorMsg != "connection failed" {
		t.Fatalf("expected 'connection failed', got %v", errorMsg)
	}
}
