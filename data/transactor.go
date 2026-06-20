// Package data 定义数据库操作的核心接口。
//
// 该包提供数据库访问的抽象层，支持不同的 ORM 实现
// (如 GORM、XORM) 而不需要修改使用方的代码。
//
// 主要接口：
//
//   - Transactor: 数据库事务操作接口
//   - Transaction: 事务接口，扩展 Transactor
//   - Rows: 查询结果集接口
//   - Row: 单行结果接口
//   - Result: 执行结果接口
//   - DBStats: 数据库统计信息
package data

import "context"

// Transactor 数据库事务操作接口
//
// 提供查询、执行和事务管理的基本操作。
// 各 ORM 实现（如 GORM、XORM）需实现此接口以接入 go-boot 数据访问层。
type Transactor interface {
	// Query 执行查询并返回多行结果集
	Query(ctx context.Context, query string, args ...any) (Rows, error)

	// QueryRow 执行查询并返回单行结果
	QueryRow(ctx context.Context, query string, args ...any) Row

	// Exec 执行写操作（INSERT/UPDATE/DELETE）并返回执行结果
	Exec(ctx context.Context, query string, args ...any) (Result, error)

	// Begin 开启一个新事务
	Begin(ctx context.Context) (Transaction, error)

	// Stats 返回数据库连接池统计信息
	Stats() DBStats

	// Close 关闭数据库连接并释放资源
	Close() error
}

// Transaction 事务接口，扩展 Transactor
//
// 在 Transactor 基础上增加 Commit 和 Rollback 操作，
// 支持显式的事务提交和回滚。
type Transaction interface {
	Transactor
	// Commit 提交当前事务
	Commit() error
	// Rollback 回滚当前事务
	Rollback() error
}

// Rows 查询结果集接口
//
// 用于遍历多行查询结果，使用模式：
//
//	rows, _ := transactor.Query(ctx, "SELECT * FROM users")
//	defer rows.Close()
//	for rows.Next() {
//	    var user User
//	    rows.Scan(&user.ID, &user.Name)
//	}
type Rows interface {
	// Next 推进到下一行，返回是否还有更多行
	Next() bool
	// Scan 将当前行的列值扫描到目标变量
	Scan(dest ...any) error
	// Close 释放结果集资源
	Close() error
	// Err 返回遍历过程中遇到的错误
	Err() error
}

// Row 单行结果接口
//
// 用于单行查询结果的扫描。
type Row interface {
	// Scan 将行的列值扫描到目标变量
	Scan(dest ...any) error
}

// Result 执行结果接口
//
// 提供 INSERT/UPDATE/DELETE 操作的执行信息。
type Result interface {
	// LastInsertId 返回 INSERT 操作生成的自增 ID
	LastInsertId() (int64, error)
	// RowsAffected 返回受影响的行数
	RowsAffected() (int64, error)
}

// DBStats 数据库连接池统计信息
type DBStats struct {
	MaxOpenConnections int // 最大打开连接数
	OpenConnections    int // 当前打开的连接数
	InUse              int // 正在使用的连接数
	Idle               int // 空闲的连接数
}
