package event

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestEventBusBuilder_Defaults(t *testing.T) {
	builder := NewEventBusBuilder()

	if builder.listeners == nil {
		t.Error("expected non-nil listeners")
	}
}

func TestEventBusBuilder_ChainConfig(t *testing.T) {
	called := false

	bus := NewEventBusBuilder().
		OnApplicationStarted(func(event ApplicationEvent) {
			called = true
		}).
		OnApplicationReady(func(event ApplicationEvent) {
			// ready handler
		}).
		OnApplicationStopped(func(event ApplicationEvent) {
			// stopped handler
		}).
		Build()

	if bus == nil {
		t.Fatal("expected non-nil bus")
	}

	// Publish application started event
	bus.Publish(NewBaseEventBuilder().Type(EventApplicationStarted).Now().Build())

	if !called {
		t.Error("expected application started listener to be called")
	}
}

func TestEventBusBuilder_CustomEventTypes(t *testing.T) {
	callCount := 0

	bus := NewEventBusBuilder().
		Subscribe("CustomEvent1", func(event ApplicationEvent) {
			callCount++
		}).
		Subscribe("CustomEvent2", func(event ApplicationEvent) {
			callCount++
		}).
		Build()

	if bus == nil {
		t.Fatal("expected non-nil bus")
	}

	bus.Publish(NewBaseEventBuilder().Type("CustomEvent1").Now().Build())
	bus.Publish(NewBaseEventBuilder().Type("CustomEvent2").Now().Build())

	if callCount != 2 {
		t.Errorf("expected callCount 2, got %d", callCount)
	}
}

func TestAsyncPublisherBuilder_Defaults(t *testing.T) {
	builder := NewAsyncPublisherBuilder()

	if builder.workerCount != 10 {
		t.Errorf("expected default workerCount 10, got %d", builder.workerCount)
	}
}

func TestAsyncPublisherBuilder_ChainConfig(t *testing.T) {
	bus := NewEventBus()

	publisher := NewAsyncPublisherBuilder().
		Bus(bus).
		WorkerCount(5).
		ErrorHandler(func(err error, event ApplicationEvent) {
			// error handler
		}).
		MustBuild()

	if publisher == nil {
		t.Fatal("expected non-nil publisher")
	}

	publisher.Close()
}

func TestAsyncPublisherBuilder_MissingBus(t *testing.T) {
	_, err := NewAsyncPublisherBuilder().Build()
	if err == nil {
		t.Error("expected error for missing bus")
	}
}

func TestAsyncPublisherBuilder_MustBuild_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	NewAsyncPublisherBuilder().MustBuild()
}

func TestBaseEventBuilder_BasicEvent(t *testing.T) {
	now := time.Now()

	event := NewBaseEventBuilder().
		Type("TestEvent").
		Timestamp(now).
		Build()

	if event.EventType != "TestEvent" {
		t.Errorf("expected EventType 'TestEvent', got %s", event.EventType)
	}

	if event.EventTime != now {
		t.Errorf("expected EventTime %v, got %v", now, event.EventTime)
	}
}

func TestBaseEventBuilder_Now(t *testing.T) {
	before := time.Now()

	event := NewBaseEventBuilder().
		Type("TestEvent").
		Now().
		Build()

	after := time.Now()

	if event.EventTime.Before(before) || event.EventTime.After(after) {
		t.Errorf("expected EventTime between %v and %v, got %v", before, after, event.EventTime)
	}
}

