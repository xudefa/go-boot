package security

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// TokenBucket 令牌桶
type TokenBucket struct {
	capacity int
	tokens   int
	rate     int
	mu       sync.Mutex
	lastTime time.Time
}

// NewTokenBucket 创建令牌桶
func NewTokenBucket(capacity, rate int) *TokenBucket {
	return &TokenBucket{
		capacity: capacity,
		tokens:   capacity,
		rate:     rate,
		lastTime: time.Now(),
	}
}

// Take 尝试获取令牌
func (b *TokenBucket) Take() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastTime).Seconds()
	newTokens := int(elapsed * float64(b.rate))

	if newTokens > 0 {
		b.tokens = min(b.capacity, b.tokens+newTokens)
		b.lastTime = now
	}

	if b.tokens > 0 {
		b.tokens--
		return true
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled      bool
	Rate         int
	Burst        int
	ExcludePaths []string
}

// RateLimitFilter 限流过滤器
type RateLimitFilter struct {
	config       RateLimitConfig
	buckets      sync.Map
	globalBucket *TokenBucket
}

// NewRateLimitFilter 创建限流过滤器
func NewRateLimitFilter(config RateLimitConfig) *RateLimitFilter {
	if config.Rate == 0 {
		config.Rate = 100
	}
	if config.Burst == 0 {
		config.Burst = 200
	}
	return &RateLimitFilter{
		config:       config,
		globalBucket: NewTokenBucket(config.Burst, config.Rate),
	}
}

// DoFilter 处理限流
func (f *RateLimitFilter) DoFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain SecurityFilterChain) error {
	if !f.config.Enabled {
		return chain.DoFilter(ctx, request, response)
	}

	uri := request.GetURI()
	for _, path := range f.config.ExcludePaths {
		if strings.HasPrefix(uri, path) {
			return chain.DoFilter(ctx, request, response)
		}
	}

	clientIP := f.getClientIP(request)

	if clientIP != "" {
		bucket, _ := f.buckets.LoadOrStore(clientIP, NewTokenBucket(f.config.Burst, f.config.Rate))
		if !bucket.(*TokenBucket).Take() {
			response.SetStatusCode(429)
			if writeErr := response.Write([]byte(`{"error":"rate limited","message":"too many requests"}`)); writeErr != nil {
				fmt.Printf("[go-boot] failed to write rate limit response: %v\n", writeErr)
			}
			return nil
		}
	} else {
		if !f.globalBucket.Take() {
			response.SetStatusCode(429)
			if writeErr := response.Write([]byte(`{"error":"rate limited","message":"too many requests"}`)); writeErr != nil {
				fmt.Printf("[go-boot] failed to write rate limit response: %v\n", writeErr)
			}
			return nil
		}
	}

	return chain.DoFilter(ctx, request, response)
}

func (f *RateLimitFilter) getClientIP(request SecurityRequest) string {
	headers := []string{"X-Forwarded-For", "X-Real-IP", "Proxy-Client-IP", "WL-Proxy-Client-IP"}
	for _, header := range headers {
		if ip := request.GetHeader(header); ip != "" {
			parts := strings.Split(ip, ",")
			clientIP := strings.TrimSpace(parts[0])
			if net.ParseIP(clientIP) != nil {
				return clientIP
			}
		}
	}
	return ""
}
