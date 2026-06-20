package container

import "reflect"

// FactoryFunc Bean 工厂函数
type FactoryFunc func() (any, error)

// InitFunc Bean 初始化回调
type InitFunc func(bean any) error

// DestroyFunc Bean 销毁回调
type DestroyFunc func(bean any)

// ConditionFunc Bean 条件判断函数
type ConditionFunc func() bool

// BeanDefinition Bean 定义
type BeanDefinition struct {
	ID          string
	Type        reflect.Type
	Scope       Scope
	Factory     FactoryFunc
	DependsOn   []string
	InitFunc    InitFunc
	DestroyFunc DestroyFunc
	Condition   ConditionFunc
	Lazy        bool
}
