package schedule

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

// TestNewScheduler_Success 验证调度器创建成功且初始状态为未运行
func TestNewScheduler_Success(t *testing.T) {
	t.Parallel()
	s := NewScheduler()
	t.Cleanup(func() { _ = s.Shutdown(context.Background()) })
	if s == nil {
		t.Fatal("expected non-nil scheduler")
	}
	if s.IsRunning() {
		t.Fatal("expected scheduler to not be running")
	}
}

// TestScheduler_StartStop_Success 验证调度器的启动和停止功能正常
func TestScheduler_StartStop_Success(t *testing.T) {
	t.Parallel()
	s := NewScheduler()
	ctx := context.Background()
	t.Cleanup(func() { _ = s.Shutdown(ctx) })
	if err := s.Start(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !s.IsRunning() {
		t.Fatal("expected scheduler to be running")
	}
	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.IsRunning() {
		t.Fatal("expected scheduler to not be running")
	}
}

// TestScheduler_RegisterAndExecute 验证注册的任务能被调度执行
func TestScheduler_RegisterAndExecute(t *testing.T) {
	t.Parallel()
	var count atomic.Int32
	task := NewTask("test", "* * * * * ?", func(ctx context.Context) error {
		count.Add(1)
		return nil
	})

	s := NewScheduler()
	ctx := context.Background()
	t.Cleanup(func() { _ = s.Shutdown(ctx) })
	if err := s.Register(task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Start(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	time.Sleep(1200 * time.Millisecond)

	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count.Load() == 0 {
		t.Fatal("expected task to be executed at least once")
	}
}

// TestScheduler_RegisterDuplicate_Error 验证重复注册同一任务名会返回错误
func TestScheduler_RegisterDuplicate_Error(t *testing.T) {
	t.Parallel()
	s := NewScheduler()
	task := NewTask("test", "0/5 * * * * ?", func(ctx context.Context) error {
		return nil
	})
	if err := s.Register(task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Register(task); err == nil {
		t.Fatal("expected error for duplicate task")
	}
}

// TestScheduler_Unregister_Success 验证任务的注销功能正常，注销不存在的任务返回 false
func TestScheduler_Unregister_Success(t *testing.T) {
	t.Parallel()
	s := NewScheduler()
	t.Cleanup(func() { _ = s.Shutdown(context.Background()) })
	task := NewTask("test", "0/5 * * * * ?", func(ctx context.Context) error {
		return nil
	})
	if err := s.Register(task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !s.Unregister("test") {
		t.Fatal("expected unregister to return true")
	}
	if s.Unregister("test") {
		t.Fatal("expected unregister of nonexistent task to return false")
	}
}

// TestScheduler_RegisteredTasks 验证 RegisteredTasks 能正确返回已注册的任务列表
func TestScheduler_RegisteredTasks(t *testing.T) {
	t.Parallel()
	s := NewScheduler()
	t.Cleanup(func() { _ = s.Shutdown(context.Background()) })
	task1 := NewTask("t1", "0/5 * * * * ?", func(ctx context.Context) error { return nil })
	task2 := NewTask("t2", "0/10 * * * * ?", func(ctx context.Context) error { return nil })
	if err := s.Register(task1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Register(task2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tasks := s.RegisteredTasks()
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
}

// TestScheduler_PoolSizeLimit 验证并发池大小限制生效，最大并发数不超过设定值
func TestScheduler_PoolSizeLimit(t *testing.T) {
	t.Parallel()
	var concurrent atomic.Int32
	var maxConcurrent atomic.Int32
	ch := make(chan struct{})

	task := NewTask("test", "* * * * * ?", func(ctx context.Context) error {
		c := concurrent.Add(1)
		defer concurrent.Add(-1)
		if c > maxConcurrent.Load() {
			maxConcurrent.Store(c)
		}
		<-ch
		return nil
	})

	s := NewScheduler(WithPoolSize(2))
	ctx := context.Background()
	t.Cleanup(func() { _ = s.Shutdown(ctx) })
	if err := s.Register(task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Start(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	time.Sleep(1500 * time.Millisecond)

	close(ch)
	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if maxConcurrent.Load() > 2 {
		t.Errorf("expected max concurrency <= 2, got %d", maxConcurrent.Load())
	}
}

// TestScheduler_Shutdown_WaitsForExecution 验证关闭调度器时会等待正在执行的任务完成
func TestScheduler_Shutdown_WaitsForExecution(t *testing.T) {
	t.Parallel()
	var executed atomic.Bool
	ch := make(chan struct{})

	task := NewTask("test", "* * * * * ?", func(ctx context.Context) error {
		<-ch
		executed.Store(true)
		return nil
	})

	s := NewScheduler()
	ctx := context.Background()
	t.Cleanup(func() { _ = s.Shutdown(ctx) })
	if err := s.Register(task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Start(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	time.Sleep(1500 * time.Millisecond)

	go func() {
		time.Sleep(100 * time.Millisecond)
		close(ch)
	}()
	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !executed.Load() {
		t.Fatal("expected task to have completed before shutdown returns")
	}
}

// TestScheduler_ErrorHandler_Success 验证任务执行出错时能正确触发错误处理函数
func TestScheduler_ErrorHandler_Success(t *testing.T) {
	t.Parallel()
	var capturedErr error
	errCh := make(chan error, 1)

	task := NewTask("test", "* * * * * ?", func(ctx context.Context) error {
		return fmt.Errorf("task failed")
	})

	s := NewScheduler(WithErrorHandler(func(task Task, err error) {
		capturedErr = err
		errCh <- err
	}))
	ctx := context.Background()
	t.Cleanup(func() { _ = s.Shutdown(ctx) })
	if err := s.Register(task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Start(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case err := <-errCh:
		if err == nil || err.Error() != "task failed" {
			t.Errorf("expected 'task failed', got %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for error handler")
	}

	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedErr == nil {
		t.Fatal("expected error handler to be called")
	}
}

// TestScheduler_PanicRecovery_Success 验证任务执行中发生 panic 时能被正确恢复并通过错误处理函数捕获
func TestScheduler_PanicRecovery_Success(t *testing.T) {
	t.Parallel()
	var capturedErr error
	errCh := make(chan error, 1)

	task := NewTask("panic", "* * * * * ?", func(ctx context.Context) error {
		panic("oops")
	})

	s := NewScheduler(WithErrorHandler(func(task Task, err error) {
		capturedErr = err
		errCh <- err
	}))
	ctx := context.Background()
	t.Cleanup(func() { _ = s.Shutdown(ctx) })
	if err := s.Register(task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Start(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case err := <-errCh:
		if err == nil || err.Error() != "schedule: panic: oops" {
			t.Errorf("expected 'schedule: panic: oops', got %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for panic handler")
	}

	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedErr == nil {
		t.Fatal("expected error handler to be called on panic")
	}
}

// TestScheduler_DoubleStart_Error 验证重复启动调度器会返回错误
func TestScheduler_DoubleStart_Error(t *testing.T) {
	t.Parallel()
	s := NewScheduler()
	ctx := context.Background()
	t.Cleanup(func() { _ = s.Shutdown(ctx) })
	if err := s.Start(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Start(ctx); err == nil {
		t.Fatal("expected error on double start")
	}
	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestScheduler_ShutdownIdle_NoError 验证关闭未启动的调度器不会返回错误
func TestScheduler_ShutdownIdle_NoError(t *testing.T) {
	t.Parallel()
	s := NewScheduler()
	t.Cleanup(func() { _ = s.Shutdown(context.Background()) })
	if err := s.Shutdown(context.Background()); err != nil {
		t.Fatalf("expected no error when shutting down idle scheduler, got %v", err)
	}
}

// TestScheduler_RegisterAfterStart_Success 验证调度器启动后注册的任务能被正常调度
func TestScheduler_RegisterAfterStart_Success(t *testing.T) {
	t.Parallel()
	var count atomic.Int32
	s := NewScheduler()
	ctx := context.Background()
	t.Cleanup(func() { _ = s.Shutdown(ctx) })
	if err := s.Start(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	task := NewTask("late", "* * * * * ?", func(ctx context.Context) error {
		count.Add(1)
		return nil
	})
	if err := s.Register(task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	time.Sleep(1200 * time.Millisecond)
	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count.Load() == 0 {
		t.Fatal("expected task registered after start to execute")
	}
}

// TestScheduler_UnregisterRunningTask_Success 验证运行中注销任务后，任务不再被调度执行
func TestScheduler_UnregisterRunningTask_Success(t *testing.T) {
	t.Parallel()
	var count atomic.Int32
	task := NewTask("removable", "0/1 * * * * ?", func(ctx context.Context) error {
		count.Add(1)
		return nil
	})

	s := NewScheduler()
	ctx := context.Background()
	t.Cleanup(func() { _ = s.Shutdown(ctx) })
	if err := s.Register(task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Start(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	time.Sleep(500 * time.Millisecond)
	s.Unregister("removable")
	prev := count.Load()
	time.Sleep(1200 * time.Millisecond)

	if count.Load() != prev {
		t.Fatal("expected task to stop after unregister")
	}
	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
