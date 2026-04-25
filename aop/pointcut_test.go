package aop

import (
	"reflect"
	"testing"
)

func TestMatchAll(t *testing.T) {
	pc := MatchAll()

	if !pc.MatchClass(reflect.TypeOf((*testing.T)(nil))) {
		t.Error("MatchAll should match any class")
	}

	m := reflect.Method{}
	if !pc.MatchMethod(m) {
		t.Error("MatchAll should match any method")
	}
}

func TestMatchByName(t *testing.T) {
	pc := MatchByName("DoSomething")

	m := reflect.Method{Name: "DoSomething"}
	if !pc.MatchMethod(m) {
		t.Error("MatchByName should match DoSomething")
	}

	m2 := reflect.Method{Name: "DoAnother"}
	if pc.MatchMethod(m2) {
		t.Error("MatchByName should not match DoAnother")
	}
}

func TestMatchByNamePrefix(t *testing.T) {
	pc := MatchByNamePrefix("Do")

	m := reflect.Method{Name: "DoSomething"}
	if !pc.MatchMethod(m) {
		t.Error("MatchByNamePrefix should match methods with Do prefix")
	}

	m2 := reflect.Method{Name: "GetValue"}
	if pc.MatchMethod(m2) {
		t.Error("MatchByNamePrefix should not match methods without Do prefix")
	}
}

func TestMatchByRegex(t *testing.T) {
	pc := MatchByRegex("^Do.*")

	m := reflect.Method{Name: "DoSomething"}
	if !pc.MatchMethod(m) {
		t.Error("MatchByRegex should match methods matching regex")
	}

	m2 := reflect.Method{Name: "GetValue"}
	if pc.MatchMethod(m2) {
		t.Error("MatchByRegex should not match methods not matching regex")
	}
}

func TestMatchInterface(t *testing.T) {
	pc := MatchInterface((*TestInterfaceForMatch)(nil))

	if !pc.MatchClass(reflect.TypeOf(&TestImplForMatch{})) {
		t.Error("MatchInterface should match implementing struct pointer")
	}

	if pc.MatchClass(reflect.TypeOf("test")) {
		t.Error("MatchInterface should not match non-implementing type")
	}
}

type TestInterfaceForMatch interface {
	DoSomething()
}

type TestImplForMatch struct{}

var _ TestInterfaceForMatch = (*TestImplForMatch)(nil)

func (s TestImplForMatch) DoSomething() {}
