package data

import (
	"context"
	"errors"
	"testing"
)

// mockTransactor 模拟事务器
type mockTransactor struct {
	beginErr error
}

func (m *mockTransactor) Query(ctx context.Context, query string, args ...any) (Rows, error) {
	return nil, nil
}

func (m *mockTransactor) QueryRow(ctx context.Context, query string, args ...any) Row {
	return nil
}

func (m *mockTransactor) Exec(ctx context.Context, query string, args ...any) (Result, error) {
	return nil, nil
}

func (m *mockTransactor) Begin(ctx context.Context) (Transaction, error) {
	if m.beginErr != nil {
		return nil, m.beginErr
	}
	return &mockTransaction{}, nil
}

func (m *mockTransactor) Stats() DBStats {
	return DBStats{}
}

func (m *mockTransactor) Close() error {
	return nil
}

// mockTransaction 模拟事务
type mockTransaction struct {
	commitErr   error
	rollbackErr error
}

func (t *mockTransaction) Query(ctx context.Context, query string, args ...any) (Rows, error) {
	return nil, nil
}

func (t *mockTransaction) QueryRow(ctx context.Context, query string, args ...any) Row {
	return nil
}

func (t *mockTransaction) Exec(ctx context.Context, query string, args ...any) (Result, error) {
	return nil, nil
}

func (t *mockTransaction) Begin(ctx context.Context) (Transaction, error) {
	return nil, errors.New("nested transactions not supported")
}

func (t *mockTransaction) Stats() DBStats {
	return DBStats{}
}

func (t *mockTransaction) Close() error {
	return nil
}

func (t *mockTransaction) Commit() error {
	return t.commitErr
}

func (t *mockTransaction) Rollback() error {
	return t.rollbackErr
}

// mockRepository 模拟仓储
type mockRepository struct{}

func (m *mockRepository) Create(entity *User) error                               { return nil }
func (m *mockRepository) CreateBatch(entities []User) error                       { return nil }
func (m *mockRepository) Delete(id any) error                                     { return nil }
func (m *mockRepository) DeleteByCondition(where any, args ...any) error          { return nil }
func (m *mockRepository) Update(entity *User) error                               { return nil }
func (m *mockRepository) UpdateByCondition(where any, args ...any) (int64, error) { return 0, nil }
func (m *mockRepository) FindByID(id any) (*User, error)                          { return &User{}, nil }
func (m *mockRepository) FindOne(where any, args ...any) (*User, error)           { return &User{}, nil }
func (m *mockRepository) FindAll(where any, args ...any) ([]User, error)          { return []User{}, nil }
func (m *mockRepository) Count(where any, args ...any) (int64, error)             { return 0, nil }
func (m *mockRepository) Raw(sql string, args ...any) ([]User, error)             { return []User{}, nil }

// User 测试实体
type User struct {
	BaseModel
	Name string
}

// mockRepoFactory 模拟仓储工厂
type mockRepoFactory struct{}

func (f *mockRepoFactory) NewRepository() Repository[User] {
	return &mockRepository{}
}

func TestTransactorBuilder_MissingDSN(t *testing.T) {
	_, err := NewTransactorBuilder().Build()
	if err == nil {
		t.Error("expected error for missing DSN")
	}
}

func TestTransactorBuilder_MustBuild_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for missing DSN")
		}
	}()

	NewTransactorBuilder().MustBuild()
}

func TestWrappedTransactor_BeginSuccess(t *testing.T) {
	beforeCalled := false

	wrapped := &wrappedTransactor{
		Transactor: &mockTransactor{},
		beforeTx: []func(ctx context.Context) error{
			func(ctx context.Context) error {
				beforeCalled = true
				return nil
			},
		},
	}

	tx, err := wrapped.Begin(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}

	if !beforeCalled {
		t.Error("expected beforeTx callback to be called")
	}
}

func TestWrappedTransactor_BeginError(t *testing.T) {
	expectedErr := errors.New("begin error")
	errorCalled := false

	wrapped := &wrappedTransactor{
		Transactor: &mockTransactor{beginErr: expectedErr},
		onError: []func(ctx context.Context, err error) error{
			func(ctx context.Context, err error) error {
				errorCalled = true
				return nil
			},
		},
	}

	_, err := wrapped.Begin(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	if !errorCalled {
		t.Error("expected onError callback to be called")
	}
}

func TestWrappedTransaction_CommitSuccess(t *testing.T) {
	afterCalled := false

	wrapped := &wrappedTransaction{
		Transaction: &mockTransaction{},
		afterTx: []func(ctx context.Context) error{
			func(ctx context.Context) error {
				afterCalled = true
				return nil
			},
		},
	}

	err := wrapped.Commit()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !afterCalled {
		t.Error("expected afterTx callback to be called")
	}
}

func TestWrappedTransaction_CommitError(t *testing.T) {
	expectedErr := errors.New("commit error")
	errorCalled := false

	wrapped := &wrappedTransaction{
		Transaction: &mockTransaction{commitErr: expectedErr},
		onError: []func(ctx context.Context, err error) error{
			func(ctx context.Context, err error) error {
				errorCalled = true
				return nil
			},
		},
	}

	err := wrapped.Commit()
	if err == nil {
		t.Fatal("expected error")
	}

	if !errorCalled {
		t.Error("expected onError callback to be called")
	}
}

func TestWrappedTransaction_AfterTxError(t *testing.T) {
	expectedErr := errors.New("after error")

	wrapped := &wrappedTransaction{
		Transaction: &mockTransaction{},
		afterTx: []func(ctx context.Context) error{
			func(ctx context.Context) error {
				return expectedErr
			},
		},
	}

	err := wrapped.Commit()
	if err == nil {
		t.Fatal("expected error")
	}

	if err.Error() != "after transaction callback failed: after error" {
		t.Errorf("expected after error, got: %v", err)
	}
}

func TestRepositoryBuilder_WithFactory(t *testing.T) {
	repo, err := NewRepositoryBuilder[User]().
		Factory(&mockRepoFactory{}).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo == nil {
		t.Fatal("expected non-nil repository")
	}
}

func TestRepositoryBuilder_MissingFactory(t *testing.T) {
	_, err := NewRepositoryBuilder[User]().Build()
	if err == nil {
		t.Error("expected error for missing factory")
	}
}

func TestRepositoryBuilder_MustBuild_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for missing factory")
		}
	}()

	NewRepositoryBuilder[User]().MustBuild()
}

func TestTransactorBuilder_ChainConfig(t *testing.T) {
	builder := NewTransactorBuilder().
		ORM("gorm").
		Driver("mysql").
		DSN("root:password@tcp(localhost:3306)/test")

	if builder.config.ORM != "gorm" {
		t.Errorf("expected ORM 'gorm', got %s", builder.config.ORM)
	}

	if builder.config.Driver != "mysql" {
		t.Errorf("expected Driver 'mysql', got %s", builder.config.Driver)
	}

	if builder.config.DSN != "root:password@tcp(localhost:3306)/test" {
		t.Errorf("expected DSN 'root:password@tcp(localhost:3306)/test', got %s", builder.config.DSN)
	}
}

func TestTransactorBuilder_WithCustomFactory(t *testing.T) {
	customFactory := NewDBFactory()
	customFactory.Register("gorm", func(cfg *DBConfig) (Transactor, error) {
		return &mockTransactor{}, nil
	})

	tx, err := NewTransactorBuilder().
		ORM("gorm").
		DSN("test").
		Factory(customFactory).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tx == nil {
		t.Fatal("expected non-nil transactor")
	}
}
