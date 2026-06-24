// Package circuit 提供了熔断器实现，用于防止级联故障
//
// 熔断器用于在后端服务不可用时快速失败，防止级联故障。
// 熔断器有三种状态：Closed（关闭）、Open（打开）、HalfOpen（半开）。
//
// 状态转换：
//   - Closed → Open: 失败率超过阈值
//   - Open → HalfOpen: 等待时间过后
//   - HalfOpen → Closed: 请求成功
//   - HalfOpen → Open: 请求失败
//
// 使用示例：
//
//	breaker := circuit.NewDefaultBreaker(
//	    circuit.WithErrorThreshold(0.5),
//	    circuit.WithWaitDuration(30 * time.Second),
//	)
//	if err := breaker.Allow(); err != nil {
//	    return errors.New("circuit breaker is open")
//	}
//	defer func() {
//	    if err != nil {
//	        breaker.RecordFailure()
//	    } else {
//	        breaker.RecordSuccess()
//	    }
//	}()
package circuit

import (
	"errors"
	"sync"
	"time"
)

var (
	// ErrCircuitOpen 熔断器打开错误
	ErrCircuitOpen = errors.New("circuit breaker is open")
	// ErrCircuitHalfOpen 熔断器半开错误
	ErrCircuitHalfOpen = errors.New("circuit breaker is half-open")
)

// State 熔断器状态
type State int

const (
	// StateClosed 关闭状态，正常处理请求
	StateClosed State = iota
	// StateOpen 打开状态，快速失败
	StateOpen
	// StateHalfOpen 半开状态，尝试恢复
	StateHalfOpen
)

// String 返回状态的字符串表示
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// Breaker 熔断器接口
type Breaker interface {
	// Allow 检查是否允许请求通过
	Allow() error
	// RecordSuccess 记录成功请求
	RecordSuccess()
	// RecordFailure 记录失败请求
	RecordFailure()
	// State 获取当前状态
	State() State
}

// DefaultBreaker 默认熔断器实现
type DefaultBreaker struct {
	mu sync.Mutex

	// 配置参数
	maxRequests    int           // 半开状态下允许的最大请求数
	errorThreshold float64       // 错误率阈值
	waitDuration   time.Duration // 打开状态等待时间

	// 状态
	state            State
	errors           int
	successes        int
	requests         int
	lastStateChange  time.Time
	halfOpenRequests int
}

// BreakerOption 熔断器选项
type BreakerOption func(*DefaultBreaker)

// WithMaxRequests 设置半开状态最大请求数
func WithMaxRequests(max int) BreakerOption {
	return func(b *DefaultBreaker) {
		b.maxRequests = max
	}
}

// WithErrorThreshold 设置错误率阈值
func WithErrorThreshold(threshold float64) BreakerOption {
	return func(b *DefaultBreaker) {
		b.errorThreshold = threshold
	}
}

// WithWaitDuration 设置打开状态等待时间
func WithWaitDuration(duration time.Duration) BreakerOption {
	return func(b *DefaultBreaker) {
		b.waitDuration = duration
	}
}

// NewDefaultBreaker 创建默认熔断器
func NewDefaultBreaker(opts ...BreakerOption) *DefaultBreaker {
	b := &DefaultBreaker{
		maxRequests:    10,
		errorThreshold: 0.5,
		waitDuration:   30 * time.Second,
		state:          StateClosed,
	}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

// Allow 检查是否允许请求通过
func (b *DefaultBreaker) Allow() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateClosed:
		return nil
	case StateOpen:
		if time.Since(b.lastStateChange) > b.waitDuration {
			b.state = StateHalfOpen
			b.halfOpenRequests = 0
			b.lastStateChange = time.Now()
			return nil
		}
		return ErrCircuitOpen
	case StateHalfOpen:
		if b.halfOpenRequests < b.maxRequests {
			b.halfOpenRequests++
			return nil
		}
		return ErrCircuitHalfOpen
	}

	return ErrCircuitOpen
}

// RecordSuccess 记录成功请求
func (b *DefaultBreaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.successes++
	b.requests++

	switch b.state {
	case StateHalfOpen:
		if b.successes >= b.maxRequests {
			b.state = StateClosed
			b.errors = 0
			b.successes = 0
			b.requests = 0
			b.lastStateChange = time.Now()
		}
	case StateClosed:
		// 成功请求不需要检查阈值，仅重置计数器防止误判
		// 使用更大的窗口避免高流量场景下频繁重置
		if b.requests >= 1000 {
			// 窗口重置，防止长期运行后阈值失效
			b.errors = 0
			b.requests = 0
			b.successes = 0
		}
	}
}

// RecordFailure 记录失败请求
func (b *DefaultBreaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.errors++
	b.requests++

	switch b.state {
	case StateHalfOpen:
		b.state = StateOpen
		b.lastStateChange = time.Now()
	case StateClosed:
		b.checkThreshold()
	}
}

// State 获取当前状态
func (b *DefaultBreaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

// FailureCount 获取失败计数
func (b *DefaultBreaker) FailureCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.errors
}

// SuccessCount 获取成功计数
func (b *DefaultBreaker) SuccessCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.successes
}

// Threshold 获取错误率阈值
func (b *DefaultBreaker) Threshold() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	// 返回基于 maxRequests 的阈值
	return b.maxRequests
}

// RecoveryTimeout 获取恢复超时
func (b *DefaultBreaker) RecoveryTimeout() time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.waitDuration
}

// LastStateChange 获取最后状态变化时间
func (b *DefaultBreaker) LastStateChange() time.Time {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.lastStateChange
}

// checkThreshold 检查是否达到错误率阈值
func (b *DefaultBreaker) checkThreshold() {
	if b.requests == 0 {
		return
	}

	errorRate := float64(b.errors) / float64(b.requests)
	if errorRate >= b.errorThreshold {
		b.state = StateOpen
		b.lastStateChange = time.Now()
	}
}
