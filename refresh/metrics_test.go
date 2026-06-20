package refresh

import (
	"testing"
	"time"
)

func TestRefreshMetrics_RecordRefresh(t *testing.T) {
	metrics := NewRefreshMetrics()

	metrics.RecordRefresh(100*time.Millisecond, true)
	metrics.RecordRefresh(200*time.Millisecond, false)
	metrics.RecordRefresh(300*time.Millisecond, true)

	if metrics.TotalRefreshes() != 3 {
		t.Errorf("Expected 3 total refreshes, got %d", metrics.TotalRefreshes())
	}

	if metrics.SuccessfulRefreshes() != 2 {
		t.Errorf("Expected 2 successful refreshes, got %d", metrics.SuccessfulRefreshes())
	}

	if metrics.FailedRefreshes() != 1 {
		t.Errorf("Expected 1 failed refresh, got %d", metrics.FailedRefreshes())
	}

	avgTime := metrics.AverageRefreshTime()
	expectedAvg := 200 * time.Millisecond
	if avgTime != expectedAvg {
		t.Errorf("Expected average time %v, got %v", expectedAvg, avgTime)
	}
}

func TestRefreshMetrics_LastRefreshTime(t *testing.T) {
	metrics := NewRefreshMetrics()

	before := time.Now()
	metrics.RecordRefresh(100*time.Millisecond, true)
	after := time.Now()

	lastTime := metrics.LastRefreshTime()
	if lastTime.Before(before) || lastTime.After(after) {
		t.Errorf("Last refresh time %v is not between %v and %v", lastTime, before, after)
	}
}

func TestRefreshMetrics_Reset(t *testing.T) {
	metrics := NewRefreshMetrics()

	metrics.RecordRefresh(100*time.Millisecond, true)
	metrics.RecordRefresh(200*time.Millisecond, false)

	metrics.Reset()

	if metrics.TotalRefreshes() != 0 {
		t.Errorf("Expected 0 total refreshes after reset, got %d", metrics.TotalRefreshes())
	}

	if metrics.SuccessfulRefreshes() != 0 {
		t.Errorf("Expected 0 successful refreshes after reset, got %d", metrics.SuccessfulRefreshes())
	}

	if metrics.FailedRefreshes() != 0 {
		t.Errorf("Expected 0 failed refreshes after reset, got %d", metrics.FailedRefreshes())
	}

	if metrics.AverageRefreshTime() != 0 {
		t.Errorf("Expected 0 average time after reset, got %v", metrics.AverageRefreshTime())
	}
}
