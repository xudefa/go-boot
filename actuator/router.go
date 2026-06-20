// Package actuator 提供应用监控端点，支持健康检查、指标采集和环境变量查看。
package actuator

import "net/http"

// RouteRegistrar 路由注册器接口
//
// 解耦路由注册逻辑，支持不同的 HTTP 框架实现。
type RouteRegistrar interface {
	// Handle 注册路由处理器
	Handle(pattern string, handler http.Handler)
}

// StdRouteRegistrar 标准库 HTTP 路由注册器
type StdRouteRegistrar struct {
	Mux interface {
		Handle(pattern string, handler http.Handler)
	}
}

// Handle 注册路由处理器
func (r *StdRouteRegistrar) Handle(pattern string, handler http.Handler) {
	r.Mux.Handle(pattern, handler)
}

// RouteConfig 路由配置
type RouteConfig struct {
	BasePath    string
	ExposeDebug bool
	Prefix      string
}

// DefaultRouteConfig 返回默认路由配置
func DefaultRouteConfig() RouteConfig {
	return RouteConfig{
		BasePath:    "/actuator",
		ExposeDebug: false,
		Prefix:      "",
	}
}
