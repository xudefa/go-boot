// Package metrics 提供轻量级的指标收集接口和默认实现。
//
// 核心接口：
//   - Counter: 计数器，只增不减（如请求总数、错误次数）
//   - Gauge: 仪表盘，可增可减（如当前连接数、内存使用量）
//   - Histogram: 直方图，记录分布情况（如请求延迟）
//   - MeterRegistry: 指标注册表，管理指标创建和收集
//   - Exporter: 指标导出器，支持将指标发送到监控系统
//
// 使用示例：
//
//	registry := metrics.NewSimpleRegistry()
//	counter := registry.Counter("http_requests_total")
//	counter.Inc()
//	gauge := registry.Gauge("memory_usage")
//	gauge.Set(1024.5)
//	histogram := registry.Histogram("request_duration")
//	histogram.Record(123.45)
//	all := registry.Collect()
package metrics

import (
	"math"
	"strings"
	"sync"
	"time"
)

// Counter 计数器接口
//
// 用于记录单调递增的数值，如请求次数、错误计数。
type Counter interface {
	// Inc 计数器加 1
	Inc()
	// Add 计数器增加指定值
	Add(v float64)
	// Value 返回当前计数值
	Value() float64
	// Reset 重置计数器为 0
	Reset()
}

// Gauge 仪表盘接口
//
// 用于记录可增可减的数值，如当前连接数、CPU 使用率。
type Gauge interface {
	// Set 设置当前值
	Set(v float64)
	// Add 增加指定值（可以为负数）
	Add(v float64)
	// Value 返回当前值
	Value() float64
}

// Histogram 直方图接口
//
// 用于记录分布情况，如请求延迟、响应大小等。
type Histogram interface {
	// Record 记录一个值
	Record(v float64)
	// RecordWithLabels 记录带标签的值
	RecordWithLabels(v float64, labels map[string]string)
	// Count 返回记录的样本数
	Count() int64
	// Sum 返回所有样本的总和
	Sum() float64
	// Reset 重置直方图
	Reset()
}

// Metric 指标快照
//
// 包含指标名称、当前值和标签信息，用于采集和上报。
type Metric struct {
	Name      string            `json:"name"`      // 指标名称
	Value     float64           `json:"value"`     // 指标当前值
	Tags      map[string]string `json:"tags"`      // 指标标签
	Type      string            `json:"type"`      // 指标类型: counter/gauge/histogram
	Timestamp int64             `json:"timestamp"` // 时间戳
	Count     int64             `json:"count"`     // 样本数量（仅直方图）
	Sum       float64           `json:"sum"`       // 样本总和（仅直方图）
}

// Exporter 指标导出器接口
type Exporter interface {
	Export(metrics []Metric) error
}

// MeterRegistry 指标注册表接口
//
// 管理 Counter、Gauge 和 Histogram 的创建与收集，支持按名称获取或创建。
// 提供指标导出功能，可将指标数据导出到不同的监控系统。
type MeterRegistry interface {
	// Counter 获取或创建指定名称的计数器
	// name: 指标名称
	// tags: 标签对，格式为 key1, value1, key2, value2...
	Counter(name string, tags ...string) Counter

	// Gauge 获取或创建指定名称的仪表盘
	// name: 指标名称
	// tags: 标签对，格式为 key1, value1, key2, value2...
	Gauge(name string, tags ...string) Gauge

	// Histogram 获取或创建指定名称的直方图
	// name: 指标名称
	// tags: 标签对，格式为 key1, value1, key2, value2...
	Histogram(name string, tags ...string) Histogram

	// Collect 收集所有已注册的指标快照
	Collect() []Metric

	// RegisterExporter 注册指标导出器
	RegisterExporter(exporter Exporter)

	// Export 导出所有指标到已注册的导出器
	Export() error

	// Reset 重置所有指标为初始状态
	Reset()
}

// simpleCounter 简单计数器实现
//
// 使用 sync.RWMutex 保证并发安全，支持原子增减操作。
// 适用于记录单调递增的数值，如请求次数、错误计数等。
type simpleCounter struct {
	mu    sync.RWMutex // 读写锁，保护 value 的读写
	value float64      // 当前计数值
}

