// Package loadbalancer 提供了多种负载均衡策略实现
//
// 支持轮询、随机、加权、最少连接、一致性哈希、会话保持、IP哈希、
// 健康感知、响应时间加权、自适应权重等负载均衡策略。
//
// 使用示例：
//
//	// 轮询
//	lb := loadbalancer.NewRoundRobin()
//	backend, err := lb.Next(backends)
//
//	// 加权轮询
//	lb := loadbalancer.NewWeightedRoundRobin()
//
//	// 最少连接
//	lb := loadbalancer.NewLeastConnections()
package loadbalancer

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ErrNoBackends 没有可用的后端
	ErrNoBackends = errors.New("no available backends")
)

// ServiceInstance 服务实例
type ServiceInstance struct {
	// ID 服务实例 ID
	ID string
	// URL 服务地址
	URL string
	// Weight 权重
	Weight int
	// Metadata 元数据
	Metadata map[string]string
	// Health 健康状态
	Health HealthStatus
	// Active 活跃连接数
	Active int64
}

// HealthStatus 健康状态
type HealthStatus int

const (
	// HealthUp 健康
	HealthUp HealthStatus = iota
	// HealthDown 不健康
	HealthDown
	// HealthUnknown 未知
	HealthUnknown
)

// Balancer 负载均衡接口
type Balancer interface {
	// Next 选择下一个服务实例
	Next(backends []*ServiceInstance) (*ServiceInstance, error)
}

// RoundRobin 轮询负载均衡器
type RoundRobin struct {
	mu      sync.Mutex
	current int64
}

// NewRoundRobin 创建轮询负载均衡器
func NewRoundRobin() *RoundRobin {
	return &RoundRobin{}
}

// Next 选择下一个后端
func (rr *RoundRobin) Next(backends []*ServiceInstance) (*ServiceInstance, error) {
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	rr.mu.Lock()
	defer rr.mu.Unlock()

	idx := rr.current % int64(len(backends))
	rr.current++
	return backends[idx], nil
}

// Random 随机负载均衡器
type Random struct{}

// NewRandom 创建随机负载均衡器
func NewRandom() *Random {
	return &Random{}
}

// Next 随机选择一个后端
func (r *Random) Next(backends []*ServiceInstance) (*ServiceInstance, error) {
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	idx := rand.Intn(len(backends))
	return backends[idx], nil
}

// WeightedRoundRobin 加权轮询负载均衡器
type WeightedRoundRobin struct {
	mu      sync.Mutex
	current int64
}

// NewWeightedRoundRobin 创建加权轮询负载均衡器
func NewWeightedRoundRobin() *WeightedRoundRobin {
	return &WeightedRoundRobin{}
}

// Next 根据权重选择后端
func (wrr *WeightedRoundRobin) Next(backends []*ServiceInstance) (*ServiceInstance, error) {
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	wrr.mu.Lock()
	defer wrr.mu.Unlock()

	totalWeight := 0
	for _, b := range backends {
		totalWeight += b.Weight
	}

	if totalWeight == 0 {
		return backends[0], nil
	}

	current := wrr.current % int64(totalWeight)
	wrr.current++

	accumulated := int64(0)
	for _, b := range backends {
		accumulated += int64(b.Weight)
		if current < accumulated {
			return b, nil
		}
	}

	return backends[0], nil
}

// LeastConnections 最少连接负载均衡器
type LeastConnections struct{}

// NewLeastConnections 创建最少连接负载均衡器
func NewLeastConnections() *LeastConnections {
	return &LeastConnections{}
}

// Next 选择连接数最少的后端
func (lc *LeastConnections) Next(backends []*ServiceInstance) (*ServiceInstance, error) {
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	var selected *ServiceInstance
	minActive := int64(^uint64(0) >> 1)

	for _, b := range backends {
		active := atomic.LoadInt64(&b.Active)
		if active < minActive {
			minActive = active
			selected = b
		}
	}

	return selected, nil
}

// ConsistentHash 一致性哈希负载均衡器
type ConsistentHash struct {
	mu       sync.Mutex
	replicas int
	hashFunc func(string) uint32
}

// NewConsistentHash 创建一致性哈希负载均衡器
func NewConsistentHash(replicas ...int) *ConsistentHash {
	r := 150
	if len(replicas) > 0 {
		r = replicas[0]
	}

	return &ConsistentHash{
		replicas: r,
		hashFunc: simpleHash,
	}
}

