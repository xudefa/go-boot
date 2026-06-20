// Package security 提供安全框架的核心接口和实现。
//
// 基于 Spring Security 设计理念，提供认证（Authentication）、授权（Authorization）、
// 安全过滤器链（SecurityFilterChain）等安全机制。
//
// 核心接口：
//   - SecurityContext: 安全上下文，管理当前认证信息
//   - Authentication: 认证信息，包含主体标识、凭证和权限
//   - AuthenticationManager: 认证管理器，处理认证请求
//   - AccessDecisionManager: 访问决策管理器，决定是否允许访问
//   - UserDetailsService: 用户详情服务，加载用户信息
//   - SecurityFilter/SecurityFilterChain: 安全过滤器及链
//   - HttpSecurity: HTTP 安全配置器，提供链式 API
//
// 提供全局安全上下文（globalSecurityContext）和基于 context.Context 的认证信息传递。
package security

import (
	"context"
	"errors"
	"sync"
)

// 错误定义
var (
	ErrAuthenticationFailed = errors.New("authentication failed") // 认证失败
	ErrAccessDenied         = errors.New("access denied")         // 访问被拒绝
	ErrUserNotFound         = errors.New("user not found")        // 用户未找到
	ErrBadCredentials       = errors.New("bad credentials")       // 凭证无效
)

// SecurityContext 安全上下文接口
// 提供对当前认证信息的访问和管理
type SecurityContext interface {
	Authentication() Authentication
	SetAuthentication(auth Authentication)
	ClearAuthentication()
}

// Authentication 认证信息接口
// 代表一个已认证主体的完整信息
type Authentication interface {
	// Principal 返回认证主体的标识，通常是用户名
	Principal() any
	// Credentials 返回凭证信息，如密码
	Credentials() any
	// Authorities 返回授权列表，如角色和权限
	Authorities() []string
	// Authenticated 是否已认证
	Authenticated() bool
	// Name 返回认证主体的名称
	Name() string
}

// AuthenticationManager 认证管理器接口
// 负责处理认证请求
type AuthenticationManager interface {
	Authenticate(ctx context.Context, authentication Authentication) (Authentication, error)
}

// AccessDecisionManager 访问决策管理器接口
// 决定是否允许访问受保护的资源
type AccessDecisionManager interface {
	Decide(ctx context.Context, authentication Authentication, object any, attributes []string) error
}

// UserDetailsService 用户详情服务接口
// 根据用户名加载用户信息
type UserDetailsService interface {
	LoadUserByUsername(ctx context.Context, username string) (UserDetails, error)
}

// UserDetails 用户详情接口
// 包含用户的完整认证和授权信息
type UserDetails interface {
	Username() string
	Password() string
	Authorities() []string
	Enabled() bool
	AccountNonExpired() bool
	CredentialsNonExpired() bool
	AccountNonLocked() bool
}

// PasswordEncoder 密码编码器接口
// 用于密码的编码和验证
type PasswordEncoder interface {
	Encode(rawPassword string) string
	Matches(rawPassword, encodedPassword string) bool
}

// SecurityFilter 安全过滤器接口
// 处理安全相关的请求过滤
type SecurityFilter interface {
	DoFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain SecurityFilterChain) error
}

// SecurityFilterChain 安全过滤器链接口
// 按顺序执行多个安全过滤器
type SecurityFilterChain interface {
	DoFilter(ctx context.Context, request SecurityRequest, response SecurityResponse) error
}

// AccessDeniedHandler 访问拒绝处理器接口
// 处理访问被拒绝的情况
type AccessDeniedHandler interface {
	Handle(ctx context.Context, request SecurityRequest, response SecurityResponse, err error) error
}

// AuthenticationEntryPoint 认证入口点接口
// 处理未认证用户访问受保护资源的情况
type AuthenticationEntryPoint interface {
	Commence(ctx context.Context, request SecurityRequest, response SecurityResponse, err error) error
}

// SecurityMetadataSource 安全元数据源接口
// 提供访问资源所需的安全属性
type SecurityMetadataSource interface {
	GetAttributes(ctx context.Context, request SecurityRequest) ([]string, error)
}

// SecurityRequest 安全请求接口
// 抽象HTTP请求，提供安全相关的访问方法
type SecurityRequest interface {
	GetMethod() string
	GetURI() string
	GetHeader(key string) string
	SetAttribute(key string, value any)
	GetAttribute(key string) (any, bool)
}

// SecurityResponse 安全响应接口
// 抽象HTTP响应，提供安全相关的修改方法
type SecurityResponse interface {
	SetStatusCode(code int)
	SetHeader(key, value string)
	Write(data []byte) error
}

