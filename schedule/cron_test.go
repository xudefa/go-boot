package schedule

import (
	"testing"
	"time"
)

// TestParse_Success 验证各类合法 Cron 表达式能被正确解析
func TestParse_Success(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		expr string
	}{
		{name: "six fields", expr: "0 0 12 * * ?"},
		{name: "question marks", expr: "0 0 12 ? 6 ?"},
		{name: "step expression", expr: "*/5 * * * * ?"},
		{name: "range step", expr: "0-10/2 * * * * ?"},
		{name: "day of week only", expr: "0 0 12 ? * 1"},
		{name: "both day specs", expr: "0 0 12 15 * 5"},
		{name: "cron with ? in day", expr: "0 0 12 ? 6 ?"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := Parse(tt.expr)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", tt.expr, err)
			}
			if expr == nil {
				t.Fatal("expected non-nil expression")
			}
		})
	}
}

// TestParse_Errors 验证各类非法 Cron 表达式能正确返回错误
func TestParse_Errors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		expr string
	}{
		{name: "5 fields", expr: "0 0 12 * *"},
		{name: "invalid hour", expr: "0 0 25 * * ?"},
		{name: "step zero", expr: "0/0 * * * * ?"},
		{name: "non-numeric", expr: "a * * * * ?"},
		{name: "reversed range", expr: "10-5 * * * * ?"},
		{name: "out of bounds step start", expr: "60/5 * * * * ?"},
		{name: "mixed list step", expr: "1,invalid * * * * ?"},
		{name: "empty expression", expr: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.expr)
			if err == nil {
				t.Errorf("Parse(%q) expected error", tt.expr)
			}
		})
	}
}

// TestNext 验证 CronExpression.Next 方法在各种场景下计算下次执行时间的正确性
//
// 测试用例覆盖：固定秒、步进、范围、列表、星期、跨月、跨年、永不匹配、月份名称、星期名称、双日匹配、每秒、分钟进位、起始步进值
func TestNext(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		expr     string
		now      time.Time
		expected time.Time
		wantZero bool
	}{
		{
			name:     "simple second",
			expr:     "0 0 12 * * ?",
			now:      time.Date(2026, 5, 12, 10, 0, 0, 0, time.UTC),
			expected: time.Date(2026, 5, 12, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "step every 5 seconds",
			expr:     "0/5 * * * * ?",
			now:      time.Date(2026, 5, 12, 10, 0, 0, 0, time.UTC),
			expected: time.Date(2026, 5, 12, 10, 0, 5, 0, time.UTC),
		},
		{
			name:     "range",
			expr:     "30 0-10 9 * * ?",
			now:      time.Date(2026, 5, 12, 9, 0, 0, 0, time.UTC),
			expected: time.Date(2026, 5, 12, 9, 0, 30, 0, time.UTC),
		},
		{
			name:     "list",
			expr:     "0 0 9,12,15 * * ?",
			now:      time.Date(2026, 5, 12, 10, 0, 0, 0, time.UTC),
			expected: time.Date(2026, 5, 12, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "day of week",
			expr:     "0 0 10 ? * 1,3,5",
			now:      time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC),
		},
		{
			name:     "cross month",
			expr:     "0 0 0 1 * ?",
			now:      time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "cross year",
			expr:     "0 0 0 1 1 ?",
			now:      time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "zero never",
			expr:     "0 0 0 30 2 ?",
			now:      time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			wantZero: true,
		},
		{
			name:     "month name",
			expr:     "0 0 0 1 JAN,MAR ?",
			now:      time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "day name",
			expr:     "0 0 10 ? * MON-FRI",
			now:      time.Date(2026, 5, 16, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2026, 5, 18, 10, 0, 0, 0, time.UTC),
		},
		{
			name:     "both day specs or match",
			expr:     "0 0 12 15 * FRI",
			now:      time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "every second",
			expr:     "* * * * * ?",
			now:      time.Date(2026, 5, 12, 10, 0, 0, 0, time.UTC),
			expected: time.Date(2026, 5, 12, 10, 0, 1, 0, time.UTC),
		},
		{
			name:     "specific minute wraps hour",
			expr:     "0 30 * * * ?",
			now:      time.Date(2026, 5, 12, 10, 31, 0, 0, time.UTC),
			expected: time.Date(2026, 5, 12, 12, 30, 0, 0, time.UTC),
		},
		{
			name:     "step with start value",
			expr:     "5/10 * * * * ?",
			now:      time.Date(2026, 5, 12, 10, 0, 0, 0, time.UTC),
			expected: time.Date(2026, 5, 12, 10, 0, 5, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := Parse(tt.expr)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", tt.expr, err)
			}
			next := expr.Next(tt.now)
			if tt.wantZero {
				if !next.IsZero() {
					t.Errorf("Next(%v) = %v, want zero time", tt.now, next)
				}
				return
			}
			if !next.Equal(tt.expected) {
				t.Errorf("Next(%v) = %v, want %v", tt.now, next, tt.expected)
			}
		})
	}
}
