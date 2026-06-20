// Package net 提供 HTTP 服务器抽象接口和默认实现。
//
// 核心接口：
//   - Server: HTTP 服务器抽象接口
package net

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Server 是 HTTP 服务器的统一接口。
//
// 所有 HTTP 框架实现都应实现此接口，
// 以便与 go-boot 容器系统集成。
type Server interface {

	// Start 启动 HTTP 服务器并开始监听请求。
	Start() error

	// Use 向服务器注册一个中间件。
	Use(m any)

	// Stop 优雅地停止服务器，等待正在处理的请求完成。
	Stop(ctx context.Context) error
}

// ServerOption 是服务器配置选项函数。
type ServerOption func(*httpServer)

// WithHost 设置服务器监听地址。
func WithHost(host string) ServerOption {
	return func(c *httpServer) {
		c.host = host
	}
}

// WithReadTimeout 设置读取超时。
func WithReadTimeout(timeout time.Duration) ServerOption {
	return func(c *httpServer) {
		c.readTimeout = timeout
	}
}

// WithWriteTimeout 设置写入超时。
func WithWriteTimeout(timeout time.Duration) ServerOption {
	return func(c *httpServer) {
		c.writeTimeout = timeout
	}
}

// WithIdleTimeout 设置空闲超时。
func WithIdleTimeout(timeout time.Duration) ServerOption {
	return func(c *httpServer) {
		c.idleTimeout = timeout
	}
}

// WithTls 设置 TLS 证书和密钥文件路径。
func WithTls(certFile, keyFile string) ServerOption {
	return func(c *httpServer) {
		c.certFile = certFile
		c.keyFile = keyFile
	}
}

// httpServer 是基于 net/http 原生包的 Server 接口实现。
type httpServer struct {
	server       *http.Server     // 底层 HTTP 服务器
	host         string           // 监听地址
	readTimeout  time.Duration    // 读取超时
	writeTimeout time.Duration    // 写入超时
	idleTimeout  time.Duration    // 空闲超时
	certFile     string           // TLS 证书文件路径
	keyFile      string           // TLS 密钥文件路径
	middlewares  []MiddlewareFunc // 中间件列表
	mu           sync.RWMutex     // 读写锁
	handler      http.Handler     // 自定义 HTTP 处理器
}

// NewHTTPServer 创建一个新的基于 net/http 的 HTTP 服务器。
func NewHTTPServer(opts ...ServerOption) Server {
	s := &httpServer{
		host:         ":8080",
		readTimeout:  30 * time.Second,
		writeTimeout: 30 * time.Second,
		idleTimeout:  120 * time.Second,
		middlewares:  make([]MiddlewareFunc, 0),
	}

	for _, opt := range opts {
		opt(s)
	}

	s.server = &http.Server{
		Addr:         s.host,
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writeTimeout,
		IdleTimeout:  s.idleTimeout,
	}

	return s
}

// Start 启动 HTTP 服务器并开始监听请求。
func (s *httpServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.handler != nil {
		s.server.Handler = s.wrapMiddleware(s.handler)
	} else {
		s.server.Handler = s.wrapMiddleware(http.DefaultServeMux)
	}

	if s.certFile != "" && s.keyFile != "" {
		return s.server.ListenAndServeTLS(s.certFile, s.keyFile)
	}
	return s.server.ListenAndServe()
}

// Use 向服务器注册一个中间件。
func (s *httpServer) Use(m any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	middleware, ok := m.(MiddlewareFunc)
	if !ok {
		panic(fmt.Sprintf("invalid middleware type: %T, expected MiddlewareFunc", m))
	}
	s.middlewares = append(s.middlewares, middleware)
}

// Stop 优雅地停止服务器，等待正在处理的请求完成。
func (s *httpServer) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fmt.Println("[go-boot] HTTP server stopping...")
	return s.server.Shutdown(ctx)
}

// SetHandler 设置自定义的 HTTP 处理器。
func (s *httpServer) SetHandler(handler http.Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handler = handler
}

// wrapMiddleware 将中间件包装到处理器中。
func (s *httpServer) wrapMiddleware(handler http.Handler) http.Handler {
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		middleware := s.middlewares[i]
		handler = &middlewareHandler{
			handler:    handler,
			middleware: middleware,
		}
	}
	return handler
}

// middlewareHandler 是一个包装器，用于在 net/http 中集成中间件。
type middlewareHandler struct {
	handler    http.Handler   // 下游 HTTP 处理器
	middleware MiddlewareFunc // 中间件函数
}

// ServeHTTP 实现 http.Handler 接口。
func (m *middlewareHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &httpContext{
		w: w,
		r: r,
	}
	m.middleware(ctx)
	if !ctx.aborted {
		m.handler.ServeHTTP(w, r)
	}
}

// httpContext 是 HandlerContext 接口的 net/http 实现。
type httpContext struct {
	w       http.ResponseWriter // 响应写入器
	r       *http.Request       // HTTP 请求
	aborted bool                // 是否已中止
	ctx     context.Context     // 自定义上下文
}

// RequestMethod 返回请求方法。
func (c *httpContext) RequestMethod() string {
	return c.r.Method
}

// RequestURI 返回请求 URI。
func (c *httpContext) RequestURI() string {
	return c.r.RequestURI
}

// Header 返回请求头。
func (c *httpContext) Header(key string) string {
	return c.r.Header.Get(key)
}

// SetStatusCode 设置响应状态码。
func (c *httpContext) SetStatusCode(code int) {
	c.w.WriteHeader(code)
}

// SetHeader 设置响应头。
func (c *httpContext) SetHeader(key, value string) {
	c.w.Header().Set(key, value)
}

// AbortWithStatus 中止请求并返回状态码。
func (c *httpContext) AbortWithStatus(code int) {
	c.aborted = true
	c.w.WriteHeader(code)
}

// AbortWithStatusJSON 中止请求并返回 JSON 响应。
func (c *httpContext) AbortWithStatusJSON(code int, body interface{}) {
	c.aborted = true
	c.w.Header().Set("Content-Type", "application/json")
	c.w.WriteHeader(code)
	if body != nil {
		if jsonData, err := json.Marshal(body); err == nil {
			if _, writeErr := c.w.Write(jsonData); writeErr != nil {
				fmt.Printf("[go-boot] failed to write JSON response: %v\n", writeErr)
			}
		}
	}
}

// Next 调用下一个中间件（在 net/http 实现中，中间件链由 middlewareHandler 处理）。
func (c *httpContext) Next() {
}

// IsAborted 返回请求是否被中止。
func (c *httpContext) IsAborted() bool {
	return c.aborted
}

// Context 返回上下文。
func (c *httpContext) Context() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	return c.r.Context()
}

// SetContext 设置上下文。
func (c *httpContext) SetContext(ctx context.Context) {
	c.ctx = ctx
	c.r = c.r.WithContext(ctx)
}
