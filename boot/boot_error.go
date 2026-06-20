package boot

import (
	"fmt"
	"strings"
)

// BootError 结构化启动错误
//
// 包含错误发生阶段、原始错误、分析结果和修复建议，便于调试和错误处理。
//
// 设计模式: Adapter（适配原始错误为结构化格式）
type BootError struct {
	Phase       string   // 错误发生的阶段
	Original    error    // 原始错误
	Analyzed    string   // FailureAnalyzer 分析结果
	Suggestions []string // 修复建议
}

// Error 实现 error 接口
func (e *BootError) Error() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "boot failed during %s: %v", e.Phase, e.Original)

	if e.Analyzed != "" {
		fmt.Fprintf(&sb, "\n\nAnalysis: %s", e.Analyzed)
	}

	if len(e.Suggestions) > 0 {
		fmt.Fprintf(&sb, "\n\nSuggestions:")
		for _, s := range e.Suggestions {
			fmt.Fprintf(&sb, "\n  - %s", s)
		}
	}

	return sb.String()
}

// Unwrap 实现 errors.Unwrap 接口
func (e *BootError) Unwrap() error {
	return e.Original
}

// NewBootError 创建结构化启动错误
func NewBootError(phase string, err error) *BootError {
	return &BootError{
		Phase:    phase,
		Original: err,
	}
}

// WithAnalysis 添加分析结果
func (e *BootError) WithAnalysis(analysis string) *BootError {
	e.Analyzed = analysis
	return e
}

// WithSuggestions 添加修复建议
func (e *BootError) WithSuggestions(suggestions ...string) *BootError {
	e.Suggestions = suggestions
	return e
}
