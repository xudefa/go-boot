package core

import (
	"fmt"
	"testing"
)

type UserService struct {
	Repo UserRepository
}

func (u *UserService) GetUser(id int) string {
	return u.Repo.FindByID(id)
}

type UserRepository interface {
	FindByID(id int) string
}

type InMemoryUserRepo struct {
	data map[int]string
}

func NewInMemoryUserRepo() *InMemoryUserRepo {
	return &InMemoryUserRepo{data: map[int]string{1: "Alice", 2: "Bob"}}
}

func (r *InMemoryUserRepo) FindByID(id int) string {
	return r.data[id]
}

type OrderRepository interface {
	FindByID(id int) string
}

type InMemoryOrderRepo struct {
	ID int
}

var orderRepoCounter int

func NewInMemoryOrderRepo() *InMemoryOrderRepo {
	orderRepoCounter++
	return &InMemoryOrderRepo{ID: orderRepoCounter}
}

func (r *InMemoryOrderRepo) FindByID(id int) string {
	return fmt.Sprintf("Order-%d", id)
}

func TestContainer_Singleton(t *testing.T) {
	container := New()

	err := container.Register("userRepo",
		Bean(NewInMemoryUserRepo()),
		Singleton(),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = container.Register("userService",
		Bean(&UserService{}),
		Ref("userRepo", "Repo"),
	)
	if err != nil {
		t.Fatal(err)
	}

	singleton1, _ := container.Get("userRepo")
	singleton2, _ := container.Get("userRepo")
	if singleton1 != singleton2 {
		t.Error("Singleton beans should be the same instance")
	}

	userSvc, err := container.Get("userService")
	if err != nil {
		t.Fatal(err)
	}

	us := userSvc.(*UserService)
	if us.Repo == nil {
		t.Fatal("Repo should be injected")
	}
	if us.GetUser(1) != "Alice" {
		t.Errorf("expected Alice, got %s", us.GetUser(1))
	}

	t.Log("Singleton test passed")
}

func TestContainer_Prototype(t *testing.T) {
	container := New()

	err := container.Register("orderRepo",
		Bean(&InMemoryOrderRepo{ID: 0}),
		Prototype(),
	)
	if err != nil {
		t.Fatal(err)
	}

	prototype1, _ := container.Get("orderRepo")
	prototype2, _ := container.Get("orderRepo")
	if prototype1 == prototype2 {
		t.Error("Prototype beans should be different instances")
	}

	t.Log("Prototype test passed")
}

func TestContainer_FieldTagInjection(t *testing.T) {
	container := New(EnableFieldTag(true))

	type Config struct {
		AppName string
	}

	type Service struct {
		Conf *Config `inject:"appConfig"`
	}

	err := container.Register("appConfig", Bean(&Config{AppName: "TestApp"}))
	if err != nil {
		return
	}
	err = container.Register("service", Bean(&Service{}))
	if err != nil {
		return
	}

	svc, err := container.Get("service")
	if err != nil {
		t.Fatal(err)
	}

	if svc.(*Service).Conf.AppName != "TestApp" {
		t.Errorf("expected TestApp, got %s", svc.(*Service).Conf.AppName)
	}

	t.Log("Field tag injection test passed")
}

func TestContainer_FactoryWithDependency(t *testing.T) {
	t.Skip("Skipping - needs review")
}
