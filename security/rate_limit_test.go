package security

import (
	"context"
	"testing"
	"time"
)

func TestSlidingWindowRateLimiter(t *testing.T) {
	limiter := NewSlidingWindowRateLimiter(1*time.Second, 3)

	for i := 0; i < 3; i++ {
		if !limiter.Allow("user1") {
			t.Errorf("expected request %d to be allowed", i+1)
		}
	}

	if limiter.Allow("user1") {
		t.Error("expected 4th request to be denied")
	}

	if !limiter.Allow("user2") {
		t.Error("expected user2 to be allowed")
	}
}

func TestSlidingWindowRateLimiter_Cleanup(t *testing.T) {
	limiter := NewSlidingWindowRateLimiter(100*time.Millisecond, 10)

	limiter.Allow("user1")
	limiter.Allow("user2")

	time.Sleep(150 * time.Millisecond)
	limiter.Cleanup()

	limiter.mu.RLock()
	count := len(limiter.windows)
	limiter.mu.RUnlock()

	if count != 0 {
		t.Errorf("expected 0 windows after cleanup, got %d", count)
	}
}

func TestSlidingWindowRateLimiter_DefaultValues(t *testing.T) {
	limiter := NewSlidingWindowRateLimiter(0, -1)

	if limiter.windowSize <= 0 {
		t.Error("expected default window size")
	}
	if limiter.maxRequests <= 0 {
		t.Error("expected default max requests")
	}
}

func TestLeakyBucketRateLimiter(t *testing.T) {
	limiter := NewLeakyBucketRateLimiter(2, 10*time.Millisecond)

	if !limiter.Allow("user1") {
		t.Error("expected 1st request to be allowed")
	}
	if !limiter.Allow("user1") {
		t.Error("expected 2nd request to be allowed")
	}

	if limiter.Allow("user1") {
		t.Error("expected 3rd request to be denied")
	}

	time.Sleep(20 * time.Millisecond)

	if !limiter.Allow("user1") {
		t.Error("expected request after leak to be allowed")
	}
}

func TestLeakyBucketRateLimiter_Cleanup(t *testing.T) {
	limiter := NewLeakyBucketRateLimiter(2, 50*time.Millisecond)

	limiter.Allow("user1")

	time.Sleep(200 * time.Millisecond)
	limiter.Cleanup()

	limiter.mu.RLock()
	count := len(limiter.buckets)
	limiter.mu.RUnlock()

	if count != 0 {
		t.Errorf("expected 0 buckets after cleanup, got %d", count)
	}
}

func TestLeakyBucketRateLimiter_DefaultValues(t *testing.T) {
	limiter := NewLeakyBucketRateLimiter(0, 0)

	if limiter.capacity <= 0 {
		t.Error("expected default capacity")
	}
	if limiter.rate <= 0 {
		t.Error("expected default rate")
	}
}

func TestFixedWindowCounterRateLimiter(t *testing.T) {
	limiter := NewFixedWindowCounterRateLimiter(1*time.Second, 2)

	if !limiter.Allow("user1") {
		t.Error("expected 1st request to be allowed")
	}
	if !limiter.Allow("user1") {
		t.Error("expected 2nd request to be allowed")
	}

	if limiter.Allow("user1") {
		t.Error("expected 3rd request to be denied")
	}
}

func TestFixedWindowCounterRateLimiter_Cleanup(t *testing.T) {
	limiter := NewFixedWindowCounterRateLimiter(100*time.Millisecond, 10)

	limiter.Allow("user1")

	time.Sleep(150 * time.Millisecond)
	limiter.Cleanup()

	limiter.mu.RLock()
	count := len(limiter.counters)
	limiter.mu.RUnlock()

	if count != 0 {
		t.Errorf("expected 0 counters after cleanup, got %d", count)
	}
}

func TestFixedWindowCounterRateLimiter_DefaultValues(t *testing.T) {
	limiter := NewFixedWindowCounterRateLimiter(0, 0)

	if limiter.windowSize <= 0 {
		t.Error("expected default window size")
	}
	if limiter.maxRequests <= 0 {
		t.Error("expected default max requests")
	}
}

