// Package actuator 提供应用监控端点，支持健康检查、指标采集和环境变量查看。
package actuator

import (
	"context"
	"time"

	"github.com/xudefa/go-boot/health"
)

// HealthIndicatorBuilder 健康指示器构建器
//
// 使用 Builder 模式简化健康指示器的创建过程。
type HealthIndicatorBuilder struct {
	name      string
	checkFunc func(context.Context) error
	timeout   time.Duration
	details   map[string]any
}

// NewHealthIndicatorBuilder 创建健康指示器构建器
func NewHealthIndicatorBuilder() *HealthIndicatorBuilder {
	return &HealthIndicatorBuilder{
		timeout: 5 * time.Second,
		details: make(map[string]any),
	}
}

// Name 设置指标名称
func (b *HealthIndicatorBuilder) Name(name string) *HealthIndicatorBuilder {
	b.name = name
	return b
}

// CheckFunc 设置检查函数
func (b *HealthIndicatorBuilder) CheckFunc(fn func(context.Context) error) *HealthIndicatorBuilder {
	b.checkFunc = fn
	return b
}

// Timeout 设置超时时间
func (b *HealthIndicatorBuilder) Timeout(timeout time.Duration) *HealthIndicatorBuilder {
	b.timeout = timeout
	return b
}

// Detail 添加详细信息
func (b *HealthIndicatorBuilder) Detail(key string, value any) *HealthIndicatorBuilder {
	b.details[key] = value
	return b
}

// Build 构建健康指示器
func (b *HealthIndicatorBuilder) Build() health.Indicator {
	return &builderIndicator{
		name:      b.name,
		checkFunc: b.checkFunc,
		timeout:   b.timeout,
		details:   b.details,
	}
}

// builderIndicator Builder 创建的健康指示器
type builderIndicator struct {
	name      string
	checkFunc func(context.Context) error
	timeout   time.Duration
	details   map[string]any
}

func (i *builderIndicator) Name() string {
	return i.name
}

func (i *builderIndicator) Health(ctx context.Context) health.Health {
	h := health.Health{
		Details:   make(map[string]any),
		Timestamp: time.Now(),
	}

	// 复制预定义的详细信息
	for k, v := range i.details {
		h.Details[k] = v
	}

	if i.checkFunc == nil {
		h.Status = health.StatusUnknown
		h.Details["error"] = "no check function provided"
		return h
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, i.timeout)
	defer cancel()

	err := i.checkFunc(ctx)
	if err != nil {
		h.Status = health.StatusDown
		h.Details["error"] = err.Error()
		h.Error = err
		return h
	}

	h.Status = health.StatusUp
	h.Details["status"] = "connected"
	h.Details["timestamp"] = time.Now().Format(time.RFC3339)
	return h
}
