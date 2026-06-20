package security

import (
	"context"
	"fmt"
	"strings"
)

// FilterChainProxy 过滤器链代理
// 管理和执行多个安全过滤器
type FilterChainProxy struct {
	filters []SecurityFilter
	chain   SecurityFilterChain
}

// NewFilterChainProxy 创建过滤器链代理
func NewFilterChainProxy(filters []SecurityFilter, chain SecurityFilterChain) *FilterChainProxy {
	return &FilterChainProxy{
		filters: filters,
		chain:   chain,
	}
}

// DoFilter 执行过滤器链
func (p *FilterChainProxy) DoFilter(ctx context.Context, request SecurityRequest, response SecurityResponse) error {
	return p.doFilterInternal(ctx, request, response, 0)
}

func (p *FilterChainProxy) doFilterInternal(ctx context.Context, request SecurityRequest, response SecurityResponse, index int) error {
	if index >= len(p.filters) {
		return p.chain.DoFilter(ctx, request, response)
	}

	filter := p.filters[index]
	return filter.DoFilter(ctx, request, response, &VirtualFilterChain{
		proxy: p,
		index: index + 1,
	})
}

// VirtualFilterChain 虚拟过滤器链
// 在过滤器执行过程中跟踪当前索引，实现链式调用
type VirtualFilterChain struct {
	proxy *FilterChainProxy // 过滤器链代理
	index int               // 当前过滤器索引
}

// DoFilter 执行下一个过滤器
func (c *VirtualFilterChain) DoFilter(ctx context.Context, request SecurityRequest, response SecurityResponse) error {
	return c.proxy.doFilterInternal(ctx, request, response, c.index)
}

// SecurityContextHolderFilter 安全上下文持有者过滤器
// 在请求开始时清除全局认证信息，请求结束后恢复
// 过滤器链内的其他过滤器通过 SetAuthentication 设置认证信息到全局上下文，
// 本过滤器负责在请求结束后清理，防止认证信息在请求之间泄漏
type SecurityContextHolderFilter struct {
	securityContext SecurityContext
}

// NewSecurityContextHolderFilter 创建安全上下文持有者过滤器
func NewSecurityContextHolderFilter(securityContext SecurityContext) *SecurityContextHolderFilter {
	return &SecurityContextHolderFilter{
		securityContext: securityContext,
	}
}

// DoFilter 执行过滤器链
// 在请求前保存全局认证信息并清除，请求结束后恢复之前的认证信息
// 当前请求的认证信息保存到请求属性中，供 SecurityFilterChainHandler 传递给 handler
func (f *SecurityContextHolderFilter) DoFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain SecurityFilterChain) error {
	prevAuth := GetAuthentication()
	ClearAuthentication()

	err := chain.DoFilter(ctx, request, response)

	currentAuth := GetAuthentication()
	if currentAuth != nil {
		request.SetAttribute("security.currentAuthentication", currentAuth)
	}

	if f.securityContext != nil {
		f.securityContext.ClearAuthentication()
		if prevAuth != nil {
			f.securityContext.SetAuthentication(prevAuth)
		}
	}

	return err
}

// AnonymousAuthenticationFilter 匿名认证过滤器
// 为未认证的用户创建匿名认证令牌
type AnonymousAuthenticationFilter struct {
	key         string
	principal   any
	authorities []string
}

// NewAnonymousAuthenticationFilter 创建匿名认证过滤器
func NewAnonymousAuthenticationFilter() *AnonymousAuthenticationFilter {
	return &AnonymousAuthenticationFilter{
		key:         "anonymousKey",
		principal:   "anonymousUser",
		authorities: []string{"ROLE_ANONYMOUS"},
	}
}

// DoFilter 如果当前没有认证则创建匿名认证
func (f *AnonymousAuthenticationFilter) DoFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain SecurityFilterChain) error {
	if GetAuthentication() == nil {
		auth := NewAnonymousAuthenticationToken(f.key, f.principal, f.authorities)
		SetAuthentication(auth)
	}
	return chain.DoFilter(ctx, request, response)
}

// AnonymousAuthenticationToken 匿名认证令牌
type AnonymousAuthenticationToken struct {
	principal     any
	authorities   []string
	authenticated bool
}

// NewAnonymousAuthenticationToken 创建匿名认证令牌
func NewAnonymousAuthenticationToken(key string, principal any, authorities []string) *AnonymousAuthenticationToken {
	return &AnonymousAuthenticationToken{
		principal:     principal,
		authorities:   authorities,
		authenticated: false,
	}
}

// Principal 返回认证主体
func (t *AnonymousAuthenticationToken) Principal() any {
	return t.principal
}

// Credentials 返回凭证
func (t *AnonymousAuthenticationToken) Credentials() any {
	return nil
}

// Authorities 返回授权列表
func (t *AnonymousAuthenticationToken) Authorities() []string {
	return t.authorities
}

