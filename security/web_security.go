package security

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

// CSRFTokenRepository CSRF令牌仓库接口
type CsrfTokenRepository interface {
	GenerateToken(ctx context.Context, request SecurityRequest) (*CsrfToken, error)
	ValidateToken(ctx context.Context, request SecurityRequest, token string) bool
	SaveToken(ctx context.Context, request SecurityRequest, response SecurityResponse, token *CsrfToken)
	ClearToken(ctx context.Context, request SecurityRequest, response SecurityResponse)
}

// CsrfToken CSRF令牌
type CsrfToken struct {
	Identifier string // 请求路径标识
	Value      string // 令牌值
}

// CsrfFilter CSRF防护过滤器
type CsrfFilter struct {
	tokenRepository CsrfTokenRepository // 令牌仓库
	excludePaths    []string            // 排除路径
}

// NewCsrfFilter 创建CSRF过滤器
func NewCsrfFilter(tokenRepository CsrfTokenRepository) *CsrfFilter {
	return &CsrfFilter{
		tokenRepository: tokenRepository,
		excludePaths:    []string{},
	}
}

// AddExcludePath 添加排除路径
func (f *CsrfFilter) AddExcludePath(paths ...string) {
	f.excludePaths = append(f.excludePaths, paths...)
}

// DoFilter 执行CSRF检查
func (f *CsrfFilter) DoFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain SecurityFilterChain) error {
	uri := request.GetURI()

	// 检查是否排除
	for _, path := range f.excludePaths {
		if strings.HasPrefix(uri, path) {
			return chain.DoFilter(ctx, request, response)
		}
	}

	method := request.GetMethod()

	// GET、HEAD、OPTIONS、TRACE请求不需要CSRF检查
	if method == "GET" || method == "HEAD" || method == "OPTIONS" || method == "TRACE" {
		token, err := f.tokenRepository.GenerateToken(ctx, request)
		if err == nil && token != nil {
			f.tokenRepository.SaveToken(ctx, request, response, token)
			request.SetAttribute("csrf.token", token.Value)
		}
		return chain.DoFilter(ctx, request, response)
	}

	// 其他请求需要验证CSRF令牌
	token := request.GetHeader("X-CSRF-Token")
	if token == "" {
		token = request.GetHeader("X-XSRF-Token")
	}
	if token == "" {
		return fmt.Errorf("missing CSRF token")
	}

	if !f.tokenRepository.ValidateToken(ctx, request, token) {
		return fmt.Errorf("invalid CSRF token")
	}

	return chain.DoFilter(ctx, request, response)
}

// CookieCsrfTokenRepository 基于Cookie的CSRF令牌仓库
type CookieCsrfTokenRepository struct {
	cookieName     string
	headerName     string
	cookieHttpOnly bool
	secure         bool
	sameSite       http.SameSite
}

// NewCookieCsrfTokenRepository 创建基于Cookie的CSRF令牌仓库
func NewCookieCsrfTokenRepository() *CookieCsrfTokenRepository {
	return &CookieCsrfTokenRepository{
		cookieName:     "_csrf_token",
		headerName:     "X-CSRF-Token",
		cookieHttpOnly: false,
		secure:         false,
		sameSite:       http.SameSiteLaxMode,
	}
}

// GenerateToken 生成新的CSRF令牌
func (r *CookieCsrfTokenRepository) GenerateToken(ctx context.Context, request SecurityRequest) (*CsrfToken, error) {
	token := generateSecureToken(32)
	return &CsrfToken{
		Identifier: request.GetURI(),
		Value:      token,
	}, nil
}

// ValidateToken 验证CSRF令牌
func (r *CookieCsrfTokenRepository) ValidateToken(ctx context.Context, request SecurityRequest, token string) bool {
	savedToken, exists := request.GetAttribute("_csrf_token")
	if !exists {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(token), []byte(savedToken.(string))) == 1
}

// SaveToken 保存CSRF令牌到Cookie
func (r *CookieCsrfTokenRepository) SaveToken(ctx context.Context, request SecurityRequest, response SecurityResponse, token *CsrfToken) {
	response.SetHeader("Set-Cookie", fmt.Sprintf("%s=%s; Path=/; HttpOnly=%v; SameSite=%v",
		r.cookieName, token.Value, r.cookieHttpOnly, r.sameSite))
}

