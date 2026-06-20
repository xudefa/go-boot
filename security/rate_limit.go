package security

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimiter 限流器接口
type RateLimiter interface {
	// Allow 判断是否允许请求
	Allow(key string) bool
	// Cleanup 清理过期数据
	Cleanup()
}

// SlidingWindowRateLimiter 滑动窗口限流器
type SlidingWindowRateLimiter struct {
	windowSize  time.Duration
	maxRequests int
	mu          sync.RWMutex
	windows     map[string]*slidingWindow
}

type slidingWindow struct {
	count     int
	startTime time.Time
}

// NewSlidingWindowRateLimiter 创建滑动窗口限流器
func NewSlidingWindowRateLimiter(windowSize time.Duration, maxRequests int) *SlidingWindowRateLimiter {
	if windowSize <= 0 {
		windowSize = 1 * time.Minute
	}
	// maxRequests=0 表示拒绝所有（用于测试或特殊场景）
	if maxRequests < 0 {
		maxRequests = 100
	}
	return &SlidingWindowRateLimiter{
		windowSize:  windowSize,
		maxRequests: maxRequests,
		windows:     make(map[string]*slidingWindow),
	}
}

// Allow 判断是否允许请求
func (r *SlidingWindowRateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	// maxRequests=0 表示拒绝所有
	if r.maxRequests <= 0 {
		return false
	}

	now := time.Now()
	window, exists := r.windows[key]

	if !exists || now.Sub(window.startTime) > r.windowSize {
		// 新窗口
		r.windows[key] = &slidingWindow{
			count:     1,
			startTime: now,
		}
		return true
	}

	if window.count < r.maxRequests {
		window.count++
		return true
	}

	return false
}

// Cleanup 清理过期窗口
func (r *SlidingWindowRateLimiter) Cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for key, window := range r.windows {
		if now.Sub(window.startTime) > r.windowSize {
			delete(r.windows, key)
		}
	}
}

// LeakyBucketRateLimiter 漏桶限流器
type LeakyBucketRateLimiter struct {
	capacity int
	rate     time.Duration
	mu       sync.RWMutex
	buckets  map[string]*leakyBucket
}

type leakyBucket struct {
	tokens   int
	lastLeak time.Time
}

// NewLeakyBucketRateLimiter 创建漏桶限流器
func NewLeakyBucketRateLimiter(capacity int, rate time.Duration) *LeakyBucketRateLimiter {
	if capacity <= 0 {
		capacity = 100
	}
	if rate <= 0 {
		rate = 100 * time.Millisecond
	}
	return &LeakyBucketRateLimiter{
		capacity: capacity,
		rate:     rate,
		buckets:  make(map[string]*leakyBucket),
	}
}

// Allow 判断是否允许请求
func (r *LeakyBucketRateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	bucket, exists := r.buckets[key]

	if !exists {
		r.buckets[key] = &leakyBucket{
			tokens:   1,
			lastLeak: now,
		}
		return true
	}

	// 漏水
	elapsed := now.Sub(bucket.lastLeak)
	leaked := int(elapsed / r.rate)
	if leaked > 0 {
		bucket.tokens = max(0, bucket.tokens-leaked)
		bucket.lastLeak = now
	}

	if bucket.tokens < r.capacity {
		bucket.tokens++
		return true
	}

	return false
}

// Cleanup 清理空桶
func (r *LeakyBucketRateLimiter) Cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for key, bucket := range r.buckets {
		if now.Sub(bucket.lastLeak) > r.rate*time.Duration(r.capacity) {
			delete(r.buckets, key)
		}
	}
}

// FixedWindowCounterRateLimiter 固定窗口计数器限流器
type FixedWindowCounterRateLimiter struct {
	windowSize  time.Duration
	maxRequests int
	mu          sync.RWMutex
	counters    map[string]*fixedWindowCounter
}

type fixedWindowCounter struct {
	count       int
	windowStart time.Time
}

// NewFixedWindowCounterRateLimiter 创建固定窗口计数器限流器
func NewFixedWindowCounterRateLimiter(windowSize time.Duration, maxRequests int) *FixedWindowCounterRateLimiter {
	if windowSize <= 0 {
		windowSize = 1 * time.Minute
	}
	if maxRequests <= 0 {
		maxRequests = 100
	}
	return &FixedWindowCounterRateLimiter{
		windowSize:  windowSize,
		maxRequests: maxRequests,
		counters:    make(map[string]*fixedWindowCounter),
	}
}

