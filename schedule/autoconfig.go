// Package schedule 定时任务调度框架的自动配置与启动器
//
// 提供 ScheduleAutoConfiguration 根据配置文件创建 Scheduler bean，
// 以及 ScheduleStarter 负责调度器的生命周期管理（启动和优雅关闭）。
// Package schedule 提供定时任务调度器的自动配置。
//
// 当 schedule.enabled=true 时自动启用，从 Environment 中读取 schedule.pool-size、schedule.scan-annotations 等配置项，
// 创建并注册 Scheduler Bean 到 IoC 容器中（Bean ID: scheduleScheduler）。
// 同时注册 ScheduleStarter 启动器，负责在应用启动时开始调度、关闭时优雅停止。
package schedule

import (
	"context"
	"log/slog"

	"github.com/xudefa/go-boot/boot"
	"github.com/xudefa/go-boot/condition"
	"github.com/xudefa/go-boot/constants"
	"github.com/xudefa/go-boot/core"
)

func init() {
	boot.RegisterAutoConfig(&ScheduleAutoConfiguration{},
		condition.OnProperty(constants.ScheduleEnabled, constants.ConditionTrue),
	)
	boot.RegisterStarter(&ScheduleStarter{})
}

// ScheduleAutoConfiguration 定时任务自动配置
//
// 根据配置创建 Scheduler bean:
//   - schedule.pool-size: 并发池大小（默认 10）
//   - schedule.scan-annotations: 是否扫描 @Scheduled 注解（默认 true）
type ScheduleAutoConfiguration struct{}

func (s *ScheduleAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	poolSize := env.GetInt(constants.SchedulePoolSize, constants.DefaultSchedulePoolSize)
	scanAnnotations := env.GetBool(constants.ScheduleScanAnnotations, constants.DefaultScheduleScanAnnotations)

	scheduler := NewScheduler(WithPoolSize(poolSize))

	if scanAnnotations {
		tasks, err := ScanScheduledTasks(ctx.Container(), ".")
		if err != nil {
			slog.Warn("schedule: failed to scan scheduled annotations", "error", err)
		} else {
			for _, t := range tasks {
				if err := scheduler.Register(t); err != nil {
					slog.Warn("schedule: failed to register task", "task", t.Name(), "error", err)
				}
			}
		}
	}

	if err := ctx.Register(constants.ScheduleSchedulerBeanID, core.Bean(scheduler), core.Singleton()); err != nil {
		return err
	}

	return nil
}

// ScheduleStarter 调度器启动器，负责启动和停止 Scheduler
type ScheduleStarter struct{}

func (s *ScheduleStarter) Name() string {
	return "ScheduleStarter"
}

func (s *ScheduleStarter) Dependencies() []string {
	return nil
}

func (s *ScheduleStarter) Configure(ctx boot.ApplicationContext) error {
	return nil
}

func (s *ScheduleStarter) Start(ctx boot.ApplicationContext) error {
	scheduler, ok := resolveScheduler(ctx)
	if !ok {
		return nil
	}
	if scheduler.IsRunning() {
		return nil
	}
	return scheduler.Start(context.Background())
}

func (s *ScheduleStarter) Stop(ctx boot.ApplicationContext) error {
	scheduler, ok := resolveScheduler(ctx)
	if !ok {
		return nil
	}
	return scheduler.Shutdown(context.Background())
}

func (s *ScheduleStarter) GetCondition() condition.Condition {
	return condition.OnProperty(constants.ScheduleEnabled, constants.ConditionTrue)
}

func resolveScheduler(ctx boot.ApplicationContext) (Scheduler, bool) {
	bean, err := ctx.Get(constants.ScheduleSchedulerBeanID)
	if err != nil {
		return nil, false
	}
	scheduler, ok := bean.(Scheduler)
	return scheduler, ok
}