// ClearToken 清除CSRF令牌
func (r *CookieCsrfTokenRepository) ClearToken(ctx context.Context, request SecurityRequest, response SecurityResponse) {
	response.SetHeader("Set-Cookie", fmt.Sprintf("%s=; Path=/; Max-Age=0", r.cookieName))
}

// generateSecureToken 生成安全随机令牌
func generateSecureToken(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	token := make([]byte, length)
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range token {
		token[i] = charset[random.Intn(len(charset))]
	}
	return string(token)
}

// LogoutHandler 登出处理器接口
type LogoutHandler interface {
	Logout(ctx context.Context, request SecurityRequest, response SecurityResponse, authentication Authentication)
}

// LogoutFilter 登出过滤器
type LogoutFilter struct {
	logoutUrl      string               // 登出URL
	handlers       []LogoutHandler      // 登出处理器
	successHandler LogoutSuccessHandler // 登出成功处理器
	filterChain    SecurityFilterChain
	httpMethods    []string // 支持的HTTP方法
}

// NewLogoutFilter 创建登出过滤器
func NewLogoutFilter(logoutUrl string, handlers []LogoutHandler) *LogoutFilter {
	return &LogoutFilter{
		logoutUrl:   logoutUrl,
		handlers:    handlers,
		httpMethods: []string{"POST", "GET"},
	}
}

// AddLogoutHandler 添加登出处理器
func (f *LogoutFilter) AddLogoutHandler(handler LogoutHandler) {
	f.handlers = append(f.handlers, handler)
}

// SetSuccessHandler 设置登出成功处理器
func (f *LogoutFilter) SetSuccessHandler(handler LogoutSuccessHandler) {
	f.successHandler = handler
}

// DoFilter 处理登出请求
func (f *LogoutFilter) DoFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain SecurityFilterChain) error {
	f.filterChain = chain

	method := request.GetMethod()
	methodAllowed := false
	for _, m := range f.httpMethods {
		if m == method {
			methodAllowed = true
			break
		}
	}
	if !methodAllowed {
		return chain.DoFilter(ctx, request, response)
	}

	uri := request.GetURI()
	if uri != f.logoutUrl {
		return chain.DoFilter(ctx, request, response)
	}

	authentication := GetAuthentication()

	for _, handler := range f.handlers {
		handler.Logout(ctx, request, response, authentication)
	}

	ClearAuthentication()

	if f.successHandler != nil {
		f.successHandler.OnLogoutSuccess(ctx, request, response, authentication)
	} else {
		response.SetStatusCode(200)
		response.SetHeader("Content-Type", "application/json")
		if writeErr := response.Write([]byte(`{"message":"logout success"}`)); writeErr != nil {
			fmt.Printf("[go-boot] failed to write logout success response: %v\n", writeErr)
		}
	}

	return nil
}

// LogoutSuccessHandler 登出成功处理器接口
type LogoutSuccessHandler interface {
	OnLogoutSuccess(ctx context.Context, request SecurityRequest, response SecurityResponse, authentication Authentication)
}

// DefaultLogoutSuccessHandler 默认登出成功处理器
type DefaultLogoutSuccessHandler struct {
	defaultTargetUrl string
}

// NewDefaultLogoutSuccessHandler 创建默认登出成功处理器
func NewDefaultLogoutSuccessHandler(defaultTargetUrl string) *DefaultLogoutSuccessHandler {
	return &DefaultLogoutSuccessHandler{
		defaultTargetUrl: defaultTargetUrl,
	}
}

// OnLogoutSuccess 处理登出成功
func (h *DefaultLogoutSuccessHandler) OnLogoutSuccess(ctx context.Context, request SecurityRequest, response SecurityResponse, authentication Authentication) {
	response.SetStatusCode(302)
	response.SetHeader("Location", h.defaultTargetUrl)
}

// SimpleLogoutSuccessHandler 简单登出成功处理器
type SimpleLogoutSuccessHandler struct {
	targetUrl string
}

// NewSimpleLogoutSuccessHandler 创建简单登出成功处理器
func NewSimpleLogoutSuccessHandler(targetUrl string) *SimpleLogoutSuccessHandler {
	return &SimpleLogoutSuccessHandler{
		targetUrl: targetUrl,
	}
}

// OnLogoutSuccess 处理登出成功，重定向到指定URL
func (h *SimpleLogoutSuccessHandler) OnLogoutSuccess(ctx context.Context, request SecurityRequest, response SecurityResponse, authentication Authentication) {
	response.SetStatusCode(302)
	response.SetHeader("Location", h.targetUrl)
}

