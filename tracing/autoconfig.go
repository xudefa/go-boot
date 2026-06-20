// Package tracing 提供追踪自动配置。
//
// 当 tracing.enabled=true 且未指定 tracing.provider 时自动启用，
// 注册 NoopTracerProvider 和默认 Tracer 到 IoC 容器。
package tracing

import (
	"context"

	"github.com/xudefa/go-boot/boot"
	"github.com/xudefa/go-boot/condition"
	"github.com/xudefa/go-boot/constants"
	"github.com/xudefa/go-boot/core"
)

// TracingAutoConfiguration 追踪自动配置
//
// 当 tracing.enabled=true 且未指定 tracing.provider 时生效，
// 注册 NoopTracerProvider 和默认 Tracer 到 IoC 容器。
type TracingAutoConfiguration struct{}

// Configure 执行自动配置逻辑，注册 TracerProvider 和默认 Tracer
func (t *TracingAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	tracerProvider := &NoopTracerProvider{}
	if err := ctx.Register(constants.TracerProviderBeanID,
		core.Bean(tracerProvider),
		core.Singleton()); err != nil {
		return err
	}

	defaultTracer := tracerProvider.Tracer("default")
	if err := ctx.Register(constants.TracerBeanID,
		core.Bean(defaultTracer),
		core.Singleton()); err != nil {
		return err
	}

	return nil
}

// TracerProviderImpl 基于 LocalTracer 的 TracerProvider 实现
type TracerProviderImpl struct{}

// Tracer 创建指定名称的本地追踪器
func (p *TracerProviderImpl) Tracer(name string) Tracer {
	return NewLocalTracer(name)
}

// Shutdown 关闭 TracerProvider，释放资源
func (p *TracerProviderImpl) Shutdown(ctx context.Context) error {
	return nil
}

func init() {
	boot.RegisterAutoConfig(
		&TracingAutoConfiguration{},
		condition.All(
			condition.OnProperty(constants.TracingEnabled, constants.ConditionTrue),
			condition.OnMissingProperty(constants.TracingProvider),
		),
	)
}
