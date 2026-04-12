package core

import (
	"reflect"
)

// Bean 创建一个注册bean实例的BuilderOption
//
// 参数:
//   - bean: bean的实例
//
// 返回值:
//   - BuilderOption: 可传递给Register的选项
//
// 示例:
//
//	container.Register("service", core.Bean(&MyService{}))
func Bean(bean interface{}) BuilderOption {
	return func(def *BeanDefinition) error {
		def.Instance = bean
		def.ConcreteType = reflect.TypeOf(bean)
		def.Factory = nil
		return nil
	}
}

// Factory 创建一个使用工厂函数创建bean的BuilderOption
//
// 参数:
//   - fn: 工厂函数,接收Container参数并返回bean实例和错误
//   - concreteType: bean的具体类型(用于反射)
//
// 返回值:
//   - BuilderOption: 可传递给Register的选项
//
// 示例:
//
//	container.Register("config", core.Factory(func(c core.Container) (interface{}, error) {
//	    return loadConfig(), nil
//	}, reflect.TypeOf((*Config)(nil)).Elem()))
func Factory(fn func(Container) (interface{}, error), concreteType reflect.Type) BuilderOption {
	return func(def *BeanDefinition) error {
		def.Factory = fn
		def.ConcreteType = concreteType
		return nil
	}
}

// Type 设置bean的类型(不创建实例,仅用于类型注册)
//
// 参数:
//   - t: bean的类型
//
// 返回值:
//   - BuilderOption: 可传递给Register的选项
//
// 注意:
//   - 单独使用Type不会创建实例,通常与Factory配合使用
//   - 也可以用于标记接口类型
func Type(t reflect.Type) BuilderOption {
	return func(def *BeanDefinition) error {
		def.ConcreteType = t
		return nil
	}
}

// SetScope 设置bean的作用域
//
// 参数:
//   - scope: 作用域,参见SingletonScope和PrototypeScope
//
// 返回值:
//   - BuilderOption: 可传递给Register的选项
//
// 注意:
//   - 单例(singleton): 容器只创建一个实例,每次Get返回同一实例
//   - 原型(prototype): 每次Get都会创建新实例
func SetScope(scope BeanScope) BuilderOption {
	return func(def *BeanDefinition) error {
		def.Scope = scope
		return nil
	}
}

// Singleton 设置bean为单例作用域(默认)
//
// 返回值:
//   - BuilderOption: 可传递给Register的选项
//
// 注意:
//   - 这是默认作用域,容器只创建一个实例并缓存
func Singleton() BuilderOption {
	return SetScope(SingletonScope)
}

// Prototype 设置bean为原型作用域
//
// 返回值:
//   - BuilderOption: 可传递给Register的选项
//
// 注意:
//   - 每次调用Get都会创建新实例
//   - 原型bean不会被缓存
func Prototype() BuilderOption {
	return SetScope(PrototypeScope)
}

// Fields 设置bean的字段注入列表
//
// 参数:
//   - fields: FieldInjection类型的可变参数
//
// 返回值:
//   - BuilderOption: 可传递给Register的选项
//
// 示例:
//
//	container.Register("service", core.Bean(&Service{}),
//	    core.Fields(core.Field("Name", "custom"), core.Ref("db")))
func Fields(fields ...FieldInjection) BuilderOption {
	return func(def *BeanDefinition) error {
		def.Fields = fields
		return nil
	}
}

// Field 设置bean的单个字段值(非引用)
//
// 参数:
//   - name: 字段名
//   - value: 字段值
//
// 返回值:
//   - BuilderOption: 可传递给Register的选项
//
// 示例:
//
//	container.Register("config", core.Bean(&Config{}),
//	    core.Field("Path", "/etc/app"))
func Field(name string, value interface{}) BuilderOption {
	return func(def *BeanDefinition) error {
		def.Fields = append(def.Fields, FieldInjection{
			Name:  name,
			Value: value,
			IsRef: false,
		})
		return nil
	}
}

// Ref 设置bean的字段为另一个bean的引用
//
// 参数:
//   - beanID: 要引用的bean的ID
//   - fieldNames: 可选的字段名,默认使用beanID作为字段名
//
// 返回值:
//   - BuilderOption: 可传递给Register的选项
//
// 示例:
//
//	// 注入名为"userRepo"的bean到同名字段
//	container.Register("service", core.Bean(&Service{}), core.Ref("userRepo"))
//
//	// 注入名为"userRepo"的bean到"Repo"字段
//	container.Register("service", core.Bean(&Service{}), core.Ref("userRepo", "Repo"))
func Ref(beanID string, fieldNames ...string) BuilderOption {
	name := beanID
	if len(fieldNames) > 0 {
		name = fieldNames[0]
	}
	return func(def *BeanDefinition) error {
		def.Fields = append(def.Fields, FieldInjection{
			Name:  name,
			Value: beanID,
			IsRef: true,
		})
		return nil
	}
}

// DependsOn 设置bean的依赖顺序
//
// 参数:
//   - beanIDs: 依赖的bean ID列表
//
// 返回值:
//   - BuilderOption: 可传递给Register的选项
//
// 注意:
//   - 容器会确保依赖的bean在当前bean之前初始化
//   - 这主要用于控制初始化顺序,不是强制的依赖检查
//
// 示例:
//
//	container.Register("service", core.Bean(&Service{}),
//	    core.DependsOn("database", "logger"))
func DependsOn(beanIDs ...string) BuilderOption {
	return func(def *BeanDefinition) error {
		def.DependsOn = beanIDs
		return nil
	}
}

// Init 设置bean的初始化函数
//
// 参数:
//   - fn: 初始化函数,接收已创建的bean实例
//
// 返回值:
//   - BuilderOption: 可传递给Register的选项
//
// 注意:
//   - 初始化函数在bean创建后、注入字段后调用
//   - 可以返回错误来中止bean的创建
//
// 示例:
//
//	container.Register("service", core.Bean(&Service{}),
//	    core.Init(func(s interface{}) error {
//	        return s.(*Service).Connect()
//	    }))
func Init(fn func(interface{}) error) BuilderOption {
	return func(def *BeanDefinition) error {
		def.Init = fn
		return nil
	}
}