// SecurityContextLogoutHandler 安全上下文登出处理器
type SecurityContextLogoutHandler struct{}

// NewSecurityContextLogoutHandler 创建安全上下文登出处理器
func NewSecurityContextLogoutHandler() *SecurityContextLogoutHandler {
	return &SecurityContextLogoutHandler{}
}

// Logout 清除安全上下文
func (h *SecurityContextLogoutHandler) Logout(ctx context.Context, request SecurityRequest, response SecurityResponse, authentication Authentication) {
	ClearAuthentication()
}

// CookieClearingLogoutHandler Cookie清除登出处理器
type CookieClearingLogoutHandler struct {
	cookieNames []string
}

// NewCookieClearingLogoutHandler 创建Cookie清除登出处理器
func NewCookieClearingLogoutHandler(cookieNames ...string) *CookieClearingLogoutHandler {
	return &CookieClearingLogoutHandler{
		cookieNames: cookieNames,
	}
}

// Logout 清除指定Cookie
func (h *CookieClearingLogoutHandler) Logout(ctx context.Context, request SecurityRequest, response SecurityResponse, authentication Authentication) {
	for _, name := range h.cookieNames {
		response.SetHeader("Set-Cookie", fmt.Sprintf("%s=; Path=/; Max-Age=0", name))
	}
}

// UsernamePasswordAuthenticationFilter 用户名密码认证过滤器
type UsernamePasswordAuthenticationFilter struct {
	loginProcessingUrl    string                // 登录处理URL
	defaultSuccessUrl     string                // 默认成功URL
	failureUrl            string                // 失败URL
	authenticationManager AuthenticationManager // 认证管理器
}

// NewUsernamePasswordAuthenticationFilterWithDefaults 创建带默认值的用户名密码认证过滤器
func NewUsernamePasswordAuthenticationFilterWithDefaults(
	loginProcessingUrl,
	defaultSuccessUrl,
	failureUrl string,
	authManager AuthenticationManager,
) *UsernamePasswordAuthenticationFilter {
	return &UsernamePasswordAuthenticationFilter{
		loginProcessingUrl:    loginProcessingUrl,
		defaultSuccessUrl:     defaultSuccessUrl,
		failureUrl:            failureUrl,
		authenticationManager: authManager,
	}
}

// DoFilter 处理用户名密码认证
func (f *UsernamePasswordAuthenticationFilter) DoFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain SecurityFilterChain) error {
	if request.GetMethod() != "POST" {
		return chain.DoFilter(ctx, request, response)
	}

	uri := request.GetURI()
	if uri != f.loginProcessingUrl {
		return chain.DoFilter(ctx, request, response)
	}

	username := request.GetHeader("username")
	password := request.GetHeader("password")

	if username == "" || password == "" {
		response.SetStatusCode(401)
		if writeErr := response.Write([]byte("missing username or password")); writeErr != nil {
			fmt.Printf("[go-boot] failed to write missing credentials response: %v\n", writeErr)
		}
		return nil
	}

	authToken := NewUsernamePasswordAuthenticationToken(username, password)
	authenticated, err := f.authenticationManager.Authenticate(ctx, authToken)

	if err != nil {
		response.SetStatusCode(401)
		response.SetHeader("Location", f.failureUrl)
		if writeErr := response.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, err))); writeErr != nil {
			fmt.Printf("[go-boot] failed to write authentication error response: %v\n", writeErr)
		}
		return nil
	}

	SetAuthentication(authenticated)

	response.SetStatusCode(302)
	response.SetHeader("Location", f.defaultSuccessUrl)
	if writeErr := response.Write([]byte(`{"message":"login success"}`)); writeErr != nil {
		fmt.Printf("[go-boot] failed to write login success response: %v\n", writeErr)
	}

	return nil
}

// BasicAuthenticationEntryPointWithRealm Basic认证入口点（带Realm）
type BasicAuthenticationEntryPointWithRealm struct {
	realmName string
}

// NewBasicAuthenticationEntryPointWithRealm 创建带Realm的Basic认证入口点
func NewBasicAuthenticationEntryPointWithRealm(realmName string) *BasicAuthenticationEntryPointWithRealm {
	return &BasicAuthenticationEntryPointWithRealm{
		realmName: realmName,
	}
}

