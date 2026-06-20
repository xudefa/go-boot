package schedule

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TaskBuilder 任务构建器，支持链式配置
type TaskBuilder struct {
	name        string
	cron        string
	fn          func(context.Context) error
	description string
	retryCount  int
	retryDelay  time.Duration
	timeout     time.Duration
	enabled     bool
	onSuccess   func(ctx context.Context)
	onFailure   func(ctx context.Context, err error)
}

// NewTaskBuilder 创建任务构建器
func NewTaskBuilder() *TaskBuilder {
	return &TaskBuilder{
		retryCount: 0,
		retryDelay: 1 * time.Second,
		timeout:    0,
		enabled:    true,
	}
}

// Name 设置任务名称
func (b *TaskBuilder) Name(name string) *TaskBuilder {
	b.name = name
	return b
}

// Cron 设置Cron表达式
func (b *TaskBuilder) Cron(expr string) *TaskBuilder {
	b.cron = expr
	return b
}

// Func 设置任务执行函数
func (b *TaskBuilder) Func(fn func(context.Context) error) *TaskBuilder {
	b.fn = fn
	return b
}

// Description 设置任务描述
func (b *TaskBuilder) Description(desc string) *TaskBuilder {
	b.description = desc
	return b
}

// Retry 设置重试次数和延迟
func (b *TaskBuilder) Retry(count int, delay time.Duration) *TaskBuilder {
	b.retryCount = count
	b.retryDelay = delay
	return b
}

// Timeout 设置任务超时时间
func (b *TaskBuilder) Timeout(timeout time.Duration) *TaskBuilder {
	b.timeout = timeout
	return b
}

// Enable 启用任务（默认启用）
func (b *TaskBuilder) Enable() *TaskBuilder {
	b.enabled = true
	return b
}

// Disable 禁用任务
func (b *TaskBuilder) Disable() *TaskBuilder {
	b.enabled = false
	return b
}

// OnSuccess 设置成功回调
func (b *TaskBuilder) OnSuccess(fn func(ctx context.Context)) *TaskBuilder {
	b.onSuccess = fn
	return b
}

// OnFailure 设置失败回调
func (b *TaskBuilder) OnFailure(fn func(ctx context.Context, err error)) *TaskBuilder {
	b.onFailure = fn
	return b
}

// Build 构建任务
func (b *TaskBuilder) Build() (Task, error) {
	if b.name == "" {
		return nil, fmt.Errorf("task name is required")
	}
	if b.cron == "" {
		return nil, fmt.Errorf("task cron is required")
	}
	if b.fn == nil {
		return nil, fmt.Errorf("task function is required")
	}

	if !b.enabled {
		return &disabledTask{name: b.name, cron: b.cron}, nil
	}

	wrappedFn := b.wrapWithRetry(b.fn)
	if b.timeout > 0 {
		wrappedFn = b.wrapWithTimeout(wrappedFn)
	}
	if b.onSuccess != nil || b.onFailure != nil {
		wrappedFn = b.wrapWithCallbacks(wrappedFn)
	}

	return NewTask(b.name, b.cron, wrappedFn), nil
}

// MustBuild 构建任务，失败则panic
func (b *TaskBuilder) MustBuild() Task {
	task, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build task: %v", err))
	}
	return task
}

func (b *TaskBuilder) wrapWithRetry(fn func(context.Context) error) func(context.Context) error {
	return func(ctx context.Context) error {
		var lastErr error
		for i := 0; i <= b.retryCount; i++ {
			if i > 0 {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(b.retryDelay):
				}
			}
			lastErr = fn(ctx)
			if lastErr == nil {
				return nil
			}
		}
		return lastErr
	}
}

func (b *TaskBuilder) wrapWithTimeout(fn func(context.Context) error) func(context.Context) error {
	return func(ctx context.Context) error {
		timeoutCtx, cancel := context.WithTimeout(ctx, b.timeout)
		defer cancel()

		done := make(chan error, 1)
		go func() {
			done <- fn(timeoutCtx)
		}()

		select {
		case err := <-done:
			return err
		case <-timeoutCtx.Done():
			return fmt.Errorf("task %q timed out after %v", b.name, b.timeout)
		}
	}
}

func (b *TaskBuilder) wrapWithCallbacks(fn func(context.Context) error) func(context.Context) error {
	return func(ctx context.Context) error {
		err := fn(ctx)
		if err == nil && b.onSuccess != nil {
			b.onSuccess(ctx)
		} else if err != nil && b.onFailure != nil {
			b.onFailure(ctx, err)
		}
		return err
	}
}

