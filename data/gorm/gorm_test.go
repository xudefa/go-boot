package gorm

import (
	"github.com/xudefa/go-boot/data"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewGormClient_UnsupportedDriver(t *testing.T) {
	is := assert.New(t)

	cfg := &data.DatabaseConfig{
		Driver:   "sqlite",
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "password",
		Name:     "testdb",
	}

	client, err := NewGormClient(cfg)

	is.Nil(client)
	is.Error(err)
	is.Contains(err.Error(), "unsupported driver")
}

func TestNewDefaultGormClient(t *testing.T) {
	cfg := &data.DatabaseConfig{Driver: "unsupported"}
	_, err := NewGormClient(cfg)
	if err == nil {
		t.Error("Expected error for unsupported driver")
	}
}

func TestDatabaseConfig_AllFields(t *testing.T) {
	is := assert.New(t)

	cfg := &data.DatabaseConfig{
		Driver:      "mysql",
		Host:        "127.0.0.1",
		Port:        3306,
		Username:    "root",
		Password:    "pass",
		Name:        "test",
		Charset:     "utf8mb4",
		MaxOpen:     100,
		MaxIdle:     10,
		MaxLife:     3600,
		MaxIdleTime: 600,
		Debug:       true,
	}

	is.Equal("mysql", cfg.Driver)
	is.Equal("127.0.0.1", cfg.Host)
	is.Equal(3306, cfg.Port)
	is.Equal("root", cfg.Username)
	is.Equal("pass", cfg.Password)
	is.Equal("test", cfg.Name)
	is.Equal("utf8mb4", cfg.Charset)
	is.Equal(100, cfg.MaxOpen)
	is.Equal(10, cfg.MaxIdle)
	is.Equal(time.Duration(3600), cfg.MaxLife)
	is.Equal(time.Duration(600), cfg.MaxIdleTime)
	is.True(cfg.Debug)
}
