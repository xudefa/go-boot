package service

import (
	"testing"

	"github.com/xudefa/go-boot/core/container"
)

func TestServiceManager_RegisterAndGet(t *testing.T) {
	c := container.New()
	manager := NewServiceManager(c)

	svc := Service{}
	manager.Register("testService", svc)

	retrieved, exists := manager.Get("testService")
	if !exists {
		t.Fatal("expected service to exist")
	}

	if retrieved.Container != svc.Container {
		t.Error("expected retrieved service to match registered service")
	}
}

func TestServiceManager_List(t *testing.T) {
	c := container.New()
	manager := NewServiceManager(c)

	manager.Register("service1", Service{})
	manager.Register("service2", Service{})

	names := manager.List()
	if len(names) != 2 {
		t.Errorf("expected 2 services, got %d", len(names))
	}
}

func TestService_GetSetContainer(t *testing.T) {
	c := container.New()
	svc := Service{}

	svc.SetContainer(c)

	if svc.GetContainer() != c {
		t.Error("expected GetContainer to return set container")
	}
}