// NewSimpleCounter 创建新的简单计数器
func NewSimpleCounter() Counter {
	return &simpleCounter{}
}

// Inc 计数器加 1
func (c *simpleCounter) Inc() {
	c.Add(1)
}

// Add 计数器增加指定值
func (c *simpleCounter) Add(v float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value += v
}

// Value 返回当前计数值
func (c *simpleCounter) Value() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.value
}

// Reset 重置计数器为 0
func (c *simpleCounter) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value = 0
}

// simpleGauge 简单仪表盘实现
//
// 使用 sync.RWMutex 保证并发安全，支持并发读取。
type simpleGauge struct {
	mu    sync.RWMutex // 读写锁，保护 value 的并发访问
	value float64      // 当前值
}

// NewSimpleGauge 创建新的简单仪表盘
func NewSimpleGauge() Gauge {
	return &simpleGauge{}
}

// Set 设置当前值
func (g *simpleGauge) Set(v float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value = v
}

// Add 增加指定值（可以为负数）
func (g *simpleGauge) Add(v float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value += v
}

// Value 返回当前值
func (g *simpleGauge) Value() float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.value
}

// simpleHistogram 简单直方图实现
//
// 使用 sync.Mutex 保证并发安全，支持基本的统计功能。
type simpleHistogram struct {
	mu    sync.Mutex
	name  string
	tags  map[string]string
	count int64
	sum   float64
	min   float64
	max   float64
}

// NewSimpleHistogram 创建新的简单直方图
func NewSimpleHistogram(name string, tags map[string]string) Histogram {
	return &simpleHistogram{
		name: name,
		tags: tags,
		min:  math.MaxFloat64,
		max:  math.Inf(-1),
	}
}

// Record 记录一个值
func (h *simpleHistogram) Record(v float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.count++
	h.sum += v
	if v < h.min {
		h.min = v
	}
	if v > h.max {
		h.max = v
	}
}

// RecordWithLabels 记录带标签的值
func (h *simpleHistogram) RecordWithLabels(v float64, labels map[string]string) {
	if len(labels) > 0 {
		h.mu.Lock()
		for k, v := range labels {
			h.tags[k] = v
		}
		h.mu.Unlock()
	}
	h.Record(v)
}

// Count 返回记录的样本数
func (h *simpleHistogram) Count() int64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.count
}

// Sum 返回所有样本的总和
func (h *simpleHistogram) Sum() float64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.sum
}

// Reset 重置直方图
func (h *simpleHistogram) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.count = 0
	h.sum = 0
	h.min = math.MaxFloat64
	h.max = math.Inf(-1)
}

// simpleRegistry 简单指标注册表实现
//
// 使用 map 存储计数器、仪表盘和直方图，sync.Mutex 保证并发安全。
// 使用 name+tags 组合作为唯一键，支持同名不同标签的指标实例。
type simpleRegistry struct {
	mu         sync.Mutex
	counters   map[string]*simpleCounter
	gauges     map[string]*simpleGauge
	histograms map[string]*simpleHistogram
	tags       map[string]map[string]string
	exporters  []Exporter
}

// NewSimpleRegistry 创建新的简单指标注册表
func NewSimpleRegistry() MeterRegistry {
	return &simpleRegistry{
		counters:   make(map[string]*simpleCounter),
		gauges:     make(map[string]*simpleGauge),
		histograms: make(map[string]*simpleHistogram),
		tags:       make(map[string]map[string]string),
		exporters:  make([]Exporter, 0),
	}
}

func parseTags(tags []string) map[string]string {
	if len(tags) == 0 {
		return nil
	}
	result := make(map[string]string, len(tags)/2)
	for i := 0; i+1 < len(tags); i += 2 {
		result[tags[i]] = tags[i+1]
	}
	return result
}

// metricKey 生成指标唯一键，由名称和标签组成
func metricKey(name string, tags map[string]string) string {
	if len(tags) == 0 {
		return name
	}
	key := name
	for k, v := range tags {
		key += "|" + k + "=" + v
	}
	return key
}

