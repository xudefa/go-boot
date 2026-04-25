package data

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultDatabaseConfig(t *testing.T) {
	is := assert.New(t)

	cfg := NewDefaultDatabaseConfig("root", "password", "testdb")

	is.Equal("mysql", cfg.Driver)
	is.Equal("localhost", cfg.Host)
	is.Equal(3306, cfg.Port)
	is.Equal("root", cfg.Username)
	is.Equal("password", cfg.Password)
	is.Equal("testdb", cfg.Name)
	is.Equal("utf8mb4", cfg.Charset)
	is.Equal(1000, cfg.MaxOpen)
}

func TestDatabaseConfig_DSN_MySQL(t *testing.T) {
	is := assert.New(t)

	cfg := &DatabaseConfig{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "password",
		Name:     "testdb",
		Charset:  "utf8mb4",
	}

	dsn := cfg.DSN()
	is.Equal("root:password@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local", dsn)
}

func TestDatabaseConfig_DSN_PostgreSQL(t *testing.T) {
	is := assert.New(t)

	cfg := &DatabaseConfig{
		Driver:   "PostgreSQL",
		Host:     "localhost",
		Port:     5432,
		Username: "postgres",
		Password: "password",
		Name:     "testdb",
	}

	dsn := cfg.DSN()
	is.Equal("host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=disable TimeZone=Asia/Shanghai", dsn)
}

func TestDatabaseConfig_DSN_Oracle(t *testing.T) {
	is := assert.New(t)

	cfg := &DatabaseConfig{
		Driver:   "oracle",
		Username: "system",
		Password: "password",
		Name:     "localhost:1521/orcl",
	}

	dsn := cfg.DSN()
	is.Equal("user=system password=password connectString=localhost:1521/orcl", dsn)
}

func TestDatabaseConfig_DSN_Unsupported(t *testing.T) {
	is := assert.New(t)

	cfg := &DatabaseConfig{
		Driver:   "sqlite",
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "password",
		Name:     "testdb",
	}

	dsn := cfg.DSN()
	is.Equal("", dsn)
}

func TestDatabaseConfig_DefaultValues(t *testing.T) {
	is := assert.New(t)

	cfg := &DatabaseConfig{
		Username: "root",
		Password: "password",
		Name:     "testdb",
	}
	is.Equal("", cfg.Driver)
	is.Equal("", cfg.Host)
	is.Equal(0, cfg.Port)
	is.Equal("", cfg.Charset)
	is.Equal(0, cfg.MaxOpen)
	is.Equal(0, cfg.MaxIdle)
	is.Equal(time.Duration(0), cfg.MaxLife)
	is.Equal(time.Duration(0), cfg.MaxIdleTime)
	is.False(cfg.Debug)
}
