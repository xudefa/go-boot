// Package exception 提供统一的应用层异常处理机制。
//
// 该包实现了类似 Spring Boot @ControllerAdvice 的全局异常处理功能，支持：
//   - 自定义异常到 HTTP 响应的映射
//   - 与日志和监控系统集成
//   - HTTP 中间件自动捕获异常
//   - 优先级排序的异常解析器链
//   - 安全模块适配器（AccessDeniedHandler、AuthenticationEntryPoint）
//
// 基本用法：
//
//	handler := exception.NewDefaultExceptionHandler()
//	resp := handler.Handle(context.Background(), exception.ErrNotFound, response)
//
// 使用 HTTP 中间件：
//
//	handler := exception.NewDefaultExceptionHandler()
//	middleware := exception.ExceptionHandlingMiddleware(handler)
//	http.Handle("/", middleware(httpHandler))
package exception

import (
	"context"
	"reflect"
)

// Logger 日志接口
//
// 抽象层，不依赖具体日志实现，支持 slog、zap 等多种日志框架适配。
type Logger interface {
	// Error 记录错误日志
	Error(ctx context.Context, msg string, keyValues ...KeyValue)
}

// KeyValue 日志键值对
type KeyValue struct {
	Key   string // 键名
	Value any    // 值
}

// MetricsRecorder 指标记录器接口
//
// 抽象层，用于记录异常指标到监控系统。
type MetricsRecorder interface {
	// RecordException 记录一次异常事件
	RecordException(exceptionType string, statusCode int)
}

// ResponseWriter 响应写入器接口
//
// 抽象层，适配不同 HTTP 框架的响应写入。
type ResponseWriter interface {
	// SetStatusCode 设置 HTTP 状态码
	SetStatusCode(code int)
	// SetHeader 设置响应头
	SetHeader(key, value string)
	// Write 写入响应体
	Write(data []byte) error
}

// ErrorResponse 统一错误响应
//
// 所有异常处理最终都转换为此结构体，确保 API 错误响应格式一致。
type ErrorResponse struct {
	Code      int    `json:"code"`                // HTTP 状态码
	Message   string `json:"message"`             // 错误消息
	RequestID string `json:"requestId,omitempty"` // 请求 ID
	TraceID   string `json:"traceId,omitempty"`   // 链路追踪 ID
	Details   any    `json:"details,omitempty"`   // 错误详情
	Timestamp int64  `json:"timestamp"`           // 时间戳
}

// ExceptionHandler 异常处理器接口
//
// 核心异常处理入口，支持注册自定义解析器和处理函数。
type ExceptionHandler interface {
	// Handle 处理异常并返回统一错误响应
	Handle(ctx context.Context, err error, response ResponseWriter) *ErrorResponse
	// RegisterResolver 注册通用异常解析器
	RegisterResolver(resolver ExceptionResolver)
	// RegisterException 为指定异常类型注册解析器
	RegisterException(exceptionType reflect.Type, resolver ExceptionResolver)
	// RegisterHandlerFunc 为指定异常类型注册处理函数
	RegisterHandlerFunc(exceptionType reflect.Type, handler func(ctx context.Context, err error) *ErrorResponse)
}

// ExceptionResolver 异常解析器接口
//
// 解析异常并生成统一错误响应，支持优先级排序。
// 多个解析器按 Order() 返回值从小到大依次尝试。
type ExceptionResolver interface {
	// Resolve 解析异常并返回错误响应
	Resolve(ctx context.Context, err error) *ErrorResponse
	// Supports 判断是否能处理该异常
	Supports(err error) bool
	// Order 返回解析器优先级，值越小优先级越高
	Order() int
}

// ExceptionHandlerConfig 异常处理器配置
type ExceptionHandlerConfig struct {
	Logger            Logger          // 日志记录器
	MetricsRecorder   MetricsRecorder // 指标记录器
	IncludeStackTrace bool            // 是否在响应中包含堆栈跟踪
}
