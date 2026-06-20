package metrics

import (
	"strings"
	"testing"
)

func TestSimpleCounter(t *testing.T) {
	c := NewSimpleCounter()
	if c.Value() != 0 {
		t.Fatalf("expected 0, got %f", c.Value())
	}

	c.Inc()
	if c.Value() != 1 {
		t.Fatalf("expected 1, got %f", c.Value())
	}

	c.Add(2.5)
	if c.Value() != 3.5 {
		t.Fatalf("expected 3.5, got %f", c.Value())
	}
}

func TestSimpleCounter_Reset(t *testing.T) {
	c := NewSimpleCounter()
	c.Add(10)
	if c.Value() != 10 {
		t.Fatalf("expected 10, got %f", c.Value())
	}

	c.Reset()
	if c.Value() != 0 {
		t.Fatalf("expected 0 after reset, got %f", c.Value())
	}
}

func TestSimpleGauge(t *testing.T) {
	g := NewSimpleGauge()
	g.Set(42.5)
	if g.Value() != 42.5 {
		t.Fatalf("expected 42.5, got %f", g.Value())
	}

	g.Add(7.5)
	if g.Value() != 50.0 {
		t.Fatalf("expected 50.0, got %f", g.Value())
	}
}

func TestSimpleHistogram(t *testing.T) {
	h := NewSimpleHistogram("test", nil)
	h.Record(10)
	h.Record(20)
	h.Record(30)

	if h.Count() != 3 {
		t.Fatalf("expected count 3, got %d", h.Count())
	}
	if h.Sum() != 60 {
		t.Fatalf("expected sum 60, got %f", h.Sum())
	}
}

func TestSimpleHistogram_Reset(t *testing.T) {
	h := NewSimpleHistogram("test", nil)
	h.Record(10)
	h.Record(20)

	h.Reset()
	if h.Count() != 0 {
		t.Fatalf("expected count 0 after reset, got %d", h.Count())
	}
	if h.Sum() != 0 {
		t.Fatalf("expected sum 0 after reset, got %f", h.Sum())
	}
}

func TestSimpleHistogram_RecordWithLabels(t *testing.T) {
	h := NewSimpleHistogram("test", map[string]string{"service": "api"})
	h.RecordWithLabels(100, map[string]string{"endpoint": "/users"})
	h.Record(200)

	if h.Count() != 2 {
		t.Fatalf("expected count 2, got %d", h.Count())
	}
}

func TestSimpleRegistry(t *testing.T) {
	r := NewSimpleRegistry()
	c := r.Counter("requests")
	c.Inc()

	metrics := r.Collect()
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].Name != "requests" {
		t.Fatalf("expected 'requests', got %s", metrics[0].Name)
	}
	if metrics[0].Type != "counter" {
		t.Fatalf("expected type 'counter', got %s", metrics[0].Type)
	}
}

func TestSimpleRegistry_Tags(t *testing.T) {
	r := NewSimpleRegistry()
	c := r.Counter("http_requests", "method", "GET", "path", "/api")
	c.Inc()

	metrics := r.Collect()
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].Tags["method"] != "GET" {
		t.Fatalf("expected tag method=GET, got %v", metrics[0].Tags)
	}
	if metrics[0].Tags["path"] != "/api" {
		t.Fatalf("expected tag path=/api, got %v", metrics[0].Tags)
	}
}

func TestSimpleRegistry_GaugeTags(t *testing.T) {
	r := NewSimpleRegistry()
	g := r.Gauge("memory_usage", "region", "us-east-1", "host", "web-1")
	g.Set(1024.5)

	metrics := r.Collect()
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].Tags["region"] != "us-east-1" {
		t.Fatalf("expected tag region=us-east-1, got %v", metrics[0].Tags)
	}
	if metrics[0].Tags["host"] != "web-1" {
		t.Fatalf("expected tag host=web-1, got %v", metrics[0].Tags)
	}
}

