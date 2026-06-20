package actuator

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xudefa/go-boot/core"
	"github.com/xudefa/go-boot/environment"
	"github.com/xudefa/go-boot/health"
	"github.com/xudefa/go-boot/metrics"
)

// testAppContext 测试用的应用上下文
type testAppContext struct {
	container core.Container
	env       *environment.Environment
}

func (t *testAppContext) Container() core.Container             { return t.container }
func (t *testAppContext) Environment() *environment.Environment { return t.env }

// newTestEnvironment 创建测试环境，移除默认的属性源
func newTestEnvironment() *environment.Environment {
	env := environment.NewEnvironment()
	env.RemovePropertySource("args")
	env.RemovePropertySource("env")
	return env
}

// testIndicator 测试用的健康指标
type testIndicator struct{}

func (t *testIndicator) Name() string {
	return "test"
}

func (t *testIndicator) Health(ctx context.Context) health.Health {
	return health.Health{
		Status: health.StatusUp,
		Details: map[string]any{
			"version": "1.0.0",
		},
	}
}

// TestNew 测试创建 Actuator 实例
func TestNew(t *testing.T) {
	container := core.New()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}

	act := New(ctx)
	if act == nil {
		t.Fatal("New() returned nil")
	}
	if act.healthAggregator == nil {
		t.Fatal("healthAggregator not initialized")
	}
	if act.metricsRegistry == nil {
		t.Fatal("metricsRegistry not initialized")
	}
	if act.appContext != ctx {
		t.Fatal("appContext not set correctly")
	}
}

