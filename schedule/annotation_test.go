package schedule

import (
	"testing"
)

// TestExtractCronFromAnnotation 测试从 @Scheduled 注解文本中提取 cron 表达式的各种情况
//
// 测试用例覆盖：双引号、单引号、多参数、无 cron 参数、空 cron、无括号、带逗号格式
func TestExtractCronFromAnnotation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "single param with double quotes", input: `@Scheduled(cron="0/5 * * * * ?")`, expected: "0/5 * * * * ?"},
		{name: "single quote", input: `@Scheduled(cron='0 0 12 * * ?')`, expected: "0 0 12 * * ?"},
		{name: "multi param", input: `@Scheduled(cron="0/5 * * * * ?", zone="UTC")`, expected: "0/5 * * * * ?"},
		{name: "no cron param", input: `@Scheduled(zone="UTC")`, expected: ""},
		{name: "empty cron", input: `@Scheduled(cron="")`, expected: ""},
		{name: "no parens", input: `@Scheduled`, expected: ""},
		{name: "cron with comma", input: `@Scheduled(cron="0,15,30,45 * * * * ?")`, expected: "0,15,30,45 * * * * ?"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractCronFromAnnotation(tt.input)
			if result != tt.expected {
				t.Errorf("extractCronFromAnnotation(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestToFirstCharLower 测试首字母转小写工具的准确性
//
// 测试用例覆盖：完整字符串、空字符串、单字符
func TestToFirstCharLower(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "service name", input: "MyService", expected: "myService"},
		{name: "empty string", input: "", expected: ""},
		{name: "single char", input: "A", expected: "a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toFirstCharLower(tt.input)
			if result != tt.expected {
				t.Errorf("toFirstCharLower(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
