// 示例: go-boot/core 依赖注入容器使用指南
//
// 本示例演示 go-boot/core 包的核心功能:
//
// 1. 创建容器并注册bean
// 2. 使用 Bean() 注册bean实例
// 3. 使用 inject 标签进行字段注入
// 4. 单例作用域的缓存行为
//
// 运行方式:
//
//	go run .
package main

import (
	"fmt"
	"github.com/xudefa/go-boot/core"
	"log"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	fmt.Println("=== Core Container Example ===")

	container := core.New()

	if err := container.Register("message", core.Bean("Hello World")); err != nil {
		return fmt.Errorf("register message failed: %w", err)
	}

	if err := container.Register("hello", core.Bean(&HelloService{})); err != nil {
		return fmt.Errorf("register hello failed: %w", err)
	}

	svc, err := container.Get("hello")
	if err != nil {
		return fmt.Errorf("get hello failed: %w", err)
	}
	fmt.Println("First get:", svc.(*HelloService).Say())

	svc2, err := container.Get("hello")
	if err != nil {
		return fmt.Errorf("get hello again failed: %w", err)
	}
	fmt.Println("Second get:", svc2.(*HelloService).Say())

	fmt.Println("=== Core Container Example ===")
	return nil
}

type HelloService struct {
	Message string `inject:"message"`
}

func (s *HelloService) Say() string {
	return s.Message
}
