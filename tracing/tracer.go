// Package tracing 提供分布式追踪的抽象接口，支持接入 OpenTelemetry 等追踪系统。
//
// 核心接口：
//   - Tracer: 追踪器，负责创建和管理 Span
//   - Span: 追踪跨度，记录操作的时间范围和元数据
//   - SpanContext: 跨度的上下文信息（TraceID、SpanID）
//   - TracerProvider: 追踪器提供者，支持生命周期管理
//
// 内置 NoopTracer / NoopSpan 默认实现，避免空指针问题。
//
// 使用示例：
//
//	tracer := &tracing.NoopTracer{}
//	ctx, span := tracer.Start(context.Background(), "operation")
//	defer span.End()
//	span.SetAttribute("key", "value")
package tracing

import "context"

// SpanStatusCode Span 状态码
//
// 表示 Span 表示的操作的执行结果状态。
type SpanStatusCode int

const (
	SpanStatusUnset    SpanStatusCode = iota // 状态未设置（默认值）
	SpanStatusOK                             // 操作成功完成
	SpanStatusError                          // 操作发生错误
	SpanStatusCanceled                       // 操作被取消
)

// SpanKind Span 类型
//
// 表示 Span 在分布式追踪中的角色。
type SpanKind int

const (
	SpanKindUnspecified SpanKind = iota // 未指定类型
	SpanKindInternal                    // 内部操作（默认）
	SpanKindServer                      // 服务器端：接收客户端请求
	SpanKindClient                      // 客户端：发起对外请求
	SpanKindProducer                    // 生产者：发送消息到队列
	SpanKindConsumer                    // 消费者：从队列接收消息
)

// SpanOption Span 创建选项
//
// 用于配置 Span 的属性和行为。
type SpanOption func(*SpanConfig)

// SpanConfig Span 创建选项的配置存储
type SpanConfig struct {
	Attributes map[string]any // Span 属性键值对
	Kind       SpanKind       // Span 类型
}

// WithAttribute 添加 Span 属性
//
// key 为属性名称，value 为属性值，支持任意类型。
func WithAttribute(key string, value any) SpanOption {
	return func(c *SpanConfig) {
		if c.Attributes == nil {
			c.Attributes = make(map[string]any)
		}
		c.Attributes[key] = value
	}
}

// WithSpanKind 设置 Span 类型
//
// 指定 Span 在分布式追踪中的角色。
func WithSpanKind(kind SpanKind) SpanOption {
	return func(c *SpanConfig) {
		c.Kind = kind
	}
}

// EventOption Span 事件选项
//
// 用于配置 Span 事件的属性。
type EventOption func(*EventConfig)

// EventConfig Span 事件选项的配置存储
type EventConfig struct {
	Attributes map[string]any // 事件属性键值对
}

// WithEventAttribute 添加事件属性
//
// key 为属性名称，value 为属性值。
func WithEventAttribute(key string, value any) EventOption {
	return func(c *EventConfig) {
		if c.Attributes == nil {
			c.Attributes = make(map[string]any)
		}
		c.Attributes[key] = value
	}
}

// Span 表示一个追踪跨度
//
// Span 是分布式追踪的基本单位，表示一个操作的执行时间范围和元数据。
// 每个 Span 包含唯一的 SpanID 和关联的 TraceID，形成完整的调用链路。
//
// 使用示例：
//
//	ctx, span := tracer.Start(ctx, "database.query",
//	  tracing.WithAttribute("db.statement", "SELECT * FROM users"),
//	  tracing.WithSpanKind(tracing.SpanKindClient))
//	defer span.End()
type Span interface {
	// End 结束 Span，记录完成时间和资源清理
	End()

	// AddEvent 向 Span 添加事件，记录操作过程中的重要时刻
	// name: 事件名称
	// opts: 事件选项，可设置事件属性
	AddEvent(name string, opts ...EventOption)

	// SetAttribute 设置 Span 属性，用于标记操作特征
	// key: 属性键名
	// value: 属性值
	SetAttribute(key string, value any)

	// GetAttributes 获取所有 Span 属性
	GetAttributes() map[string]any

	// RecordError 记录错误信息，但不改变 Span 状态
	RecordError(err error)

	// SetError 设置错误状态，将 Span 标记为错误
	SetError(err error)

	// SetStatus 设置 Span 状态码
	SetStatus(code SpanStatusCode)

	// SpanContext 获取 Span 上下文信息
	SpanContext() SpanContext

	// GetTraceID 获取关联的 TraceID
	GetTraceID() string

	// GetSpanID 获取当前 Span 的 ID
	GetSpanID() string
}

