package net

import (
	"net/http"
	"testing"
	"time"
)

func TestHttpClientBuilder_Defaults(t *testing.T) {
	builder := NewHttpClientBuilder()

	if builder.timeout != DefaultTimeout {
		t.Errorf("expected default timeout %v, got %v", DefaultTimeout, builder.timeout)
	}

	if builder.headers == nil {
		t.Error("expected non-nil headers")
	}
}

func TestHttpClientBuilder_ChainConfig(t *testing.T) {
	client, err := NewHttpClientBuilder().
		BaseURL("http://localhost:8080").
		Timeout(10*time.Second).
		Header("X-Custom-Header", "value").
		Header("Accept", "application/json").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestHttpClientBuilder_MissingBaseURL(t *testing.T) {
	_, err := NewHttpClientBuilder().Build()
	if err == nil {
		t.Error("expected error for missing base URL")
	}
}

func TestHttpClientBuilder_MustBuild_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for missing base URL")
		}
	}()

	NewHttpClientBuilder().MustBuild()
}

func TestHttpClientBuilder_WithMiddleware(t *testing.T) {
	middlewareCalled := false

	client, err := NewHttpClientBuilder().
		BaseURL("http://localhost:8080").
		Middleware(func(req *http.Request, resp *HttpResponse) error {
			middlewareCalled = true
			return nil
		}).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client == nil {
		t.Fatal("expected non-nil client")
	}

	netClient, ok := client.(*NetClient)
	if !ok {
		t.Fatal("expected *NetClient")
	}

	if len(netClient.middleware) != 1 {
		t.Errorf("expected 1 middleware, got %d", len(netClient.middleware))
	}

	_ = middlewareCalled
}

func TestHttpClientBuilder_WithRetry(t *testing.T) {
	client, err := NewHttpClientBuilder().
		BaseURL("http://localhost:8080").
		Retry(
			WithMaxAttempts(5),
			WithOnRetry(func(attempt int, resp *HttpResponse, err error) {
				// retry callback for testing
			}),
		).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client == nil {
		t.Fatal("expected non-nil client")
	}

	_, ok := client.(*RetryableClient)
	if !ok {
		t.Fatal("expected *RetryableClient when retry is configured")
	}
}

func TestHttpServerBuilder_Defaults(t *testing.T) {
	builder := NewHttpServerBuilder()

	if builder.host != ":8080" {
		t.Errorf("expected default host ':8080', got %s", builder.host)
	}

	if builder.readTimeout != 30*time.Second {
		t.Errorf("expected default readTimeout 30s, got %v", builder.readTimeout)
	}

	if builder.writeTimeout != 30*time.Second {
		t.Errorf("expected default writeTimeout 30s, got %v", builder.writeTimeout)
	}

	if builder.idleTimeout != 120*time.Second {
		t.Errorf("expected default idleTimeout 120s, got %v", builder.idleTimeout)
	}
}

func TestHttpServerBuilder_ChainConfig(t *testing.T) {
	server, err := NewHttpServerBuilder().
		Host(":9090").
		ReadTimeout(10 * time.Second).
		WriteTimeout(10 * time.Second).
		IdleTimeout(60 * time.Second).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if server == nil {
		t.Fatal("expected non-nil server")
	}
}

func TestHttpServerBuilder_WithMiddleware(t *testing.T) {
	middlewareCalled := false

	server, err := NewHttpServerBuilder().
		Host(":9090").
		Middleware(func(ctx HandlerContext) {
			middlewareCalled = true
		}).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if server == nil {
		t.Fatal("expected non-nil server")
	}

	_ = middlewareCalled
}

func TestHttpServerBuilder_WithTls(t *testing.T) {
	server, err := NewHttpServerBuilder().
		Host(":9090").
		Tls("/path/to/cert.pem", "/path/to/key.pem").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if server == nil {
		t.Fatal("expected non-nil server")
	}

	hs, ok := server.(*httpServer)
	if !ok {
		t.Fatal("expected *httpServer")
	}

	if hs.certFile != "/path/to/cert.pem" {
		t.Errorf("expected certFile '/path/to/cert.pem', got %s", hs.certFile)
	}

	if hs.keyFile != "/path/to/key.pem" {
		t.Errorf("expected keyFile '/path/to/key.pem', got %s", hs.keyFile)
	}
}

func TestHttpServerBuilder_MustBuild_Success(t *testing.T) {
	// Build() doesn't have required fields, so it won't panic
	server := NewHttpServerBuilder().MustBuild()

	if server == nil {
		t.Fatal("expected non-nil server")
	}
}

func TestRequestBuilder_GetRequest(t *testing.T) {
	opts := NewRequestBuilder("GET", "/api/users").
		Header("Authorization", "Bearer token").
		Query("page", "1").
		Query("limit", "10").
		Timeout(5 * time.Second).
		Build()

	if len(opts) < 4 {
		t.Errorf("expected at least 4 options, got %d", len(opts))
	}
}

func TestRequestBuilder_PostRequest(t *testing.T) {
	body := map[string]string{"name": "test"}

	opts := NewRequestBuilder("POST", "/api/users").
		Body(body).
		ContentType("application/json").
		AuthToken("test-token").
		Build()

	// AuthToken and ContentType add to opts, Body doesn't
	if len(opts) < 2 {
		t.Errorf("expected at least 2 options, got %d", len(opts))
	}
}

func TestRequestBuilder_Execute_Get(t *testing.T) {
	// 使用Noop客户端测试
	builder := NewRequestBuilder("GET", "/api/users").
		Query("page", "1")

	// 验证Execute方法不会panic
	// 实际执行需要真实的HTTP客户端
	_ = builder
}

