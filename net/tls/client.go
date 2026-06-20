// Package tls 提供 TLS 客户端构建器，封装 net.NetClient 并配置 TLS 连接。
//
// 核心组件：
//   - ClientBuilder: HTTPS 客户端构建器
//   - ClientOption: 客户端配置选项函数
//
// 使用示例：
//
//	client := tls.NewClient("https://example.com",
//	    tls.WithTLSConfig(tlsCfg),
//	    tls.WithTimeout(30*time.Second),
//	)
//	resp, err := client.Get(ctx, "/api/users")
package tls

import (
	"crypto/tls"
	goNet "net"
	"net/http"
	"time"

	"github.com/xudefa/go-boot/net"
)

// ClientOption 是 HTTPS 客户端配置选项。
type ClientOption func(*ClientBuilder)

// ClientBuilder 是 HTTPS 客户端构建器，封装 net.NetClient 并配置 TLS。
//
// 字段说明:
//   - baseURL: 基础 URL（应以 https:// 开头）
//   - tlsConfig: TLS 配置
//   - timeout: 请求超时时间
//   - transport: 自定义 HTTP Transport
//   - headers: 默认请求头
type ClientBuilder struct {
	baseURL   string
	tlsConfig *tls.Config
	timeout   time.Duration
	transport *http.Transport
	headers   map[string]string
}

// NewClient 创建 HTTPS 客户端构建器。
//
// 参数:
//   - baseURL: 基础 URL，如 "https://example.com"
//   - opts: 可选配置选项
//
// 返回值:
//   - *net.NetClient: 配置好的 HTTPS 客户端实例
//
// 示例:
//
//	client := https.NewClient("https://api.example.com",
//	    https.WithInsecureTLS(),
//	    https.WithTimeout(30*time.Second),
//	)
//	resp, err := client.Get(ctx, "/api/users")
func NewClient(baseURL string, opts ...ClientOption) *net.NetClient {
	b := &ClientBuilder{
		baseURL: baseURL,
		timeout: net.DefaultTimeout,
	}

	for _, opt := range opts {
		opt(b)
	}

	transport := b.transport
	if transport == nil {
		transport = &http.Transport{
			TLSClientConfig: b.tlsConfig,
			DialContext: (&goNet.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
	}

	header := make(http.Header)
	for k, v := range b.headers {
		header.Set(k, v)
	}

	return net.NewClient(baseURL,
		net.WithClientTimeout(b.timeout),
		net.WithHeaders(header),
		func(c *net.NetClient) {
			c.HttpClient.Transport = transport
		},
	)
}

// WithTLSConfig 设置 TLS 配置。
//
// 参数:
//   - tlsConfig: TLS 配置
//
// 返回值:
//   - ClientOption: 客户端配置选项
func WithTLSConfig(tlsConfig *tls.Config) ClientOption {
	return func(b *ClientBuilder) {
		b.tlsConfig = tlsConfig
	}
}

// WithInsecureTLS 跳过 TLS 证书验证（仅用于开发测试）。
//
// 返回值:
//   - ClientOption: 客户端配置选项
func WithInsecureTLS() ClientOption {
	return func(b *ClientBuilder) {
		b.tlsConfig = InsecureTLSConfig()
	}
}

// WithTimeout 设置请求超时时间。
//
// 参数:
//   - timeout: 超时时间
//
// 返回值:
//   - ClientOption: 客户端配置选项
func WithTimeout(timeout time.Duration) ClientOption {
	return func(b *ClientBuilder) {
		b.timeout = timeout
	}
}

// WithDefaultHeader 设置默认请求头。
//
// 参数:
//   - key: 请求头名称
//   - value: 请求头值
//
// 返回值:
//   - ClientOption: 客户端配置选项
func WithDefaultHeader(key, value string) ClientOption {
	return func(b *ClientBuilder) {
		if b.headers == nil {
			b.headers = make(map[string]string)
		}
		b.headers[key] = value
	}
}

// WithTransport 设置自定义 HTTP Transport。
//
// 参数:
//   - transport: HTTP Transport
//
// 返回值:
//   - ClientOption: 客户端配置选项
func WithTransport(transport *http.Transport) ClientOption {
	return func(b *ClientBuilder) {
		b.transport = transport
	}
}
