// Package schedule 定时任务调度框架的边缘情况测试
package schedule

import (
	"context"
	"testing"
	"time"
)

func TestScheduler_RegisterInvalidCron(t *testing.T) {
	s := NewScheduler()
	t.Cleanup(func() { _ = s.Shutdown(context.Background()) })

	task := NewTask("invalid", "invalid cron", func(ctx context.Context) error {
		return nil
	})

	err := s.Register(task)
	if err == nil {
		t.Fatal("expected error for invalid cron")
	}
}

func TestScheduler_RegisterNeverFireTask(t *testing.T) {
	s := NewScheduler()
	t.Cleanup(func() { _ = s.Shutdown(context.Background()) })

	task := NewTask("never", "0 0 0 30 2 ?", func(ctx context.Context) error {
		return nil
	})

	err := s.Register(task)
	if err == nil {
		t.Fatal("expected error for task that will never fire")
	}
}

func TestScheduler_ExecuteTaskError(t *testing.T) {
	s := NewScheduler()
	ctx := context.Background()
	t.Cleanup(func() { _ = s.Shutdown(ctx) })

	errCount := 0
	task := NewTask("error", "* * * * * ?", func(ctx context.Context) error {
		errCount++
		return context.DeadlineExceeded
	})

	if err := s.Register(task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Start(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	time.Sleep(1200 * time.Millisecond)

	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if errCount == 0 {
		t.Fatal("expected task to be executed and return error")
	}
}

func TestCronExpression_Next_MaxSearchRange(t *testing.T) {
	expr, err := Parse("* * * * * ?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Use a far future time that would exceed 4-year search range
	farFuture := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	next := expr.Next(farFuture)

	if next.IsZero() {
		t.Fatal("expected non-zero time for valid cron expression")
	}
}

func TestCronExpression_Parse_EmptyExpression(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Fatal("expected error for empty expression")
	}
}

func TestCronExpression_Parse_TooManyFields(t *testing.T) {
	_, err := Parse("* * * * * * extra")
	if err == nil {
		t.Fatal("expected error for too many fields")
	}
}

func TestCronExpression_Parse_InvalidRange(t *testing.T) {
	_, err := Parse("* * * 50 * ?") // day of month 50 is invalid
	if err == nil {
		t.Fatal("expected error for invalid range")
	}
}

func TestCronExpression_Parse_InvalidStep(t *testing.T) {
	_, err := Parse("* * * * */0 ?") // step of 0 is invalid
	if err == nil {
		t.Fatal("expected error for invalid step")
	}
}

func TestCronExpression_Parse_InvalidMonth(t *testing.T) {
	_, err := Parse("* * * * INVALID ?") // invalid month
	if err == nil {
		t.Fatal("expected error for invalid month")
	}
}

func TestCronExpression_Parse_InvalidDayOfWeek(t *testing.T) {
	_, err := Parse("* * * * * INVALID") // invalid day of week
	if err == nil {
		t.Fatal("expected error for invalid day of week")
	}
}

func TestCronExpression_Parse_InvalidValue(t *testing.T) {
	_, err := Parse("INVALID * * * * ?") // invalid value
	if err == nil {
		t.Fatal("expected error for invalid value")
	}
}
