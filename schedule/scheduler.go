// Package schedule 定时任务调度框架
//
// 提供 6 字段 Spring 风格 Cron 表达式解析（CronExpression）、
// 任务调度器（Scheduler）、任务接口（Task）以及 @Scheduled 注解扫描。
package schedule

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"
)

// Scheduler 定时任务调度器接口
type Scheduler interface {
	// Start 启动调度器，开始触发定时任务
	Start(ctx context.Context) error
	// Shutdown 优雅关闭，等待正在执行的任务完成
	Shutdown(ctx context.Context) error
	// Register 注册定时任务，任务名唯一，重复返回 error
	Register(task Task) error
	// Unregister 注销定时任务
	Unregister(name string) bool
	// IsRunning 返回调度器是否正在运行
	IsRunning() bool
	// RegisteredTasks 返回所有已注册任务
	RegisteredTasks() []Task
}

type schedulerOptions struct {
	poolSize     int
	errorHandler func(task Task, err error)
}

// SchedulerOption 调度器选项函数
type SchedulerOption func(*schedulerOptions)

// WithPoolSize 设置任务执行池大小（最大并发数，默认 10）
func WithPoolSize(n int) SchedulerOption {
	return func(o *schedulerOptions) {
		if n > 0 {
			o.poolSize = n
		}
	}
}

// WithErrorHandler 设置执行错误处理函数
func WithErrorHandler(h func(task Task, err error)) SchedulerOption {
	return func(o *schedulerOptions) {
		o.errorHandler = h
	}
}

// scheduledTask 已注册的定时任务，包含下次执行时间
type scheduledTask struct {
	task Task
	expr *CronExpression
	next time.Time
	idx  int
}

// taskHeap 基于最小堆的 scheduledTask 优先队列，按下次执行时间排序
type taskHeap []*scheduledTask

func (h taskHeap) Len() int           { return len(h) }
func (h taskHeap) Less(i, j int) bool { return h[i].next.Before(h[j].next) }
func (h taskHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i]; h[i].idx = i; h[j].idx = j }
func (h *taskHeap) Push(x any) {
	item, ok := x.(*scheduledTask)
	if !ok {
		panic(fmt.Sprintf("heap: expected *scheduledTask, got %T", x))
	}
	n := len(*h)
	item.idx = n
	*h = append(*h, item)
}
func (h *taskHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.idx = -1
	*h = old[:n-1]
	return item
}

type schedulerImpl struct {
	tasks   map[string]*scheduledTask
	heap    taskHeap
	timer   *time.Timer
	wakeCh  chan struct{}
	ctx     context.Context
	cancel  context.CancelFunc
	running bool
	mu      sync.RWMutex
	wg      sync.WaitGroup
	sem     chan struct{}
	opts    schedulerOptions
}

// NewScheduler 创建调度器
func NewScheduler(opts ...SchedulerOption) Scheduler {
	o := schedulerOptions{poolSize: 10}
	for _, opt := range opts {
		opt(&o)
	}
	if o.errorHandler == nil {
		o.errorHandler = func(task Task, err error) {}
	}
	return &schedulerImpl{
		tasks:  make(map[string]*scheduledTask),
		heap:   make(taskHeap, 0),
		wakeCh: make(chan struct{}, 1),
		sem:    make(chan struct{}, o.poolSize),
		opts:   o,
	}
}

// Start 启动调度器，创建内部上下文并启动主循环
func (s *schedulerImpl) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return fmt.Errorf("schedule: scheduler already running")
	}
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.running = true
	s.timer = time.NewTimer(0)
	if !s.timer.Stop() {
		<-s.timer.C
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.loop()
	}()
	return nil
}

// Shutdown 优雅关闭调度器
//
// 先取消内部上下文阻止新任务派发，再等待所有执行中的任务完成。
// 如果 ctx 超时，返回 ctx.Err() 但调度器仍会继续等待后台任务。
func (s *schedulerImpl) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	s.mu.Unlock()

	s.cancel()
	s.timer.Stop()
	s.wake()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		go func() {
			<-done
		}()
		return ctx.Err()
	}
}

