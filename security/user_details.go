package security

import (
	"context"
	"sync"
)

// InMemoryUserDetails 内存用户详情实现
// 存储用户的完整认证和授权信息
type InMemoryUserDetails struct {
	username              string
	password              string
	authorities           []string
	enabled               bool
	accountNonExpired     bool
	credentialsNonExpired bool
	accountNonLocked      bool
}

// NewInMemoryUserDetails 创建内存用户详情
// 默认值: enabled=true, accountNonExpired=true, credentialsNonExpired=true, accountNonLocked=true
func NewInMemoryUserDetails(username, password string, authorities []string) *InMemoryUserDetails {
	return &InMemoryUserDetails{
		username:              username,
		password:              password,
		authorities:           authorities,
		enabled:               true,
		accountNonExpired:     true,
		credentialsNonExpired: true,
		accountNonLocked:      true,
	}
}

// Username 返回用户名
func (u *InMemoryUserDetails) Username() string {
	return u.username
}

// Password 返回密码
func (u *InMemoryUserDetails) Password() string {
	return u.password
}

// Authorities 返回授权列表
func (u *InMemoryUserDetails) Authorities() []string {
	return u.authorities
}

// Enabled 返回是否启用
func (u *InMemoryUserDetails) Enabled() bool {
	return u.enabled
}

// AccountNonExpired 返回账户是否未过期
func (u *InMemoryUserDetails) AccountNonExpired() bool {
	return u.accountNonExpired
}

// CredentialsNonExpired 返回凭证是否未过期
func (u *InMemoryUserDetails) CredentialsNonExpired() bool {
	return u.credentialsNonExpired
}

// AccountNonLocked 返回账户是否未锁定
func (u *InMemoryUserDetails) AccountNonLocked() bool {
	return u.accountNonLocked
}

// InMemoryUserDetailsService 内存用户详情服务
// 使用内存map存储用户信息，适用于开发和测试
type InMemoryUserDetailsService struct {
	users map[string]UserDetails
	mu    sync.RWMutex
}

// NewInMemoryUserDetailsService 创建内存用户详情服务
func NewInMemoryUserDetailsService() *InMemoryUserDetailsService {
	return &InMemoryUserDetailsService{
		users: make(map[string]UserDetails),
	}
}

// CreateUser 创建新用户
func (s *InMemoryUserDetailsService) CreateUser(username, password string, authorities []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[username] = NewInMemoryUserDetails(username, password, authorities)
}

// LoadUserByUsername 根据用户名加载用户
// 如果用户不存在返回ErrUserNotFound
func (s *InMemoryUserDetailsService) LoadUserByUsername(ctx context.Context, username string) (UserDetails, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[username]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// DeleteUser 删除用户
func (s *InMemoryUserDetailsService) DeleteUser(username string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.users, username)
}

// UserCount 返回用户总数
func (s *InMemoryUserDetailsService) UserCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.users)
}

// Role 角色实现
// 实现GrantedAuthority接口
type Role struct {
	authority string
}

// NewRole 创建角色
func NewRole(authority string) *Role {
	return &Role{authority: authority}
}

// Authority 返回权限字符串
func (r *Role) Authority() string {
	return r.authority
}

// NewAuthority 创建权限（GrantedAuthority接口）
func NewAuthority(authority string) GrantedAuthority {
	return NewRole(authority)
}
