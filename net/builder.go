package net

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// HttpClientBuilder HTTP客户端构建器，支持链式配置
type HttpClientBuilder struct {
	baseURL     string
	timeout     time.Duration
	headers     http.Header
	middleware  []ClientMiddlewareFunc
	retryConfig *RetryConfig
}

// NewHttpClientBuilder 创建HTTP客户端构建器
func NewHttpClientBuilder() *HttpClientBuilder {
	return &HttpClientBuilder{
		timeout: DefaultTimeout,
		headers: make(http.Header),
	}
}

// BaseURL 设置基础URL
func (b *HttpClientBuilder) BaseURL(url string) *HttpClientBuilder {
	b.baseURL = url
	return b
}

// Timeout 设置请求超时时间
func (b *HttpClientBuilder) Timeout(timeout time.Duration) *HttpClientBuilder {
	b.timeout = timeout
	return b
}

// Header 添加默认请求头
func (b *HttpClientBuilder) Header(key, value string) *HttpClientBuilder {
	b.headers.Set(key, value)
	return b
}

// Middleware 添加客户端中间件
func (b *HttpClientBuilder) Middleware(m ClientMiddlewareFunc) *HttpClientBuilder {
	b.middleware = append(b.middleware, m)
	return b
}

// Retry 配置重试策略
func (b *HttpClientBuilder) Retry(opts ...RetryOption) *HttpClientBuilder {
	cfg := &RetryConfig{
		maxAttempts: 3,
		strategy:    NewExponentialBackoff(100*time.Millisecond, 10*time.Second),
	}
	for _, opt := range opts {
		opt(cfg)
	}
	b.retryConfig = cfg
	return b
}

// Build 构建HTTP客户端
func (b *HttpClientBuilder) Build() (HttpClient, error) {
	if b.baseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	client := NewClient(b.baseURL,
		WithClientTimeout(b.timeout),
		WithHeaders(b.headers),
	)

	for _, m := range b.middleware {
		client.WithMiddleware(m)
	}

	// 如果配置了重试，返回可重试客户端
	if b.retryConfig != nil {
		retryableClient := NewRetryableClient(client,
			WithMaxAttempts(b.retryConfig.maxAttempts),
			WithRetryStrategy(b.retryConfig.strategy),
		)
		if b.retryConfig.onRetry != nil {
			// 重新应用onRetry回调
			retryableClient = NewRetryableClient(client,
				WithMaxAttempts(b.retryConfig.maxAttempts),
				WithRetryStrategy(b.retryConfig.strategy),
				WithOnRetry(b.retryConfig.onRetry),
			)
		}
		return retryableClient, nil
	}

	return client, nil
}

// MustBuild 构建HTTP客户端，失败则panic
func (b *HttpClientBuilder) MustBuild() HttpClient {
	client, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build HTTP client: %v", err))
	}
	return client
}

// HttpServerBuilder HTTP服务器构建器，支持链式配置
type HttpServerBuilder struct {
	host         string
	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration
	certFile     string
	keyFile      string
	middlewares  []MiddlewareFunc
	handler      http.Handler
}

// NewHttpServerBuilder 创建HTTP服务器构建器
func NewHttpServerBuilder() *HttpServerBuilder {
	return &HttpServerBuilder{
		host:         ":8080",
		readTimeout:  30 * time.Second,
		writeTimeout: 30 * time.Second,
		idleTimeout:  120 * time.Second,
	}
}

// Host 设置监听地址
func (b *HttpServerBuilder) Host(host string) *HttpServerBuilder {
	b.host = host
	return b
}

// ReadTimeout 设置读取超时
func (b *HttpServerBuilder) ReadTimeout(timeout time.Duration) *HttpServerBuilder {
	b.readTimeout = timeout
	return b
}

// WriteTimeout 设置写入超时
func (b *HttpServerBuilder) WriteTimeout(timeout time.Duration) *HttpServerBuilder {
	b.writeTimeout = timeout
	return b
}

// IdleTimeout 设置空闲超时
func (b *HttpServerBuilder) IdleTimeout(timeout time.Duration) *HttpServerBuilder {
	b.idleTimeout = timeout
	return b
}

