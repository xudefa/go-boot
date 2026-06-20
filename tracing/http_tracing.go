// Package tracing 提供 HTTP 追踪中间件和上下文传播功能。
//
// 包含 HTTP 请求/响应的追踪属性收集、上下文传播器、
// 以及适用于标准库和 go-boot 网络框架的追踪中间件。
package tracing

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	net_pkg "github.com/xudefa/go-boot/net"
)

// TextMapCarrier 文本映射载体，用于在 HTTP header 中提取和注入追踪上下文
type TextMapCarrier interface {
	Get(key string) string
	Set(key string, value string)
	Keys() []string
}

// HTTPHeadersCarrier 基于 http.Header 的 TextMapCarrier 实现
type HTTPHeadersCarrier struct {
	headers http.Header
}

// NewHTTPHeadersCarrier 创建基于 http.Header 的载体实例
func NewHTTPHeadersCarrier(headers http.Header) *HTTPHeadersCarrier {
	return &HTTPHeadersCarrier{headers: headers}
}

func (c *HTTPHeadersCarrier) Get(key string) string {
	return c.headers.Get(key)
}

func (c *HTTPHeadersCarrier) Set(key string, value string) {
	c.headers.Set(key, value)
}

func (c *HTTPHeadersCarrier) Keys() []string {
	keys := make([]string, 0, len(c.headers))
	for k := range c.headers {
		keys = append(keys, k)
	}
	return keys
}

// TraceContextPropagator 追踪上下文传播器
// 用于在服务之间传递追踪上下文
type TraceContextPropagator interface {
	Extract(ctx context.Context, carrier TextMapCarrier) context.Context
	Inject(ctx context.Context, carrier TextMapCarrier)
}

// NoopPropagator 空操作的上下文传播器
type NoopPropagator struct{}

// Extract 直接返回原始上下文，不做任何提取操作
func (n *NoopPropagator) Extract(ctx context.Context, carrier TextMapCarrier) context.Context {
	return ctx
}

// Inject 不执行任何注入操作
func (n *NoopPropagator) Inject(ctx context.Context, carrier TextMapCarrier) {}

var defaultPropagator TraceContextPropagator = &NoopPropagator{}

// SetPropagator 设置全局上下文传播器
func SetPropagator(p TraceContextPropagator) {
	defaultPropagator = p
}

// GetPropagator 获取全局上下文传播器
func GetPropagator() TraceContextPropagator {
	return defaultPropagator
}

// ExtractTraceContext 从 carrier 中提取追踪上下文
func ExtractTraceContext(ctx context.Context, carrier TextMapCarrier) context.Context {
	return defaultPropagator.Extract(ctx, carrier)
}

// InjectTraceContext 将追踪上下文注入到 carrier 中
func InjectTraceContext(ctx context.Context, carrier TextMapCarrier) {
	defaultPropagator.Inject(ctx, carrier)
}

// HTTPStatusToSpanStatusCode 将 HTTP 状态码转换为 Span 状态码
func HTTPStatusToSpanStatusCode(statusCode int) SpanStatusCode {
	switch {
	case statusCode >= http.StatusInternalServerError:
		return SpanStatusError
	case statusCode >= http.StatusBadRequest:
		return SpanStatusUnset
	default:
		return SpanStatusOK
	}
}

// HTTPTraceAttributes HTTP 追踪属性集合
type HTTPTraceAttributes struct {
	Method     string
	Target     string
	Host       string
	StatusCode int
	UserAgent  string
	ClientIP   string
}

// SetAttributes 设置所有 HTTP 相关属性到 Span
func (h *HTTPTraceAttributes) SetAttributes(span Span) {
	if h.Method != "" {
		span.SetAttribute("http.method", h.Method)
	}
	if h.Target != "" {
		span.SetAttribute("http.target", h.Target)
	}
	if h.Host != "" {
		span.SetAttribute("http.host", h.Host)
	}
	if h.StatusCode != 0 {
		span.SetAttribute("http.status_code", h.StatusCode)
		span.SetStatus(HTTPStatusToSpanStatusCode(h.StatusCode))
	}
	if h.UserAgent != "" {
		span.SetAttribute("http.user_agent", h.UserAgent)
	}
	if h.ClientIP != "" {
		span.SetAttribute("http.client_ip", h.ClientIP)
	}
}

