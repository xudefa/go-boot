package security

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
)

// contextKey 定义上下文键类型，避免与其他包的字符串键冲突
type contextKey string

// HttpRequestAdapter HTTP请求适配器
// 将标准库的http.Request适配为SecurityRequest接口
type HttpRequestAdapter struct {
	request *http.Request
}

// NewHttpRequestAdapter 创建HTTP请求适配器
func NewHttpRequestAdapter(request *http.Request) *HttpRequestAdapter {
	return &HttpRequestAdapter{request: request}
}

// GetMethod 返回HTTP方法
func (a *HttpRequestAdapter) GetMethod() string {
	return a.request.Method
}

// GetURI 返回请求URI
func (a *HttpRequestAdapter) GetURI() string {
	return a.request.URL.Path
}

// GetHeader 返回请求头
func (a *HttpRequestAdapter) GetHeader(key string) string {
	return a.request.Header.Get(key)
}

// SetAttribute 设置请求属性
func (a *HttpRequestAdapter) SetAttribute(key string, value any) {
	ctx := context.WithValue(a.request.Context(), contextKey(key), value)
	*a.request = *a.request.WithContext(ctx)
}

// GetAttribute 获取请求属性
func (a *HttpRequestAdapter) GetAttribute(key string) (any, bool) {
	value := a.request.Context().Value(contextKey(key))
	return value, value != nil
}

// HttpResponseAdapter HTTP响应适配器
// 将标准库的http.ResponseWriter适配为SecurityResponse接口
type HttpResponseAdapter struct {
	responseWriter http.ResponseWriter
	statusCode     int
	written        bool
}

// NewHttpResponseAdapter 创建HTTP响应适配器
func NewHttpResponseAdapter(responseWriter http.ResponseWriter) *HttpResponseAdapter {
	return &HttpResponseAdapter{responseWriter: responseWriter, statusCode: 200}
}

// SetStatusCode 设置响应状态码
func (a *HttpResponseAdapter) SetStatusCode(code int) {
	a.statusCode = code
	a.responseWriter.WriteHeader(code)
	a.written = true
}

// StatusCode 返回当前响应状态码
func (a *HttpResponseAdapter) StatusCode() int {
	return a.statusCode
}

// SetHeader 设置响应头
func (a *HttpResponseAdapter) SetHeader(key, value string) {
	a.responseWriter.Header().Set(key, value)
}

// Write 写入响应体
func (a *HttpResponseAdapter) Write(data []byte) error {
	_, err := a.responseWriter.Write(data)
	a.written = true
	return err
}

// headersWritten 检查是否已写入响应头
func (a *HttpResponseAdapter) headersWritten() bool {
	return a.written
}

// SecurityFilterChainHandler 安全过滤器链处理器
// 将SecurityFilterChain适配为标准库的http.Handler
type SecurityFilterChainHandler struct {
	securityFilterChain SecurityFilterChain // 安全过滤器链
	nextHandler         http.Handler        // 下一个处理器
}

// NewSecurityFilterChainHandler 创建安全过滤器链处理器
func NewSecurityFilterChainHandler(securityFilterChain SecurityFilterChain, nextHandler http.Handler) *SecurityFilterChainHandler {
	return &SecurityFilterChainHandler{
		securityFilterChain: securityFilterChain,
		nextHandler:         nextHandler,
	}
}

// ServeHTTP 实现http.Handler接口
func (h *SecurityFilterChainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request := NewHttpRequestAdapter(r)
	response := NewHttpResponseAdapter(w)

	ctx := r.Context()
	err := h.securityFilterChain.DoFilter(ctx, request, response)
	if err != nil {
		if !response.headersWritten() {
			response.SetStatusCode(500)
			if writeErr := response.Write([]byte(err.Error())); writeErr != nil {
				fmt.Printf("[go-boot] failed to write error response: %v\n", writeErr)
			}
		}
		return
	}

	if response.StatusCode() >= 400 {
		return
	}

	if h.nextHandler != nil {
		if authVal, ok := request.GetAttribute("security.currentAuthentication"); ok {
			if auth, ok := authVal.(Authentication); ok {
				r = r.WithContext(ContextWithAuthentication(r.Context(), auth))
			}
		}
		h.nextHandler.ServeHTTP(w, r)
	}
}

// SetNextHandler 设置下一个处理器
func (h *SecurityFilterChainHandler) SetNextHandler(handler http.Handler) {
	h.nextHandler = handler
}

// BasicAuthenticationFilter Basic认证过滤器
// 从Authorization头中提取Basic认证信息并进行认证
type BasicAuthenticationFilter struct {
	authenticationManager AuthenticationManager // 认证管理器
}

// NewBasicAuthenticationFilter 创建Basic认证过滤器
func NewBasicAuthenticationFilter(authenticationManager AuthenticationManager) *BasicAuthenticationFilter {
	return &BasicAuthenticationFilter{
		authenticationManager: authenticationManager,
	}
}

// DoFilter 处理Basic认证
func (f *BasicAuthenticationFilter) DoFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain SecurityFilterChain) error {
	authHeader := request.GetHeader("Authorization")
	if authHeader == "" {
		return chain.DoFilter(ctx, request, response)
	}

	username, password, err := f.extractBasicAuth(authHeader)
	if err != nil {
		return err
	}

	authToken := NewUsernamePasswordAuthenticationToken(username, password)
	authenticated, err := f.authenticationManager.Authenticate(ctx, authToken)
	if err != nil {
		return err
	}

	SetAuthentication(authenticated)

	return chain.DoFilter(ctx, request, response)
}

// extractBasicAuth 从Authorization头中提取用户名和密码
func (f *BasicAuthenticationFilter) extractBasicAuth(authHeader string) (string, string, error) {
	const prefix = "Basic "
	if len(authHeader) < len(prefix) || authHeader[:len(prefix)] != prefix {
		return "", "", ErrBadCredentials
	}

	encoded := authHeader[len(prefix):]
	decoded, err := f.decodeBase64(encoded)
	if err != nil {
		return "", "", ErrBadCredentials
	}

	credentials := string(decoded)
	sepIndex := -1
	for i, c := range credentials {
		if c == ':' {
			sepIndex = i
			break
		}
	}

	if sepIndex == -1 {
		return "", "", ErrBadCredentials
	}

	username := credentials[:sepIndex]
	password := credentials[sepIndex+1:]

	return username, password, nil
}

// decodeBase64 Base64解码
func (f *BasicAuthenticationFilter) decodeBase64(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}
