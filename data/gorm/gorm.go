// Package gorm 提供基于 GORM 的数据库访问层实现
package gorm

import (
	"fmt"

	"github.com/xudefa/go-boot/data"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GormClient GORM 数据库客户端
//
// 创建方式：
//
//	client, _ := gorm.NewGormClient(cfg)
//	client, _ := gorm.NewDefaultGormClient("user", "pass", "database")
type GormClient struct {
	DB *gorm.DB
}

var _ data.RepositoryFactory[any] = (*GormClient)(nil)
var _ data.Transaction = (*GormTransaction)(nil)

// GormTransaction GORM 事务
//
// 通过 Client.Begin() 创建，提供事务控制方法
type GormTransaction struct {
	DB *gorm.DB
}

// NewGormClient 创建 GORM 客户端
//
// 使用 DatabaseConfig 配置连接：
//
//	cfg := &data.DatabaseConfig{
//	    Driver:   "mysql",
//	    Host:     "localhost",
//	    Port:     3306,
//	    Username: "user",
//	    Password: "pass",
//	    Name:     "database",
//	}
//	client, _ := gorm.NewGormClient(cfg)
func NewGormClient(cfg *data.DatabaseConfig) (*GormClient, error) {
	dsn := cfg.DSN()
	if dsn == "" {
		return nil, fmt.Errorf("unsupported driver: %s", cfg.Driver)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetConnMaxLifetime(cfg.MaxLife)
	sqlDB.SetConnMaxIdleTime(cfg.MaxIdleTime)

	if !cfg.Debug {
		db.Config.Logger = logger.Default.LogMode(logger.Warn)
	}

	return &GormClient{DB: db}, nil
}

// NewDefaultGormClient 创建 GORM 客户端（使用默认配置）
//
// 默认使用 MySQL，localhost:3306，utf8mb4 编码
func NewDefaultGormClient(username, password, database string) (*GormClient, error) {
	dsn := data.NewDefaultDatabaseConfig(username, password, database).DSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}
	sqlDB.SetMaxOpenConns(1000)
	return &GormClient{DB: db}, nil
}

// NewRepository 创建 Repository 实例
//
//	repo := client.NewRepository().(data.Repository[User])
func (c *GormClient) NewRepository() data.Repository[any] {
	return NewBaseRepository[any](c.DB)
}

// Begin 开启事务
//
//	tx, _ := client.Begin()
//	defer tx.Close()
//	tx.Commit()
func (c *GormClient) Begin() (data.Transaction, error) {
	tx := c.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &GormTransaction{DB: tx}, nil
}

// Commit 提交事务
func (t *GormTransaction) Commit() error {
	return t.DB.Commit().Error
}

// Rollback 回滚事务
func (t *GormTransaction) Rollback() error {
	return t.DB.Rollback().Error
}

// Close 关闭事务（自动回滚未提交的事务）
func (t *GormTransaction) Close() {
	if err := t.DB.Error; err != nil && err != gorm.ErrRecordNotFound {
		t.DB.Rollback()
	}
}

// Tx 获取事务的 *gorm.DB
func (t *GormTransaction) Tx() *gorm.DB {
	return t.DB
}
