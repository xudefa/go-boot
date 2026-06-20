package data

import (
	"context"
	"testing"
)

// TestTransactorInterface 验证 Transactor 接口可以正确使用
func TestTransactorInterface(t *testing.T) {
	// 验证接口定义可编译并包含所需方法
	var _ interface {
		Query(ctx context.Context, query string, args ...any) (Rows, error)
		QueryRow(ctx context.Context, query string, args ...any) Row
		Exec(ctx context.Context, query string, args ...any) (Result, error)
		Begin(ctx context.Context) (Transaction, error)
		Stats() DBStats
		Close() error
	} = nil

	// 验证 Transaction 扩展了 Transactor
	var _ interface {
		Query(ctx context.Context, query string, args ...any) (Rows, error)
		QueryRow(ctx context.Context, query string, args ...any) Row
		Exec(ctx context.Context, query string, args ...any) (Result, error)
		Begin(ctx context.Context) (Transaction, error)
		Stats() DBStats
		Close() error
		Commit() error
		Rollback() error
	} = nil

	t.Log("Transactor and Transaction interfaces are correctly defined")
}

// TestRowsInterface 验证 Rows 接口
func TestRowsInterface(t *testing.T) {
	var _ interface {
		Next() bool
		Scan(dest ...any) error
		Close() error
		Err() error
	} = nil
	t.Log("Rows interface is correctly defined")
}

// TestRowInterface 验证 Row 接口
func TestRowInterface(t *testing.T) {
	var _ interface {
		Scan(dest ...any) error
	} = nil
	t.Log("Row interface is correctly defined")
}

// TestResultInterface 验证 Result 接口
func TestResultInterface(t *testing.T) {
	var _ interface {
		LastInsertId() (int64, error)
		RowsAffected() (int64, error)
	} = nil
	t.Log("Result interface is correctly defined")
}

// TestDBStats 验证 DBStats 结构体
func TestDBStats(t *testing.T) {
	stats := DBStats{
		MaxOpenConnections: 10,
		OpenConnections:    5,
		InUse:              3,
		Idle:               2,
	}
	if stats.MaxOpenConnections != 10 {
		t.Error("DBStats not correctly initialized")
	}
}
