package metrics

import (
	"testing"
	"time"
)

func TestCounter(t *testing.T) {
	c := NewCounter("test_counter")

	c.Inc(1)
	c.Inc(2)

	if c.Value() != 3 {
		t.Errorf("expected 3, got %f", c.Value())
	}

	c.Reset()
	if c.Value() != 0 {
		t.Errorf("expected 0 after reset, got %f", c.Value())
	}
}

func TestGauge(t *testing.T) {
	g := NewGauge("test_gauge")

	g.Set(42.5)
	if g.Value() != 42.5 {
		t.Errorf("expected 42.5, got %f", g.Value())
	}

	g.Set(10.0)
	if g.Value() != 10.0 {
		t.Errorf("expected 10.0, got %f", g.Value())
	}
}

func TestHistogram(t *testing.T) {
	h := NewHistogram("test_histogram")

	h.Observe(10.0)
	h.Observe(20.0)
	h.Observe(30.0)

	if h.Count() != 3 {
		t.Errorf("expected count 3, got %d", h.Count())
	}
	if h.Sum() != 60.0 {
		t.Errorf("expected sum 60.0, got %f", h.Sum())
	}
	if h.Min() != 10.0 {
		t.Errorf("expected min 10.0, got %f", h.Min())
	}
	if h.Max() != 30.0 {
		t.Errorf("expected max 30.0, got %f", h.Max())
	}
	if h.Value() != 20.0 {
		t.Errorf("expected avg 20.0, got %f", h.Value())
	}
}

func TestTimer(t *testing.T) {
	timer := NewTimer("test_timer")
	timer.Start()
	time.Sleep(10 * time.Millisecond)
	timer.Stop()

	if timer.Duration() == 0 {
		t.Error("expected non-zero duration")
	}
	if timer.Value() == 0 {
		t.Error("expected non-zero value")
	}
}

func TestMetricsRegistry(t *testing.T) {
	registry := NewMetricsRegistry()

	counter := NewCounter("reg_counter")
	registry.Register(counter)

	metric, exists := registry.Get("reg_counter")
	if !exists {
		t.Fatal("expected metric to exist")
	}
	if metric.Name() != "reg_counter" {
		t.Errorf("expected name 'reg_counter', got %s", metric.Name())
	}

	metrics := registry.List()
	if len(metrics) != 1 {
		t.Errorf("expected 1 metric, got %d", len(metrics))
	}
}
