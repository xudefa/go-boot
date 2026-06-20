package health

import (
	"context"
	"time"
)

// IndicatorBuilder 健康指标构建器
//
// 使用 Builder 模式简化健康指标创建流程，支持链式配置。
// 内置超时控制，防止健康检查阻塞。
//
// 使用示例：
//
//	indicator := NewIndicatorBuilder().
//	    Name("database").
//	    CheckFunc(func(ctx context.Context) error {
//	        return db.PingContext(ctx)
//	    }).
//	    Timeout(5 * time.Second).
//	    Detail("type", "postgres").
//	    Build()
type IndicatorBuilder struct {
	name      string
	checkFunc func(context.Context) error
	timeout   time.Duration
	details   map[string]any
}

// NewIndicatorBuilder 创建健康指标构建器
//
// 返回:
//   - *IndicatorBuilder: 构建器实例，默认超时 5 秒
func NewIndicatorBuilder() *IndicatorBuilder {
	return &IndicatorBuilder{
		timeout: 5 * time.Second,
		details: make(map[string]any),
	}
}

// Name 设置指标名称
//
// 参数:
//   - name: 指标名称
//
// 返回:
//   - *IndicatorBuilder: 构建器自身，支持链式调用
func (b *IndicatorBuilder) Name(name string) *IndicatorBuilder {
	b.name = name
	return b
}

// CheckFunc 设置检查函数
//
// 参数:
//   - fn: 检查函数，接收上下文参数，返回错误表示健康检查失败
//
// 返回:
//   - *IndicatorBuilder: 构建器自身，支持链式调用
func (b *IndicatorBuilder) CheckFunc(fn func(context.Context) error) *IndicatorBuilder {
	b.checkFunc = fn
	return b
}

// Timeout 设置超时时间
//
// 参数:
//   - d: 超时时间
//
// 返回:
//   - *IndicatorBuilder: 构建器自身，支持链式调用
func (b *IndicatorBuilder) Timeout(d time.Duration) *IndicatorBuilder {
	b.timeout = d
	return b
}

// Detail 添加详细信息
//
// 参数:
//   - key: 信息键名
//   - value: 信息值
//
// 返回:
//   - *IndicatorBuilder: 构建器自身，支持链式调用
func (b *IndicatorBuilder) Detail(key string, value any) *IndicatorBuilder {
	b.details[key] = value
	return b
}

// Build 构建健康指标
//
// 返回:
//   - Indicator: 健康指标实例
func (b *IndicatorBuilder) Build() Indicator {
	return &builderIndicator{
		name:      b.name,
		checkFunc: b.checkFunc,
		timeout:   b.timeout,
		details:   b.details,
	}
}

// builderIndicator 构建器创建的健康指标实现
type builderIndicator struct {
	name      string
	checkFunc func(context.Context) error
	timeout   time.Duration
	details   map[string]any
}

// Name 返回指标名称
func (i *builderIndicator) Name() string {
	return i.name
}

// Health 执行健康检查
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - Health: 健康状态
func (i *builderIndicator) Health(ctx context.Context) Health {
	// 检查函数为空时返回未知状态
	if i.checkFunc == nil {
		return Health{
			Status:    StatusUnknown,
			Details:   i.details,
			Timestamp: time.Now(),
		}
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, i.timeout)
	defer cancel()

	// 执行检查
	err := i.checkFunc(ctx)
	if err != nil {
		return Health{
			Status:    StatusDown,
			Details:   i.details,
			Error:     err,
			Timestamp: time.Now(),
		}
	}

	return Health{
		Status:    StatusUp,
		Details:   i.details,
		Timestamp: time.Now(),
	}
}
