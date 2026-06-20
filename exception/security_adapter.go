package exception

import (
	"context"
	"encoding/json"
	"fmt"
)

// AccessDeniedHandlerAdapter 适配 AccessDeniedHandler 到异常处理器
//
// AccessDeniedHandlerAdapter 是一个适配器，将安全模块的 AccessDeniedHandler 接口适配到异常处理器。
// 这样可以将访问拒绝异常统一通过异常处理器处理，实现一致的错误响应格式。
type AccessDeniedHandlerAdapter struct {
	handler ExceptionHandler
}

// NewAccessDeniedHandlerAdapter 创建 AccessDeniedHandler 适配器
//
// 返回一个 AccessDeniedHandlerAdapter 实例，该实例会使用提供的异常处理器处理异常。
func NewAccessDeniedHandlerAdapter(handler ExceptionHandler) *AccessDeniedHandlerAdapter {
	return &AccessDeniedHandlerAdapter{
		handler: handler,
	}
}

// Handle 处理访问拒绝异常
//
// 当访问被拒绝时，调用此方法处理异常。
// 如果异常不为 nil，会通过异常处理器处理，并将响应写入 ResponseWriter。
func (a *AccessDeniedHandlerAdapter) Handle(response ResponseWriter, err error) {
	if err == nil {
		return
	}

	ctx := context.Background()
	resp := a.handler.Handle(ctx, err, response)

	if resp != nil {
		response.SetStatusCode(resp.Code)
		response.SetHeader("Content-Type", "application/json")
		data, marshalErr := json.Marshal(resp)
		if marshalErr != nil {
			fmt.Printf("[go-boot] failed to marshal access denied response: %v\n", marshalErr)
		}
		if writeErr := response.Write(data); writeErr != nil {
			fmt.Printf("[go-boot] failed to write access denied response: %v\n", writeErr)
		}
	}
}

// AuthenticationEntryPointAdapter 适配 AuthenticationEntryPoint 到异常处理器
//
// AuthenticationEntryPointAdapter 是一个适配器，将安全模块的 AuthenticationEntryPoint 接口适配到异常处理器。
// 这样可以将认证入口点异常统一通过异常处理器处理，实现一致的错误响应格式。
type AuthenticationEntryPointAdapter struct {
	handler ExceptionHandler
}

// NewAuthenticationEntryPointAdapter 创建 AuthenticationEntryPoint 适配器
//
// 返回一个 AuthenticationEntryPointAdapter 实例，该实例会使用提供的异常处理器处理异常。
func NewAuthenticationEntryPointAdapter(handler ExceptionHandler) *AuthenticationEntryPointAdapter {
	return &AuthenticationEntryPointAdapter{
		handler: handler,
	}
}

// Commence 处理认证入口点异常
//
// 当认证失败时，调用此方法处理异常。
// 如果异常不为 nil，会通过异常处理器处理，并将响应写入 ResponseWriter。
func (a *AuthenticationEntryPointAdapter) Commence(response ResponseWriter, err error) {
	if err == nil {
		return
	}

	ctx := context.Background()
	resp := a.handler.Handle(ctx, err, response)

	if resp != nil {
		response.SetStatusCode(resp.Code)
		response.SetHeader("Content-Type", "application/json")
		data, marshalErr := json.Marshal(resp)
		if marshalErr != nil {
			fmt.Printf("[go-boot] failed to marshal authentication entry point response: %v\n", marshalErr)
		}
		if writeErr := response.Write(data); writeErr != nil {
			fmt.Printf("[go-boot] failed to write authentication entry point response: %v\n", writeErr)
		}
	}
}
