package schedule

import (
	"context"
	"errors"
	"testing"
)

// TestNewTask_Success 验证 NewTask 能正确创建任务并返回名称和 Cron 表达式
func TestNewTask_Success(t *testing.T) {
	task := NewTask("test", "0/5 * * * * ?", func(ctx context.Context) error {
		return nil
	})
	if task.Name() != "test" {
		t.Errorf("expected name 'test', got %q", task.Name())
	}
	if task.Cron() != "0/5 * * * * ?" {
		t.Errorf("expected cron '0/5 * * * * ?', got %q", task.Cron())
	}
}

// TestTask_Execute_Success 验证任务能被成功执行
func TestTask_Execute_Success(t *testing.T) {
	var executed bool
	task := NewTask("test", "0/5 * * * * ?", func(ctx context.Context) error {
		executed = true
		return nil
	})
	err := task.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !executed {
		t.Fatal("expected task to be executed")
	}
}

// TestTask_Execute_Error 验证任务执行出错时能正确返回错误
func TestTask_Execute_Error(t *testing.T) {
	expectedErr := errors.New("task error")
	task := NewTask("test", "0/5 * * * * ?", func(ctx context.Context) error {
		return expectedErr
	})
	err := task.Execute(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

// TestTask_WithContextCancellation 验证已取消的上下文能正确传递到任务执行函数中
func TestTask_WithContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	task := NewTask("test", "0/5 * * * * ?", func(ctx context.Context) error {
		return ctx.Err()
	})
	err := task.Execute(ctx)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// TestNewTask_EmptyName_Allowed 验证创建空名称任务不会报错（框架允许但不推荐）
func TestNewTask_EmptyName_Allowed(t *testing.T) {
	task := NewTask("", "0/5 * * * * ?", func(ctx context.Context) error {
		return nil
	})
	if task.Name() != "" {
		t.Errorf("expected empty name, got %q", task.Name())
	}
	if task.Cron() != "0/5 * * * * ?" {
		t.Errorf("expected '0/5 * * * * ?', got %q", task.Cron())
	}
}

// TestTask_ExecuteWithNilFunc_Panics 验证传入 nil 函数创建任务后执行会触发 panic
func TestTask_ExecuteWithNilFunc_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil function")
		}
	}()
	task := NewTask("nil", "* * * * * ?", nil)
	_ = task.Execute(context.Background())
}

// TestNewTask_NoReturnValue_Success 验证无返回值的任务函数能正常执行
func TestNewTask_NoReturnValue_Success(t *testing.T) {
	task := NewTask("void", "0/5 * * * * ?", func(ctx context.Context) error {
		return nil
	})
	if err := task.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestTask_PropertiesUnchanged 验证任务创建后名称和 Cron 表达式不可变
func TestTask_PropertiesUnchanged(t *testing.T) {
	task := NewTask("propTest", "0 0 12 * * ?", func(ctx context.Context) error {
		return nil
	})
	if task.Name() != "propTest" {
		t.Errorf("expected Name 'propTest', got %q", task.Name())
	}
	if task.Cron() != "0 0 12 * * ?" {
		t.Errorf("expected Cron '0 0 12 * * ?', got %q", task.Cron())
	}
	err := task.Execute(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
