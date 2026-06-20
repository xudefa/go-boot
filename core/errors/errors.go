package errors

import "fmt"

// 容器错误
var (
	ErrBeanNotFound       = fmt.Errorf("bean not found")
	ErrDuplicateBean      = fmt.Errorf("duplicate bean definition")
	ErrCircularDependency = fmt.Errorf("circular dependency detected")
	ErrInvalidBeanType    = fmt.Errorf("invalid bean type")
	ErrBeanCreation       = fmt.Errorf("bean creation failed")
)

// 配置错误
var (
	ErrInvalidConfiguration = fmt.Errorf("invalid configuration")
	ErrPropertyNotFound     = fmt.Errorf("property not found")
	ErrTypeConversion       = fmt.Errorf("type conversion failed")
)

// AOP 错误
var (
	ErrProxyCreation   = fmt.Errorf("proxy creation failed")
	ErrInvalidPointcut = fmt.Errorf("invalid pointcut expression")
)

// 生命周期错误
var (
	ErrInvalidPhase    = fmt.Errorf("invalid lifecycle phase")
	ErrPhaseTransition = fmt.Errorf("phase transition failed")
)

// ErrorBuilder 结构化错误构建器
type ErrorBuilder struct {
	code    string
	message string
	cause   error
	context map[string]any
}

// NewError 创建错误构建器
func NewError(code string) *ErrorBuilder {
	return &ErrorBuilder{
		code:    code,
		context: make(map[string]any),
	}
}

// Message 设置错误消息
func (b *ErrorBuilder) Message(msg string) *ErrorBuilder {
	b.message = msg
	return b
}

// Cause 设置根本原因
func (b *ErrorBuilder) Cause(err error) *ErrorBuilder {
	b.cause = err
	return b
}

// Context 添加上下文信息
func (b *ErrorBuilder) Context(key string, value any) *ErrorBuilder {
	b.context[key] = value
	return b
}

// Build 构建最终错误
func (b *ErrorBuilder) Build() error {
	if b.cause != nil {
		return &StructuredError{
			Code:    b.code,
			Message: b.message,
			Cause:   b.cause,
			Context: b.context,
		}
	}
	return &StructuredError{
		Code:    b.code,
		Message: b.message,
		Context: b.context,
	}
}

// StructuredError 结构化错误
type StructuredError struct {
	Code    string
	Message string
	Cause   error
	Context map[string]any
}

func (e *StructuredError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *StructuredError) Unwrap() error {
	return e.Cause
}
