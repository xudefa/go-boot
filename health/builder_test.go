package health

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestIndicatorBuilder_Build_Up(t *testing.T) {
	indicator := NewIndicatorBuilder().
		Name("database").
		CheckFunc(func(ctx context.Context) error {
			return nil
		}).
		Timeout(5*time.Second).
		Detail("type", "postgres").
		Build()

	h := indicator.Health(context.Background())

	if h.Status != StatusUp {
		t.Errorf("expected UP, got %s", h.Status)
	}
	if h.Details["type"] != "postgres" {
		t.Errorf("expected postgres, got %v", h.Details["type"])
	}
	if h.Error != nil {
		t.Errorf("expected no error, got %v", h.Error)
	}
}

func TestIndicatorBuilder_Build_Down(t *testing.T) {
	indicator := NewIndicatorBuilder().
		Name("redis").
		CheckFunc(func(ctx context.Context) error {
			return errors.New("connection refused")
		}).
		Build()

	h := indicator.Health(context.Background())

	if h.Status != StatusDown {
		t.Errorf("expected DOWN, got %s", h.Status)
	}
	if h.Error == nil {
		t.Error("expected error")
	}
}

func TestIndicatorBuilder_Build_Timeout(t *testing.T) {
	indicator := NewIndicatorBuilder().
		Name("slow-service").
		CheckFunc(func(ctx context.Context) error {
			select {
			case <-time.After(10 * time.Second):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}).
		Timeout(100 * time.Millisecond).
		Build()

	start := time.Now()
	h := indicator.Health(context.Background())
	elapsed := time.Since(start)

	if h.Status != StatusDown {
		t.Errorf("expected DOWN due to timeout, got %s", h.Status)
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("expected timeout around 100ms, took %v", elapsed)
	}
}

func TestIndicatorBuilder_Build_NoCheckFunc(t *testing.T) {
	indicator := NewIndicatorBuilder().
		Name("unknown").
		Build()

	h := indicator.Health(context.Background())

	if h.Status != StatusUnknown {
		t.Errorf("expected UNKNOWN, got %s", h.Status)
	}
}

func TestIndicatorBuilder_ChainConfiguration(t *testing.T) {
	indicator := NewIndicatorBuilder().
		Name("test-service").
		Timeout(3*time.Second).
		Detail("version", "1.0.0").
		Detail("env", "production").
		CheckFunc(func(ctx context.Context) error {
			return nil
		}).
		Build()

	h := indicator.Health(context.Background())

	if h.Status != StatusUp {
		t.Errorf("expected UP, got %s", h.Status)
	}
	if h.Details["version"] != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %v", h.Details["version"])
	}
	if h.Details["env"] != "production" {
		t.Errorf("expected env production, got %v", h.Details["env"])
	}
}

func TestIndicatorBuilder_MultipleIndicators(t *testing.T) {
	registry := NewAggregator()

	dbIndicator := NewIndicatorBuilder().
		Name("database").
		CheckFunc(func(ctx context.Context) error {
			return nil
		}).
		Build()

	redisIndicator := NewIndicatorBuilder().
		Name("redis").
		CheckFunc(func(ctx context.Context) error {
			return errors.New("redis down")
		}).
		Build()

	registry.AddIndicator(dbIndicator)
	registry.AddIndicator(redisIndicator)

	h := registry.Aggregate(context.Background())

	// 数据库应该 UP，Redis 应该 DOWN
	details := h.Details["database"].(map[string]any)
	if details["status"] != "UP" {
		t.Errorf("expected database UP, got %s", details["status"])
	}

	redisDetails := h.Details["redis"].(map[string]any)
	if redisDetails["status"] != "DOWN" {
		t.Errorf("expected redis DOWN, got %s", redisDetails["status"])
	}
}
