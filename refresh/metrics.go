package refresh

import (
	"sync/atomic"
	"time"
)

// RefreshMetrics 刷新指标收集器
//
// 使用原子操作记录刷新次数、耗时等指标，支持并发安全访问。
type RefreshMetrics struct {
	totalRefreshes      atomic.Int64 // 总刷新次数
	successfulRefreshes atomic.Int64 // 成功刷新次数
	failedRefreshes     atomic.Int64 // 失败刷新次数
	totalRefreshTime    atomic.Int64 // 总刷新耗时（纳秒）
	lastRefreshTime     atomic.Int64 // 最后刷新时间（Unix 纳秒）
}

// NewRefreshMetrics 创建刷新指标收集器
func NewRefreshMetrics() *RefreshMetrics {
	return &RefreshMetrics{}
}

// RecordRefresh 记录一次刷新操作
//
// 参数：
//   - duration: 刷新耗时
//   - success: 是否刷新成功
func (m *RefreshMetrics) RecordRefresh(duration time.Duration, success bool) {
	m.totalRefreshes.Add(1)
	m.totalRefreshTime.Add(int64(duration))
	m.lastRefreshTime.Store(time.Now().UnixNano())

	if success {
		m.successfulRefreshes.Add(1)
	} else {
		m.failedRefreshes.Add(1)
	}
}

// TotalRefreshes 返回总刷新次数
func (m *RefreshMetrics) TotalRefreshes() int64 {
	return m.totalRefreshes.Load()
}

// SuccessfulRefreshes 返回成功刷新次数
func (m *RefreshMetrics) SuccessfulRefreshes() int64 {
	return m.successfulRefreshes.Load()
}

// FailedRefreshes 返回失败刷新次数
func (m *RefreshMetrics) FailedRefreshes() int64 {
	return m.failedRefreshes.Load()
}

// AverageRefreshTime 返回平均刷新耗时
func (m *RefreshMetrics) AverageRefreshTime() time.Duration {
	total := m.totalRefreshes.Load()
	if total == 0 {
		return 0
	}
	return time.Duration(m.totalRefreshTime.Load() / total)
}

// LastRefreshTime 返回最后刷新时间
func (m *RefreshMetrics) LastRefreshTime() time.Time {
	nanos := m.lastRefreshTime.Load()
	if nanos == 0 {
		return time.Time{}
	}
	return time.Unix(0, nanos)
}

// Reset 重置所有指标
func (m *RefreshMetrics) Reset() {
	m.totalRefreshes.Store(0)
	m.successfulRefreshes.Store(0)
	m.failedRefreshes.Store(0)
	m.totalRefreshTime.Store(0)
	m.lastRefreshTime.Store(0)
}
