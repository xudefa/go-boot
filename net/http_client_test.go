package net

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	c := NewClient("http://localhost:8080")
	if c == nil {
		t.Fatal("NewClient() returned nil")
	}
	if c.baseURL != "http://localhost:8080" {
		t.Errorf("baseURL = %v, want http://localhost:8080", c.baseURL)
	}
}

func TestNewClientWithOptions(t *testing.T) {
	headers := make(http.Header)
	headers.Set("Authorization", "Bearer test-token")

	c := NewClient("http://api.example.com",
		WithClientTimeout(10*time.Second),
		WithHeaders(headers),
	)
	if c == nil {
		t.Fatal("NewClient() returned nil")
	}
	if c.baseURL != "http://api.example.com" {
		t.Errorf("baseURL = %v, want http://api.example.com", c.baseURL)
	}
	if c.HttpClient.Timeout != 10*time.Second {
		t.Errorf("Timeout = %v, want 10s", c.HttpClient.Timeout)
	}
	if c.headers.Get("Authorization") != "Bearer test-token" {
		t.Errorf("Authorization header = %v", c.headers.Get("Authorization"))
	}
}

func TestClientBuildURL(t *testing.T) {
	c := NewClient("http://localhost:8080")

	tests := []struct {
		path     string
		query    url.Values
		expected string
	}{
		{"/api/users", nil, "http://localhost:8080/api/users"},
		{"api/users", nil, "http://localhost:8080/api/users"},
		{"http://other.com/api", nil, "http://other.com/api"},
		{"/api/search", url.Values{"q": []string{"test"}}, "http://localhost:8080/api/search?q=test"},
	}

	for _, tt := range tests {
		got := c.buildURL(tt.path, tt.query)
		if got != tt.expected {
			t.Errorf("buildURL(%q) = %v, want %v", tt.path, got, tt.expected)
		}
	}
}

func TestClientWithMiddleware(t *testing.T) {
	c := NewClient("http://localhost:8080")
	called := false
	c.WithMiddleware(func(req *http.Request, resp *HttpResponse) error {
		called = true
		return nil
	})
	if len(c.middleware) != 1 {
		t.Errorf("middleware count = %v, want 1", len(c.middleware))
	}
	_ = called
}

