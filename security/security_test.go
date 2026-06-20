package security

import (
	"context"
	"testing"
)

func TestInMemoryUserDetailsService(t *testing.T) {
	service := NewInMemoryUserDetailsService()

	service.CreateUser("admin", "admin123", []string{"ROLE_ADMIN", "ROLE_USER"})
	service.CreateUser("user", "user123", []string{"ROLE_USER"})

	if service.UserCount() != 2 {
		t.Errorf("Expected 2 users, got %d", service.UserCount())
	}

	ctx := context.Background()
	user, err := service.LoadUserByUsername(ctx, "admin")
	if err != nil {
		t.Fatalf("Failed to load user: %v", err)
	}

	if user.Username() != "admin" {
		t.Errorf("Expected username 'admin', got '%s'", user.Username())
	}

	if user.Password() != "admin123" {
		t.Errorf("Expected password 'admin123', got '%s'", user.Password())
	}

	expectedAuthorities := []string{"ROLE_ADMIN", "ROLE_USER"}
	if len(user.Authorities()) != len(expectedAuthorities) {
		t.Errorf("Expected %d authorities, got %d", len(expectedAuthorities), len(user.Authorities()))
	}

	_, err = service.LoadUserByUsername(ctx, "nonexistent")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestNoOpPasswordEncoder(t *testing.T) {
	encoder := NewNoOpPasswordEncoder()

	password := "myPassword"
	encoded := encoder.Encode(password)

	if encoded != password {
		t.Errorf("Expected encoded password to be same as raw password")
	}

	if !encoder.Matches(password, encoded) {
		t.Errorf("Expected password to match")
	}

	if encoder.Matches("wrongPassword", encoded) {
		t.Errorf("Expected wrong password to not match")
	}
}

func TestBCryptPasswordEncoder(t *testing.T) {
	encoder := NewBCryptPasswordEncoder()

	password := "myPassword"
	encoded := encoder.Encode(password)

	if encoded == password {
		t.Errorf("Expected encoded password to be different from raw password")
	}

	if !encoder.Matches(password, encoded) {
		t.Errorf("Expected password to match")
	}

	if encoder.Matches("wrongPassword", encoded) {
		t.Errorf("Expected wrong password to not match")
	}
}

func TestStandardPasswordEncoder(t *testing.T) {
	encoder := NewStandardPasswordEncoder("mySecret")

	password := "myPassword"
	encoded := encoder.Encode(password)

	if encoded == password {
		t.Errorf("Expected encoded password to be different from raw password")
	}

	if !encoder.Matches(password, encoded) {
		t.Errorf("Expected password to match")
	}

	if encoder.Matches("wrongPassword", encoded) {
		t.Errorf("Expected wrong password to not match")
	}

	encoder2 := NewStandardPasswordEncoder("differentSecret")
	encoded2 := encoder2.Encode(password)

	if encoded == encoded2 {
		t.Errorf("Expected different secrets to produce different encoded passwords")
	}
}

func TestDelegatingPasswordEncoder(t *testing.T) {
	noopEncoder := NewNoOpPasswordEncoder()
	bcryptEncoder := NewBCryptPasswordEncoder()

	encoders := map[string]PasswordEncoder{
		"noop":   noopEncoder,
		"bcrypt": bcryptEncoder,
	}

	encoder := NewDelegatingPasswordEncoder("bcrypt", encoders)

	password := "myPassword"
	encoded := encoder.Encode(password)

	if !encoder.Matches(password, encoded) {
		t.Errorf("Expected password to match")
	}

	if encoder.Matches("wrongPassword", encoded) {
		t.Errorf("Expected wrong password to not match")
	}
}

func TestDaoAuthenticationProvider(t *testing.T) {
	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	provider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder)

	ctx := context.Background()

	authToken := NewUsernamePasswordAuthenticationToken("admin", "admin123")
	authenticated, err := provider.Authenticate(ctx, authToken)
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}

	if !authenticated.Authenticated() {
		t.Errorf("Expected authentication to be successful")
	}

	if authenticated.Name() != "admin" {
		t.Errorf("Expected username 'admin', got '%s'", authenticated.Name())
	}

	wrongAuthToken := NewUsernamePasswordAuthenticationToken("admin", "wrongPassword")
	_, err = provider.Authenticate(ctx, wrongAuthToken)
	if err != ErrBadCredentials {
		t.Errorf("Expected ErrBadCredentials, got %v", err)
	}

	nonExistentAuthToken := NewUsernamePasswordAuthenticationToken("nonexistent", "password")
	_, err = provider.Authenticate(ctx, nonExistentAuthToken)
	if err != ErrBadCredentials {
		t.Errorf("Expected ErrBadCredentials, got %v", err)
	}
}

