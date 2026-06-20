package security

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

// TestPasswordEncoding 测试密码编码
// 演示不同密码编码器的使用
func TestPasswordEncoding(t *testing.T) {
	noop := NewNoOpPasswordEncoder()
	noopEncoded := noop.Encode("password123")
	if !noop.Matches("password123", noopEncoded) {
		t.Error("NoOp encoder should match")
	}

	bcrypt := NewBCryptPasswordEncoder()
	bcryptEncoded := bcrypt.Encode("password123")
	if bcryptEncoded == "password123" {
		t.Error("BCrypt encoder should hash the password")
	}
	if !bcrypt.Matches("password123", bcryptEncoded) {
		t.Error("BCrypt encoder should match")
	}

	standard := NewStandardPasswordEncoder("secret")
	standardEncoded := standard.Encode("password123")
	if !standard.Matches("password123", standardEncoded) {
		t.Error("Standard encoder should match")
	}

	delegating := NewDelegatingPasswordEncoder("bcrypt", map[string]PasswordEncoder{
		"bcrypt": bcrypt,
		"noop":   noop,
	})
	delegatedEncoded := delegating.Encode("password123")
	if !delegating.Matches("password123", delegatedEncoded) {
		t.Error("Delegating encoder should match")
	}
}

// TestMultipleProviderManager 测试多个认证提供者
// 演示多个认证提供者的使用
func TestMultipleProviderManager(t *testing.T) {
	InitSecurityContext()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})
	userDetailsService.CreateUser("user", "user123", []string{"ROLE_USER"})

	passwordEncoder := NewNoOpPasswordEncoder()
	daoProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder)
	anonymousProvider := NewAnonymousAuthenticationProvider()

	manager := NewProviderManager(daoProvider, anonymousProvider)

	ctx := context.Background()

	authToken := NewUsernamePasswordAuthenticationToken("admin", "admin123")
	authenticated, err := manager.Authenticate(ctx, authToken)
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}

	if authenticated.Name() != "admin" {
		t.Errorf("Expected username 'admin', got '%s'", authenticated.Name())
	}
}

// TestAccessDecisionVoters 测试访问决策投票者
// 演示不同类型的投票者
func TestAccessDecisionVoters(t *testing.T) {
	ctx := context.Background()
	auth := NewAuthenticatedUsernamePasswordAuthenticationToken("user", []string{"ROLE_USER", "ROLE_ADMIN"})

	voter := NewWebExpressionVoter()

	result := voter.Vote(ctx, auth, nil, []string{"permitAll"})
	if result != ACCESS_GRANTED {
		t.Errorf("permitAll should be granted")
	}

	result = voter.Vote(ctx, auth, nil, []string{"denyAll"})
	if result != ACCESS_DENIED {
		t.Errorf("denyAll should be denied")
	}

	result = voter.Vote(ctx, auth, nil, []string{"authenticated"})
	if result != ACCESS_GRANTED {
		t.Errorf("authenticated should be granted")
	}

	result = voter.Vote(ctx, auth, nil, []string{"hasRole('USER')"})
	if result != ACCESS_GRANTED {
		t.Errorf("hasRole('USER') should be granted")
	}

	result = voter.Vote(ctx, auth, nil, []string{"hasRole('SUPERUSER')"})
	if result != ACCESS_DENIED {
		t.Errorf("hasRole('SUPERUSER') should be denied")
	}

	result = voter.Vote(ctx, auth, nil, []string{"hasAnyRole('ADMIN','SUPERUSER')"})
	if result != ACCESS_GRANTED {
		t.Errorf("hasAnyRole('ADMIN','SUPERUSER') should be granted")
	}
}

