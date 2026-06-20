package aop

import (
	"testing"
)

func TestBefore(t *testing.T) {
	t.Parallel()
	called := false

	advice := Before(func(jp JoinPoint) {
		called = true
	})

	if advice.Type() != AdviceBefore {
		t.Error("Before advice type should be 'before'")
	}

	inv := &invocation{
		method: nil,
		args:   nil,
		target: nil,
		sig:    &methodSignature{name: "Test"},
	}
	advice.Apply(inv, nil)

	if !called {
		t.Error("Before advice should have been called")
	}
}

func TestAfter(t *testing.T) {
	t.Parallel()
	called := false

	advice := After(func(jp JoinPoint) {
		called = true
	})

	if advice.Type() != AdviceAfter {
		t.Error("After advice type should be 'after'")
	}

	inv := &invocation{
		method: nil,
		args:   nil,
		target: nil,
		sig:    &methodSignature{name: "Test"},
	}
	advice.Apply(inv, nil)

	if !called {
		t.Error("After advice should have been called")
	}
}

func TestAfterReturning(t *testing.T) {
	t.Parallel()
	var receivedResult any
	expectedResult := "test result"

	advice := AfterReturning(func(jp JoinPoint, result any) {
		receivedResult = result
	})

	if advice.Type() != AdviceAfterReturning {
		t.Error("AfterReturning advice type should be 'after_returning'")
	}

	targetFunc := func(args ...any) any {
		return expectedResult
	}

	inv := &invocation{
		method: nil,
		args:   nil,
		target: nil,
		sig:    &methodSignature{name: "Test"},
	}
	result := advice.Apply(inv, targetFunc)

	if receivedResult != expectedResult {
		t.Errorf("expected result %v, got %v", expectedResult, receivedResult)
	}

	if result != expectedResult {
		t.Errorf("expected return result %v, got %v", expectedResult, result)
	}
}

func TestAfterThrowing(t *testing.T) {
	t.Parallel()
	var receivedError error
	testError := testError{"test error"}

	advice := AfterThrowing(func(jp JoinPoint, err error) {
		receivedError = err
	})

	if advice.Type() != AdviceAfterThrowing {
		t.Error("AfterThrowing advice type should be 'after_throwing'")
	}

	targetFunc := func(args ...any) any {
		return testError
	}

	inv := &invocation{
		method: nil,
		args:   nil,
		target: nil,
		sig:    &methodSignature{name: "Test"},
	}
	advice.Apply(inv, targetFunc)

	if receivedError != testError {
		t.Errorf("expected error %v, got %v", testError, receivedError)
	}
}

func TestAround(t *testing.T) {
	t.Parallel()
	var beforeCalled, afterCalled bool
	expectedResult := "test result"

	advice := Around(func(jp JoinPoint, proceed ProceedFunc) any {
		beforeCalled = true
		result := proceed(jp.Args()...)
		afterCalled = true
		return result
	})

	if advice.Type() != AdviceAround {
		t.Error("Around advice type should be 'around'")
	}

	targetFunc := func(args ...any) any {
		return expectedResult
	}

	inv := &invocation{
		method: nil,
		args:   nil,
		target: nil,
		sig:    &methodSignature{name: "Test"},
	}
	result := advice.Apply(inv, targetFunc)

	if !beforeCalled {
		t.Error("Before part of Around advice should have been called")
	}

	if !afterCalled {
		t.Error("After part of Around advice should have been called")
	}

	if result != expectedResult {
		t.Errorf("expected result %v, got %v", expectedResult, result)
	}
}

func TestAroundWithArgs(t *testing.T) {
	t.Parallel()
	var passedArgs []any

	advice := Around(func(jp JoinPoint, proceed ProceedFunc) any {
		return proceed("arg1", 42)
	})

	targetFunc := func(args ...any) any {
		passedArgs = args
		return nil
	}

	inv := &invocation{
		method: nil,
		args:   nil,
		target: nil,
		sig:    &methodSignature{name: "Test"},
	}
	advice.Apply(inv, targetFunc)

	if len(passedArgs) != 2 || passedArgs[0] != "arg1" || passedArgs[1] != 42 {
		t.Errorf("expected args [arg1, 42], got %v", passedArgs)
	}
}

func TestAfterReturning_NilProceed(t *testing.T) {
	t.Parallel()
	advice := AfterReturning(func(jp JoinPoint, result any) {
		if result != nil {
			t.Errorf("expected nil result in callback, got %v", result)
		}
	})

	inv := &invocation{
		method: nil,
		args:   nil,
		target: nil,
		sig:    &methodSignature{name: "Test"},
	}
	result := advice.Apply(inv, nil)

	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestAfterThrowing_NoError(t *testing.T) {
	t.Parallel()
	var receivedError error

	advice := AfterThrowing(func(jp JoinPoint, err error) {
		receivedError = err
	})

	targetFunc := func(args ...any) any {
		return "success"
	}

	inv := &invocation{
		method: nil,
		args:   nil,
		target: nil,
		sig:    &methodSignature{name: "Test"},
	}
	advice.Apply(inv, targetFunc)

	if receivedError != nil {
		t.Errorf("expected nil error, got %v", receivedError)
	}
}

func TestAfterThrowing_MultiResultWithError(t *testing.T) {
	t.Parallel()
	testErr := testError{"multi error"}
	var receivedError error

	advice := AfterThrowing(func(jp JoinPoint, err error) {
		receivedError = err
	})

	targetFunc := func(args ...any) any {
		return []any{"ok", testErr}
	}

	inv := &invocation{
		method: nil,
		args:   nil,
		target: nil,
		sig:    &methodSignature{name: "Test"},
	}
	advice.Apply(inv, targetFunc)

	if receivedError != testErr {
		t.Errorf("expected error %v, got %v", testErr, receivedError)
	}
}

func TestAfterThrowing_WithNilProceed(t *testing.T) {
	t.Parallel()
	var receivedError error

	advice := AfterThrowing(func(jp JoinPoint, err error) {
		receivedError = err
	})

	inv := &invocation{
		method: nil,
		args:   nil,
		target: nil,
		sig:    &methodSignature{name: "Test"},
	}
	advice.Apply(inv, nil)

	if receivedError != nil {
		t.Errorf("expected nil error with nil proceed, got %v", receivedError)
	}
}

func TestAround_NilProceed(t *testing.T) {
	t.Parallel()
	var aroundCalled bool

	advice := Around(func(jp JoinPoint, proceed ProceedFunc) any {
		aroundCalled = true
		return nil
	})

	inv := &invocation{
		method: nil,
		args:   nil,
		target: nil,
		sig:    &methodSignature{name: "Test"},
	}
	result := advice.Apply(inv, nil)

	if !aroundCalled {
		t.Error("Around advice should have been called")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestAdviceType_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		adviceType AdviceType
		expected   string
	}{
		{AdviceBefore, "before"},
		{AdviceAfter, "after"},
		{AdviceAround, "around"},
		{AdviceAfterReturning, "after_returning"},
		{AdviceAfterThrowing, "after_throwing"},
	}

	for _, tt := range tests {
		if tt.adviceType != AdviceType(tt.expected) {
			t.Errorf("AdviceType(%q) mismatch", tt.expected)
		}
	}
}

type testError struct {
	msg string
}

func (e testError) Error() string {
	return e.msg
}
