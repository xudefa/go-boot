// Package refresh 提供配置热刷新机制，参考 Spring Cloud 的 @RefreshScope。
//
// 当配置中心（如 Nacos）的配置发生变更时，RefreshScopeManager 会：
//  1. 接收配置变更事件
//  2. 标记受影响的 Bean 为需要刷新
//  3. 通过代理模式在下次访问时重建 Bean 实例
//
// 核心组件：
//   - RefreshScopeManager: 刷新作用域管理器，协调刷新流程
//   - RefreshProxy: 代理模式，延迟刷新 Bean 实例
//   - RefreshableBean: 可刷新 Bean 接口
//   - RefreshMetrics: 刷新指标收集
package refresh

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xudefa/go-boot/core"
)

// RefreshScopeManager 刷新作用域管理器
//
// 参考 Spring Cloud 的 @RefreshScope，管理可刷新 Bean 的生命周期。
// 当配置变更时，标记受影响的 Bean 并在下次访问时重建实例。
//
// 使用代理模式实现延迟刷新：Bean 的消费者持有代理引用，
// 代理内部在刷新标记被设置后重新创建目标实例。
type RefreshScopeManager struct {
	beanCreator      core.BeanCreator           // Bean 创建器，用于重建实例
	config           *RefreshConfig             // 刷新配置
	refreshFlags     map[string]bool            // Bean 刷新标记
	beanVersions     map[string]*atomic.Int64   // Bean 版本号（递增）
	activeProxies    map[string]*RefreshProxy   // 活跃的代理
	refreshableBeans map[string]RefreshableBean // 可刷新 Bean 注册表
	metrics          *RefreshMetrics            // 刷新指标
	mu               sync.RWMutex               // 保护并发访问
	logger           *slog.Logger               // 日志记录器
}

// NewRefreshScopeManager 创建刷新作用域管理器
//
// 参数：
//   - beanCreator: Bean 创建器，用于重建 Bean 实例
//   - logger: 日志记录器，nil 时使用配置中的默认值
//   - opts: 刷新配置选项
func NewRefreshScopeManager(beanCreator core.BeanCreator, logger *slog.Logger, opts ...RefreshOption) *RefreshScopeManager {
	config := DefaultRefreshConfig()
	config.ApplyOptions(opts)

	if logger != nil {
		config.Logger = logger
	}

	return &RefreshScopeManager{
		beanCreator:      beanCreator,
		config:           config,
		refreshFlags:     make(map[string]bool),
		beanVersions:     make(map[string]*atomic.Int64),
		activeProxies:    make(map[string]*RefreshProxy),
		refreshableBeans: make(map[string]RefreshableBean),
		metrics:          NewRefreshMetrics(),
		logger:           config.Logger,
	}
}

// MarkBeanForRefresh 标记 Bean 为需要刷新
//
// 设置刷新标记并通知代理，代理会在下次 GetTarget 时重建实例。
func (m *RefreshScopeManager) MarkBeanForRefresh(beanID string) {
	startTime := time.Now()
	defer func() {
		m.metrics.RecordRefresh(time.Since(startTime), true)
	}()

	m.mu.Lock()
	defer m.mu.Unlock()

	m.refreshFlags[beanID] = true

	if proxy, ok := m.activeProxies[beanID]; ok {
		proxy.MarkForRefresh()
	}

	m.logger.Info("Bean marked for refresh", "beanID", beanID)
}

// GetRefreshedBean 获取刷新后的 Bean 实例
//
// 通过代理获取最新的 Bean 实例。如果代理不存在，返回错误。
func (m *RefreshScopeManager) GetRefreshedBean(beanID string) (any, error) {
	m.mu.RLock()
	proxy, ok := m.activeProxies[beanID]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("bean not found: %s", beanID)
	}

	return proxy.GetTarget(), nil
}

// RegisterRefreshableBean 注册可刷新 Bean
//
// 注册后，该 Bean 会在配置变更时收到 OnConfigChange 通知。
func (m *RefreshScopeManager) RegisterRefreshableBean(beanID string, bean RefreshableBean) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.refreshableBeans[beanID] = bean
	m.logger.Info("Refreshable bean registered", "beanID", beanID)
}

// createBean 创建新的 Bean 实例（内部方法）
func (m *RefreshScopeManager) createBean(beanID string) (any, error) {
	return m.beanCreator.CreateBean(beanID)
}

// incrementBeanVersion 增加 Bean 版本号
func (m *RefreshScopeManager) incrementBeanVersion(beanID string) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	version, ok := m.beanVersions[beanID]
	if !ok {
		version = &atomic.Int64{}
		m.beanVersions[beanID] = version
	}

	return version.Add(1)
}

// OnConfigChange 处理配置变更事件
//
// 遍历所有已注册的可刷新 Bean，逐一通知配置变更。
// 如果某个 Bean 通知失败，记录错误日志但不中断其他 Bean 的通知。
func (m *RefreshScopeManager) OnConfigChange(event ConfigChangeEvent) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 通知所有 RefreshableBean
	for beanID, bean := range m.refreshableBeans {
		if err := bean.OnConfigChange(event); err != nil {
			m.logger.Error("Failed to notify refreshable bean",
				"beanID", beanID,
				"error", err,
			)
		}
	}
}

// Metrics 返回刷新指标
func (m *RefreshScopeManager) Metrics() *RefreshMetrics {
	return m.metrics
}
