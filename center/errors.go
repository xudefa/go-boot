// Package center 提供注册中心抽象，包括服务注册、发现和负载均衡选择器。
// Registry 接口定义服务注册与发现操作，Selector 接口定义负载均衡策略。
package center

import "errors"

var (
	// ErrNoInstances 表示没有可用的服务实例可供选择。
	// 当 Selector.Select() 传入空实例列表或 nil 时返回此错误。
	ErrNoInstances = errors.New("center: no available instances")

	// ErrInvalidInstance 表示服务实例信息无效。
	// 当注册或发现过程中遇到格式错误或不完整的实例信息时返回此错误。
	ErrInvalidInstance = errors.New("center: invalid instance info")
)
