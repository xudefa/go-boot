package refresh

import (
	"log/slog"
	"sync"
	"sync/atomic"
)

// RefreshProxy Bean 刷新代理
//
// 使用代理模式实现延迟刷新。消费者持有代理引用，
// 代理内部在刷新标记被设置后重新创建目标实例。
//
// 刷新流程：
//  1. MarkForRefresh 设置刷新标记
//  2. GetTarget 检测到刷新标记，加锁创建新实例
//  3. 使用双重检查锁定避免不必要的加锁
type RefreshProxy struct {
	target       atomic.Value         // 存储实际的 Bean 实例
	needsRefresh atomic.Bool          // 是否需要刷新
	mu           sync.Mutex           // 保护刷新过程
	beanID       string               // Bean 标识
	manager      *RefreshScopeManager // 所属的刷新作用域管理器
	logger       *slog.Logger         // 日志记录器
}

// NewRefreshProxy 创建刷新代理
//
// 参数：
//   - beanID: Bean 标识
//   - target: 初始 Bean 实例
//   - manager: 刷新作用域管理器
//   - logger: 日志记录器
func NewRefreshProxy(beanID string, target any, manager *RefreshScopeManager, logger *slog.Logger) *RefreshProxy {
	proxy := &RefreshProxy{
		beanID:  beanID,
		manager: manager,
		logger:  logger,
	}
	proxy.target.Store(target)
	return proxy
}

// GetTarget 获取目标 Bean 实例
//
// 如果需要刷新，使用双重检查锁定模式创建新实例：
//  1. 快速路径：无需刷新时直接返回当前实例（无锁）
//  2. 慢速路径：加锁后再次检查，确认需要刷新后创建新实例
//  3. 创建失败时返回旧实例，保证服务可用性
func (p *RefreshProxy) GetTarget() any {
	if !p.needsRefresh.Load() {
		return p.target.Load()
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.needsRefresh.Load() {
		return p.target.Load()
	}

	newBean, err := p.manager.createBean(p.beanID)
	if err != nil {
		p.logger.Error("Failed to refresh bean", "beanID", p.beanID, "error", err)
		return p.target.Load()
	}

	p.target.Store(newBean)
	p.needsRefresh.Store(false)
	p.logger.Info("Bean refreshed successfully", "beanID", p.beanID)
	return newBean
}

// MarkForRefresh 标记代理需要在下次 GetTarget 时刷新
func (p *RefreshProxy) MarkForRefresh() {
	p.needsRefresh.Store(true)
}
