package security

import (
	"context"
	"strings"
)

// 访问决策投票结果常量
const (
	ACCESS_GRANTED = 1
	ACCESS_DENIED  = -1
	ACCESS_ABSTAIN = 0
)

// AccessDecisionVoter 访问决策投票者接口
// 投票决定是否授予访问权限
type AccessDecisionVoter interface {
	Vote(ctx context.Context, authentication Authentication, object any, attributes []string) int
}

// WebExpressionVoter Web表达式投票者
// 支持Spring Security Web表达式语言
type WebExpressionVoter struct{}

func NewWebExpressionVoter() *WebExpressionVoter {
	return &WebExpressionVoter{}
}

// Vote 投票决定访问权限
// 支持的表达式:
//   - permitAll: 允许所有人访问
//   - denyAll: 拒绝所有人访问
//   - authenticated: 仅允许已认证用户
//   - hasRole('ROLE'): 检查是否具有指定角色
//   - hasAnyRole('ROLE1','ROLE2'): 检查是否具有任一指定角色
//   - hasAuthority('AUTHORITY'): 检查是否具有指定权限
//   - hasAnyAuthority('AUTH1','AUTH2'): 检查是否具有任一指定权限
func (v *WebExpressionVoter) Vote(ctx context.Context, authentication Authentication, object any, attributes []string) int {
	if len(attributes) == 0 {
		return ACCESS_ABSTAIN
	}

	for _, attribute := range attributes {
		if attribute == "permitAll" {
			return ACCESS_GRANTED
		}

		if attribute == "denyAll" {
			return ACCESS_DENIED
		}

		if attribute == "authenticated" {
			if authentication != nil && authentication.Authenticated() {
				return ACCESS_GRANTED
			}
			return ACCESS_DENIED
		}

		if strings.HasPrefix(attribute, "hasRole('") && strings.HasSuffix(attribute, "')") {
			role := strings.TrimPrefix(attribute, "hasRole('")
			role = strings.TrimSuffix(role, "')")
			if v.hasRole(authentication, role) {
				return ACCESS_GRANTED
			}
			return ACCESS_DENIED
		}

		if strings.HasPrefix(attribute, "hasAnyRole('") && strings.HasSuffix(attribute, "')") {
			rolesStr := strings.TrimPrefix(attribute, "hasAnyRole('")
			rolesStr = strings.TrimSuffix(rolesStr, "')")
			roles := strings.Split(rolesStr, "','")
			if v.hasAnyRole(authentication, roles) {
				return ACCESS_GRANTED
			}
			return ACCESS_DENIED
		}

		if strings.HasPrefix(attribute, "hasAuthority('") && strings.HasSuffix(attribute, "')") {
			authority := strings.TrimPrefix(attribute, "hasAuthority('")
			authority = strings.TrimSuffix(authority, "')")
			if v.hasAuthority(authentication, authority) {
				return ACCESS_GRANTED
			}
			return ACCESS_DENIED
		}

		if strings.HasPrefix(attribute, "hasAnyAuthority('") && strings.HasSuffix(attribute, "')") {
			authoritiesStr := strings.TrimPrefix(attribute, "hasAnyAuthority('")
			authoritiesStr = strings.TrimSuffix(authoritiesStr, "')")
			authorities := strings.Split(authoritiesStr, "','")
			if v.hasAnyAuthority(authentication, authorities) {
				return ACCESS_GRANTED
			}
			return ACCESS_DENIED
		}
	}

	return ACCESS_ABSTAIN
}

// hasRole 检查是否具有指定角色
// 支持带ROLE_前缀和不带前缀两种写法
func (v *WebExpressionVoter) hasRole(authentication Authentication, role string) bool {
	if authentication == nil {
		return false
	}

	authorities := authentication.Authorities()
	for _, auth := range authorities {
		if auth == "ROLE_"+role || auth == role {
			return true
		}
	}
	return false
}

// hasAnyRole 检查是否具有任一指定角色
func (v *WebExpressionVoter) hasAnyRole(authentication Authentication, roles []string) bool {
	if authentication == nil {
		return false
	}

	for _, role := range roles {
		if v.hasRole(authentication, role) {
			return true
		}
	}
	return false
}

// hasAuthority 检查是否具有指定权限
func (v *WebExpressionVoter) hasAuthority(authentication Authentication, authority string) bool {
	if authentication == nil {
		return false
	}

	authorities := authentication.Authorities()
	for _, auth := range authorities {
		if auth == authority {
			return true
		}
	}
	return false
}