// GrantedAuthority 授予权限接口
// 代表一个具体的权限
type GrantedAuthority interface {
	Authority() string
}

// HttpSecurity HTTP安全配置接口
// 提供链式API配置HTTP安全规则
type HttpSecurity interface {
	AuthenticationManager(authManager AuthenticationManager) HttpSecurity
	UserDetailsService(userDetailsService UserDetailsService) HttpSecurity
	PasswordEncoder(encoder PasswordEncoder) HttpSecurity
	AccessDecisionManager(manager AccessDecisionManager) HttpSecurity
	SecurityMetadataSource(source SecurityMetadataSource) HttpSecurity
	AuthorizeRequests(authorizer AuthorizeRequests) HttpSecurity
	AddFilter(filter SecurityFilter) HttpSecurity
	AddFilterBefore(filter SecurityFilter, beforeFilter SecurityFilter) HttpSecurity
	AddFilterAfter(filter SecurityFilter, afterFilter SecurityFilter) HttpSecurity
	Anonymous() HttpSecurity
	ExceptionHandling(handler AccessDeniedHandler, entryPoint AuthenticationEntryPoint) HttpSecurity
	Csrf() HttpSecurity
	Logout(logoutUrl string, successHandler ...LogoutSuccessHandler) HttpSecurity
	FormLogin(loginProcessingUrl string, defaultSuccessUrl ...string) HttpSecurity
	HttpBasic() HttpSecurity
	Build() (SecurityFilterChain, error)
}

// AuthorizeRequests 授权请求配置接口
// 配置URL路径的访问规则
type AuthorizeRequests interface {
	AntMatchers(patterns ...string) ExpressionInterceptUrlRegistry
	AnyRequest() ExpressionInterceptUrlRegistry
}

// ExpressionInterceptUrlRegistry URL拦截注册接口
// 配置特定URL的访问表达式
type ExpressionInterceptUrlRegistry interface {
	PermitAll() HttpSecurity
	Authenticated() HttpSecurity
	HasRole(role string) HttpSecurity
	HasAnyRole(roles ...string) HttpSecurity
	HasAuthority(authority string) HttpSecurity
	HasAnyAuthority(authorities ...string) HttpSecurity
	DenyAll() HttpSecurity
}

// SecurityConfig 安全配置接口
// 提供自定义安全配置的入口
type SecurityConfig interface {
	Configure(http HttpSecurity) error
}

// securityContext 安全上下文实现
// 线程安全的认证信息存储
type securityContext struct {
	authentication Authentication
	mu             sync.RWMutex
}

func (s *securityContext) Authentication() Authentication {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.authentication
}

func (s *securityContext) SetAuthentication(auth Authentication) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.authentication = auth
}

func (s *securityContext) ClearAuthentication() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.authentication = nil
}

// contextKey 定义上下文键类型
type requestContextKey string

const authContextKey requestContextKey = "security.authentication"

// GetAuthenticationFromContext 从 context.Context 获取认证信息
func GetAuthenticationFromContext(ctx context.Context) Authentication {
	if val, ok := ctx.Value(authContextKey).(Authentication); ok {
		return val
	}
	return nil
}

// ContextWithAuthentication 将认证信息存入 context.Context
func ContextWithAuthentication(ctx context.Context, auth Authentication) context.Context {
	return context.WithValue(ctx, authContextKey, auth)
}

// 全局安全上下文实例
var globalSecurityContext = &securityContext{}

// GetSecurityContext 获取全局安全上下文
func GetSecurityContext() SecurityContext {
	return globalSecurityContext
}

// SetSecurityContext 设置全局安全上下文
func SetSecurityContext(ctx SecurityContext) {
	if sc, ok := ctx.(*securityContext); ok {
		auth := sc.Authentication()
		globalSecurityContext.SetAuthentication(auth)
	}
}

// GetAuthentication 获取当前认证信息
// 优先从请求上下文获取，如果不存在则从全局安全上下文获取
func GetAuthentication() Authentication {
	return globalSecurityContext.Authentication()
}

// SetAuthentication 设置当前认证信息到全局安全上下文
func SetAuthentication(auth Authentication) {
	globalSecurityContext.SetAuthentication(auth)
}

// ClearAuthentication 清除当前全局安全上下文中的认证信息
func ClearAuthentication() {
	globalSecurityContext.ClearAuthentication()
}

// InitSecurityContext 初始化安全上下文
// 用于测试环境重置安全状态
func InitSecurityContext() {
	globalSecurityContext = &securityContext{}
}
