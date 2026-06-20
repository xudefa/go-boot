package event

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestEventBus_StressHighConcurrencyPublish 测试高并发发布事件
func TestEventBus_StressHighConcurrencyPublish(t *testing.T) {
	bus := NewEventBus()
	var called int32

	bus.Subscribe("stress.event", func(e ApplicationEvent) {
		atomic.AddInt32(&called, 1)
	})

	for i := 0; i < 10000; i++ {
		go func() {
			bus.Publish(&BaseEvent{
				EventType: "stress.event",
				EventTime: time.Now(),
			})
		}()
	}

	time.Sleep(200 * time.Millisecond)
	if atomic.LoadInt32(&called) != 10000 {
		t.Errorf("Expected 10000 calls, got %d", called)
	}
}

// TestEventBus_StressHighConcurrencySubscribe 测试高并发订阅
func TestEventBus_StressHighConcurrencySubscribe(t *testing.T) {
	bus := NewEventBus()
	var called int32
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			bus.Subscribe("stress.subscribe", func(e ApplicationEvent) {
				atomic.AddInt32(&called, 1)
			})
		}(i)
	}

	wg.Wait()

	bus.Publish(&BaseEvent{
		EventType: "stress.subscribe",
		EventTime: time.Now(),
	})

	time.Sleep(10 * time.Millisecond)
	if atomic.LoadInt32(&called) != 1000 {
		t.Errorf("Expected 1000 calls, got %d", called)
	}
}

// TestEventBus_StressMultipleEventTypes 测试多种事件类型的压力测试
func TestEventBus_StressMultipleEventTypes(t *testing.T) {
	bus := NewEventBus()
	var called int32

	for i := 0; i < 10; i++ {
		eventType := fmt.Sprintf("stress.event.%d", i)
		bus.Subscribe(eventType, func(e ApplicationEvent) {
			atomic.AddInt32(&called, 1)
		})
	}

	for i := 0; i < 10000; i++ {
		go func(id int) {
			eventType := fmt.Sprintf("stress.event.%d", id%10)
			bus.Publish(&BaseEvent{
				EventType: eventType,
				EventTime: time.Now(),
			})
		}(i)
	}

	time.Sleep(200 * time.Millisecond)
	if atomic.LoadInt32(&called) != 10000 {
		t.Errorf("Expected 10000 calls, got %d", called)
	}
}
