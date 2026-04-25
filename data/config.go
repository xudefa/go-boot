// Package data 提供数据库访问层实现
package data

import (
	"fmt"
	"time"
)

// DatabaseConfig 数据库配置结构
type DatabaseConfig struct {
	Dsn         string        // dsn 数据连接的dsn地址
	Driver      string        // 数据库驱动 (mysql/postgres等，默认mysql)
	Host        string        // 数据库地址 (默认localhost)
	Port        int           // 数据库端口 （默认3306）
	Username    string        // 数据库用户名
	Password    string        // 数据库密码
	Name        string        // 数据库名称 （默认data）
	Charset     string        // 字符编码 (默认utf-8)
	MaxOpen     int           // 最大打开连接数 (go默认是0，不限制，这里默认 1000)
	MaxIdle     int           // 最大空闲连接数（go默认 2)
	MaxLife     time.Duration // 连接最大生命时间 （go默认 0，不关闭）
	MaxIdleTime time.Duration // 最大空闲连接时间 （go默认 0，不关闭）
	Debug       bool          // 是否开启调试模式 (默认false)
}

// NewDefaultDatabaseConfig 创建默认数据库配置
// username: 数据库用户名
// password: 数据库密码
// database: 数据库名称
func NewDefaultDatabaseConfig(username, password, database string) *DatabaseConfig {
	return &DatabaseConfig{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Username: username,
		Password: password,
		Name:     database,
		Charset:  "utf8mb4",
		MaxOpen:  1000,
	}
}

// DSN 返回数据库连接字符串
// 返回: 根据Driver类型返回对应的DSN字符串
func (c *DatabaseConfig) DSN() string {
	if c.Dsn != "" {
		return c.Dsn
	}
	switch c.Driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			c.Username, c.Password, c.Host, c.Port, c.Name, c.Charset)
	case "PostgreSQL":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
			c.Host, c.Port, c.Username, c.Password, c.Name)
	case "oracle":
		return fmt.Sprintf("user=%s password=%s connectString=%s",
			c.Username, c.Password, c.Name)
	default:
		return ""
	}
}
