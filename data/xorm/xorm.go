// Package xorm 提供基于 XORM 的数据库访问层实现
package xorm

import (
	"fmt"
	"time"

	"github.com/xudefa/go-boot/data"
	"xorm.io/xorm"
)

// XormClient XORM 数据库客户端
//
// 创建方式：
//
//	client, _ := xorm.NewXormClient(cfg)
//	client, _ := xorm.NewDefaultXormClient("user", "pass", "database")
type XormClient struct {
	Engine *xorm.Engine
}

// XormTransaction XORM 事务
//
// 通过 Client.Begin() 创建，提供事务控制方法
type XormTransaction struct {
	session *xorm.Session
}

// NewXormClient 创建 XORM 客户端
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
//	client, _ := xorm.NewXormClient(cfg)
func NewXormClient(cfg *data.DatabaseConfig) (*XormClient, error) {
	dsn := cfg.DSN()
	if dsn == "" {
		return nil, fmt.Errorf("unsupported driver: %s", cfg.Driver)
	}

	engine, err := xorm.NewEngine(cfg.Driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create engine: %w", err)
	}

	engine.SetMaxOpenConns(cfg.MaxOpen)
	engine.SetMaxIdleConns(cfg.MaxIdle)
	engine.SetConnMaxLifetime(cfg.MaxLife)

	if cfg.Debug {
		engine.ShowSQL(true)
	}

	return &XormClient{Engine: engine}, nil
}

// NewDefaultXormClient 创建 XORM 客户端（使用默认配置）
//
// 默认使用 MySQL，localhost:3306，utf8mb4 编码
func NewDefaultXormClient(username, password, database string) (*XormClient, error) {
	cfg := data.NewDefaultDatabaseConfig(username, password, database)
	cfg.MaxOpen = 1000
	engine, err := xorm.NewEngine(cfg.Driver, cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to create engine: %w", err)
	}

	engine.SetMaxOpenConns(cfg.MaxOpen)
	engine.SetMaxIdleConns(cfg.MaxIdle)
	engine.SetConnMaxLifetime(time.Hour)

	return &XormClient{Engine: engine}, nil
}

// NewSession 创建新的数据库会话
//
// 与事务不同，用于批量操作等场景
func (c *XormClient) NewSession() *xorm.Session {
	return c.Engine.NewSession()
}

// NewRepository 创建 Repository 实例
func (c *XormClient) NewRepository() data.Repository[any] {
	return NewBaseRepository[any](c.Engine)
}

// Begin 开启事务
//
//	tx, _ := client.Begin()
//	defer tx.Close()
//	tx.Commit()
func (c *XormClient) Begin() (data.Transaction, error) {
	session := c.Engine.NewSession()
	err := session.Begin()
	if err != nil {
		return nil, err
	}
	return &XormTransaction{session: session}, nil
}

// Commit 提交事务
func (t *XormTransaction) Commit() error {
	return t.session.Commit()
}

// Rollback 回滚事务
func (t *XormTransaction) Rollback() error {
	return t.session.Rollback()
}

// Close 关闭事务（自动回滚未提交的事务）
func (t *XormTransaction) Close() {
	err := t.session.Close()
	if err != nil {
		return
	}
}

// Session 获取事务的 *xorm.Session
func (t *XormTransaction) Session() *xorm.Session {
	return t.session
}

func (s *XormTransaction) Insert(bean any) (int64, error) {
	return s.session.Insert(bean)
}

func (s *XormTransaction) InsertBatch(beans []any) error {
	if len(beans) == 0 {
		return nil
	}
	_, err := s.session.Insert(&beans)
	return err
}

func (s *XormTransaction) Delete(bean any) (int64, error) {
	return s.session.Delete(bean)
}

func (s *XormTransaction) Update(bean any) (int64, error) {
	return s.session.Update(bean)
}

func (s *XormTransaction) Find(bean any) error {
	return s.session.Find(bean)
}

func (s *XormTransaction) Get(bean any) (bool, error) {
	return s.session.Get(bean)
}

func (s *XormTransaction) Count(bean any) (int64, error) {
	return s.session.Count(bean)
}
