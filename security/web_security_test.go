package security

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"
)

func TestCsrfFilter(t *testing.T) {
	InitSecurityContext()
	ctx := context.Background()

	tokenRepo := NewCookieCsrfTokenRepository()
	csrfFilter := NewCsrfFilter(tokenRepo)

	t.Run("GET request should generate token", func(t *testing.T) {
		req := &mockSecurityRequest{method: "GET", uri: "/api/test"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}

		err := csrfFilter.DoFilter(ctx, req, resp, chain)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if !chain.called {
			t.Error("Expected filter chain to be called")
		}

		token, exists := req.GetAttribute("csrf.token")
		if !exists || token == nil {
			t.Error("Expected CSRF token to be generated")
		}
	})

	t.Run("POST request without token should fail", func(t *testing.T) {
		req := &mockSecurityRequest{method: "POST", uri: "/api/test"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}

		err := csrfFilter.DoFilter(ctx, req, resp, chain)
		if err == nil {
			t.Error("Expected error for missing CSRF token")
		}
		if !strings.Contains(err.Error(), "missing CSRF token") {
			t.Errorf("Expected 'missing CSRF token' error, got %v", err)
		}
	})

	t.Run("Exclude path should skip CSRF", func(t *testing.T) {
		csrfFilterWithExclude := NewCsrfFilter(tokenRepo)
		csrfFilterWithExclude.AddExcludePath("/public")

		req := &mockSecurityRequest{method: "POST", uri: "/public/api"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}

		err := csrfFilterWithExclude.DoFilter(ctx, req, resp, chain)
		if err != nil {
			t.Fatalf("Expected no error for excluded path, got %v", err)
		}
	})
}

func TestCookieCsrfTokenRepository(t *testing.T) {
	repo := NewCookieCsrfTokenRepository()
	ctx := context.Background()

	t.Run("Generate token", func(t *testing.T) {
		req := &mockSecurityRequest{method: "POST", uri: "/api/test"}
		token, err := repo.GenerateToken(ctx, req)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}
		if token.Value == "" {
			t.Error("Expected non-empty token value")
		}
		if token.Identifier != "/api/test" {
			t.Errorf("Expected identifier '/api/test', got '%s'", token.Identifier)
		}
	})

	t.Run("Validate token", func(t *testing.T) {
		req := &mockSecurityRequest{method: "POST", uri: "/api/test"}
		req.SetAttribute("_csrf_token", "test-token")
		valid := repo.ValidateToken(ctx, req, "test-token")
		if !valid {
			t.Error("Expected token to be valid")
		}

		valid = repo.ValidateToken(ctx, req, "wrong-token")
		if valid {
			t.Error("Expected token to be invalid")
		}
	})

	t.Run("Save token", func(t *testing.T) {
		req := &mockSecurityRequest{method: "POST", uri: "/api/test"}
		resp := &mockSecurityResponse{}
		token := &CsrfToken{Identifier: "/api/test", Value: "test-token-value"}
		repo.SaveToken(ctx, req, resp, token)

		cookie := resp.headers["Set-Cookie"]
		if cookie == "" {
			t.Error("Expected Set-Cookie header")
		}
		if !strings.Contains(cookie, "test-token-value") {
			t.Errorf("Expected token value in cookie, got %s", cookie)
		}
	})
}