// Allow 判断是否允许请求
func (r *FixedWindowCounterRateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	// maxRequests=0 表示拒绝所有
	if r.maxRequests <= 0 {
		return false
	}

	now := time.Now()
	counter, exists := r.counters[key]

	if !exists || now.Sub(counter.windowStart) > r.windowSize {
		// 新窗口
		r.counters[key] = &fixedWindowCounter{
			count:       1,
			windowStart: now,
		}
		return true
	}

	if counter.count < r.maxRequests {
		counter.count++
		return true
	}

	return false
}

// Cleanup 清理过期计数器
func (r *FixedWindowCounterRateLimiter) Cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for key, counter := range r.counters {
		if now.Sub(counter.windowStart) > r.windowSize {
			delete(r.counters, key)
		}
	}
}

// RateLimitStrategy 限流策略接口（用于 RateLimitFilter）
type RateLimitStrategy interface {
	// Allow 判断是否允许请求
	Allow(key string) bool
}

// StrategyRateLimiterAdapter 限流器适配器
type StrategyRateLimiterAdapter struct {
	limiter RateLimiter
}

// NewStrategyRateLimiterAdapter 创建适配器
func NewStrategyRateLimiterAdapter(limiter RateLimiter) *StrategyRateLimiterAdapter {
	return &StrategyRateLimiterAdapter{limiter: limiter}
}

// Allow 判断是否允许请求
func (a *StrategyRateLimiterAdapter) Allow(key string) bool {
	return a.limiter.Allow(key)
}

// EnhancedRateLimitFilter 增强版限流过滤器
type EnhancedRateLimitFilter struct {
	strategy     RateLimitStrategy
	excludePaths []string
	onRateLimit  func(ctx context.Context, request SecurityRequest, response SecurityResponse)
}

// EnhancedRateLimitOption 增强限流配置选项
type EnhancedRateLimitOption func(*EnhancedRateLimitFilter)

// WithExcludePaths 设置排除路径
func WithExcludePaths(paths ...string) EnhancedRateLimitOption {
	return func(f *EnhancedRateLimitFilter) {
		f.excludePaths = append(f.excludePaths, paths...)
	}
}

// WithOnRateLimit 设置限流回调
func WithOnRateLimit(fn func(ctx context.Context, request SecurityRequest, response SecurityResponse)) EnhancedRateLimitOption {
	return func(f *EnhancedRateLimitFilter) {
		f.onRateLimit = fn
	}
}

