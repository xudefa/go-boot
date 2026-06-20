package core

import (
	"fmt"
	"reflect"
)

// FieldInjector 统一的字段注入器
//
// 消除 tag 注入和注解注入的重复逻辑。
//
// 设计模式: Template Method
type FieldInjector struct {
	container *container
}

// NewFieldInjector 创建字段注入器
func NewFieldInjector(c *container) *FieldInjector {
	return &FieldInjector{container: c}
}

// InjectField 注入单个字段
//
// 参数:
//   - target: 目标结构体值
//   - field: 字段反射信息
//   - beanName: 指定的 Bean 名称（tag 注入时使用，空表示按类型查找）
//   - injectionType: 注入类型描述（用于错误信息）
func (fi *FieldInjector) InjectField(target reflect.Value, field reflect.StructField, beanName string, injectionType string) error {
	fieldValue := target.FieldByIndex(field.Index)
	if !fieldValue.CanSet() {
		return nil
	}

	fieldType := field.Type
	var bean any
	var err error

	if beanName != "" {
		bean, err = fi.container.Get(beanName)
	} else {
		bean, err = fi.container.getBeanByType(fieldType, field.Name, injectionType)
	}

	if err != nil {
		return fmt.Errorf("failed to inject field %s.%s: %w", target.Type().Name(), field.Name, err)
	}

	if err := fi.setFieldValue(fieldValue, bean); err != nil {
		return fmt.Errorf("failed to set field %s.%s: %w", target.Type().Name(), field.Name, err)
	}

	return nil
}

// setFieldValue 设置字段值，处理指针/值转换
func (fi *FieldInjector) setFieldValue(fieldValue reflect.Value, bean any) error {
	beanVal := reflect.ValueOf(bean)
	fieldType := fieldValue.Type()

	switch {
	case fieldType.Kind() == reflect.Ptr && beanVal.Kind() != reflect.Ptr:
		if !beanVal.CanAddr() {
			ptr := reflect.New(fieldType.Elem())
			ptr.Elem().Set(beanVal)
			fieldValue.Set(ptr)
			return nil
		}
		fieldValue.Set(beanVal.Addr())

	case fieldType.Kind() != reflect.Ptr && beanVal.Kind() == reflect.Ptr:
		if beanVal.IsNil() {
			fieldValue.Set(reflect.Zero(fieldType))
			return nil
		}
		fieldValue.Set(beanVal.Elem())

	default:
		fieldValue.Set(beanVal)
	}

	return nil
}

// getBeanByType 按类型获取 Bean
func (c *container) getBeanByType(targetType reflect.Type, fieldName string, context string) (any, error) {
	c.registry.lock.RLock()
	ids, ok := c.registry.typeToIDs[targetType]
	c.registry.lock.RUnlock()

	if !ok || len(ids) == 0 {
		return nil, fmt.Errorf("no bean of type %s found for injection", targetType)
	}

	if len(ids) > 1 {
		return nil, fmt.Errorf("multiple beans of type %s found: %v", targetType, ids)
	}

	return c.Get(ids[0])
}
