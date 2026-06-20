package data

import (
	"context"
	"testing"
)

func TestGetTableName_ModelInterface(t *testing.T) {

	t.Run("with TableName method", func(t *testing.T) {
		type UserModel struct {
			BaseModel
			Name string
		}
		// BaseModel 返回空字符串，所以会使用类型名转换
		name := GetTableName[UserModel]()
		if name != "user_model" {
			t.Errorf("expected user_model, got %s", name)
		}
	})

	t.Run("with custom TableName", func(t *testing.T) {
		type CustomTable struct{}
		// CustomTable 没有实现 TableName，所以会使用类型名转换
		name := GetTableName[CustomTable]()
		if name != "custom_table" {
			t.Errorf("expected custom_table, got %s", name)
		}
	})
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"single word", "User", "user"},
		{"camelCase", "UserName", "user_name"},
		{"PascalCase", "UserName", "user_name"},
		{"with acronym", "UserID", "user_id"},
		{"multiple acronyms", "HTMLParser", "html_parser"},
		{"single acronym", "ID", "id"},
		{"complex", "UserIDString", "user_id_string"},
		{"already snake", "user_name", "user_name"},
		{"mixed case", "HTTPServer", "http_server"},
		{"all caps", "HTTP", "http"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToSnakeCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDBFactory_Create_NilConfig(t *testing.T) {
	factory := NewDBFactory()
	_, err := factory.Create(nil)
	if err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestDBFactory_Create_Unregistered(t *testing.T) {
	factory := NewDBFactory()
	cfg := &DBConfig{
		ORM:    "unknown",
		Driver: "mysql",
		DSN:    "test-dsn",
	}
	_, err := factory.Create(cfg)
	if err == nil {
		t.Fatal("expected error for unregistered ORM")
	}
}

func TestDBFactory_RegisterAndCreate(t *testing.T) {
	factory := NewDBFactory()
	factory.Register("gorm", func(cfg *DBConfig) (Transactor, error) {
		if cfg.DSN != "test-dsn" {
			t.Errorf("expected DSN test-dsn, got %s", cfg.DSN)
		}
		return &stubTransactor{}, nil
	})

	cfg := &DBConfig{
		ORM:    "gorm",
		Driver: "mysql",
		DSN:    "test-dsn",
	}
	transactor, err := factory.Create(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if transactor == nil {
		t.Fatal("expected transactor, got nil")
	}
}

func TestDBFactory_ConcurrentRegisterAndCreate(t *testing.T) {
	factory := NewDBFactory()

	t.Run("concurrent register", func(t *testing.T) {
		t.Parallel()
		factory.Register("test1", func(cfg *DBConfig) (Transactor, error) {
			return &stubTransactor{}, nil
		})
	})

	t.Run("concurrent create", func(t *testing.T) {
		t.Parallel()
		factory.Register("test2", func(cfg *DBConfig) (Transactor, error) {
			return &stubTransactor{}, nil
		})

		_, err := factory.Create(&DBConfig{ORM: "test2", Driver: "mysql", DSN: "dsn"})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	})
}

func TestDBFactory_DefaultToGorm(t *testing.T) {
	factory := NewDBFactory()
	var called bool
	factory.Register("gorm", func(cfg *DBConfig) (Transactor, error) {
		called = true
		return &stubTransactor{}, nil
	})

	cfg := &DBConfig{
		Driver: "mysql",
		DSN:    "test-dsn",
	}
	_, err := factory.Create(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatal("expected gorm to be called as default")
	}
}

type stubTransactor struct{}

func (s *stubTransactor) Query(ctx context.Context, query string, args ...any) (Rows, error) {
	return nil, nil
}

func (s *stubTransactor) QueryRow(ctx context.Context, query string, args ...any) Row {
	return nil
}

func (s *stubTransactor) Exec(ctx context.Context, query string, args ...any) (Result, error) {
	return nil, nil
}

func (s *stubTransactor) Begin(ctx context.Context) (Transaction, error) {
	return nil, nil
}

func (s *stubTransactor) Stats() DBStats {
	return DBStats{}
}

func (s *stubTransactor) Close() error {
	return nil
}
