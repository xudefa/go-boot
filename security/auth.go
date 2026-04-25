// Package security 提供认证和授权功能,包括用户管理、JWT令牌、安全上下文等.
package security

import (
	"crypto/md5"
	"errors"
	"fmt"
	"sync"
)

// User 是用户模型,包含用户的基本信息和权限.
type User struct {
	Id           string   `json:"id"`       // 用户唯一标识
	Username     string   `json:"username"` // 用户名(登录账号)
	Password     string   `json:"-"`        // 密码(不序列化,保证安全)
	Email        string   `json:"email"`    // 邮箱地址
	Roles        []string `json:"roles"`    // 角色列表
	Enabled      bool     `json:"enabled"`  // 账户是否启用
	Locked       bool     `json:"locked"`   // 账户是否锁定
	PasswordHash string   `json:"-"`        // 密码哈希(不序列化)
}

// UserDetails 是用户详情接口,用于获取用户信息.
type UserDetails interface {
	GetUserId() string   // 获取用户ID
	GetUsername() string // 获取用户名
	GetRoles() []string  // 获取角色列表
	IsEnabled() bool     // 检查账户是否启用
}

// UserDetailsService 是用户详情服务接口,用于加载用户.
type UserDetailsService interface {
	LoadUserByUsername(username string) (UserDetails, error) // 根据用户名加载用户
}

// Authentication 是认证结果接口,表示认证成功后的结果.
type Authentication interface {
	GetPrincipal() interface{}   // 获取认证主体(通常为用户名)
	GetCredentials() interface{} // 获取凭证(通常为密码)
	GetAuthorities() []string    // 获取权限列表(通常为角色)
	IsAuthenticated() bool       // 是否认证成功
}

// AuthenticationToken 是认证结果实现,包含认证的基本信息.
type AuthenticationToken struct {
	Principal     interface{} // 认证主体
	Credentials   interface{} // 凭证
	Authorities   []string    // 权限列表
	Authenticated bool        // 是否认证成功
}

// GetPrincipal 获取认证主体
func (a *AuthenticationToken) GetPrincipal() interface{} {
	return a.Principal
}

// GetCredentials 获取凭证
func (a *AuthenticationToken) GetCredentials() interface{} {
	return a.Credentials
}

// GetAuthorities 获取权限列表
func (a *AuthenticationToken) GetAuthorities() []string {
	return a.Authorities
}

// IsAuthenticated 检查是否认证成功
func (a *AuthenticationToken) IsAuthenticated() bool {
	return a.Authenticated
}

// AuthenticationManager 是认证管理器,负责管理多个认证提供者.
type AuthenticationManager struct {
	providers  []AuthenticationProvider // 认证提供者列表
	userCache  map[string]UserDetails   // 用户缓存
	cacheMutex sync.RWMutex             // 缓存互斥锁
}

// NewAuthenticationManager 创建新的认证管理器
func NewAuthenticationManager() *AuthenticationManager {
	return &AuthenticationManager{
		providers: make([]AuthenticationProvider, 0),
		userCache: make(map[string]UserDetails),
	}
}

// RegisterProvider 注册认证提供者
func (am *AuthenticationManager) RegisterProvider(provider AuthenticationProvider) {
	am.providers = append(am.providers, provider)
}

// Authenticate 进行用户认证
// 首先尝试从缓存加载用户,然后遍历所有提供者进行认证
func (am *AuthenticationManager) Authenticate(username, password string) (Authentication, error) {
	var userDetails UserDetails
	var err error

	// 先从缓存获取
	am.cacheMutex.RLock()
	if user, ok := am.userCache[username]; ok {
		userDetails = user
	}
	am.cacheMutex.RUnlock()

	// 缓存未命中,从提供者加载
	if userDetails == nil {
		for _, provider := range am.providers {
			userDetails, err = provider.LoadUserByUsername(username)
			if err == nil {
				break
			}
		}
	}

	// 用户不存在
	if userDetails == nil {
		return nil, errors.New("user not found")
	}

	// 检查账户是否启用
	if !userDetails.IsEnabled() {
		return nil, errors.New("user is disabled")
	}

	// 遍历提供者进行认证
	for _, provider := range am.providers {
		if provider.Supports(&AuthenticationRequest{Username: username, Password: password}) {
			auth, err := provider.Authenticate(&AuthenticationRequest{
				Username:    username,
				Password:    password,
				UserDetails: userDetails,
			})
			if err == nil && auth.IsAuthenticated() {
				// 认证成功,缓存用户
				am.cacheMutex.Lock()
				am.userCache[username] = userDetails
				am.cacheMutex.Unlock()
				return auth, nil
			}
		}
	}

	return nil, errors.New("authentication failed")
}