// hasAnyAuthority 检查是否具有任一指定权限
func (v *WebExpressionVoter) hasAnyAuthority(authentication Authentication, authorities []string) bool {
	if authentication == nil {
		return false
	}

	for _, authority := range authorities {
		if v.hasAuthority(authentication, authority) {
			return true
		}
	}
	return false
}

// RoleVoter 角色投票者
// 只处理具有指定前缀(默认ROLE_)的角色
type RoleVoter struct {
	rolePrefix string
}

// NewRoleVoter 创建角色投票者，默认前缀为ROLE_
func NewRoleVoter() *RoleVoter {
	return &RoleVoter{
		rolePrefix: "ROLE_",
	}
}

// Vote 投票决定访问权限
// 只处理rolePrefix前缀的属性，其他属性返回ABSTAIN
func (v *RoleVoter) Vote(ctx context.Context, authentication Authentication, object any, attributes []string) int {
	if len(attributes) == 0 {
		return ACCESS_ABSTAIN
	}

	for _, attribute := range attributes {
		if v.supports(attribute) {
			role := strings.TrimPrefix(attribute, v.rolePrefix)
			if v.hasRole(authentication, role) {
				return ACCESS_GRANTED
			}
			return ACCESS_DENIED
		}
	}

	return ACCESS_ABSTAIN
}

// supports 检查属性是否以此投票者支持的前缀开头
func (v *RoleVoter) supports(attribute string) bool {
	return strings.HasPrefix(attribute, v.rolePrefix)
}

// hasRole 检查是否具有指定角色
func (v *RoleVoter) hasRole(authentication Authentication, role string) bool {
	if authentication == nil {
		return false
	}

	authorities := authentication.Authorities()
	for _, auth := range authorities {
		if auth == v.rolePrefix+role || auth == role {
			return true
		}
	}
	return false
}

// SetRolePrefix 设置角色前缀
func (v *RoleVoter) SetRolePrefix(prefix string) {
	v.rolePrefix = prefix
}

// AuthenticatedVoter 认证投票者
// 检查用户是否满足特定的认证级别
type AuthenticatedVoter struct{}

func NewAuthenticatedVoter() *AuthenticatedVoter {
	return &AuthenticatedVoter{}
}

// Vote 投票决定访问权限
// 支持的表达式:
//   - IS_AUTHENTICATED_FULLY: 要求完全认证（非匿名）
//   - IS_AUTHENTICATED_REMEMBERED: 允许记住我或完全认证
//   - IS_AUTHENTICATED_ANONYMOUSLY: 允许匿名用户访问（即所有人都可以访问）
func (v *AuthenticatedVoter) Vote(ctx context.Context, authentication Authentication, object any, attributes []string) int {
	if len(attributes) == 0 {
		return ACCESS_ABSTAIN
	}

	for _, attribute := range attributes {
		if attribute == "IS_AUTHENTICATED_FULLY" {
			if authentication != nil && authentication.Authenticated() {
				return ACCESS_GRANTED
			}
			return ACCESS_DENIED
		}

		if attribute == "IS_AUTHENTICATED_REMEMBERED" {
			if authentication != nil && authentication.Authenticated() {
				return ACCESS_GRANTED
			}
			return ACCESS_DENIED
		}

		if attribute == "IS_AUTHENTICATED_ANONYMOUSLY" {
			return ACCESS_GRANTED
		}
	}

	return ACCESS_ABSTAIN
}

// AffirmativeBased 肯定优先访问决策管理器
// 只要有一个投票者授予访问权限就允许访问
type AffirmativeBased struct {
	decisionVoters             []AccessDecisionVoter
	allowIfAllAbstainDecisions bool
}

// NewAffirmativeBased 创建肯定优先决策管理器
func NewAffirmativeBased(voters ...AccessDecisionVoter) *AffirmativeBased {
	return &AffirmativeBased{
		decisionVoters:             voters,
		allowIfAllAbstainDecisions: false,
	}
}

// Decide 决定是否授予访问权限
// 策略: 只要有一个投票者授予权限就通过，有一个拒绝就拒绝
func (m *AffirmativeBased) Decide(ctx context.Context, authentication Authentication, object any, attributes []string) error {
	grant := 0
	deny := 0
	abstain := 0

	for _, voter := range m.decisionVoters {
		result := voter.Vote(ctx, authentication, object, attributes)
		switch result {
		case ACCESS_GRANTED:
			grant++
		case ACCESS_DENIED:
			deny++
		case ACCESS_ABSTAIN:
			abstain++
		}
	}

	if grant > 0 {
		return nil
	}

	if deny > 0 {
		return ErrAccessDenied
	}

	if m.allowIfAllAbstainDecisions {
		return nil
	}

	return ErrAccessDenied
}

