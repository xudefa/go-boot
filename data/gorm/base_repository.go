// Package gorm 提供基于 GORM 的数据库访问层实现
package gorm

import (
	"github.com/xudefa/go-boot/data"
	"gorm.io/gorm"
)

// BaseRepository 泛型 Repository 实现
//
// 支持事务模式，通过 NewBaseRepositoryWithTransaction 创建
type BaseRepository[T any] struct {
	db    *gorm.DB
	tx    *gorm.DB
	table string
}

// NewBaseRepository 创建 Repository 实例
//
//	repo := gorm.NewBaseRepository[User](client.DB)
func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{
		db:    db,
		table: data.GetTableName[T](),
	}
}

// NewBaseRepositoryWithTransaction 创建基于事务的 Repository 实例
//
//	txRepo := gorm.NewBaseRepositoryWithTransaction[User](tx.(*gorm.GormTransaction).Tx())
func NewBaseRepositoryWithTransaction[T any](tx *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{
		tx:    tx,
		table: data.GetTableName[T](),
	}
}

func (r *BaseRepository[T]) engine() *gorm.DB {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// Create 创建单条记录
func (r *BaseRepository[T]) Create(entity *T) error {
	return r.engine().Create(entity).Error
}

// CreateBatch 批量创建记录
func (r *BaseRepository[T]) CreateBatch(entities []T) error {
	return r.engine().Create(&entities).Error
}

// Delete 根据 ID 删除记录
func (r *BaseRepository[T]) Delete(id interface{}) error {
	return r.engine().Table(r.table).Where("id = ?", id).Delete(new(T)).Error
}

// DeleteByCondition 根据条件删除记录
func (r *BaseRepository[T]) DeleteByCondition(where interface{}, args ...interface{}) error {
	return r.engine().Table(r.table).Where(where, args...).Delete(new(T)).Error
}

// Update 更新记录（需包含 ID）
// 只更新非零值字段，以符合 GORM 的只更新 changed 字段的行为
func (r *BaseRepository[T]) Update(entity *T) error {
	return r.engine().Table(r.table).Updates(entity).Error
}

// UpdateByCondition 根据条件更新记录
//
//	rows, _ := repo.UpdateByCondition("name = ?", "John", map[string]any{"age": 30})
func (r *BaseRepository[T]) UpdateByCondition(where interface{}, args ...interface{}) (int64, error) {
	if len(args) == 0 {
		return 0, nil
	}
	result := r.engine().Table(r.table).Where(where, args[:len(args)-1]...).Updates(args[len(args)-1].(map[string]interface{}))
	return result.RowsAffected, result.Error
}

// FindByID 根据 ID 查询单条记录
func (r *BaseRepository[T]) FindByID(id interface{}) (*T, error) {
	var entity T
	err := r.engine().Table(r.table).Where("id = ?", id).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// FindOne 根据条件查询单条记录
func (r *BaseRepository[T]) FindOne(where interface{}, args ...interface{}) (*T, error) {
	var entity T
	err := r.engine().Table(r.table).Where(where, args...).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// FindAll 根据条件查询多条记录
func (r *BaseRepository[T]) FindAll(where interface{}, args ...interface{}) ([]T, error) {
	var entities []T
	err := r.engine().Table(r.table).Where(where, args...).Find(&entities).Error
	return entities, err
}

// Count 统计记录数量
func (r *BaseRepository[T]) Count(where interface{}, args ...interface{}) (int64, error) {
	var count int64
	err := r.engine().Table(r.table).Model(new(T)).Where(where, args...).Count(&count).Error
	return count, err
}

// Raw 执行原生 SQL 查询
func (r *BaseRepository[T]) Raw(sql string, args ...interface{}) ([]T, error) {
	var entities []T
	err := r.engine().Table(r.table).Raw(sql, args...).Find(&entities).Error
	return entities, err
}
