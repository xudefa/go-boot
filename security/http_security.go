package security

import (
	"context"
	"fmt"
)

// httpSecurity HTTP安全配置实现
// 提供链式API配置HTTP安全规则
type httpSecurity struct {
	authenticationManager       AuthenticationManager          // 认证管理器
	userDetailsService          UserDetailsService             // 用户详情服务
	passwordEncoder             PasswordEncoder                // 密码编码器
	accessDecisionManager       AccessDecisionManager          // 访问决策管理器
	securityMetadataSource      SecurityMetadataSource         // 安全元数据源
	filters                     []SecurityFilter               // 自定义过滤器
	anonymousFilter             *AnonymousAuthenticationFilter // 匿名认证过滤器
	exceptionTranslationFilter  *ExceptionTranslationFilter    // 异常转换过滤器
	filterSecurityInterceptor   *FilterSecurityInterceptor     // 过滤器安全拦截器
	securityContextHolderFilter *SecurityContextHolderFilter   // 安全上下文持有者过滤器

	// CSRF配置
	csrfEnabled         bool                // 是否启用CSRF防护
	csrfTokenRepository CsrfTokenRepository // CSRF令牌仓库

	// Logout配置
	logoutUrl            string               // 登出URL
	logoutHandlers       []LogoutHandler      // 登出处理器
	logoutSuccessHandler LogoutSuccessHandler // 登出成功处理器

	// FormLogin配置
	formLoginEnabled   bool   // 是否启用表单登录
	loginProcessingUrl string // 登录处理URL
	defaultSuccessUrl  string // 默认成功URL
	failureUrl         string // 登录失败URL

	// HttpBasic配置
	httpBasicEnabled bool   // 是否启用HTTP Basic认证
	realmName        string // Realm名称
}

// NewHttpSecurity 创建新的HTTP安全配置
func NewHttpSecurity() HttpSecurity {
	return &httpSecurity{
		filters:             make([]SecurityFilter, 0),
		csrfTokenRepository: NewCookieCsrfTokenRepository(),
	}
}

// AuthenticationManager 设置认证管理器
func (h *httpSecurity) AuthenticationManager(authManager AuthenticationManager) HttpSecurity {
	h.authenticationManager = authManager
	return h
}

// UserDetailsService 设置用户详情服务
func (h *httpSecurity) UserDetailsService(userDetailsService UserDetailsService) HttpSecurity {
	h.userDetailsService = userDetailsService
	return h
}

// PasswordEncoder 设置密码编码器
func (h *httpSecurity) PasswordEncoder(encoder PasswordEncoder) HttpSecurity {
	h.passwordEncoder = encoder
	return h
}

// AccessDecisionManager 设置访问决策管理器
func (h *httpSecurity) AccessDecisionManager(manager AccessDecisionManager) HttpSecurity {
	h.accessDecisionManager = manager
	return h
}

// SecurityMetadataSource 设置安全元数据源
func (h *httpSecurity) SecurityMetadataSource(source SecurityMetadataSource) HttpSecurity {
	h.securityMetadataSource = source
	return h
}

// AuthorizeRequests 配置授权请求
func (h *httpSecurity) AuthorizeRequests(authorizer AuthorizeRequests) HttpSecurity {
	if httpSecurityAuthorizer, ok := authorizer.(*httpSecurityAuthorizer); ok {
		httpSecurityAuthorizer.httpSecurity = h
	}
	return h
}

// AddFilter 添加自定义过滤器
func (h *httpSecurity) AddFilter(filter SecurityFilter) HttpSecurity {
	h.filters = append(h.filters, filter)
	return h
}

// AddFilterBefore 在指定过滤器之前添加过滤器
func (h *httpSecurity) AddFilterBefore(filter SecurityFilter, beforeFilter SecurityFilter) HttpSecurity {
	newFilters := make([]SecurityFilter, 0, len(h.filters)+1)
	for _, f := range h.filters {
		if f == beforeFilter {
			newFilters = append(newFilters, filter)
		}
		newFilters = append(newFilters, f)
	}
	h.filters = newFilters
	return h
}

