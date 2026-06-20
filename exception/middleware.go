package exception

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ExceptionHandlingMiddleware HTTP 异常处理中间件
//
// ExceptionHandlingMiddleware 是一个 HTTP 中间件，用于自动捕获和处理 HTTP 处理过程中的异常。
// 它会捕获 panic 和返回的异常，并通过 ExceptionHandler 处理它们。
//
// 使用示例：
//
//	handler := exception.NewDefaultExceptionHandler()
//	middleware := exception.ExceptionHandlingMiddleware(handler)
//	http.Handle("/", middleware(httpHandler))
func ExceptionHandlingMiddleware(handler ExceptionHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					err := fmt.Errorf("panic: %v", rec)
					resp := handler.Handle(r.Context(), err, &httpResponseWriter{w})
					if resp != nil {
						w.WriteHeader(resp.Code)
						if encErr := json.NewEncoder(w).Encode(resp); encErr != nil {
							fmt.Printf("[go-boot] failed to encode exception response: %v\n", encErr)
						}
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// httpResponseWriter 适配 http.ResponseWriter 到 ResponseWriter 接口
//
// httpResponseWriter 是一个适配器，将标准库的 http.ResponseWriter 适配到包中的 ResponseWriter 接口。
// 这使得异常处理器可以与标准 HTTP 处理器无缝集成。
type httpResponseWriter struct {
	http.ResponseWriter
}

// SetStatusCode 设置 HTTP 状态码
//
// 调用底层 http.ResponseWriter 的 WriteHeader 方法。
func (w *httpResponseWriter) SetStatusCode(code int) {
	w.WriteHeader(code)
}

// SetHeader 设置 HTTP 头
//
// 调用底层 http.ResponseWriter 的 Header().Set 方法。
func (w *httpResponseWriter) SetHeader(key, value string) {
	w.Header().Set(key, value)
}

// Write 写入响应体
//
// 调用底层 http.ResponseWriter 的 Write 方法。
func (w *httpResponseWriter) Write(data []byte) error {
	_, err := w.ResponseWriter.Write(data)
	return err
}
