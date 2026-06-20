# Security 模块

提供全面的安全认证和授权功能，灵感来源于Spring Security，采用接口设计，方便快速集成到项目中。

## 核心组件

### 认证(Authentication)

| 接口/类型 | 说明 |
|---------|------|
| `Authentication` | 认证信息接口，包含principal、credentials、authorities |
| `AuthenticationManager` | 认证管理器，处理认证请求 |
| `AuthenticationProvider` | 认证提供者接口，定义认证逻辑 |
| `UsernamePasswordAuthenticationToken` | 用户名密码认证令牌 |

### 授权(Authorization)

| 接口/类型 | 说明 |
|---------|------|
| `AccessDecisionManager` | 访问决策管理器，决定资源访问权限 |
| `AccessDecisionVoter` | 访问决策投票者，投票决定访问权限 |
| `WebExpressionVoter` | Web表达式投票者，支持hasRole、hasAuthority等表达式 |
| `RoleVoter` | 角色投票者，检查角色权限 |
| `AuthenticatedVoter` | 认证投票者，检查认证级别 |

### 用户管理(UserDetails)

| 接口/类型 | 说明 |
|---------|------|
| `UserDetails` | 用户详情接口 |
| `UserDetailsService` | 用户详情服务接口 |
| `InMemoryUserDetails` | 内存用户详情实现 |
| `InMemoryUserDetailsService` | 内存用户详情服务 |
| `DaoAuthenticationProvider` | DAO认证提供者 |

### 密码编码(PasswordEncoder)

| 类型 | 说明 |
|-----|------|
| `NoOpPasswordEncoder` | 不编码，用于开发测试 |
| `BCryptPasswordEncoder` | SHA256密码编码器 |
| `StandardPasswordEncoder` | 标准密码编码器，支持密钥加盐 |
| `DelegatingPasswordEncoder` | 委托密码编码器，支持多编码器 |

### 过滤器(Filter)

| 类型 | 说明 |
|-----|------|
| `FilterChainProxy` | 过滤器链代理 |
| `SecurityContextHolderFilter` | 安全上下文持有者过滤器 |
| `AnonymousAuthenticationFilter` | 匿名认证过滤器 |
| `ExceptionTranslationFilter` | 异常转换过滤器 |
| `FilterSecurityInterceptor` | 过滤器安全拦截器 |
| `BasicAuthenticationFilter` | Basic认证过滤器 |
| `JwtAuthenticationFilter` | JWT认证过滤器（jwt模块提供） |

**过滤器执行顺序**（自动配置时）：
1. CORS过滤器（如启用）
2. 限流过滤器（如启用）
3. JWT认证过滤器（如启用jwt模块）
4. SecurityContextHolder过滤器
5. Anonymous认证过滤器
6. ExceptionTranslation过滤器
7. FilterSecurityInterceptor

### HTTP安全(HttpSecurity)

链式API配置HTTP安全规则：

```go
security := NewHttpSecurity()
chain, err := security.
    AuthenticationManager(authManager).
    UserDetailsService(userDetailsService).
    PasswordEncoder(passwordEncoder).
    Anonymous().
    AuthorizeRequests(func(reg AuthorizeRequests) {
        reg.AntMatchers("/public/**").PermitAll()
        reg.AntMatchers("/admin/**").HasRole("ADMIN")
        reg.AnyRequest().Authenticated()
    }).
    ExceptionHandling(handler, entryPoint).
    Build()
```

## 快速开始

### 1. 基本配置

```go
userDetailsService := NewInMemoryUserDetailsService()
userDetailsService.CreateUser("admin", encodePassword("123456"), []string{"ROLE_ADMIN", "ROLE_USER"})

passwordEncoder := NewBCryptPasswordEncoder()
authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder)
authManager := NewProviderManager(authProvider)

security := NewHttpSecurity()
chain, err := security.
    AuthenticationManager(authManager).
    UserDetailsService(userDetailsService).
    PasswordEncoder(passwordEncoder).
    Anonymous().
    Build()
```

### 2. 集成HTTP服务器

```go
handler := NewSecurityFilterChainHandler(chain, yourAppHandler)
http.ListenAndServe(":8080", handler)
```

### 3. URL权限配置

