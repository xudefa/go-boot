// Package gin 提供 Gin 框架的 HTTP 客户端实现。
//
// 该包提供 RESTful 风格的 HTTP 客户端，支持 GET、POST、PUT、DELETE、PATCH 等方法。
// 支持全局默认客户端和独立实例两种使用方式。
package gin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/xudefa/go-boot/ingress"
)

const (
	// DefaultTimeout 默认请求超时时间。
	DefaultTimeout = 30 * time.Second
)

var (
	// defaultClient 全局默认客户端。
	defaultClient *Client

	// defaultBaseURL 全局默认基础 URL。
	defaultBaseURL = "http://localhost:8080"
)

// Client 是 HTTP 客户端。
//
// 提供 RESTful 请求方法和请求选项配置。
type Client struct {
	baseURL    string
	httpClient *http.Client
	headers    http.Header
	middleware []MiddlewareFunc
}

// MiddlewareFunc 是中间件函数类型。
type MiddlewareFunc func(*http.Request, *Response) error

// Config 用于配置客户端。
type Config struct {
	BaseURL string
	Timeout time.Duration
	Headers http.Header
}

// Response 是 HTTP 响应封装。
type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// NewClient 创建新的客户端。
//
// 参数:
//   - baseURL: 基础 URL，如 "http://localhost:8080"
//
// 返回配置好的客户端实例。
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		headers: make(http.Header),
	}
}

// NewClientWithConfig 使用配置创建客户端。
func NewClientWithConfig(cfg *Config) *Client {
	client := &Client{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		headers: make(http.Header),
	}

	if cfg.Timeout > 0 {
		client.httpClient.Timeout = cfg.Timeout
	}

	if cfg.Headers != nil {
		client.headers = cfg.Headers
	}

	return client
}

// SetDefaultClient 设置全局默认客户端。
func SetDefaultClient(client *Client) {
	defaultClient = client
}

// DefaultClient 获取全局默认客户端。
//
// 如果未设置全局客户端，则返回 nil。
func DefaultClient() *Client {
	return defaultClient
}

// SetBaseURL 设置全局默认基础 URL。
func SetBaseURL(baseURL string) {
	defaultBaseURL = baseURL
}

// GetBaseURL 获取全局默认基础 URL。
func GetBaseURL() string {
	return defaultBaseURL
}

// SetHeader 设置全局请求头。
func SetHeader(key, value string) {
	if defaultClient == nil {
		defaultClient = NewClient(defaultBaseURL)
	}
	defaultClient.headers.Set(key, value)
}

// WithMiddleware 添加中间件。
func (c *Client) WithMiddleware(m MiddlewareFunc) *Client {
	c.middleware = append(c.middleware, m)
	return c
}

// buildURL 构建完整 URL。
func (c *Client) buildURL(path string, query url.Values) string {
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		path = c.baseURL + path
	}

	if len(query) > 0 {
		path = path + "?" + query.Encode()
	}

	return path
}

// buildRequest 构建 HTTP 请求。
func (c *Client) buildRequest(ctx context.Context, method, path string, body interface{}, opts ...ingress.RequestOption) (*http.Request, error) {
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

	req, err := http.NewRequestWithContext(ctx, method, c.buildURL(path, nil), reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	for key, values := range c.headers {
		for _, value := range values {
			req.Header.Set(key, value)
		}
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	cfg := &ingress.RequestConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	for key, values := range cfg.Header {
		for _, value := range values {
			req.Header.Set(key, value)
		}
	}

	if cfg.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.AuthToken)
	}

	if cfg.BasicAuth != nil {
		username, password := cfg.BasicAuth()
		req.SetBasicAuth(username, password)
	}

	return req, nil
}

// do 执行 HTTP 请求并返回响应。
func (c *Client) do(ctx context.Context, req *http.Request) (*Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	httpResp := &Response{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       body,
	}

	for _, m := range c.middleware {
		if err := m(req, httpResp); err != nil {
			return nil, err
		}
	}

	return httpResp, nil
}

