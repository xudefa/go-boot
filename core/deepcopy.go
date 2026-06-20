package core

import (
	"reflect"
)

// 不可复制的类型列表
var uncopyableTypes = map[string]bool{
	"sync.Mutex":     true,
	"sync.RWMutex":   true,
	"sync.Cond":      true,
	"sync.WaitGroup": true,
	"sync.Once":      true,
	"sync.Map":       true,
	"sync.Pool":      true,
}

// deepCopy 递归深拷贝任意值
//
// 支持: 指针、slice、map、结构体、基本类型
// 跳过: sync 包中的不可复制类型
func deepCopy(src reflect.Value) reflect.Value {
	if !src.IsValid() {
		return src
	}

	// 处理指针，解引用到实际值
	for src.Kind() == reflect.Ptr {
		if src.IsNil() {
			return src
		}
		src = src.Elem()
	}

	switch src.Kind() {
	case reflect.Slice:
		return deepCopySlice(src)
	case reflect.Map:
		return deepCopyMap(src)
	case reflect.Struct:
		return deepCopyStructValue(src)
	default:
		return src
	}
}

func deepCopySlice(src reflect.Value) reflect.Value {
	if src.IsNil() {
		return src
	}
	dst := reflect.MakeSlice(src.Type(), src.Len(), src.Cap())
	for i := 0; i < src.Len(); i++ {
		dst.Index(i).Set(deepCopy(src.Index(i)))
	}
	return dst
}

func deepCopyMap(src reflect.Value) reflect.Value {
	if src.IsNil() {
		return src
	}
	dst := reflect.MakeMap(src.Type())
	for _, key := range src.MapKeys() {
		dst.SetMapIndex(key, deepCopy(src.MapIndex(key)))
	}
	return dst
}

func deepCopyStructValue(src reflect.Value) reflect.Value {
	srcType := src.Type()
	typeKey := srcType.PkgPath() + "." + srcType.Name()

	if uncopyableTypes[typeKey] {
		return src
	}

	dst := reflect.New(srcType).Elem()
	for i := 0; i < src.NumField(); i++ {
		field := srcType.Field(i)
		if !field.IsExported() {
			continue
		}
		dst.Field(i).Set(deepCopy(src.Field(i)))
	}
	return dst
}

// IsUncopyable 检查类型是否不可复制
func IsUncopyable(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return false
	}
	typeKey := t.PkgPath() + "." + t.Name()
	return uncopyableTypes[typeKey]
}