// WithHTTPAttributes 从 HTTP 请求创建追踪属性
func WithHTTPAttributes(r *http.Request) *HTTPTraceAttributes {
	return &HTTPTraceAttributes{
		Method:    r.Method,
		Target:    r.URL.Path,
		Host:      r.Host,
		UserAgent: r.UserAgent(),
		ClientIP:  getClientIP(r),
	}
}

// WithHTTPResponseAttributes 从 HTTP 响应创建追踪属性
func WithHTTPResponseAttributes(statusCode int) *HTTPTraceAttributes {
	return &HTTPTraceAttributes{
		StatusCode: statusCode,
	}
}

// StartHTTPServerSpan 创建 HTTP 服务器端 Span
func StartHTTPServerSpan(ctx context.Context, tracer Tracer, spanName, method, target, host string) (context.Context, Span) {
	attrs := &HTTPTraceAttributes{
		Method: method,
		Target: target,
		Host:   host,
	}

	ctx, span := tracer.Start(ctx, spanName,
		WithSpanKind(SpanKindServer))

	attrs.SetAttributes(span)
	return ctx, span
}

// StartHTTPClientSpan 创建 HTTP 客户端 Span
func StartHTTPClientSpan(ctx context.Context, tracer Tracer, spanName, method, target, host string) (context.Context, Span) {
	attrs := &HTTPTraceAttributes{
		Method: method,
		Target: target,
		Host:   host,
	}

	ctx, span := tracer.Start(ctx, spanName,
		WithSpanKind(SpanKindClient))

	attrs.SetAttributes(span)
	return ctx, span
}

// HTTPTraceMiddleware HTTP 追踪中间件
//
// 自动为 HTTP 请求创建追踪 Span，记录请求的关键信息。
// 适用于标准库 http.Handler 接口。
type HTTPTraceMiddleware struct {
	tracer Tracer
}

// NewHTTPTraceMiddleware 创建 HTTP 追踪中间件
func NewHTTPTraceMiddleware(tracer Tracer) *HTTPTraceMiddleware {
	return &HTTPTraceMiddleware{
		tracer: tracer,
	}
}

// Middleware 实现 HTTP 中间件接口
func (m *HTTPTraceMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 创建追踪上下文
		attrs := WithHTTPAttributes(r)

		ctx, span := StartHTTPServerSpan(r.Context(), m.tracer,
			r.Method+" "+r.URL.Path,
			r.Method, r.URL.Path, r.Host)
		defer span.End()

		// 设置请求属性
		attrs.SetAttributes(span)

		// 创建包装的 ResponseWriter 来捕获状态码
		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

		// 执行下一个处理器
		next.ServeHTTP(wrapped, r.WithContext(ctx))

		// 设置响应属性
		respAttrs := WithHTTPResponseAttributes(wrapped.statusCode)
		respAttrs.SetAttributes(span)
	})
}

// responseWriter 包装 ResponseWriter 以捕获状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.statusCode = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

// getClientIP 获取客户端 IP 地址
func getClientIP(r *http.Request) string {
	// 检查 X-Forwarded-For 头
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	// 检查 X-Real-IP 头
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// 使用 RemoteAddr
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

// NetTraceMiddleware 为 net.HandlerContext 设计的追踪中间件
//
// 适用于 go-boot 网络框架的中间件接口。
func NetTraceMiddleware(tracer Tracer, operationName string) net_pkg.MiddlewareFunc {
	return func(hc net_pkg.HandlerContext) {
		// 尝试从 HandlerContext 获取请求信息以构建操作名称
		method := hc.RequestMethod()
		uri := hc.RequestURI()
		spanName := operationName
		if spanName == "" {
			spanName = fmt.Sprintf("%s %s", method, uri)
		}

		// 开始追踪 Span
		ctx, span := tracer.Start(context.Background(), spanName)
		defer span.End()

		// 设置基本属性
		span.SetAttribute("http.method", method)
		span.SetAttribute("http.url", uri)

		// 设置上下文并继续处理
		hc.SetContext(ctx)
		hc.Next()
	}
}
