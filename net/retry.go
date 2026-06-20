package net

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// RetryStrategy 重试策略接口
type RetryStrategy interface {
	// ShouldRetry 判断是否应该重试
	ShouldRetry(resp *HttpResponse, err error, attempt int) bool
	// Delay 计算下次重试的延迟时间
	Delay(attempt int) time.Duration
}

// RetryOption 重试配置选项
type RetryOption func(*RetryConfig)

// RetryConfig 重试配置
type RetryConfig struct {
	maxAttempts int
	strategy    RetryStrategy
	onRetry     func(attempt int, resp *HttpResponse, err error)
}

// WithMaxAttempts 设置最大重试次数
func WithMaxAttempts(n int) RetryOption {
	return func(c *RetryConfig) {
		c.maxAttempts = n
	}
}

// WithRetryStrategy 设置重试策略
func WithRetryStrategy(strategy RetryStrategy) RetryOption {
	return func(c *RetryConfig) {
		c.strategy = strategy
	}
}

// WithOnRetry 设置重试回调
func WithOnRetry(fn func(attempt int, resp *HttpResponse, err error)) RetryOption {
	return func(c *RetryConfig) {
		c.onRetry = fn
	}
}

// ExponentialBackoff 指数退避重试策略
type ExponentialBackoff struct {
	baseDelay       time.Duration
	maxDelay        time.Duration
	retryableStatus []int
}

// NewExponentialBackoff 创建指数退避策略
func NewExponentialBackoff(baseDelay, maxDelay time.Duration, retryableStatus ...int) *ExponentialBackoff {
	if baseDelay <= 0 {
		baseDelay = 100 * time.Millisecond
	}
	if maxDelay <= 0 {
		maxDelay = 10 * time.Second
	}
	if len(retryableStatus) == 0 {
		retryableStatus = []int{500, 502, 503, 504} // 默认重试服务器错误
	}
	return &ExponentialBackoff{
		baseDelay:       baseDelay,
		maxDelay:        maxDelay,
		retryableStatus: retryableStatus,
	}
}

// ShouldRetry 判断是否应该重试
func (e *ExponentialBackoff) ShouldRetry(resp *HttpResponse, err error, attempt int) bool {
	// 网络错误总是重试
	if err != nil {
		return true
	}

	// 检查状态码是否可重试
	if resp != nil {
		for _, status := range e.retryableStatus {
			if resp.StatusCode == status {
				return true
			}
		}
	}

	return false
}

// Delay 计算延迟时间（指数退避 + 抖动）
func (e *ExponentialBackoff) Delay(attempt int) time.Duration {
	// 指数退避：baseDelay * 2^attempt
	delay := e.baseDelay * time.Duration(math.Pow(2, float64(attempt)))

	// 限制最大延迟
	if delay > e.maxDelay {
		delay = e.maxDelay
	}

	// 添加随机抖动（±50%）
	jitter := delay / 2
	delay = delay - jitter + time.Duration(rand.Int63n(int64(jitter*2)))

	return delay
}

// FixedDelay 固定延迟重试策略
type FixedDelay struct {
	delay           time.Duration
	retryableStatus []int
}

// NewFixedDelay 创建固定延迟策略
func NewFixedDelay(delay time.Duration, retryableStatus ...int) *FixedDelay {
	if delay <= 0 {
		delay = 1 * time.Second
	}
	if len(retryableStatus) == 0 {
		retryableStatus = []int{500, 502, 503, 504}
	}
	return &FixedDelay{
		delay:           delay,
		retryableStatus: retryableStatus,
	}
}

// ShouldRetry 判断是否应该重试
func (f *FixedDelay) ShouldRetry(resp *HttpResponse, err error, attempt int) bool {
	if err != nil {
		return true
	}
	if resp != nil {
		for _, status := range f.retryableStatus {
			if resp.StatusCode == status {
				return true
			}
		}
	}
	return false
}

// Delay 返回固定延迟
func (f *FixedDelay) Delay(attempt int) time.Duration {
	return f.delay
}

// RetryableClient 支持重试的 HTTP 客户端
type RetryableClient struct {
	client *NetClient
	config RetryConfig
}

