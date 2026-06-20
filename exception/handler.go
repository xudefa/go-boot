package exception

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
)

// DefaultExceptionHandler 默认异常处理器实现
//
// DefaultExceptionHandler 是 ExceptionHandler 接口的主要实现，提供了完整的异常处理功能：
// - 使用解析器链处理不同类型的异常
// - 支持日志和监控集成
// - 自动写入 HTTP 响应
// - 支持基于类型的异常处理函数注册
type DefaultExceptionHandler struct {
	chain   *ResolverChain
	config  ExceptionHandlerConfig
	typeMap map[reflect.Type]func(ctx context.Context, err error) *ErrorResponse
}

// NewDefaultExceptionHandler 创建默认异常处理器
//
// 支持通过 Option 函数配置日志记录器、指标记录器和堆栈跟踪选项。
// 默认情况下，会注册内置异常解析器和默认异常解析器。
//
// 示例：
//
//	handler := exception.NewDefaultExceptionHandler(
//		exception.WithLogger(logger),
//		exception.WithMetricsRecorder(metrics),
//		exception.WithIncludeStackTrace(true),
//	)
func NewDefaultExceptionHandler(opts ...Option) ExceptionHandler {
	config := ExceptionHandlerConfig{
		IncludeStackTrace: false,
	}

	for _, opt := range opts {
		opt(&config)
	}

	chain := NewResolverChain()
	chain.AddResolver(NewBuiltinExceptionResolver())
	chain.AddResolver(NewDefaultExceptionResolver())

	return &DefaultExceptionHandler{
		chain:   chain,
		config:  config,
		typeMap: make(map[reflect.Type]func(ctx context.Context, err error) *ErrorResponse),
	}
}

// Handle 处理异常
//
// Handle 方法是异常处理的核心方法，它会：
// 1. 检查异常是否为 nil
// 2. 查找类型匹配的处理函数
// 3. 使用解析器链查找合适的解析器
// 4. 如果没有找到解析器，返回 500 错误
// 5. 将响应写入 ResponseWriter
// 6. 记录日志和指标（如果配置）
func (h *DefaultExceptionHandler) Handle(ctx context.Context, err error, response ResponseWriter) *ErrorResponse {
	if err == nil {
		return nil
	}

	var resp *ErrorResponse

	errType := reflect.TypeOf(err)
	if handlerFunc, ok := h.typeMap[errType]; ok {
		resp = handlerFunc(ctx, err)
	}

	if resp == nil {
		resp = h.chain.Resolve(ctx, err)
	}

	if resp == nil {
		resp = NewErrorResponse(500, "Internal Server Error", "", "", nil)
	}

	if response != nil {
		response.SetStatusCode(resp.Code)
		response.SetHeader("Content-Type", "application/json")
		data, marshalErr := json.Marshal(resp)
		if marshalErr != nil {
			fmt.Printf("[go-boot] failed to marshal exception response: %v\n", marshalErr)
		}
		if writeErr := response.Write(data); writeErr != nil {
			fmt.Printf("[go-boot] failed to write exception response: %v\n", writeErr)
		}
	}

	if h.config.Logger != nil {
		h.config.Logger.Error(ctx, "exception handled",
			KeyValue{Key: "exception_type", Value: reflect.TypeOf(err).String()},
			KeyValue{Key: "message", Value: resp.Message},
			KeyValue{Key: "code", Value: resp.Code},
		)
	}

	if h.config.MetricsRecorder != nil {
		h.config.MetricsRecorder.RecordException(reflect.TypeOf(err).String(), resp.Code)
	}

	return resp
}

// RegisterResolver 注册解析器
//
// 将自定义解析器添加到解析器链中，解析器会自动按 Order 排序。
// 优先级数值越小，优先级越高。
func (h *DefaultExceptionHandler) RegisterResolver(resolver ExceptionResolver) {
	h.chain.AddResolver(resolver)
}

// RegisterException 注册异常类型和解析器
//
// 为特定异常类型注册专用解析器，该方法内部调用 RegisterResolver。
func (h *DefaultExceptionHandler) RegisterException(exceptionType reflect.Type, resolver ExceptionResolver) {
	h.chain.AddResolver(resolver)
}

// RegisterHandlerFunc 注册异常类型和处理函数
//
// 为特定异常类型注册处理函数，该函数会优先于解析器链执行。
// 这是一种更直接、更高效的异常处理方式。
func (h *DefaultExceptionHandler) RegisterHandlerFunc(exceptionType reflect.Type, handler func(ctx context.Context, err error) *ErrorResponse) {
	h.typeMap[exceptionType] = handler
}

// Option 配置选项
//
// Option 是用于配置 ExceptionHandler 的函数类型。
type Option func(*ExceptionHandlerConfig)

// WithLogger 设置日志记录器
//
// 为异常处理器配置日志记录器，所有异常处理都会被记录。
func WithLogger(logger Logger) Option {
	return func(c *ExceptionHandlerConfig) {
		c.Logger = logger
	}
}

// WithMetricsRecorder 设置指标记录器
//
// 为异常处理器配置指标记录器，所有异常都会被记录为指标。
func WithMetricsRecorder(recorder MetricsRecorder) Option {
	return func(c *ExceptionHandlerConfig) {
		c.MetricsRecorder = recorder
	}
}

// WithIncludeStackTrace 设置是否包含堆栈信息
//
// 配置是否在错误响应中包含堆栈跟踪信息。
// 注意：生产环境中建议关闭此选项以避免泄露敏感信息。
func WithIncludeStackTrace(include bool) Option {
	return func(c *ExceptionHandlerConfig) {
		c.IncludeStackTrace = include
	}
}