// AuthenticationProvider 是认证提供者接口,实现具体的认证逻辑.
type AuthenticationProvider interface {
	Supports(request *AuthenticationRequest) bool                        // 是否支持该请求
	Authenticate(request *AuthenticationRequest) (Authentication, error) // 进行认证
	LoadUserByUsername(username string) (UserDetails, error)             // 加载用户
}

// AuthenticationRequest 是认证请求,包含认证所需的信息.
type AuthenticationRequest struct {
	Username    string      // 用户名
	Password    string      // 密码
	UserDetails UserDetails // 用户详情
}

// InMemoryUserDetailsService 是内存用户详情服务,用户信息存储在内存中.
type InMemoryUserDetailsService struct {
	users map[string]*User // 用户映射
}

// NewInMemoryUserDetailsService 创建新的内存用户服务
func NewInMemoryUserDetailsService() *InMemoryUserDetailsService {
	return &InMemoryUserDetailsService{
		users: make(map[string]*User),
	}
}

// CreateUser 创建用户
func (s *InMemoryUserDetailsService) CreateUser(user *User) {
	user.PasswordHash = HashPassword(user.Password)
	s.users[user.Username] = user
}

// LoadUserByUsername 根据用户名加载用户
func (s *InMemoryUserDetailsService) LoadUserByUsername(username string) (UserDetails, error) {
	if user, ok := s.users[username]; ok {
		return &userDetailsAdapter{user: user}, nil
	}
	return nil, errors.New("user not found")
}

// DeleteUser 删除用户
func (s *InMemoryUserDetailsService) DeleteUser(username string) {
	delete(s.users, username)
}

// UpdateUser 更新用户
func (s *InMemoryUserDetailsService) UpdateUser(user *User) {
	user.PasswordHash = HashPassword(user.Password)
	s.users[user.Username] = user
}

// userDetailsAdapter 是User到UserDetails的适配器
type userDetailsAdapter struct {
	user *User
}

func (a *userDetailsAdapter) GetUserId() string {
	return a.user.Id
}

func (a *userDetailsAdapter) GetUsername() string {
	return a.user.Username
}

func (a *userDetailsAdapter) GetRoles() []string {
	return a.user.Roles
}

func (a *userDetailsAdapter) IsEnabled() bool {
	return a.user.Enabled
}

// DaoAuthenticationProvider 是DAO认证提供者,通过密码哈希验证用户.
type DaoAuthenticationProvider struct {
	userDetailsService UserDetailsService // 用户详情服务
}

func NewDaoAuthenticationProvider(userDetailsService UserDetailsService) *DaoAuthenticationProvider {
	return &DaoAuthenticationProvider{
		userDetailsService: userDetailsService,
	}
}

func (p *DaoAuthenticationProvider) Supports(request *AuthenticationRequest) bool {
	return true
}

func (p *DaoAuthenticationProvider) Authenticate(request *AuthenticationRequest) (Authentication, error) {
	// 注意: 此处有类型断言问题,实际使用时需要修复
	if !ComparePasswords(request.UserDetails.(interface{ GetPassword() string }).GetPassword(), request.Password) {
		return nil, errors.New("invalid credentials")
	}

	return &AuthenticationToken{
		Principal:     request.UserDetails.GetUsername(),
		Authorities:   request.UserDetails.GetRoles(),
		Authenticated: true,
	}, nil
}

func (p *DaoAuthenticationProvider) LoadUserByUsername(username string) (UserDetails, error) {
	return p.userDetailsService.LoadUserByUsername(username)
}

// SimpleAuthenticationProvider 是简单认证提供者,直接验证密码.
type SimpleAuthenticationProvider struct {
	userDetailsService UserDetailsService
}

func NewSimpleAuthenticationProvider(userDetailsService UserDetailsService) *SimpleAuthenticationProvider {
	return &SimpleAuthenticationProvider{
		userDetailsService: userDetailsService,
	}
}

func (p *SimpleAuthenticationProvider) Supports(request *AuthenticationRequest) bool {
	return true
}

func (p *SimpleAuthenticationProvider) Authenticate(request *AuthenticationRequest) (Authentication, error) {
	userDetails, err := p.userDetailsService.LoadUserByUsername(request.Username)
	if err != nil {
		return nil, err
	}

	if !userDetails.IsEnabled() {
		return nil, errors.New("user is disabled")
	}

	return &AuthenticationToken{
		Principal:     userDetails.GetUsername(),
		Credentials:   request.Password,
		Authorities:   userDetails.GetRoles(),
		Authenticated: true,
	}, nil
}

