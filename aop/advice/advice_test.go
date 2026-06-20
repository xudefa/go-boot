package advice

import (
	"reflect"
	"testing"
)

type testAdviceTarget struct{}

func (t *testAdviceTarget) GetValue() string {
	return "value"
}

func TestSimpleMethodInvocation_Proceed(t *testing.T) {
	target := &testAdviceTarget{}
	targetType := reflect.TypeOf(target)
	method, _ := targetType.MethodByName("GetValue")

	invocation := &SimpleMethodInvocation{
		Target: target,
		Method: method,
		Args:   []reflect.Value{reflect.ValueOf(target)},
	}

	result, err := invocation.Proceed()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "value" {
		t.Errorf("expected 'value', got %v", result)
	}
}

func TestSimpleMethodInvocation_GetMethod(t *testing.T) {
	target := &testAdviceTarget{}
	targetType := reflect.TypeOf(target)
	method, _ := targetType.MethodByName("GetValue")

	invocation := &SimpleMethodInvocation{
		Target: target,
		Method: method,
	}

	if invocation.GetMethod().Name != "GetValue" {
		t.Errorf("expected method name 'GetValue', got %s", invocation.GetMethod().Name)
	}
}

func TestSimpleMethodInvocation_GetTarget(t *testing.T) {
	target := &testAdviceTarget{}
	invocation := &SimpleMethodInvocation{
		Target: target,
	}

	if invocation.GetTarget() != target {
		t.Error("expected GetTarget to return original target")
	}
}
