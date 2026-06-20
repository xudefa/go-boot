// Package validation 提供基于结构体标签的数据验证功能。
//
// 支持丰富的验证规则，包括：
//   - 必填验证：required
//   - 长度验证：min、max、len、gt、gte、lt、lte
//   - 格式验证：email、url、ip、regexp
//   - 范围验证：oneof
//   - 跨字段验证：field
//   - 条件验证：when
//   - 自定义验证：通过 ValidatorRegistry 注册
//
// 使用示例：
//
//	type User struct {
//	    Name  string `validate:"required,min=2,max=50"`
//	    Email string `validate:"required,email"`
//	    Age   int    `validate:"required,gte=0,lte=150"`
//	}
//	err := validation.ValidateStruct(user)
package validation

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Validator 验证器接口，定义了验证操作的标准接口
type Validator interface {
	Validate(obj interface{}) error
}

// ValidationError 验证错误结构，包含字段名称、错误消息和实际值
type ValidationError struct {
	Field   string      `json:"field"`           // 字段名称
	Message string      `json:"message"`         // 错误消息
	Value   interface{} `json:"value,omitempty"` // 实际值
}

// Error 实现错误接口，返回格式化的错误字符串
func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors 验证错误集合，实现了错误接口
type ValidationErrors []ValidationError

