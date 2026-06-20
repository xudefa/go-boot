package actuator

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xudefa/go-boot/health"
)

func TestHealthIndicatorBuilder_Build_Success(t *testing.T) {
	indicator := NewHealthIndicatorBuilder().
		Name("test").
		CheckFunc(func(ctx context.Context) error { return nil }).
		Timeout(2*time.Second).
		Detail("key", "value").
		Build()

	if indicator.Name() != "test" {
		t.Errorf("Expected name 'test', got %s", indicator.Name())
	}

	h := indicator.Health(context.Background())
	if h.Status != health.StatusUp {
		t.Errorf("Expected status UP, got %s", h.Status)
	}
	if h.Details["key"] != "value" {
		t.Errorf("Expected detail key=value, got %v", h.Details["key"])
	}
}

func TestHealthIndicatorBuilder_Build_Failure(t *testing.T) {
	indicator := NewHealthIndicatorBuilder().
		Name("test").
		CheckFunc(func(ctx context.Context) error { return errors.New("connection failed") }).
		Build()

	h := indicator.Health(context.Background())
	if h.Status != health.StatusDown {
		t.Errorf("Expected status DOWN, got %s", h.Status)
	}
	if h.Details["error"] != "connection failed" {
		t.Errorf("Expected error 'connection failed', got %v", h.Details["error"])
	}
}

func TestHealthIndicatorBuilder_Build_NoCheckFunc(t *testing.T) {
	indicator := NewHealthIndicatorBuilder().
		Name("test").
		Build()

	h := indicator.Health(context.Background())
	if h.Status != health.StatusUnknown {
		t.Errorf("Expected status UNKNOWN, got %s", h.Status)
	}
}

func TestHealthIndicatorBuilder_Build_Timeout(t *testing.T) {
	indicator := NewHealthIndicatorBuilder().
		Name("test").
		CheckFunc(func(ctx context.Context) error {
			// 模拟一个会检查上下文的操作
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(200 * time.Millisecond):
				return nil
			}
		}).
		Timeout(50 * time.Millisecond).
		Build()

	h := indicator.Health(context.Background())
	if h.Status != health.StatusDown {
		t.Errorf("Expected status DOWN due to timeout, got %s", h.Status)
	}
}
