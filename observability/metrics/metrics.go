package metrics

import (
	"sync"
	"time"
)

// Metric 指标接口
type Metric interface {
	Name() string
	Value() float64
}

// Counter 计数器
type Counter struct {
	name  string
	value float64
	mu    sync.Mutex
}

// NewCounter 创建计数器
func NewCounter(name string) *Counter {
	return &Counter{name: name}
}

// Name 实现 Metric.Name
func (c *Counter) Name() string {
	return c.name
}

// Value 实现 Metric.Value
func (c *Counter) Value() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

// Inc 增加计数
func (c *Counter) Inc(delta float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value += delta
}

// Reset 重置计数
func (c *Counter) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value = 0
}

// Gauge 仪表盘
type Gauge struct {
	name  string
	value float64
	mu    sync.Mutex
}

// NewGauge 创建仪表盘
func NewGauge(name string) *Gauge {
	return &Gauge{name: name}
}

// Name 实现 Metric.Name
func (g *Gauge) Name() string {
	return g.name
}

// Value 实现 Metric.Value
func (g *Gauge) Value() float64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.value
}

// Set 设置值
func (g *Gauge) Set(value float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value = value
}

// Histogram 直方图
type Histogram struct {
	name  string
	count int64
	sum   float64
	min   float64
	max   float64
	mu    sync.Mutex
}

// NewHistogram 创建直方图
func NewHistogram(name string) *Histogram {
	return &Histogram{
		name: name,
		min:  float64(^uint(0) >> 1),
	}
}

// Name 实现 Metric.Name
func (h *Histogram) Name() string {
	return h.name
}

// Value 实现 Metric.Value (返回平均值)
func (h *Histogram) Value() float64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.count == 0 {
		return 0
	}
	return h.sum / float64(h.count)
}

// Observe 记录观测值
func (h *Histogram) Observe(value float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.count++
	h.sum += value
	if value < h.min {
		h.min = value
	}
	if value > h.max {
		h.max = value
	}
}

// Count 获取计数
func (h *Histogram) Count() int64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.count
}

// Sum 获取总和
func (h *Histogram) Sum() float64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.sum
}

// Min 获取最小值
func (h *Histogram) Min() float64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.min
}

// Max 获取最大值
func (h *Histogram) Max() float64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.max
}

// Timer 计时器
type Timer struct {
	name      string
	startTime time.Time
	duration  time.Duration
	mu        sync.Mutex
}

// NewTimer 创建计时器
func NewTimer(name string) *Timer {
	return &Timer{name: name}
}

// Name 实现 Metric.Name
func (t *Timer) Name() string {
	return t.name
}

// Value 实现 Metric.Value (返回毫秒)
func (t *Timer) Value() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return float64(t.duration.Milliseconds())
}

// Start 开始计时
func (t *Timer) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.startTime = time.Now()
}

// Stop 停止计时
func (t *Timer) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.duration = time.Since(t.startTime)
}

// Duration 获取持续时间
func (t *Timer) Duration() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.duration
}

// MetricsRegistry 指标注册表
type MetricsRegistry struct {
	metrics map[string]Metric
	mu      sync.RWMutex
}

// NewMetricsRegistry 创建指标注册表
func NewMetricsRegistry() *MetricsRegistry {
	return &MetricsRegistry{
		metrics: make(map[string]Metric),
	}
}

// Register 注册指标
func (r *MetricsRegistry) Register(metric Metric) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metrics[metric.Name()] = metric
}

// Get 获取指标
func (r *MetricsRegistry) Get(name string) (Metric, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	metric, exists := r.metrics[name]
	return metric, exists
}

// List 列出所有指标
func (r *MetricsRegistry) List() []Metric {
	r.mu.RLock()
	defer r.mu.RUnlock()
	metrics := make([]Metric, 0, len(r.metrics))
	for _, m := range r.metrics {
		metrics = append(metrics, m)
	}
	return metrics
}
