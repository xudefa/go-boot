// Package metrics 提供 Metrics 自动配置。
//
// 通过 boot.RegisterAutoConfig 注册，由 boot.Application 在启动时自动执行。
// 仅当 metrics.enabled=true 时生效，注册 SimpleRegistry 为单例 Bean。
package metrics

import (
	"github.com/xudefa/go-boot/boot"
	"github.com/xudefa/go-boot/condition"
	"github.com/xudefa/go-boot/constants"
	"github.com/xudefa/go-boot/core"
)

// MetricsAutoConfiguration Metrics 自动配置
type MetricsAutoConfiguration struct{}

// Configure 注册 SimpleRegistry 为单例 Bean
func (m *MetricsAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	registry := NewSimpleRegistry()
	if err := ctx.Register(constants.MeterRegistryBeanID,
		core.Bean(registry),
		core.Singleton()); err != nil {
		return err
	}

	return nil
}

func init() {
	boot.RegisterAutoConfig(
		&MetricsAutoConfiguration{},
		condition.OnProperty(constants.MetricsEnabled, constants.ConditionTrue),
	)
}