func TestProviderManager(t *testing.T) {
	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder)
	manager := NewProviderManager(authProvider)

	ctx := context.Background()

	authToken := NewUsernamePasswordAuthenticationToken("admin", "admin123")
	authenticated, err := manager.Authenticate(ctx, authToken)
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}

	if !authenticated.Authenticated() {
		t.Errorf("Expected authentication to be successful")
	}

	wrongAuthToken := NewUsernamePasswordAuthenticationToken("admin", "wrongPassword")
	_, err = manager.Authenticate(ctx, wrongAuthToken)
	if err != ErrBadCredentials {
		t.Errorf("Expected ErrBadCredentials, got %v", err)
	}
}

func TestWebExpressionVoter(t *testing.T) {
	voter := NewWebExpressionVoter()

	auth := NewAuthenticatedUsernamePasswordAuthenticationToken("user", []string{"ROLE_USER", "ROLE_ADMIN"})

	ctx := context.Background()

	result := voter.Vote(ctx, auth, "/test", []string{"permitAll"})
	if result != ACCESS_GRANTED {
		t.Errorf("Expected ACCESS_GRANTED for permitAll")
	}

	result = voter.Vote(ctx, auth, "/test", []string{"denyAll"})
	if result != ACCESS_DENIED {
		t.Errorf("Expected ACCESS_DENIED for denyAll")
	}

	result = voter.Vote(ctx, auth, "/test", []string{"authenticated"})
	if result != ACCESS_GRANTED {
		t.Errorf("Expected ACCESS_GRANTED for authenticated")
	}

	result = voter.Vote(ctx, nil, "/test", []string{"authenticated"})
	if result != ACCESS_DENIED {
		t.Errorf("Expected ACCESS_DENIED for unauthenticated user")
	}

	result = voter.Vote(ctx, auth, "/test", []string{"hasRole('ADMIN')"})
	if result != ACCESS_GRANTED {
		t.Errorf("Expected ACCESS_GRANTED for hasRole('ADMIN')")
	}

	result = voter.Vote(ctx, auth, "/test", []string{"hasRole('USER')"})
	if result != ACCESS_GRANTED {
		t.Errorf("Expected ACCESS_GRANTED for hasRole('USER')")
	}

	result = voter.Vote(ctx, auth, "/test", []string{"hasRole('GUEST')"})
	if result != ACCESS_DENIED {
		t.Errorf("Expected ACCESS_DENIED for hasRole('GUEST')")
	}

	result = voter.Vote(ctx, auth, "/test", []string{"hasAnyRole('ADMIN','USER')"})
	if result != ACCESS_GRANTED {
		t.Errorf("Expected ACCESS_GRANTED for hasAnyRole('ADMIN','USER')")
	}

	result = voter.Vote(ctx, auth, "/test", []string{"hasAuthority('ROLE_ADMIN')"})
	if result != ACCESS_GRANTED {
		t.Errorf("Expected ACCESS_GRANTED for hasAuthority('ROLE_ADMIN')")
	}
}