func TestLogoutFilter(t *testing.T) {
	InitSecurityContext()
	ctx := context.Background()

	t.Run("Logout success", func(t *testing.T) {
		auth := NewAuthenticatedUsernamePasswordAuthenticationToken("admin", []string{"ROLE_ADMIN"})
		SetAuthentication(auth)

		req := &mockSecurityRequest{method: "POST", uri: "/logout"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}

		logoutFilter := NewLogoutFilter("/logout", []LogoutHandler{NewSecurityContextLogoutHandler()})

		err := logoutFilter.DoFilter(ctx, req, resp, chain)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if chain.called {
			t.Error("Filter chain should not be called after logout")
		}

		currentAuth := GetAuthentication()
		if currentAuth != nil {
			t.Error("Expected authentication to be cleared")
		}
	})

	t.Run("Non-logout URL should skip", func(t *testing.T) {
		req := &mockSecurityRequest{method: "POST", uri: "/api/test"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}

		logoutFilter := NewLogoutFilter("/logout", []LogoutHandler{})

		err := logoutFilter.DoFilter(ctx, req, resp, chain)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !chain.called {
			t.Error("Filter chain should be called for non-logout URL")
		}
	})

	t.Run("Custom success handler", func(t *testing.T) {
		auth := NewAuthenticatedUsernamePasswordAuthenticationToken("admin", []string{"ROLE_ADMIN"})
		SetAuthentication(auth)

		req := &mockSecurityRequest{method: "POST", uri: "/logout"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}

		customHandler := &mockLogoutSuccessHandler{targetCalled: "/custom"}
		logoutFilter := NewLogoutFilter("/logout", []LogoutHandler{})
		logoutFilter.SetSuccessHandler(customHandler)

		err := logoutFilter.DoFilter(ctx, req, resp, chain)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if resp.headers["Location"] != "/custom" {
			t.Errorf("Expected Location '/custom', got '%s'", resp.headers["Location"])
		}
	})
}

func TestLogoutSuccessHandler(t *testing.T) {
	ctx := context.Background()

	t.Run("DefaultLogoutSuccessHandler", func(t *testing.T) {
		handler := NewDefaultLogoutSuccessHandler("/login?logout")
		req := &mockSecurityRequest{method: "POST", uri: "/logout"}
		resp := &mockSecurityResponse{}
		auth := NewAuthenticatedUsernamePasswordAuthenticationToken("admin", []string{"ROLE_ADMIN"})

		handler.OnLogoutSuccess(ctx, req, resp, auth)

		if resp.statusCode != 302 {
			t.Errorf("Expected status code 302, got %d", resp.statusCode)
		}
		if resp.headers["Location"] != "/login?logout" {
			t.Errorf("Expected Location '/login?logout', got '%s'", resp.headers["Location"])
		}
	})

	t.Run("SimpleLogoutSuccessHandler", func(t *testing.T) {
		handler := NewSimpleLogoutSuccessHandler("/home")
		req := &mockSecurityRequest{method: "POST", uri: "/logout"}
		resp := &mockSecurityResponse{}
		auth := NewAuthenticatedUsernamePasswordAuthenticationToken("admin", []string{"ROLE_ADMIN"})

		handler.OnLogoutSuccess(ctx, req, resp, auth)

		if resp.statusCode != 302 {
			t.Errorf("Expected status code 302, got %d", resp.statusCode)
		}
		if resp.headers["Location"] != "/home" {
			t.Errorf("Expected Location '/home', got '%s'", resp.headers["Location"])
		}
	})
}

func TestCookieClearingLogoutHandler(t *testing.T) {
	ctx := context.Background()
	req := &mockSecurityRequest{method: "POST", uri: "/logout"}
	resp := &mockSecurityResponse{}
	auth := NewAuthenticatedUsernamePasswordAuthenticationToken("admin", []string{"ROLE_ADMIN"})

	handler := NewCookieClearingLogoutHandler("session_id")

	handler.Logout(ctx, req, resp, auth)

	cookie := resp.headers["Set-Cookie"]
	if cookie == "" {
		t.Error("Expected Set-Cookie header")
	}
	if !strings.Contains(cookie, "session_id") {
		t.Errorf("Expected cookie 'session_id' to be cleared, got %s", cookie)
	}
	if !strings.Contains(cookie, "Max-Age=0") {
		t.Errorf("Expected cookie to have Max-Age=0, got %s", cookie)
	}
}

