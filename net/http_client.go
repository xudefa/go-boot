// Package net 提供 HTTP 服务器/客户端的抽象接口和默认实现。
//
// 核心接口：
//   - Server: HTTP 服务器抽象接口
//   - HttpClient: RESTful 客户端接口，支持 GET、POST、PUT、DELETE、PATCH 等方法
//   - ClientMiddlewareFunc: 客户端中间件
//
// NetClient 是默认实现，支持连接池、超时、认证、中间件链等特性。
//
// 使用示例：
//
//	client := net.NewClient("http://localhost:8080")
//	resp, err := client.Get(ctx, "/api/users")
//	status := resp.StatusCode
package net

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/xudefa/go-boot/log"
)

const (
	// DefaultTimeout 默认请求超时时间。
	DefaultTimeout = 30 * time.Second
)

// HttpClient 是 HTTP 客户端的统一接口。
//
// 提供 RESTful 请求方法，支持连接池管理。
type HttpClient interface {
	// Get 发送 GET 请求。
	Get(ctx context.Context, url string, opts ...RequestOption) (*HttpResponse, error)

	// Head 发送 HEAD 请求。
	Head(ctx context.Context, url string, opts ...RequestOption) (*HttpResponse, error)

	// Post 发送 POST 请求。
	Post(ctx context.Context, url string, body any, opts ...RequestOption) (*HttpResponse, error)

	// Put 发送 PUT 请求。
	Put(ctx context.Context, url string, body any, opts ...RequestOption) (*HttpResponse, error)

	// Patch 发送 PATCH 请求。
	Patch(ctx context.Context, url string, body any, opts ...RequestOption) (*HttpResponse, error)

	// Delete 发送 DELETE 请求。
	Delete(ctx context.Context, url string, opts ...RequestOption) (*HttpResponse, error)

	// Options 发送 OPTIONS 请求。
	Options(ctx context.Context, url string, opts ...RequestOption) (*HttpResponse, error)

	// Do 发送自定义 HTTP 请求。
	Do(ctx context.Context, req any) (*HttpResponse, error)

	// Close 关闭客户端并释放资源。
	Close() error
}

// NetClient 是 HTTP 客户端。
//
// 提供 RESTful 请求方法和请求选项配置。
type NetClient struct {
	mu         sync.RWMutex           // 读写锁
	baseURL    string                 // 基础 URL
	HttpClient *http.Client           // 底层 HTTP 客户端
	headers    http.Header            // 默认请求头
	middleware []ClientMiddlewareFunc // 中间件列表
}

// RequestOption 是请求选项配置函数。
type RequestOption func(*HttpRequest)

// BasicAuth 表示 HTTP 基本认证凭据。
type BasicAuth struct {
	Username string // 用户名
	Password string // 密码
}

// HttpRequest 是 HTTP 请求封装。
type HttpRequest struct {
	Header      http.Header   // 请求头
	Query       url.Values    // 查询参数
	Timeout     time.Duration // 请求超时
	AuthToken   string        // Bearer 认证令牌
	ContentType string        // 请求 Content-Type
	BasicAuth   BasicAuth     // 基本认证凭据
}

// HttpResponse 是 HTTP 响应封装。
type HttpResponse struct {
	StatusCode int         // HTTP 状态码
	Header     http.Header // 响应头
	Body       []byte      // 响应体
}

// ClientMiddlewareFunc 是 HTTP 客户端中间件函数类型。
type ClientMiddlewareFunc func(*http.Request, *HttpResponse) error

// ClientOption 定义客户端配置选项
type ClientOption func(*NetClient)

// WithClientTimeout 设置客户端请求超时时间
func WithClientTimeout(timeout time.Duration) ClientOption {
	return func(c *NetClient) {
		c.HttpClient.Timeout = timeout
	}
}

// WithHeaders 设置默认请求头
func WithHeaders(headers http.Header) ClientOption {
	return func(c *NetClient) {
		c.headers = headers.Clone()
	}
}

