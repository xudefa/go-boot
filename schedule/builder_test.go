package schedule

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTaskBuilder_BasicTask(t *testing.T) {
	executed := false
	task, err := NewTaskBuilder().
		Name("test-task").
		Cron("0 0 0 * * *").
		Func(func(ctx context.Context) error {
			executed = true
			return nil
		}).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if task.Name() != "test-task" {
		t.Errorf("expected name 'test-task', got %s", task.Name())
	}

	if task.Cron() != "0 0 0 * * *" {
		t.Errorf("expected cron '0 0 0 * * *', got %s", task.Cron())
	}

	err = task.Execute(context.Background())
	if err != nil {
		t.Errorf("unexpected execute error: %v", err)
	}

	if !executed {
		t.Error("expected task to be executed")
	}
}

func TestTaskBuilder_MissingName(t *testing.T) {
	_, err := NewTaskBuilder().
		Cron("0 0 0 * * *").
		Func(func(ctx context.Context) error { return nil }).
		Build()

	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestTaskBuilder_MissingCron(t *testing.T) {
	_, err := NewTaskBuilder().
		Name("test").
		Func(func(ctx context.Context) error { return nil }).
		Build()

	if err == nil {
		t.Error("expected error for missing cron")
	}
}

func TestTaskBuilder_MissingFunc(t *testing.T) {
	_, err := NewTaskBuilder().
		Name("test").
		Cron("0 0 0 * * *").
		Build()

	if err == nil {
		t.Error("expected error for missing func")
	}
}

func TestTaskBuilder_DisabledTask(t *testing.T) {
	task, err := NewTaskBuilder().
		Name("disabled-task").
		Cron("0 0 0 * * *").
		Func(func(ctx context.Context) error {
			t.Error("disabled task should not execute")
			return nil
		}).
		Disable().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = task.Execute(context.Background())
	if err != nil {
		t.Errorf("unexpected execute error: %v", err)
	}
}

func TestTaskBuilder_WithRetry(t *testing.T) {
	callCount := 0
	task, err := NewTaskBuilder().
		Name("retry-task").
		Cron("0 0 0 * * *").
		Func(func(ctx context.Context) error {
			callCount++
			if callCount < 3 {
				return errors.New("temporary error")
			}
			return nil
		}).
		Retry(3, 10*time.Millisecond).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = task.Execute(context.Background())
	if err != nil {
		t.Errorf("unexpected execute error: %v", err)
	}

	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestTaskBuilder_WithTimeout(t *testing.T) {
	task, err := NewTaskBuilder().
		Name("timeout-task").
		Cron("0 0 0 * * *").
		Func(func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		}).
		Timeout(50 * time.Millisecond).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = task.Execute(context.Background())
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestTaskBuilder_OnSuccess(t *testing.T) {
	successCalled := false
	task, err := NewTaskBuilder().
		Name("success-task").
		Cron("0 0 0 * * *").
		Func(func(ctx context.Context) error {
			return nil
		}).
		OnSuccess(func(ctx context.Context) {
			successCalled = true
		}).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = task.Execute(context.Background())
	if err != nil {
		t.Errorf("unexpected execute error: %v", err)
	}

	if !successCalled {
		t.Error("expected onSuccess callback to be called")
	}
}

func TestTaskBuilder_OnFailure(t *testing.T) {
	failureCalled := false
	var failureErr error
	task, err := NewTaskBuilder().
		Name("failure-task").
		Cron("0 0 0 * * *").
		Func(func(ctx context.Context) error {
			return errors.New("test error")
		}).
		OnFailure(func(ctx context.Context, err error) {
			failureCalled = true
			failureErr = err
		}).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = task.Execute(context.Background())
	if err == nil {
		t.Error("expected task error")
	}

	if !failureCalled {
		t.Error("expected onFailure callback to be called")
	}

	if failureErr == nil {
		t.Error("expected failure error to be passed to callback")
	}
}

func TestTaskBuilder_MustBuild(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid task")
		}
	}()

	NewTaskBuilder().MustBuild()
}

func TestTaskBuilder_Description(t *testing.T) {
	task, err := NewTaskBuilder().
		Name("test").
		Cron("0 0 0 * * *").
		Func(func(ctx context.Context) error { return nil }).
		Description("Test task").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if task == nil {
		t.Error("expected non-nil task")
	}
}