// NewEnhancedRateLimitFilter 创建增强版限流过滤器
func NewEnhancedRateLimitFilter(strategy RateLimitStrategy, opts ...EnhancedRateLimitOption) *EnhancedRateLimitFilter {
	f := &EnhancedRateLimitFilter{
		strategy: strategy,
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

// DoFilter 处理限流
func (f *EnhancedRateLimitFilter) DoFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain SecurityFilterChain) error {
	uri := request.GetURI()
	for _, path := range f.excludePaths {
		if uri == path || (len(path) > 0 && uri[:len(path)] == path) {
			return chain.DoFilter(ctx, request, response)
		}
	}

	key := request.GetHeader("X-Real-IP")
	if key == "" {
		key = request.GetHeader("X-Forwarded-For")
	}
	if key == "" {
		key = "global"
	}

	if !f.strategy.Allow(key) {
		if f.onRateLimit != nil {
			f.onRateLimit(ctx, request, response)
		} else {
			response.SetStatusCode(429)
			if writeErr := response.Write([]byte(`{"error":"rate_limited","message":"too many requests"}`)); writeErr != nil {
				fmt.Printf("[go-boot] failed to write rate limit response: %v\n", writeErr)
			}
		}
		return nil
	}

	return chain.DoFilter(ctx, request, response)
}

// SecurityBuilder 安全配置构建器
type SecurityBuilder struct {
	authManager        AuthenticationManager
	userDetailsService UserDetailsService
	passwordEncoder    PasswordEncoder
	accessDecisionMgr  AccessDecisionManager
	filters            []filterEntry
	anonymous          bool
	csrf               bool
	formLogin          *formLoginConfig
	httpBasic          bool
	logoutConfig       *logoutConfig
}

type filterEntry struct {
	filter SecurityFilter
	before SecurityFilter
	after  SecurityFilter
}

type formLoginConfig struct {
	processingUrl     string
	defaultSuccessUrl string
}

type logoutConfig struct {
	url            string
	successHandler LogoutSuccessHandler
}

// NewSecurityBuilder 创建安全配置构建器
func NewSecurityBuilder() *SecurityBuilder {
	return &SecurityBuilder{}
}

// AuthenticationManager 设置认证管理器
func (b *SecurityBuilder) AuthenticationManager(manager AuthenticationManager) *SecurityBuilder {
	b.authManager = manager
	return b
}

// UserDetailsService 设置用户详情服务
func (b *SecurityBuilder) UserDetailsService(service UserDetailsService) *SecurityBuilder {
	b.userDetailsService = service
	return b
}

// PasswordEncoder 设置密码编码器
func (b *SecurityBuilder) PasswordEncoder(encoder PasswordEncoder) *SecurityBuilder {
	b.passwordEncoder = encoder
	return b
}

// AccessDecisionManager 设置访问决策管理器
func (b *SecurityBuilder) AccessDecisionManager(manager AccessDecisionManager) *SecurityBuilder {
	b.accessDecisionMgr = manager
	return b
}

// AddFilter 添加过滤器
func (b *SecurityBuilder) AddFilter(filter SecurityFilter) *SecurityBuilder {
	b.filters = append(b.filters, filterEntry{filter: filter})
	return b
}

// AddFilterBefore 在指定过滤器前添加
func (b *SecurityBuilder) AddFilterBefore(filter SecurityFilter, before SecurityFilter) *SecurityBuilder {
	b.filters = append(b.filters, filterEntry{filter: filter, before: before})
	return b
}

// AddFilterAfter 在指定过滤器后添加
func (b *SecurityBuilder) AddFilterAfter(filter SecurityFilter, after SecurityFilter) *SecurityBuilder {
	b.filters = append(b.filters, filterEntry{filter: filter, after: after})
	return b
}

// EnableAnonymous 启用匿名访问
func (b *SecurityBuilder) EnableAnonymous() *SecurityBuilder {
	b.anonymous = true
	return b
}

// EnableCsrf 启用 CSRF 保护
func (b *SecurityBuilder) EnableCsrf() *SecurityBuilder {
	b.csrf = true
	return b
}

// EnableFormLogin 启用表单登录
func (b *SecurityBuilder) EnableFormLogin(processingUrl string, defaultSuccessUrl ...string) *SecurityBuilder {
	cfg := &formLoginConfig{processingUrl: processingUrl}
	if len(defaultSuccessUrl) > 0 {
		cfg.defaultSuccessUrl = defaultSuccessUrl[0]
	}
	b.formLogin = cfg
	return b
}

// EnableHttpBasic 启用 HTTP Basic 认证
func (b *SecurityBuilder) EnableHttpBasic() *SecurityBuilder {
	b.httpBasic = true
	return b
}

// EnableLogout 启用登出
func (b *SecurityBuilder) EnableLogout(url string, successHandler ...LogoutSuccessHandler) *SecurityBuilder {
	cfg := &logoutConfig{url: url}
	if len(successHandler) > 0 {
		cfg.successHandler = successHandler[0]
	}
	b.logoutConfig = cfg
	return b
}

// Build 构建安全配置
func (b *SecurityBuilder) Build() SecurityConfig {
	return &builtSecurityConfig{builder: b}
}

type builtSecurityConfig struct {
	builder *SecurityBuilder
}

func (c *builtSecurityConfig) Configure(http HttpSecurity) error {
	if c.builder.authManager != nil {
		http.AuthenticationManager(c.builder.authManager)
	}
	if c.builder.userDetailsService != nil {
		http.UserDetailsService(c.builder.userDetailsService)
	}
	if c.builder.passwordEncoder != nil {
		http.PasswordEncoder(c.builder.passwordEncoder)
	}
	if c.builder.accessDecisionMgr != nil {
		http.AccessDecisionManager(c.builder.accessDecisionMgr)
	}

	for _, entry := range c.builder.filters {
		if entry.before != nil {
			http.AddFilterBefore(entry.filter, entry.before)
		} else if entry.after != nil {
			http.AddFilterAfter(entry.filter, entry.after)
		} else {
			http.AddFilter(entry.filter)
		}
	}

	if c.builder.anonymous {
		http.Anonymous()
	}
	if c.builder.csrf {
		http.Csrf()
	}
	if c.builder.formLogin != nil {
		if c.builder.formLogin.defaultSuccessUrl != "" {
			http.FormLogin(c.builder.formLogin.processingUrl, c.builder.formLogin.defaultSuccessUrl)
		} else {
			http.FormLogin(c.builder.formLogin.processingUrl)
		}
	}
	if c.builder.httpBasic {
		http.HttpBasic()
	}
	if c.builder.logoutConfig != nil {
		if c.builder.logoutConfig.successHandler != nil {
			http.Logout(c.builder.logoutConfig.url, c.builder.logoutConfig.successHandler)
		} else {
			http.Logout(c.builder.logoutConfig.url)
		}
	}

	return nil
}

// max 辅助函数
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