func TestRoleVoter(t *testing.T) {
	voter := NewRoleVoter()

	auth := NewAuthenticatedUsernamePasswordAuthenticationToken("user", []string{"ROLE_ADMIN"})

	ctx := context.Background()

	result := voter.Vote(ctx, auth, "/test", []string{"ROLE_ADMIN"})
	if result != ACCESS_GRANTED {
		t.Errorf("Expected ACCESS_GRANTED for ROLE_ADMIN")
	}

	result = voter.Vote(ctx, auth, "/test", []string{"ROLE_USER"})
	if result != ACCESS_DENIED {
		t.Errorf("Expected ACCESS_DENIED for ROLE_USER")
	}

	result = voter.Vote(ctx, nil, "/test", []string{"ROLE_ADMIN"})
	if result != ACCESS_DENIED {
		t.Errorf("Expected ACCESS_DENIED for nil authentication")
	}

	result = voter.Vote(ctx, auth, "/test", []string{"permitAll"})
	if result != ACCESS_ABSTAIN {
		t.Errorf("Expected ACCESS_ABSTAIN for non-role attribute")
	}
}

func TestAuthenticatedVoter(t *testing.T) {
	voter := NewAuthenticatedVoter()

	auth := NewAuthenticatedUsernamePasswordAuthenticationToken("user", []string{"ROLE_USER"})

	ctx := context.Background()

	result := voter.Vote(ctx, auth, "/test", []string{"IS_AUTHENTICATED_FULLY"})
	if result != ACCESS_GRANTED {
		t.Errorf("Expected ACCESS_GRANTED for IS_AUTHENTICATED_FULLY")
	}

	result = voter.Vote(ctx, auth, "/test", []string{"IS_AUTHENTICATED_REMEMBERED"})
	if result != ACCESS_GRANTED {
		t.Errorf("Expected ACCESS_GRANTED for IS_AUTHENTICATED_REMEMBERED")
	}

	result = voter.Vote(ctx, nil, "/test", []string{"IS_AUTHENTICATED_ANONYMOUSLY"})
	if result != ACCESS_GRANTED {
		t.Errorf("Expected ACCESS_GRANTED for IS_AUTHENTICATED_ANONYMOUSLY")
	}
}

func TestAffirmativeBased(t *testing.T) {
	webExpressionVoter := NewWebExpressionVoter()
	authenticatedVoter := NewAuthenticatedVoter()
	roleVoter := NewRoleVoter()

	manager := NewAffirmativeBased(webExpressionVoter, authenticatedVoter, roleVoter)

	auth := NewAuthenticatedUsernamePasswordAuthenticationToken("user", []string{"ROLE_USER"})

	ctx := context.Background()

	err := manager.Decide(ctx, auth, "/test", []string{"permitAll"})
	if err != nil {
		t.Errorf("Expected no error for permitAll, got %v", err)
	}

	err = manager.Decide(ctx, auth, "/test", []string{"denyAll"})
	if err != ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied for denyAll, got %v", err)
	}

	err = manager.Decide(ctx, auth, "/test", []string{"hasRole('USER')"})
	if err != nil {
		t.Errorf("Expected no error for hasRole('USER'), got %v", err)
	}

	err = manager.Decide(ctx, auth, "/test", []string{"hasRole('ADMIN')"})
	if err != ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied for hasRole('ADMIN'), got %v", err)
	}
}

func TestUnanimousBased(t *testing.T) {
	webExpressionVoter := NewWebExpressionVoter()
	authenticatedVoter := NewAuthenticatedVoter()
	roleVoter := NewRoleVoter()

	manager := NewUnanimousBased(webExpressionVoter, authenticatedVoter, roleVoter)

	auth := NewAuthenticatedUsernamePasswordAuthenticationToken("user", []string{"ROLE_USER"})

	ctx := context.Background()

	err := manager.Decide(ctx, auth, "/test", []string{"permitAll"})
	if err != nil {
		t.Errorf("Expected no error for permitAll, got %v", err)
	}

	err = manager.Decide(ctx, auth, "/test", []string{"denyAll"})
	if err != ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied for denyAll, got %v", err)
	}

	err = manager.Decide(ctx, auth, "/test", []string{"authenticated", "hasRole('USER')"})
	if err != nil {
		t.Errorf("Expected no error for matching conditions, got %v", err)
	}
}