// Next 使用一致性哈希选择后端
func (ch *ConsistentHash) Next(backends []*ServiceInstance) (*ServiceInstance, error) {
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	ch.mu.Lock()
	defer ch.mu.Unlock()

	// 使用默认键进行哈希（基于当前计数器）
	key := fmt.Sprintf("default-%d", ch.replicas)
	hash := ch.hashFunc(key)
	idx := int(hash) % len(backends)
	return backends[idx], nil
}

// NextByKey 根据键选择后端
func (ch *ConsistentHash) NextByKey(backends []*ServiceInstance, key string) (*ServiceInstance, error) {
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	ch.mu.Lock()
	defer ch.mu.Unlock()

	hash := ch.hashFunc(key)
	idx := int(hash) % len(backends)
	return backends[idx], nil
}

// simpleHash 简单的哈希函数
func simpleHash(key string) uint32 {
	var hash uint32
	for _, c := range key {
		hash = hash*31 + uint32(c)
	}
	return hash
}

// ResponseTimeWeighted 响应时间加权负载均衡器
type ResponseTimeWeighted struct {
	mu sync.RWMutex
	// avgResponseTimes 记录每个后端的平均响应时间（毫秒）
	avgResponseTimes map[string]float64
	// lastUpdated 最后更新时间
	lastUpdated map[string]time.Time
	// decay 衰减因子，用于平滑响应时间变化
	decay float64
}

// NewResponseTimeWeighted 创建响应时间加权负载均衡器
func NewResponseTimeWeighted(decay ...float64) *ResponseTimeWeighted {
	d := 0.9 // 默认衰减因子
	if len(decay) > 0 {
		d = decay[0]
	}

	return &ResponseTimeWeighted{
		avgResponseTimes: make(map[string]float64),
		lastUpdated:      make(map[string]time.Time),
		decay:            d,
	}
}

// Next 选择响应时间最短的后端
func (rtw *ResponseTimeWeighted) Next(backends []*ServiceInstance) (*ServiceInstance, error) {
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	rtw.mu.RLock()
	defer rtw.mu.RUnlock()

	type backendWeight struct {
		backend *ServiceInstance
		weight  float64
	}

	weighted := make([]backendWeight, 0, len(backends))
	totalWeight := 0.0

	for _, b := range backends {
		avgTime := rtw.avgResponseTimes[b.URL]
		if avgTime <= 0 {
			avgTime = 1 // 默认值，避免除零
		}

		weight := 1.0 / avgTime
		weighted = append(weighted, backendWeight{backend: b, weight: weight})
		totalWeight += weight
	}

	if totalWeight <= 0 {
		return backends[0], nil
	}

	r := rand.Float64() * totalWeight
	accumulated := 0.0

	for _, bw := range weighted {
		accumulated += bw.weight
		if r <= accumulated {
			return bw.backend, nil
		}
	}

	return backends[len(backends)-1], nil
}

// RecordResponseTime 记录后端响应时间
func (rtw *ResponseTimeWeighted) RecordResponseTime(backendURL string, responseTimeMs float64) {
	rtw.mu.Lock()
	defer rtw.mu.Unlock()

	oldAvg, exists := rtw.avgResponseTimes[backendURL]
	if !exists {
		rtw.avgResponseTimes[backendURL] = responseTimeMs
	} else {
		rtw.avgResponseTimes[backendURL] = rtw.decay*oldAvg + (1-rtw.decay)*responseTimeMs
	}

	rtw.lastUpdated[backendURL] = time.Now()
}

// GetAvgResponseTime 获取后端平均响应时间
func (rtw *ResponseTimeWeighted) GetAvgResponseTime(backendURL string) (float64, bool) {
	rtw.mu.RLock()
	defer rtw.mu.RUnlock()

	avgTime, exists := rtw.avgResponseTimes[backendURL]
	return avgTime, exists
}

// StickySession 会话保持负载均衡器
type StickySession struct {
	mu sync.RWMutex
	// sessionCookieName 会话 Cookie 名称
	sessionCookieName string
	// sessionBackendMap 会话到后端的映射
	sessionBackendMap map[string]*ServiceInstance
}

// NewStickySession 创建会话保持负载均衡器
func NewStickySession(sessionCookieName string) *StickySession {
	if sessionCookieName == "" {
		sessionCookieName = "JSESSIONID"
	}

	return &StickySession{
		sessionCookieName: sessionCookieName,
		sessionBackendMap: make(map[string]*ServiceInstance),
	}
}

