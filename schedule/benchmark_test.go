// Package schedule 性能基准测试
package schedule

import (
	"context"
	"testing"
	"time"
)

// BenchmarkScheduler_RegisterManyTasks 基准测试：注册大量任务的性能
func BenchmarkScheduler_RegisterManyTasks(b *testing.B) {
	s := NewScheduler()
	ctx := context.Background()

	// 预先创建多个任务
	tasks := make([]Task, 100)
	for i := 0; i < 100; i++ {
		task := NewTask(
			"task_"+string(rune('0'+i)),
			"* * * * * ?",
			func(ctx context.Context) error {
				return nil
			},
		)
		tasks[i] = task
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for _, task := range tasks {
			_ = s.Register(task)
		}
	}

	b.StopTimer()
	_ = s.Shutdown(ctx)
}

// BenchmarkScheduler_RegisterAndStart 基准测试：注册并启动调度器的性能
func BenchmarkScheduler_RegisterAndStart(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		s := NewScheduler()

		task := NewTask(
			"benchmark_task",
			"* * * * * ?",
			func(ctx context.Context) error {
				return nil
			},
		)

		_ = s.Register(task)
		_ = s.Start(ctx)

		_ = s.Shutdown(ctx)
	}
}

// BenchmarkCron_Parse 基准测试：解析cron表达式的性能
func BenchmarkCron_Parse(b *testing.B) {
	expressions := []string{
		"0 0/5 * * * ?",
		"0 0 3 * * ?",
		"0 0 3 * * 1-5",
		"0 0/30 8-17 * * ?",
		"0 0 12 1W * ?",
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for _, expr := range expressions {
			_, _ = Parse(expr)
		}
	}
}

// BenchmarkCron_Next 基准测试：计算下次执行时间的性能
func BenchmarkCron_Next(b *testing.B) {
	expressions := []string{
		"0 0/5 * * * ?",
		"0 0 3 * * ?",
		"0 0/30 8-17 * * ?",
	}

	parsedExprs := make([]*CronExpression, len(expressions))
	for i, expr := range expressions {
		parsedExprs[i], _ = Parse(expr)
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for _, expr := range parsedExprs {
			_ = expr.Next(time.Now())
		}
	}
}

// BenchmarkAnnotation_Scan 基准测试：注解扫描性能
func BenchmarkAnnotation_Scan(b *testing.B) {
	container := &MockContainer{
		beams: map[string]interface{}{
			"testService": &TestService{},
		},
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_, _ = ScanScheduledTasks(container, ".")
	}
}