func TestUsernamePasswordAuthenticationFilter(t *testing.T) {
	InitSecurityContext()
	ctx := context.Background()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder)
	authManager := NewProviderManager(authProvider)

	t.Run("Successful authentication", func(t *testing.T) {
		req := &mockSecurityRequest{
			method: "POST",
			uri:    "/login",
		}
		req.SetHeader("username", "admin")
		req.SetHeader("password", "admin123")
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}

		filter := NewUsernamePasswordAuthenticationFilterWithDefaults("/login", "/home", "/login?error", authManager)

		err := filter.DoFilter(ctx, req, resp, chain)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if resp.statusCode != 302 {
			t.Errorf("Expected status code 302, got %d", resp.statusCode)
		}
		if resp.headers["Location"] != "/home" {
			t.Errorf("Expected Location '/home', got '%s'", resp.headers["Location"])
		}

		auth := GetAuthentication()
		if auth == nil {
			t.Error("Expected authentication to be set")
		}
		if auth.Name() != "admin" {
			t.Errorf("Expected username 'admin', got '%s'", auth.Name())
		}
	})

	t.Run("Failed authentication", func(t *testing.T) {
		req := &mockSecurityRequest{
			method: "POST",
			uri:    "/login",
		}
		req.SetHeader("username", "admin")
		req.SetHeader("password", "wrongpassword")
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}

		filter := NewUsernamePasswordAuthenticationFilterWithDefaults("/login", "/home", "/login?error", authManager)

		err := filter.DoFilter(ctx, req, resp, chain)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if resp.statusCode != 401 {
			t.Errorf("Expected status code 401, got %d", resp.statusCode)
		}
	})

	t.Run("Non-login URL should skip", func(t *testing.T) {
		req := &mockSecurityRequest{
			method: "POST",
			uri:    "/api/test",
		}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}

		filter := NewUsernamePasswordAuthenticationFilterWithDefaults("/login", "/home", "/login?error", authManager)

		err := filter.DoFilter(ctx, req, resp, chain)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !chain.called {
			t.Error("Filter chain should be called for non-login URL")
		}
	})
}

func TestBasicAuthenticationFilter(t *testing.T) {
	InitSecurityContext()
	ctx := context.Background()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder)
	authManager := NewProviderManager(authProvider)

	t.Run("Successful Basic auth", func(t *testing.T) {
		encoded := base64.StdEncoding.EncodeToString([]byte("admin:admin123"))
		req := &mockSecurityRequest{
			method: "GET",
			uri:    "/api/test",
		}
		req.SetHeader("Authorization", "Basic "+encoded)
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}

		filter := NewBasicAuthenticationFilterWithRealm(authManager, "Test Realm")

		err := filter.DoFilter(ctx, req, resp, chain)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		auth := GetAuthentication()
		if auth == nil {
			t.Error("Expected authentication to be set")
		}
		if auth.Name() != "admin" {
			t.Errorf("Expected username 'admin', got '%s'", auth.Name())
		}
	})

	t.Run("Missing Authorization header returns error", func(t *testing.T) {
		req := &mockSecurityRequest{
			method: "GET",
			uri:    "/api/test",
		}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}

		filter := NewBasicAuthenticationFilterWithRealm(authManager, "Test Realm")

		err := filter.DoFilter(ctx, req, resp, chain)
		if err == nil {
			t.Error("Expected error for missing Authorization header")
		}
		if resp.statusCode != 401 {
			t.Errorf("Expected status 401, got %d", resp.statusCode)
		}
	})

	t.Run("Invalid credentials returns error", func(t *testing.T) {
		encoded := base64.StdEncoding.EncodeToString([]byte("admin:wrongpassword"))
		req := &mockSecurityRequest{
			method: "GET",
			uri:    "/api/test",
		}
		req.SetHeader("Authorization", "Basic "+encoded)
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}

		filter := NewBasicAuthenticationFilterWithRealm(authManager, "Test Realm")

		err := filter.DoFilter(ctx, req, resp, chain)
		if err == nil {
			t.Error("Expected error for invalid credentials")
		}
		if resp.statusCode != 401 {
			t.Errorf("Expected status 401, got %d", resp.statusCode)
		}
	})
}

