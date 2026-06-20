// Package core 的 Builder 模式支持.
//
// BuilderOption 函数用于配置 BeanDefinition 的各个属性,
// 可以在 Register 时作为可变参数传入,实现链式配置.
//
// # 使用示例
//
//	container.Register("userService",
//	    core.Bean(&UserService{}),
//	    core.Singleton(),
//	    core.DependsOn("db", "logger"),
//	    core.Init(func(s any) error { return s.(*UserService).Init() }),
//	)
//
// # 泛型支持
//
// 泛型版本的函数(如 BeanOf、FactoryOf、TypeOf)提供类型安全的API,
// 在编译期检查类型,避免反射错误.
package core

import "reflect"

// Bean 创建一个使用现有实例的构造器选项
//
// 将已有的实例注册到容器中.该实例会被直接使用,不会创建副本.
// 适用于单例对象的注册.
//
// 参数:
//   - bean: 要注册的实例,可以是任意类型的指针或值
//
// 返回值:
//   - BuilderOption: 可用于 Register 方法的构建器选项
//
// 注意:
//   - 如果注册为单例(Singleton),每次获取都返回同一个实例
//   - 如果注册为原型(Prototype),需要配合 Factory 使用
//
// 示例:
//
//	container.Register("service", core.Bean(&UserService{}))
func Bean(bean any) BuilderOption {
	t := reflect.TypeOf(bean)
	return func(def *BeanDefinition) error {
		def.Instance = bean
		def.ConcreteType = t
		def.Factory = nil
		return nil
	}
}

// BeanOf 创建一个使用现有实例的构造器选项（泛型版本）
//
// 与 Bean 功能相同，但通过泛型参数在编译期确定类型，提供类型安全。
// 类型参数 T 指定 Bean 的类型。
func BeanOf[T any](bean T) BuilderOption {
	return func(def *BeanDefinition) error {
		def.Instance = bean
		def.ConcreteType = reflect.TypeOf(bean)
		if def.ConcreteType == nil {
			def.ConcreteType = reflect.TypeFor[T]()
		}
		def.Factory = nil
		return nil
	}
}

// Factory 创建一个使用工厂函数的构造器选项
//
// 工厂函数在首次获取 Bean 时调用，用于创建复杂对象或需要依赖容器的对象。
//
// 参数:
//   - fn: 工厂函数，接收 Container 参数，返回创建的对象和错误
//   - concreteType: 工厂函数返回的具体类型
func Factory(fn func(Container) (any, error), concreteType reflect.Type) BuilderOption {
	return func(def *BeanDefinition) error {
		def.Factory = fn
		def.ConcreteType = concreteType
		return nil
	}
}

// FactoryOf 创建一个使用工厂函数的构造器选项（泛型版本）
func FactoryOf[T any](fn func(Container) (T, error)) BuilderOption {
	return func(def *BeanDefinition) error {
		def.Factory = func(c Container) (any, error) {
			return fn(c)
		}
		var zero T
		def.ConcreteType = reflect.TypeOf(zero)
		// 当 T 是接口时,reflect.TypeOf(nil) 返回 nil,
		// 尝试从函数签名推断返回类型
		if def.ConcreteType == nil {
			fnType := reflect.TypeFor[func(Container) (T, error)]()
			if fnType.Kind() == reflect.Func && fnType.NumOut() > 0 {
				def.ConcreteType = fnType.Out(0)
			}
		}
		return nil
	}
}

// Type 创建一个使用具体类型的构造器选项
//
// 仅指定 Bean 的类型，不提供实例或工厂函数，容器在获取时通过反射创建。适用于框架内部自动注册场景。
func Type(t reflect.Type) BuilderOption {
	return func(def *BeanDefinition) error {
		def.ConcreteType = t
		return nil
	}
}

// TypeOf 创建一个使用具体类型的构造器选项（泛型版本）
//
// 与 Type 功能相同，但通过泛型参数确定类型，避免手动构建反射类型。
func TypeOf[T any]() BuilderOption {
	var zero T
	return func(def *BeanDefinition) error {
		def.ConcreteType = reflect.TypeOf(zero)
		return nil
	}
}

// SetScope 设置bean的作用域
func SetScope(scope BeanScope) BuilderOption {
	return func(def *BeanDefinition) error {
		def.Scope = scope
		return nil
	}
}

// Singleton 将bean设置为单例模式
func Singleton() BuilderOption {
	return SetScope(SingletonScope)
}

// Prototype 将bean设置为原型模式
func Prototype() BuilderOption {
	return SetScope(PrototypeScope)
}

// Fields 设置字段注入列表
//
// 直接使用 FieldInjection 结构体列表配置字段注入，适用于更复杂的注入场景。
func Fields(fields ...FieldInjection) BuilderOption {
	return func(def *BeanDefinition) error {
		def.Fields = append(def.Fields, fields...)
		return nil
	}
}

// Field 创建一个值注入字段配置
//
// 向 Bean 的指定字段注入一个固定值，而非引用容器中的其他 Bean。
//
// 参数:
//   - name: 字段名称
//   - value: 要注入的值
func Field(name string, value any) BuilderOption {
	return func(def *BeanDefinition) error {
		def.Fields = append(def.Fields, FieldInjection{
			Name:  name,
			Value: value,
			IsRef: false,
		})
		return nil
	}
}

// Ref 创建一个引用注入字段配置
//
// 向 Bean 的指定字段注入容器中另一个 Bean 的引用。
//
// 参数:
//   - beanID: 被引用的 Bean ID
//   - fieldNames: 目标字段名（可选，不指定时使用 beanID 作为字段名）
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

// DependsOn 设置bean的依赖关系
func DependsOn(beanIDs ...string) BuilderOption {
	return func(def *BeanDefinition) error {
		def.DependsOn = append(def.DependsOn, beanIDs...)
		return nil
	}
}

// Init 设置bean的初始化函数
func Init(fn func(any) error) BuilderOption {
	return func(def *BeanDefinition) error {
		def.Init = fn
		return nil
	}
}

// Condition 设置bean的条件函数
func Condition(fn func(Container) bool) BuilderOption {
	return func(def *BeanDefinition) error {
		def.Condition = fn
		return nil
	}
}

// PostProcessor 添加bean后置处理器
func PostProcessor(processors ...BeanPostProcessor) BuilderOption {
	return func(def *BeanDefinition) error {
		def.PostProcessors = append(def.PostProcessors, processors...)
		return nil
	}
}

// RefreshScope 标记 Bean 支持刷新
func RefreshScope() BuilderOption {
	return func(def *BeanDefinition) error {
		def.RefreshScope = true
		return nil
	}
}

// ConfigKeys 设置 Bean 依赖的配置键
func ConfigKeys(keys ...string) BuilderOption {
	return func(def *BeanDefinition) error {
		def.ConfigKeys = keys
		return nil
	}
}