// Authenticated 返回是否已认证
func (t *AnonymousAuthenticationToken) Authenticated() bool {
	return t.authenticated
}

// Name 返回认证主体名称
func (t *AnonymousAuthenticationToken) Name() string {
	if name, ok := t.principal.(string); ok {
		return name
	}
	return ""
}

// ExceptionTranslationFilter 异常转换过滤器
// 将安全异常转换为适当的HTTP响应
type ExceptionTranslationFilter struct {
	accessDeniedHandler      AccessDeniedHandler
	authenticationEntryPoint AuthenticationEntryPoint
}

// NewExceptionTranslationFilter 创建异常转换过滤器
func NewExceptionTranslationFilter(accessDeniedHandler AccessDeniedHandler, authenticationEntryPoint AuthenticationEntryPoint) *ExceptionTranslationFilter {
	return &ExceptionTranslationFilter{
		accessDeniedHandler:      accessDeniedHandler,
		authenticationEntryPoint: authenticationEntryPoint,
	}
}

// DoFilter 捕获安全异常并转换为HTTP响应
func (f *ExceptionTranslationFilter) DoFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain SecurityFilterChain) error {
	err := chain.DoFilter(ctx, request, response)
	if err != nil {
		if err == ErrAccessDenied {
			auth := GetAuthentication()
			if auth == nil || !auth.Authenticated() {
				if f.authenticationEntryPoint != nil {
					return f.authenticationEntryPoint.Commence(ctx, request, response, err)
				}
				return err
			}
			if f.accessDeniedHandler != nil {
				return f.accessDeniedHandler.Handle(ctx, request, response, err)
			}
		}
		return err
	}
	return nil
}

// FilterSecurityInterceptor 过滤器安全拦截器
// 执行访问决策检查，决定是否允许请求继续
type FilterSecurityInterceptor struct {
	securityMetadataSource SecurityMetadataSource // 安全元数据源
	accessDecisionManager  AccessDecisionManager  // 访问决策管理器
	authenticationManager  AuthenticationManager  // 认证管理器
	observeOncePerRequest  bool                   // 是否每个请求只执行一次
}

// NewFilterSecurityInterceptor 创建过滤器安全拦截器
func NewFilterSecurityInterceptor(securityMetadataSource SecurityMetadataSource, accessDecisionManager AccessDecisionManager, authenticationManager AuthenticationManager) *FilterSecurityInterceptor {
	return &FilterSecurityInterceptor{
		securityMetadataSource: securityMetadataSource,
		accessDecisionManager:  accessDecisionManager,
		authenticationManager:  authenticationManager,
		observeOncePerRequest:  true,
	}
}

// DoFilter 执行访问决策
func (f *FilterSecurityInterceptor) DoFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain SecurityFilterChain) error {
	attributes, err := f.securityMetadataSource.GetAttributes(ctx, request)
	if err != nil {
		return err
	}

	if f.observeOncePerRequest {
		attrKey := "FILTER_APPLIED"
		if _, exists := request.GetAttribute(attrKey); exists {
			return chain.DoFilter(ctx, request, response)
		}
		request.SetAttribute(attrKey, true)
	}

	auth := GetAuthentication()

	if len(attributes) == 0 {
		return chain.DoFilter(ctx, request, response)
	}

	if err := f.accessDecisionManager.Decide(ctx, auth, request, attributes); err != nil {
		return err
	}

	return chain.DoFilter(ctx, request, response)
}

func (f *FilterSecurityInterceptor) SetSecurityMetadataSource(source SecurityMetadataSource) {
	f.securityMetadataSource = source
}

func (f *FilterSecurityInterceptor) SetAccessDecisionManager(manager AccessDecisionManager) {
	f.accessDecisionManager = manager
}

func (f *FilterSecurityInterceptor) SetAuthenticationManager(manager AuthenticationManager) {
	f.authenticationManager = manager
}

// ExpressionBasedFilterInvocationSecurityMetadataSource 基于表达式的过滤器调用安全元数据源
// 使用Ant风格路径模式配置URL安全属性
type ExpressionBasedFilterInvocationSecurityMetadataSource struct {
	requestMap map[string][]string
}

// NewExpressionBasedFilterInvocationSecurityMetadataSource 创建元数据源
func NewExpressionBasedFilterInvocationSecurityMetadataSource() *ExpressionBasedFilterInvocationSecurityMetadataSource {
	return &ExpressionBasedFilterInvocationSecurityMetadataSource{
		requestMap: make(map[string][]string),
	}
}

// AddMapping 添加URL路径和安全属性映射
// 格式: "GET::/path" 或 "/path" (默认匹配所有HTTP方法)
// 示例: AddMapping("/admin/**", []string{"hasRole('ADMIN')"})
func (s *ExpressionBasedFilterInvocationSecurityMetadataSource) AddMapping(pattern string, attributes []string) {
	if !strings.Contains(pattern, "::") {
		pattern = "**::" + pattern
	}
	s.requestMap[pattern] = attributes
}