// Error 实现错误接口，返回所有验证错误的组合字符串
func (e ValidationErrors) Error() string {
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// TagValidator 基于标签的验证器，支持多种验证规则
type TagValidator struct {
	registry *ValidatorRegistry
}

// NewTagValidator 创建新的标签验证器实例
func NewTagValidator() *TagValidator {
	return &TagValidator{
		registry: NewValidatorRegistry(),
	}
}

// NewTagValidatorWithRegistry 创建带有注册表的标签验证器实例
func NewTagValidatorWithRegistry(registry *ValidatorRegistry) *TagValidator {
	return &TagValidator{
		registry: registry,
	}
}

// Validate 验证对象，对结构体的字段进行验证
func (v *TagValidator) Validate(obj interface{}) error {
	if obj == nil {
		return nil
	}

	rv := reflect.ValueOf(obj)
	rt := reflect.TypeOf(obj)

	// 解引用指针
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
		rt = rt.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return errors.New("validation: only structs are supported")
	}

	var errs ValidationErrors

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rt.Field(i)

		// 检查是否存在验证标签
		tag := fieldType.Tag.Get("validate")
		if tag == "" {
			continue
		}

		// 获取字段名，优先使用json标签
		fieldName := fieldType.Name
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
		}

		fieldErrors := v.validateField(field, tag, fieldName, obj)
		errs = append(errs, fieldErrors...)
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// validateField 验证单个字段，解析验证规则并执行验证
func (v *TagValidator) validateField(field reflect.Value, tag, fieldName string, obj interface{}) []ValidationError {
	var errs []ValidationError

	// 分割验证规则
	rules := strings.Split(tag, ",")

	// 检查字段是否必需
	isRequired := false
	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "required" {
			isRequired = true
			break
		}
	}

	// 检查字段是否为空值
	isEmpty := !v.isRequiredValid(field)

	// 如果字段是必需的且为空，则报告错误
	if isRequired && isEmpty {
		errs = append(errs, ValidationError{
			Field:   fieldName,
			Message: "字段是必需的",
			Value:   field.Interface(),
		})
		// 即使字段为空，也要检查其他规则（比如长度规则），因为某些规则可能适用于空值
	}

	// 遍历所有规则进行验证
	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}

		// 处理条件依赖验证（必须在跨字段验证之前处理，因为 when 可能包含 field 规则）
		if strings.HasPrefix(rule, "when=") {
			whenErrs := v.validateWhenCondition(field, rule, fieldName, obj)
			if len(whenErrs) > 0 {
				errs = append(errs, whenErrs...)
			}
			continue
		}

		// 处理跨字段验证
		if strings.HasPrefix(rule, "field") {
			err := v.validateCrossField(field, rule, fieldName, obj)
			if err != nil {
				errs = append(errs, err.(ValidationError))
			}
			continue
		}

		if rule == "required" {
			continue
		}

		// 解析规则（支持带参数的规则，如 min=10, max=100）
		if strings.Contains(rule, "=") {
			parts := strings.SplitN(rule, "=", 2)
			ruleName := parts[0]
			ruleValue := parts[1]

			switch ruleName {
			case "min":
				if !v.isMinValid(field, ruleValue) {
					min, _ := strconv.Atoi(ruleValue)
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: fmt.Sprintf("字段值必须大于或等于 %d", min),
						Value:   field.Interface(),
					})
				}
			case "max":
				if !v.isMaxValid(field, ruleValue) {
					max, _ := strconv.Atoi(ruleValue)
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: fmt.Sprintf("字段值必须小于或等于 %d", max),
						Value:   field.Interface(),
					})
				}
			case "len":
				if !v.isLenValid(field, ruleValue) {
					length, _ := strconv.Atoi(ruleValue)
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: fmt.Sprintf("字段长度必须为 %d", length),
						Value:   field.Interface(),
					})
				}
			case "email":
				if !v.isEmailValid(field) {
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: "字段必须是有效的邮箱地址",
						Value:   field.Interface(),
					})
				}
			case "regexp":
				if !v.isRegexpValid(field, ruleValue) {
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: fmt.Sprintf("字段不匹配正则表达式: %s", ruleValue),
						Value:   field.Interface(),
					})
				}
			case "gt": // Greater Than
				if !v.isGtValid(field, ruleValue) {
					gt, _ := strconv.Atoi(ruleValue)
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: fmt.Sprintf("字段值必须大于 %d", gt),
						Value:   field.Interface(),
					})
				}
			case "gte": // Greater Than or Equal
				if !v.isGteValid(field, ruleValue) {
					gte, _ := strconv.Atoi(ruleValue)
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: fmt.Sprintf("字段值必须大于或等于 %d", gte),
						Value:   field.Interface(),
					})
				}
			case "lt": // Less Than
				if !v.isLtValid(field, ruleValue) {
					lt, _ := strconv.Atoi(ruleValue)
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: fmt.Sprintf("字段值必须小于 %d", lt),
						Value:   field.Interface(),
					})
				}
			case "lte": // Less Than or Equal
				if !v.isLteValid(field, ruleValue) {
					lte, _ := strconv.Atoi(ruleValue)
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: fmt.Sprintf("字段值必须小于或等于 %d", lte),
						Value:   field.Interface(),
					})
				}
			case "url":
				if !v.isUrlValid(field) {
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: "字段必须是有效的URL地址",
						Value:   field.Interface(),
					})
				}
			case "ip":
				if !v.isIpValid(field) {
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: "字段必须是有效的IP地址",
						Value:   field.Interface(),
					})
				}
			case "oneof":
				if !v.isOneOfValid(field, ruleValue) {
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: fmt.Sprintf("字段值必须是以下选项之一: %s", ruleValue),
						Value:   field.Interface(),
					})
				}
			}
		} else {
			// 无参数的规则（但不包括required，已在上面处理）
			switch rule {
			case "email":
				if !v.isEmailValid(field) {
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: "字段必须是有效的邮箱地址",
						Value:   field.Interface(),
					})
				}
			case "url":
				if !v.isUrlValid(field) {
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: "字段必须是有效的URL地址",
						Value:   field.Interface(),
					})
				}
			case "ip":
				if !v.isIpValid(field) {
					errs = append(errs, ValidationError{
						Field:   fieldName,
						Message: "字段必须是有效的IP地址",
						Value:   field.Interface(),
					})
				}
			default:
				// 检查是否为自定义验证器
				if v.registry != nil {
					if customValidator, ok := v.registry.GetFunc(rule); ok {
						if valid, msg := customValidator(field, ""); !valid {
							errs = append(errs, ValidationError{
								Field:   fieldName,
								Message: msg,
								Value:   field.Interface(),
							})
						}
					}
				}
			}
		}
	}

	return errs
}

// isRequiredValid 验证字段是否必需（非零值）
func (v *TagValidator) isRequiredValid(field reflect.Value) bool {
	switch field.Kind() {
	case reflect.String:
		return field.String() != ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return field.Float() != 0
	case reflect.Bool:
		return field.Bool()
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return !field.IsNil()
	default:
		return field.IsValid() && field.Interface() != reflect.Zero(field.Type()).Interface()
	}
}

