# go-boot-data

数据库访问层模块，支持 GORM 和 XORM 两种 ORM 框架。

## 模块结构

```
data/
├── repository.go    # 公共接口定义
├── config.go       # 数据库配置
├── gorm/           # GORM 实现
│   ├── base_repository.go
│   └── gorm.go
└── xorm/           # XORM 实现
    ├── base_repository.go
    └── xorm.go
```

## 核心接口

### Repository[T]

通用 CRUD 接口，所有 ORM 实现需实现此接口。

```go
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
```

### Transaction

事务接口，提供事务的提交、回滚和关闭。

```go
type Transaction interface {
    Commit() error
    Rollback() error
    Close()
}
```

### Transactor

事务管理器，支持开启事务。

```go
type Transactor interface {
    Begin() (Transaction, error)
    Transaction
}
```

## 使用示例

### GORM

```go
import (
    "github.com/xudefa/go-boot/data"
    "github.com/xudefa/go-boot/data/gorm"
)

type User struct {
    data.BaseModel
    Name  string `gorm:"size:100" json:"name"`
    Email string `gorm:"size:100" json:"email"`
}

func (User) TableName() string {
    return "users"
}

client, _ := gorm.NewDefaultGormClient("user", "pass", "database")
repo := gorm.NewBaseRepository[User](client.DB)

user := &User{Name: "John", Email: "john@example.com"}
repo.Create(user)

found, _ := repo.FindByID(user.ID)
_ = found

count, _ := repo.Count("name = ?", "John")
_ = count

tx, _ := client.Begin()
txRepo := gorm.NewBaseRepositoryWithTransaction[User](tx.(*gorm.GormTransaction).Tx())
txRepo.Create(&User{Name: "Tx1", Email: "tx1@example.com"})
tx.Commit()
```

### XORM

```go
import (
    "github.com/xudefa/go-boot/data"
    "github.com/xudefa/go-boot/data/xorm"
)

client, _ := xorm.NewDefaultXormClient("user", "pass", "database")
repo := xorm.NewBaseRepository[User](client.Engine)

user := &User{Name: "Alice", Email: "alice@example.com"}
repo.Create(user)

found, _ := repo.FindByID(user.ID)
_ = found

tx, _ := client.Begin()
txRepo := xorm.NewBaseRepositoryWithTransaction[User](tx.(*xorm.XormTransaction).Session())
txRepo.Create(&User{Name: "TxA", Email: "txa@example.com"})
tx.Commit()
```

## BaseModel

提供基础模型，包含通用字段：

```go
type BaseModel struct {
    ID        uint   `gorm:"primaryKey" json:"id"`
    CreatedAt int64  `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt int64  `gorm:"autoUpdateTime" json:"updated_at"`
    DeletedAt *int64 `gorm:"index" json:"deleted_at"`
}
```

## 配置

```go
cfg := &data.DatabaseConfig{
    Driver:   "mysql",
    Host:     "localhost",
    Port:     3306,
    Username: "user",
    Password: "pass",
    Name:     "database",
    Charset:  "utf8mb4",
    MaxOpen:  1000,
    MaxIdle:  10,
    Debug:    false,
}

client, _ := gorm.NewGormClient(cfg)
```

快捷配置：

```go
cfg := data.NewDefaultDatabaseConfig("user", "pass", "database")
```

## 依赖

```go
require (
    github.com/xudefa/go-boot/data
    gorm.io/gorm
    xorm.io/xorm
)
```