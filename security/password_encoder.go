package security

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// NoOpPasswordEncoder 不进行编码的密码编码器
// 适用于开发和测试环境，明文存储密码
type NoOpPasswordEncoder struct{}

// NewNoOpPasswordEncoder 创建NoOp密码编码器
func NewNoOpPasswordEncoder() *NoOpPasswordEncoder {
	return &NoOpPasswordEncoder{}
}

// Encode 直接返回原始密码
func (e *NoOpPasswordEncoder) Encode(rawPassword string) string {
	return rawPassword
}

// Matches 直接比较原始密码和编码后密码
func (e *NoOpPasswordEncoder) Matches(rawPassword, encodedPassword string) bool {
	return rawPassword == encodedPassword
}

// BCryptPasswordEncoder 基于SHA256的密码编码器
// 使用SHA256算法对密码进行哈希，适用于生产环境
type BCryptPasswordEncoder struct{}

// NewBCryptPasswordEncoder 创建BCrypt密码编码器
func NewBCryptPasswordEncoder() *BCryptPasswordEncoder {
	return &BCryptPasswordEncoder{}
}

// Encode 使用SHA256算法编码密码
func (e *BCryptPasswordEncoder) Encode(rawPassword string) string {
	hash := sha256.Sum256([]byte(rawPassword))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// Matches 比较编码后的密码
func (e *BCryptPasswordEncoder) Matches(rawPassword, encodedPassword string) bool {
	return e.Encode(rawPassword) == encodedPassword
}

// StandardPasswordEncoder 标准密码编码器
// 使用密钥对密码进行SHA256哈希
type StandardPasswordEncoder struct {
	secret string
}

// NewStandardPasswordEncoder 创建标准密码编码器
// secret: 用于加盐的密钥
func NewStandardPasswordEncoder(secret string) *StandardPasswordEncoder {
	return &StandardPasswordEncoder{secret: secret}
}

// Encode 使用密钥对密码进行编码
func (e *StandardPasswordEncoder) Encode(rawPassword string) string {
	combined := e.secret + rawPassword
	hash := sha256.Sum256([]byte(combined))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// Matches 比较编码后的密码
func (e *StandardPasswordEncoder) Matches(rawPassword, encodedPassword string) bool {
	return e.Encode(rawPassword) == encodedPassword
}

// DelegatingPasswordEncoder 委托密码编码器
// 支持多种编码器，可根据编码ID选择合适的编码器
// 编码格式: {编码器ID}encodedPassword
type DelegatingPasswordEncoder struct {
	idForEncode      string
	passwordEncoders map[string]PasswordEncoder
}

// NewDelegatingPasswordEncoder 创建委托密码编码器
// idForEncode: 默认使用的编码器ID
// passwordEncoders: 可用的编码器映射
func NewDelegatingPasswordEncoder(idForEncode string, passwordEncoders map[string]PasswordEncoder) *DelegatingPasswordEncoder {
	return &DelegatingPasswordEncoder{
		idForEncode:      idForEncode,
		passwordEncoders: passwordEncoders,
	}
}

// Encode 使用默认编码器编码密码
func (e *DelegatingPasswordEncoder) Encode(rawPassword string) string {
	encoder, exists := e.passwordEncoders[e.idForEncode]
	if !exists {
		panic(fmt.Sprintf("encoder not found for id: %s", e.idForEncode))
	}
	encoded := encoder.Encode(rawPassword)
	return fmt.Sprintf("%s{%s}", e.idForEncode, encoded)
}

// Matches 根据编码ID选择合适的编码器进行匹配
func (e *DelegatingPasswordEncoder) Matches(rawPassword, encodedPassword string) bool {
	id, password, err := e.extractId(encodedPassword)
	if err != nil {
		return false
	}

	encoder, exists := e.passwordEncoders[id]
	if !exists {
		return false
	}

	return encoder.Matches(rawPassword, password)
}

// extractId 从编码后的密码中提取编码器ID和实际密码
// 格式: {id}password
func (e *DelegatingPasswordEncoder) extractId(encodedPassword string) (string, string, error) {
	for i := 0; i < len(encodedPassword); i++ {
		if encodedPassword[i] == '{' {
			id := encodedPassword[:i]
			password := encodedPassword[i+1 : len(encodedPassword)-1]
			return id, password, nil
		}
	}
	return "", "", fmt.Errorf("invalid encoded password format")
}