func TestSecurityContext(t *testing.T) {
	InitSecurityContext()

	ctx := GetSecurityContext()

	if ctx.Authentication() != nil {
		t.Errorf("Expected no authentication initially")
	}

	auth := NewAuthenticatedUsernamePasswordAuthenticationToken("user", []string{"ROLE_USER"})
	ctx.SetAuthentication(auth)

	if ctx.Authentication() == nil {
		t.Errorf("Expected authentication to be set")
	}

	if ctx.Authentication().Name() != "user" {
		t.Errorf("Expected username 'user', got '%s'", ctx.Authentication().Name())
	}

	ctx.ClearAuthentication()

	if ctx.Authentication() != nil {
		t.Errorf("Expected authentication to be cleared")
	}
}

func TestExpressionBasedFilterInvocationSecurityMetadataSource(t *testing.T) {
	source := NewExpressionBasedFilterInvocationSecurityMetadataSource()

	source.AddMapping("GET::/api/public/**", []string{"permitAll"})
	source.AddMapping("POST::/api/admin/**", []string{"hasRole('ADMIN')"})
	source.AddMapping("**::/api/user/**", []string{"hasRole('USER')"})

	ctx := context.Background()

	request := &mockSecurityRequest{
		method: "GET",
		uri:    "/api/public/data",
	}

	attributes, err := source.GetAttributes(ctx, request)
	if err != nil {
		t.Fatalf("Failed to get attributes: %v", err)
	}

	if len(attributes) != 1 || attributes[0] != "permitAll" {
		t.Errorf("Expected permitAll, got %v", attributes)
	}

	request = &mockSecurityRequest{
		method: "POST",
		uri:    "/api/admin/users",
	}

	attributes, err = source.GetAttributes(ctx, request)
	if err != nil {
		t.Fatalf("Failed to get attributes: %v", err)
	}

	if len(attributes) != 1 || attributes[0] != "hasRole('ADMIN')" {
		t.Errorf("Expected hasRole('ADMIN'), got %v", attributes)
	}

	request = &mockSecurityRequest{
		method: "DELETE",
		uri:    "/api/user/profile",
	}

	attributes, err = source.GetAttributes(ctx, request)
	if err != nil {
		t.Fatalf("Failed to get attributes: %v", err)
	}

	if len(attributes) != 1 || attributes[0] != "hasRole('USER')" {
		t.Errorf("Expected hasRole('USER'), got %v", attributes)
	}
}

type mockSecurityRequest struct {
	method  string
	uri     string
	attrs   map[string]any
	headers map[string]string
}

func (m *mockSecurityRequest) GetMethod() string {
	return m.method
}

func (m *mockSecurityRequest) GetURI() string {
	return m.uri
}

func (m *mockSecurityRequest) GetHeader(key string) string {
	if m.headers == nil {
		return ""
	}
	return m.headers[key]
}

func (m *mockSecurityRequest) SetHeader(key, value string) {
	if m.headers == nil {
		m.headers = make(map[string]string)
	}
	m.headers[key] = value
}

func (m *mockSecurityRequest) SetAttribute(key string, value any) {
	if m.attrs == nil {
		m.attrs = make(map[string]any)
	}
	m.attrs[key] = value
}

func (m *mockSecurityRequest) GetAttribute(key string) (any, bool) {
	if m.attrs == nil {
		return nil, false
	}
	val, exists := m.attrs[key]
	return val, exists
}

type mockSecurityResponse struct {
	statusCode int
	headers    map[string]string
	body       []byte
}

func (m *mockSecurityResponse) SetStatusCode(code int) {
	m.statusCode = code
}

func (m *mockSecurityResponse) SetHeader(key, value string) {
	if m.headers == nil {
		m.headers = make(map[string]string)
	}
	m.headers[key] = value
}

func (m *mockSecurityResponse) Write(data []byte) error {
	m.body = data
	return nil
}
