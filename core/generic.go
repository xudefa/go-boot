package core

import (
	"encoding/json"
	"reflect"
)

type GenericFactory[T any] func() (T, error)

func ZeroOf[T any]() T {
	var zero T
	return zero
}

func TypeOfGeneric[T any]() reflect.Type {
	var t T
	return reflect.TypeOf(t)
}

func ValueOfGeneric[T any](t T) reflect.Value {
	return reflect.ValueOf(t)
}

func Clone[T any](src *T) *T {
	if src == nil {
		return nil
	}
	b, err := json.Marshal(src)
	if err != nil {
		return src
	}
	var dst T
	json.Unmarshal(b, &dst)
	return &dst
}