func TestStrategyRateLimiterAdapter(t *testing.T) {
	limiter := NewSlidingWindowRateLimiter(1*time.Second, 100)
	adapter := NewStrategyRateLimiterAdapter(limiter)

	if !adapter.Allow("user1") {
		t.Error("expected adapter to allow request")
	}
}

func TestEnhancedRateLimitFilter_Allow(t *testing.T) {
	limiter := NewSlidingWindowRateLimiter(1*time.Second, 100)
	adapter := NewStrategyRateLimiterAdapter(limiter)

	filter := NewEnhancedRateLimitFilter(adapter)

	req := &mockSecurityRequest{
		uri:    "/api/test",
		method: "GET",
		headers: map[string]string{
			"X-Real-IP": "192.168.1.1",
		},
	}
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	err := filter.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !chain.called {
		t.Error("expected chain to be called")
	}
}

func TestEnhancedRateLimitFilter_ExcludePath(t *testing.T) {
	limiter := NewSlidingWindowRateLimiter(1*time.Second, 0)
	adapter := NewStrategyRateLimiterAdapter(limiter)

	filter := NewEnhancedRateLimitFilter(adapter,
		WithExcludePaths("/health", "/metrics"),
	)

	req := &mockSecurityRequest{
		uri:    "/health",
		method: "GET",
	}
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	err := filter.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !chain.called {
		t.Error("expected excluded path to bypass rate limiting")
	}
}

func TestEnhancedRateLimitFilter_RateLimited(t *testing.T) {
	limiter := NewSlidingWindowRateLimiter(1*time.Second, 0)
	adapter := NewStrategyRateLimiterAdapter(limiter)

	filter := NewEnhancedRateLimitFilter(adapter)

	req := &mockSecurityRequest{
		uri:    "/api/test",
		method: "GET",
		headers: map[string]string{
			"X-Real-IP": "192.168.1.1",
		},
	}
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	err := filter.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp.statusCode != 429 {
		t.Errorf("expected status 429, got %d", resp.statusCode)
	}
	if chain.called {
		t.Error("expected chain not to be called when rate limited")
	}
}

func TestEnhancedRateLimitFilter_CustomCallback(t *testing.T) {
	limiter := NewSlidingWindowRateLimiter(1*time.Second, 0)
	adapter := NewStrategyRateLimiterAdapter(limiter)

	callbackCalled := false
	filter := NewEnhancedRateLimitFilter(adapter,
		WithOnRateLimit(func(ctx context.Context, request SecurityRequest, response SecurityResponse) {
			callbackCalled = true
			response.SetStatusCode(503)
		}),
	)

	req := &mockSecurityRequest{
		uri:    "/api/test",
		method: "GET",
	}
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	_ = filter.DoFilter(context.Background(), req, resp, chain)

	if !callbackCalled {
		t.Error("expected custom callback to be called")
	}
	if resp.statusCode != 503 {
		t.Errorf("expected status 503 from callback, got %d", resp.statusCode)
	}
}

func TestSecurityBuilder_BasicConfig(t *testing.T) {
	authManager := &testAuthManager{}
	userDetailsService := &testUserDetailsService{}
	passwordEncoder := NewNoOpPasswordEncoder()

	config := NewSecurityBuilder().
		AuthenticationManager(authManager).
		UserDetailsService(userDetailsService).
		PasswordEncoder(passwordEncoder).
		EnableAnonymous().
		EnableHttpBasic().
		Build()

	if config == nil {
		t.Fatal("expected non-nil config")
	}

	http := NewHttpSecurity()
	err := config.Configure(http)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = http.Build()
	if err != nil {
		t.Errorf("unexpected build error: %v", err)
	}
}

func TestSecurityBuilder_WithFilters(t *testing.T) {
	filter1 := &testSecurityFilter{}
	filter2 := &testSecurityFilter{}

	config := NewSecurityBuilder().
		AuthenticationManager(&testAuthManager{}).
		AddFilter(filter1).
		AddFilterAfter(filter2, filter1).
		Build()

	http := NewHttpSecurity()
	err := config.Configure(http)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = http.Build()
	if err != nil {
		t.Errorf("unexpected build error: %v", err)
	}
}

