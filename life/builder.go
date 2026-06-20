package life

import (
	"fmt"
	"sync"
)

// LifecycleBuilder 生命周期构建器，支持链式配置
type LifecycleBuilder struct {
	initialPhase ApplicationPhase
	listeners    []PhaseListener
	onError      func(oldPhase, newPhase ApplicationPhase, err error)
	mu           sync.Mutex
}

// NewLifecycleBuilder 创建生命周期构建器
func NewLifecycleBuilder() *LifecycleBuilder {
	return &LifecycleBuilder{
		initialPhase: PhaseInitializing,
		listeners:    make([]PhaseListener, 0),
	}
}

// InitialPhase 设置初始阶段
func (b *LifecycleBuilder) InitialPhase(phase ApplicationPhase) *LifecycleBuilder {
	b.initialPhase = phase
	return b
}

// Listener 添加监听器
func (b *LifecycleBuilder) Listener(listener PhaseListener) *LifecycleBuilder {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.listeners = append(b.listeners, listener)
	return b
}

// OnError 设置错误处理回调
func (b *LifecycleBuilder) OnError(handler func(oldPhase, newPhase ApplicationPhase, err error)) *LifecycleBuilder {
	b.onError = handler
	return b
}

// Build 构建生命周期管理器
func (b *LifecycleBuilder) Build() *LifecycleManager {
	mgr := NewLifecycleManager()
	mgr.phase = b.initialPhase

	for _, listener := range b.listeners {
		mgr.AddListener(listener)
	}

	if b.onError != nil {
		mgr.SetErrorHandler(b.onError)
	}

	return mgr
}

// PhaseChecker 阶段检查器，提供便捷的阶段判断方法
type PhaseChecker struct {
	manager *LifecycleManager
}

// NewPhaseChecker 创建阶段检查器
func NewPhaseChecker(manager *LifecycleManager) *PhaseChecker {
	return &PhaseChecker{manager: manager}
}

// IsInitializing 是否处于初始化阶段
func (c *PhaseChecker) IsInitializing() bool {
	return c.manager.GetPhase() == PhaseInitializing
}

// IsConfiguring 是否处于配置阶段
func (c *PhaseChecker) IsConfiguring() bool {
	return c.manager.GetPhase() == PhaseConfiguring
}

// IsContextRefreshed 是否处于上下文刷新完成阶段
func (c *PhaseChecker) IsContextRefreshed() bool {
	return c.manager.GetPhase() == PhaseContextRefreshed
}

// IsReady 是否处于就绪阶段
func (c *PhaseChecker) IsReady() bool {
	return c.manager.GetPhase() == PhaseReady
}

// IsRunning 是否处于运行阶段
func (c *PhaseChecker) IsRunning() bool {
	return c.manager.GetPhase() == PhaseRunning
}

// IsStopping 是否处于停止阶段
func (c *PhaseChecker) IsStopping() bool {
	return c.manager.GetPhase() == PhaseStopping
}

// IsStopped 是否已停止
func (c *PhaseChecker) IsStopped() bool {
	return c.manager.GetPhase() == PhaseStopped
}

// IsBefore 检查当前阶段是否在指定阶段之前
func (c *PhaseChecker) IsBefore(phase ApplicationPhase) bool {
	return c.manager.GetPhase() < phase
}

// IsAfter 检查当前阶段是否在指定阶段之后
func (c *PhaseChecker) IsAfter(phase ApplicationPhase) bool {
	return c.manager.GetPhase() > phase
}

// IsAtOrBefore 检查当前阶段是否在指定阶段之前或等于
func (c *PhaseChecker) IsAtOrBefore(phase ApplicationPhase) bool {
	return c.manager.GetPhase() <= phase
}

// IsAtOrAfter 检查当前阶段是否在指定阶段之后或等于
func (c *PhaseChecker) IsAtOrAfter(phase ApplicationPhase) bool {
	return c.manager.GetPhase() >= phase
}

// PhaseTransition 阶段转换辅助结构
type PhaseTransition struct {
	OldPhase ApplicationPhase
	NewPhase ApplicationPhase
}

// String 返回转换描述
func (t PhaseTransition) String() string {
	return fmt.Sprintf("%s -> %s", t.OldPhase, t.NewPhase)
}

// PhaseTransitionRecorder 阶段转换记录器
type PhaseTransitionRecorder struct {
	mu          sync.RWMutex
	transitions []PhaseTransition
}

// NewPhaseTransitionRecorder 创建阶段转换记录器
func NewPhaseTransitionRecorder() *PhaseTransitionRecorder {
	return &PhaseTransitionRecorder{}
}

// OnPhaseChange 实现 PhaseListener 接口
func (r *PhaseTransitionRecorder) OnPhaseChange(oldPhase, newPhase ApplicationPhase) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.transitions = append(r.transitions, PhaseTransition{
		OldPhase: oldPhase,
		NewPhase: newPhase,
	})
	return nil
}

// GetTransitions 获取所有转换记录
func (r *PhaseTransitionRecorder) GetTransitions() []PhaseTransition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]PhaseTransition, len(r.transitions))
	copy(result, r.transitions)
	return result
}

// GetLastTransition 获取最后一次转换
func (r *PhaseTransitionRecorder) GetLastTransition() (PhaseTransition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.transitions) == 0 {
		return PhaseTransition{}, false
	}
	return r.transitions[len(r.transitions)-1], true
}

// Count 获取转换次数
func (r *PhaseTransitionRecorder) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.transitions)
}
