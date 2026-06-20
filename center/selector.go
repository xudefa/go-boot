package center

import (
	"math/rand/v2"
	"sync"
	"sync/atomic"
)

// Selector 负载均衡选择器接口。
// 从一组服务实例中按策略选出一个目标实例。
//
// 使用示例：
//
//	sel := center.RandomSelect
//	inst, err := sel.Select(instances)
type Selector interface {
	Select(instances []InstanceInfo) (InstanceInfo, error)
}

var (
	// RandomSelect 随机选择器，从实例列表中随机选取一个。
	// 内部使用独立的 Rand 实例保证并发安全。
	RandomSelect Selector = &randomSelector{rng: rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))}
	// RoundRobinSelect 轮询选择器，按顺序依次选取实例。
	// 内部使用原子计数器保证并发安全。
	RoundRobinSelect Selector = &roundRobinSelector{}
)

// NewWeightedRandomSelect 创建加权随机选择器。
// 权重越高的实例被选中的概率越大。
// 实例的 Weight <= 0 时按 1 处理。
func NewWeightedRandomSelect() Selector {
	return &weightedRandomSelector{rng: rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))}
}

// randomSelector 随机选择器实现。
// 使用 Mutex 保护 Rand 实例，确保并发安全。
type randomSelector struct {
	mu  sync.Mutex
	rng *rand.Rand
}

func (s *randomSelector) Select(instances []InstanceInfo) (InstanceInfo, error) {
	if len(instances) == 0 {
		return InstanceInfo{}, ErrNoInstances
	}
	s.mu.Lock()
	idx := s.rng.IntN(len(instances))
	s.mu.Unlock()
	return instances[idx], nil
}

// roundRobinSelector 轮询选择器实现。
// 使用原子计数器，每次调用时递增并取模得到目标下标。
type roundRobinSelector struct {
	counter atomic.Uint64
}

func (s *roundRobinSelector) Select(instances []InstanceInfo) (InstanceInfo, error) {
	if len(instances) == 0 {
		return InstanceInfo{}, ErrNoInstances
	}
	idx := s.counter.Add(1) - 1
	return instances[idx%uint64(len(instances))], nil
}

// weightedRandomSelector 加权随机选择器实现。
// 算法：计算总权重，生成 [0, total) 之间的随机数，按权重累加找到对应的实例。
type weightedRandomSelector struct {
	mu  sync.Mutex
	rng *rand.Rand
}

func (s *weightedRandomSelector) Select(instances []InstanceInfo) (InstanceInfo, error) {
	n := len(instances)
	if n == 0 {
		return InstanceInfo{}, ErrNoInstances
	}
	// 第一步：计算总权重，权重 <= 0 的按 1 计算
	total := 0
	for _, inst := range instances {
		if inst.Weight <= 0 {
			total++
		} else {
			total += inst.Weight
		}
	}
	// 第二步：生成随机目标值
	s.mu.Lock()
	target := s.rng.IntN(total)
	s.mu.Unlock()
	// 第三步：遍历实例，累减权重直到找到目标实例
	for _, inst := range instances {
		w := inst.Weight
		if w <= 0 {
			w = 1
		}
		target -= w
		if target < 0 {
			return inst, nil
		}
	}
	// 兜底返回最后一个实例（正常情况下不会执行到此）
	return instances[n-1], nil
}