func TestSecurityBuilder_FormLogin(t *testing.T) {
	config := NewSecurityBuilder().
		AuthenticationManager(&testAuthManager{}).
		EnableFormLogin("/login", "/home").
		EnableCsrf().
		EnableLogout("/logout").
		Build()

	http := NewHttpSecurity()
	err := config.Configure(http)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	chain, err := http.Build()
	if err != nil {
		t.Errorf("unexpected build error: %v", err)
	}
	if chain == nil {
		t.Error("expected non-nil filter chain")
	}
}

func TestSecurityBuilder_LogoutWithHandler(t *testing.T) {
	handler := &testLogoutSuccessHandler{}
	config := NewSecurityBuilder().
		AuthenticationManager(&testAuthManager{}).
		EnableLogout("/logout", handler).
		Build()

	http := NewHttpSecurity()
	err := config.Configure(http)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	chain, err := http.Build()
	if err != nil {
		t.Errorf("unexpected build error: %v", err)
	}
	if chain == nil {
		t.Error("expected non-nil filter chain")
	}
}

func TestSecurityBuilder_AddFilterBefore(t *testing.T) {
	filter1 := &testSecurityFilter{}
	filter2 := &testSecurityFilter{}

	config := NewSecurityBuilder().
		AuthenticationManager(&testAuthManager{}).
		AddFilter(filter1).
		AddFilterBefore(filter2, filter1).
		Build()

	http := NewHttpSecurity()
	err := config.Configure(http)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = http.Build()
	if err != nil {
		t.Errorf("unexpected build error: %v", err)
	}
}

func TestSecurityBuilder_AccessDecisionManager(t *testing.T) {
	accessDecisionMgr := &testAccessDecisionManager{}
	config := NewSecurityBuilder().
		AuthenticationManager(&testAuthManager{}).
		AccessDecisionManager(accessDecisionMgr).
		Build()

	http := NewHttpSecurity()
	err := config.Configure(http)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = http.Build()
	if err != nil {
		t.Errorf("unexpected build error: %v", err)
	}
}

func TestSecurityBuilder_EmptyBuild(t *testing.T) {
	config := NewSecurityBuilder().Build()
	http := NewHttpSecurity()
	err := config.Configure(http)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = http.Build()
	if err == nil {
		t.Error("expected build to fail without auth manager")
	}
}

// Test mock implementations

type testAuth struct {
	principal   string
	credentials string
}

func (a *testAuth) Principal() any        { return a.principal }
func (a *testAuth) Credentials() any      { return a.credentials }
func (a *testAuth) Authorities() []string { return []string{"ROLE_USER"} }
func (a *testAuth) Authenticated() bool   { return true }
func (a *testAuth) Name() string          { return a.principal }

type testAuthManager struct{}

func (m *testAuthManager) Authenticate(ctx context.Context, auth Authentication) (Authentication, error) {
	return &testAuth{principal: "testuser", credentials: "testpass"}, nil
}

type testUserDetailsService struct{}

func (s *testUserDetailsService) LoadUserByUsername(ctx context.Context, username string) (UserDetails, error) {
	return nil, nil
}

type testAccessDecisionManager struct{}

func (m *testAccessDecisionManager) Decide(ctx context.Context, auth Authentication, object any, attrs []string) error {
	return nil
}

type testSecurityFilter struct{}

func (f *testSecurityFilter) DoFilter(ctx context.Context, req SecurityRequest, resp SecurityResponse, chain SecurityFilterChain) error {
	return chain.DoFilter(ctx, req, resp)
}

type testLogoutSuccessHandler struct{}

func (h *testLogoutSuccessHandler) OnLogoutSuccess(ctx context.Context, req SecurityRequest, resp SecurityResponse, auth Authentication) {
	resp.SetStatusCode(302)
	resp.SetHeader("Location", "/login")
}
