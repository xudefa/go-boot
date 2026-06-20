package validation

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// validateCrossField 验证跨字段规则
func (v *TagValidator) validateCrossField(field reflect.Value, rule, fieldName string, obj interface{}) error {
	if !strings.HasPrefix(rule, "field") {
		return nil
	}

	parts := strings.SplitN(rule, "=", 2)
	if len(parts) != 2 {
		return nil
	}

	targetFieldName := parts[1]
	targetValue := getFieldValue(obj, targetFieldName)
	if targetValue == nil {
		return ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("字段 %s 不存在", targetFieldName),
			Value:   field.Interface(),
		}
	}

	targetReflectValue := reflect.ValueOf(targetValue)

	switch parts[0] {
	case "fieldmatch":
		if !compareValues(field, targetReflectValue, "==") {
			return ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("字段值必须等于 %s 的值", targetFieldName),
				Value:   field.Interface(),
			}
		}
	case "fieldne":
		if !compareValues(field, targetReflectValue, "!=") {
			return ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("字段值必须不等于 %s 的值", targetFieldName),
				Value:   field.Interface(),
			}
		}
	case "fieldgt":
		if !compareValues(field, targetReflectValue, ">") {
			return ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("字段值必须大于 %s 的值", targetFieldName),
				Value:   field.Interface(),
			}
		}
	case "fieldlt":
		if !compareValues(field, targetReflectValue, "<") {
			return ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("字段值必须小于 %s 的值", targetFieldName),
				Value:   field.Interface(),
			}
		}
	case "fieldgte":
		if !compareValues(field, targetReflectValue, ">=") {
			return ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("字段值必须大于等于 %s 的值", targetFieldName),
				Value:   field.Interface(),
			}
		}
	case "fieldlte":
		if !compareValues(field, targetReflectValue, "<=") {
			return ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("字段值必须小于等于 %s 的值", targetFieldName),
				Value:   field.Interface(),
			}
		}
	}

	return nil
}

// validateWhenCondition 验证条件依赖规则
func (v *TagValidator) validateWhenCondition(field reflect.Value, rule, fieldName string, obj interface{}) []ValidationError {
	if !strings.HasPrefix(rule, "when=") {
		return nil
	}

	conditionStr := strings.TrimPrefix(rule, "when=")
	parts := strings.SplitN(conditionStr, ":", 2)
	if len(parts) != 2 {
		return nil
	}

	condition := parts[0]
	rulesStr := parts[1]

	conditionMet := evaluateCondition(condition, obj)
	if !conditionMet {
		return nil
	}

	// 将分号替换为逗号，因为 validateField 使用逗号分隔规则
	normalizedRules := strings.ReplaceAll(rulesStr, ";", ",")

	// 检查是否包含 required 规则
	rules := strings.Split(normalizedRules, ",")
	isRequired := false
	for _, r := range rules {
		r = strings.TrimSpace(r)
		if r == "required" {
			isRequired = true
			break
		}
	}

	// 如果字段是必需的且为空，则报告错误
	if isRequired && !v.isRequiredValid(field) {
		return []ValidationError{
			{
				Field:   fieldName,
				Message: "字段是必需的",
				Value:   field.Interface(),
			},
		}
	}

	return v.validateField(field, normalizedRules, fieldName, obj)
}

// getFieldValue 获取对象中指定字段的值
func getFieldValue(obj interface{}, fieldName string) interface{} {
	rv := reflect.ValueOf(obj)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return nil
	}

	field := rv.FieldByName(fieldName)
	if !field.IsValid() {
		return nil
	}

	return field.Interface()
}

// compareValues 比较两个字段的值
func compareValues(field1, field2 reflect.Value, operator string) bool {
	v1 := field1.Interface()
	v2 := field2.Interface()

	switch operator {
	case "==":
		return compareEqual(v1, v2)
	case "!=":
		return !compareEqual(v1, v2)
	case ">":
		return compareNumeric(v1, v2, ">")
	case "<":
		return compareNumeric(v1, v2, "<")
	case ">=":
		return compareNumeric(v1, v2, ">=")
	case "<=":
		return compareNumeric(v1, v2, "<=")
	}

	return false
}

// compareEqual 比较两个值是否相等（类型安全）
func compareEqual(v1, v2 interface{}) bool {
	rv1 := reflect.ValueOf(v1)
	rv2 := reflect.ValueOf(v2)

	if rv1.Type() != rv2.Type() {
		return false
	}

	switch rv1.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv1.Int() == rv2.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv1.Uint() == rv2.Uint()
	case reflect.Float32, reflect.Float64:
		return rv1.Float() == rv2.Float()
	case reflect.String:
		return rv1.String() == rv2.String()
	case reflect.Bool:
		return rv1.Bool() == rv2.Bool()
	default:
		return fmt.Sprintf("%v", v1) == fmt.Sprintf("%v", v2)
	}
}