// GetAttributes 获取请求对应的安全属性
func (s *ExpressionBasedFilterInvocationSecurityMetadataSource) GetAttributes(ctx context.Context, request SecurityRequest) ([]string, error) {
	uri := request.GetURI()
	method := request.GetMethod()

	for pattern, attributes := range s.requestMap {
		if s.matches(pattern, uri, method) {
			return attributes, nil
		}
	}

	return []string{}, nil
}

// matches 检查URL和方法是否匹配模式
// 支持: /**, /path/*, /path/**
func (s *ExpressionBasedFilterInvocationSecurityMetadataSource) matches(pattern, uri, method string) bool {
	parts := strings.Split(pattern, "::")
	if len(parts) != 2 {
		return false
	}

	methodPattern := parts[0]
	pathPattern := parts[1]

	if methodPattern != method && methodPattern != "**" {
		return false
	}

	if pathPattern == "/**" {
		return true
	}

	if pathPattern == uri {
		return true
	}

	if strings.HasSuffix(pathPattern, "/**") {
		prefix := strings.TrimSuffix(pathPattern, "/**")
		return strings.HasPrefix(uri, prefix) || uri == strings.TrimSuffix(prefix, "/")
	}

	if strings.Contains(pathPattern, "*") {
		return s.matchWildcard(pathPattern, uri)
	}

	return false
}

func (s *ExpressionBasedFilterInvocationSecurityMetadataSource) matchWildcard(pattern, str string) bool {
	patternParts := strings.Split(pattern, "/")
	strParts := strings.Split(str, "/")

	if len(patternParts) != len(strParts) {
		return false
	}

	for i, patternPart := range patternParts {
		if patternPart != "*" && patternPart != strParts[i] {
			return false
		}
	}

	return true
}

// Http403ForbiddenEntryPoint 403禁止访问入口点
type Http403ForbiddenEntryPoint struct{}

// NewHttp403ForbiddenEntryPoint 创建403入口点
func NewHttp403ForbiddenEntryPoint() *Http403ForbiddenEntryPoint {
	return &Http403ForbiddenEntryPoint{}
}

// Commence 返回403状态码
func (e *Http403ForbiddenEntryPoint) Commence(ctx context.Context, request SecurityRequest, response SecurityResponse, err error) error {
	response.SetStatusCode(403)
	if writeErr := response.Write([]byte("403 Forbidden")); writeErr != nil {
		fmt.Printf("[go-boot] failed to write 403 response: %v\n", writeErr)
	}
	return nil
}

// Http401UnauthorizedEntryPoint 401未认证入口点
type Http401UnauthorizedEntryPoint struct{}

// NewHttp401UnauthorizedEntryPoint 创建401入口点
func NewHttp401UnauthorizedEntryPoint() *Http401UnauthorizedEntryPoint {
	return &Http401UnauthorizedEntryPoint{}
}

// Commence 返回401状态码
func (e *Http401UnauthorizedEntryPoint) Commence(ctx context.Context, request SecurityRequest, response SecurityResponse, err error) error {
	response.SetStatusCode(401)
	if writeErr := response.Write([]byte("401 Unauthorized")); writeErr != nil {
		fmt.Printf("[go-boot] failed to write 401 response: %v\n", writeErr)
	}
	return nil
}

// Http403ForbiddenAccessDeniedHandler 403禁止访问拒绝处理器
type Http403ForbiddenAccessDeniedHandler struct{}

// NewHttp403ForbiddenAccessDeniedHandler 创建403拒绝处理器
func NewHttp403ForbiddenAccessDeniedHandler() *Http403ForbiddenAccessDeniedHandler {
	return &Http403ForbiddenAccessDeniedHandler{}
}

// Handle 处理访问拒绝
func (e *Http403ForbiddenAccessDeniedHandler) Handle(ctx context.Context, request SecurityRequest, response SecurityResponse, err error) error {
	response.SetStatusCode(403)
	if writeErr := response.Write([]byte("403 Forbidden")); writeErr != nil {
		fmt.Printf("[go-boot] failed to write 403 access denied response: %v\n", writeErr)
	}
	return nil
}

// LoginUrlAuthenticationEntryPoint 登录URL认证入口点
// 未认证用户访问受保护资源时重定向到登录页
type LoginUrlAuthenticationEntryPoint struct {
	loginFormUrl string // 登录表单URL
}

// NewLoginUrlAuthenticationEntryPoint 创建登录入口点
func NewLoginUrlAuthenticationEntryPoint(loginFormUrl string) *LoginUrlAuthenticationEntryPoint {
	return &LoginUrlAuthenticationEntryPoint{
		loginFormUrl: loginFormUrl,
	}
}

// Commence 重定向到登录页
func (e *LoginUrlAuthenticationEntryPoint) Commence(ctx context.Context, request SecurityRequest, response SecurityResponse, err error) error {
	response.SetStatusCode(302)
	response.SetHeader("Location", e.loginFormUrl)
	return nil
}