func TestBasicAuthenticationEntryPoint(t *testing.T) {
	ctx := context.Background()

	t.Run("Send challenge", func(t *testing.T) {
		entryPoint := NewBasicAuthenticationEntryPointWithRealm("Test Realm")
		req := &mockSecurityRequest{method: "GET", uri: "/api/test"}
		resp := &mockSecurityResponse{}

		err := entryPoint.Commence(ctx, req, resp, ErrBadCredentials)
		if err == nil {
			t.Error("Expected error to be returned")
		}

		if resp.statusCode != 401 {
			t.Errorf("Expected status code 401, got %d", resp.statusCode)
		}

		wwwAuth := resp.headers["WWW-Authenticate"]
		if !strings.Contains(wwwAuth, `Basic realm="Test Realm"`) {
			t.Errorf("Expected WWW-Authenticate header with realm, got '%s'", wwwAuth)
		}
	})
}

func TestHttpSecurityConfiguration(t *testing.T) {
	InitSecurityContext()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder)
	authManager := NewProviderManager(authProvider)

	t.Run("Build with all features", func(t *testing.T) {
		chain, err := NewHttpSecurity().
			AuthenticationManager(authManager).
			Csrf().
			FormLogin("/api/login", "/dashboard").
			Logout("/api/logout").
			HttpBasic().
			Build()

		if err != nil {
			t.Fatalf("Failed to build security chain: %v", err)
		}
		if chain == nil {
			t.Error("Expected security chain to be non-nil")
		}
	})

	t.Run("Custom FormLogin URL", func(t *testing.T) {
		chain, err := NewHttpSecurity().
			AuthenticationManager(authManager).
			FormLogin("/custom-login").
			Build()

		if err != nil {
			t.Fatalf("Failed to build security chain: %v", err)
		}
		if chain == nil {
			t.Error("Expected security chain to be non-nil")
		}
	})

	t.Run("Custom Logout URL", func(t *testing.T) {
		chain, err := NewHttpSecurity().
			AuthenticationManager(authManager).
			Logout("/custom-logout", NewSimpleLogoutSuccessHandler("/goodbye")).
			Build()

		if err != nil {
			t.Fatalf("Failed to build security chain: %v", err)
		}
		if chain == nil {
			t.Error("Expected security chain to be non-nil")
		}
	})
}

func TestCsrfTokenManager(t *testing.T) {
	manager := NewCsrfTokenManager()

	t.Run("Generate and validate token", func(t *testing.T) {
		token := manager.GenerateToken("user1")
		if token == "" {
			t.Error("Expected non-empty token")
		}

		if !manager.ValidateToken("user1", token) {
			t.Error("Expected token to be valid")
		}

		if manager.ValidateToken("user1", "invalid-token") {
			t.Error("Expected invalid token to fail validation")
		}
	})

	t.Run("Remove token", func(t *testing.T) {
		token := manager.GenerateToken("user2")
		manager.RemoveToken("user2")

		if manager.ValidateToken("user2", token) {
			t.Error("Expected token to be invalid after removal")
		}
	})

	t.Run("Different users have different tokens", func(t *testing.T) {
		token1 := manager.GenerateToken("user3")
		token2 := manager.GenerateToken("user4")

		if token1 == token2 {
			t.Error("Expected different tokens for different users")
		}
	})
}

func TestSecurityContextLogoutHandler(t *testing.T) {
	ctx := context.Background()
	InitSecurityContext()

	handler := NewSecurityContextLogoutHandler()

	auth := NewAuthenticatedUsernamePasswordAuthenticationToken("admin", []string{"ROLE_ADMIN"})
	SetAuthentication(auth)

	req := &mockSecurityRequest{method: "POST", uri: "/logout"}
	resp := &mockSecurityResponse{}

	handler.Logout(ctx, req, resp, auth)

	currentAuth := GetAuthentication()
	if currentAuth != nil {
		t.Error("Expected authentication to be cleared")
	}
}

type mockLogoutSuccessHandler struct {
	targetCalled string
}

func (h *mockLogoutSuccessHandler) OnLogoutSuccess(ctx context.Context, request SecurityRequest, response SecurityResponse, authentication Authentication) {
	response.SetStatusCode(302)
	response.SetHeader("Location", h.targetCalled)
}

type mockSecurityFilterChain struct {
	called bool
}

func (m *mockSecurityFilterChain) DoFilter(ctx context.Context, request SecurityRequest, response SecurityResponse) error {
	m.called = true
	return nil
}