// Counter 获取或创建指定名称的计数器
//
// 使用 name+tags 组合作为唯一键，同名不同标签会创建不同的计数器实例。
func (r *simpleRegistry) Counter(name string, tags ...string) Counter {
	r.mu.Lock()
	defer r.mu.Unlock()
	parsedTags := parseTags(tags)
	key := metricKey(name, parsedTags)
	if c, ok := r.counters[key]; ok {
		return c
	}
	c := &simpleCounter{}
	r.counters[key] = c
	if len(tags) > 0 {
		r.tags[key] = parsedTags
	}
	return c
}

// Gauge 获取或创建指定名称的仪表盘
//
// 使用 name+tags 组合作为唯一键，同名不同标签会创建不同的仪表盘实例。
func (r *simpleRegistry) Gauge(name string, tags ...string) Gauge {
	r.mu.Lock()
	defer r.mu.Unlock()
	parsedTags := parseTags(tags)
	key := metricKey(name, parsedTags)
	if g, ok := r.gauges[key]; ok {
		return g
	}
	g := &simpleGauge{}
	r.gauges[key] = g
	if len(tags) > 0 {
		r.tags[key] = parsedTags
	}
	return g
}

// Histogram 获取或创建指定名称的直方图
//
// 使用 name+tags 组合作为唯一键，同名不同标签会创建不同的直方图实例。
func (r *simpleRegistry) Histogram(name string, tags ...string) Histogram {
	r.mu.Lock()
	defer r.mu.Unlock()
	parsedTags := parseTags(tags)
	key := metricKey(name, parsedTags)
	if h, ok := r.histograms[key]; ok {
		return h
	}
	h := NewSimpleHistogram(name, parsedTags)
	r.histograms[key] = h.(*simpleHistogram)
	if len(tags) > 0 {
		r.tags[key] = parsedTags
	}
	return h
}

// metricNameFromKey 从 metricKey 中提取纯指标名称
func metricNameFromKey(key string) string {
	if idx := strings.Index(key, "|"); idx >= 0 {
		return key[:idx]
	}
	return key
}

// Collect 收集所有已注册的指标快照
//
// 返回当前所有计数器、仪表盘和直方图的快照列表。
func (r *simpleRegistry) Collect() []Metric {
	r.mu.Lock()
	defer r.mu.Unlock()

	metrics := make([]Metric, 0, len(r.counters)+len(r.gauges)+len(r.histograms))
	now := time.Now().UnixMilli()

	for key, c := range r.counters {
		m := Metric{
			Name:      metricNameFromKey(key),
			Value:     c.Value(),
			Type:      "counter",
			Timestamp: now,
		}
		if tags, ok := r.tags[key]; ok {
			m.Tags = tags
		}
		metrics = append(metrics, m)
	}

	for key, g := range r.gauges {
		m := Metric{
			Name:      metricNameFromKey(key),
			Value:     g.Value(),
			Type:      "gauge",
			Timestamp: now,
		}
		if tags, ok := r.tags[key]; ok {
			m.Tags = tags
		}
		metrics = append(metrics, m)
	}

	for key, h := range r.histograms {
		count := h.Count()
		var avg float64
		if count > 0 {
			avg = h.Sum() / float64(count)
		}
		m := Metric{
			Name:      metricNameFromKey(key),
			Value:     avg,
			Type:      "histogram",
			Timestamp: now,
			Count:     count,
			Sum:       h.Sum(),
		}
		if tags, ok := r.tags[key]; ok {
			m.Tags = tags
		}
		metrics = append(metrics, m)
	}

	return metrics
}

// RegisterExporter 注册指标导出器
func (r *simpleRegistry) RegisterExporter(exporter Exporter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.exporters = append(r.exporters, exporter)
}

// Export 导出所有指标到已注册的导出器
func (r *simpleRegistry) Export() error {
	metrics := r.Collect()
	for _, exporter := range r.exporters {
		if err := exporter.Export(metrics); err != nil {
			return err
		}
	}
	return nil
}

// Reset 重置所有指标
func (r *simpleRegistry) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range r.counters {
		c.Reset()
	}
	for _, h := range r.histograms {
		h.Reset()
	}
}

// ConsoleExporter 控制台导出器
type ConsoleExporter struct{}

func NewConsoleExporter() Exporter {
	return &ConsoleExporter{}
}

func (e *ConsoleExporter) Export(metrics []Metric) error {
	return nil
}