// SpanContext 跨度的上下文信息
//
// 包含 TraceID、SpanID 和 TraceFlags，用于跨服务传递追踪上下文。
type SpanContext struct {
	TraceID    string // 追踪 ID，唯一标识一个分布式追踪
	SpanID     string // 跨度 ID，唯一标识一个 Span
	TraceFlags byte   // 追踪标志位
}

// Tracer 追踪器接口
//
// 负责创建和管理 Span，是分布式追踪的核心组件。
type Tracer interface {
	Start(ctx context.Context, spanName string, opts ...SpanOption) (context.Context, Span) // 创建并启动一个新 Span
	CurrentSpan(ctx context.Context) Span                                                   // 获取当前上下文中的 Span
	Finish()                                                                                // 清理资源（如果需要）
}

// TracerProvider 追踪器提供者
//
// 负责创建和管理 Tracer 的生命周期。
type TracerProvider interface {
	Tracer(name string) Tracer          // 创建指定名称的 Tracer
	Shutdown(ctx context.Context) error // 关闭提供者，释放资源
}

// NoopTracer 空操作的追踪器（默认实现）
//
// 所有方法都不执行任何实际操作，用于默认场景避免空指针。
type NoopTracer struct{}

func (n *NoopTracer) Start(ctx context.Context, spanName string, opts ...SpanOption) (context.Context, Span) {
	return ctx, &NoopSpan{}
}

func (n *NoopTracer) CurrentSpan(ctx context.Context) Span {
	return &NoopSpan{}
}

func (n *NoopTracer) Finish() {}

// NoopSpan 空操作的 Span
//
// 所有方法都不执行任何实际操作。
type NoopSpan struct{}

func (n *NoopSpan) End()                                      {}
func (n *NoopSpan) AddEvent(name string, opts ...EventOption) {}
func (n *NoopSpan) SetAttribute(key string, value any)        {}
func (n *NoopSpan) GetAttributes() map[string]any             { return nil }
func (n *NoopSpan) RecordError(err error)                     {}
func (n *NoopSpan) SetError(err error)                        {}
func (n *NoopSpan) SetStatus(code SpanStatusCode)             {}
func (n *NoopSpan) SpanContext() SpanContext                  { return SpanContext{} }
func (n *NoopSpan) GetTraceID() string                        { return "" }
func (n *NoopSpan) GetSpanID() string                         { return "" }

// NoopTracerProvider 空操作的 TracerProvider
//
// 返回 NoopTracer，不执行任何实际追踪操作。
type NoopTracerProvider struct{}

func (n *NoopTracerProvider) Tracer(name string) Tracer {
	return &NoopTracer{}
}

func (n *NoopTracerProvider) Shutdown(ctx context.Context) error {
	return nil
}

// spanContextKey 用于在 context 中存储 Span 的键类型
type spanContextKey struct{}

// _spanContextKey 用于在 context 中存储 Span 的键实例
var _spanContextKey = spanContextKey{}

// SpanFromContext 从上下文中获取 Span
// 如果上下文中没有 Span，返回 nil
func SpanFromContext(ctx context.Context) Span {
	if span, ok := ctx.Value(_spanContextKey).(Span); ok {
		return span
	}
	return nil
}

// SetSpanToContext 将 Span 设置到上下文中
func SetSpanToContext(ctx context.Context, span Span) context.Context {
	return context.WithValue(ctx, _spanContextKey, span)
}

var defaultTracerProvider TracerProvider = &NoopTracerProvider{}

// SetTracerProvider 设置全局 TracerProvider
func SetTracerProvider(provider TracerProvider) {
	defaultTracerProvider = provider
}

// GetTracerProvider 获取全局 TracerProvider
func GetTracerProvider() TracerProvider {
	return defaultTracerProvider
}

// GetTracer 获取指定名称的 Tracer
func GetTracer(name string) Tracer {
	return defaultTracerProvider.Tracer(name)
}
