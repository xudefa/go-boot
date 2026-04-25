// Package xorm 提供基于 XORM 的数据库访问层实现
package xorm

import (
	"reflect"

	"github.com/xudefa/go-boot/data"
	"xorm.io/xorm"
)

var _ data.Repository[any] = (*BaseRepository[any])(nil)

// BaseRepository 泛型 Repository 实现
//
// 支持事务模式，通过 NewBaseRepositoryWithTransaction 创建
type BaseRepository[T any] struct {
	engine  *xorm.Engine
	session *xorm.Session
	table   string
}

// NewBaseRepository 创建 Repository 实例
//
//	repo := xorm.NewBaseRepository[User](client.Engine)
func NewBaseRepository[T any](engine *xorm.Engine) *BaseRepository[T] {
	return &BaseRepository[T]{
		engine: engine,
		table:  data.GetTableName[T](),
	}
}

// NewBaseRepositoryWithTransaction 创建基于事务的 Repository 实例
//
//	txRepo := xorm.NewBaseRepositoryWithTransaction[User](tx.(*xorm.XormTransaction).Session())
func NewBaseRepositoryWithTransaction[T any](session *xorm.Session) *BaseRepository[T] {
	return &BaseRepository[T]{
		session: session,
		table:   data.GetTableName[T](),
	}
}

// Create 创建单条记录
func (r *BaseRepository[T]) Create(entity *T) error {
	if r.session != nil {
		_, err := r.session.Table(r.table).Insert(entity)
		return err
	}
	_, err := r.engine.Table(r.table).Insert(entity)
	return err
}

// CreateBatch 批量创建记录
func (r *BaseRepository[T]) CreateBatch(entities []T) error {
	if len(entities) == 0 {
		return nil
	}
	if r.session != nil {
		_, err := r.session.Insert(&entities)
		return err
	}
	_, err := r.engine.Insert(&entities)
	return err
}

// Delete 根据 ID 删除记录
func (r *BaseRepository[T]) Delete(id any) error {
	if r.session != nil {
		_, err := r.session.Table(r.table).ID(id).Delete(r.newEntity())
		return err
	}
	_, err := r.engine.Table(r.table).ID(id).Delete(r.newEntity())
	return err
}

// DeleteByCondition 根据条件删除记录
func (r *BaseRepository[T]) DeleteByCondition(where any, args ...any) error {
	if r.session != nil {
		_, err := r.session.Where(where, args...).Delete(r.newEntity())
		return err
	}
	_, err := r.engine.Where(where, args...).Delete(r.newEntity())
	return err
}

// Update 更新记录（需包含 ID）
func (r *BaseRepository[T]) Update(entity *T) error {
	if r.session != nil {
		_, err := r.session.ID(getID(entity)).Update(entity)
		return err
	}
	_, err := r.engine.ID(getID(entity)).Update(entity)
	return err
}

// UpdateByCondition 根据条件更新记录
func (r *BaseRepository[T]) UpdateByCondition(where any, args ...any) (int64, error) {
	if len(args) == 0 {
		return 0, nil
	}
	if r.session != nil {
		return r.session.Where(where, args[:len(args)-1]...).Update(r.newEntity(), args[len(args)-1].(map[string]interface{}))
	}
	return r.engine.Where(where, args[:len(args)-1]...).Update(r.newEntity(), args[len(args)-1].(map[string]interface{}))
}

// FindByID 根据 ID 查询单条记录
func (r *BaseRepository[T]) FindByID(id any) (*T, error) {
	var entity T
	var has bool
	var err error
	if r.session != nil {
		has, err = r.session.ID(id).Get(&entity)
	} else {
		has, err = r.engine.ID(id).Get(&entity)
	}
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return &entity, nil
}

// FindOne 根据条件查询单条记录
func (r *BaseRepository[T]) FindOne(where any, args ...any) (*T, error) {
	var entity T
	var has bool
	var err error
	if r.session != nil {
		has, err = r.session.Where(where, args...).Get(&entity)
	} else {
		has, err = r.engine.Where(where, args...).Get(&entity)
	}
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return &entity, nil
}

// FindAll 根据条件查询多条记录
func (r *BaseRepository[T]) FindAll(where any, args ...any) ([]T, error) {
	var entities []T
	var err error
	if r.session != nil {
		err = r.session.Where(where, args...).Find(&entities)
	} else {
		err = r.engine.Where(where, args...).Find(&entities)
	}
	return entities, err
}

// Count 统计记录数量
func (r *BaseRepository[T]) Count(where any, args ...any) (int64, error) {
	if r.session != nil {
		return r.session.Where(where, args...).Count(r.newEntity())
	}
	return r.engine.Where(where, args...).Count(r.newEntity())
}

// Raw 执行原生 SQL 查询
func (r *BaseRepository[T]) Raw(sql string, args ...any) ([]T, error) {
	var entities []T
	var err error
	if r.session != nil {
		err = r.session.SQL(sql, args...).Find(&entities)
	} else {
		err = r.engine.SQL(sql, args...).Find(&entities)
	}
	return entities, err
}

func (r *BaseRepository[T]) newEntity() *T {
	var entity T
	return &entity
}

func getID(entity any) any {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	f := v.FieldByName("ID")
	if f.IsValid() && f.CanInt() {
		return f.Interface()
	}
	return nil
}