// Register 注册定时任务
//
// 任务名必须唯一。如果调度器已运行，注册后立即唤醒主循环重新计算最近执行时间。
func (s *schedulerImpl) Register(task Task) error {
	expr, err := Parse(task.Cron())
	if err != nil {
		return err
	}

	next := expr.Next(time.Now())
	if next.IsZero() {
		return fmt.Errorf("schedule: task %q will never fire (cron: %s)", task.Name(), task.Cron())
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[task.Name()]; exists {
		return fmt.Errorf("schedule: task %q already registered", task.Name())
	}

	st := &scheduledTask{
		task: task,
		expr: expr,
		next: next,
		idx:  -1,
	}
	s.tasks[task.Name()] = st
	heap.Push(&s.heap, st)

	if s.running && len(s.heap) > 0 && s.heap[0] == st {
		s.wake()
	}

	return nil
}

// Unregister 注销定时任务，返回任务是否存在
//
// 如果移除的是堆顶任务且调度器正在运行，会唤醒主循环重新计算。
func (s *schedulerImpl) Unregister(name string) bool {
	s.mu.Lock()
	st, exists := s.tasks[name]
	if !exists {
		s.mu.Unlock()
		return false
	}
	delete(s.tasks, name)
	wasTop := s.heap.Len() > 0 && s.heap[0] == st
	if st.idx >= 0 && st.idx < len(s.heap) {
		heap.Remove(&s.heap, st.idx)
	}
	running := s.running
	s.mu.Unlock()

	if running && (wasTop || s.heap.Len() == 0) {
		s.wake()
	}
	return true
}

// IsRunning 返回调度器是否正在运行
func (s *schedulerImpl) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// RegisteredTasks 返回所有已注册任务的快照
func (s *schedulerImpl) RegisteredTasks() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tasks := make([]Task, 0, len(s.tasks))
	for _, st := range s.tasks {
		tasks = append(tasks, st.task)
	}
	return tasks
}

// loop 调度器主循环
//
// 不断从堆顶取出最近需执行的任务，等待至执行时间后派发执行。
// 任务执行后重新计算下次执行时间并调整堆。
func (s *schedulerImpl) loop() {
	for {
		s.mu.Lock()
		if !s.running {
			s.mu.Unlock()
			return
		}
		nextTask := s.peekTask()
		if nextTask == nil {
			s.mu.Unlock()
			if !s.waitForWake() {
				return
			}
			continue
		}
		st := nextTask
		now := time.Now()
		if now.After(st.next) || now.Equal(st.next) {
			s.handleDueTask(st)
			s.mu.Unlock()
			s.executeTask(st.task)
			continue
		}
		s.mu.Unlock()
		if !s.waitForTimer(st.next.Sub(now)) {
			return
		}
	}
}

// peekTask 查看堆顶任务，堆空时返回 nil
func (s *schedulerImpl) peekTask() *scheduledTask {
	if s.heap.Len() == 0 {
		return nil
	}
	return s.heap[0]
}

// handleDueTask 处理到期任务：重新计算下次执行时间，如果永不到期则移除
func (s *schedulerImpl) handleDueTask(st *scheduledTask) {
	newNext := st.expr.Next(time.Now())
	if newNext.IsZero() {
		heap.Remove(&s.heap, 0)
		delete(s.tasks, st.task.Name())
	} else {
		st.next = newNext
		heap.Fix(&s.heap, 0)
	}
}

// waitForWake 等待唤醒信号或关闭信号
func (s *schedulerImpl) waitForWake() bool {
	select {
	case <-s.wakeCh:
		return true
	case <-s.ctx.Done():
		return false
	}
}

// waitForTimer 等待定时器到期、唤醒信号或关闭信号
func (s *schedulerImpl) waitForTimer(d time.Duration) bool {
	s.timer.Reset(d)
	select {
	case <-s.timer.C:
		return true
	case <-s.wakeCh:
		s.drainTimer()
		return true
	case <-s.ctx.Done():
		s.drainTimer()
		return false
	}
}

// drainTimer 安全排空定时器通道
func (s *schedulerImpl) drainTimer() {
	if !s.timer.Stop() {
		select {
		case <-s.timer.C:
		default:
		}
	}
}

// executeTask 异步执行任务，受并发池限制
//
// 先注册到 WaitGroup 再尝试获取信号量，确保 Shutdown 能正确等待。
// 如果上下文已取消则跳过执行，防止关闭期间启动新任务。
func (s *schedulerImpl) executeTask(task Task) {
	s.wg.Go(func() {
		select {
		case s.sem <- struct{}{}:
		case <-s.ctx.Done():
			return
		}
		defer func() { <-s.sem }()
		defer func() {
			if r := recover(); r != nil {
				var err error
				switch x := r.(type) {
				case error:
					err = x
				default:
					err = fmt.Errorf("schedule: panic: %v", x)
				}
				s.opts.errorHandler(task, err)
			}
		}()
		if err := task.Execute(s.ctx); err != nil {
			s.opts.errorHandler(task, err)
		}
	})
}

// wake 发送非阻塞唤醒信号给主循环
func (s *schedulerImpl) wake() {
	select {
	case s.wakeCh <- struct{}{}:
	default:
	}
}
