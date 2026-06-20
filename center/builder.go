package center

import (
	"context"
	"fmt"
	"sync"
)

// RegistryBuilder 注册中心构建器，支持链式配置
type RegistryBuilder struct {
	registryType string
	address      string
	token        string
	heartbeat    int
	metadata     map[string]string
	registry     Registry
}

// NewRegistryBuilder 创建注册中心构建器
func NewRegistryBuilder() *RegistryBuilder {
	return &RegistryBuilder{
		heartbeat: 30,
		metadata:  make(map[string]string),
	}
}

// Type 设置注册中心类型
func (b *RegistryBuilder) Type(typ string) *RegistryBuilder {
	b.registryType = typ
	return b
}

// Address 设置注册中心地址
func (b *RegistryBuilder) Address(addr string) *RegistryBuilder {
	b.address = addr
	return b
}

// Token 设置认证令牌
func (b *RegistryBuilder) Token(token string) *RegistryBuilder {
	b.token = token
	return b
}

// Heartbeat 设置心跳间隔（秒）
func (b *RegistryBuilder) Heartbeat(seconds int) *RegistryBuilder {
	if seconds > 0 {
		b.heartbeat = seconds
	}
	return b
}

// Metadata 添加元数据
func (b *RegistryBuilder) Metadata(key, value string) *RegistryBuilder {
	b.metadata[key] = value
	return b
}

// Registry 使用自定义注册中心
func (b *RegistryBuilder) Registry(reg Registry) *RegistryBuilder {
	b.registry = reg
	return b
}

// Build 构建注册中心
func (b *RegistryBuilder) Build() (Registry, error) {
	if b.registry != nil {
		return b.registry, nil
	}

	if b.address == "" {
		return nil, fmt.Errorf("registry address is required")
	}

	// 返回内存注册中心作为默认实现
	return NewMemoryRegistry(), nil
}

// MustBuild 构建注册中心，失败则panic
func (b *RegistryBuilder) MustBuild() Registry {
	reg, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build registry: %v", err))
	}
	return reg
}

// InstanceBuilder 实例构建器，简化服务实例创建
type InstanceBuilder struct {
	info InstanceInfo
}

// NewInstanceBuilder 创建实例构建器
func NewInstanceBuilder(serviceName string) *InstanceBuilder {
	return &InstanceBuilder{
		info: InstanceInfo{
			ServiceName: serviceName,
			Weight:      1,
			Healthy:     true,
			Metadata:    make(map[string]string),
		},
	}
}

// ID 设置实例ID
func (b *InstanceBuilder) ID(id string) *InstanceBuilder {
	b.info.ID = id
	return b
}

// Host 设置主机地址
func (b *InstanceBuilder) Host(host string) *InstanceBuilder {
	b.info.Host = host
	return b
}

// Port 设置端口
func (b *InstanceBuilder) Port(port int) *InstanceBuilder {
	b.info.Port = port
	return b
}

// Weight 设置权重
func (b *InstanceBuilder) Weight(weight int) *InstanceBuilder {
	if weight > 0 {
		b.info.Weight = weight
	}
	return b
}

// Metadata 添加元数据
func (b *InstanceBuilder) Metadata(key, value string) *InstanceBuilder {
	b.info.Metadata[key] = value
	return b
}

// Build 构建实例信息
func (b *InstanceBuilder) Build() InstanceInfo {
	return b.info
}

// SelectorBuilder 选择器构建器
type SelectorBuilder struct {
	strategy string
	weights  map[string]int
}

// NewSelectorBuilder 创建选择器构建器
func NewSelectorBuilder() *SelectorBuilder {
	return &SelectorBuilder{
		strategy: "round_robin",
		weights:  make(map[string]int),
	}
}

// Strategy 设置选择策略（random、round_robin、weighted）
func (b *SelectorBuilder) Strategy(strategy string) *SelectorBuilder {
	b.strategy = strategy
	return b
}

