package refresh

import (
	"sync"
	"time"

	"github.com/xudefa/go-boot/event"
)

// EventRouter 配置变更事件路由器
//
// 维护 Bean 与配置键的双向映射关系，当配置变更事件到达时，
// 查找受影响的 Bean 并发布 BeanRefreshEvent。
type EventRouter struct {
	eventBus      *event.EventBus     // 事件总线
	beanConfigMap map[string][]string // Bean ID -> 依赖的配置键列表
	configBeanMap map[string][]string // 配置键 -> 依赖该键的 Bean ID 列表
	mu            sync.RWMutex        // 保护映射表的并发访问
}

// NewEventRouter 创建事件路由器
//
// 创建后自动订阅事件总线的 "ConfigChange" 事件。
func NewEventRouter(eventBus *event.EventBus) *EventRouter {
	router := &EventRouter{
		eventBus:      eventBus,
		beanConfigMap: make(map[string][]string),
		configBeanMap: make(map[string][]string),
	}

	eventBus.Subscribe("ConfigChange", router.onConfigChange)

	return router
}

// RegisterBean 注册 Bean 及其依赖的配置键
//
// 建立 Bean 与配置键的双向映射，用于后续的事件路由。
func (r *EventRouter) RegisterBean(beanID string, configKeys []string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.beanConfigMap[beanID] = configKeys

	for _, key := range configKeys {
		r.configBeanMap[key] = append(r.configBeanMap[key], beanID)
	}
}

// onConfigChange 处理配置变更事件
//
// 查找受影响的 Bean，为每个 Bean 发布 BeanRefreshEvent。
func (r *EventRouter) onConfigChange(e event.ApplicationEvent) {
	event, ok := e.(*ConfigChangeEvent)
	if !ok {
		return
	}

	affectedBeans := r.findAffectedBeans(event.Keys)

	for _, beanID := range affectedBeans {
		refreshEvent := &BeanRefreshEvent{
			BeanID:     beanID,
			ConfigKeys: event.Keys,
			OldValues:  event.OldValues,
			NewValues:  event.NewValues,
		}
		r.eventBus.Publish(refreshEvent)
	}
}

// findAffectedBeans 查找受配置变更影响的 Bean
func (r *EventRouter) findAffectedBeans(keys []string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	affected := make(map[string]bool)
	for _, key := range keys {
		if beanIDs, ok := r.configBeanMap[key]; ok {
			for _, beanID := range beanIDs {
				affected[beanID] = true
			}
		}
	}

	result := make([]string, 0, len(affected))
	for beanID := range affected {
		result = append(result, beanID)
	}
	return result
}

// BeanRefreshEvent Bean 刷新事件
//
// 当配置变更导致 Bean 需要刷新时发布此事件。
type BeanRefreshEvent struct {
	BeanID     string         // 需要刷新的 Bean 标识
	ConfigKeys []string       // 触发刷新的配置键
	OldValues  map[string]any // 变更前的值
	NewValues  map[string]any // 变更后的值
}

// Type 返回事件类型标识
func (e *BeanRefreshEvent) Type() string {
	return "BeanRefresh"
}

// Timestamp 返回事件发生时间
func (e *BeanRefreshEvent) Timestamp() time.Time {
	return time.Now()
}
