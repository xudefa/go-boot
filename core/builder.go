package core

import "reflect"

// Bean 创建一个使用现有实例的构造器选项
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
func BeanOf[T any](bean T) BuilderOption {
	var zero T
	return func(def *BeanDefinition) error {
		def.Instance = bean
		def.ConcreteType = reflect.TypeOf(zero)
		def.Factory = nil
		return nil
	}
}

// Factory 创建一个使用工厂函数的构造器选项
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
		return nil
	}
}

// Type 创建一个使用具体类型的构造器选项
func Type(t reflect.Type) BuilderOption {
	return func(def *BeanDefinition) error {
		def.ConcreteType = t
		return nil
	}
}

// TypeOf 创建一个使用具体类型的构造器选项（泛型版本）
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

// Fields 设置字段注入
func Fields(fields ...FieldInjection) BuilderOption {
	return func(def *BeanDefinition) error {
		def.Fields = fields
		return nil
	}
}

// Field 创建一个字段注入（值注入）
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

// Ref 创建一个字段注入（引用注入）
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
		def.DependsOn = beanIDs
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
