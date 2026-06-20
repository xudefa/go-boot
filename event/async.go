package event

import (
	"context"
	"log/slog"
	"sync"
)

// AsyncPublisher 异步事件发布器
//
// 提供异步事件发布功能，支持上下文超时控制和错误处理。
// 使用工作协程池处理事件，避免阻塞发布者。
//
// 使用示例：
//
//	bus := event.NewEventBus()
//	publisher := event.NewAsyncPublisher(bus,
//	    event.WithWorkerCount(5),
//	    event.WithErrorHandler(func(err error, e event.ApplicationEvent) {
//	        log.Printf("event error: %v", err)
//	    }),
//	)
//	defer publisher.Close()
//
//	ctx := context.Background()
//	publisher.Publish(ctx, &event.BaseEvent{EventType: "MyEvent"})
type AsyncPublisher struct {
	bus        *EventBus
	worker     chan func()
	done       chan struct{}
	wg         sync.WaitGroup
	errHandler func(error, ApplicationEvent)
}

// AsyncPublisherOption 异步发布器选项函数
type AsyncPublisherOption func(*AsyncPublisher)

// WithWorkerCount 设置工作协程池大小
//
// 参数:
//   - n: 工作协程数量
//
// 返回:
//   - AsyncPublisherOption: 选项函数
func WithWorkerCount(n int) AsyncPublisherOption {
	return func(p *AsyncPublisher) {
		p.worker = make(chan func(), n)
	}
}

// WithErrorHandler 设置错误处理器
//
// 参数:
//   - handler: 错误处理函数，接收错误和事件作为参数
//
// 返回:
//   - AsyncPublisherOption: 选项函数
func WithErrorHandler(handler func(error, ApplicationEvent)) AsyncPublisherOption {
	return func(p *AsyncPublisher) {
		p.errHandler = handler
	}
}

// NewAsyncPublisher 创建异步事件发布器
//
// 参数:
//   - bus: 事件总线
//   - opts: 可选配置项
//
// 返回:
//   - *AsyncPublisher: 异步发布器实例
func NewAsyncPublisher(bus *EventBus, opts ...AsyncPublisherOption) *AsyncPublisher {
	p := &AsyncPublisher{
		bus:    bus,
		worker: make(chan func(), 10), // 默认缓冲 10
		done:   make(chan struct{}),
	}
	for _, opt := range opts {
		opt(p)
	}

	// 启动工作协程
	p.wg.Add(1)
	go p.run()

	return p
}

// run 工作协程主循环
func (p *AsyncPublisher) run() {
	defer p.wg.Done()
	for {
		select {
		case fn := <-p.worker:
			fn()
		case <-p.done:
			return
		}
	}
}

// Publish 异步发布事件
//
// 将事件发布到工作队列，由工作协程异步处理。
// 支持上下文超时控制，超时后调用错误处理器。
//
// 参数:
//   - ctx: 上下文，用于超时控制
//   - event: 要发布的事件
func (p *AsyncPublisher) Publish(ctx context.Context, event ApplicationEvent) {
	// 先检查上下文是否已经完成
	select {
	case <-ctx.Done():
		if p.errHandler != nil {
			p.errHandler(ctx.Err(), event)
		}
		return
	default:
	}

	p.wg.Add(1)
	select {
	case p.worker <- func() {
		defer p.wg.Done()
		p.publishEvent(event)
	}:
	case <-ctx.Done():
		p.wg.Done()
		if p.errHandler != nil {
			p.errHandler(ctx.Err(), event)
		}
	}
}

// publishEvent 发布单个事件，包含 panic 恢复逻辑
func (p *AsyncPublisher) publishEvent(event ApplicationEvent) {
	defer func() {
		if r := recover(); r != nil {
			if p.errHandler != nil {
				p.errHandler(nil, event)
			}
			slog.Error("event handler panic", "event", event.Type(), "recover", r)
		}
	}()
	p.bus.Publish(event)
}

// Close 关闭异步发布器
//
// 等待所有待处理的事件处理完成后返回。
func (p *AsyncPublisher) Close() {
	close(p.done)
	// 等待工作协程退出
	p.wg.Wait()
}