// Next 选择后端（不带会话信息）
func (ss *StickySession) Next(backends []*ServiceInstance) (*ServiceInstance, error) {
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	return backends[0], nil
}

// NextWithSession 根据会话 ID 选择后端
func (ss *StickySession) NextWithSession(backends []*ServiceInstance, sessionID string) (*ServiceInstance, error) {
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	if sessionID == "" {
		return backends[0], nil
	}

	ss.mu.RLock()
	backend, exists := ss.sessionBackendMap[sessionID]
	ss.mu.RUnlock()

	if exists {
		for _, b := range backends {
			if b.URL == backend.URL {
				return b, nil
			}
		}
		ss.mu.Lock()
		delete(ss.sessionBackendMap, sessionID)
		ss.mu.Unlock()
	}

	ss.mu.Lock()
	defer ss.mu.Unlock()

	idx := int(simpleHash(sessionID)) % len(backends)
	backend = backends[idx]
	ss.sessionBackendMap[sessionID] = backend

	return backend, nil
}

// GetSessionBackend 获取会话绑定的后端
func (ss *StickySession) GetSessionBackend(sessionID string) (*ServiceInstance, bool) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	backend, exists := ss.sessionBackendMap[sessionID]
	return backend, exists
}

// RemoveSession 移除会话绑定
func (ss *StickySession) RemoveSession(sessionID string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	delete(ss.sessionBackendMap, sessionID)
}

// GetSessionCount 获取会话数量
func (ss *StickySession) GetSessionCount() int {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	return len(ss.sessionBackendMap)
}

// IPHash IP 哈希负载均衡器
type IPHash struct {
	mu sync.Mutex
}

// NewIPHash 创建 IP 哈希负载均衡器
func NewIPHash() *IPHash {
	return &IPHash{}
}

// Next 选择后端（不带 IP 信息）
func (ih *IPHash) Next(backends []*ServiceInstance) (*ServiceInstance, error) {
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	return backends[0], nil
}

// NextByIP 根据客户端 IP 选择后端
func (ih *IPHash) NextByIP(backends []*ServiceInstance, clientIP string) (*ServiceInstance, error) {
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	if clientIP == "" {
		return backends[0], nil
	}

	ih.mu.Lock()
	defer ih.mu.Unlock()

	hash := simpleHash(clientIP)
	idx := int(hash) % len(backends)

	return backends[idx], nil
}

// HealthAware 健康感知负载均衡器
type HealthAware struct {
	mu      sync.RWMutex
	inner   Balancer
	failure map[string]int
}

// NewHealthAware 创建健康感知负载均衡器
func NewHealthAware(inner Balancer) *HealthAware {
	return &HealthAware{
		inner:   inner,
		failure: make(map[string]int),
	}
}

// Next 选择健康的后端
func (ha *HealthAware) Next(backends []*ServiceInstance) (*ServiceInstance, error) {
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	ha.mu.Lock()
	defer ha.mu.Unlock()

	healthyBackends := make([]*ServiceInstance, 0)
	for _, b := range backends {
		if b.Health == HealthUp {
			healthyBackends = append(healthyBackends, b)
		}
	}

	if len(healthyBackends) == 0 {
		healthyBackends = backends
	}

	return ha.inner.Next(healthyBackends)
}

// RecordFailure 记录后端失败
func (ha *HealthAware) RecordFailure(backendURL string) {
	ha.mu.Lock()
	defer ha.mu.Unlock()

	ha.failure[backendURL]++
}

// RecordSuccess 记录后端成功
func (ha *HealthAware) RecordSuccess(backendURL string) {
	ha.mu.Lock()
	defer ha.mu.Unlock()

	delete(ha.failure, backendURL)
}

// GetFailureCount 获取失败计数
func (ha *HealthAware) GetFailureCount(backendURL string) int {
	ha.mu.RLock()
	defer ha.mu.RUnlock()

	return ha.failure[backendURL]
}

// AdaptiveWeight 自适应权重负载均衡器
type AdaptiveWeight struct {
	mu sync.RWMutex
	// weights 后端权重映射
	weights map[string]float64
	// stats 后端统计信息
	stats map[string]*BackendStats
}

// BackendStats 后端统计信息
type BackendStats struct {
	// TotalRequests 总请求数
	TotalRequests int64
	// FailedRequests 失败请求数
	FailedRequests int64
	// AvgResponseTime 平均响应时间（毫秒）
	AvgResponseTime float64
	// ActiveConnections 活跃连接数
	ActiveConnections int64
	// LastUpdated 最后更新时间
	LastUpdated time.Time
}

