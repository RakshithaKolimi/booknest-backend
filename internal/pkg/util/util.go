package util

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"booknest/internal/domain"
)

var beginTx = func(ctx context.Context, pool *pgxpool.Pool) (pgx.Tx, error) {
	return pool.Begin(ctx)
}

type pgxTransactionManager struct {
	pool *pgxpool.Pool
}

func NewTransactionManager(pool *pgxpool.Pool) domain.TransactionManager {
	return &pgxTransactionManager{pool: pool}
}

func (m *pgxTransactionManager) InTransaction(
	ctx context.Context,
	fn func(ctx context.Context) error,
) error {

	// Begin the transaction
	tx, err := beginTx(ctx, m.pool)
	if err != nil {
		return err
	}

	// Add value to context
	ctx = context.WithValue(ctx, domain.TxKey, tx)

	// If an error occurs, rollback
	if err := fn(ctx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	// Commit the transaction
	return tx.Commit(ctx)
}

// WithTransaction is kept for backward compatibility.
func WithTransaction(
	ctx context.Context,
	pool *pgxpool.Pool,
	fn func(ctx context.Context) error,
) error {
	return NewTransactionManager(pool).InTransaction(ctx, fn)
}
