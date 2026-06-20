package tracing

import (
	"context"
	"fmt"
)

// TracerBuilder 追踪器构建器，支持链式配置
type TracerBuilder struct {
	serviceName    string
	serviceVersion string
	environment    string
	sampleRate     float64
	exporterType   string
	endpoint       string
	headers        map[string]string
	provider       TracerProvider
}

// NewTracerBuilder 创建追踪器构建器
func NewTracerBuilder() *TracerBuilder {
	return &TracerBuilder{
		sampleRate: 1.0,
		headers:    make(map[string]string),
	}
}

// ServiceName 设置服务名称
func (b *TracerBuilder) ServiceName(name string) *TracerBuilder {
	b.serviceName = name
	return b
}

// ServiceVersion 设置服务版本
func (b *TracerBuilder) ServiceVersion(version string) *TracerBuilder {
	b.serviceVersion = version
	return b
}

// Environment 设置环境
func (b *TracerBuilder) Environment(env string) *TracerBuilder {
	b.environment = env
	return b
}

// SampleRate 设置采样率（0.0-1.0）
func (b *TracerBuilder) SampleRate(rate float64) *TracerBuilder {
	if rate >= 0.0 && rate <= 1.0 {
		b.sampleRate = rate
	}
	return b
}

// ExporterType 设置导出器类型
func (b *TracerBuilder) ExporterType(typ string) *TracerBuilder {
	b.exporterType = typ
	return b
}

// Endpoint 设置导出器端点
func (b *TracerBuilder) Endpoint(endpoint string) *TracerBuilder {
	b.endpoint = endpoint
	return b
}

// Header 添加请求头
func (b *TracerBuilder) Header(key, value string) *TracerBuilder {
	b.headers[key] = value
	return b
}

// Provider 使用自定义提供者
func (b *TracerBuilder) Provider(p TracerProvider) *TracerBuilder {
	b.provider = p
	return b
}

// Build 构建追踪器
func (b *TracerBuilder) Build() (Tracer, error) {
	if b.provider != nil {
		return b.provider.Tracer(b.serviceName), nil
	}

	// 如果没有自定义提供者，返回 NoopTracer
	if b.serviceName == "" {
		return nil, fmt.Errorf("service name is required when using default provider")
	}

	return &NoopTracer{}, nil
}

// MustBuild 构建追踪器，失败则panic
func (b *TracerBuilder) MustBuild() Tracer {
	tracer, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build tracer: %v", err))
	}
	return tracer
}

// BuildAndRegister 构建并注册为全局提供者
func (b *TracerBuilder) BuildAndRegister() (Tracer, error) {
	tracer, err := b.Build()
	if err != nil {
		return nil, err
	}

	if b.provider != nil {
		SetTracerProvider(b.provider)
	}

	return tracer, nil
}

// SpanBuilder Span构建器，简化Span创建
type SpanBuilder struct {
	tracer     Tracer
	spanName   string
	attributes map[string]any
	kind       SpanKind
	parentCtx  context.Context
}

// NewSpanBuilder 创建Span构建器
func NewSpanBuilder(tracer Tracer, spanName string) *SpanBuilder {
	return &SpanBuilder{
		tracer:     tracer,
		spanName:   spanName,
		attributes: make(map[string]any),
		parentCtx:  context.Background(),
	}
}

// Context 设置父上下文
func (b *SpanBuilder) Context(ctx context.Context) *SpanBuilder {
	b.parentCtx = ctx
	return b
}

// Attribute 添加属性
func (b *SpanBuilder) Attribute(key string, value any) *SpanBuilder {
	b.attributes[key] = value
	return b
}

// Kind 设置Span类型
func (b *SpanBuilder) Kind(kind SpanKind) *SpanBuilder {
	b.kind = kind
	return b
}

// Start 启动Span
func (b *SpanBuilder) Start() (context.Context, Span) {
	opts := make([]SpanOption, 0)
	for k, v := range b.attributes {
		opts = append(opts, WithAttribute(k, v))
	}
	if b.kind != SpanKindUnspecified {
		opts = append(opts, WithSpanKind(b.kind))
	}

	ctx, span := b.tracer.Start(b.parentCtx, b.spanName, opts...)
	return SetSpanToContext(ctx, span), span
}

// TraceHelper 追踪辅助函数
type TraceHelper struct {
	tracer Tracer
}

// NewTraceHelper 创建追踪辅助函数
func NewTraceHelper(tracer Tracer) *TraceHelper {
	return &TraceHelper{tracer: tracer}
}

// Trace 执行带追踪的函数
func (h *TraceHelper) Trace(ctx context.Context, spanName string, fn func(ctx context.Context) error, opts ...SpanOption) error {
	ctx, span := h.tracer.Start(ctx, spanName, opts...)
	defer span.End()

	err := fn(ctx)
	if err != nil {
		span.SetError(err)
	}
	return err
}

// TraceResult 追踪结果
type TraceResult struct {
	Value any
	Err   error
}

// TraceWithCallback 执行带追踪和返回值的函数（使用回调方式）
func (h *TraceHelper) TraceWithCallback(ctx context.Context, spanName string, fn func(ctx context.Context) (any, error), opts ...SpanOption) TraceResult {
	ctx, span := h.tracer.Start(ctx, spanName, opts...)
	defer span.End()

	result, err := fn(ctx)
	if err != nil {
		span.SetError(err)
	}
	return TraceResult{Value: result, Err: err}
}