// WithHeader 设置请求头。
func WithHeader(key, value string) RequestOption {
	return func(c *HttpRequest) {
		if c.Header == nil {
			c.Header = make(http.Header)
		}
		c.Header.Set(key, value)
	}
}

// WithQuery 设置查询参数。
func WithQuery(key, value string) RequestOption {
	return func(c *HttpRequest) {
		if c.Query == nil {
			c.Query = make(url.Values)
		}
		c.Query.Set(key, value)
	}
}

// WithTimeout 设置请求超时。
func WithTimeout(timeout time.Duration) RequestOption {
	return func(c *HttpRequest) {
		c.Timeout = timeout
	}
}

// WithAuthToken 设置认证令牌。
func WithAuthToken(token string) RequestOption {
	return func(c *HttpRequest) {
		c.AuthToken = token
	}
}

// WithContentType 设置请求 Content-Type。
func WithContentType(contentType string) RequestOption {
	return func(c *HttpRequest) {
		c.ContentType = contentType
	}
}

// WithBasicAuth 设置基本认证。
func WithBasicAuth(username, password string) RequestOption {
	return func(c *HttpRequest) {
		c.BasicAuth = BasicAuth{Username: username, Password: password}
	}
}

// IsSuccess 判断响应是否为成功状态码 (2xx)。
func (r *HttpResponse) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsRedirect 判断响应是否为重定向状态码 (3xx)。
func (r *HttpResponse) IsRedirect() bool {
	return r.StatusCode >= 300 && r.StatusCode < 400
}

// IsClientError 判断响应是否为客户端错误状态码 (4xx)。
func (r *HttpResponse) IsClientError() bool {
	return r.StatusCode >= 400 && r.StatusCode < 500
}

// IsServerError 判断响应是否为服务端错误状态码 (5xx)。
func (r *HttpResponse) IsServerError() bool {
	return r.StatusCode >= 500 && r.StatusCode < 600
}

// Bind 绑定响应体到目标结构体。
func (r *HttpResponse) Bind(v any) error {
	if len(r.Body) == 0 {
		return nil
	}
	return json.Unmarshal(r.Body, v)
}

// String 获取响应体字符串。
func (r *HttpResponse) String() string {
	return string(r.Body)
}

// NewClient 创建新的客户端。
//
// 参数:
//   - baseURL: 基础 URL，如 "http://localhost:8080"
//   - opts: 可选配置选项
//
// 返回配置好的客户端实例。
func NewClient(baseURL string, opts ...ClientOption) *NetClient {
	c := &NetClient{
		baseURL: baseURL,
		HttpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		headers: make(http.Header),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithMiddleware 添加中间件。
func (c *NetClient) WithMiddleware(m ClientMiddlewareFunc) *NetClient {
	c.mu.Lock()
	c.middleware = append(c.middleware, m)
	c.mu.Unlock()
	return c
}

// buildURL 构建完整 URL。
func (c *NetClient) buildURL(path string, query url.Values) string {
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		base := strings.TrimSuffix(c.baseURL, "/")
		path = strings.TrimPrefix(path, "/")
		path = base + "/" + path
	}

	if len(query) > 0 {
		separator := "?"
		if strings.Contains(path, "?") {
			separator = "&"
		}
		path = path + separator + query.Encode()
	}

	return path
}

// buildRequest 构建 HTTP 请求。
func (c *NetClient) buildRequest(ctx context.Context, method, path string, body any, cfg *HttpRequest) (*http.Request, error) {
	var reqBody io.Reader
	contentType := "application/json"

	if body != nil {
		switch v := body.(type) {
		case string:
			reqBody = strings.NewReader(v)
			contentType = "text/plain"
		case []byte:
			reqBody = bytes.NewReader(v)
			contentType = "application/octet-stream"
		case url.Values:
			reqBody = strings.NewReader(v.Encode())
			contentType = "application/x-www-form-urlencoded"
		default:
			data, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("marshal body failed: %w", err)
			}
			reqBody = bytes.NewReader(data)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, c.buildURL(path, cfg.Query), reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	c.mu.RLock()
	req.Header = c.headers.Clone()
	c.mu.RUnlock()

	if cfg.ContentType != "" {
		req.Header.Set("Content-Type", cfg.ContentType)
	} else {
		req.Header.Set("Content-Type", contentType)
	}

	for key, values := range cfg.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	if cfg.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.AuthToken)
	}

	if cfg.BasicAuth.Username != "" || cfg.BasicAuth.Password != "" {
		req.SetBasicAuth(cfg.BasicAuth.Username, cfg.BasicAuth.Password)
	}

	return req, nil
}