// isMinValid 验证最小值
func (v *TagValidator) isMinValid(field reflect.Value, minStr string) bool {
	min, err := strconv.Atoi(minStr)
	if err != nil {
		return false
	}

	switch field.Kind() {
	case reflect.String:
		return len([]rune(field.String())) >= min // 使用rune来计算字符数，而不是字节数
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() >= int64(min)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() >= uint64(min)
	case reflect.Float32, reflect.Float64:
		return field.Float() >= float64(min)
	default:
		return false
	}
}

// isMaxValid 验证最大值
func (v *TagValidator) isMaxValid(field reflect.Value, maxStr string) bool {
	max, err := strconv.Atoi(maxStr)
	if err != nil {
		return false
	}

	switch field.Kind() {
	case reflect.String:
		return len([]rune(field.String())) <= max // 使用rune来计算字符数，而不是字节数
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() <= int64(max)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() <= uint64(max)
	case reflect.Float32, reflect.Float64:
		return field.Float() <= float64(max)
	default:
		return false
	}
}

// isLenValid 验证长度
func (v *TagValidator) isLenValid(field reflect.Value, lenStr string) bool {
	length, err := strconv.Atoi(lenStr)
	if err != nil {
		return false
	}

	switch field.Kind() {
	case reflect.String:
		return len([]rune(field.String())) == length // 使用rune来计算字符数，而不是字节数
	case reflect.Array, reflect.Slice:
		return field.Len() == length
	default:
		return false
	}
}

// isEmailValid 验证邮箱格式
func (v *TagValidator) isEmailValid(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}
	email := field.String()
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// isRegexpValid 验证正则表达式匹配
func (v *TagValidator) isRegexpValid(field reflect.Value, pattern string) bool {
	if field.Kind() != reflect.String {
		return false
	}
	matched, err := regexp.MatchString(pattern, field.String())
	return err == nil && matched
}

// isGtValid 验证大于指定值
func (v *TagValidator) isGtValid(field reflect.Value, valueStr string) bool {
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return false
	}

	switch field.Kind() {
	case reflect.String:
		return len([]rune(field.String())) > value // 使用rune来计算字符数，而不是字节数
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() > int64(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() > uint64(value)
	case reflect.Float32, reflect.Float64:
		return field.Float() > float64(value)
	default:
		return false
	}
}

// isGteValid 验证大于等于指定值
func (v *TagValidator) isGteValid(field reflect.Value, valueStr string) bool {
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return false
	}

	switch field.Kind() {
	case reflect.String:
		return len([]rune(field.String())) >= value // 使用rune来计算字符数，而不是字节数
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() >= int64(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() >= uint64(value)
	case reflect.Float32, reflect.Float64:
		return field.Float() >= float64(value)
	default:
		return false
	}
}

// isLtValid 验证小于指定值
func (v *TagValidator) isLtValid(field reflect.Value, valueStr string) bool {
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return false
	}

	switch field.Kind() {
	case reflect.String:
		return len([]rune(field.String())) < value // 使用rune来计算字符数，而不是字节数
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() < int64(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() < uint64(value)
	case reflect.Float32, reflect.Float64:
		return field.Float() < float64(value)
	default:
		return false
	}
}

// isLteValid 验证小于等于指定值
func (v *TagValidator) isLteValid(field reflect.Value, valueStr string) bool {
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return false
	}

	switch field.Kind() {
	case reflect.String:
		return len([]rune(field.String())) <= value // 使用rune来计算字符数，而不是字节数
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() <= int64(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() <= uint64(value)
	case reflect.Float32, reflect.Float64:
		return field.Float() <= float64(value)
	default:
		return false
	}
}

// isUrlValid 验证URL格式
func (v *TagValidator) isUrlValid(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}
	url := field.String()
	urlRegex := regexp.MustCompile(`^https?:\/\/(?:[-\w.])+(?:\:[0-9]+)?(?:\/(?:[\w\/_.])*(?:\?(?:[\w&=%.])*)?(?:\#(?:[\w.])*)?)?$`)
	return urlRegex.MatchString(url)
}

// isIpValid 验证IP地址格式
func (v *TagValidator) isIpValid(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}
	ip := field.String()
	ipRegex := regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
	return ipRegex.MatchString(ip)
}

