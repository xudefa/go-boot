package environment

import (
	"fmt"
	"reflect"
	"strconv"
)

// TypeConverter 类型转换器
//
// 统一的类型转换逻辑，消除重复代码。
type TypeConverter struct{}

// NewTypeConverter 创建类型转换器
func NewTypeConverter() *TypeConverter {
	return &TypeConverter{}
}

// ConvertTo 将值转换为目标类型
func (c *TypeConverter) ConvertTo(val any, targetType reflect.Type) (reflect.Value, error) {
	rv := reflect.ValueOf(val)

	// 如果类型已经匹配，直接返回
	if rv.Type().AssignableTo(targetType) {
		return rv, nil
	}

	// 特殊处理：数值类型到 string 的转换（避免 ASCII 转换）
	if targetType.Kind() == reflect.String && isNumeric(rv.Kind()) {
		return c.toString(val)
	}

	// 尝试标准转换
	if rv.Type().ConvertibleTo(targetType) {
		return rv.Convert(targetType), nil
	}

	// 特殊处理
	return c.specialConvert(val, targetType)
}

// isNumeric 检查是否为数值类型
func isNumeric(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// specialConvert 处理特殊类型转换
func (c *TypeConverter) specialConvert(val any, targetType reflect.Type) (reflect.Value, error) {
	switch targetType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return c.toInt(val, targetType)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return c.toUint(val, targetType)
	case reflect.Float32, reflect.Float64:
		return c.toFloat(val, targetType)
	case reflect.Bool:
		return c.toBool(val)
	case reflect.String:
		return c.toString(val)
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %T to %s", val, targetType)
}

func (c *TypeConverter) toInt(val any, targetType reflect.Type) (reflect.Value, error) {
	var n int64
	switch v := val.(type) {
	case int:
		n = int64(v)
	case int8:
		n = int64(v)
	case int16:
		n = int64(v)
	case int32:
		n = int64(v)
	case int64:
		n = v
	case float64:
		n = int64(v)
	case float32:
		n = int64(v)
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert string %q to int: %w", v, err)
		}
		n = parsed
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %T to int", val)
	}

	switch targetType.Kind() {
	case reflect.Int:
		return reflect.ValueOf(int(n)), nil
	case reflect.Int8:
		return reflect.ValueOf(int8(n)), nil
	case reflect.Int16:
		return reflect.ValueOf(int16(n)), nil
	case reflect.Int32:
		return reflect.ValueOf(int32(n)), nil
	case reflect.Int64:
		return reflect.ValueOf(n), nil
	}

	return reflect.Value{}, fmt.Errorf("unsupported int type: %s", targetType)
}

func (c *TypeConverter) toUint(val any, targetType reflect.Type) (reflect.Value, error) {
	var n uint64
	switch v := val.(type) {
	case uint:
		n = uint64(v)
	case uint8:
		n = uint64(v)
	case uint16:
		n = uint64(v)
	case uint32:
		n = uint64(v)
	case uint64:
		n = v
	case float64:
		n = uint64(v)
	case float32:
		n = uint64(v)
	case int:
		n = uint64(v)
	case int64:
		n = uint64(v)
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %T to uint", val)
	}

	switch targetType.Kind() {
	case reflect.Uint:
		return reflect.ValueOf(uint(n)), nil
	case reflect.Uint8:
		return reflect.ValueOf(uint8(n)), nil
	case reflect.Uint16:
		return reflect.ValueOf(uint16(n)), nil
	case reflect.Uint32:
		return reflect.ValueOf(uint32(n)), nil
	case reflect.Uint64:
		return reflect.ValueOf(n), nil
	}

	return reflect.Value{}, fmt.Errorf("unsupported uint type: %s", targetType)
}

func (c *TypeConverter) toFloat(val any, targetType reflect.Type) (reflect.Value, error) {
	var f float64
	switch v := val.(type) {
	case float64:
		f = v
	case float32:
		f = float64(v)
	case int:
		f = float64(v)
	case int64:
		f = float64(v)
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert string %q to float: %w", v, err)
		}
		f = parsed
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %T to float", val)
	}

	if targetType.Kind() == reflect.Float32 {
		return reflect.ValueOf(float32(f)), nil
	}
	return reflect.ValueOf(f), nil
}

func (c *TypeConverter) toBool(val any) (reflect.Value, error) {
	switch v := val.(type) {
	case bool:
		return reflect.ValueOf(v), nil
	case string:
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert string %q to bool: %w", v, err)
		}
		return reflect.ValueOf(parsed), nil
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %T to bool", val)
}

func (c *TypeConverter) toString(val any) (reflect.Value, error) {
	switch v := val.(type) {
	case string:
		return reflect.ValueOf(v), nil
	case int:
		return reflect.ValueOf(strconv.Itoa(v)), nil
	case int8:
		return reflect.ValueOf(strconv.FormatInt(int64(v), 10)), nil
	case int16:
		return reflect.ValueOf(strconv.FormatInt(int64(v), 10)), nil
	case int32:
		return reflect.ValueOf(strconv.FormatInt(int64(v), 10)), nil
	case int64:
		return reflect.ValueOf(strconv.FormatInt(v, 10)), nil
	case uint:
		return reflect.ValueOf(strconv.FormatUint(uint64(v), 10)), nil
	case uint8:
		return reflect.ValueOf(strconv.FormatUint(uint64(v), 10)), nil
	case uint16:
		return reflect.ValueOf(strconv.FormatUint(uint64(v), 10)), nil
	case uint32:
		return reflect.ValueOf(strconv.FormatUint(uint64(v), 10)), nil
	case uint64:
		return reflect.ValueOf(strconv.FormatUint(v, 10)), nil
	case float64:
		return reflect.ValueOf(strconv.FormatFloat(v, 'f', -1, 64)), nil
	case float32:
		return reflect.ValueOf(strconv.FormatFloat(float64(v), 'f', -1, 32)), nil
	case bool:
		return reflect.ValueOf(strconv.FormatBool(v)), nil
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %T to string", val)
}
