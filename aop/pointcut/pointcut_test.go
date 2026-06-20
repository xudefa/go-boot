package pointcut

import (
	"reflect"
	"testing"
)

type TestService struct{}

func (s *TestService) GetUser() string {
	return "user"
}

func (s *TestService) SaveUser() error {
	return nil
}

func (s *TestService) DeleteUser() error {
	return nil
}

func TestByName(t *testing.T) {
	pc := ByName("Get.*")
	method, _ := reflect.TypeOf(&TestService{}).MethodByName("GetUser")

	if !pc.Matches(method) {
		t.Error("expected ByName to match GetUser")
	}

	pc2 := ByName("Save.*")
	if pc2.Matches(method) {
		t.Error("expected ByName to not match GetUser")
	}
}

func TestByPrefix(t *testing.T) {
	pc := ByPrefix("Get")
	method, _ := reflect.TypeOf(&TestService{}).MethodByName("GetUser")

	if !pc.Matches(method) {
		t.Error("expected ByPrefix to match GetUser")
	}

	pc2 := ByPrefix("Save")
	if pc2.Matches(method) {
		t.Error("expected ByPrefix to not match GetUser")
	}
}

func TestByRegex(t *testing.T) {
	pc, err := ByRegex("^Get.*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	method, _ := reflect.TypeOf(&TestService{}).MethodByName("GetUser")

	if !pc.Matches(method) {
		t.Error("expected ByRegex to match GetUser")
	}
}

func TestAll(t *testing.T) {
	pc := All(ByPrefix("Get"), ByName("Get.*"))
	method, _ := reflect.TypeOf(&TestService{}).MethodByName("GetUser")

	if !pc.Matches(method) {
		t.Error("expected All to match GetUser")
	}
}

func TestAny(t *testing.T) {
	pc := Any(ByPrefix("Get"), ByPrefix("Save"))
	method, _ := reflect.TypeOf(&TestService{}).MethodByName("SaveUser")

	if !pc.Matches(method) {
		t.Error("expected Any to match SaveUser")
	}
}
