package domain

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Transaction key constants
type TxKeyType string

const TxKey TxKeyType = "BookNest-Transactioner"

// DBExecer defines the interface for database operations -> to maintain abstraction over pgxpool.Pool and pgx.Tx and also to facilitate testing
type DBExecer interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// TransactionManager abstracts transaction handling so services don't depend on DB drivers.
type TransactionManager interface {
	InTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
