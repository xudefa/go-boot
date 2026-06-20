// Package net 提供 HTTP 服务器/客户端抽象接口、中间件和 WebSocket 接口定义。
//
// 核心组件：
//   - HandlerContext: HTTP 请求处理上下文抽象接口
//   - MiddlewareFunc: 中间件函数类型
//   - Server: HTTP 服务器抽象接口
//   - HttpClient: RESTful 客户端接口
//   - WebSocketServer: WebSocket 服务器抽象接口
package net

import "context"

// HandlerContext HTTP 请求处理上下文抽象接口
//
// 封装 HTTP 请求和响应操作，支持多种 HTTP 框架适配。
// 所有 HTTP 框架（如 Gin、Hertz）都应实现此接口，
// 以便与 go-boot 中间件体系集成。
type HandlerContext interface {
	// RequestMethod 返回请求方法（GET、POST 等）
	RequestMethod() string

	// RequestURI 返回请求 URI
	RequestURI() string

	// Header 获取指定请求头的值
	Header(key string) string

	// SetStatusCode 设置响应状态码
	SetStatusCode(code int)

	// SetHeader 设置响应头
	SetHeader(key, value string)

	// AbortWithStatus 中止请求并返回指定状态码
	AbortWithStatus(code int)

	// AbortWithStatusJSON 中止请求并返回 JSON 响应
	AbortWithStatusJSON(code int, body interface{})

	// Next 调用下一个中间件或处理器
	Next()

	// IsAborted 判断请求是否已被中止
	IsAborted() bool

	// Context 获取请求上下文
	Context() context.Context

	// SetContext 设置请求上下文
	SetContext(ctx context.Context)
}

// MiddlewareFunc 中间件函数类型
//
// 接收 HandlerContext 并在处理过程中调用 ctx.Next() 传递到下一个中间件。
type MiddlewareFunc func(HandlerContext)