// AddVoter 添加投票者
func (m *AffirmativeBased) AddVoter(voter AccessDecisionVoter) {
	m.decisionVoters = append(m.decisionVoters, voter)
}

// SetAllowIfAllAbstainDecisions 设置是否在所有投票者都弃权时允许访问
func (m *AffirmativeBased) SetAllowIfAllAbstainDecisions(allow bool) {
	m.allowIfAllAbstainDecisions = allow
}

// UnanimousBased 一致通过访问决策管理器
// 只有所有投票者都不拒绝才允许访问
type UnanimousBased struct {
	decisionVoters             []AccessDecisionVoter
	allowIfAllAbstainDecisions bool
}

// NewUnanimousBased 创建一致通过决策管理器
func NewUnanimousBased(voters ...AccessDecisionVoter) *UnanimousBased {
	return &UnanimousBased{
		decisionVoters:             voters,
		allowIfAllAbstainDecisions: false,
	}
}

// Decide 决定是否授予访问权限
// 策略: 所有投票者都不拒绝，且至少有一个授予权限时通过
func (m *UnanimousBased) Decide(ctx context.Context, authentication Authentication, object any, attributes []string) error {
	deny := 0
	grant := 0
	abstain := 0

	for _, voter := range m.decisionVoters {
		result := voter.Vote(ctx, authentication, object, attributes)
		switch result {
		case ACCESS_GRANTED:
			grant++
		case ACCESS_DENIED:
			deny++
		case ACCESS_ABSTAIN:
			abstain++
		}
	}

	if deny > 0 {
		return ErrAccessDenied
	}

	if grant > 0 {
		return nil
	}

	if m.allowIfAllAbstainDecisions {
		return nil
	}

	return ErrAccessDenied
}

// AddVoter 添加投票者
func (m *UnanimousBased) AddVoter(voter AccessDecisionVoter) {
	m.decisionVoters = append(m.decisionVoters, voter)
}

// SetAllowIfAllAbstainDecisions 设置是否在所有投票者都弃权时允许访问
func (m *UnanimousBased) SetAllowIfAllAbstainDecisions(allow bool) {
	m.allowIfAllAbstainDecisions = allow
}

// ConsensusBased 共识优先访问决策管理器
// 根据多数投票结果决定访问权限
type ConsensusBased struct {
	decisionVoters             []AccessDecisionVoter
	allowIfEqualGrantedDenied  bool
	allowIfAllAbstainDecisions bool
}

// NewConsensusBased 创建共识优先决策管理器
func NewConsensusBased(voters ...AccessDecisionVoter) *ConsensusBased {
	return &ConsensusBased{
		decisionVoters:             voters,
		allowIfEqualGrantedDenied:  false,
		allowIfAllAbstainDecisions: false,
	}
}

// Decide 决定是否授予访问权限
// 策略: 根据多数票决定，相同数量时按配置处理
func (m *ConsensusBased) Decide(ctx context.Context, authentication Authentication, object any, attributes []string) error {
	grant := 0
	deny := 0
	abstain := 0

	for _, voter := range m.decisionVoters {
		result := voter.Vote(ctx, authentication, object, attributes)
		switch result {
		case ACCESS_GRANTED:
			grant++
		case ACCESS_DENIED:
			deny++
		case ACCESS_ABSTAIN:
			abstain++
		}
	}

	if grant > deny {
		return nil
	}

	if deny > grant {
		return ErrAccessDenied
	}

	if grant == deny {
		if m.allowIfEqualGrantedDenied {
			return nil
		}
		return ErrAccessDenied
	}

	if m.allowIfAllAbstainDecisions {
		return nil
	}

	return ErrAccessDenied
}

// AddVoter 添加投票者
func (m *ConsensusBased) AddVoter(voter AccessDecisionVoter) {
	m.decisionVoters = append(m.decisionVoters, voter)
}

// SetAllowIfEqualGrantedDenied 设置当授权和拒绝票数相同时的处理方式
func (m *ConsensusBased) SetAllowIfEqualGrantedDenied(allow bool) {
	m.allowIfEqualGrantedDenied = allow
}

// SetAllowIfAllAbstainDecisions 设置是否在所有投票者都弃权时允许访问
func (m *ConsensusBased) SetAllowIfAllAbstainDecisions(allow bool) {
	m.allowIfAllAbstainDecisions = allow
}