// Tls 设置TLS证书
func (b *HttpServerBuilder) Tls(certFile, keyFile string) *HttpServerBuilder {
	b.certFile = certFile
	b.keyFile = keyFile
	return b
}

// Middleware 添加中间件
func (b *HttpServerBuilder) Middleware(m MiddlewareFunc) *HttpServerBuilder {
	b.middlewares = append(b.middlewares, m)
	return b
}

// Handler 设置HTTP处理器
func (b *HttpServerBuilder) Handler(handler http.Handler) *HttpServerBuilder {
	b.handler = handler
	return b
}

// Build 构建HTTP服务器
func (b *HttpServerBuilder) Build() (Server, error) {
	server := NewHTTPServer(
		WithHost(b.host),
		WithReadTimeout(b.readTimeout),
		WithWriteTimeout(b.writeTimeout),
		WithIdleTimeout(b.idleTimeout),
	)

	if b.certFile != "" && b.keyFile != "" {
		WithTls(b.certFile, b.keyFile)(server.(*httpServer))
	}

	if b.handler != nil {
		server.(*httpServer).SetHandler(b.handler)
	}

	for _, m := range b.middlewares {
		server.Use(m)
	}

	return server, nil
}

// MustBuild 构建HTTP服务器，失败则panic
func (b *HttpServerBuilder) MustBuild() Server {
	server, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build HTTP server: %v", err))
	}
	return server
}

// RequestBuilder 请求构建器，简化HTTP请求构建
type RequestBuilder struct {
	method  string
	path    string
	body    any
	opts    []RequestOption
	headers http.Header
	query   map[string]string
}

// NewRequestBuilder 创建请求构建器
func NewRequestBuilder(method, path string) *RequestBuilder {
	return &RequestBuilder{
		method:  method,
		path:    path,
		headers: make(http.Header),
		query:   make(map[string]string),
	}
}

// Body 设置请求体
func (b *RequestBuilder) Body(body any) *RequestBuilder {
	b.body = body
	return b
}

// Header 添加请求头
func (b *RequestBuilder) Header(key, value string) *RequestBuilder {
	b.headers.Set(key, value)
	return b
}

// Query 添加查询参数
func (b *RequestBuilder) Query(key, value string) *RequestBuilder {
	b.query[key] = value
	return b
}

// AuthToken 设置认证令牌
func (b *RequestBuilder) AuthToken(token string) *RequestBuilder {
	b.opts = append(b.opts, WithAuthToken(token))
	return b
}

// ContentType 设置Content-Type
func (b *RequestBuilder) ContentType(contentType string) *RequestBuilder {
	b.opts = append(b.opts, WithContentType(contentType))
	return b
}

// Timeout 设置请求超时
func (b *RequestBuilder) Timeout(timeout time.Duration) *RequestBuilder {
	b.opts = append(b.opts, WithTimeout(timeout))
	return b
}

// Build 构建请求选项
func (b *RequestBuilder) Build() []RequestOption {
	opts := make([]RequestOption, len(b.opts))
	copy(opts, b.opts)

	for key, values := range b.headers {
		for _, value := range values {
			k, v := key, value
			opts = append(opts, func(req *HttpRequest) {
				if req.Header == nil {
					req.Header = make(http.Header)
				}
				req.Header.Add(k, v)
			})
		}
	}

	for key, value := range b.query {
		k, v := key, value
		opts = append(opts, WithQuery(k, v))
	}

	return opts
}

// Execute 执行请求
func (b *RequestBuilder) Execute(ctx context.Context, client HttpClient) (*HttpResponse, error) {
	opts := b.Build()

	switch b.method {
	case "GET":
		return client.Get(ctx, b.path, opts...)
	case "POST":
		return client.Post(ctx, b.path, b.body, opts...)
	case "PUT":
		return client.Put(ctx, b.path, b.body, opts...)
	case "PATCH":
		return client.Patch(ctx, b.path, b.body, opts...)
	case "DELETE":
		return client.Delete(ctx, b.path, opts...)
	case "HEAD":
		return client.Head(ctx, b.path, opts...)
	case "OPTIONS":
		return client.Options(ctx, b.path, opts...)
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", b.method)
	}
}
