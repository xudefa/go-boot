// Package hertz 提供 Hertz 框架的 HTTP 客户端实现。
//
// 该包提供 RESTful 风格的 HTTP 客户端，支持 GET、POST、PUT、DELETE、PATCH 等方法。
// 支持全局默认客户端和独立实例两种使用方式。
package hertz

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

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
	baseURL     string
	hertzClient *client.Client
	opts        []config.ClientOption
}

// MiddlewareFunc 是中间件函数类型。
type MiddlewareFunc func(*protocol.Request, *protocol.Response) error

// Config 用于配置客户端。
type Config struct {
	BaseURL string
	Timeout time.Duration
}

// Response 是 HTTP 响应封装。
type Response struct {
	StatusCode int
	Header     []byte
	Body       []byte
}

// NewClient 创建新的客户端。
//
// 参数:
//   - baseURL: 基础 URL，如 "http://localhost:8080"
//
// 返回配置好的客户端实例。
func NewClient(baseURL string) (*Client, error) {
	c, err := client.NewClient()
	if err != nil {
		return nil, err
	}
	return &Client{
		baseURL:     baseURL,
		hertzClient: c,
	}, nil
}

// NewClientWithConfig 使用配置创建客户端。
func NewClientWithConfig(cfg *Config) (*Client, error) {
	opts := []config.ClientOption{
		client.WithDialTimeout(DefaultTimeout),
	}
	if cfg.Timeout > 0 {
		opts = append(opts, client.WithDialTimeout(cfg.Timeout))
	}

	c, err := client.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		baseURL:     cfg.BaseURL,
		hertzClient: c,
		opts:        opts,
	}, nil
}

// SetDefaultClient 设置全局默认客户端。
func SetDefaultClient(c *Client) {
	defaultClient = c
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

// buildURL 构建完整 URL。
func (c *Client) buildURL(path string) string {
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		return c.baseURL + path
	}
	return path
}

// buildRequest 构建 Hertz 请求。
func (c *Client) buildRequest(method, path string, body interface{}, opts ...ingress.RequestOption) (*protocol.Request, error) {
	req := protocol.AcquireRequest()

	req.SetMethod(method)
	req.SetRequestURI(c.buildURL(path))

	cfg := &ingress.RequestConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	if body != nil {
		switch v := body.(type) {
		case string:
			req.SetBody([]byte(v))
			req.SetHeader("Content-Type", "text/plain")
		case []byte:
			req.SetBody(v)
			req.SetHeader("Content-Type", "application/octet-stream")
		default:
			data, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("marshal body failed: %w", err)
			}
			req.SetBody(data)
			req.SetHeader("Content-Type", "application/json")
		}
	}

	if cfg.AuthToken != "" {
		req.SetHeader("Authorization", "Bearer "+cfg.AuthToken)
	}

	if cfg.BasicAuth != nil {
		username, password := cfg.BasicAuth()
		req.SetBasicAuth(username, password)
	}

	for key, values := range cfg.Header {
		for _, value := range values {
			req.SetHeader(key, value)
		}
	}

	return req, nil
}

// do 执行 HTTP 请求并返回响应。
func (c *Client) do(ctx context.Context, req *protocol.Request) (*Response, error) {
	resp := protocol.AcquireResponse()

	err := c.hertzClient.Do(ctx, req, resp)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode(),
		Body:       resp.Body(),
	}, nil
}

// Get 发送 GET 请求。
func (c *Client) Get(path string, opts ...ingress.RequestOption) (*Response, error) {
	return c.Do(consts.MethodGet, path, nil, opts...)
}

// Post 发送 POST 请求。
func (c *Client) Post(path string, body interface{}, opts ...ingress.RequestOption) (*Response, error) {
	return c.Do(consts.MethodPost, path, body, opts...)
}

// Put 发送 PUT 请求。
func (c *Client) Put(path string, body interface{}, opts ...ingress.RequestOption) (*Response, error) {
	return c.Do(consts.MethodPut, path, body, opts...)
}

// Delete 发送 DELETE 请求。
func (c *Client) Delete(path string, opts ...ingress.RequestOption) (*Response, error) {
	return c.Do(consts.MethodDelete, path, nil, opts...)
}

// Patch 发送 PATCH 请求。
func (c *Client) Patch(path string, body interface{}, opts ...ingress.RequestOption) (*Response, error) {
	return c.Do(consts.MethodPatch, path, body, opts...)
}

// Head 发送 HEAD 请求。
func (c *Client) Head(path string, opts ...ingress.RequestOption) (*Response, error) {
	return c.Do(consts.MethodHead, path, nil, opts...)
}

// Options 发送 OPTIONS 请求。
func (c *Client) Options(path string, opts ...ingress.RequestOption) (*Response, error) {
	return c.Do(consts.MethodOptions, path, nil, opts...)
}

// Do 发送自定义 HTTP 请求。
func (c *Client) Do(method, path string, body interface{}, opts ...ingress.RequestOption) (*Response, error) {
	ctx := context.Background()
	req, err := c.buildRequest(method, path, body, opts...)
	if err != nil {
		return nil, err
	}

	return c.do(ctx, req)
}

// DoWithContext 发送带上下文的 HTTP 请求。
func (c *Client) DoWithContext(ctx context.Context, method, path string, body interface{}, opts ...ingress.RequestOption) (*Response, error) {
	req, err := c.buildRequest(method, path, body, opts...)
	if err != nil {
		return nil, err
	}

	return c.do(ctx, req)
}

// Close 关闭客户端。
func (c *Client) Close() error {
	c.hertzClient.CloseIdleConnections()
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
		var err error
		defaultClient, err = NewClient(defaultBaseURL)
		if err != nil {
			return nil, err
		}
	}
	return defaultClient.Get(path, opts...)
}

// Post 全局 POST 请求。
func Post(path string, body interface{}, opts ...ingress.RequestOption) (*Response, error) {
	if defaultClient == nil {
		var err error
		defaultClient, err = NewClient(defaultBaseURL)
		if err != nil {
			return nil, err
		}
	}
	return defaultClient.Post(path, body, opts...)
}

// Put 全局 PUT 请求。
func Put(path string, body interface{}, opts ...ingress.RequestOption) (*Response, error) {
	if defaultClient == nil {
		var err error
		defaultClient, err = NewClient(defaultBaseURL)
		if err != nil {
			return nil, err
		}
	}
	return defaultClient.Put(path, body, opts...)
}

// Delete 全局 DELETE 请求。
func Delete(path string, opts ...ingress.RequestOption) (*Response, error) {
	if defaultClient == nil {
		var err error
		defaultClient, err = NewClient(defaultBaseURL)
		if err != nil {
			return nil, err
		}
	}
	return defaultClient.Delete(path, opts...)
}

// Patch 全局 PATCH 请求。
func Patch(path string, body interface{}, opts ...ingress.RequestOption) (*Response, error) {
	if defaultClient == nil {
		var err error
		defaultClient, err = NewClient(defaultBaseURL)
		if err != nil {
			return nil, err
		}
	}
	return defaultClient.Patch(path, body, opts...)
}

// Do 全局自定义 HTTP 请求。
func Do(method, path string, body interface{}, opts ...ingress.RequestOption) (*Response, error) {
	if defaultClient == nil {
		var err error
		defaultClient, err = NewClient(defaultBaseURL)
		if err != nil {
			return nil, err
		}
	}
	return defaultClient.Do(method, path, body, opts...)
}