// isOneOfValid 验证值是否在指定的选项中
func (v *TagValidator) isOneOfValid(field reflect.Value, optionsStr string) bool {
	options := strings.Split(optionsStr, " ")

	// 移除每个选项前后的空格
	for i, opt := range options {
		options[i] = strings.TrimSpace(opt)
	}

	switch field.Kind() {
	case reflect.String:
		value := field.String()
		for _, opt := range options {
			if value == opt {
				return true
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value := field.Int()
		for _, opt := range options {
			if optNum, err := strconv.ParseInt(opt, 10, 64); err == nil && value == optNum {
				return true
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value := field.Uint()
		for _, opt := range options {
			if optNum, err := strconv.ParseUint(opt, 10, 64); err == nil && value == optNum {
				return true
			}
		}
	case reflect.Float32, reflect.Float64:
		value := field.Float()
		for _, opt := range options {
			if optNum, err := strconv.ParseFloat(opt, 64); err == nil && value == optNum {
				return true
			}
		}
	}

	return false
}

// ValidateStruct 用于验证结构体的便捷函数
func ValidateStruct(obj interface{}) error {
	validator := NewTagValidator()
	return validator.Validate(obj)
}

// Validate 验证单个值是否符合规则
func Validate(value interface{}, rules string) error {
	// 由于反射无法直接对基本类型应用标签验证，
	// 我们需要实现一个简化版本的验证逻辑
	ruleList := strings.Split(rules, ",")
	for _, rule := range ruleList {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}

		// 检查规则格式
		if strings.Contains(rule, "=") {
			parts := strings.SplitN(rule, "=", 2)
			ruleName := parts[0]
			ruleValue := parts[1]

			var isValid bool
			switch ruleName {
			case "min":
				isValid = isMinValidForValue(value, ruleValue)
			case "max":
				isValid = isMaxValidForValue(value, ruleValue)
			case "len":
				isValid = isLenValidForValue(value, ruleValue)
			case "email":
				isValid = isEmailValidForValue(value)
			case "regexp":
				isValid = isRegexpValidForValue(value, ruleValue)
			case "gt":
				isValid = isGtValidForValue(value, ruleValue)
			case "gte":
				isValid = isGteValidForValue(value, ruleValue)
			case "lt":
				isValid = isLtValidForValue(value, ruleValue)
			case "lte":
				isValid = isLteValidForValue(value, ruleValue)
			case "url":
				isValid = isUrlValidForValue(value)
			case "ip":
				isValid = isIpValidForValue(value)
			case "oneof":
				isValid = isOneOfValidForValue(value, ruleValue)
			default:
				continue
			}

			if !isValid {
				return fmt.Errorf("validation failed for rule: %s", rule)
			}
		} else {
			// 无参数的规则
			var isValid bool
			switch rule {
			case "required":
				isValid = isRequiredValidForValue(value)
			case "email":
				isValid = isEmailValidForValue(value)
			case "url":
				isValid = isUrlValidForValue(value)
			case "ip":
				isValid = isIpValidForValue(value)
			default:
				continue
			}

			if !isValid {
				return fmt.Errorf("validation failed for rule: %s", rule)
			}
		}
	}

	return nil
}

// 以下是辅助函数，用于验证单个值
func isRequiredValidForValue(value interface{}) bool {
	if value == nil {
		return false
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return v.String() != ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return v.Float() != 0
	case reflect.Bool:
		return v.Bool()
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return !v.IsNil()
	default:
		return v.IsValid() && value != reflect.Zero(v.Type()).Interface()
	}
}

func isMinValidForValue(value interface{}, minStr string) bool {
	min, err := strconv.Atoi(minStr)
	if err != nil {
		return false
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return len([]rune(v.String())) >= min // 使用rune来计算字符数，而不是字节数
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() >= int64(min)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() >= uint64(min)
	case reflect.Float32, reflect.Float64:
		return v.Float() >= float64(min)
	default:
		return false
	}
}

func isMaxValidForValue(value interface{}, maxStr string) bool {
	max, err := strconv.Atoi(maxStr)
	if err != nil {
		return false
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return len([]rune(v.String())) <= max // 使用rune来计算字符数，而不是字节数
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() <= int64(max)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() <= uint64(max)
	case reflect.Float32, reflect.Float64:
		return v.Float() <= float64(max)
	default:
		return false
	}
}

func isLenValidForValue(value interface{}, lenStr string) bool {
	length, err := strconv.Atoi(lenStr)
	if err != nil {
		return false
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return len([]rune(v.String())) == length // 使用rune来计算字符数，而不是字节数
	case reflect.Array, reflect.Slice:
		return v.Len() == length
	default:
		return false
	}
}

func isEmailValidForValue(value interface{}) bool {
	if value == nil {
		return false
	}

	v := reflect.ValueOf(value)
	if v.Kind() != reflect.String {
		return false
	}
	email := v.String()
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func isRegexpValidForValue(value interface{}, pattern string) bool {
	if value == nil {
		return false
	}

	v := reflect.ValueOf(value)
	if v.Kind() != reflect.String {
		return false
	}
	matched, err := regexp.MatchString(pattern, v.String())
	return err == nil && matched
}

func isGtValidForValue(value interface{}, valueStr string) bool {
	val, err := strconv.Atoi(valueStr)
	if err != nil {
		return false
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return len([]rune(v.String())) > val // 使用rune来计算字符数，而不是字节数
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() > int64(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() > uint64(val)
	case reflect.Float32, reflect.Float64:
		return v.Float() > float64(val)
	default:
		return false
	}
}

func isGteValidForValue(value interface{}, valueStr string) bool {
	val, err := strconv.Atoi(valueStr)
	if err != nil {
		return false
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return len([]rune(v.String())) >= val // 使用rune来计算字符数，而不是字节数
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() >= int64(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() >= uint64(val)
	case reflect.Float32, reflect.Float64:
		return v.Float() >= float64(val)
	default:
		return false
	}
}

func isLtValidForValue(value interface{}, valueStr string) bool {
	val, err := strconv.Atoi(valueStr)
	if err != nil {
		return false
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return len([]rune(v.String())) < val // 使用rune来计算字符数，而不是字节数
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() < int64(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() < uint64(val)
	case reflect.Float32, reflect.Float64:
		return v.Float() < float64(val)
	default:
		return false
	}
}

func isLteValidForValue(value interface{}, valueStr string) bool {
	val, err := strconv.Atoi(valueStr)
	if err != nil {
		return false
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return len([]rune(v.String())) <= val // 使用rune来计算字符数，而不是字节数
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() <= int64(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() <= uint64(val)
	case reflect.Float32, reflect.Float64:
		return v.Float() <= float64(val)
	default:
		return false
	}
}

func isUrlValidForValue(value interface{}) bool {
	if value == nil {
		return false
	}

	v := reflect.ValueOf(value)
	if v.Kind() != reflect.String {
		return false
	}
	url := v.String()
	urlRegex := regexp.MustCompile(`^https?:\/\/(?:[-\w.])+(?:\:[0-9]+)?(?:\/(?:[\w\/_.])*(?:\?(?:[\w&=%.])*)?(?:\#(?:[\w.])*)?)?$`)
	return urlRegex.MatchString(url)
}

func isIpValidForValue(value interface{}) bool {
	if value == nil {
		return false
	}

	v := reflect.ValueOf(value)
	if v.Kind() != reflect.String {
		return false
	}
	ip := v.String()
	ipRegex := regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
	return ipRegex.MatchString(ip)
}

func isOneOfValidForValue(value interface{}, optionsStr string) bool {
	if value == nil {
		return false
	}

	options := strings.Split(optionsStr, " ")

	// 移除每个选项前后的空格
	for i, opt := range options {
		options[i] = strings.TrimSpace(opt)
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		strVal := v.String()
		for _, opt := range options {
			if strVal == opt {
				return true
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal := v.Int()
		for _, opt := range options {
			if optNum, err := strconv.ParseInt(opt, 10, 64); err == nil && intVal == optNum {
				return true
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal := v.Uint()
		for _, opt := range options {
			if optNum, err := strconv.ParseUint(opt, 10, 64); err == nil && uintVal == optNum {
				return true
			}
		}
	case reflect.Float32, reflect.Float64:
		floatVal := v.Float()
		for _, opt := range options {
			if optNum, err := strconv.ParseFloat(opt, 64); err == nil && floatVal == optNum {
				return true
			}
		}
	}

	return false
}