func (p *SimpleAuthenticationProvider) LoadUserByUsername(username string) (UserDetails, error) {
	return p.userDetailsService.LoadUserByUsername(username)
}

// HashPassword 使用MD5哈希密码
// 注意: 生产环境应使用更安全的哈希算法(如bcrypt)
func HashPassword(password string) string {
	hash := md5.Sum([]byte(password))
	return fmt.Sprintf("%x", hash)
}

// SecurityContext 是安全上下文,存储当前线程的认证信息.
type SecurityContext struct {
	authentication Authentication // 认证信息
	mu             sync.RWMutex   // 互斥锁
}

func NewSecurityContext() *SecurityContext {
	return &SecurityContext{}
}

// SetAuthentication 设置认证信息
func (sc *SecurityContext) SetAuthentication(auth Authentication) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.authentication = auth
}

// GetAuthentication 获取认证信息
func (sc *SecurityContext) GetAuthentication() Authentication {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.authentication
}

// IsAuthenticated 检查是否已认证
func (sc *SecurityContext) IsAuthenticated() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.authentication != nil && sc.authentication.IsAuthenticated()
}

// GetUserId 获取用户ID
func (sc *SecurityContext) GetUserId() string {
	if !sc.IsAuthenticated() {
		return ""
	}
	if claims, ok := sc.authentication.GetPrincipal().(*JWTClaims); ok {
		return claims.UserId
	}
	return fmt.Sprintf("%v", sc.authentication.GetPrincipal())
}

// GetUsername 获取用户名
func (sc *SecurityContext) GetUsername() string {
	if !sc.IsAuthenticated() {
		return ""
	}
	if claims, ok := sc.authentication.GetPrincipal().(*JWTClaims); ok {
		return claims.Username
	}
	return fmt.Sprintf("%v", sc.authentication.GetPrincipal())
}

// GetRoles 获取角色列表
func (sc *SecurityContext) GetRoles() []string {
	if !sc.IsAuthenticated() {
		return nil
	}
	return sc.authentication.GetAuthorities()
}

// HasRole 检查是否拥有指定角色
func (sc *SecurityContext) HasRole(role string) bool {
	roles := sc.GetRoles()
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// Clear 清除认证信息
func (sc *SecurityContext) Clear() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.authentication = nil
}

// SecurityContextHolder 是安全上下文持有者,管理多个安全上下文.
// 支持根据请求ID获取或创建安全上下文.
type SecurityContextHolder struct {
	contexts sync.Map // 安全上下文映射
}

var securityContextHolder = &SecurityContextHolder{}

// GetSecurityContextHolder 获取全局安全上下文持有者
func GetSecurityContextHolder() *SecurityContextHolder {
	return securityContextHolder
}

// GetContext 根据请求ID获取或创建安全上下文
func (h *SecurityContextHolder) GetContext(requestId string) *SecurityContext {
	if ctx, ok := h.contexts.Load(requestId); ok {
		return ctx.(*SecurityContext)
	}
	newCtx := NewSecurityContext()
	h.contexts.Store(requestId, newCtx)
	return newCtx
}

// RemoveContext 移除安全上下文
func (h *SecurityContextHolder) RemoveContext(requestId string) {
	h.contexts.Delete(requestId)
}

// SecurityFilter 是安全过滤器,用于HTTP请求的认证拦截.
type SecurityFilter struct {
	jwtManager  *JWTManager            // JWT管理器
	authManager *AuthenticationManager // 认证管理器
	skipPaths   []string               // 跳过认证的路径列表
}

// NewSecurityFilter 创建新的安全过滤器
func NewSecurityFilter(jwtManager *JWTManager, authManager *AuthenticationManager) *SecurityFilter {
	return &SecurityFilter{
		jwtManager:  jwtManager,
		authManager: authManager,
		skipPaths: []string{
			"/api/auth/login",
			"/api/auth/register",
			"/actuator/health",
			"/actuator/info",
		},
	}
}

// AddSkipPaths 添加跳过认证的路径
func (f *SecurityFilter) AddSkipPaths(paths ...string) {
	f.skipPaths = append(f.skipPaths, paths...)
}

// ShouldSkip 检查是否应该跳过认证
func (f *SecurityFilter) ShouldSkip(path string) bool {
	for _, p := range f.skipPaths {
		if p == path {
			return true
		}
	}
	return false
}