// AddFilterAfter 在指定过滤器之后添加过滤器
func (h *httpSecurity) AddFilterAfter(filter SecurityFilter, afterFilter SecurityFilter) HttpSecurity {
	newFilters := make([]SecurityFilter, 0, len(h.filters)+1)
	for _, f := range h.filters {
		newFilters = append(newFilters, f)
		if f == afterFilter {
			newFilters = append(newFilters, filter)
		}
	}
	h.filters = newFilters
	return h
}

// Anonymous 启用匿名认证
func (h *httpSecurity) Anonymous() HttpSecurity {
	h.anonymousFilter = NewAnonymousAuthenticationFilter()
	return h
}

// ExceptionHandling 配置异常处理
func (h *httpSecurity) ExceptionHandling(handler AccessDeniedHandler, entryPoint AuthenticationEntryPoint) HttpSecurity {
	h.exceptionTranslationFilter = NewExceptionTranslationFilter(handler, entryPoint)
	return h
}

// Csrf 启用CSRF防护
func (h *httpSecurity) Csrf() HttpSecurity {
	h.csrfEnabled = true
	return h
}

// Logout 配置登出
// 可选参数: logoutUrl 登出URL, successHandler 登出成功处理器
func (h *httpSecurity) Logout(logoutUrl string, successHandler ...LogoutSuccessHandler) HttpSecurity {
	h.logoutUrl = "/logout"
	if logoutUrl != "" {
		h.logoutUrl = logoutUrl
	}
	h.logoutHandlers = make([]LogoutHandler, 0)
	h.logoutSuccessHandler = nil
	if len(successHandler) > 0 {
		h.logoutSuccessHandler = successHandler[0]
	} else {
		h.logoutSuccessHandler = NewDefaultLogoutSuccessHandler("/login?logout")
	}
	return h
}

// FormLogin 配置表单登录
// 可选参数: loginProcessingUrl 登录处理URL, defaultSuccessUrl 默认成功URL, failureUrl 失败URL
func (h *httpSecurity) FormLogin(loginProcessingUrl string, defaultSuccessUrl ...string) HttpSecurity {
	h.formLoginEnabled = true
	h.loginProcessingUrl = "/login"
	h.defaultSuccessUrl = "/"
	h.failureUrl = "/login?error"

	if loginProcessingUrl != "" {
		h.loginProcessingUrl = loginProcessingUrl
	}
	if len(defaultSuccessUrl) > 0 && defaultSuccessUrl[0] != "" {
		h.defaultSuccessUrl = defaultSuccessUrl[0]
	}
	return h
}

// HttpBasic 配置HTTP Basic认证
func (h *httpSecurity) HttpBasic() HttpSecurity {
	h.httpBasicEnabled = true
	h.realmName = "Secured Area"
	return h
}