// Get 发送 GET 请求。
func (c *Client) Get(path string, opts ...ingress.RequestOption) (*Response, error) {
	return c.Do("GET", path, nil, opts...)
}

// Post 发送 POST 请求。
func (c *Client) Post(path string, body interface{}, opts ...ingress.RequestOption) (*Response, error) {
	return c.Do("POST", path, body, opts...)
}

// Put 发送 PUT 请求。
func (c *Client) Put(path string, body interface{}, opts ...ingress.RequestOption) (*Response, error) {
	return c.Do("PUT", path, body, opts...)
}

// Delete 发送 DELETE 请求。
func (c *Client) Delete(path string, opts ...ingress.RequestOption) (*Response, error) {
	return c.Do("DELETE", path, nil, opts...)
}

// Patch 发送 PATCH 请求。
func (c *Client) Patch(path string, body interface{}, opts ...ingress.RequestOption) (*Response, error) {
	return c.Do("PATCH", path, body, opts...)
}

// Head 发送 HEAD 请求。
func (c *Client) Head(path string, opts ...ingress.RequestOption) (*Response, error) {
	return c.Do("HEAD", path, nil, opts...)
}

// Options 发送 OPTIONS 请求。
func (c *Client) Options(path string, opts ...ingress.RequestOption) (*Response, error) {
	return c.Do("OPTIONS", path, nil, opts...)
}

// Do 发送自定义 HTTP 请求。
func (c *Client) Do(method, path string, body interface{}, opts ...ingress.RequestOption) (*Response, error) {
	ctx := context.Background()
	req, err := c.buildRequest(ctx, method, path, body, opts...)
	if err != nil {
		return nil, err
	}

	return c.do(ctx, req)
}

// DoWithContext 发送带上下文的 HTTP 请求。
func (c *Client) DoWithContext(ctx context.Context, method, path string, body interface{}, opts ...ingress.RequestOption) (*Response, error) {
	req, err := c.buildRequest(ctx, method, path, body, opts...)
	if err != nil {
		return nil, err
	}

	return c.do(ctx, req)
}

// Close 关闭客户端。
func (c *Client) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}

// Bind 绑定响应体到目标结构体。
func (r *Response) Bind(v interface{}) error {
	if len(r.Body) == 0 {
		return nil
	}
	return json.Unmarshal(r.Body, v)
}

// String 获取响应体字符串。
func (r *Response) String() string {
	return string(r.Body)
}

// 全局默认客户端方法

// Get 全局 GET 请求。
func Get(path string, opts ...ingress.RequestOption) (*Response, error) {
	if defaultClient == nil {
		defaultClient = NewClient(defaultBaseURL)
	}
	return defaultClient.Get(path, opts...)
}

// Post 全局 POST 请求。
func Post(path string, body interface{}, opts ...ingress.RequestOption) (*Response, error) {
	if defaultClient == nil {
		defaultClient = NewClient(defaultBaseURL)
	}
	return defaultClient.Post(path, body, opts...)
}

// Put 全局 PUT 请求。
func Put(path string, body interface{}, opts ...ingress.RequestOption) (*Response, error) {
	if defaultClient == nil {
		defaultClient = NewClient(defaultBaseURL)
	}
	return defaultClient.Put(path, body, opts...)
}

// Delete 全局 DELETE 请求。
func Delete(path string, opts ...ingress.RequestOption) (*Response, error) {
	if defaultClient == nil {
		defaultClient = NewClient(defaultBaseURL)
	}
	return defaultClient.Delete(path, opts...)
}

// Patch 全局 PATCH 请求。
func Patch(path string, body interface{}, opts ...ingress.RequestOption) (*Response, error) {
	if defaultClient == nil {
		defaultClient = NewClient(defaultBaseURL)
	}
	return defaultClient.Patch(path, body, opts...)
}

// Do 全局自定义 HTTP 请求。
func Do(method, path string, body interface{}, opts ...ingress.RequestOption) (*Response, error) {
	if defaultClient == nil {
		defaultClient = NewClient(defaultBaseURL)
	}
	return defaultClient.Do(method, path, body, opts...)
}
