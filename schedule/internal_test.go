// Package schedule 定时任务调度框架的内部功能测试
package schedule

import (
	"container/heap"
	"context"
	"testing"
	"time"
)

// TestScheduler_InternalMethods 主要测试内部方法的功能
func TestScheduler_InternalMethods(t *testing.T) {
	// 测试 heap 的 Swap 方法
	heap := make(taskHeap, 2)
	task1 := &scheduledTask{next: time.Now()}
	task2 := &scheduledTask{next: time.Now().Add(1 * time.Hour)}

	heap[0] = task1
	heap[1] = task2

	// 测试原始顺序
	if !heap.Less(0, 1) {
		t.Error("expected task1 to be less than task2 based on time")
	}

	// 手动交换
	heap.Swap(0, 1)

	if heap[0] != task2 || heap[1] != task1 {
		t.Error("swap didn't work correctly")
	}

	if heap[0].idx != 0 || heap[1].idx != 1 {
		t.Error("indices not updated correctly after swap")
	}
}

// TestScheduler_HandleDueTask_NeverFire 测试处理永远不会执行的任务
func TestScheduler_HandleDueTask_NeverFire(t *testing.T) {
	s := NewScheduler().(*schedulerImpl)

	task := NewTask("never", "0 0 0 30 2 ?", func(ctx context.Context) error {
		return nil
	})

	expr, err := Parse("0 0 0 30 2 ?") // Feb 30th never happens
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	next := expr.Next(time.Now())
	if !next.IsZero() {
		t.Skip("This expression unexpectedly has a next time, skipping test")
	}

	st := &scheduledTask{
		task: task,
		expr: expr,
		next: time.Now(),
		idx:  -1,
	}

	s.tasks[task.Name()] = st
	s.heap = append(s.heap, st)

	// Manually call handleDueTask to test the case where expr.Next returns zero time
	s.mu.Lock()
	s.handleDueTask(st)
	s.mu.Unlock()

	// The task should be removed from both tasks map and heap
	s.mu.RLock()
	_, exists := s.tasks[task.Name()]
	s.mu.RUnlock()

	if exists {
		t.Error("expected task to be removed from tasks map when it never fires")
	}
}

// TestScheduler_WaitForWake 测试等待唤醒功能
func TestScheduler_WaitForWake(t *testing.T) {
	s := NewScheduler().(*schedulerImpl)

	// Start the scheduler to initialize context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.mu.Lock()
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.running = true
	s.mu.Unlock()

	// Test wake functionality
	go func() {
		time.Sleep(100 * time.Millisecond)
		s.wake()
	}()

	done := make(chan bool, 1)
	go func() {
		result := s.waitForWake()
		done <- result
	}()

	select {
	case result := <-done:
		if !result {
			t.Error("expected waitForWake to return true when woken up")
		}
	case <-time.After(2 * time.Second):
		t.Error("waitForWake didn't return in time")
	}
}

// TestScheduler_WaitForTimer_Timeout 测试定时器超时功能
func TestScheduler_WaitForTimer_Timeout(t *testing.T) {
	s := NewScheduler().(*schedulerImpl)

	// Start the scheduler to initialize context and timer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := s.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start scheduler: %v", err)
	}
	defer func() {
		_ = s.Shutdown(context.Background())
	}()

	done := make(chan bool, 1)
	go func() {
		result := s.waitForTimer(100 * time.Millisecond) // Short timeout
		done <- result
	}()

	select {
	case result := <-done:
		if !result {
			t.Error("expected waitForTimer to return true after timeout")
		}
	case <-time.After(2 * time.Second):
		t.Error("waitForTimer didn't return in time")
	}
}

// TestScheduler_WaitForTimer_Wake 测试定时器被提前唤醒功能
func TestScheduler_WaitForTimer_Wake(t *testing.T) {
	s := NewScheduler().(*schedulerImpl)

	s.mu.Lock()
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.running = true
	s.timer = time.NewTimer(0)
	if !s.timer.Stop() {
		<-s.timer.C
	}
	s.mu.Unlock()

	defer func() {
		s.cancel()
		s.running = false
	}()

	done := make(chan bool, 1)
	go func() {
		result := s.waitForTimer(5 * time.Second) // Long timeout
		done <- result
	}()

	// Wake up the timer early
	time.Sleep(100 * time.Millisecond)
	s.wake()

	select {
	case result := <-done:
		if !result {
			t.Error("expected waitForTimer to return true when woken up")
		}
	case <-time.After(2 * time.Second):
		t.Error("waitForTimer didn't return in time after wake")
	}
}

// TestScheduler_CancelContextDuringWait 测试在等待时取消上下文
func TestScheduler_CancelContextDuringWait(t *testing.T) {
	s := NewScheduler().(*schedulerImpl)

	// Start the scheduler to initialize context and timer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := s.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start scheduler: %v", err)
	}
	defer func() {
		_ = s.Shutdown(context.Background())
	}()

	done := make(chan bool, 1)
	go func() {
		result := s.waitForTimer(5 * time.Second) // Long timeout
		done <- result
	}()

	// Cancel the context to interrupt the wait
	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case result := <-done:
		if result {
			t.Error("expected waitForTimer to return false when context cancelled")
		}
	case <-time.After(2 * time.Second):
		t.Error("waitForTimer didn't return in time after context cancel")
	}
}

// TestScheduler_PushPopHeap 测试堆的推送和弹出操作
func TestScheduler_PushPopHeap(t *testing.T) {
	heap := &taskHeap{}

	task := &scheduledTask{
		next: time.Now().Add(1 * time.Hour),
	}

	// Test push
	heap.Push(task)
	if len(*heap) != 1 {
		t.Error("push didn't add task to heap")
	}

	if (*heap)[0].idx != 0 {
		t.Error("index not set correctly after push")
	}

	// Test pop
	popped := heap.Pop()
	if popped != task {
		t.Error("pop didn't return the correct task")
	}

	if len(*heap) != 0 {
		t.Error("pop didn't remove task from heap")
	}

	if task.idx != -1 {
		t.Error("index not reset after pop")
	}
}

// TestScheduler_HeapOrdering 测试堆的排序功能
func TestScheduler_HeapOrdering(t *testing.T) {
	heap := &taskHeap{}

	now := time.Now()
	task1 := &scheduledTask{next: now.Add(3 * time.Hour)}
	task2 := &scheduledTask{next: now.Add(1 * time.Hour)}
	task3 := &scheduledTask{next: now.Add(2 * time.Hour)}

	heap.Push(task1)
	heap.Push(task2)
	heap.Push(task3)

	// Build heap order
	heap.init()

	// After heap initialization, the earliest task should be at index 0
	if (*heap)[0] != task2 { // task2 is earliest
		t.Error("heap ordering is incorrect - earliest task should be at top")
	}
}

// Helper to initialize heap
func (h *taskHeap) init() {
	heap.Init(h)
}