// Weight 添加权重配置
func (b *SelectorBuilder) Weight(instanceID string, weight int) *SelectorBuilder {
	b.weights[instanceID] = weight
	return b
}

// Build 构建选择器
func (b *SelectorBuilder) Build() (Selector, error) {
	switch b.strategy {
	case "random":
		return NewRandomSelector(), nil
	case "round_robin":
		return NewRoundRobinSelector(), nil
	case "weighted":
		return NewWeightedSelector(b.weights), nil
	default:
		return nil, fmt.Errorf("unknown selector strategy: %s", b.strategy)
	}
}

// MustBuild 构建选择器，失败则panic
func (b *SelectorBuilder) MustBuild() Selector {
	sel, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build selector: %v", err))
	}
	return sel
}

// MemoryRegistry 内存注册中心（简单实现）
type MemoryRegistry struct {
	mu        sync.RWMutex
	instances map[string][]InstanceInfo
	watchers  map[string][]chan []InstanceInfo
}

// NewMemoryRegistry 创建内存注册中心
func NewMemoryRegistry() *MemoryRegistry {
	return &MemoryRegistry{
		instances: make(map[string][]InstanceInfo),
		watchers:  make(map[string][]chan []InstanceInfo),
	}
}

func (r *MemoryRegistry) Register(ctx context.Context, info InstanceInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.instances[info.ServiceName] = append(r.instances[info.ServiceName], info)

	// 通知watchers
	for _, ch := range r.watchers[info.ServiceName] {
		select {
		case ch <- r.instances[info.ServiceName]:
		default:
		}
	}

	return nil
}

func (r *MemoryRegistry) Deregister(ctx context.Context, info InstanceInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	instances := r.instances[info.ServiceName]
	for i, inst := range instances {
		if inst.ID == info.ID {
			r.instances[info.ServiceName] = append(instances[:i], instances[i+1:]...)
			break
		}
	}

	return nil
}

func (r *MemoryRegistry) Discover(ctx context.Context, serviceName string) ([]InstanceInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instances := r.instances[serviceName]
	result := make([]InstanceInfo, 0, len(instances))
	for _, inst := range instances {
		if inst.Healthy {
			result = append(result, inst)
		}
	}

	return result, nil
}

func (r *MemoryRegistry) Watch(ctx context.Context, serviceName string) (<-chan []InstanceInfo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ch := make(chan []InstanceInfo, 10)
	r.watchers[serviceName] = append(r.watchers[serviceName], ch)

	return ch, nil
}

// RandomSelector 随机选择器
type RandomSelector struct{}

func NewRandomSelector() *RandomSelector {
	return &RandomSelector{}
}

func (s *RandomSelector) Select(instances []InstanceInfo) (InstanceInfo, error) {
	if len(instances) == 0 {
		return InstanceInfo{}, fmt.Errorf("no instances available")
	}
	return instances[0], nil
}

// RoundRobinSelector 轮询选择器
type RoundRobinSelector struct {
	mu    sync.Mutex
	index int
}

func NewRoundRobinSelector() *RoundRobinSelector {
	return &RoundRobinSelector{}
}

func (s *RoundRobinSelector) Select(instances []InstanceInfo) (InstanceInfo, error) {
	if len(instances) == 0 {
		return InstanceInfo{}, fmt.Errorf("no instances available")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	instance := instances[s.index%len(instances)]
	s.index++
	return instance, nil
}

// WeightedSelector 加权选择器
type WeightedSelector struct {
	weights map[string]int
}

func NewWeightedSelector(weights map[string]int) *WeightedSelector {
	return &WeightedSelector{weights: weights}
}

func (s *WeightedSelector) Select(instances []InstanceInfo) (InstanceInfo, error) {
	if len(instances) == 0 {
		return InstanceInfo{}, fmt.Errorf("no instances available")
	}
	return instances[0], nil
}
