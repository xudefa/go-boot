// Package core 的泛型工具函数.
//
// 提供类型安全的泛型API,用于获取类型信息、零值和对象克隆.
//
// # 使用示例
//
//	// 获取类型的零值
//	var zero int = core.ZeroOf[int]() // zero == 0
//
//	// 获取反射类型
//	t := core.TypeOfGeneric[UserService]()
//
//	// 克隆对象
//	original := &User{Name: "Alice"}
//	cloned, _ := core.Clone(original)
//
// # 注意事项
//
//   - Clone 使用JSON序列化/反序列化实现,要求字段必须是可导出的
//   - 泛型函数在编译期进行类型检查,比反射更安全
package core

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// GenericFactory 泛型工厂函数类型
//
// 用于创建指定类型的实例，可以返回错误。
// 通常与 FactoryOf 配合使用。
type GenericFactory[T any] func() (T, error)

// ZeroOf 获取类型的零值
//
// 返回类型参数的零值.
// 对于指针类型是nil,对于结构体类型是各字段为零值的结构体.
//
// 类型参数:
//   - T: 任意类型
//
// 返回值:
//   - T: 该类型的零值
//
// 示例:
//
//	var s string = core.ZeroOf[string]()  // ""
//	var i int = core.ZeroOf[int]()        // 0
//	var p *T = core.ZeroOf[*T]()         // nil
func ZeroOf[T any]() T {
	var zero T
	return zero
}

// TypeOfGeneric 获取泛型参数的反射类型
//
// 类型参数:
//   - T: 任意类型
//
// 返回值:
//   - reflect.Type: 类型的反射表示
//
// 示例:
//
//	t := core.TypeOfGeneric[UserService]()
//	fmt.Println(t.Name()) // "UserService"
func TypeOfGeneric[T any]() reflect.Type {
	var t T
	return reflect.TypeOf(t)
}

// ValueOfGeneric 获取泛型值的反射值
//
// 参数:
//   - t: 泛型值
//
// 返回值:
//   - reflect.Value: 值的反射表示
func ValueOfGeneric[T any](t T) reflect.Value {
	return reflect.ValueOf(t)
}

// Clone 深克隆对象
//
// 使用JSON序列化/反序列化实现深克隆.
// 要求对象的所有字段都是可导出的(公开字段).
//
// 类型参数:
//   - T: 要克隆的对象类型,必须是指针类型
//
// 参数:
//   - src: 源对象指针
//
// 返回值:
//   - *T: 克隆后的新对象指针
//   - error: 克隆失败时返回错误(如JSON序列化失败)
//
// 注意:
//   - 此方法通过json.Marshal和json.Unmarshal实现
//   - 字段必须是公开的(大写开头)才能被克隆
//   - 不会克隆非导出字段
//
// 示例:
//
//	type User struct {
//	    Name string
//	    Age  int
//	}
//	original := &User{Name: "Alice", Age: 30}
//	cloned, err := core.Clone(original)
//	if err != nil {
//	    log.Fatal(err)
//	}
func Clone[T any](src *T) (*T, error) {
	if src == nil {
		return nil, nil
	}
	b, err := json.Marshal(src)
	if err != nil {
		return nil, fmt.Errorf("marshal failed: %w", err)
	}
	var dst T
	err = json.Unmarshal(b, &dst)
	if err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}
	return &dst, nil
}