// TestActuator_NilContext 测试当应用上下文为 nil 时的错误处理
func TestActuator_NilContext(t *testing.T) {
	act := &Actuator{}

	t.Run("EnviHandler returns 500", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/actuator/env", nil)
		act.EnvHandler(w, r)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("BeansHandler returns 500", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/actuator/beans", nil)
		act.BeansHandler(w, r)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}

// TestHealthHandler_NilAggregator 测试健康聚合器为 nil 时的错误处理
func TestHealthHandler_NilAggregator(t *testing.T) {
	act := &Actuator{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/health", nil)
	act.HealthHandler(w, r)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// TestMetricsHandler_NilRegistry 测试指标注册表为 nil 时的错误处理
func TestMetricsHandler_NilRegistry(t *testing.T) {
	act := &Actuator{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/metrics", nil)
	act.MetricsHandler(w, r)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// TestHealthHandler_EmptyAggregator 测试空健康聚合器的处理
func TestHealthHandler_EmptyAggregator(t *testing.T) {
	container := core.New()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/health", nil)
	act.HealthHandler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result health.Health
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.Status != health.StatusUp {
		t.Fatalf("expected UP, got %s", result.Status)
	}
}

// TestHealthHandler_WithIndicators 测试带健康指标的处理
func TestHealthHandler_WithIndicators(t *testing.T) {
	container := core.New()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	agg := health.NewAggregator()
	agg.AddIndicator(&testIndicator{})
	act.SetHealthAggregator(agg)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/health", nil)
	act.HealthHandler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result health.Health
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.Status != health.StatusUp {
		t.Fatalf("expected UP, got %s", result.Status)
	}
	if result.Details == nil {
		t.Fatal("expected details, got nil")
	}
}

// TestMetricsHandler_Empty 测试空指标的处理
func TestMetricsHandler_Empty(t *testing.T) {
	container := core.New()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/metrics", nil)
	act.MetricsHandler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result []metrics.Metric
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 metrics, got %d", len(result))
	}
}

// TestMetricsHandler_WithData 测试带指标数据的处理
func TestMetricsHandler_WithData(t *testing.T) {
	container := core.New()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	reg := metrics.NewSimpleRegistry()
	reg.Counter("requests_total").Inc()
	reg.Gauge("memory_usage").Set(1024.5)
	act.SetMetricsRegistry(reg)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/metrics", nil)
	act.MetricsHandler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result []metrics.Metric
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(result))
	}
}

// TestEnvHandler 测试环境信息处理器
func TestEnvHandler(t *testing.T) {
	container := core.New()
	env := newTestEnvironment()
	env.AddPropertySource(environment.NewDefaultPropertySource("test", map[string]any{
		"server.port": "8080",
		"app.name":    "test-app",
	}))
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/env", nil)
	act.EnvHandler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result []map[string]any
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least one property source")
	}

	found := false
	for _, src := range result {
		if src["name"] == "test" {
			found = true
			props, ok := src["properties"].([]any)
			if !ok {
				t.Fatal("expected properties array")
			}
			if len(props) != 2 {
				t.Fatalf("expected 2 properties, got %d", len(props))
			}
			break
		}
	}
	if !found {
		t.Fatal("expected to find 'test' property source")
	}
}

func TestEnvHandler_NoAppContext(t *testing.T) {
	act := &Actuator{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/env", nil)
	act.EnvHandler(w, r)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestBeansHandler_EmptyContainer(t *testing.T) {
	container := core.New()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/beans", nil)
	act.BeansHandler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result map[string]any
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	beans, ok := result["beans"].([]any)
	if !ok {
		t.Fatal("expected beans array")
	}
	if len(beans) != 0 {
		t.Fatalf("expected 0 beans, got %d", len(beans))
	}
}

func TestBeansHandler_WithBeans(t *testing.T) {
	container := core.New()
	_ = container.Register("service", core.Bean(&struct{ Name string }{Name: "test"}))
	_ = container.Register("db", core.Bean(&struct{}{}))
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/beans", nil)
	act.BeansHandler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result map[string]any
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	beans, ok := result["beans"].([]any)
	if !ok {
		t.Fatal("expected beans array")
	}
	if len(beans) != 2 {
		t.Fatalf("expected 2 beans, got %d", len(beans))
	}
}

func TestBeansHandler_NoAppContext(t *testing.T) {
	act := &Actuator{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/beans", nil)
	act.BeansHandler(w, r)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestRegisterRoutes(t *testing.T) {
	container := core.New()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	mux := http.NewServeMux()
	config := RouteConfig{
		BasePath:    "/actuator",
		ExposeDebug: false,
	}
	act.RegisterRoutes(&StdRouteRegistrar{Mux: mux}, config)

	routes := []string{
		"/actuator/health",
		"/actuator/metrics",
		"/actuator/env",
		"/actuator/beans",
	}

	for _, route := range routes {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, route, nil)
		mux.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("route %s returned %d", route, w.Code)
		}
	}
}

func TestSetHealthAggregator(t *testing.T) {
	container := core.New()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	agg := health.NewAggregator()
	agg.AddIndicator(&testIndicator{})
	act.SetHealthAggregator(agg)

	if act.healthAggregator == nil {
		t.Fatal("healthAggregator should not be nil")
	}
	h := act.healthAggregator.Aggregate(context.Background())
	if h.Status != health.StatusUp {
		t.Fatalf("expected UP, got %s", h.Status)
	}
}

func TestSetMetricsRegistry(t *testing.T) {
	container := core.New()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	reg := metrics.NewSimpleRegistry()
	reg.Counter("test").Inc()
	act.SetMetricsRegistry(reg)

	got := act.MetricsRegistry()
	if got == nil {
		t.Fatal("MetricsRegistry() returned nil")
	}
	collected := got.Collect()
	if len(collected) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(collected))
	}
}

func TestMetricsRegistry(t *testing.T) {
	container := core.New()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	reg := act.MetricsRegistry()
	if reg == nil {
		t.Fatal("MetricsRegistry() returned nil")
	}
}

func TestDatabaseHealthIndicator(t *testing.T) {
	t.Run("nil check function returns unknown", func(t *testing.T) {
		ind := NewDatabaseHealthIndicator(nil)
		h := ind.Health(context.Background())
		if h.Status != health.StatusUnknown {
			t.Fatalf("expected UNKNOWN, got %s", h.Status)
		}
	})

	t.Run("successful check returns up", func(t *testing.T) {
		ind := NewDatabaseHealthIndicator(func(ctx context.Context) error {
			return nil
		})
		h := ind.Health(context.Background())
		if h.Status != health.StatusUp {
			t.Fatalf("expected UP, got %s", h.Status)
		}
	})

	t.Run("failed check returns down", func(t *testing.T) {
		ind := NewDatabaseHealthIndicator(func(ctx context.Context) error {
			return assertAnError("connection refused")
		})
		h := ind.Health(context.Background())
		if h.Status != health.StatusDown {
			t.Fatalf("expected DOWN, got %s", h.Status)
		}
		if h.Details == nil {
			t.Fatal("expected details")
		}
	})
}

func TestRedisHealthIndicator(t *testing.T) {
	t.Run("nil check function returns unknown", func(t *testing.T) {
		ind := NewRedisHealthIndicator(nil)
		h := ind.Health(context.Background())
		if h.Status != health.StatusUnknown {
			t.Fatalf("expected UNKNOWN, got %s", h.Status)
		}
	})

	t.Run("successful check returns up", func(t *testing.T) {
		ind := NewRedisHealthIndicator(func(ctx context.Context) error {
			return nil
		})
		h := ind.Health(context.Background())
		if h.Status != health.StatusUp {
			t.Fatalf("expected UP, got %s", h.Status)
		}
	})

	t.Run("failed check returns down", func(t *testing.T) {
		ind := NewRedisHealthIndicator(func(ctx context.Context) error {
			return assertAnError("timeout")
		})
		h := ind.Health(context.Background())
		if h.Status != health.StatusDown {
			t.Fatalf("expected DOWN, got %s", h.Status)
		}
		if h.Details == nil {
			t.Fatal("expected details")
		}
	})
}

func TestDatabaseHealthIndicator_Name(t *testing.T) {
	ind := NewDatabaseHealthIndicator(nil)
	if ind.Name() != "database" {
		t.Fatalf("expected 'database', got %s", ind.Name())
	}
}

func TestRedisHealthIndicator_Name(t *testing.T) {
	ind := NewRedisHealthIndicator(nil)
	if ind.Name() != "redis" {
		t.Fatalf("expected 'redis', got %s", ind.Name())
	}
}

// assertAnError 返回一个固定的 error，用于测试
type assertAnError string

func (e assertAnError) Error() string { return string(e) }

// TestInfoHandler 测试应用信息处理器
func TestInfoHandler(t *testing.T) {
	container := core.New()
	env := newTestEnvironment()
	env.AddPropertySource(environment.NewDefaultPropertySource("test", map[string]any{
		"app.name":    "test-app",
		"app.version": "2.0.0",
		"build.time":  "2024-01-01T00:00:00Z",
	}))
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/info", nil)
	act.InfoHandler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result map[string]any
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	app, ok := result["app"].(map[string]any)
	if !ok {
		t.Fatal("expected app object")
	}
	if app["name"] != "test-app" {
		t.Fatalf("expected app.name 'test-app', got %v", app["name"])
	}
	if app["version"] != "2.0.0" {
		t.Fatalf("expected app.version '2.0.0', got %v", app["version"])
	}
}

// TestInfoHandler_NoAppContext 测试应用上下文为 nil 时的错误处理
func TestInfoHandler_NoAppContext(t *testing.T) {
	act := &Actuator{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/info", nil)
	act.InfoHandler(w, r)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// TestPrometheusHandler 测试 Prometheus 格式指标处理器
func TestPrometheusHandler(t *testing.T) {
	container := core.New()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	reg := metrics.NewSimpleRegistry()
	reg.Counter("requests_total", "method", "GET").Inc()
	reg.Counter("requests_total", "method", "POST").Add(2)
	reg.Gauge("memory_usage").Set(1024.5)
	act.SetMetricsRegistry(reg)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	act.PrometheusHandler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/plain; version=0.0.4; charset=utf-8" {
		t.Fatalf("expected Content-Type 'text/plain; version=0.0.4; charset=utf-8', got %s", contentType)
	}

	body := w.Body.String()
	if body == "" {
		t.Fatal("expected non-empty response body")
	}

	// 验证包含指标名称
	if !contains(body, "requests_total") {
		t.Fatal("expected 'requests_total' in response")
	}
	if !contains(body, "memory_usage") {
		t.Fatal("expected 'memory_usage' in response")
	}
}

// TestPrometheusHandler_NilRegistry 测试指标注册表为 nil 时的错误处理
func TestPrometheusHandler_NilRegistry(t *testing.T) {
	act := &Actuator{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	act.PrometheusHandler(w, r)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// TestPprofHandlers 测试 pprof 端点处理器
func TestPprofHandlers(t *testing.T) {
	container := core.New()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	handlers := act.PprofHandlers()
	if handlers == nil {
		t.Fatal("PprofHandlers() returned nil")
	}

	expectedPaths := []string{
		"/debug/pprof/",
		"/debug/pprof/cmdline",
		"/debug/pprof/profile",
		"/debug/pprof/symbol",
		"/debug/pprof/trace",
	}

	for _, path := range expectedPaths {
		if _, ok := handlers[path]; !ok {
			t.Fatalf("expected path %s not found in handlers", path)
		}
	}

	if len(handlers) != len(expectedPaths) {
		t.Fatalf("expected %d handlers, got %d", len(expectedPaths), len(handlers))
	}
}

// TestRegisterDebugRoutes 测试注册调试路由
func TestRegisterDebugRoutes(t *testing.T) {
	container := core.New()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	mux := http.NewServeMux()
	act.RegisterDebugRoutes(&StdRouteRegistrar{Mux: mux})

	expectedPaths := []string{
		"/debug/pprof/",
		"/debug/pprof/cmdline",
		"/debug/pprof/symbol",
	}

	for _, path := range expectedPaths {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, path, nil)
		mux.ServeHTTP(w, r)

		if w.Code == http.StatusNotFound {
			t.Fatalf("path %s returned 404, expected handler to be registered", path)
		}
	}

	t.Run("profile endpoint registered", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/debug/pprof/profile?seconds=0.1", nil)
		mux.ServeHTTP(w, r)

		if w.Code == http.StatusNotFound {
			t.Fatal("profile endpoint should be registered")
		}
	})

	t.Run("trace endpoint registered", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/debug/pprof/trace", nil)
		mux.ServeHTTP(w, r)

		if w.Code == http.StatusNotFound {
			t.Fatal("trace endpoint should be registered")
		}
	})
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
