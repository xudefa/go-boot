package exception

import (
	"encoding/json"
	"time"
)

// NewErrorResponse 创建新的错误响应
//
// 创建一个 ErrorResponse 实例，自动设置时间戳为当前时间。
//
// 参数：
// - code: HTTP 状态码
// - message: 错误消息
// - requestID: 请求 ID（可选）
// - traceID: 追踪 ID（可选）
// - details: 详细信息（可选，可以是任意类型）
//
// 返回一个初始化好的 ErrorResponse 实例。
func NewErrorResponse(code int, message, requestID, traceID string, details interface{}) *ErrorResponse {
	return &ErrorResponse{
		Code:      code,
		Message:   message,
		RequestID: requestID,
		TraceID:   traceID,
		Details:   details,
		Timestamp: time.Now().UnixMilli(),
	}
}

// ToJSON 将错误响应转换为JSON字节
func (e *ErrorResponse) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}
