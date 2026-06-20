// Package migrations 提供数据库迁移启动器。
package migrations

import (
	"context"
	"database/sql"

	"github.com/xudefa/go-boot/boot"
	"github.com/xudefa/go-boot/condition"
)

// MigrationStarter 数据库迁移启动器
type MigrationStarter struct {
	config MigrationConfig
	db     *sql.DB
}

// NewMigrationStarter 创建迁移启动器
func NewMigrationStarter(cfg MigrationConfig, db *sql.DB) *MigrationStarter {
	return &MigrationStarter{
		config: cfg,
		db:     db,
	}
}

// Name 返回启动器名称
func (s *MigrationStarter) Name() string {
	return "migration"
}

// Dependencies 返回依赖的启动器名称
func (s *MigrationStarter) Dependencies() []string {
	return []string{"database"}
}

// Configure 在配置阶段调用
func (s *MigrationStarter) Configure(ctx boot.ApplicationContext) error {
	// 从环境加载迁移配置
	env := ctx.Environment()
	auto := env.GetBool("db.migrations.auto", false)
	if auto {
		s.config.Auto = true
	}

	path := env.GetString("db.migrations.path", "")
	if path != "" {
		s.config.Path = path
	}

	return nil
}

// Start 在启动阶段调用
func (s *MigrationStarter) Start(ctx boot.ApplicationContext) error {
	if !s.config.Auto {
		return nil // 默认在生产环境中不自动迁移
	}

	migrator := NewSQLMigrator(s.db, s.config)

	// 应用启动时执行迁移
	return migrator.Up(context.Background())
}

// Stop 在停止阶段调用
func (s *MigrationStarter) Stop(ctx boot.ApplicationContext) error {
	return nil
}

// GetCondition 返回启动条件
func (s *MigrationStarter) GetCondition() condition.Condition {
	return condition.OnProperty("db.migrations.enabled", "true")
}

// AutoMigrate auto migrate function, can be called manually after app starts
func AutoMigrate(ctx context.Context, db *sql.DB, config MigrationConfig) error {
	migrator := NewSQLMigrator(db, config)
	return migrator.Up(ctx)
}
