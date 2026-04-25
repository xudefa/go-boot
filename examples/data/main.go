// Package main 演示 data 模块的 CRUD 和事务功能
//
// 支持 Gorm 和 Xorm 两种数据库 ORM 实现:
//
//	gcd examples/data && go run .
package main

import (
	"fmt"
	"log"

	"github.com/xudefa/go-boot/data"
	"github.com/xudefa/go-boot/data/gorm"
	"github.com/xudefa/go-boot/data/xorm"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	fmt.Println("=== Data Repository Example ===")

	if err := gormExample(); err != nil {
		return fmt.Errorf("gorm example failed: %w", err)
	}

	if err := xormExample(); err != nil {
		return fmt.Errorf("xorm example failed: %w", err)
	}

	fmt.Println("=== Data Repository Example ===")
	return nil
}

type User struct {
	data.BaseModel
	Name  string `gorm:"size:100;not null" json:"name"`
	Email string `gorm:"size:100" json:"email"`
}

func (User) TableName() string {
	return "users"
}

func gormExample() error {
	fmt.Println("\n--- Gorm Example ---")

	client, err := gorm.NewDefaultGormClient("gate", "123456", "gate")
	if err != nil {
		return fmt.Errorf("connect database failed: %w", err)
	}

	repo := gorm.NewBaseRepository[User](client.DB)

	user := &User{Name: "John", Email: "john@example.com"}
	if err := repo.Create(user); err != nil {
		return fmt.Errorf("create user failed: %w", err)
	}
	fmt.Printf("Created: ID=%d, Name=%s\n", user.ID, user.Name)

	found, err := repo.FindByID(user.ID)
	if err != nil {
		return fmt.Errorf("find user failed: %w", err)
	}
	fmt.Printf("Found: ID=%d, Name=%s\n", found.ID, found.Name)

	user.Name = "Jane"
	if err := repo.Update(user); err != nil {
		return fmt.Errorf("update user failed: %w", err)
	}
	fmt.Println("Updated user")

	count, err := repo.Count("name = ?", "Jane")
	if err != nil {
		return fmt.Errorf("count failed: %w", err)
	}
	fmt.Printf("Count: %d\n", count)

	fmt.Println("\n--- Gorm Transaction Example ---")

	tx, err := client.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Close()

	txRepo := gorm.NewBaseRepositoryWithTransaction[User](tx.(*gorm.GormTransaction).Tx())
	err = txRepo.Create(&User{Name: "Tx1", Email: "tx1@example.com"})
	if err != nil {
		return fmt.Errorf("create user failed: %w", err)
	}
	err = txRepo.Create(&User{Name: "Tx2", Email: "tx2@example.com"})
	if err != nil {
		return fmt.Errorf("create user failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}
	fmt.Println("Transaction committed")

	users, _ := repo.FindAll("name IN (?)", []string{"Tx1", "Tx2"})
	fmt.Printf("Found %d users in transaction\n", len(users))

	if err := repo.DeleteByCondition("name IN (?)", []string{"Tx1", "Tx2"}); err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}

	return nil
}

func xormExample() error {
	fmt.Println("\n--- Xorm Example ---")

	client, err := xorm.NewDefaultXormClient("gate", "123456", "gate")
	if err != nil {
		return fmt.Errorf("connect database failed: %w", err)
	}

	repo := xorm.NewBaseRepository[User](client.Engine)

	user := &User{Name: "Alice", Email: "alice@example.com"}
	if err := repo.Create(user); err != nil {
		return fmt.Errorf("create user failed: %w", err)
	}
	fmt.Printf("Created: ID=%d, Name=%s\n", user.ID, user.Name)

	found, err := repo.FindByID(user.ID)
	if err != nil {
		return fmt.Errorf("find user failed: %w", err)
	}
	fmt.Printf("Found: ID=%d, Name=%s\n", found.ID, found.Name)

	user.Name = "Bob"
	if err := repo.Update(user); err != nil {
		return fmt.Errorf("update user failed: %w", err)
	}
	fmt.Println("Updated user")

	fmt.Println("\n--- Xorm Transaction Example ---")

	tx, err := client.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Close()

	txRepo := xorm.NewBaseRepositoryWithTransaction[User](tx.(*xorm.XormTransaction).Session())
	err = txRepo.Create(&User{Name: "TxA", Email: "txa@example.com"})
	if err != nil {
		return fmt.Errorf("create user failed: %w", err)
	}
	err = txRepo.Create(&User{Name: "TxB", Email: "txb@example.com"})
	if err != nil {
		return fmt.Errorf("create user failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}
	fmt.Println("Transaction committed")

	users, _ := repo.FindAll("name IN (?)", []string{"TxA", "TxB"})
	fmt.Printf("Found %d users in transaction\n", len(users))

	if err := repo.DeleteByCondition("name IN (?)", []string{"TxA", "TxB"}); err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}

	return nil
}