func TestSimpleRegistry_CollectMultiple(t *testing.T) {
	r := NewSimpleRegistry()
	r.Counter("requests_total").Add(10)
	r.Counter("errors_total").Add(2)
	r.Gauge("active_connections").Set(5)
	r.Gauge("memory_mb").Set(256)

	metrics := r.Collect()
	if len(metrics) != 4 {
		t.Fatalf("expected 4 metrics, got %d", len(metrics))
	}
}

func TestSimpleRegistry_CounterReuse(t *testing.T) {
	r := NewSimpleRegistry()
	c1 := r.Counter("requests")
	c1.Add(5)
	c2 := r.Counter("requests")
	c2.Add(3)

	if c2.Value() != 8 {
		t.Fatalf("expected 8, got %f", c2.Value())
	}
}

func TestSimpleRegistry_GaugeReuse(t *testing.T) {
	r := NewSimpleRegistry()
	g1 := r.Gauge("memory")
	g1.Set(100)
	g2 := r.Gauge("memory")
	g2.Add(50)

	if g2.Value() != 150 {
		t.Fatalf("expected 150, got %f", g2.Value())
	}
}

func TestSimpleRegistry_NoTagsReturnsNilMap(t *testing.T) {
	r := NewSimpleRegistry()
	r.Counter("requests").Inc()

	metrics := r.Collect()
	if metrics[0].Tags != nil {
		t.Fatal("expected nil tags when no tags provided")
	}
}

func TestSimpleRegistry_OddTagsIgnored(t *testing.T) {
	r := NewSimpleRegistry()
	c := r.Counter("requests", "key1", "val1", "key2")
	c.Inc()

	metrics := r.Collect()
	if len(metrics[0].Tags) != 1 {
		t.Fatalf("expected 1 tag pair, got %d", len(metrics[0].Tags))
	}
}

func TestCounter_ConcurrentSafe(t *testing.T) {
	c := NewSimpleCounter()
	done := make(chan struct{})

	go func() {
		for range 1000 {
			c.Inc()
		}
		done <- struct{}{}
	}()

	go func() {
		for range 1000 {
			c.Add(1)
		}
		done <- struct{}{}
	}()

	go func() {
		for range 1000 {
			_ = c.Value()
		}
		done <- struct{}{}
	}()

	for range 3 {
		<-done
	}

	if c.Value() != 2000 {
		t.Fatalf("expected 2000, got %f", c.Value())
	}
}

func TestGauge_ConcurrentSafe(t *testing.T) {
	g := NewSimpleGauge()
	done := make(chan struct{})

	go func() {
		for range 1000 {
			g.Add(1)
		}
		done <- struct{}{}
	}()

	go func() {
		for range 1000 {
			g.Add(-1)
		}
		done <- struct{}{}
	}()

	go func() {
		for range 1000 {
			_ = g.Value()
		}
		done <- struct{}{}
	}()

	for range 3 {
		<-done
	}

	if g.Value() != 0 {
		t.Fatalf("expected 0, got %f", g.Value())
	}
}

func TestRegistry_ConcurrentSafe(t *testing.T) {
	r := NewSimpleRegistry()
	done := make(chan struct{})

	go func() {
		for range 100 {
			r.Counter("requests").Inc()
		}
		done <- struct{}{}
	}()

	go func() {
		for i := range 100 {
			r.Gauge("memory").Set(float64(i))
		}
		done <- struct{}{}
	}()

	go func() {
		for range 100 {
			_ = r.Collect()
		}
		done <- struct{}{}
	}()

	for range 3 {
		<-done
	}
}

func TestRegistry_Histogram(t *testing.T) {
	r := NewSimpleRegistry()
	h := r.Histogram("request_duration", "service", "api")
	h.Record(100.5)
	h.Record(200.5)

	metrics := r.Collect()
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].Type != "histogram" {
		t.Fatalf("expected type 'histogram', got %s", metrics[0].Type)
	}
	if metrics[0].Count != 2 {
		t.Fatalf("expected count 2, got %d", metrics[0].Count)
	}
	if metrics[0].Sum != 301.0 {
		t.Fatalf("expected sum 301.0, got %f", metrics[0].Sum)
	}
}

