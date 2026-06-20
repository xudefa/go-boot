package actuator

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStdRouteRegistrar_Handle(t *testing.T) {
	mux := http.NewServeMux()
	registrar := &StdRouteRegistrar{Mux: mux}

	// 注册路由
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	registrar.Handle("/test", handler)

	// 验证路由已注册（通过发送请求测试）
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestDefaultRouteConfig(t *testing.T) {
	config := DefaultRouteConfig()

	if config.BasePath != "/actuator" {
		t.Errorf("Expected BasePath '/actuator', got %s", config.BasePath)
	}
	if config.ExposeDebug != false {
		t.Errorf("Expected ExposeDebug false, got %v", config.ExposeDebug)
	}
	if config.Prefix != "" {
		t.Errorf("Expected Prefix empty, got %s", config.Prefix)
	}
}
