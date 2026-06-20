package security

import (
	"fmt"
	"strings"

	"github.com/xudefa/go-boot/boot"
	"github.com/xudefa/go-boot/condition"
	"github.com/xudefa/go-boot/constants"
	"github.com/xudefa/go-boot/core"
	"github.com/xudefa/go-boot/environment"
)

func init() {
	boot.RegisterAutoConfig(&SecurityAutoConfiguration{},
		condition.OnProperty(constants.SecurityEnabled, constants.ConditionTrue),
	)
}

// SecurityAutoConfiguration 安全模块自动配置
// 当 security.enabled=true 时自动装配以下组件：
// - UserDetailsService: 默认使用内存存储
// - PasswordEncoder: 默认使用BCrypt编码器
// - AuthenticationManager: 包含DaoAuthenticationProvider和AnonymousAuthenticationProvider
// - SecurityFilterChain: 默认安全过滤器链（支持CORS、限流）
type SecurityAutoConfiguration struct{}

// Configure 执行自动配置
func (c *SecurityAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	container := ctx.Container()
	env := ctx.Environment()

	// 获取或注册 UserDetailsService
	userDetailsService := c.getUserDetailsService(ctx, container)

	// 获取或注册 PasswordEncoder
	passwordEncoder := c.getPasswordEncoder(ctx, container)

	// 注册 AuthenticationManager
	authManager := c.buildAuthenticationManager(userDetailsService, passwordEncoder)
	if regErr := ctx.Register(constants.AuthenticationManagerBeanID, core.Bean(authManager), core.Singleton()); regErr != nil {
		return regErr
	}

	filterChain := c.getOrBuildSecurityFilterChain(ctx, authManager, userDetailsService, env, container)

	if regErr := ctx.Register(constants.SecurityFilterChainBeanID, core.Bean(filterChain), core.Singleton()); regErr != nil {
		return regErr
	}

	return nil
}

func (c *SecurityAutoConfiguration) getUserDetailsService(ctx boot.ApplicationContext, container core.Container) UserDetailsService {
	beans, err := container.GetAll((*UserDetailsService)(nil))
	if err != nil || len(beans) == 0 {
		service := NewInMemoryUserDetailsService()
		if regErr := ctx.Register(constants.UserDetailsServiceBeanID, core.Bean(service), core.Singleton()); regErr != nil {
			fmt.Printf("[go-boot] failed to register UserDetailsService bean: %v\n", regErr)
		}
		return service
	}
	return beans[0].(UserDetailsService)
}

func (c *SecurityAutoConfiguration) getPasswordEncoder(ctx boot.ApplicationContext, container core.Container) PasswordEncoder {
	beans, err := container.GetAll((*PasswordEncoder)(nil))
	if err != nil || len(beans) == 0 {
		encoder := NewBCryptPasswordEncoder()
		if regErr := ctx.Register(constants.PasswordEncoderBeanID, core.Bean(encoder), core.Singleton()); regErr != nil {
			fmt.Printf("[go-boot] failed to register PasswordEncoder bean: %v\n", regErr)
		}
		return encoder
	}
	return beans[0].(PasswordEncoder)
}

// buildAuthenticationManager 构建 AuthenticationManager
func (c *SecurityAutoConfiguration) buildAuthenticationManager(userDetailsService UserDetailsService, passwordEncoder PasswordEncoder) AuthenticationManager {
	daoAuthProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder)
	anonymousProvider := NewAnonymousAuthenticationProvider()
	return NewProviderManager(daoAuthProvider, anonymousProvider)
}

// getOrBuildSecurityFilterChain 获取或构建 SecurityFilterChain
func (c *SecurityAutoConfiguration) getOrBuildSecurityFilterChain(ctx boot.ApplicationContext, authManager AuthenticationManager, userDetailsService UserDetailsService, env *environment.Environment, container core.Container) SecurityFilterChain {
	beans, err := container.GetAll((*SecurityFilterChain)(nil))
	if err != nil || len(beans) == 0 {
		return c.buildSecurityFilterChain(authManager, userDetailsService, env, container)
	}
	// 如果用户注入的过滤器链中不存在CORS、限流过滤器，则添加默认的
	return c.addDefaultFiltersIfMissing(beans[0].(SecurityFilterChain), env, container)
}