// NewRetryableClient 创建可重试的 HTTP 客户端
func NewRetryableClient(client *NetClient, opts ...RetryOption) *RetryableClient {
	cfg := RetryConfig{
		maxAttempts: 3,
		strategy:    NewExponentialBackoff(100*time.Millisecond, 10*time.Second),
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	return &RetryableClient{
		client: client,
		config: cfg,
	}
}

// Get 发送 GET 请求（支持重试）
func (c *RetryableClient) Get(ctx context.Context, url string, opts ...RequestOption) (*HttpResponse, error) {
	return c.doWithRetry(ctx, func(ctx context.Context) (*HttpResponse, error) {
		return c.client.Get(ctx, url, opts...)
	})
}

// Post 发送 POST 请求（支持重试）
func (c *RetryableClient) Post(ctx context.Context, url string, body any, opts ...RequestOption) (*HttpResponse, error) {
	return c.doWithRetry(ctx, func(ctx context.Context) (*HttpResponse, error) {
		return c.client.Post(ctx, url, body, opts...)
	})
}

// Put 发送 PUT 请求（支持重试）
func (c *RetryableClient) Put(ctx context.Context, url string, body any, opts ...RequestOption) (*HttpResponse, error) {
	return c.doWithRetry(ctx, func(ctx context.Context) (*HttpResponse, error) {
		return c.client.Put(ctx, url, body, opts...)
	})
}

// Delete 发送 DELETE 请求（支持重试）
func (c *RetryableClient) Delete(ctx context.Context, url string, opts ...RequestOption) (*HttpResponse, error) {
	return c.doWithRetry(ctx, func(ctx context.Context) (*HttpResponse, error) {
		return c.client.Delete(ctx, url, opts...)
	})
}

// Patch 发送 PATCH 请求（支持重试）
func (c *RetryableClient) Patch(ctx context.Context, url string, body any, opts ...RequestOption) (*HttpResponse, error) {
	return c.doWithRetry(ctx, func(ctx context.Context) (*HttpResponse, error) {
		return c.client.Patch(ctx, url, body, opts...)
	})
}

// Head 发送 HEAD 请求（支持重试）
func (c *RetryableClient) Head(ctx context.Context, url string, opts ...RequestOption) (*HttpResponse, error) {
	return c.doWithRetry(ctx, func(ctx context.Context) (*HttpResponse, error) {
		return c.client.Head(ctx, url, opts...)
	})
}

// Options 发送 OPTIONS 请求（支持重试）
func (c *RetryableClient) Options(ctx context.Context, url string, opts ...RequestOption) (*HttpResponse, error) {
	return c.doWithRetry(ctx, func(ctx context.Context) (*HttpResponse, error) {
		return c.client.Options(ctx, url, opts...)
	})
}

// Do 执行自定义请求（支持重试）
func (c *RetryableClient) Do(ctx context.Context, req any) (*HttpResponse, error) {
	return c.doWithRetry(ctx, func(ctx context.Context) (*HttpResponse, error) {
		return c.client.Do(ctx, req)
	})
}

// doWithRetry 执行带重试的请求
func (c *RetryableClient) doWithRetry(ctx context.Context, fn func(context.Context) (*HttpResponse, error)) (*HttpResponse, error) {
	var lastResp *HttpResponse
	var lastErr error

	for attempt := 0; attempt < c.config.maxAttempts; attempt++ {
		lastResp, lastErr = fn(ctx)

		// 判断是否需要重试
		if !c.config.strategy.ShouldRetry(lastResp, lastErr, attempt) {
			return lastResp, lastErr
		}

		// 调用重试回调
		if c.config.onRetry != nil {
			c.config.onRetry(attempt, lastResp, lastErr)
		}

		// 如果是最后一次尝试，不等待直接返回
		if attempt == c.config.maxAttempts-1 {
			break
		}

		// 等待下次重试
		delay := c.config.strategy.Delay(attempt)
		select {
		case <-ctx.Done():
			return lastResp, ctx.Err()
		case <-time.After(delay):
			// 继续重试
		}
	}

	return lastResp, lastErr
}

// Close 关闭客户端
func (c *RetryableClient) Close() error {
	return c.client.Close()
}

// CircuitBreaker 断路器
type CircuitBreaker struct {
	maxFailures  int
	resetTimeout time.Duration
	failures     int
	lastFailure  time.Time
	state        CircuitState
}

// CircuitState 断路器状态
type CircuitState int

const (
	CircuitClosed   CircuitState = iota // 关闭状态（正常请求）
	CircuitOpen                         // 打开状态（拒绝请求）
	CircuitHalfOpen                     // 半开状态（允许探测请求）
)

// NewCircuitBreaker 创建断路器
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	if maxFailures <= 0 {
		maxFailures = 5
	}
	if resetTimeout <= 0 {
		resetTimeout = 30 * time.Second
	}
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        CircuitClosed,
	}
}

// AllowRequest 判断是否允许请求
func (cb *CircuitBreaker) AllowRequest() bool {
	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		// 检查是否超过重置超时
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = CircuitHalfOpen
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	}
	return false
}

