// Package ingress 提供统一的 HTTP 服务器抽象接口。
//
// 该包定义了 HTTP 服务器的标准接口，允许使用不同的 HTTP 框架
// (如 Hertz、Gin) 而无需修改应用代码。
package ingress

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/xudefa/go-boot/core"
)

// Server 是 HTTP 服务器的统一接口。
//
// 所有 HTTP 框架实现都应实现此接口，
// 以便与 go-boot 容器系统集成。
type Server interface {
	// Run 启动 HTTP 服务器并阻塞，直到收到关闭信号。
	Run() error

	// Use 向服务器的中间件链追加一个中间件。
	Use(m any) Server

	// UseGlobal 向服务器的中间件链 prepend 一个中间件。
	UseGlobal(m any) Server

	// Register 在容器中注册一个处理函数。
	Register(fn func(core.Container) error) Server

	// Container 返回 go-boot 容器实例。
	Container() core.Container
}

// HTTPClient 是 HTTP 客户端的统一接口。
//
// 提供 RESTful 请求方法，支持连接池管理。
type HTTPClient interface {
	// Get 发送 GET 请求。
	Get(ctx context.Context, url string, opts ...RequestOption) (*Response, error)

	// Post 发送 POST 请求。
	Post(ctx context.Context, url string, body any, opts ...RequestOption) (*Response, error)

	// Put 发送 PUT 请求。
	Put(ctx context.Context, url string, body any, opts ...RequestOption) (*Response, error)

	// Delete 发送 DELETE 请求。
	Delete(ctx context.Context, url string, opts ...RequestOption) (*Response, error)

	// Do 发送自定义 HTTP 请求。
	Do(ctx context.Context, req *http.Request) (*Response, error)

	// Close 关闭客户端并释放资源。
	Close() error
}

// Response 是 HTTP 响应封装。
type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	BodyReader io.Reader
}

// RequestOption 是请求选项配置函数。
type RequestOption func(*RequestConfig)

// RequestConfig 是请求配置。
type RequestConfig struct {
	Header      http.Header
	QueryParams map[string]string
	Timeout     time.Duration
	AuthToken   string
	ContentType string
	BasicAuth   (func() (string, string))
}

// WithHeader 设置请求头。
func WithHeader(key, value string) RequestOption {
	return func(c *RequestConfig) {
		if c.Header == nil {
			c.Header = make(http.Header)
		}
		c.Header.Set(key, value)
	}
}

// WithQuery 设置查询参数。
func WithQuery(key, value string) RequestOption {
	return func(c *RequestConfig) {
		if c.QueryParams == nil {
			c.QueryParams = make(map[string]string)
		}
		c.QueryParams[key] = value
	}
}

// WithTimeout 设置请求超时。
func WithTimeout(timeout time.Duration) RequestOption {
	return func(c *RequestConfig) {
		c.Timeout = timeout
	}
}

// WithAuthToken 设置认证令牌。
func WithAuthToken(token string) RequestOption {
	return func(c *RequestConfig) {
		c.AuthToken = token
	}
}

// WithBasicAuth 设置基本认证。
func WithBasicAuth(username, password string) RequestOption {
	return func(c *RequestConfig) {
		c.BasicAuth = func() (string, string) {
			return username, password
		}
	}
}
