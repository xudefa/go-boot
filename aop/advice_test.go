package aop

import (
	"testing"
)

func TestBefore(t *testing.T) {
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
	var receivedResult interface{}
	expectedResult := "test result"

	advice := AfterReturning(func(jp JoinPoint, result interface{}) {
		receivedResult = result
	})

	if advice.Type() != AdviceAfterReturning {
		t.Error("AfterReturning advice type should be 'after_returning'")
	}

	targetFunc := func(args ...interface{}) interface{} {
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
	var receivedError error
	testError := testError{"test error"}

	advice := AfterThrowing(func(jp JoinPoint, err error) {
		receivedError = err
	})

	if advice.Type() != AdviceAfterThrowing {
		t.Error("AfterThrowing advice type should be 'after_throwing'")
	}

	targetFunc := func(args ...interface{}) interface{} {
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
	var beforeCalled, afterCalled bool
	expectedResult := "test result"

	advice := Around(func(jp JoinPoint, proceed ProceedFunc) interface{} {
		beforeCalled = true
		result := proceed(jp.Args()...)
		afterCalled = true
		return result
	})

	if advice.Type() != AdviceAround {
		t.Error("Around advice type should be 'around'")
	}

	targetFunc := func(args ...interface{}) interface{} {
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
	var passedArgs []interface{}

	advice := Around(func(jp JoinPoint, proceed ProceedFunc) interface{} {
		return proceed("arg1", 42)
	})

	targetFunc := func(args ...interface{}) interface{} {
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

type testError struct {
	msg string
}

func (e testError) Error() string {
	return e.msg
}