func TestTaskChain_AddTasks(t *testing.T) {
	task1 := NewTask("task1", "0 0 0 * * *", func(ctx context.Context) error { return nil })
	task2 := NewTask("task2", "0 0 1 * * *", func(ctx context.Context) error { return nil })

	chain := NewTaskChain().
		Add(task1).
		Add(task2)

	tasks := chain.Tasks()
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestTaskChain_Dependencies(t *testing.T) {
	task1 := NewTask("task1", "0 0 0 * * *", func(ctx context.Context) error { return nil })
	task2 := NewTask("task2", "0 0 1 * * *", func(ctx context.Context) error { return nil })

	chain := NewTaskChain().
		Add(task1).
		Add(task2).
		DependsOn("task2", "task1")

	deps := chain.Dependencies()
	if len(deps["task2"]) != 1 || deps["task2"][0] != "task1" {
		t.Errorf("expected task2 to depend on task1, got %v", deps)
	}
}

func TestTaskChain_OrderedTasks_NoDeps(t *testing.T) {
	task1 := NewTask("task1", "0 0 0 * * *", func(ctx context.Context) error { return nil })
	task2 := NewTask("task2", "0 0 1 * * *", func(ctx context.Context) error { return nil })

	chain := NewTaskChain().
		Add(task1).
		Add(task2)

	ordered, err := chain.OrderedTasks()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ordered) != 2 {
		t.Errorf("expected 2 ordered tasks, got %d", len(ordered))
	}
}

func TestTaskChain_OrderedTasks_WithDeps(t *testing.T) {
	task1 := NewTask("task1", "0 0 0 * * *", func(ctx context.Context) error { return nil })
	task2 := NewTask("task2", "0 0 1 * * *", func(ctx context.Context) error { return nil })
	task3 := NewTask("task3", "0 0 2 * * *", func(ctx context.Context) error { return nil })

	chain := NewTaskChain().
		Add(task1).
		Add(task2).
		Add(task3).
		DependsOn("task2", "task1").
		DependsOn("task3", "task2")

	ordered, err := chain.OrderedTasks()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ordered) != 3 {
		t.Fatalf("expected 3 ordered tasks, got %d", len(ordered))
	}

	if ordered[0].Name() != "task1" {
		t.Errorf("expected first task to be task1, got %s", ordered[0].Name())
	}
	if ordered[1].Name() != "task2" {
		t.Errorf("expected second task to be task2, got %s", ordered[1].Name())
	}
	if ordered[2].Name() != "task3" {
		t.Errorf("expected third task to be task3, got %s", ordered[2].Name())
	}
}

func TestTaskChain_CircularDependency(t *testing.T) {
	task1 := NewTask("task1", "0 0 0 * * *", func(ctx context.Context) error { return nil })
	task2 := NewTask("task2", "0 0 1 * * *", func(ctx context.Context) error { return nil })

	chain := NewTaskChain().
		Add(task1).
		Add(task2).
		DependsOn("task1", "task2").
		DependsOn("task2", "task1")

	_, err := chain.OrderedTasks()
	if err == nil {
		t.Error("expected circular dependency error")
	}
}

func TestTaskChain_MissingDependency(t *testing.T) {
	task1 := NewTask("task1", "0 0 0 * * *", func(ctx context.Context) error { return nil })

	chain := NewTaskChain().
		Add(task1).
		DependsOn("task1", "nonexistent")

	_, err := chain.OrderedTasks()
	if err == nil {
		t.Error("expected missing dependency error")
	}
}

func TestScheduleBuilder_Default(t *testing.T) {
	scheduler := NewScheduleBuilder().Build()

	if scheduler == nil {
		t.Fatal("expected non-nil scheduler")
	}

	if scheduler.IsRunning() {
		t.Error("expected scheduler not to be running initially")
	}
}

func TestScheduleBuilder_WithPoolSize(t *testing.T) {
	scheduler := NewScheduleBuilder().
		PoolSize(5).
		Build()

	if scheduler == nil {
		t.Fatal("expected non-nil scheduler")
	}
}

func TestScheduleBuilder_WithErrorHandler(t *testing.T) {
	scheduler := NewScheduleBuilder().
		ErrorHandler(func(task Task, err error) {
			// error handler
		}).
		Build()

	if scheduler == nil {
		t.Fatal("expected non-nil scheduler")
	}
}

func TestScheduleBuilder_BuildAndRegister(t *testing.T) {
	task := NewTask("test-task", "0 0 0 * * *", func(ctx context.Context) error { return nil })

	chain := NewTaskChain().Add(task)

	scheduler, err := NewScheduleBuilder().
		BuildAndRegister(chain)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if scheduler == nil {
		t.Fatal("expected non-nil scheduler")
	}

	tasks := scheduler.RegisteredTasks()
	if len(tasks) != 1 {
		t.Errorf("expected 1 registered task, got %d", len(tasks))
	}
}

func TestScheduleBuilder_BuildAndRegister_InvalidOrder(t *testing.T) {
	task1 := NewTask("task1", "0 0 0 * * *", func(ctx context.Context) error { return nil })
	task2 := NewTask("task2", "0 0 1 * * *", func(ctx context.Context) error { return nil })

	chain := NewTaskChain().
		Add(task1).
		Add(task2).
		DependsOn("task1", "task2").
		DependsOn("task2", "task1")

	_, err := NewScheduleBuilder().BuildAndRegister(chain)
	if err == nil {
		t.Error("expected error for circular dependency")
	}
}
