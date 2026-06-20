// Package migrations 提供数据库迁移功能。
package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Migrator 定义数据库迁移接口
type Migrator interface {
	Up(ctx context.Context) error
	Down(ctx context.Context, steps int) error
	Status(ctx context.Context) ([]MigrationStatus, error)
	Version(ctx context.Context) (int, error)
	SetVersion(ctx context.Context, version int, dirty bool) error
}

// MigrationStatus 表示迁移状态
type MigrationStatus struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Applied bool   `json:"applied"`
	Error   string `json:"error,omitempty"`
}

// MigrationConfig 迁移配置
type MigrationConfig struct {
	Path      string // 迁移文件路径
	TableName string // 迁移记录表名，默认：migrations
	Auto      bool   // 是否自动迁移，默认：false
}

// SQLMigrator 基于SQL的迁移器
type SQLMigrator struct {
	db     *sql.DB
	config MigrationConfig
}

// NewSQLMigrator 创建新的SQL迁移器
func NewSQLMigrator(db *sql.DB, config MigrationConfig) *SQLMigrator {
	if config.TableName == "" {
		config.TableName = "migrations"
	}
	return &SQLMigrator{
		db:     db,
		config: config,
	}
}

// Up 应用未应用的迁移
func (m *SQLMigrator) Up(ctx context.Context) error {
	if err := m.ensureMigrationTable(ctx); err != nil {
		return err
	}

	files, err := m.getMigrationFiles()
	if err != nil {
		return err
	}

	currentVersion, err := m.Version(ctx)
	if err != nil {
		return err
	}

	for _, file := range files {
		id, _, err := parseMigrationFilename(file)
		if err != nil {
			continue
		}

		if id <= currentVersion {
			continue
		}

		if err := m.applyMigration(ctx, file, id); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", file, err)
		}

		if err := m.SetVersion(ctx, id, false); err != nil {
			return err
		}
	}

	return nil
}

// Down 回滚指定数量的迁移
func (m *SQLMigrator) Down(ctx context.Context, steps int) error {
	if steps <= 0 {
		return nil
	}

	currentVersion, err := m.Version(ctx)
	if err != nil {
		return err
	}

	files, err := m.getMigrationFiles()
	if err != nil {
		return err
	}

	// 收集当前已应用的迁移（版本 <= currentVersion）
	appliedMigrations := []string{}
	for _, file := range files {
		id, _, err := parseMigrationFilename(file)
		if err != nil {
			continue
		}
		if id <= currentVersion {
			appliedMigrations = append(appliedMigrations, file)
		}
	}

	// 回滚最后N个迁移（从最高到最低）
	rollbackCount := 0
	for i := len(appliedMigrations) - 1; i >= 0 && rollbackCount < steps; i-- {
		file := appliedMigrations[i]
		id, _, err := parseMigrationFilename(file)
		if err != nil {
			continue
		}

		if err := m.rollbackMigration(ctx, file, id); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", file, err)
		}

		// 将版本更新为此迁移的前一个版本
		newVersion := id - 1
		if newVersion < 0 {
			newVersion = 0
		}
		if err := m.SetVersion(ctx, newVersion, false); err != nil {
			return err
		}
		rollbackCount++
	}

	return nil
}

// Status 获取迁移状态
func (m *SQLMigrator) Status(ctx context.Context) ([]MigrationStatus, error) {
	files, err := m.getMigrationFiles()
	if err != nil {
		return nil, err
	}

	currentVersion, err := m.Version(ctx)
	if err != nil {
		return nil, err
	}

	var statuses []MigrationStatus
	for _, file := range files {
		id, name, err := parseMigrationFilename(file)
		if err != nil {
			continue
		}

		statuses = append(statuses, MigrationStatus{
			ID:      id,
			Name:    name,
			Applied: id <= currentVersion,
		})
	}

	return statuses, nil
}

// Version 获取当前迁移版本
func (m *SQLMigrator) Version(ctx context.Context) (int, error) {
	query := fmt.Sprintf("SELECT COALESCE(MAX(version), 0) FROM %s", m.config.TableName)
	var version int
	err := m.db.QueryRowContext(ctx, query).Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// SetVersion 设置当前迁移版本
func (m *SQLMigrator) SetVersion(ctx context.Context, version int, dirty bool) error {
	// 首先删除现有记录
	deleteQuery := fmt.Sprintf("DELETE FROM %s", m.config.TableName)
	_, err := m.db.ExecContext(ctx, deleteQuery)
	if err != nil {
		return err
	}

	// 插入新版本记录
	insertQuery := fmt.Sprintf("INSERT INTO %s (version, dirty) VALUES (?, ?)", m.config.TableName)
	_, err = m.db.ExecContext(ctx, insertQuery, version, dirty)
	return err
}

// ensureMigrationTable 确保迁移表存在
func (m *SQLMigrator) ensureMigrationTable(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY DEFAULT 1,
			version INTEGER NOT NULL DEFAULT 0,
			dirty BOOLEAN NOT NULL DEFAULT FALSE,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_%s_version ON %s(version);
	`, m.config.TableName, m.config.TableName, m.config.TableName)
	_, err := m.db.ExecContext(ctx, query)
	return err
}

// getMigrationFiles gets migration file list
func (m *SQLMigrator) getMigrationFiles() ([]string, error) {
	var files []string
	err := filepath.Walk(m.config.Path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".sql") && !strings.Contains(path, ".down.sql") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Sort by version number
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			idI, _, _ := parseMigrationFilename(files[i])
			idJ, _, _ := parseMigrationFilename(files[j])
			if idI > idJ {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	return files, nil
}

// applyMigration applies a single migration
func (m *SQLMigrator) applyMigration(ctx context.Context, filePath string, id int) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	query := string(content)
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && rbErr != sql.ErrTxDone {
			fmt.Printf("[go-boot] failed to rollback migration transaction: %v\n", rbErr)
		}
	}()

	if _, err := tx.ExecContext(ctx, query); err != nil {
		return err
	}

	return tx.Commit()
}

// rollbackMigration rolls back a single migration
func (m *SQLMigrator) rollbackMigration(ctx context.Context, filePath string, id int) error {
	rollbackPath := strings.TrimSuffix(filePath, ".sql") + ".down.sql"

	if _, err := os.Stat(rollbackPath); err == nil {
		content, err := os.ReadFile(rollbackPath)
		if err != nil {
			return err
		}

		query := string(content)
		tx, err := m.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() {
			if rbErr := tx.Rollback(); rbErr != nil && rbErr != sql.ErrTxDone {
				fmt.Printf("[go-boot] failed to rollback migration transaction: %v\n", rbErr)
			}
		}()

		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}

		return tx.Commit()
	} else {
		return fmt.Errorf("rollback file not found: %s", rollbackPath)
	}
}

// parseMigrationFilename parses migration filename (format: 001_create_users_table.sql)
func parseMigrationFilename(filename string) (int, string, error) {
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	parts := strings.SplitN(name, "_", 2)
	if len(parts) < 2 {
		return 0, "", fmt.Errorf("invalid migration filename: %s", filename)
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", fmt.Errorf("invalid migration id in filename: %s", filename)
	}

	return id, parts[1], nil
}
