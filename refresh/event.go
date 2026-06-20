package refresh

import "time"

// ConfigChangeEvent 配置变更事件
//
// 当配置中心（如 Nacos、Etcd）的配置发生变更时发布此事件。
// 包含变更类型、变更键列表、新旧值和来源信息。
type ConfigChangeEvent struct {
	EventType string            // 事件类型："modify"、"delete"、"create"
	Keys      []string          // 变更的配置键列表
	OldValues map[string]any    // 变更前的值
	NewValues map[string]any    // 变更后的值
	Source    string            // 配置源类型（如 "nacos"、"etcd"）
	timestamp time.Time         // 事件发生时间
	Metadata  map[string]string // 额外元数据
}

// Type 返回事件类型标识
func (e *ConfigChangeEvent) Type() string {
	return "ConfigChange"
}

// Timestamp 返回事件发生时间
func (e *ConfigChangeEvent) Timestamp() time.Time {
	return e.timestamp
}

// BeanRefreshedEvent Bean 刷新完成事件
//
// 当 Bean 成功或失败地完成刷新后发布此事件。
type BeanRefreshedEvent struct {
	BeanID      string    // Bean 标识
	OldVersion  int64     // 刷新前版本号
	NewVersion  int64     // 刷新后版本号
	RefreshTime time.Time // 刷新完成时间
	Success     bool      // 是否刷新成功
	Error       error     // 刷新失败的错误信息
}

// RefreshFailedEvent 刷新失败事件
//
// 当 Bean 刷新过程中发生错误时发布此事件。
type RefreshFailedEvent struct {
	BeanID     string    // 失败的 Bean 标识
	ConfigKeys []string  // 触发刷新的配置键
	Error      error     // 失败原因
	Timestamp  time.Time // 失败时间
}

// RefreshableBean 可刷新 Bean 接口
//
// 实现 此接口 的 Bean 会在配置变更时收到 OnConfigChange 回调，
// 可以据此更新内部状态。
type RefreshableBean interface {
	// OnConfigChange 处理配置变更事件
	OnConfigChange(event ConfigChangeEvent) error
}

// NewConfigChangeEvent 创建配置变更事件
//
// 参数：
//   - eventType: 事件类型（"modify"、"delete"、"create"）
//   - keys: 变更的配置键列表
//   - oldValues: 变更前的值
//   - newValues: 变更后的值
//   - source: 配置源类型
func NewConfigChangeEvent(eventType string, keys []string, oldValues, newValues map[string]any, source string) ConfigChangeEvent {
	return ConfigChangeEvent{
		EventType: eventType,
		Keys:      keys,
		OldValues: oldValues,
		NewValues: newValues,
		Source:    source,
		timestamp: time.Now(),
		Metadata:  make(map[string]string),
	}
}
