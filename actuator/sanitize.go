// Package actuator 提供应用监控端点，支持健康检查、指标采集和环境变量查看。
package actuator

import (
	"encoding/base64"
	"strings"
	"sync"
)

const redactedValue = "***REDACTED***"

// Sanitizer 敏感信息检测器
//
// 使用策略模式管理多种敏感信息检测规则，
// 支持自定义检测策略。
type Sanitizer struct {
	mu         sync.RWMutex
	keywords   []string
	strategies []SanitizeStrategy
}

// SanitizeStrategy 敏感信息检测策略接口
type SanitizeStrategy interface {
	// IsSensitive 检查键值对是否包含敏感信息
	IsSensitive(key string, value any) bool
}

// NewSanitizer 创建敏感信息检测器
func NewSanitizer() *Sanitizer {
	return &Sanitizer{
		keywords: defaultKeywords(),
	}
}

// AddStrategy 添加自定义检测策略
func (s *Sanitizer) AddStrategy(strategy SanitizeStrategy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.strategies = append(s.strategies, strategy)
}

// Sanitize 掩盖敏感值
func (s *Sanitizer) Sanitize(key string, value any) any {
	if value == nil {
		return value
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// 检查自定义策略
	for _, strategy := range s.strategies {
		if strategy.IsSensitive(key, value) {
			return redactedValue
		}
	}

	// 检查关键词
	keyLower := strings.ToLower(key)
	for _, keyword := range s.keywords {
		if strings.Contains(keyLower, keyword) {
			return redactedValue
		}
	}

	// 检查值本身
	if str, ok := value.(string); ok {
		if looksLikeSensitiveData(str) {
			return redactedValue
		}
	}

	return value
}

// defaultKeywords 返回默认敏感关键词列表
func defaultKeywords() []string {
	return []string{
		"password", "secret", "token", "key", "auth",
		"credential", "private", "api_key", "access_token",
		"client_secret", "oauth", "bearer", "jwt",
	}
}

// looksLikeSensitiveData 检查字符串是否看起来像敏感数据
func looksLikeSensitiveData(value string) bool {
	if len(value) > 32 {
		if strings.HasPrefix(value, "-----BEGIN") && strings.Contains(value, "PRIVATE KEY") {
			return true
		}
		if isRandomLookingString(value) {
			return true
		}
		if isTokenFormat(value) {
			return true
		}
	}
	return false
}

// isRandomLookingString 检查字符串是否看起来像随机字符串
func isRandomLookingString(s string) bool {
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= '0' && r <= '9':
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial && len(s) > 16
}

// isTokenFormat 检查字符串是否为令牌格式
func isTokenFormat(s string) bool {
	// JWT 令牌格式检测
	if strings.Count(s, ".") >= 2 {
		parts := strings.Split(s, ".")
		if len(parts) == 3 {
			for _, part := range parts {
				if !isValidBase64(part) {
					return false
				}
			}
			return true
		}
	}
	return false
}

// isValidBase64 检查字符串是否为有效的 Base64 编码
func isValidBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}