func TestClientClose(t *testing.T) {
	c := NewClient("http://localhost:8080")
	if err := c.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestResponseBind(t *testing.T) {
	resp := &HttpResponse{
		StatusCode: 200,
		Body:       []byte(`{"name": "test", "value": 123}`),
	}

	var result struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	if err := resp.Bind(&result); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	if result.Name != "test" {
		t.Errorf("Name = %v, want test", result.Name)
	}
	if result.Value != 123 {
		t.Errorf("Value = %v, want 123", result.Value)
	}
}

func TestResponseBindEmpty(t *testing.T) {
	resp := &HttpResponse{
		StatusCode: 200,
		Body:       nil,
	}

	var result struct {
		Name string `json:"name"`
	}
	if err := resp.Bind(&result); err != nil {
		t.Errorf("Bind() error on empty body: %v", err)
	}
}

func TestResponseString(t *testing.T) {
	resp := &HttpResponse{
		StatusCode: 200,
		Body:       []byte("hello world"),
	}
	if resp.String() != "hello world" {
		t.Errorf("String() = %v, want 'hello world'", resp.String())
	}
}
func newCfg(opts ...RequestOption) *HttpRequest {
	cfg := &HttpRequest{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

func TestBuildRequest(t *testing.T) {
	c := NewClient("http://localhost:8080")

	tests := []struct {
		name    string
		method  string
		path    string
		body    any
		wantErr bool
	}{
		{"get request", "GET", "/api/users", nil, false},
		{"post json", "POST", "/api/users", map[string]string{"name": "test"}, false},
		{"post string", "POST", "/api/data", "raw string", false},
		{"post bytes", "POST", "/api/data", []byte("binary"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := c.buildRequest(context.Background(), tt.method, tt.path, tt.body, newCfg())
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("buildRequest() error = %v", err)
			}
			if req.Method != tt.method {
				t.Errorf("method = %v, want %v", req.Method, tt.method)
			}
		})
	}
}

func TestBuildRequestWithOptions(t *testing.T) {
	c := NewClient("http://localhost:8080")

	cfg := newCfg(
		WithHeader("X-Custom", "value"),
		WithQuery("page", "1"),
		WithTimeout(5*time.Second),
		WithAuthToken("test-token"),
	)

	req, err := c.buildRequest(
		context.Background(),
		"GET",
		"/api/users",
		nil,
		cfg,
	)
	if err != nil {
		t.Fatalf("buildRequest() error = %v", err)
	}
	if req.Header.Get("X-Custom") != "value" {
		t.Errorf("X-Custom header = %v", req.Header.Get("X-Custom"))
	}
	if req.Header.Get("Authorization") != "Bearer test-token" {
		t.Errorf("Authorization = %v", req.Header.Get("Authorization"))
	}
}

func TestBuildRequestWithBasicAuth(t *testing.T) {
	c := NewClient("http://localhost:8080")

	cfg := newCfg(WithBasicAuth("admin", "secret"))
	req, err := c.buildRequest(
		context.Background(),
		"GET",
		"/api/users",
		nil,
		cfg,
	)
	if err != nil {
		t.Fatalf("buildRequest() error = %v", err)
	}
	username, password, ok := req.BasicAuth()
	if !ok {
		t.Error("BasicAuth not set")
	}
	if username != "admin" || password != "secret" {
		t.Errorf("BasicAuth = %s:%s, want admin:secret", username, password)
	}
}

func TestRequestOptions(t *testing.T) {
	cfg := &HttpRequest{}

	WithHeader("X-Test", "value")(cfg)
	if cfg.Header.Get("X-Test") != "value" {
		t.Errorf("Header = %v", cfg.Header.Get("X-Test"))
	}

	WithQuery("key", "val")(cfg)
	if cfg.Query.Get("key") != "val" {
		t.Errorf("Query = %v", cfg.Query)
	}

	WithTimeout(30 * time.Second)(cfg)
	if cfg.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v", cfg.Timeout)
	}

	WithAuthToken("my-token")(cfg)
	if cfg.AuthToken != "my-token" {
		t.Errorf("AuthToken = %v", cfg.AuthToken)
	}

	WithContentType("application/xml")(cfg)
	if cfg.ContentType != "application/xml" {
		t.Errorf("ContentType = %v", cfg.ContentType)
	}

	WithBasicAuth("user", "pass")(cfg)
	if cfg.BasicAuth.Username != "user" || cfg.BasicAuth.Password != "pass" {
		t.Errorf("BasicAuth = %s:%s", cfg.BasicAuth.Username, cfg.BasicAuth.Password)
	}
}

func TestResponseHelpers(t *testing.T) {
	tests := []struct {
		statusCode int
		success    bool
		redirect   bool
		clientErr  bool
		serverErr  bool
	}{
		{200, true, false, false, false},
		{204, true, false, false, false},
		{301, false, true, false, false},
		{400, false, false, true, false},
		{404, false, false, true, false},
		{500, false, false, false, true},
		{503, false, false, false, true},
	}

	for _, tt := range tests {
		resp := &HttpResponse{StatusCode: tt.statusCode}
		if resp.IsSuccess() != tt.success {
			t.Errorf("IsSuccess() for %d = %v, want %v", tt.statusCode, resp.IsSuccess(), tt.success)
		}
		if resp.IsRedirect() != tt.redirect {
			t.Errorf("IsRedirect() for %d = %v, want %v", tt.statusCode, resp.IsRedirect(), tt.redirect)
		}
		if resp.IsClientError() != tt.clientErr {
			t.Errorf("IsClientError() for %d = %v, want %v", tt.statusCode, resp.IsClientError(), tt.clientErr)
		}
		if resp.IsServerError() != tt.serverErr {
			t.Errorf("IsServerError() for %d = %v, want %v", tt.statusCode, resp.IsServerError(), tt.serverErr)
		}
	}
}

func TestBuildRequestWithFormBody(t *testing.T) {
	c := NewClient("http://localhost:8080")

	form := url.Values{}
	form.Set("key", "value")

	req, err := c.buildRequest(context.Background(), "POST", "/api/form", form, newCfg())
	if err != nil {
		t.Fatalf("buildRequest() error = %v", err)
	}
	if req.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		t.Errorf("Content-Type = %v", req.Header.Get("Content-Type"))
	}
}

func TestClientDoWithServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"method":"GET"}`))
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"method":"POST"}`))
		case http.MethodPut:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"method":"PUT"}`))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodPatch:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"method":"PATCH"}`))
		case http.MethodHead:
			w.WriteHeader(http.StatusOK)
		case http.MethodOptions:
			w.Header().Set("Allow", "GET,POST,PUT,DELETE,PATCH,HEAD,OPTIONS")
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer ts.Close()

	c := NewClient(ts.URL)

	t.Run("GET", func(t *testing.T) {
		resp, err := c.Get(context.Background(), "/")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("POST", func(t *testing.T) {
		resp, err := c.Post(context.Background(), "/", map[string]string{"key": "value"})
		if err != nil {
			t.Fatalf("Post() error = %v", err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusCreated)
		}
	})

	t.Run("PUT", func(t *testing.T) {
		resp, err := c.Put(context.Background(), "/", map[string]string{"key": "value"})
		if err != nil {
			t.Fatalf("Put() error = %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("DELETE", func(t *testing.T) {
		resp, err := c.Delete(context.Background(), "/")
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusNoContent)
		}
	})

	t.Run("PATCH", func(t *testing.T) {
		resp, err := c.Patch(context.Background(), "/", map[string]string{"key": "value"})
		if err != nil {
			t.Fatalf("Patch() error = %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("HEAD", func(t *testing.T) {
		resp, err := c.Head(context.Background(), "/")
		if err != nil {
			t.Fatalf("Head() error = %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("OPTIONS", func(t *testing.T) {
		resp, err := c.Options(context.Background(), "/")
		if err != nil {
			t.Fatalf("Options() error = %v", err)
		}
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusNoContent)
		}
	})
}

func TestClientDoWithCustomRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	req, err := http.NewRequestWithContext(context.Background(), "GET", ts.URL, nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext error = %v", err)
	}

	c := NewClient("")
	resp, err := c.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if !resp.IsSuccess() {
		t.Error("IsSuccess() should be true")
	}
}

func TestClientDoWithHeaderResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom", "test-value")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := NewClient(ts.URL)
	resp, err := c.Get(context.Background(), "/")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if resp.Header.Get("X-Custom") != "test-value" {
		t.Errorf("X-Custom header = %q, want %q", resp.Header.Get("X-Custom"), "test-value")
	}
}

func TestClientDoWithTimeout(t *testing.T) {
	t.Parallel()

	// 创建一个慢服务器的测试：延迟 100ms 返回
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := NewClient(ts.URL)

	t.Run("request timeout triggers cancel", func(t *testing.T) {
		ctx := context.Background()
		_, err := c.Get(ctx, "/", WithTimeout(1*time.Millisecond))
		if err == nil {
			t.Error("expected timeout error")
		}
	})

	t.Run("no timeout succeeds", func(t *testing.T) {
		ctx := context.Background()
		resp, err := c.Get(ctx, "/")
		if err != nil {
			t.Fatalf("Get() without timeout error = %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("client-level timeout still works", func(t *testing.T) {
		c2 := NewClient(ts.URL, WithClientTimeout(1*time.Millisecond))
		_, err := c2.Get(context.Background(), "/")
		if err == nil {
			t.Error("expected timeout error with client-level timeout")
		}
	})
}
