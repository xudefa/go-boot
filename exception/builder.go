package exception

import (
	"context"
	"fmt"
	"reflect"
)

// ExceptionHandlerBuilder 异常处理器构建器，支持链式配置
type ExceptionHandlerBuilder struct {
	opts              []Option
	resolvers         []ExceptionResolver
	exceptionHandlers map[reflect.Type]func(ctx context.Context, err error) *ErrorResponse
}

// NewExceptionHandlerBuilder 创建异常处理器构建器
func NewExceptionHandlerBuilder() *ExceptionHandlerBuilder {
	return &ExceptionHandlerBuilder{
		exceptionHandlers: make(map[reflect.Type]func(ctx context.Context, err error) *ErrorResponse),
	}
}

// Logger 设置日志记录器
func (b *ExceptionHandlerBuilder) Logger(logger Logger) *ExceptionHandlerBuilder {
	b.opts = append(b.opts, WithLogger(logger))
	return b
}

// MetricsRecorder 设置指标记录器
func (b *ExceptionHandlerBuilder) MetricsRecorder(recorder MetricsRecorder) *ExceptionHandlerBuilder {
	b.opts = append(b.opts, WithMetricsRecorder(recorder))
	return b
}

// IncludeStackTrace 设置是否包含堆栈跟踪
func (b *ExceptionHandlerBuilder) IncludeStackTrace(include bool) *ExceptionHandlerBuilder {
	b.opts = append(b.opts, WithIncludeStackTrace(include))
	return b
}

// Resolver 添加异常解析器
func (b *ExceptionHandlerBuilder) Resolver(resolver ExceptionResolver) *ExceptionHandlerBuilder {
	b.resolvers = append(b.resolvers, resolver)
	return b
}

// ExceptionHandler 为指定异常类型注册处理函数
func (b *ExceptionHandlerBuilder) ExceptionHandler(exceptionType reflect.Type, handler func(ctx context.Context, err error) *ErrorResponse) *ExceptionHandlerBuilder {
	b.exceptionHandlers[exceptionType] = handler
	return b
}

// Build 构建异常处理器
func (b *ExceptionHandlerBuilder) Build() (ExceptionHandler, error) {
	handler := NewDefaultExceptionHandler(b.opts...)

	// 注册解析器
	for _, resolver := range b.resolvers {
		handler.RegisterResolver(resolver)
	}

	// 注册异常处理函数
	for exceptionType, handlerFunc := range b.exceptionHandlers {
		handler.RegisterHandlerFunc(exceptionType, handlerFunc)
	}

	return handler, nil
}

// MustBuild 构建异常处理器，失败则panic
func (b *ExceptionHandlerBuilder) MustBuild() ExceptionHandler {
	handler, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build exception handler: %v", err))
	}
	return handler
}

// ErrorResponseBuilder 错误响应构建器
type ErrorResponseBuilder struct {
	response ErrorResponse
}

// NewErrorResponseBuilder 创建错误响应构建器
func NewErrorResponseBuilder() *ErrorResponseBuilder {
	return &ErrorResponseBuilder{}
}

// Code 设置HTTP状态码
func (b *ErrorResponseBuilder) Code(code int) *ErrorResponseBuilder {
	b.response.Code = code
	return b
}

// Message 设置错误消息
func (b *ErrorResponseBuilder) Message(message string) *ErrorResponseBuilder {
	b.response.Message = message
	return b
}

// RequestID 设置请求ID
func (b *ErrorResponseBuilder) RequestID(requestID string) *ErrorResponseBuilder {
	b.response.RequestID = requestID
	return b
}

// TraceID 设置链路追踪ID
func (b *ErrorResponseBuilder) TraceID(traceID string) *ErrorResponseBuilder {
	b.response.TraceID = traceID
	return b
}

// Details 设置错误详情
func (b *ErrorResponseBuilder) Details(details any) *ErrorResponseBuilder {
	b.response.Details = details
	return b
}

// Build 构建错误响应
func (b *ErrorResponseBuilder) Build() ErrorResponse {
	return b.response
}

// ToJSON 将错误响应转换为JSON字节
func (b *ErrorResponseBuilder) ToJSON() ([]byte, error) {
	return b.response.ToJSON()
}
