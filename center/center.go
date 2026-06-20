// Package center 提供服务注册与发现的抽象接口，以及负载均衡选择器。
//
// 核心接口：
//   - Registry: 注册中心，支持 Register、Deregister、Discover、Watch
//   - Selector: 负载均衡选择器，支持随机、轮询、加权随机等策略
//
// 内置选择器：
//   - center.RandomSelect: 随机选择
//   - center.RoundRobinSelect: 轮询选择
//   - center.NewWeightedRandomSelect(): 加权随机选择
package center

import "context"

// InstanceInfo 注册中心中的服务实例信息
//
// 包含服务名、实例 ID、网络地址、权重、健康状态和元数据。
type InstanceInfo struct {
	ServiceName string            // 服务名称
	ID          string            // 实例唯一标识
	Host        string            // 主机地址
	Port        int               // 端口号
	Weight      int               // 负载均衡权重
	Healthy     bool              // 是否健康
	Metadata    map[string]string // 扩展元数据
}

// Registry 注册中心的核心抽象接口
//
// 提供服务实例的注册、注销、发现和监听功能。
//
// 使用示例：
//
//	var reg center.Registry
//	err := reg.Register(ctx, center.InstanceInfo{
//	    ServiceName: "user-service",
//	    ID:          "192.168.1.1:8080",
//	    Host:        "192.168.1.1",
//	    Port:        8080,
//	})
//	instances, err := reg.Discover(ctx, "user-service")
type Registry interface {
	// Register 注册服务实例到注册中心
	Register(ctx context.Context, info InstanceInfo) error
	// Deregister 从注册中心注销服务实例
	Deregister(ctx context.Context, info InstanceInfo) error
	// Discover 发现指定服务的所有实例
	Discover(ctx context.Context, serviceName string) ([]InstanceInfo, error)
	// Watch 监听服务实例变更，返回实例列表变更通道
	Watch(ctx context.Context, serviceName string) (<-chan []InstanceInfo, error)
}
