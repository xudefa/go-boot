package event

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestAsyncPublisher_Publish(t *testing.T) {
	bus := NewEventBus()
	publisher := NewAsyncPublisher(bus)
	defer publisher.Close()

	var mu sync.Mutex
	received := 0

	bus.Subscribe("TestEvent", func(e ApplicationEvent) {
		mu.Lock()
		defer mu.Unlock()
		received++
	})

	ctx := context.Background()
	publisher.Publish(ctx, &BaseEvent{EventType: "TestEvent"})

	// 等待异步处理
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	if received != 1 {
		t.Errorf("expected 1, got %d", received)
	}
	mu.Unlock()
}

func TestAsyncPublisher_MultipleEvents(t *testing.T) {
	bus := NewEventBus()
	publisher := NewAsyncPublisher(bus)
	defer publisher.Close()

	var mu sync.Mutex
	received := 0

	bus.Subscribe("MultiEvent", func(e ApplicationEvent) {
		mu.Lock()
		defer mu.Unlock()
		received++
	})

	ctx := context.Background()
	for i := 0; i < 10; i++ {
		publisher.Publish(ctx, &BaseEvent{EventType: "MultiEvent"})
	}

	// 等待异步处理
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	if received != 10 {
		t.Errorf("expected 10, got %d", received)
	}
	mu.Unlock()
}

func TestAsyncPublisher_ContextTimeout(t *testing.T) {
	bus := NewEventBus()
	var errHandled error
	var mu sync.Mutex

	publisher := NewAsyncPublisher(bus,
		WithErrorHandler(func(err error, e ApplicationEvent) {
			mu.Lock()
			defer mu.Unlock()
			errHandled = err
		}),
	)
	defer publisher.Close()

	// 创建一个已经取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	publisher.Publish(ctx, &BaseEvent{EventType: "TimeoutEvent"})

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if errHandled == nil {
		t.Error("expected context canceled error")
	}
}

func TestAsyncPublisher_PanicRecovery(t *testing.T) {
	bus := NewEventBus()
	var mu sync.Mutex
	panicHandled := false

	bus.Subscribe("PanicEvent", func(e ApplicationEvent) {
		panic("test panic")
	})

	publisher := NewAsyncPublisher(bus,
		WithErrorHandler(func(err error, e ApplicationEvent) {
			mu.Lock()
			panicHandled = true
			mu.Unlock()
		}),
	)
	defer publisher.Close()

	ctx := context.Background()
	publisher.Publish(ctx, &BaseEvent{EventType: "PanicEvent"})

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if !panicHandled {
		t.Error("expected panic to be handled")
	}
}

func TestAsyncPublisher_WorkerCount(t *testing.T) {
	bus := NewEventBus()
	publisher := NewAsyncPublisher(bus, WithWorkerCount(5))
	defer publisher.Close()

	var mu sync.Mutex
	received := 0

	bus.Subscribe("WorkerEvent", func(e ApplicationEvent) {
		mu.Lock()
		defer mu.Unlock()
		received++
	})

	ctx := context.Background()
	for i := 0; i < 20; i++ {
		publisher.Publish(ctx, &BaseEvent{EventType: "WorkerEvent"})
	}

	// 等待异步处理
	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	if received != 20 {
		t.Errorf("expected 20, got %d", received)
	}
	mu.Unlock()
}

func TestAsyncPublisher_Close(t *testing.T) {
	bus := NewEventBus()
	publisher := NewAsyncPublisher(bus)

	var mu sync.Mutex
	received := 0

	bus.Subscribe("CloseEvent", func(e ApplicationEvent) {
		mu.Lock()
		defer mu.Unlock()
		received++
	})

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		publisher.Publish(ctx, &BaseEvent{EventType: "CloseEvent"})
	}

	// 等待所有事件被处理
	time.Sleep(100 * time.Millisecond)

	// 关闭发布器
	publisher.Close()

	mu.Lock()
	if received != 5 {
		t.Errorf("expected 5, got %d", received)
	}
	mu.Unlock()
}

func TestAsyncPublisher_NoErrorHandler(t *testing.T) {
	bus := NewEventBus()
	// 不设置错误处理器
	publisher := NewAsyncPublisher(bus)
	defer publisher.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond)

	// 应该不会 panic
	publisher.Publish(ctx, &BaseEvent{EventType: "NoHandlerEvent"})
}
