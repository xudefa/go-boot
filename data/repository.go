// Package data 提供数据库访问层实现
//
// 提供通用的 Repository 接口和事务支持，可对接多种 ORM 框架（如 GORM、XORM）。
//
// 核心接口：
//
//   - Repository[T] - 通用 CRUD 接口
//   - Transaction  - 事务接口（提交/回滚/关闭）
//   - Transactor  - 事务管理器
//
// 使用示例：
//
//	client, _ := gorm.NewDefaultGormClient("user", "pass", "db")
//	repo := gorm.NewBaseRepository[User](client.DB)
//
//	user := &User{Name: "John"}
//	repo.Create(user)
//
//	tx, _ := client.Begin()
//	txRepo := gorm.NewBaseRepositoryWithTransaction[User](tx.(*gorm.GormTransaction).Tx())
//	txRepo.Create(&User{Name: "Tx"})
//	tx.Commit()
package data

import "reflect"

// Repository 定义通用 CRUD 接口
//
// T 为实体类型，实现 Model 接口或自动推导表名
type Repository[T any] interface {
	Create(entity *T) error
	CreateBatch(entities []T) error
	Delete(id any) error
	DeleteByCondition(where any, args ...any) error
	Update(entity *T) error
	UpdateByCondition(where any, args ...any) (int64, error)
	FindByID(id any) (*T, error)
	FindOne(where any, args ...any) (*T, error)
	FindAll(where any, args ...any) ([]T, error)
	Count(where any, args ...any) (int64, error)
	Raw(sql string, args ...any) ([]T, error)
}

// Transactor 事务管理器，支持开启事务
type Transactor interface {
	Begin() (Transaction, error)
	Transaction
}

// Transaction 事务接口
//
// 提供事务控制方法，事务开启后需在 defer 中调用 Close
type Transaction interface {
	// Commit 提交事务
	Commit() error
	// Rollback 回滚事务
	Rollback() error
	// Close 关闭事务
	//
	// 通常在 defer 中调用：
	//
	//	tx, _ := client.Begin()
	//	defer tx.Close()
	//	// ...
	//	tx.Commit()
	Close()
}

// RepositoryFactory 仓库工厂
//
// 客户端应实现此接口以创建 Repository 实例
type RepositoryFactory[T any] interface {
	NewRepository() Repository[T]
}

// Model 定义表名接口
//
// 实体实现此接口可自定义表名，否则自动推导
type Model interface {
	TableName() string
}

// BaseModel 基础模型
//
// 提供通用字段：ID、CreatedAt、UpdatedAt、DeletedAt
//
// 实体可嵌入此结构体：
//
//	type User struct {
//	    data.BaseModel
//	    Name string
//	}
type BaseModel struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	CreatedAt int64  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt *int64 `gorm:"index" json:"deleted_at"`
}

func (BaseModel) TableName() string {
	return ""
}

// GetTableName 获取实体对应的表名
//
// 优先使用 Model.TableName()，否则将类型名转换为蛇形
func GetTableName[T any]() string {
	var t T
	if m, ok := any(&t).(Model); ok {
		name := m.TableName()
		if name != "" {
			return name
		}
	}
	var zero T
	typeName := reflect.TypeOf(zero).Name()
	return ToSnakeCase(typeName)
}

// ToSnakeCase 将驼峰命名转换为蛇形命名
//
//	UserName -> user_name
//	ID      -> i_d
func ToSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	if len(result) == 0 {
		return ""
	}
	lower := make([]rune, len(result))
	for i, r := range result {
		if r >= 'A' && r <= 'Z' {
			lower[i] = r + 32
		} else {
			lower[i] = r
		}
	}
	return string(lower)
}
