package event

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestEventBus_SubscribePublish(t *testing.T) {
	bus := NewEventBus()
	var called int32

	bus.Subscribe("test.event", func(e ApplicationEvent) {
		atomic.AddInt32(&called, 1)
	})

	bus.Publish(&BaseEvent{
		EventType: "test.event",
		EventTime: time.Now(),
	})

	if atomic.LoadInt32(&called) != 1 {
		t.Fatal("expected listener to be called once")
	}
}

func TestEventBus_MultipleListeners(t *testing.T) {
	bus := NewEventBus()
	var count int32

	for range 3 {
		bus.Subscribe("multi", func(e ApplicationEvent) {
			atomic.AddInt32(&count, 1)
		})
	}

	bus.Publish(&BaseEvent{EventType: "multi"})

	if atomic.LoadInt32(&count) != 3 {
		t.Fatalf("expected 3 calls, got %d", count)
	}
}

func TestEventBus_Unsubscribe(t *testing.T) {
	bus := NewEventBus()
	var called int32

	listener := func(e ApplicationEvent) {
		atomic.AddInt32(&called, 1)
	}

	bus.Subscribe("test", listener)
	bus.Publish(&BaseEvent{EventType: "test"})

	if atomic.LoadInt32(&called) != 1 {
		t.Fatal("expected listener to be called once before unsubscribe")
	}

	bus.Unsubscribe("test", listener)
	bus.Publish(&BaseEvent{EventType: "test"})

	if atomic.LoadInt32(&called) != 1 {
		t.Fatal("expected listener not to be called after unsubscribe")
	}
}

func TestEventBus_UnsubscribeNil(t *testing.T) {
	bus := NewEventBus()
	bus.Subscribe("test", func(e ApplicationEvent) {})

	bus.Unsubscribe("test", nil)
	bus.Unsubscribe("nonexistent", nil)
}

func TestEventBus_NoPanicOnUnsubscribedEvent(t *testing.T) {
	bus := NewEventBus()
	bus.Publish(&BaseEvent{EventType: "nonexistent"})
	// 不应该触发 panic
}

func TestBaseEvent_Timestamp(t *testing.T) {
	now := time.Now()
	e := &BaseEvent{EventType: "test", EventTime: now}

	if e.Type() != "test" {
		t.Fatalf("expected type test, got %s", e.Type())
	}
	if !e.Timestamp().Equal(now) {
		t.Fatal("timestamp mismatch")
	}

	// 零值时间戳应该自动填充
	e2 := &BaseEvent{EventType: "test2"}
	if e2.Timestamp().IsZero() {
		t.Fatal("expected auto-populated timestamp")
	}
}