```go
security.AuthorizeRequests(func(reg AuthorizeRequests) {
    reg.AntMatchers("/public/**").PermitAll()
    reg.AntMatchers("/admin/**").HasRole("ADMIN")
    reg.AntMatchers("/user/**").HasAnyRole("ADMIN", "USER")
    reg.AntMatchers("/api/**").hasAuthority("ACCESS_API")
    reg.AnyRequest().Authenticated()
})
```

## Web表达式支持

| 表达式 | 说明 |
|-------|------|
| `permitAll` | 允许所有人访问 |
| `denyAll` | 拒绝所有人访问 |
| `authenticated` | 仅允许已认证用户 |
| `hasRole('ROLE')` | 检查是否具有指定角色 |
| `hasAnyRole('ROLE1','ROLE2')` | 检查是否具有任一角色 |
| `hasAuthority('AUTHORITY')` | 检查是否具有指定权限 |
| `hasAnyAuthority('AUTH1','AUTH2')` | 检查是否具有任一权限 |

## 访问决策策略

| 类型 | 说明 |
|-----|------|
| `AffirmativeBased` | 肯定优先，只要有投票者授予权限就通过 |
| `UnanimousBased` | 一致通过，所有投票者都不拒绝才通过 |
| `ConsensusBased` | 多数通过，多数投票者授予权限才通过 |

## 错误处理

| 错误 | 说明 |
|-----|------|
| `ErrAuthenticationFailed` | 认证失败 |
| `ErrAccessDenied` | 访问被拒绝 |
| `ErrUserNotFound` | 用户不存在 |
| `ErrBadCredentials` | 凭证错误 |

## 安全上下文

线程安全的认证信息存储：

```go
ctx := GetSecurityContext()
ctx.SetAuthentication(auth)
auth := ctx.Authentication()
ctx.ClearAuthentication()
```

全局认证快捷方法：

```go
SetAuthentication(auth)
GetAuthentication() Authentication
ClearAuthentication()
```

## 自动配置

通过 `condition.OnProperty("security.enabled", "true")` 控制，启用时自动配置以下组件：

```yaml
# application.yaml
security:
  enabled: true
  cors:
    enabled: false
    allowed-origins: "*"
    allowed-methods: "GET,POST,PUT,DELETE,OPTIONS"
    allowed-headers: "Content-Type,Authorization,X-Requested-With"
    allow-credentials: false
    max-age: 3600
  rate-limit:
    enabled: false
    rate: 100          # 每秒请求数
    burst: 200         # 突发请求数
    exclude-paths: "/health,/actuator/health"
  rules: "/public/**->permitAll,/admin/**->hasRole('ADMIN'),/api/**->authenticated"
  login-url: "/login"
```

**自动配置注册的 Bean**：

| Bean ID | 说明 | 是否可自定义 |
|---------|------|-------------|
| `userDetailsService` | 用户详情服务 | 是 |
| `passwordEncoder` | 密码编码器（BCrypt） | 是 |
| `authenticationManager` | 认证管理器 | 否 |
| `securityFilterChain` | 安全过滤器链 | 是 |

**配置选项说明**：

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `security.enabled` | 是否启用安全模块 | `false` |
| `security.cors.enabled` | 是否启用 CORS | `false` |
| `security.cors.allowed-origins` | 允许的源 | `*` |
| `security.cors.allowed-methods` | 允许的方法 | `GET,POST,PUT,DELETE,OPTIONS` |
| `security.cors.allowed-headers` | 允许的请求头 | `Content-Type,Authorization,X-Requested-With` |
| `security.cors.allow-credentials` | 是否允许凭证 | `false` |
| `security.cors.max-age` | 预检请求缓存时间（秒） | `3600` |
| `security.rate-limit.enabled` | 是否启用限流 | `false` |
| `security.rate-limit.rate` | 限流速率（每秒） | `100` |
| `security.rate-limit.burst` | 突发容量 | `200` |
| `security.rate-limit.exclude-paths` | 排除路径 | `/health,/actuator/health` |
| `security.rules` | URL安全规则 | 空 |
| `security.login-url` | 登录页面URL | `/login` |

### 与 JWT 模块集成

当同时启用安全模块和 JWT 模块时，JWT 认证过滤器会自动集成到安全过滤器链中：

```yaml
# application.yaml
security:
  enabled: true
jwt:
  enabled: true
  secret-key: my-secret-key-123456
  exclude-paths: /login,/register
```

**集成效果**：
- JWT 过滤器会自动添加到安全过滤器链中
- 支持配置排除路径（不需要 JWT 认证的路径）
- 认证成功后，用户信息自动设置到安全上下文