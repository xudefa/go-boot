// Package life 定义应用生命周期阶段管理和阶段监听器。
//
// 生命周期按以下顺序正向流转：
// PhaseInitializing → PhaseConfiguring → PhaseContextRefreshed → PhaseReady → PhaseRunning → PhaseStopping → PhaseStopped
//
// LifecycleManager 管理阶段转换并通知 PhaseListener，仅允许正向转换。
package life

import (
	"fmt"
	"sync"
)

// ApplicationPhase 应用生命周期阶段
type ApplicationPhase int

// 应用生命周期阶段定义，按时间顺序流转：
// PhaseInitializing → PhaseConfiguring → PhaseContextRefreshed → PhaseReady → PhaseRunning → PhaseStopping → PhaseStopped
const (
	PhaseInitializing     ApplicationPhase = iota // 初始化阶段：创建容器和基础组件
	PhaseConfiguring                              // 配置阶段：加载配置、注册 Bean
	PhaseContextRefreshed                         // 上下文刷新完成：所有 Bean 已注册
	PhaseReady                                    // 就绪阶段：应用准备就绪但尚未开始服务
	PhaseRunning                                  // 运行阶段：应用正常运行，处理请求
	PhaseStopping                                 // 停止阶段：应用正在停止，释放资源
	PhaseStopped                                  // 已停止：应用完全停止
)

// isForwardTransition 检查是否为有效的正向阶段转换。
// 规则：newPhase > oldPhase 且在当前阶段定义范围内为有效转换。
func isForwardTransition(oldPhase, newPhase ApplicationPhase) bool {
	return oldPhase >= PhaseInitializing && newPhase > oldPhase && newPhase <= PhaseStopped
}

var phaseNames = map[ApplicationPhase]string{
	PhaseInitializing:     "INITIALIZING",
	PhaseConfiguring:      "CONFIGURING",
	PhaseContextRefreshed: "CONTEXT_REFRESHED",
	PhaseReady:            "READY",
	PhaseRunning:          "RUNNING",
	PhaseStopping:         "STOPPING",
	PhaseStopped:          "STOPPED",
}

func (p ApplicationPhase) String() string {
	if name, ok := phaseNames[p]; ok {
		return name
	}
	return fmt.Sprintf("UNKNOWN(%d)", p)
}

// PhaseListener 生命周期阶段监听器
//
// 当应用生命周期阶段发生转换时回调。
// 如果回调返回错误，SetPhase 会返回第一个非 nil 错误。
type PhaseListener interface {
	// OnPhaseChange 处理阶段变更事件
	OnPhaseChange(oldPhase, newPhase ApplicationPhase) error
}

// LifecycleManager 生命周期管理器
//
// 管理应用生命周期阶段的转换和监听器通知。
// 仅允许正向阶段转换（PhaseInitializing → PhaseStopped），
// 逆向转换会返回错误。
type LifecycleManager struct {
	mu        sync.RWMutex                                         // 保护 phase 和 listeners 的并发访问
	phase     ApplicationPhase                                     // 当前阶段
	listeners []PhaseListener                                      // 阶段变更监听器列表
	onError   func(oldPhase, newPhase ApplicationPhase, err error) // 错误处理回调
}

// NewLifecycleManager 创建生命周期管理器
//
// 初始阶段为 PhaseInitializing。
func NewLifecycleManager() *LifecycleManager {
	return &LifecycleManager{
		phase: PhaseInitializing,
	}
}

// GetPhase 返回当前生命周期阶段
func (m *LifecycleManager) GetPhase() ApplicationPhase {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.phase
}

// SetPhase 设置新的生命周期阶段并通知所有监听器
//
// 仅允许正向转换，逆向转换返回错误。
// 监听器按注册顺序依次调用，第一个返回的错误会被记录。
func (m *LifecycleManager) SetPhase(newPhase ApplicationPhase) error {
	m.mu.Lock()
	oldPhase := m.phase
	if !isForwardTransition(oldPhase, newPhase) {
		m.mu.Unlock()
		return fmt.Errorf("invalid phase transition from %s to %s", oldPhase, newPhase)
	}
	m.phase = newPhase
	listeners := make([]PhaseListener, len(m.listeners))
	copy(listeners, m.listeners)
	m.mu.Unlock()

	var err error
	for _, listener := range listeners {
		if e := listener.OnPhaseChange(oldPhase, newPhase); e != nil && err == nil {
			err = e
		}
	}

	if err != nil && m.onError != nil {
		m.mu.RLock()
		handler := m.onError
		m.mu.RUnlock()
		if handler != nil {
			handler(oldPhase, newPhase, err)
		}
	}

	return err
}

// AddListener 添加生命周期阶段监听器
func (m *LifecycleManager) AddListener(listener PhaseListener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners = append(m.listeners, listener)
}

// SetErrorHandler 设置错误处理回调
func (m *LifecycleManager) SetErrorHandler(handler func(oldPhase, newPhase ApplicationPhase, err error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onError = handler
}
