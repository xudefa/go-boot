// Package security 提供认证和授权功能,包括JWT令牌管理等.
// 本文件包含JWT相关功能.
package security

import (
	"crypto/subtle"
	"errors"
	"github.com/google/uuid"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT相关错误
var (
	ErrTokenExpired     = errors.New("token has expired")      // 令牌已过期
	ErrTokenMalformed   = errors.New("token is malformed")     // 令牌格式错误
	ErrTokenInvalid     = errors.New("token is invalid")       // 令牌无效
	ErrTokenNotValidYet = errors.New("token is not valid yet") // 令牌尚未生效
	ErrSignatureInvalid = errors.New("signature is invalid")   // 签名无效
)

// JWTClaims 是JWT声明,包含用户信息和有效期等.
type JWTClaims struct {
	jwt.RegisteredClaims
	UserId   string   `json:"user_id"`   // 用户ID
	Username string   `json:"username"`  // 用户名
	Roles    []string `json:"roles"`     // 角色列表
	ExpireAt int64    `json:"expire_at"` // 过期时间戳
}

// JWTManager 是JWT管理器,负责令牌的生成和验证.
type JWTManager struct {
	secretKey     []byte        // 密钥
	issuer        string        // 发行者
	expiration    time.Duration // 访问令牌有效期
	refreshExpiry time.Duration // 刷新令牌有效期
}

// JwtConfig 是JWT配置,可从配置文件加载.
type JwtConfig struct {
	SecretKey     string        `mapstructure:"secretKey" json:"-"`                       // 密钥(不在JSON中输出)
	Issuer        string        `mapstructure:"issuer" json:"issuer,omitempty"`           // 发行者
	Expiration    time.Duration `mapstructure:"expireTime" json:"expireTime,omitempty"`   // 访问令牌有效期
	RefreshExpiry time.Duration `mapstructure:"refreshTime" json:"refreshTime,omitempty"` // 刷新令牌有效期
}

// NewJWTManager 创建新的JWT管理器
// 默认访问令牌有效期为24小时,刷新令牌有效期为7天
func NewJWTManager(cfg *JwtConfig) *JWTManager {
	if cfg.Expiration == 0 {
		cfg.Expiration = 24 * time.Hour
	}
	if cfg.RefreshExpiry == 0 {
		cfg.RefreshExpiry = 7 * 24 * time.Hour
	}

	return &JWTManager{
		secretKey:     []byte(cfg.SecretKey),
		issuer:        cfg.Issuer,
		expiration:    cfg.Expiration,
		refreshExpiry: cfg.RefreshExpiry,
	}
}

// GenerateToken 生成访问令牌
// 包含用户ID、用户名、角色列表和过期时间
func (m *JWTManager) GenerateToken(userId, username string, roles []string) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(), // 唯一ID
			Issuer:    m.issuer,            // 发行者
			Subject:   userId,              // 主题(用户ID)
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expiration)),
			NotBefore: jwt.NewNumericDate(now),
		},
		UserId:   userId,
		Username: username,
		Roles:    roles,
		ExpireAt: now.Add(m.expiration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// GenerateRefreshToken 生成刷新令牌
// 刷新令牌有效期更长,但不包含用户信息
func (m *JWTManager) GenerateRefreshToken(userId string) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Issuer:    m.issuer,
			Subject:   userId,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshExpiry)),
			NotBefore: jwt.NewNumericDate(now),
		},
		UserId:   userId,
		ExpireAt: now.Add(m.refreshExpiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// ValidateToken 验证令牌并返回声明
// 验证签名、过期时间等
func (m *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return m.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrTokenMalformed
		}
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, ErrSignatureInvalid
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	if claims.NotBefore != nil && claims.NotBefore.Time.After(time.Now()) {
		return nil, ErrTokenNotValidYet
	}

	return claims, nil
}

// RefreshToken 使用刷新令牌获取新的访问令牌和刷新令牌
func (m *JWTManager) RefreshToken(refreshToken string) (string, string, error) {
	claims, err := m.ValidateToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	if claims.UserId == "" {
		return "", "", ErrTokenInvalid
	}

	newAccessToken, err := m.GenerateToken(claims.UserId, claims.Username, claims.Roles)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := m.GenerateRefreshToken(claims.UserId)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

// ExtractClaims 提取令牌声明(同ValidateToken)
func (m *JWTManager) ExtractClaims(tokenString string) (*JWTClaims, error) {
	return m.ValidateToken(tokenString)
}

// ComparePasswords 使用常量时间比较,防止时序攻击
func ComparePasswords(hashedPassword, plainPassword string) bool {
	return subtle.ConstantTimeCompare([]byte(hashedPassword), []byte(plainPassword)) == 1
}
