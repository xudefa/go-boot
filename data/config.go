// Package data 提供数据库配置和工厂支持。
//
// 包括 DatabaseConfig（连接配置）、DatabaseOption（函数式选项）、
// DBFactory（根据 ORM 类型动态创建数据访问层实例）。
package data

import (
	"fmt"
	"strings"
	"time"
)

// DatabaseOption 数据库配置选项函数
type DatabaseOption func(*DatabaseConfig)

// WithHost 设置数据库主机地址。
func WithHost(host string) DatabaseOption {
	return func(c *DatabaseConfig) { c.Host = host }
}

// WithPort 设置数据库端口。
func WithPort(port int) DatabaseOption {
	return func(c *DatabaseConfig) { c.Port = port }
}

// WithDriver 设置数据库驱动类型（如 mysql、postgres、oracle）。
func WithDriver(driver string) DatabaseOption {
	return func(c *DatabaseConfig) { c.Driver = driver }
}

// WithCharset 设置字符编码（默认 utf8mb4）。
func WithCharset(charset string) DatabaseOption {
	return func(c *DatabaseConfig) { c.Charset = charset }
}

// WithMaxOpenConns 设置最大打开连接数（默认 1000）。
func WithMaxOpenConns(n int) DatabaseOption {
	return func(c *DatabaseConfig) { c.MaxOpen = n }
}

// WithMaxIdleConns 设置最大空闲连接数（Go 默认 2）。
func WithMaxIdleConns(n int) DatabaseOption {
	return func(c *DatabaseConfig) { c.MaxIdle = n }
}

// WithConnMaxLifetime 设置连接最大生命时间（0 表示不限制）。
func WithConnMaxLifetime(d time.Duration) DatabaseOption {
	return func(c *DatabaseConfig) { c.MaxLife = d }
}

// WithConnMaxIdleTime 设置最大空闲连接时间（0 表示不限制）。
func WithConnMaxIdleTime(d time.Duration) DatabaseOption {
	return func(c *DatabaseConfig) { c.MaxIdleTime = d }
}

// WithDebug 设置是否开启调试模式，开启后打印 SQL 日志。
func WithDebug(debug bool) DatabaseOption {
	return func(c *DatabaseConfig) { c.Debug = debug }
}

// DatabaseConfig 数据库配置结构体
//
// 支持 MySQL、PostgreSQL、Oracle 等主流数据库的连接配置。
// 通过 DSN() 方法自动生成对应驱动的连接字符串。
type DatabaseConfig struct {
	Dsn         string        // 完整的数据连接 dsn 地址（设置后忽略其他连接参数）
	Driver      string        // 数据库驱动类型：mysql / postgres / oracle（默认 mysql）
	Host        string        // 数据库地址（默认 localhost）
	Port        int           // 数据库端口（默认 3306）
	Username    string        // 数据库用户名
	Password    string        // 数据库密码
	Name        string        // 数据库名称（默认 gate）
	Charset     string        // 字符编码（默认 utf8mb4）
	MaxOpen     int           // 最大打开连接数（默认 1000）
	MaxIdle     int           // 最大空闲连接数（Go 默认 2）
	MaxLife     time.Duration // 连接最大生命时间（0 表示不限制）
	MaxIdleTime time.Duration // 最大空闲连接时间（0 表示不限制）
	Debug       bool          // 是否开启调试模式，打印 SQL 日志（默认 false）
}

// NewDefaultDatabaseConfig 创建默认数据库配置
// username: 数据库用户名
// password: 数据库密码
// database: 数据库名称
// opts: 可选配置选项
func NewDefaultDatabaseConfig(username, password, database string, opts ...DatabaseOption) *DatabaseConfig {
	cfg := &DatabaseConfig{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Username: username,
		Password: password,
		Name:     database,
		Charset:  "utf8mb4",
		MaxOpen:  1000,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// DSN 根据数据库驱动类型生成对应的连接字符串
//
// 生成流程：
//  1. 如果已设置 Dsn 字段，直接返回
//  2. 根据 Driver 类型选择对应的连接串格式
//  3. MySQL 返回 "user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
//  4. PostgreSQL 返回 "host= port= user= password= dbname= sslmode=disable"
//  5. Oracle 返回 "user= password= connectString="
func (c *DatabaseConfig) DSN() string {
	if c.Dsn != "" {
		return c.Dsn
	}
	switch strings.ToLower(c.Driver) {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			c.Username, c.Password, c.Host, c.Port, c.Name, c.Charset)
	case "postgresql", "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
			c.Host, c.Port, c.Username, c.Password, c.Name)
	case "oracle":
		return fmt.Sprintf("user=%s password=%s connectString=%s",
			c.Username, c.Password, c.Name)
	default:
		return ""
	}
}