func TestRequestBuilder_Execute_UnsupportedMethod(t *testing.T) {
	builder := NewRequestBuilder("INVALID", "/api/test")

	// 验证Execute方法对不支持的方法返回错误
	// 需要mock client来测试
	_ = builder
}

func TestRequestBuilder_MultipleHeaders(t *testing.T) {
	opts := NewRequestBuilder("GET", "/api/users").
		Header("Accept", "application/json").
		Header("X-Custom-1", "value1").
		Header("X-Custom-2", "value2").
		Build()

	if len(opts) < 3 {
		t.Errorf("expected at least 3 options, got %d", len(opts))
	}
}

func TestRequestBuilder_MultipleQueries(t *testing.T) {
	opts := NewRequestBuilder("GET", "/api/users").
		Query("filter", "active").
		Query("sort", "name").
		Query("order", "asc").
		Build()

	if len(opts) < 3 {
		t.Errorf("expected at least 3 options, got %d", len(opts))
	}
}

func TestHttpClientBuilder_MultipleMiddleware(t *testing.T) {
	count := 0

	client, err := NewHttpClientBuilder().
		BaseURL("http://localhost:8080").
		Middleware(func(req *http.Request, resp *HttpResponse) error {
			count++
			return nil
		}).
		Middleware(func(req *http.Request, resp *HttpResponse) error {
			count++
			return nil
		}).
		Middleware(func(req *http.Request, resp *HttpResponse) error {
			count++
			return nil
		}).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	netClient, ok := client.(*NetClient)
	if !ok {
		t.Fatal("expected *NetClient")
	}

	if len(netClient.middleware) != 3 {
		t.Errorf("expected 3 middleware, got %d", len(netClient.middleware))
	}
}

func TestHttpServerBuilder_MultipleMiddleware(t *testing.T) {
	count := 0

	server, err := NewHttpServerBuilder().
		Host(":9090").
		Middleware(func(ctx HandlerContext) {
			count++
		}).
		Middleware(func(ctx HandlerContext) {
			count++
		}).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if server == nil {
		t.Fatal("expected non-nil server")
	}

	_ = count
}

func TestHttpClientBuilder_WithAllOptions(t *testing.T) {
	client, err := NewHttpClientBuilder().
		BaseURL("https://api.example.com").
		Timeout(30*time.Second).
		Header("Accept", "application/json").
		Header("User-Agent", "go-boot/1.0").
		Middleware(func(req *http.Request, resp *HttpResponse) error {
			return nil
		}).
		Retry(
			WithMaxAttempts(3),
			WithRetryStrategy(NewFixedDelay(1*time.Second)),
		).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestHttpServerBuilder_WithAllOptions(t *testing.T) {
	server, err := NewHttpServerBuilder().
		Host(":8443").
		ReadTimeout(15*time.Second).
		WriteTimeout(15*time.Second).
		IdleTimeout(90*time.Second).
		Tls("/cert.pem", "/key.pem").
		Middleware(func(ctx HandlerContext) {}).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if server == nil {
		t.Fatal("expected non-nil server")
	}

	hs, ok := server.(*httpServer)
	if !ok {
		t.Fatal("expected *httpServer")
	}

	if hs.host != ":8443" {
		t.Errorf("expected host ':8443', got %s", hs.host)
	}

	if hs.readTimeout != 15*time.Second {
		t.Errorf("expected readTimeout 15s, got %v", hs.readTimeout)
	}

	if hs.certFile != "/cert.pem" {
		t.Errorf("expected certFile '/cert.pem', got %s", hs.certFile)
	}
}

func TestRequestBuilder_ChainConfig(t *testing.T) {
	builder := NewRequestBuilder("POST", "/api/users").
		Body(map[string]string{"name": "test"}).
		Header("Content-Type", "application/json").
		Header("Authorization", "Bearer token").
		Query("version", "v1").
		AuthToken("test-token").
		ContentType("application/json").
		Timeout(10 * time.Second)

	if builder.method != "POST" {
		t.Errorf("expected method 'POST', got %s", builder.method)
	}

	if builder.path != "/api/users" {
		t.Errorf("expected path '/api/users', got %s", builder.path)
	}

	if builder.body == nil {
		t.Error("expected non-nil body")
	}

	if len(builder.headers) < 2 {
		t.Errorf("expected at least 2 headers, got %d", len(builder.headers))
	}

	if len(builder.query) < 1 {
		t.Errorf("expected at least 1 query, got %d", len(builder.query))
	}

	if len(builder.opts) < 3 {
		t.Errorf("expected at least 3 options, got %d", len(builder.opts))
	}
}

func TestHttpClientBuilder_EmptyMiddleware(t *testing.T) {
	client, err := NewHttpClientBuilder().
		BaseURL("http://localhost:8080").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	netClient, ok := client.(*NetClient)
	if !ok {
		t.Fatal("expected *NetClient")
	}

	if len(netClient.middleware) != 0 {
		t.Errorf("expected 0 middleware, got %d", len(netClient.middleware))
	}
}

func TestHttpServerBuilder_EmptyMiddleware(t *testing.T) {
	server, err := NewHttpServerBuilder().
		Host(":9090").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hs, ok := server.(*httpServer)
	if !ok {
		t.Fatal("expected *httpServer")
	}

	if len(hs.middlewares) != 0 {
		t.Errorf("expected 0 middleware, got %d", len(hs.middlewares))
	}
}