// Build 构建安全过滤器链
// 自动配置默认过滤器和安全组件
func (h *httpSecurity) Build() (SecurityFilterChain, error) {
	if h.authenticationManager == nil {
		return nil, fmt.Errorf("authentication manager is required")
	}

	if h.securityMetadataSource == nil {
		h.securityMetadataSource = NewExpressionBasedFilterInvocationSecurityMetadataSource()
	}

	if h.accessDecisionManager == nil {
		webExpressionVoter := NewWebExpressionVoter()
		authenticatedVoter := NewAuthenticatedVoter()
		roleVoter := NewRoleVoter()
		h.accessDecisionManager = NewAffirmativeBased(webExpressionVoter, authenticatedVoter, roleVoter)
	}

	h.securityContextHolderFilter = NewSecurityContextHolderFilter(GetSecurityContext())

	if h.anonymousFilter == nil {
		h.anonymousFilter = NewAnonymousAuthenticationFilter()
	}

	h.filterSecurityInterceptor = NewFilterSecurityInterceptor(
		h.securityMetadataSource,
		h.accessDecisionManager,
		h.authenticationManager,
	)

	if h.exceptionTranslationFilter == nil {
		accessDeniedHandler := NewHttp403ForbiddenAccessDeniedHandler()
		unauthorizedEntryPoint := NewHttp401UnauthorizedEntryPoint()
		h.exceptionTranslationFilter = NewExceptionTranslationFilter(accessDeniedHandler, unauthorizedEntryPoint)
	}

	// 构建默认过滤器列表
	defaultFilters := []SecurityFilter{
		h.securityContextHolderFilter,
		h.anonymousFilter,
	}

	// CSRF过滤器（如果启用）
	if h.csrfEnabled {
		csrfFilter := NewCsrfFilter(h.csrfTokenRepository)
		defaultFilters = append(defaultFilters, csrfFilter)
	}

	// Logout过滤器（如果配置了）
	if h.logoutUrl != "" {
		logoutFilter := NewLogoutFilter(h.logoutUrl, h.logoutHandlers)
		if h.logoutSuccessHandler != nil {
			logoutFilter.SetSuccessHandler(h.logoutSuccessHandler)
		}
		defaultFilters = append(defaultFilters, logoutFilter)
	}

	// FormLogin过滤器（如果启用）
	if h.formLoginEnabled {
		formLoginFilter := NewUsernamePasswordAuthenticationFilterWithDefaults(
			h.loginProcessingUrl,
			h.defaultSuccessUrl,
			h.failureUrl,
			h.authenticationManager,
		)
		defaultFilters = append(defaultFilters, formLoginFilter)
	}

	// HttpBasic过滤器（如果启用）
	if h.httpBasicEnabled {
		basicFilter := NewBasicAuthenticationFilterWithRealm(h.authenticationManager, h.realmName)
		defaultFilters = append(defaultFilters, basicFilter)
	}

	defaultFilters = append(defaultFilters, h.exceptionTranslationFilter)
	defaultFilters = append(defaultFilters, h.filterSecurityInterceptor)

	allFilters := make([]SecurityFilter, 0, len(defaultFilters)+len(h.filters))
	allFilters = append(allFilters, defaultFilters...)
	allFilters = append(allFilters, h.filters...)

	return NewFilterChainProxy(allFilters, &DefaultSecurityFilterChain{}), nil
}

// DefaultSecurityFilterChain 默认安全过滤器链
// 当所有过滤器都执行完毕后调用
type DefaultSecurityFilterChain struct{}

func (c *DefaultSecurityFilterChain) DoFilter(ctx context.Context, request SecurityRequest, response SecurityResponse) error {
	return nil
}

// httpSecurityAuthorizer HTTP安全授权器
// 实现AuthorizeRequests接口
type httpSecurityAuthorizer struct {
	httpSecurity *httpSecurity
}

func (a *httpSecurityAuthorizer) AntMatchers(patterns ...string) ExpressionInterceptUrlRegistry {
	return &expressionInterceptUrlRegistry{
		httpSecurity: a.httpSecurity,
		patterns:     patterns,
	}
}

func (a *httpSecurityAuthorizer) AnyRequest() ExpressionInterceptUrlRegistry {
	return &expressionInterceptUrlRegistry{
		httpSecurity: a.httpSecurity,
		patterns:     []string{"/**"},
	}
}

// expressionInterceptUrlRegistry URL拦截注册
// 配置特定URL的访问规则
type expressionInterceptUrlRegistry struct {
	httpSecurity *httpSecurity // HTTP安全配置
	patterns     []string      // URL模式列表
}

