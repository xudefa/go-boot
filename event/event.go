// Package event 提供应用事件驱动支持，参考 Spring 的 ApplicationEvent/ApplicationListener 模式。
//
// 核心功能：
//   - ApplicationEvent: 应用事件接口，所有自定义事件需实现此接口
//   - EventBus: 事件总线，支持事件的发布和订阅
//   - BaseEvent: 基础事件实现，可直接使用或嵌入自定义事件结构体
//
// 内置事件类型：
//   - EventEnvironmentPrepared: 环境准备完成
//   - EventContextRefreshed: 上下文刷新完成
//   - EventApplicationStarted: 应用启动
//   - EventApplicationReady: 应用就绪
//   - EventApplicationStopped: 应用停止
//
// 使用示例：
//
//	// 创建事件总线
//	bus := event.NewEventBus()
//
//	// 订阅事件
//	bus.Subscribe(event.EventApplicationStarted, func(e event.ApplicationEvent) {
//	    fmt.Println("应用已启动")
//	})
//
//	// 发布事件
//	bus.Publish(&event.BaseEvent{EventType: event.EventApplicationStarted})
package event

import (
	"reflect"
	"sync"
	"time"
)

// ApplicationEvent 应用事件接口
//
// 所有应用事件需实现此接口，通过 Type() 返回的事件类型字符串进行路由。
type ApplicationEvent interface {
	// Type 返回事件类型字符串，用于事件路由和匹配
	Type() string
	// Timestamp 返回事件发生的时间戳
	Timestamp() time.Time
}

// EventListener 事件监听器函数类型
//
// 接收 ApplicationEvent 参数，处理事件通知。
type EventListener func(event ApplicationEvent)

// EventBus 事件总线
//
// 负责事件的发布与订阅管理，支持多监听器注册。
// 线程安全，支持并发发布和订阅。
type EventBus struct {
	mu        sync.RWMutex               // 保护 listeners 的读写锁
	listeners map[string][]EventListener // 事件类型到监听器列表的映射
}

// NewEventBus 创建新的事件总线实例
func NewEventBus() *EventBus {
	return &EventBus{
		listeners: make(map[string][]EventListener),
	}
}

// Publish 发布事件，通知所有订阅了该事件类型的监听器
//
// 发布流程：
//  1. 获取事件类型的监听器列表（读锁保护）
//  2. 如果无监听器则直接返回
//  3. 遍历监听器列表并逐个调用
func (b *EventBus) Publish(event ApplicationEvent) {
	b.mu.RLock()
	listeners, ok := b.listeners[event.Type()]
	if !ok {
		b.mu.RUnlock()
		return
	}
	snapshot := make([]EventListener, len(listeners))
	copy(snapshot, listeners)
	b.mu.RUnlock()
	for _, listener := range snapshot {
		listener(event)
	}
}

// Subscribe 订阅指定类型的事件
//
// 参数:
//   - eventType: 事件类型字符串，与 ApplicationEvent.Type() 返回值对应
//   - listener: 事件监听器函数
func (b *EventBus) Subscribe(eventType string, listener EventListener) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.listeners[eventType] = append(b.listeners[eventType], listener)
}

// Unsubscribe 取消订阅指定类型的事件
//
// 使用 reflect 比较函数指针来定位要移除的监听器。
// 因为 Go 语言中函数类型不支持 == 直接比较，需要通过 reflect.ValueOf().Pointer() 获取函数指针进行比较。
func (b *EventBus) Unsubscribe(eventType string, target EventListener) {
	if target == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	listeners := b.listeners[eventType]
	targetPtr := reflect.ValueOf(target).Pointer()
	for i, listener := range listeners {
		if reflect.ValueOf(listener).Pointer() == targetPtr {
			b.listeners[eventType] = append(listeners[:i], listeners[i+1:]...)
			return
		}
	}
}

// BaseEvent 基础事件实现
//
// 可直接使用，也支持嵌入到自定义事件结构体中。
// 如果 EventTime 未设置，Timestamp() 会自动返回当前时间。
type BaseEvent struct {
	EventType string    // 事件类型
	EventTime time.Time // 事件发生时间（可选，为空时自动使用当前时间）
}

func (e *BaseEvent) Type() string {
	return e.EventType
}

func (e *BaseEvent) Timestamp() time.Time {
	if e.EventTime.IsZero() {
		return time.Now()
	}
	return e.EventTime
}

// 内置事件类型常量
const (
	EventEnvironmentPrepared = "EnvironmentPrepared"
	EventContextRefreshed    = "ContextRefreshed"
	EventApplicationStarted  = "ApplicationStarted"
	EventApplicationReady    = "ApplicationReady"
	EventApplicationStopped  = "ApplicationStopped"
)
