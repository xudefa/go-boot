// Package schedule 定时任务调度框架
//
// 提供 6 字段 Spring 风格 Cron 表达式解析（CronExpression）、
// 任务调度器（Scheduler）、任务接口（Task）以及 @Scheduled 注解扫描。
package schedule

import "context"

// Task 定时任务接口
//
// 每个定时任务包含名称、Cron 表达式和执行逻辑。
// 通过 NewTask 创建基于函数的任务实例。
type Task interface {
	// Name 返回任务名称，用于注册和查找
	Name() string
	// Cron 返回任务的 cron 表达式（6字段 Spring 风格）
	Cron() string
	// Execute 执行任务逻辑
	Execute(ctx context.Context) error
}

type task struct {
	name string
	cron string
	fn   func(context.Context) error
}

func (t *task) Name() string                      { return t.name }
func (t *task) Cron() string                      { return t.cron }
func (t *task) Execute(ctx context.Context) error { return t.fn(ctx) }

// NewTask 创建基于函数的定时任务
//
//	name: 任务名称，必须唯一
//	cronExpr: 6字段 Spring 风格 Cron 表达式
//	fn: 任务执行函数
func NewTask(name, cronExpr string, fn func(context.Context) error) Task {
	return &task{
		name: name,
		cron: cronExpr,
		fn:   fn,
	}
}