// compareNumeric 比较数值类型
func compareNumeric(v1, v2 interface{}, operator string) bool {
	n1, err1 := toFloat64(v1)
	n2, err2 := toFloat64(v2)

	if err1 != nil || err2 != nil {
		return false
	}

	switch operator {
	case ">":
		return n1 > n2
	case "<":
		return n1 < n2
	case ">=":
		return n1 >= n2
	case "<=":
		return n1 <= n2
	}

	return false
}

// toFloat64 将值转换为 float64
func toFloat64(v interface{}) (float64, error) {
	rv := reflect.ValueOf(v)

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(rv.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return rv.Float(), nil
	default:
		return 0, fmt.Errorf("不支持的类型: %T", v)
	}
}

// evaluateCondition 评估条件表达式
func evaluateCondition(condition string, obj interface{}) bool {
	parts := strings.Split(condition, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if evaluateSingleCondition(part, obj) {
			return true
		}
	}
	return false
}

// evaluateSingleCondition 评估单个条件
func evaluateSingleCondition(condition string, obj interface{}) bool {
	if strings.Contains(condition, "==") {
		return evaluateComparisonCondition(condition, obj, "==")
	}

	if strings.Contains(condition, "!=") {
		return evaluateComparisonCondition(condition, obj, "!=")
	}

	if strings.Contains(condition, ">=") {
		return evaluateNumericCondition(condition, obj, ">=")
	}

	if strings.Contains(condition, "<=") {
		return evaluateNumericCondition(condition, obj, "<=")
	}

	if strings.Contains(condition, "<") {
		return evaluateNumericCondition(condition, obj, "<")
	}

	if strings.Contains(condition, ">") {
		return evaluateNumericCondition(condition, obj, ">")
	}

	return false
}

// evaluateComparisonCondition 评估比较条件（== 或 !=）
func evaluateComparisonCondition(condition string, obj interface{}, operator string) bool {
	var parts []string
	if operator == "==" {
		parts = strings.SplitN(condition, "==", 2)
	} else {
		parts = strings.SplitN(condition, "!=", 2)
	}

	if len(parts) != 2 {
		return false
	}

	fieldName := strings.TrimSpace(parts[0])
	expectedValue := strings.TrimSpace(parts[1])

	fieldValue := getFieldValue(obj, fieldName)
	if fieldValue == nil {
		return false
	}

	// 处理布尔值比较
	if boolValue, ok := fieldValue.(bool); ok {
		switch expectedValue {
		case "true":
			if operator == "==" {
				return boolValue
			}
			return !boolValue
		case "false":
			if operator == "==" {
				return !boolValue
			}
			return boolValue
		}
		return false
	}

	// 处理数值比较
	if numValue, ok := fieldValue.(int); ok {
		if expectedNum, err := strconv.Atoi(expectedValue); err == nil {
			if operator == "==" {
				return numValue == expectedNum
			}
			return numValue != expectedNum
		}
	}

	// 默认字符串比较
	actualValue := fmt.Sprintf("%v", fieldValue)
	if operator == "==" {
		return actualValue == expectedValue
	}
	return actualValue != expectedValue
}

// evaluateNumericCondition 评估数值条件（>=, <=, <, >）
func evaluateNumericCondition(condition string, obj interface{}, operator string) bool {
	var parts []string
	switch operator {
	case ">=":
		parts = strings.SplitN(condition, ">=", 2)
	case "<=":
		parts = strings.SplitN(condition, "<=", 2)
	case "<":
		parts = strings.SplitN(condition, "<", 2)
	case ">":
		parts = strings.SplitN(condition, ">", 2)
	}

	if len(parts) != 2 {
		return false
	}

	fieldName := strings.TrimSpace(parts[0])
	expectedValue := strings.TrimSpace(parts[1])

	fieldValue := getFieldValue(obj, fieldName)
	if fieldValue == nil {
		return false
	}

	expected, err := strconv.ParseFloat(expectedValue, 64)
	if err != nil {
		return false
	}

	actual, err := toFloat64(fieldValue)
	if err != nil {
		return false
	}

	switch operator {
	case ">=":
		return actual >= expected
	case "<=":
		return actual <= expected
	case "<":
		return actual < expected
	case ">":
		return actual > expected
	}

	return false
}
