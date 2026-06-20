// Package health 提供健康检查的核心接口和聚合器。
//
// Indicator 接口由各集成模块实现（如数据库、Redis 连接检查），
// Aggregator 聚合所有指标的健康状态，支持组合判断。
// 健康状态枚举：StatusUp > StatusDegraded > StatusDown > StatusOutage > StatusUnknown。
//
// 使用示例：
//
//	agg := health.NewAggregator()
//	agg.AddIndicator(&myIndicator{})
//	h := agg.Aggregate(ctx)
//	fmt.Println(h.Status) // UP / DOWN / DEGRADED
package health

import (
	"context"
	"sync"
	"time"
)

// Status 健康状态枚举
//
// 表示组件或应用的整体健康程度，从优到劣依次为：
// StatusUp > StatusDegraded > StatusDown > StatusOutage > StatusUnknown。
//
// 使用示例：
//
//	status := health.StatusUp
//	fmt.Println(status) // 输出: UP
type Status int

const (
	// StatusUp 表示组件运行正常，服务完全可用
	// 应用于：数据库连接正常、缓存服务可用、外部API正常响应等
	StatusUp Status = iota // 正常：组件运行正常

	// StatusDown 表示组件不可用，服务中断
	// 应用于：数据库连接失败、关键服务不可达等
	StatusDown // 异常：组件不可用

	// StatusDegraded 表示组件部分功能可用，服务质量下降
	// 应用于：响应时间变慢、部分功能受限等
	StatusDegraded // 降级：组件部分功能可用

	// StatusOutage 表示组件完全不可用，服务停服
	// 应用于：严重故障导致服务完全不可用
	StatusOutage // 停服：组件完全不可用

	// StatusUnknown 表示无法确定组件状态
	// 应用于：检查函数为空、超时等不确定状态
	StatusUnknown // 未知：无法确定组件状态
)

var statusNames = map[Status]string{
	StatusUp:       "UP",
	StatusDown:     "DOWN",
	StatusDegraded: "DEGRADED",
	StatusOutage:   "OUTAGE",
	StatusUnknown:  "UNKNOWN",
}

func (s Status) String() string {
	if name, ok := statusNames[s]; ok {
		return name
	}
	return "UNKNOWN"
}

// Health 健康信息
//
// 包含组件的健康状态、详细信息、错误信息和时间戳。
// 用于在健康检查聚合器中传递和展示健康状态。
type Health struct {
	// Status 组件的健康状态
	Status Status `json:"status"`

	// Details 组件健康检查的详细信息，包含各个子组件的状态
	Details map[string]any `json:"details,omitempty"`

	// Error 检查过程中发生的错误信息
	Error error `json:"-"`

	// Timestamp 健康检查完成的时间戳
	Timestamp time.Time `json:"timestamp"`
}

// Indicator 健康指标接口
//
// 各集成模块实现此接口提供组件健康状态。
// 例如：数据库连接检查、Redis 连通性检查。
type Indicator interface {
	Name() string
	Health(ctx context.Context) Health
}

// Aggregator 健康指标聚合器
//
// 聚合所有 Indicator 的健康状态。
// 全部 UP 则整体 UP，任一 DOWN 则整体 DOWN。
// 线程安全，可在并发环境下使用。
type Aggregator struct {
	mu         sync.RWMutex // 读写锁，保护并发访问
	indicators []Indicator  // 健康指标切片
}

// NewAggregator 创建聚合器
func NewAggregator() *Aggregator {
	return &Aggregator{}
}

// AddIndicator 添加健康指标
func (a *Aggregator) AddIndicator(indicator Indicator) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.indicators = append(a.indicators, indicator)
}

// Indicators 返回所有指标的副本，防止外部修改
func (a *Aggregator) Indicators() []Indicator {
	a.mu.RLock()
	defer a.mu.RUnlock()
	result := make([]Indicator, len(a.indicators))
	copy(result, a.indicators)
	return result
}

// Aggregate 聚合所有指标的健康状态
func (a *Aggregator) Aggregate(ctx context.Context) Health {
	indicators := a.Indicators()

	overall := StatusUp
	details := make(map[string]any)

	for _, ind := range indicators {
		h := ind.Health(ctx)
		d := map[string]any{
			"status": h.Status.String(),
			"detail": h.Details,
		}
		if h.Error != nil {
			d["error"] = h.Error.Error()
		}
		details[ind.Name()] = d
		switch h.Status {
		case StatusOutage:
			overall = StatusOutage
		case StatusDown:
			if overall != StatusOutage {
				overall = StatusDown
			}
		case StatusDegraded:
			if overall != StatusOutage && overall != StatusDown {
				overall = StatusDegraded
			}
		}
	}

	return Health{
		Status:    overall,
		Details:   details,
		Timestamp: time.Now(),
	}
}
