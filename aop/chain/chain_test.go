package chain

import (
	"reflect"
	"testing"

	"github.com/xudefa/go-boot/aop/advice"
	"github.com/xudefa/go-boot/aop/pointcut"
)

type mockAdvice struct {
	called bool
}

func (m *mockAdvice) Invoke(invocation advice.MethodInvocation) (any, error) {
	m.called = true
	return invocation.Proceed()
}

type mockTarget struct{}

func (m *mockTarget) DoWork() string {
	return "work"
}

func TestAdviceChain_Proceed(t *testing.T) {
	advice1 := &mockAdvice{}
	advice2 := &mockAdvice{}

	advisors := []Advisor{
		{Pointcut: pointcut.ByName("Do.*"), Advice: advice1},
		{Pointcut: pointcut.ByName("Do.*"), Advice: advice2},
	}

	chain := NewAdviceChain(advisors)

	method, _ := reflect.TypeOf(&mockTarget{}).MethodByName("DoWork")
	invocation := chain.CreateInvocation(&mockTarget{}, method, []reflect.Value{reflect.ValueOf(&mockTarget{})})

	result, err := invocation.Proceed()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "work" {
		t.Errorf("expected 'work', got %v", result)
	}
	if !advice1.called {
		t.Error("expected first advice to be called")
	}
	if !advice2.called {
		t.Error("expected second advice to be called")
	}
}

func TestAdviceChain_Matches(t *testing.T) {
	pc := pointcut.ByName("Do.*")
	method, _ := reflect.TypeOf(&mockTarget{}).MethodByName("DoWork")

	if !pc.Matches(method) {
		t.Error("expected pointcut to match DoWork")
	}
}