// NewAdaptiveWeight 创建自适应权重负载均衡器
func NewAdaptiveWeight() *AdaptiveWeight {
	return &AdaptiveWeight{
		weights: make(map[string]float64),
		stats:   make(map[string]*BackendStats),
	}
}

// Next 选择权重最高的后端
func (aw *AdaptiveWeight) Next(backends []*ServiceInstance) (*ServiceInstance, error) {
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	aw.mu.RLock()
	defer aw.mu.RUnlock()

	type backendScore struct {
		backend *ServiceInstance
		score   float64
	}

	scores := make([]backendScore, 0, len(backends))
	totalScore := 0.0

	for _, b := range backends {
		stats := aw.stats[b.URL]
		if stats == nil {
			scores = append(scores, backendScore{backend: b, score: 1.0})
			totalScore += 1.0
			continue
		}

		score := aw.calculateScore(stats)
		scores = append(scores, backendScore{backend: b, score: score})
		totalScore += score
	}

	if totalScore <= 0 {
		return backends[0], nil
	}

	r := rand.Float64() * totalScore
	accumulated := 0.0

	for _, bs := range scores {
		accumulated += bs.score
		if r <= accumulated {
			return bs.backend, nil
		}
	}

	return backends[len(backends)-1], nil
}

// calculateScore 计算后端综合得分
func (aw *AdaptiveWeight) calculateScore(stats *BackendStats) float64 {
	errorRate := 0.0
	if stats.TotalRequests > 0 {
		errorRate = float64(stats.FailedRequests) / float64(stats.TotalRequests)
	}

	responseTimeScore := 1.0
	if stats.AvgResponseTime > 0 {
		responseTimeScore = 1000.0 / (stats.AvgResponseTime + 1)
	}

	connectionScore := 1.0
	if stats.ActiveConnections > 0 {
		connectionScore = 100.0 / (float64(stats.ActiveConnections) + 1)
	}

	score := (1 - errorRate) * responseTimeScore * connectionScore

	return math.Max(score, 0.01)
}

// RecordRequest 记录请求统计
func (aw *AdaptiveWeight) RecordRequest(backendURL string, responseTimeMs float64, failed bool) {
	aw.mu.Lock()
	defer aw.mu.Unlock()

	stats, exists := aw.stats[backendURL]
	if !exists {
		stats = &BackendStats{}
		aw.stats[backendURL] = stats
	}

	stats.TotalRequests++
	if failed {
		stats.FailedRequests++
	}

	if stats.AvgResponseTime == 0 {
		stats.AvgResponseTime = responseTimeMs
	} else {
		stats.AvgResponseTime = 0.9*stats.AvgResponseTime + 0.1*responseTimeMs
	}

	stats.LastUpdated = time.Now()
}

// RecordConnection 记录连接数变化
func (aw *AdaptiveWeight) RecordConnection(backendURL string, delta int64) {
	aw.mu.Lock()
	defer aw.mu.Unlock()

	stats, exists := aw.stats[backendURL]
	if !exists {
		stats = &BackendStats{}
		aw.stats[backendURL] = stats
	}

	stats.ActiveConnections += delta
	if stats.ActiveConnections < 0 {
		stats.ActiveConnections = 0
	}
}

// GetStats 获取后端统计信息
func (aw *AdaptiveWeight) GetStats(backendURL string) (*BackendStats, bool) {
	aw.mu.RLock()
	defer aw.mu.RUnlock()

	stats, exists := aw.stats[backendURL]
	return stats, exists
}

// GetAllStats 获取所有后端统计信息
func (aw *AdaptiveWeight) GetAllStats() map[string]*BackendStats {
	aw.mu.RLock()
	defer aw.mu.RUnlock()

	result := make(map[string]*BackendStats)
	for url, stats := range aw.stats {
		result[url] = stats
	}

	return result
}

// SortByResponseTime 按响应时间排序后端
func SortByResponseTime(backends []*ServiceInstance, stats map[string]*BackendStats) []*ServiceInstance {
	sorted := make([]*ServiceInstance, len(backends))
	copy(sorted, backends)

	sort.Slice(sorted, func(i, j int) bool {
		statsI := stats[sorted[i].URL]
		statsJ := stats[sorted[j].URL]

		if statsI == nil {
			return true
		}
		if statsJ == nil {
			return false
		}

		return statsI.AvgResponseTime < statsJ.AvgResponseTime
	})

	return sorted
}
