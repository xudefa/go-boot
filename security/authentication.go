package security

import (
	"context"
	"errors"
)

// UsernamePasswordAuthenticationToken 用户名密码认证令牌
// 用于在认证过程中传递和存储用户凭证
type UsernamePasswordAuthenticationToken struct {
	principal     any
	credentials   any
	authorities   []string
	authenticated bool
}

// NewUsernamePasswordAuthenticationToken 创建未认证的用户名密码令牌
// 用于认证请求的初始阶段
func NewUsernamePasswordAuthenticationToken(principal, credentials any) *UsernamePasswordAuthenticationToken {
	return &UsernamePasswordAuthenticationToken{
		principal:     principal,
		credentials:   credentials,
		authorities:   []string{},
		authenticated: false,
	}
}

// NewAuthenticatedUsernamePasswordAuthenticationToken 创建已认证的用户名密码令牌
// 用于认证成功后的令牌
func NewAuthenticatedUsernamePasswordAuthenticationToken(principal any, authorities []string) *UsernamePasswordAuthenticationToken {
	return &UsernamePasswordAuthenticationToken{
		principal:     principal,
		credentials:   nil,
		authorities:   authorities,
		authenticated: true,
	}
}

func (t *UsernamePasswordAuthenticationToken) Principal() any {
	return t.principal
}

func (t *UsernamePasswordAuthenticationToken) Credentials() any {
	return t.credentials
}

func (t *UsernamePasswordAuthenticationToken) Authorities() []string {
	return t.authorities
}

func (t *UsernamePasswordAuthenticationToken) Authenticated() bool {
	return t.authenticated
}

// Name 返回认证主体名称
// 支持字符串和UserDetails两种principal类型
func (t *UsernamePasswordAuthenticationToken) Name() string {
	if name, ok := t.principal.(string); ok {
		return name
	}
	if userDetails, ok := t.principal.(UserDetails); ok {
		return userDetails.Username()
	}
	return ""
}

// SetAuthenticated 设置认证状态
func (t *UsernamePasswordAuthenticationToken) SetAuthenticated(authenticated bool) {
	t.authenticated = authenticated
}

// SetAuthorities 设置授权列表
func (t *UsernamePasswordAuthenticationToken) SetAuthorities(authorities []string) {
	t.authorities = authorities
}

// ProviderManager 认证提供者管理器
// 管理多个认证提供者，按顺序尝试认证
type ProviderManager struct {
	providers []AuthenticationProvider
}

// NewProviderManager 创建认证提供者管理器
func NewProviderManager(providers ...AuthenticationProvider) *ProviderManager {
	return &ProviderManager{
		providers: providers,
	}
}

// Authenticate 尝试通过配置的提供者进行认证
// 遍历所有提供者，返回第一个成功认证的结果
func (m *ProviderManager) Authenticate(ctx context.Context, authentication Authentication) (Authentication, error) {
	var lastErr error

	for _, provider := range m.providers {
		if provider.Supports(authentication) {
			result, err := provider.Authenticate(ctx, authentication)
			if err != nil {
				lastErr = err
				continue
			}
			if result != nil {
				return result, nil
			}
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return nil, ErrAuthenticationFailed
}

// AddProvider 添加认证提供者
func (m *ProviderManager) AddProvider(provider AuthenticationProvider) {
	m.providers = append(m.providers, provider)
}

// AuthenticationProvider 认证提供者接口
// 定义认证逻辑
type AuthenticationProvider interface {
	Authenticate(ctx context.Context, authentication Authentication) (Authentication, error)
	Supports(authentication Authentication) bool
}

// DaoAuthenticationProvider 基于DAO的认证提供者
// 从UserDetailsService加载用户并验证密码
type DaoAuthenticationProvider struct {
	userDetailsService UserDetailsService
	passwordEncoder    PasswordEncoder
}

// NewDaoAuthenticationProvider 创建DAO认证提供者
func NewDaoAuthenticationProvider(userDetailsService UserDetailsService, passwordEncoder PasswordEncoder) *DaoAuthenticationProvider {
	return &DaoAuthenticationProvider{
		userDetailsService: userDetailsService,
		passwordEncoder:    passwordEncoder,
	}
}

// Authenticate 执行认证逻辑
// 1. 根据用户名加载用户信息
// 2. 验证密码
// 3. 检查用户状态
func (p *DaoAuthenticationProvider) Authenticate(ctx context.Context, authentication Authentication) (Authentication, error) {
	username := authentication.Name()
	if username == "" {
		return nil, ErrBadCredentials
	}

	user, err := p.userDetailsService.LoadUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrBadCredentials
		}
		return nil, err
	}

	presentedPassword := ""
	if creds, ok := authentication.Credentials().(string); ok {
		presentedPassword = creds
	}

	if !p.passwordEncoder.Matches(presentedPassword, user.Password()) {
		return nil, ErrBadCredentials
	}

	if !user.Enabled() {
		return nil, errors.New("user is disabled")
	}

	if !user.AccountNonLocked() {
		return nil, errors.New("user account is locked")
	}

	return NewAuthenticatedUsernamePasswordAuthenticationToken(user, user.Authorities()), nil
}

// Supports 判断是否支持该认证方式
func (p *DaoAuthenticationProvider) Supports(authentication Authentication) bool {
	_, ok := authentication.(*UsernamePasswordAuthenticationToken)
	return ok
}

// AnonymousAuthenticationProvider 匿名认证提供者
// 为匿名用户创建认证令牌，当其他认证提供者都无法处理时使用
type AnonymousAuthenticationProvider struct{}

// NewAnonymousAuthenticationProvider 创建匿名认证提供者
func NewAnonymousAuthenticationProvider() *AnonymousAuthenticationProvider {
	return &AnonymousAuthenticationProvider{}
}

// Authenticate 为匿名用户创建认证令牌
// 只有当传入的authentication为nil或未认证时，才创建匿名令牌
func (p *AnonymousAuthenticationProvider) Authenticate(ctx context.Context, authentication Authentication) (Authentication, error) {
	if authentication == nil || !authentication.Authenticated() {
		return NewAuthenticatedUsernamePasswordAuthenticationToken("anonymousUser", []string{"ROLE_ANONYMOUS"}), nil
	}
	return nil, nil
}

// Supports 只支持UsernamePasswordAuthenticationToken类型
func (p *AnonymousAuthenticationProvider) Supports(authentication Authentication) bool {
	_, ok := authentication.(*UsernamePasswordAuthenticationToken)
	return ok
}