// TestAccessDecisionManagers 测试访问决策管理器
// 演示不同的决策策略
func TestAccessDecisionManagers(t *testing.T) {
	ctx := context.Background()
	auth := NewAuthenticatedUsernamePasswordAuthenticationToken("user", []string{"ROLE_USER"})

	webExpressionVoter := NewWebExpressionVoter()
	roleVoter := NewRoleVoter()

	t.Run("AffirmativeBased", func(t *testing.T) {
		manager := NewAffirmativeBased(webExpressionVoter, roleVoter)

		err := manager.Decide(ctx, auth, nil, []string{"permitAll"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		err = manager.Decide(ctx, auth, nil, []string{"hasRole('USER')"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("UnanimousBased", func(t *testing.T) {
		manager := NewUnanimousBased(webExpressionVoter, roleVoter)

		err := manager.Decide(ctx, auth, nil, []string{"permitAll"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		err = manager.Decide(ctx, auth, nil, []string{"hasRole('USER')"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("ConsensusBased", func(t *testing.T) {
		manager := NewConsensusBased(webExpressionVoter, roleVoter)

		err := manager.Decide(ctx, auth, nil, []string{"permitAll"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		err = manager.Decide(ctx, auth, nil, []string{"hasRole('USER')"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

// TestSecurityContextThreadSafety 测试安全上下文线程安全
// 演示安全上下文的线程安全操作
func TestSecurityContextThreadSafety(t *testing.T) {
	InitSecurityContext()

	ctx := GetSecurityContext()
	if ctx.Authentication() != nil {
		t.Error("Expected no initial authentication")
	}

	auth := NewAuthenticatedUsernamePasswordAuthenticationToken("user", []string{"ROLE_USER"})
	ctx.SetAuthentication(auth)

	if ctx.Authentication() == nil {
		t.Error("Expected authentication to be set")
	}

	if ctx.Authentication().Name() != "user" {
		t.Errorf("Expected username 'user', got '%s'", ctx.Authentication().Name())
	}

	ctx.ClearAuthentication()
	if ctx.Authentication() != nil {
		t.Error("Expected no authentication after clear")
	}
}

// TestFilterInvocationSecurityMetadataSource 测试URL模式匹配
// 演示Ant风格路径模式的使用
func TestFilterInvocationSecurityMetadataSource(t *testing.T) {
	source := NewExpressionBasedFilterInvocationSecurityMetadataSource()

	source.AddMapping("/public/**", []string{"permitAll"})
	source.AddMapping("/admin/**", []string{"hasRole('ADMIN')"})
	source.AddMapping("/api/**", []string{"authenticated"})
	source.AddMapping("/health", []string{"permitAll"})

	ctx := context.Background()
	req := NewHttpRequestAdapter(&http.Request{})

	t.Run("exact match", func(t *testing.T) {
		req.request = &http.Request{URL: mustParseURL("/health")}
		attrs, err := source.GetAttributes(ctx, req)
		if err != nil {
			t.Fatalf("Error getting attributes: %v", err)
		}
		if len(attrs) != 1 || attrs[0] != "permitAll" {
			t.Errorf("Expected permitAll, got %v", attrs)
		}
	})

	t.Run("wildcard match", func(t *testing.T) {
		req.request = &http.Request{URL: mustParseURL("/admin/dashboard")}
		attrs, err := source.GetAttributes(ctx, req)
		if err != nil {
			t.Fatalf("Error getting attributes: %v", err)
		}
		if len(attrs) != 1 || attrs[0] != "hasRole('ADMIN')" {
			t.Errorf("Expected hasRole('ADMIN'), got %v", attrs)
		}
	})

	t.Run("no match", func(t *testing.T) {
		req.request = &http.Request{URL: mustParseURL("/unknown")}
		attrs, err := source.GetAttributes(ctx, req)
		if err != nil {
			t.Fatalf("Error getting attributes: %v", err)
		}
		if len(attrs) != 0 {
			t.Errorf("Expected empty attributes, got %v", attrs)
		}
	})
}

// TestHttpSecurityChain 测试HTTP安全链构建
// 演示完整的HTTP安全配置
func TestHttpSecurityChain(t *testing.T) {
	InitSecurityContext()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})
	userDetailsService.CreateUser("user", "user123", []string{"ROLE_USER"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder)
	authManager := NewProviderManager(authProvider)

	httpSecurity := NewHttpSecurity()
	httpSecurity.AuthenticationManager(authManager)

	metadataSource := NewExpressionBasedFilterInvocationSecurityMetadataSource()
	metadataSource.AddMapping("/public/**", []string{"permitAll"})
	metadataSource.AddMapping("/admin/**", []string{"hasRole('ADMIN')"})
	metadataSource.AddMapping("/api/**", []string{"authenticated"})
	httpSecurity.SecurityMetadataSource(metadataSource)

	securityFilterChain, err := httpSecurity.Build()
	if err != nil {
		t.Fatalf("Failed to build security: %v", err)
	}

	if securityFilterChain == nil {
		t.Error("Expected security filter chain")
	}
}

// TestSecurityIntegration 完整集成测试
// 演示从认证到授权的完整流程
func TestSecurityIntegration(t *testing.T) {
	InitSecurityContext()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN", "ROLE_USER"})
	userDetailsService.CreateUser("user", "user123", []string{"ROLE_USER"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder)
	authManager := NewProviderManager(authProvider)

	ctx := context.Background()

	authToken := NewUsernamePasswordAuthenticationToken("admin", "admin123")
	authenticated, err := authManager.Authenticate(ctx, authToken)
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}
	if !authenticated.Authenticated() {
		t.Error("Expected authenticated")
	}

	webExpressionVoter := NewWebExpressionVoter()
	authenticatedVoter := NewAuthenticatedVoter()
	roleVoter := NewRoleVoter()
	accessDecisionManager := NewAffirmativeBased(webExpressionVoter, authenticatedVoter, roleVoter)

	err = accessDecisionManager.Decide(ctx, authenticated, "/admin", []string{"hasRole('ADMIN')"})
	if err != nil {
		t.Errorf("Admin access should be granted: %v", err)
	}

	err = accessDecisionManager.Decide(ctx, authenticated, "/admin", []string{"hasRole('USER')"})
	if err != nil {
		t.Errorf("User access should be granted: %v", err)
	}
}

// ExampleUserDetailsService 使用示例：用户详情服务
// 展示如何创建和管理用户
func ExampleUserDetailsService() {
	service := NewInMemoryUserDetailsService()

	service.CreateUser("john", "password123", []string{"ROLE_USER"})
	service.CreateUser("jane", "password456", []string{"ROLE_USER", "ROLE_ADMIN"})

	ctx := context.Background()

	john, _ := service.LoadUserByUsername(ctx, "john")
	fmt.Printf("Username: %s\n", john.Username())
	fmt.Printf("Authorities: %v\n", john.Authorities())

	fmt.Printf("Total users: %d\n", service.UserCount())

	service.DeleteUser("john")
	fmt.Printf("Users after delete: %d\n", service.UserCount())
}

// ExamplePasswordEncoder 使用示例：密码编码
// 展示不同密码编码器的使用
func ExamplePasswordEncoder() {
	noop := NewNoOpPasswordEncoder()
	encoded := noop.Encode("secret")
	fmt.Printf("NoOp encoded: %s\n", encoded)
	fmt.Printf("NoOp matches: %v\n", noop.Matches("secret", encoded))

	bcrypt := NewBCryptPasswordEncoder()
	encoded = bcrypt.Encode("secret")
	fmt.Printf("BCrypt encoded: %s\n", encoded[:20]+"...")
	fmt.Printf("BCrypt matches: %v\n", bcrypt.Matches("secret", encoded))

	standard := NewStandardPasswordEncoder("mysecret")
	encoded = standard.Encode("secret")
	fmt.Printf("Standard encoded: %s\n", encoded[:20]+"...")
	fmt.Printf("Standard matches: %v\n", standard.Matches("secret", encoded))
}

// ExampleAuthentication 使用示例：用户认证
// 展示完整的认证流程
func ExampleAuthentication() {
	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN", "ROLE_USER"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder)
	authManager := NewProviderManager(authProvider)

	ctx := context.Background()

	authToken := NewUsernamePasswordAuthenticationToken("admin", "admin123")
	authenticated, err := authManager.Authenticate(ctx, authToken)
	if err != nil {
		fmt.Printf("Authentication failed: %v\n", err)
		return
	}

	fmt.Printf("Authentication successful!\n")
	fmt.Printf("Principal: %v\n", authenticated.Principal())
	fmt.Printf("Authorities: %v\n", authenticated.Authorities())
	fmt.Printf("Authenticated: %v\n", authenticated.Authenticated())
	fmt.Printf("Name: %s\n", authenticated.Name())
}

// TestRoleBasedDecision 测试基于角色的访问决策
// 展示基于角色的访问控制
func TestRoleBasedDecision(t *testing.T) {
	user := NewAuthenticatedUsernamePasswordAuthenticationToken("user", []string{"ROLE_USER", "ROLE_ADMIN"})

	webExpressionVoter := NewWebExpressionVoter()
	roleVoter := NewRoleVoter()
	accessDecisionManager := NewAffirmativeBased(webExpressionVoter, roleVoter)

	ctx := context.Background()

	type testCase struct {
		path       string
		attributes []string
		expected   string
	}

	testCases := []testCase{
		{"/public", []string{"permitAll"}, "granted"},
		{"/admin", []string{"hasRole('ADMIN')"}, "granted"},
		{"/user", []string{"hasRole('USER')"}, "granted"},
		{"/secret", []string{"hasRole('SUPERUSER')"}, "denied"},
	}

	for _, tc := range testCases {
		err := accessDecisionManager.Decide(ctx, user, tc.path, tc.attributes)
		if err != nil {
			fmt.Printf("%s with %v: %s\n", tc.path, tc.attributes, err)
		} else {
			fmt.Printf("%s with %v: granted\n", tc.path, tc.attributes)
		}
	}
}

// ExampleHttpSecurity 使用示例：HTTP安全配置
// 展示链式API配置HTTP安全规则
func ExampleHttpSecurity() {
	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})
	userDetailsService.CreateUser("user", "user123", []string{"ROLE_USER"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder)
	authManager := NewProviderManager(authProvider)

	metadataSource := NewExpressionBasedFilterInvocationSecurityMetadataSource()
	metadataSource.AddMapping("/public/**", []string{"permitAll"})
	metadataSource.AddMapping("/admin/**", []string{"hasRole('ADMIN')"})
	metadataSource.AddMapping("/api/**", []string{"authenticated"})

	httpSecurity := NewHttpSecurity()
	httpSecurity.
		AuthenticationManager(authManager).
		SecurityMetadataSource(metadataSource)

	chain, err := httpSecurity.Build()
	if err != nil {
		fmt.Printf("Failed to build: %v\n", err)
		return
	}

	fmt.Printf("Security chain built successfully: %v\n", chain != nil)
}

// 辅助函数
func mustParseURL(raw string) *url.URL {
	u, _ := url.Parse(raw)
	return u
}
