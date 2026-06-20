// Package config 定义配置中心的核心接口和模型。
//
// ConfigCenter 接口由各集成模块（如 Nacos、Etcd、Consul）实现，
// 提供远程配置的加载、监听和关闭能力。
// ConfigCenterConfig 统一描述配置中心的连接参数。
package config

import (
	"strings"
	"time"
)

// ConfigData 配置数据，键值对形式存储从配置中心加载的配置项。
type ConfigData map[string]any

// ConfigCenter 配置中心接口
//
// 定义远程配置的加载、监听和关闭行为。
// 各配置中心实现（如 Nacos、Etcd、Consul）需实现此接口。
type ConfigCenter interface {
	// Load 从配置中心加载配置数据
	Load() (ConfigData, error)

	// Watch 监听指定键的配置变更，变更时触发回调
	Watch(key string, callback func(ConfigData)) error

	// Close 关闭配置中心连接并释放资源
	Close() error
}

// ConfigCenterConfig 配置中心连接配置
//
// 字段说明：
//   - Endpoints: 配置中心服务端地址列表
//   - Namespace: 命名空间（Nacos 使用）
//   - Timeout: 连接超时时间
//   - DataID: 数据 ID（Nacos 使用）
//   - Group: 配置分组（Nacos 使用）
//   - Prefix: 配置前缀（Etcd/Consul 使用）
type ConfigCenterConfig struct {
	Endpoints []string
	Namespace string
	Timeout   time.Duration
	DataID    string
	Group     string
	Prefix    string
}

// WithDataID 设置配置中心的数据 ID（Nacos 使用）。
func WithDataID(dataID string) func(*ConfigCenterConfig) {
	return func(cfg *ConfigCenterConfig) {
		cfg.DataID = dataID
	}
}

// WithGroup 设置配置中心的分组（Nacos 使用）。
func WithGroup(group string) func(*ConfigCenterConfig) {
	return func(cfg *ConfigCenterConfig) {
		cfg.Group = group
	}
}

// WithFormat 设置配置中心的前缀（Etcd/Consul 使用）。
func WithFormat(format string) func(*ConfigCenterConfig) {
	return func(cfg *ConfigCenterConfig) {
		cfg.Prefix = format
	}
}

// WithProfiles 设置配置中心的命名空间（将 Profile 列表以逗号连接）。
func WithProfiles(profiles []string) func(*ConfigCenterConfig) {
	return func(cfg *ConfigCenterConfig) {
		cfg.Namespace = strings.Join(profiles, ",")
	}
}
