package event

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestEventBus_AsyncPublish 测试异步发布事件
func TestEventBus_AsyncPublish(t *testing.T) {
	bus := NewEventBus()
	var called int32

	bus.Subscribe("async.event", func(e ApplicationEvent) {
		atomic.AddInt32(&called, 1)
	})

	for i := 0; i < 100; i++ {
		go func() {
			bus.Publish(&BaseEvent{
				EventType: "async.event",
				EventTime: time.Now(),
			})
		}()
	}

	time.Sleep(100 * time.Millisecond)
	if atomic.LoadInt32(&called) != 100 {
		t.Errorf("Expected 100 calls, got %d", called)
	}
}

// TestEventBus_ConcurrentSubscribe 测试并发订阅
func TestEventBus_ConcurrentSubscribe(t *testing.T) {
	bus := NewEventBus()
	var called int32
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			bus.Subscribe("concurrent.event", func(e ApplicationEvent) {
				atomic.AddInt32(&called, 1)
			})
		}(i)
	}

	wg.Wait()

	bus.Publish(&BaseEvent{
		EventType: "concurrent.event",
		EventTime: time.Now(),
	})

	time.Sleep(10 * time.Millisecond)
	if atomic.LoadInt32(&called) != 100 {
		t.Errorf("Expected 100 calls, got %d", called)
	}
}

// TestEventBus_MixedAsyncOperations 测试混合异步操作
func TestEventBus_MixedAsyncOperations(t *testing.T) {
	bus := NewEventBus()
	var subscribeCount int32
	var publishCount int32
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			bus.Subscribe("mixed.event", func(e ApplicationEvent) {
				atomic.AddInt32(&subscribeCount, 1)
			})
		}(i)
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bus.Publish(&BaseEvent{
				EventType: "mixed.event",
				EventTime: time.Now(),
			})
			atomic.AddInt32(&publishCount, 1)
		}()
	}

	wg.Wait()
	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&publishCount) != 50 {
		t.Errorf("Expected 50 publishes, got %d", publishCount)
	}
}
