// Package data 提供数据库访问层实现
//
// 提供通用的 Repository 接口和事务支持，可对接多种 ORM 框架（如 GORM、XORM）。
//
// 核心接口：
//
//   - Repository[T] - 通用 CRUD 接口
//   - Transactor   - 数据库事务操作接口（定义在 transactor.go）
//   - Transaction  - 事务接口（定义在 transactor.go）
//   - RepositoryFactory - 仓库工厂
package data

import "reflect"

// Repository 通用 CRUD 接口
//
// T 为实体类型，实现 Model 接口或自动推导表名。
// 各 ORM 实现需提供此接口的具体实现。
type Repository[T any] interface {
	// Create 创建单个实体
	Create(entity *T) error
	// CreateBatch 批量创建实体
	CreateBatch(entities []T) error
	// Delete 根据主键删除实体
	Delete(id any) error
	// DeleteByCondition 根据条件删除实体
	DeleteByCondition(where any, args ...any) error
	// Update 更新实体
	Update(entity *T) error
	// UpdateByCondition 根据条件更新，返回受影响行数
	UpdateByCondition(where any, args ...any) (int64, error)
	// FindByID 根据主键查询单个实体
	FindByID(id any) (*T, error)
	// FindOne 根据条件查询单个实体
	FindOne(where any, args ...any) (*T, error)
	// FindAll 根据条件查询多个实体
	FindAll(where any, args ...any) ([]T, error)
	// Count 根据条件统计实体数量
	Count(where any, args ...any) (int64, error)
	// Raw 执行原生 SQL 查询
	Raw(sql string, args ...any) ([]T, error)
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
	ID        uint   `json:"id"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
	DeletedAt *int64 `json:"deleted_at"`
}

// TableName 返回表名.
//
// 返回空字符串,表示需要由具体实体类型自行覆盖或自动推导.
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
	typ := reflect.TypeOf(zero)
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	return ToSnakeCase(typ.Name())
}

// ToSnakeCase 将驼峰命名转换为蛇形命名
//
//	UserName     -> user_name
//	UserID       -> user_id
//	HTMLParser   -> html_parser
//	ID           -> id
//	UserIDString -> user_id_string
func ToSnakeCase(s string) string {
	if len(s) == 0 {
		return s
	}

	var result []rune
	runes := []rune(s)
	for i, r := range runes {
		if r >= 'A' && r <= 'Z' {
			lower := r + 'a' - 'A'
			// 在以下情况插入下划线：
			// 1. 非首位且前一个字符不是下划线
			// 2. 当前大写字母后面跟着小写字母（表示单词边界）
			needUnderscore := i > 0 && runes[i-1] != '_'
			isWordBoundary := i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z'

			if needUnderscore && (isWordBoundary || (runes[i-1] >= 'a' && runes[i-1] <= 'z')) {
				result = append(result, '_')
			}
			result = append(result, lower)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}
