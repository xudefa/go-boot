// Package actuator 提供 Actuator 自动配置。
//
// 通过 boot.RegisterAutoConfig 注册，由 boot.Application 在启动时自动执行。
// 仅当 actuator.enabled=true 时生效。
package actuator

import (
	"github.com/xudefa/go-boot/boot"
	"github.com/xudefa/go-boot/condition"
	"github.com/xudefa/go-boot/constants"
	"github.com/xudefa/go-boot/core"
	"github.com/xudefa/go-boot/health"
)

// ActuatorAutoConfiguration Actuator 自动配置
type ActuatorAutoConfiguration struct{}

// Configure 创建 Actuator 实例并注册为 Bean
func (a *ActuatorAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	act := New(ctx)

	agg := health.NewAggregator()

	if indicators, err := ctx.Container().GetAll((*health.Indicator)(nil)); err == nil && len(indicators) > 0 {
		for _, ind := range indicators {
			if h, ok := ind.(health.Indicator); ok {
				agg.AddIndicator(h)
			}
		}
	}
	act.SetHealthAggregator(agg)

	return ctx.Register(constants.ActuatorBeanID, core.Bean(act))
}

func init() {
	boot.RegisterAutoConfig(
		&ActuatorAutoConfiguration{},
		condition.OnProperty(constants.ActuatorEnabled, constants.ConditionTrue),
	)
}
