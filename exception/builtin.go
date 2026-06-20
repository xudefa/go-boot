package exception

import (
	"context"
	"errors"
)

var (
	ErrNotFound       = errors.New("resource not found")
	ErrBadRequest     = errors.New("bad request")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrConflict       = errors.New("conflict")
	ErrInternalServer = errors.New("internal server error")
)

// DefaultExceptionResolver 默认异常解析器（兜底）
//
// DefaultExceptionResolver 是一个兜底解析器，支持所有异常类型。
// 它总是返回 500 错误，优先级最低（Order = 1000），确保在没有其他解析器匹配时也能返回响应。
type DefaultExceptionResolver struct{}

// NewDefaultExceptionResolver 创建默认解析器
//
// 返回一个 DefaultExceptionResolver 实例。
func NewDefaultExceptionResolver() ExceptionResolver {
	return &DefaultExceptionResolver{}
}

// Resolve 返回 500 错误
//
// 无论什么异常，都返回 500 Internal Server Error。
func (r *DefaultExceptionResolver) Resolve(ctx context.Context, err error) *ErrorResponse {
	return NewErrorResponse(500, "Internal Server Error", "", "", nil)
}

// Supports 支持所有异常
//
// 始终返回 true，确保这个解析器可以处理任何异常。
func (r *DefaultExceptionResolver) Supports(err error) bool {
	return true
}

// Order 最低优先级
//
// 返回 1000，确保在其他解析器都不匹配时才使用这个解析器。
func (r *DefaultExceptionResolver) Order() int {
	return 1000
}

// BuiltinExceptionResolver 内置异常解析器
//
// BuiltinExceptionResolver 处理包中定义的内置异常类型：
// - ErrNotFound → 404
// - ErrBadRequest → 400
// - ErrUnauthorized → 401
// - ErrForbidden → 403
// - ErrConflict → 409
// - ErrInternalServer → 500
type BuiltinExceptionResolver struct{}

// NewBuiltinExceptionResolver 创建内置解析器
//
// 返回一个 BuiltinExceptionResolver 实例。
func NewBuiltinExceptionResolver() ExceptionResolver {
	return &BuiltinExceptionResolver{}
}

// Resolve 根据异常类型返回对应响应
//
// 根据异常类型返回相应的 HTTP 状态码和错误消息。
func (r *BuiltinExceptionResolver) Resolve(ctx context.Context, err error) *ErrorResponse {
	switch {
	case errors.Is(err, ErrNotFound):
		return NewErrorResponse(404, "Resource not found", "", "", nil)
	case errors.Is(err, ErrBadRequest):
		return NewErrorResponse(400, "Bad request", "", "", nil)
	case errors.Is(err, ErrUnauthorized):
		return NewErrorResponse(401, "Unauthorized", "", "", nil)
	case errors.Is(err, ErrForbidden):
		return NewErrorResponse(403, "Forbidden", "", "", nil)
	case errors.Is(err, ErrConflict):
		return NewErrorResponse(409, "Conflict", "", "", nil)
	case errors.Is(err, ErrInternalServer):
		return NewErrorResponse(500, "Internal server error", "", "", nil)
	default:
		return NewErrorResponse(500, "Internal Server Error", "", "", nil)
	}
}

// Supports 支持内置异常
//
// 检查异常是否是包中定义的内置异常之一。
func (r *BuiltinExceptionResolver) Supports(err error) bool {
	return errors.Is(err, ErrNotFound) ||
		errors.Is(err, ErrBadRequest) ||
		errors.Is(err, ErrUnauthorized) ||
		errors.Is(err, ErrForbidden) ||
		errors.Is(err, ErrConflict) ||
		errors.Is(err, ErrInternalServer)
}

// Order 中等优先级
//
// 返回 100，高于默认解析器，低于自定义解析器。
func (r *BuiltinExceptionResolver) Order() int {
	return 100
}
