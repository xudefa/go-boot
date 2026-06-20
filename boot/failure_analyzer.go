package boot

import (
	"fmt"
	"strings"
	"sync"
)

// FailureReport 失败报告
//
// 参考 Spring Boot 的 FailureAnalysis，在应用启动失败时提供结构化的错误信息。
// 包含错误描述、建议动作、根因和可能的解决方案。
type FailureReport struct {
	Headline          string         `json:"headline"`                    // 报告标题
	Description       string         `json:"description"`                 // 错误描述
	Action            string         `json:"action"`                      // 建议动作
	Cause             string         `json:"cause"`                       // 根因
	Details           map[string]any `json:"details,omitempty"`           // 附加详情
	StackTrace        string         `json:"stackTrace,omitempty"`        // 堆栈跟踪
	PossibleSolutions []string       `json:"possibleSolutions,omitempty"` // 可能的解决方案列表
}

// FailureAnalyzer 失败分析器接口
//
// 参考 Spring Boot 的 FailureAnalyzer。
// 在应用启动失败时提供友好的错误提示，帮助开发者快速定位问题。
type FailureAnalyzer interface {
	// CanAnalyze 检查是否能分析该错误
	CanAnalyze(err error) bool

	// Analyze 分析错误并返回失败报告
	Analyze(err error) *FailureReport
}

// SimpleFailureAnalyzer 简单的失败分析器
//
// 通过传入的分析函数创建分析器，适用于简单的错误分析场景。
type SimpleFailureAnalyzer struct {
	analyzeFn func(err error) *FailureReport
}

// NewSimpleFailureAnalyzer 创建简单失败分析器
//
// 参数：
//   - analyzeFn: 分析函数，接收错误返回失败报告，返回 nil 表示无法分析
func NewSimpleFailureAnalyzer(analyzeFn func(err error) *FailureReport) *SimpleFailureAnalyzer {
	return &SimpleFailureAnalyzer{analyzeFn: analyzeFn}
}

// CanAnalyze 检查分析函数是否能处理该错误
func (s *SimpleFailureAnalyzer) CanAnalyze(err error) bool {
	return s.analyzeFn(err) != nil
}

// Analyze 使用分析函数分析错误
func (s *SimpleFailureAnalyzer) Analyze(err error) *FailureReport {
	return s.analyzeFn(err)
}

// FailureAnalyzerRegistry 失败分析器注册表
//
// 管理所有 FailureAnalyzer 的注册和查询。
// 分析时按注册顺序遍历，返回第一个匹配的失败报告。
type FailureAnalyzerRegistry struct {
	mu        sync.RWMutex
	analyzers []FailureAnalyzer
}

var globalAnalyzerRegistry = NewFailureAnalyzerRegistry()

// NewFailureAnalyzerRegistry 创建失败分析器注册表
func NewFailureAnalyzerRegistry() *FailureAnalyzerRegistry {
	return &FailureAnalyzerRegistry{}
}

// Register 注册失败分析器
func (r *FailureAnalyzerRegistry) Register(analyzer FailureAnalyzer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.analyzers = append(r.analyzers, analyzer)
}

// RegisterFailureAnalyzer 注册到全局注册表
//
// 兼容旧版接口，包装为 compatFailureAnalyzer。
func RegisterFailureAnalyzer(analyzer interface {
	Analyze(err error) *FailureReport
}) {
	globalAnalyzerRegistry.Register(&compatFailureAnalyzer{inner: analyzer})
}

// compatFailureAnalyzer 兼容旧版 FailureAnalyzer 的适配器
type compatFailureAnalyzer struct {
	inner interface {
		Analyze(err error) *FailureReport
	}
}

func (c *compatFailureAnalyzer) CanAnalyze(err error) bool {
	return c.inner.Analyze(err) != nil
}

func (c *compatFailureAnalyzer) Analyze(err error) *FailureReport {
	return c.inner.Analyze(err)
}

// Analyze 分析错误，返回第一个匹配的失败报告
func (r *FailureAnalyzerRegistry) Analyze(err error) *FailureReport {
	r.mu.RLock()
	analyzers := make([]FailureAnalyzer, len(r.analyzers))
	copy(analyzers, r.analyzers)
	r.mu.RUnlock()

	for _, analyzer := range analyzers {
		if analyzer.CanAnalyze(err) {
			return analyzer.Analyze(err)
		}
	}
	return nil
}

// GlobalAnalyzerRegistry 返回全局失败分析器注册表
func GlobalAnalyzerRegistry() *FailureAnalyzerRegistry {
	return globalAnalyzerRegistry
}

// formatFailure 格式化失败报告为可读字符串
func formatFailure(report *FailureReport) string {
	var result strings.Builder
	fmt.Fprintf(&result, `
====================
APPLICATION FAILED TO START
====================

描述: %s

动作: %s

原因: %s
`, report.Description, report.Action, report.Cause)
	if len(report.PossibleSolutions) > 0 {
		result.WriteString("\n可能的解决方案:\n")
		for i, sol := range report.PossibleSolutions {
			fmt.Fprintf(&result, "  %d. %s\n", i+1, sol)
		}
	}
	return result.String()
}

// ReportFailure 分析并格式化输出失败报告
//
// 如果没有匹配的分析器，返回简单的错误信息字符串。
func ReportFailure(err error) string {
	report := globalAnalyzerRegistry.Analyze(err)
	if report == nil {
		return fmt.Sprintf("Error: %s", err)
	}
	return formatFailure(report)
}