func TestBaseEventBuilder_Publish(t *testing.T) {
	called := false
	var receivedEvent ApplicationEvent

	bus := NewEventBus()
	bus.Subscribe("TestEvent", func(event ApplicationEvent) {
		called = true
		receivedEvent = event
	})

	NewBaseEventBuilder().
		Type("TestEvent").
		Now().
		Publish(bus)

	if !called {
		t.Error("expected listener to be called")
	}

	if receivedEvent == nil {
		t.Error("expected non-nil received event")
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test all helper functions create valid events
	events := []struct {
		name  string
		event *BaseEventBuilder
	}{
		{"ApplicationStarted", ApplicationStartedEvent()},
		{"ApplicationReady", ApplicationReadyEvent()},
		{"ApplicationStopped", ApplicationStoppedEvent()},
		{"ContextRefreshed", ContextRefreshedEvent()},
		{"EnvironmentPrepared", EnvironmentPreparedEvent()},
	}

	for _, tt := range events {
		t.Run(tt.name, func(t *testing.T) {
			event := tt.event.Build()

			if event.EventType == "" {
				t.Errorf("expected non-empty EventType for %s", tt.name)
			}

			if event.EventTime.IsZero() {
				t.Errorf("expected non-zero EventTime for %s", tt.name)
			}
		})
	}
}

func TestEventPublisher(t *testing.T) {
	bus := NewEventBus()
	publisher := NewEventPublisher(bus)

	called := false
	bus.Subscribe("TestEvent", func(event ApplicationEvent) {
		called = true
	})

	publisher.PublishEvent("TestEvent")

	if !called {
		t.Error("expected listener to be called")
	}
}

func TestEventPublisher_BuiltInEvents(t *testing.T) {
	bus := NewEventBus()
	publisher := NewEventPublisher(bus)

	startedCalled := false
	readyCalled := false
	stoppedCalled := false

	bus.Subscribe(EventApplicationStarted, func(event ApplicationEvent) {
		startedCalled = true
	})
	bus.Subscribe(EventApplicationReady, func(event ApplicationEvent) {
		readyCalled = true
	})
	bus.Subscribe(EventApplicationStopped, func(event ApplicationEvent) {
		stoppedCalled = true
	})

	publisher.PublishStarted()
	publisher.PublishReady()
	publisher.PublishStopped()

	if !startedCalled {
		t.Error("expected started listener to be called")
	}

	if !readyCalled {
		t.Error("expected ready listener to be called")
	}

	if !stoppedCalled {
		t.Error("expected stopped listener to be called")
	}
}

func TestAsyncEventPublisher(t *testing.T) {
	bus := NewEventBus()
	asyncPublisher := NewAsyncPublisherBuilder().
		Bus(bus).
		WorkerCount(5).
		MustBuild()

	asyncEventPublisher := NewAsyncEventPublisher(asyncPublisher)

	var wg sync.WaitGroup
	wg.Add(1)

	bus.Subscribe("AsyncTest", func(event ApplicationEvent) {
		wg.Done()
	})

	ctx := context.Background()
	asyncEventPublisher.PublishEvent(ctx, "AsyncTest")

	wg.Wait()
	asyncEventPublisher.Close()
}

func TestAsyncEventPublisher_Publish(t *testing.T) {
	bus := NewEventBus()
	asyncPublisher := NewAsyncPublisherBuilder().
		Bus(bus).
		MustBuild()

	asyncEventPublisher := NewAsyncEventPublisher(asyncPublisher)

	var wg sync.WaitGroup
	wg.Add(1)

	event := NewBaseEventBuilder().Type("DirectAsync").Now().Build()
	bus.Subscribe("DirectAsync", func(event ApplicationEvent) {
		wg.Done()
	})

	ctx := context.Background()
	asyncEventPublisher.Publish(ctx, event)

	wg.Wait()
	asyncEventPublisher.Close()
}

func TestEventBusBuilder_MultipleListenersSameEvent(t *testing.T) {
	callCount := 0

	bus := NewEventBusBuilder().
		Subscribe("MultiListener", func(event ApplicationEvent) {
			callCount++
		}).
		Subscribe("MultiListener", func(event ApplicationEvent) {
			callCount++
		}).
		Subscribe("MultiListener", func(event ApplicationEvent) {
			callCount++
		}).
		Build()

	bus.Publish(NewBaseEventBuilder().Type("MultiListener").Now().Build())

	if callCount != 3 {
		t.Errorf("expected callCount 3, got %d", callCount)
	}
}

func TestAsyncPublisherBuilder_AllOptions(t *testing.T) {
	bus := NewEventBus()

	errorHandlerCalled := false
	var capturedErr error
	var capturedEvent ApplicationEvent

	publisher := NewAsyncPublisherBuilder().
		Bus(bus).
		WorkerCount(20).
		ErrorHandler(func(err error, event ApplicationEvent) {
			errorHandlerCalled = true
			capturedErr = err
			capturedEvent = event
		}).
		MustBuild()

	if publisher == nil {
		t.Fatal("expected non-nil publisher")
	}

	publisher.Close()

	_ = errorHandlerCalled
	_ = capturedErr
	_ = capturedEvent
}

func TestBaseEventBuilder_EmptyEvent(t *testing.T) {
	event := NewBaseEventBuilder().Build()

	if event.EventType != "" {
		t.Errorf("expected empty EventType, got %s", event.EventType)
	}

	if !event.EventTime.IsZero() {
		t.Errorf("expected zero EventTime, got %v", event.EventTime)
	}
}

func TestEventPublisher_PublishDirectEvent(t *testing.T) {
	bus := NewEventBus()
	publisher := NewEventPublisher(bus)

	called := false
	var receivedEvent ApplicationEvent

	bus.Subscribe("DirectEvent", func(event ApplicationEvent) {
		called = true
		receivedEvent = event
	})

	event := NewBaseEventBuilder().Type("DirectEvent").Now().Build()
	publisher.Publish(event)

	if !called {
		t.Error("expected listener to be called")
	}

	if receivedEvent == nil {
		t.Error("expected non-nil received event")
	}

	if receivedEvent.Type() != "DirectEvent" {
		t.Errorf("expected event type 'DirectEvent', got %s", receivedEvent.Type())
	}
}