// Commence 返回Basic认证失败响应，包含WWW-Authenticate头和Realm信息
func (e *BasicAuthenticationEntryPointWithRealm) Commence(ctx context.Context, request SecurityRequest, response SecurityResponse, err error) error {
	response.SetStatusCode(401)
	response.SetHeader("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, e.realmName))
	if writeErr := response.Write([]byte("Authentication required")); writeErr != nil {
		fmt.Printf("[go-boot] failed to write authentication required response: %v\n", writeErr)
	}

	if err == nil {
		return ErrBadCredentials
	}
	return err
}

// BasicAuthenticationFilterWithRealm 带Realm的Basic认证过滤器
type BasicAuthenticationFilterWithRealm struct {
	authenticationManager AuthenticationManager    // 认证管理器
	entryPoint            AuthenticationEntryPoint // 认证入口点
}

// NewBasicAuthenticationFilterWithRealm 创建带Realm的Basic认证过滤器
func NewBasicAuthenticationFilterWithRealm(authManager AuthenticationManager, realmName string) *BasicAuthenticationFilterWithRealm {
	return &BasicAuthenticationFilterWithRealm{
		authenticationManager: authManager,
		entryPoint:            NewBasicAuthenticationEntryPointWithRealm(realmName),
	}
}

// DoFilter 处理Basic认证
func (f *BasicAuthenticationFilterWithRealm) DoFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain SecurityFilterChain) error {
	authHeader := request.GetHeader("Authorization")
	if authHeader == "" {
		return f.entryPoint.Commence(ctx, request, response, ErrBadCredentials)
	}

	const prefix = "Basic "
	if len(authHeader) < len(prefix) || authHeader[:len(prefix)] != prefix {
		return f.entryPoint.Commence(ctx, request, response, ErrBadCredentials)
	}

	encoded := authHeader[len(prefix):]
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return f.entryPoint.Commence(ctx, request, response, ErrBadCredentials)
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
		return f.entryPoint.Commence(ctx, request, response, ErrBadCredentials)
	}

	username := credentials[:sepIndex]
	password := credentials[sepIndex+1:]

	authToken := NewUsernamePasswordAuthenticationToken(username, password)
	authenticated, err := f.authenticationManager.Authenticate(ctx, authToken)

	if err != nil {
		return f.entryPoint.Commence(ctx, request, response, err)
	}

	SetAuthentication(authenticated)

	return chain.DoFilter(ctx, request, response)
}

// SessionAuthenticationStrategy 会话认证策略接口
type SessionAuthenticationStrategy interface {
	OnAuthentication(ctx context.Context, authentication Authentication, request SecurityRequest, response SecurityResponse)
}

// CsrfAuthenticationStrategy CSRF令牌会话认证策略
type CsrfAuthenticationStrategy struct {
	mu           sync.Mutex
	tokenManager *CsrfTokenManager
}

// NewCsrfAuthenticationStrategy 创建CSRF认证策略
func NewCsrfAuthenticationStrategy() *CsrfAuthenticationStrategy {
	return &CsrfAuthenticationStrategy{
		tokenManager: NewCsrfTokenManager(),
	}
}

// OnAuthentication 认证成功后生成新的CSRF令牌
func (s *CsrfAuthenticationStrategy) OnAuthentication(ctx context.Context, authentication Authentication, request SecurityRequest, response SecurityResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()

	token := s.tokenManager.GenerateToken(authentication.Name())
	request.SetAttribute("csrf.token", token)
}

// CsrfTokenManager CSRF令牌管理器
type CsrfTokenManager struct {
	tokens map[string]string
	mu     sync.RWMutex
}

// NewCsrfTokenManager 创建CSRF令牌管理器
func NewCsrfTokenManager() *CsrfTokenManager {
	return &CsrfTokenManager{
		tokens: make(map[string]string),
	}
}

// GenerateToken 为用户生成CSRF令牌
func (m *CsrfTokenManager) GenerateToken(principal string) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	token := generateSecureToken(32)
	m.tokens[principal] = token
	return token
}

// ValidateToken 验证CSRF令牌
func (m *CsrfTokenManager) ValidateToken(principal, token string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	savedToken, exists := m.tokens[principal]
	if !exists {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(token), []byte(savedToken)) == 1
}

// RemoveToken 移除用户的CSRF令牌
func (m *CsrfTokenManager) RemoveToken(principal string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tokens, principal)
}