// RecordSuccess 记录成功
func (cb *CircuitBreaker) RecordSuccess() {
	cb.failures = 0
	cb.state = CircuitClosed
}

// RecordFailure 记录失败
func (cb *CircuitBreaker) RecordFailure() {
	cb.failures++
	cb.lastFailure = time.Now()

	if cb.failures >= cb.maxFailures {
		cb.state = CircuitOpen
	}
}

// GetState 获取当前状态
func (cb *CircuitBreaker) GetState() CircuitState {
	return cb.state
}

// CircuitBreakerClient 带断路器的 HTTP 客户端
type CircuitBreakerClient struct {
	client   *NetClient
	breaker  *CircuitBreaker
	fallback func(ctx context.Context) (*HttpResponse, error)
}

// CircuitBreakerOption 断路器配置选项
type CircuitBreakerOption func(*CircuitBreakerConfig)

// CircuitBreakerConfig 断路器配置
type CircuitBreakerConfig struct {
	maxFailures  int
	resetTimeout time.Duration
	fallback     func(ctx context.Context) (*HttpResponse, error)
}

// WithCircuitMaxFailures 设置最大失败次数
func WithCircuitMaxFailures(n int) CircuitBreakerOption {
	return func(c *CircuitBreakerConfig) {
		c.maxFailures = n
	}
}

// WithCircuitResetTimeout 设置重置超时
func WithCircuitResetTimeout(d time.Duration) CircuitBreakerOption {
	return func(c *CircuitBreakerConfig) {
		c.resetTimeout = d
	}
}

// WithFallback 设置降级函数
func WithFallback(fn func(ctx context.Context) (*HttpResponse, error)) CircuitBreakerOption {
	return func(c *CircuitBreakerConfig) {
		c.fallback = fn
	}
}

// NewCircuitBreakerClient 创建带断路器的 HTTP 客户端
func NewCircuitBreakerClient(client *NetClient, opts ...CircuitBreakerOption) *CircuitBreakerClient {
	cfg := CircuitBreakerConfig{
		maxFailures:  5,
		resetTimeout: 30 * time.Second,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	return &CircuitBreakerClient{
		client:   client,
		breaker:  NewCircuitBreaker(cfg.maxFailures, cfg.resetTimeout),
		fallback: cfg.fallback,
	}
}

// Get 发送 GET 请求（带断路器保护）
func (c *CircuitBreakerClient) Get(ctx context.Context, url string, opts ...RequestOption) (*HttpResponse, error) {
	return c.execute(ctx, func(ctx context.Context) (*HttpResponse, error) {
		return c.client.Get(ctx, url, opts...)
	})
}

// Post 发送 POST 请求（带断路器保护）
func (c *CircuitBreakerClient) Post(ctx context.Context, url string, body any, opts ...RequestOption) (*HttpResponse, error) {
	return c.execute(ctx, func(ctx context.Context) (*HttpResponse, error) {
		return c.client.Post(ctx, url, body, opts...)
	})
}

// Put 发送 PUT 请求（带断路器保护）
func (c *CircuitBreakerClient) Put(ctx context.Context, url string, body any, opts ...RequestOption) (*HttpResponse, error) {
	return c.execute(ctx, func(ctx context.Context) (*HttpResponse, error) {
		return c.client.Put(ctx, url, body, opts...)
	})
}

// Delete 发送 DELETE 请求（带断路器保护）
func (c *CircuitBreakerClient) Delete(ctx context.Context, url string, opts ...RequestOption) (*HttpResponse, error) {
	return c.execute(ctx, func(ctx context.Context) (*HttpResponse, error) {
		return c.client.Delete(ctx, url, opts...)
	})
}

// execute 执行带断路器保护的请求
func (c *CircuitBreakerClient) execute(ctx context.Context, fn func(context.Context) (*HttpResponse, error)) (*HttpResponse, error) {
	if !c.breaker.AllowRequest() {
		// 断路器打开，调用降级函数
		if c.fallback != nil {
			return c.fallback(ctx)
		}
		return nil, fmt.Errorf("circuit breaker is open")
	}

	resp, err := fn(ctx)

	if err != nil || (resp != nil && resp.IsServerError()) {
		c.breaker.RecordFailure()
	} else {
		c.breaker.RecordSuccess()
	}

	return resp, err
}

// Close 关闭客户端
func (c *CircuitBreakerClient) Close() error {
	return c.client.Close()
}

// GetCircuitState 获取断路器状态
func (c *CircuitBreakerClient) GetCircuitState() CircuitState {
	return c.breaker.GetState()
}
