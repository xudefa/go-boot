// Package tracing 提供本地追踪器实现。
//
// LocalTracer 是不依赖外部追踪系统的轻量级追踪器，
// 生成随机的 TraceID 和 SpanID，适用于不需要接入外部追踪系统的场景。
package tracing

import (
	"context"
	"crypto/rand"
	"fmt"
	mrand "math/rand/v2"
)

// LocalTracer 本地追踪器实现（不依赖外部追踪系统）
//
// 提供轻量级的追踪功能，生成随机的 TraceID 和 SpanID，
// 适用于不需要接入外部追踪系统的场景。
type LocalTracer struct {
	name string
}

// NewTracer 创建追踪器（便捷函数）
//
// 根据配置自动选择使用本地追踪器或 OpenTelemetry 追踪器。
// 当 OpenTelemetry 可用时优先使用，否则使用本地追踪器。
func NewTracer(name string) Tracer {
	// 检查是否已配置 OpenTelemetry
	if provider := GetTracerProvider(); provider != nil {
		if _, ok := provider.(*NoopTracerProvider); !ok {
			return provider.Tracer(name)
		}
	}
	// 默认使用本地追踪器
	return NewLocalTracer(name)
}

// NewLocalTracer 创建本地追踪器实例
func NewLocalTracer(name string) *LocalTracer {
	return &LocalTracer{name: name}
}

// Start 创建并启动一个新的 Span
//
// ctx: 父上下文，如果包含 Span 则作为父 Span
// spanName: Span 名称，通常是操作名称
// opts: Span 创建选项，如属性、类型等
//
// 返回新的上下文（包含新 Span）和 Span 实例
func (t *LocalTracer) Start(ctx context.Context, spanName string, opts ...SpanOption) (context.Context, Span) {
	// 生成随机的 TraceID 和 SpanID
	traceID := generateID(32)
	spanID := generateID(16)

	// 解析选项配置
	config := &SpanConfig{}
	for _, opt := range opts {
		opt(config)
	}

	// 创建 Span 实例
	span := &localSpan{
		traceID:    traceID,
		spanID:     spanID,
		name:       spanName,
		attributes: config.Attributes,
	}

	// 将 Span 注入到上下文中并返回
	return SetSpanToContext(ctx, span), span
}

// CurrentSpan 获取当前上下文中的 Span
//
// 如果上下文中没有 Span，返回 NoopSpan
func (t *LocalTracer) CurrentSpan(ctx context.Context) Span {
	span := SpanFromContext(ctx)
	if span != nil {
		return span
	}
	return &NoopSpan{}
}

// Finish 清理追踪器资源
//
// 本地追踪器无需特殊清理操作
func (t *LocalTracer) Finish() {}

// localSpan 本地 Span 实现
type localSpan struct {
	traceID    string
	spanID     string
	name       string
	attributes map[string]any
	statusCode SpanStatusCode
	errorMsg   string
}

// End 结束 Span
//
// 本地实现不执行任何操作，Span 创建后即视为完成
func (s *localSpan) End() {}

// AddEvent 添加事件到 Span
//
// 本地实现不执行任何操作
func (s *localSpan) AddEvent(name string, opts ...EventOption) {}

// SetError 设置错误状态
//
// 记录错误信息并设置状态码为 SpanStatusError
func (s *localSpan) SetError(err error) {
	if err != nil {
		s.errorMsg = err.Error()
		s.statusCode = SpanStatusError
		s.SetAttribute("error", err.Error())
	}
}

// GetTraceID 获取 TraceID
func (s *localSpan) GetTraceID() string {
	return s.traceID
}

// GetSpanID 获取 SpanID
func (s *localSpan) GetSpanID() string {
	return s.spanID
}

// SetAttribute 设置属性
//
// 如果属性 map 为 nil，先初始化
func (s *localSpan) SetAttribute(key string, value any) {
	if s.attributes == nil {
		s.attributes = make(map[string]any)
	}
	s.attributes[key] = value
}

// GetAttributes 获取所有属性
//
// 返回属性 map 的副本，避免外部修改影响内部状态
func (s *localSpan) GetAttributes() map[string]any {
	if s.attributes == nil {
		return make(map[string]any)
	}
	return s.attributes
}

// RecordError 记录错误
//
// 将错误信息添加为属性
func (s *localSpan) RecordError(err error) {
	s.SetAttribute("error", err.Error())
}

// SetStatus 设置状态码
func (s *localSpan) SetStatus(code SpanStatusCode) {
	s.statusCode = code
}

// SpanContext 获取 Span 上下文
func (s *localSpan) SpanContext() SpanContext {
	return SpanContext{
		TraceID:    s.traceID,
		SpanID:     s.spanID,
		TraceFlags: 0x01, // 采样标志位
	}
}

// generateID 生成指定长度的随机十六进制字符串
//
// length: 生成字符串的长度（字符数）
//
// 使用 crypto/rand 生成安全的随机数，如果失败则降级使用 math/rand
func generateID(length int) string {
	// 计算需要的字节数（每个十六进制字符占 4 位，即 0.5 字节）
	b := make([]byte, length/2)

	// 使用 crypto/rand 生成安全随机数
	_, err := rand.Read(b)
	if err != nil {
		// 降级使用 math/rand
		for i := range b {
			b[i] = byte(mrand.Uint64() >> ((i % 8) * 8))
		}
	}

	// 转换为十六进制字符串
	return fmt.Sprintf("%x", b)
}