// buildSecurityFilterChain 构建安全过滤器链
func (c *SecurityAutoConfiguration) buildSecurityFilterChain(authManager AuthenticationManager, userDetailsService UserDetailsService, env *environment.Environment, container core.Container) SecurityFilterChain {
	filters := []SecurityFilter{}

	// 添加CORS过滤器
	if corsFilter := c.getOrCreateCorsFilter(env, container); corsFilter != nil {
		filters = append(filters, corsFilter)
	}

	// 添加限流过滤器
	if rateLimitFilter := c.getOrCreateRateLimitFilter(env, container); rateLimitFilter != nil {
		filters = append(filters, rateLimitFilter)
	}

	// 添加安全上下文持有者过滤器（必须在最前面，保存初始状态）
	securityContext := GetSecurityContext()
	contextHolderFilter := NewSecurityContextHolderFilter(securityContext)
	filters = append(filters, contextHolderFilter)

	// 添加JWT认证过滤器（如果容器中存在）
	if jwtFilter := c.getJwtAuthenticationFilter(container); jwtFilter != nil {
		filters = append(filters, jwtFilter)
	}

	// 添加其他默认安全过滤器
	defaultFilters := c.buildDefaultSecurityFilters(authManager, env, true) // skipContextFilter=true
	filters = append(filters, defaultFilters...)

	return NewFilterChainProxy(filters, &DefaultSecurityFilterChain{})
}

// getJwtAuthenticationFilter 从容器中获取JWT认证过滤器
func (c *SecurityAutoConfiguration) getJwtAuthenticationFilter(container core.Container) SecurityFilter {
	// 使用类型断言查找JWT过滤器
	beans, err := container.GetAll((*SecurityFilter)(nil))
	if err != nil || len(beans) == 0 {
		return nil
	}
	for _, bean := range beans {
		if jwtFilter, ok := bean.(SecurityFilter); ok {
			// 检查是否是JWT过滤器（通过类型断言）
			if _, isJwt := jwtFilter.(interface{ IsJwtFilter() bool }); isJwt {
				return jwtFilter
			}
		}
	}
	return nil
}

// buildDefaultSecurityFilters 构建默认安全过滤器（不含CORS和限流）
func (c *SecurityAutoConfiguration) buildDefaultSecurityFilters(authManager AuthenticationManager, env *environment.Environment, skipContextFilter ...bool) []SecurityFilter {
	filters := []SecurityFilter{}

	skip := false
	if len(skipContextFilter) > 0 {
		skip = skipContextFilter[0]
	}

	if !skip {
		securityContext := GetSecurityContext()
		contextHolderFilter := NewSecurityContextHolderFilter(securityContext)
		filters = append(filters, contextHolderFilter)
	}

	anonymousFilter := NewAnonymousAuthenticationFilter()
	filters = append(filters, anonymousFilter)

	webExpressionVoter := NewWebExpressionVoter()
	authenticatedVoter := NewAuthenticatedVoter()
	roleVoter := NewRoleVoter()
	accessDecisionManager := NewAffirmativeBased(webExpressionVoter, authenticatedVoter, roleVoter)

	metadataSource := NewExpressionBasedFilterInvocationSecurityMetadataSource()
	if rules := env.GetString("security.rules", ""); rules != "" {
		for _, rule := range parseStringSlice(rules) {
			parts := strings.SplitN(rule, "->", 2)
			if len(parts) == 2 {
				pattern := strings.TrimSpace(parts[0])
				expression := strings.TrimSpace(parts[1])
				metadataSource.AddMapping(pattern, []string{expression})
			}
		}
	}

	accessDeniedHandler := NewHttp403ForbiddenAccessDeniedHandler()
	unauthorizedEntryPoint := NewHttp401UnauthorizedEntryPoint()
	exceptionTranslationFilter := NewExceptionTranslationFilter(accessDeniedHandler, unauthorizedEntryPoint)
	filters = append(filters, exceptionTranslationFilter)

	filterSecurityInterceptor := NewFilterSecurityInterceptor(
		metadataSource,
		accessDecisionManager,
		authManager,
	)
	filters = append(filters, filterSecurityInterceptor)

	return filters
}

// getOrCreateCorsFilter 获取或创建CORS过滤器
func (c *SecurityAutoConfiguration) getOrCreateCorsFilter(env *environment.Environment, container core.Container) *CorsFilter {
	if !env.GetBool("security.cors.enabled", false) {
		return nil
	}
	// 优先从容器获取
	beans, _ := container.GetAll((*CorsFilter)(nil))
	if len(beans) > 0 {
		return beans[0].(*CorsFilter)
	}
	// 创建默认的CORS过滤器
	return c.createDefaultCorsFilter(env)
}

