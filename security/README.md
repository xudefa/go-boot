# Security 模块

提供认证和授权功能,包括用户管理、JWT令牌、安全上下文等。

## 概述

security模块提供以下功能:

- 用户认证与授权
- JWT令牌生成与验证
- 安全上下文管理
- 多种认证提供者

## 主要组件

### 用户模型

```go
type User struct {
    Id           string   // 用户ID
    Username     string   // 用户名
    Password     string   // 密码(不序列化)
    Email        string   // 邮箱
    Roles        []string // 角色列表
    Enabled      bool     // 是否启用
    Locked       bool     // 是否锁定
    PasswordHash string   // 密码哈希(不序列化)
}
```

### 认证接口

```go
// 用户详情接口
type UserDetails interface {
    GetUserId() string
    GetUsername() string
    GetRoles() []string
    IsEnabled() bool
}

// 认证结果接口
type Authentication interface {
    GetPrincipal() interface{}
    GetCredentials() interface{}
    GetAuthorities() []string
    IsAuthenticated() bool
}

// 认证提供者接口
type AuthenticationProvider interface {
    Supports(request *AuthenticationRequest) bool
    Authenticate(request *AuthenticationRequest) (Authentication, error)
    LoadUserByUsername(username string) (UserDetails, error)
}
```

## 使用方法

### 创建认证管理器

```go
import "github.com/xudefa/go-boot/security"

// 创建内存用户服务
userService := security.NewInMemoryUserDetailsService()
userService.CreateUser(&security.User{
    Id:       "1",
    Username: "admin",
    Password: "password",
    Email:    "admin@example.com",
    Roles:    []string{"ADMIN", "USER"},
    Enabled:  true,
})

// 创建DAO认证提供者
daoProvider := security.NewDaoAuthenticationProvider(userService)

// 创建简单认证提供者
simpleProvider := security.NewSimpleAuthenticationProvider(userService)

// 创建认证管理器
authManager := security.NewAuthenticationManager()
authManager.RegisterProvider(daoProvider)

// 认证用户
auth, err := authManager.Authenticate("admin", "password")
if err != nil {
    // 认证失败
}
```

### 使用JWT

```go
// 创建JWT配置
cfg := &security.JwtConfig{
    SecretKey:     "your-secret-key",
    Issuer:        "your-app",
    Expiration:    24 * time.Hour,
    RefreshExpiry: 7 * 24 * time.Hour,
}

// 创建JWT管理器
jwtManager := security.NewJWTManager(cfg)

// 生成访问令牌
token, err := jwtManager.GenerateToken("1", "admin", []string{"ADMIN", "USER"})
if err != nil {
    // 生成失败
}

// 验证令牌
claims, err := jwtManager.ValidateToken(token)
if err != nil {
    // 验证失败
}

// 刷新令牌
newAccessToken, newRefreshToken, err := jwtManager.RefreshToken(refreshToken)
```

### 使用安全上下文

```go
// 创建安全上下文
securityContext := security.NewSecurityContext()

// 设置认证信息
securityContext.SetAuthentication(auth)

// 获取当前用户信息
userId := securityContext.GetUserId()
username := securityContext.GetUsername()
roles := securityContext.GetRoles()

// 检查角色
hasAdmin := securityContext.HasRole("ADMIN")

// 获取安全上下文持有者
holder := security.GetSecurityContextHolder()
ctx := holder.GetContext(requestId)
```

### 使用安全过滤器

```go
// 创建安全过滤器
securityFilter := security.NewSecurityFilter(jwtManager, authManager)

// 添加跳过路径
securityFilter.AddSkipPaths("/api/public/*")

// 检查是否跳过
if securityFilter.ShouldSkip(path) {
    // 跳过认证
}
```

### 密码操作

```go
// 密码哈希
hash := security.HashPassword("password")

// 密码比较
isValid := security.ComparePasswords(hash, "password")
```

## JWT配置说明

| 字段            | 类型            | 说明      | 默认值 |
|---------------|---------------|---------|-----|
| SecretKey     | string        | 密钥      | -   |
| Issuer        | string        | 发行者     | -   |
| Expiration    | time.Duration | 访问令牌有效期 | 24h |
| RefreshExpiry | time.Duration | 刷新令牌有效期 | 7d  |

## 错误类型

| 错误                  | 说明     |
|---------------------|--------|
| ErrTokenExpired     | 令牌已过期  |
| ErrTokenMalformed   | 令牌格式错误 |
| ErrTokenInvalid     | 令牌无效   |
| ErrTokenNotValidYet | 令牌尚未生效 |
| ErrSignatureInvalid | 签名无效   |

## 特性

- 支持多种认证提供者(Dao、Simple)
- 支持用户缓存
- 支持JWT访问令牌和刷新令牌
- 支持安全上下文线程安全
- 支持跳过路径配置