// PermitAll 允许所有用户访问
func (r *expressionInterceptUrlRegistry) PermitAll() HttpSecurity {
	for _, pattern := range r.patterns {
		r.httpSecurity.securityMetadataSource.(*ExpressionBasedFilterInvocationSecurityMetadataSource).AddMapping(
			pattern,
			[]string{"permitAll"},
		)
	}
	return r.httpSecurity
}

// Authenticated 要求用户已认证
func (r *expressionInterceptUrlRegistry) Authenticated() HttpSecurity {
	for _, pattern := range r.patterns {
		r.httpSecurity.securityMetadataSource.(*ExpressionBasedFilterInvocationSecurityMetadataSource).AddMapping(
			pattern,
			[]string{"authenticated"},
		)
	}
	return r.httpSecurity
}

// HasRole 要求具有指定角色
func (r *expressionInterceptUrlRegistry) HasRole(role string) HttpSecurity {
	for _, pattern := range r.patterns {
		r.httpSecurity.securityMetadataSource.(*ExpressionBasedFilterInvocationSecurityMetadataSource).AddMapping(
			pattern,
			[]string{fmt.Sprintf("hasRole('%s')", role)},
		)
	}
	return r.httpSecurity
}

// HasAnyRole 要求具有任一指定角色
func (r *expressionInterceptUrlRegistry) HasAnyRole(roles ...string) HttpSecurity {
	for _, pattern := range r.patterns {
		rolesStr := ""
		for i, role := range roles {
			if i > 0 {
				rolesStr += "','"
			}
			rolesStr += role
		}
		r.httpSecurity.securityMetadataSource.(*ExpressionBasedFilterInvocationSecurityMetadataSource).AddMapping(
			pattern,
			[]string{fmt.Sprintf("hasAnyRole('%s')", rolesStr)},
		)
	}
	return r.httpSecurity
}

// HasAuthority 要求具有指定权限
func (r *expressionInterceptUrlRegistry) HasAuthority(authority string) HttpSecurity {
	for _, pattern := range r.patterns {
		r.httpSecurity.securityMetadataSource.(*ExpressionBasedFilterInvocationSecurityMetadataSource).AddMapping(
			pattern,
			[]string{fmt.Sprintf("hasAuthority('%s')", authority)},
		)
	}
	return r.httpSecurity
}

// HasAnyAuthority 要求具有任一指定权限
func (r *expressionInterceptUrlRegistry) HasAnyAuthority(authorities ...string) HttpSecurity {
	for _, pattern := range r.patterns {
		authStr := ""
		for i, auth := range authorities {
			if i > 0 {
				authStr += "','"
			}
			authStr += auth
		}
		r.httpSecurity.securityMetadataSource.(*ExpressionBasedFilterInvocationSecurityMetadataSource).AddMapping(
			pattern,
			[]string{fmt.Sprintf("hasAnyAuthority('%s')", authStr)},
		)
	}
	return r.httpSecurity
}

// DenyAll 拒绝所有用户访问
func (r *expressionInterceptUrlRegistry) DenyAll() HttpSecurity {
	for _, pattern := range r.patterns {
		r.httpSecurity.securityMetadataSource.(*ExpressionBasedFilterInvocationSecurityMetadataSource).AddMapping(
			pattern,
			[]string{"denyAll"},
		)
	}
	return r.httpSecurity
}

// WebSecurity Web安全配置入口
// 提供创建HTTP安全配置的工厂方法
type WebSecurity struct{}

// NewWebSecurity 创建Web安全配置入口
func NewWebSecurity() *WebSecurity {
	return &WebSecurity{}
}

// HttpSecurity 创建HTTP安全配置
func (w *WebSecurity) HttpSecurity() *httpSecurity {
	return &httpSecurity{
		filters: make([]SecurityFilter, 0),
	}
}

// Build 构建安全过滤器链（已废弃，请使用HttpSecurity().Build()）
func (w *WebSecurity) Build() (SecurityFilterChain, error) {
	return nil, fmt.Errorf("use HttpSecurity().Build() instead")
}