// createDefaultCorsFilter 创建默认的CORS过滤器
func (c *SecurityAutoConfiguration) createDefaultCorsFilter(env *environment.Environment) *CorsFilter {
	corsConfig := CorsConfig{
		AllowedOrigins:   parseStringSlice(env.GetString("security.cors.allowed-origins", "*")),
		AllowedMethods:   parseStringSlice(env.GetString("security.cors.allowed-methods", "GET,POST,PUT,DELETE,OPTIONS")),
		AllowedHeaders:   parseStringSlice(env.GetString("security.cors.allowed-headers", "Content-Type,Authorization,X-Requested-With")),
		ExposedHeaders:   parseStringSlice(env.GetString("security.cors.exposed-headers", "")),
		AllowCredentials: env.GetBool("security.cors.allow-credentials", false),
		MaxAge:           env.GetInt("security.cors.max-age", 3600),
	}
	return NewCorsFilter(corsConfig)
}

// getOrCreateRateLimitFilter 获取或创建限流过滤器
func (c *SecurityAutoConfiguration) getOrCreateRateLimitFilter(env *environment.Environment, container core.Container) *RateLimitFilter {
	if !env.GetBool("security.rate-limit.enabled", false) {
		return nil
	}
	// 优先从容器获取
	beans, _ := container.GetAll((*RateLimitFilter)(nil))
	if len(beans) > 0 {
		return beans[0].(*RateLimitFilter)
	}
	// 创建默认的限流过滤器
	return c.createDefaultRateLimitFilter(env)
}

// createDefaultRateLimitFilter 创建默认的限流过滤器
func (c *SecurityAutoConfiguration) createDefaultRateLimitFilter(env *environment.Environment) *RateLimitFilter {
	rateLimitConfig := RateLimitConfig{
		Enabled:      true,
		Rate:         env.GetInt("security.rate-limit.rate", 100),
		Burst:        env.GetInt("security.rate-limit.burst", 200),
		ExcludePaths: parseStringSlice(env.GetString("security.rate-limit.exclude-paths", "/health,/actuator/health")),
	}
	return NewRateLimitFilter(rateLimitConfig)
}

// addDefaultFiltersIfMissing 如果用户注入的过滤器链中不存在CORS、限流过滤器，则添加默认的
func (c *SecurityAutoConfiguration) addDefaultFiltersIfMissing(filterChain SecurityFilterChain, env *environment.Environment, container core.Container) SecurityFilterChain {
	// 尝试将用户的过滤器链转换为 FilterChainProxy
	proxy, ok := filterChain.(*FilterChainProxy)
	if !ok {
		return filterChain
	}

	// 检查是否已存在CORS过滤器
	hasCors := c.hasFilter(proxy.filters, (*CorsFilter)(nil))

	// 检查是否已存在限流过滤器
	hasRateLimit := c.hasFilter(proxy.filters, (*RateLimitFilter)(nil))

	// 如果不存在CORS过滤器且配置启用了CORS，则添加默认的
	if !hasCors && env.GetBool("security.cors.enabled", false) {
		corsFilter := c.getOrCreateCorsFilter(env, container)
		if corsFilter != nil {
			// 在最前面插入CORS过滤器
			proxy.filters = append([]SecurityFilter{corsFilter}, proxy.filters...)
		}
	}

	// 如果不存在限流过滤器且配置启用了限流，则添加默认的
	if !hasRateLimit && env.GetBool("security.rate-limit.enabled", false) {
		rateLimitFilter := c.getOrCreateRateLimitFilter(env, container)
		if rateLimitFilter != nil {
			// 在CORS过滤器之后（如果存在）或最前面插入限流过滤器
			insertIndex := 0
			if hasCors || (!hasCors && env.GetBool("security.cors.enabled", false)) {
				insertIndex = 1
			}
			proxy.filters = append(proxy.filters[:insertIndex], append([]SecurityFilter{rateLimitFilter}, proxy.filters[insertIndex:]...)...)
		}
	}

	return proxy
}

// hasFilter 检查过滤器列表中是否包含指定类型的过滤器
func (c *SecurityAutoConfiguration) hasFilter(filters []SecurityFilter, target interface{}) bool {
	for _, f := range filters {
		switch target.(type) {
		case *CorsFilter:
			if _, ok := f.(*CorsFilter); ok {
				return true
			}
		case *RateLimitFilter:
			if _, ok := f.(*RateLimitFilter); ok {
				return true
			}
		}
	}
	return false
}

func parseStringSlice(value string) []string {
	if value == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
