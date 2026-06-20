package data

import (
	"context"
	"fmt"
)

// TransactorBuilder 事务器构建器，支持链式配置
type TransactorBuilder struct {
	config   *DBConfig
	factory  *DBFactory
	beforeTx []func(ctx context.Context) error
	afterTx  []func(ctx context.Context) error
	onError  []func(ctx context.Context, err error) error
}

// NewTransactorBuilder 创建事务器构建器
func NewTransactorBuilder() *TransactorBuilder {
	return &TransactorBuilder{
		config:  &DBConfig{},
		factory: NewDBFactory(),
	}
}

// ORM 设置ORM类型
func (b *TransactorBuilder) ORM(name string) *TransactorBuilder {
	b.config.ORM = name
	return b
}

// Driver 设置数据库驱动
func (b *TransactorBuilder) Driver(name string) *TransactorBuilder {
	b.config.Driver = name
	return b
}

// DSN 设置数据库连接字符串
func (b *TransactorBuilder) DSN(dsn string) *TransactorBuilder {
	b.config.DSN = dsn
	return b
}

// Factory 使用自定义工厂
func (b *TransactorBuilder) Factory(f *DBFactory) *TransactorBuilder {
	b.factory = f
	return b
}

// BeforeTx 添加事务前回调
func (b *TransactorBuilder) BeforeTx(fn func(ctx context.Context) error) *TransactorBuilder {
	b.beforeTx = append(b.beforeTx, fn)
	return b
}

// AfterTx 添加事务后回调
func (b *TransactorBuilder) AfterTx(fn func(ctx context.Context) error) *TransactorBuilder {
	b.afterTx = append(b.afterTx, fn)
	return b
}

// OnError 添加错误回调
func (b *TransactorBuilder) OnError(fn func(ctx context.Context, err error) error) *TransactorBuilder {
	b.onError = append(b.onError, fn)
	return b
}

// Build 构建事务器
func (b *TransactorBuilder) Build() (Transactor, error) {
	if b.config.DSN == "" {
		return nil, fmt.Errorf("DSN is required")
	}

	tx, err := b.factory.Create(b.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	if len(b.beforeTx) > 0 || len(b.afterTx) > 0 || len(b.onError) > 0 {
		tx = &wrappedTransactor{
			Transactor: tx,
			beforeTx:   b.beforeTx,
			afterTx:    b.afterTx,
			onError:    b.onError,
		}
	}

	return tx, nil
}

// MustBuild 构建事务器，失败则panic
func (b *TransactorBuilder) MustBuild() Transactor {
	tx, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build transactor: %v", err))
	}
	return tx
}

// wrappedTransactor 包装的事务器，支持回调
type wrappedTransactor struct {
	Transactor
	beforeTx []func(ctx context.Context) error
	afterTx  []func(ctx context.Context) error
	onError  []func(ctx context.Context, err error) error
}

// Begin 开启事务，添加回调支持
func (w *wrappedTransactor) Begin(ctx context.Context) (Transaction, error) {
	for _, before := range w.beforeTx {
		if err := before(ctx); err != nil {
			return nil, fmt.Errorf("before transaction callback failed: %w", err)
		}
	}

	tx, err := w.Transactor.Begin(ctx)
	if err != nil {
		for _, onError := range w.onError {
			if cbErr := onError(ctx, err); cbErr != nil {
				return nil, fmt.Errorf("transaction begin failed: %w, error callback error: %v", err, cbErr)
			}
		}
		return nil, err
	}

	return &wrappedTransaction{
		Transaction: tx,
		afterTx:     w.afterTx,
		onError:     w.onError,
	}, nil
}

// wrappedTransaction 包装的事务
type wrappedTransaction struct {
	Transaction
	afterTx []func(ctx context.Context) error
	onError []func(ctx context.Context, err error) error
}

// Commit 提交事务，添加回调支持
func (w *wrappedTransaction) Commit() error {
	err := w.Transaction.Commit()
	if err != nil {
		for _, onError := range w.onError {
			if cbErr := onError(context.Background(), err); cbErr != nil {
				return fmt.Errorf("transaction commit failed: %w, error callback error: %v", err, cbErr)
			}
		}
		return err
	}

	for _, after := range w.afterTx {
		if err := after(context.Background()); err != nil {
			return fmt.Errorf("after transaction callback failed: %w", err)
		}
	}

	return nil
}

// RepositoryBuilder 仓储构建器
type RepositoryBuilder[T any] struct {
	transactor Transactor
	factory    RepositoryFactory[T]
}

// NewRepositoryBuilder 创建仓储构建器
func NewRepositoryBuilder[T any]() *RepositoryBuilder[T] {
	return &RepositoryBuilder[T]{}
}

// Transactor 设置事务器
func (b *RepositoryBuilder[T]) Transactor(tx Transactor) *RepositoryBuilder[T] {
	b.transactor = tx
	return b
}

// Factory 设置仓储工厂
func (b *RepositoryBuilder[T]) Factory(f RepositoryFactory[T]) *RepositoryBuilder[T] {
	b.factory = f
	return b
}

// Build 构建仓储
func (b *RepositoryBuilder[T]) Build() (Repository[T], error) {
	if b.factory != nil {
		return b.factory.NewRepository(), nil
	}
	return nil, fmt.Errorf("factory is required, use Transactor() or Factory() to configure")
}

// MustBuild 构建仓储，失败则panic
func (b *RepositoryBuilder[T]) MustBuild() Repository[T] {
	repo, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build repository: %v", err))
	}
	return repo
}