func TestRegistry_Reset(t *testing.T) {
	r := NewSimpleRegistry()
	r.Counter("requests").Add(10)
	r.Histogram("latency").Record(100)

	r.Reset()

	metrics := r.Collect()
	for _, m := range metrics {
		if m.Type == "counter" && m.Value != 0 {
			t.Fatalf("expected counter value 0 after reset, got %f", m.Value)
		}
		if m.Type == "histogram" && m.Count != 0 {
			t.Fatalf("expected histogram count 0 after reset, got %d", m.Count)
		}
	}
}

func TestRegistry_Export(t *testing.T) {
	r := NewSimpleRegistry()
	r.Counter("requests").Inc()

	exporter := NewConsoleExporter()
	r.RegisterExporter(exporter)

	err := r.Export()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSimpleCounter_AddNegativeValue(t *testing.T) {
	c := NewSimpleCounter()
	c.Add(10)
	c.Add(-5)

	if c.Value() != 5 {
		t.Fatalf("expected 5, got %f", c.Value())
	}
}

func TestSimpleCounter_ConcurrentSafe(t *testing.T) {
	c := NewSimpleCounter()
	const goroutines = 10
	const increments = 100

	done := make(chan struct{})

	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < increments; j++ {
				c.Inc()
			}
			done <- struct{}{}
		}()
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}

	expected := float64(goroutines * increments)
	if c.Value() != expected {
		t.Fatalf("expected %f, got %f", expected, c.Value())
	}
}

func TestSimpleRegistry_HistogramStats(t *testing.T) {
	r := NewSimpleRegistry()
	h := r.Histogram("request_duration_seconds")

	// 添加多个值
	values := []float64{100, 200, 300, 400, 500}
	for _, v := range values {
		h.Record(v)
	}

	metrics := r.Collect()
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}

	m := metrics[0]
	if m.Count != int64(len(values)) {
		t.Fatalf("expected count %d, got %d", len(values), m.Count)
	}

	if m.Sum != 1500 { // 100+200+300+400+500
		t.Fatalf("expected sum 1500, got %f", m.Sum)
	}
}

func TestPrometheusExporter(t *testing.T) {
	// 创建一个简单的内存写入器
	var buffer strings.Builder

	exporter := NewPrometheusExporter(&buffer)

	// 创建测试指标
	metrics := []Metric{
		{
			Name:  "test_counter",
			Value: 42.0,
			Type:  "counter",
			Tags:  map[string]string{"method": "GET", "path": "/api"},
		},
		{
			Name:  "test_gauge",
			Value: 123.45,
			Type:  "gauge",
			Tags:  nil,
		},
	}

	err := exporter.Export(metrics)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buffer.String()
	if !strings.Contains(output, `test_counter{method="GET",path="/api"} 42`) {
		t.Fatalf("expected counter with labels in output, got: %s", output)
	}
	if !strings.Contains(output, `test_gauge 123.45`) {
		t.Fatalf("expected gauge in output, got: %s", output)
	}
}

func TestRegistry_MultipleExporters(t *testing.T) {
	r := NewSimpleRegistry()
	r.Counter("requests").Inc()

	// 创建多个导出器
	var buffer1, buffer2 strings.Builder
	exporter1 := NewPrometheusExporter(&buffer1)
	exporter2 := NewPrometheusExporter(&buffer2)

	r.RegisterExporter(exporter1)
	r.RegisterExporter(exporter2)

	err := r.Export()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 检查两个导出器都收到了数据
	output1 := buffer1.String()
	output2 := buffer2.String()

	if !strings.Contains(output1, "requests") {
		t.Fatalf("expected requests in exporter1 output, got: %s", output1)
	}
	if !strings.Contains(output2, "requests") {
		t.Fatalf("expected requests in exporter2 output, got: %s", output2)
	}
}