// disabledTask 禁用的任务
type disabledTask struct {
	name string
	cron string
}

func (t *disabledTask) Name() string                      { return t.name }
func (t *disabledTask) Cron() string                      { return t.cron }
func (t *disabledTask) Execute(ctx context.Context) error { return nil }

// TaskChain 任务链，支持任务依赖和顺序执行
type TaskChain struct {
	tasks        []Task
	dependencies map[string][]string
	mu           sync.RWMutex
}

// NewTaskChain 创建任务链
func NewTaskChain() *TaskChain {
	return &TaskChain{
		tasks:        make([]Task, 0),
		dependencies: make(map[string][]string),
	}
}

// Add 添加任务到链中
func (c *TaskChain) Add(task Task) *TaskChain {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tasks = append(c.tasks, task)
	return c
}

// DependsOn 设置任务依赖
func (c *TaskChain) DependsOn(taskName string, dependsOn ...string) *TaskChain {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dependencies[taskName] = append(c.dependencies[taskName], dependsOn...)
	return c
}

// Tasks 获取所有任务
func (c *TaskChain) Tasks() []Task {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]Task, len(c.tasks))
	copy(result, c.tasks)
	return result
}

// Dependencies 获取任务依赖
func (c *TaskChain) Dependencies() map[string][]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string][]string)
	for k, v := range c.dependencies {
		result[k] = append([]string(nil), v...)
	}
	return result
}

// OrderedTasks 返回按依赖顺序排序的任务列表
func (c *TaskChain) OrderedTasks() ([]Task, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	taskMap := make(map[string]Task)
	for _, t := range c.tasks {
		taskMap[t.Name()] = t
	}

	visited := make(map[string]bool)
	inStack := make(map[string]bool)
	var ordered []Task

	var visit func(name string) error
	visit = func(name string) error {
		if inStack[name] {
			return fmt.Errorf("circular dependency detected for task %q", name)
		}
		if visited[name] {
			return nil
		}

		inStack[name] = true
		for _, dep := range c.dependencies[name] {
			if _, exists := taskMap[dep]; !exists {
				return fmt.Errorf("dependency %q not found for task %q", dep, name)
			}
			if err := visit(dep); err != nil {
				return err
			}
		}
		inStack[name] = false
		visited[name] = true

		if task, exists := taskMap[name]; exists {
			ordered = append(ordered, task)
		}
		return nil
	}

	for _, t := range c.tasks {
		if err := visit(t.Name()); err != nil {
			return nil, err
		}
	}

	return ordered, nil
}

// ScheduleBuilder 调度器构建器
type ScheduleBuilder struct {
	scheduler    Scheduler
	poolSize     int
	errorHandler func(task Task, err error)
}

// NewScheduleBuilder 创建调度器构建器
func NewScheduleBuilder() *ScheduleBuilder {
	return &ScheduleBuilder{
		poolSize: 10,
	}
}

// PoolSize 设置并发池大小
func (b *ScheduleBuilder) PoolSize(size int) *ScheduleBuilder {
	b.poolSize = size
	return b
}

// ErrorHandler 设置错误处理器
func (b *ScheduleBuilder) ErrorHandler(handler func(task Task, err error)) *ScheduleBuilder {
	b.errorHandler = handler
	return b
}

// Scheduler 使用自定义调度器实现
func (b *ScheduleBuilder) Scheduler(s Scheduler) *ScheduleBuilder {
	b.scheduler = s
	return b
}

// Build 构建调度器
func (b *ScheduleBuilder) Build() Scheduler {
	if b.scheduler != nil {
		return b.scheduler
	}

	opts := make([]SchedulerOption, 0)
	opts = append(opts, WithPoolSize(b.poolSize))
	if b.errorHandler != nil {
		opts = append(opts, WithErrorHandler(b.errorHandler))
	}

	return NewScheduler(opts...)
}

// BuildAndRegister 构建调度器并注册任务链
func (b *ScheduleBuilder) BuildAndRegister(chain *TaskChain) (Scheduler, error) {
	scheduler := b.Build()

	orderedTasks, err := chain.OrderedTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to order tasks: %w", err)
	}

	for _, task := range orderedTasks {
		if err := scheduler.Register(task); err != nil {
			return nil, fmt.Errorf("failed to register task %q: %w", task.Name(), err)
		}
	}

	return scheduler, nil
}