// Get 发送 GET 请求。
func (c *NetClient) Get(ctx context.Context, path string, opts ...RequestOption) (*HttpResponse, error) {
	return c.do(ctx, "GET", path, nil, opts...)
}

// Head 发送 HEAD 请求。
func (c *NetClient) Head(ctx context.Context, path string, opts ...RequestOption) (*HttpResponse, error) {
	return c.do(ctx, "HEAD", path, nil, opts...)
}

// Post 发送 POST 请求。
func (c *NetClient) Post(ctx context.Context, path string, body any, opts ...RequestOption) (*HttpResponse, error) {
	return c.do(ctx, "POST", path, body, opts...)
}

// Put 发送 PUT 请求。
func (c *NetClient) Put(ctx context.Context, path string, body any, opts ...RequestOption) (*HttpResponse, error) {
	return c.do(ctx, "PUT", path, body, opts...)
}

// Patch 发送 PATCH 请求。
func (c *NetClient) Patch(ctx context.Context, path string, body any, opts ...RequestOption) (*HttpResponse, error) {
	return c.do(ctx, "PATCH", path, body, opts...)
}

// Delete 发送 DELETE 请求。
func (c *NetClient) Delete(ctx context.Context, path string, opts ...RequestOption) (*HttpResponse, error) {
	return c.do(ctx, "DELETE", path, nil, opts...)
}

// Options 发送 OPTIONS 请求。
func (c *NetClient) Options(ctx context.Context, path string, opts ...RequestOption) (*HttpResponse, error) {
	return c.do(ctx, "OPTIONS", path, nil, opts...)
}

// do 执行 HTTP 请求并返回响应。
func (c *NetClient) do(ctx context.Context, method, path string, body any, opts ...RequestOption) (*HttpResponse, error) {
	// 预解析 opts 传递至 buildRequest，避免重复解析
	cfg := &HttpRequest{}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}

	req, err := c.buildRequest(ctx, method, path, body, cfg)
	if err != nil {
		return nil, err
	}
	return c.Do(ctx, req)
}

// Do 执行自定义 HTTP 请求并返回响应。
func (c *NetClient) Do(ctx context.Context, request any) (*HttpResponse, error) {
	if c.HttpClient == nil {
		return nil, fmt.Errorf("HttpClient is nil")
	}
	req, ok := request.(*http.Request)
	if !ok {
		return nil, fmt.Errorf("invalid request type, expected *http.Request")
	}
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.mu.RLock()
			if len(c.middleware) > 0 {
				log.Build().Error(ctx, "close response body failed",
					log.KeyValue{Key: "close_error", Value: closeErr.Error()},
					log.KeyValue{Key: "read_error", Value: err.Error()},
				)
			}
			c.mu.RUnlock()
		}
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	if err := resp.Body.Close(); err != nil {
		return nil, fmt.Errorf("close response body failed: %w", err)
	}

	httpResp := &HttpResponse{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       body,
	}

	c.mu.RLock()
	middleware := make([]ClientMiddlewareFunc, len(c.middleware))
	copy(middleware, c.middleware)
	c.mu.RUnlock()

	for _, m := range middleware {
		if err := m(req, httpResp); err != nil {
			return nil, err
		}
	}

	return httpResp, nil
}

// Close 关闭客户端。
func (c *NetClient) Close() error {
	c.HttpClient.CloseIdleConnections()
	return nil
}
