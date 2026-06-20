// Package config 提供配置热重载支持。
//
// WatchManager 管理配置变更的监听和通知，支持多配置源和多监听器。
package config

import (
	"sync"
)

// 热重载事件类型常量.
const (
	EventModify = "modify" // 配置修改
	EventDelete = "delete" // 配置删除
	EventCreate = "create" // 配置创建
)

// WatchManager 配置热重载管理器.
//
// 管理配置变更的监听和通知.
type WatchManager struct {
	callbacks map[string]func(WatchEvent)
	mu        sync.RWMutex
	sources   map[string]chan WatchEvent
	closed    bool
}

// NewWatchManager 创建热重载管理器实例.
//
// 返回:
//   - *WatchManager: 管理器实例
func NewWatchManager() *WatchManager {
	return &WatchManager{
		callbacks: make(map[string]func(WatchEvent)),
		sources:   make(map[string]chan WatchEvent),
	}
}

// Register 注册配置变更监听器.
//
// 参数:
//   - key: 监听器标识
//   - callback: 回调函数
func (m *WatchManager) Register(key string, callback func(WatchEvent)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return
	}
	m.callbacks[key] = callback
}

// Unregister 取消配置变更监听.
//
// 参数:
//   - key: 监听器标识
func (m *WatchManager) Unregister(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return
	}
	delete(m.callbacks, key)
}

// Notify 通知所有监听器配置变更.
//
// 参数:
//   - event: 变更事件
func (m *WatchManager) Notify(event WatchEvent) {
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return
	}
	callbacks := make([]func(WatchEvent), 0, len(m.callbacks))
	for _, cb := range m.callbacks {
		if cb != nil {
			callbacks = append(callbacks, cb)
		}
	}
	m.mu.RUnlock()
	for _, cb := range callbacks {
		cb(event)
	}
}

// AddSource 添加配置源.
//
// 参数:
//   - name: 配置源名称
//   - ch: 事件通道
func (m *WatchManager) AddSource(name string, ch chan WatchEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return
	}
	m.sources[name] = ch
}

// GetSource 获取配置源.
//
// 参数:
//   - name: 配置源名称
//
// 返回:
//   - chan WatchEvent: 事件通道
//   - bool: 是否存在
func (m *WatchManager) GetSource(name string) (chan WatchEvent, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ch, ok := m.sources[name]
	return ch, ok
}

// Close 关闭热重载管理器.
func (m *WatchManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return
	}
	m.closed = true

	for name, ch := range m.sources {
		if ch != nil {
			close(ch)
		}
		delete(m.sources, name)
	}
	m.sources = nil
	m.callbacks = nil
}
