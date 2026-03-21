package repository

import (
	"context"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	pgxmock "github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"

	"booknest/internal/domain"
)

func TestOrderRepo_UpdatePaymentAndStatus(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &orderRepo{db: mock, sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
	orderID := uuid.New()

	mock.ExpectExec("UPDATE orders").
		WithArgs(domain.PaymentPaid, domain.PaymentCOD, pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	require.NoError(t, repo.UpdateOrderPayment(context.Background(), orderID, domain.PaymentPaid, domain.PaymentCOD))

	mock.ExpectExec("UPDATE orders").
		WithArgs(domain.OrderCompleted, pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	require.NoError(t, repo.UpdateOrderStatus(context.Background(), orderID, domain.OrderCompleted))

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepo_DecrementStock(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &orderRepo{db: mock, sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
	bookID := uuid.New()

	mock.ExpectExec("UPDATE books").
		WithArgs(2, bookID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	require.NoError(t, repo.DecrementStock(context.Background(), []domain.OrderItem{{BookID: bookID, PurchaseCount: 2}}))

	bookID2 := uuid.New()
	mock.ExpectExec("UPDATE books").
		WithArgs(3, bookID2).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))
	err = repo.DecrementStock(context.Background(), []domain.OrderItem{{BookID: bookID2, PurchaseCount: 3}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient stock")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepo_HasUserPurchasedBook(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &orderRepo{db: mock, sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
	userID := uuid.New()
	bookID := uuid.New()

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(userID, bookID, domain.OrderCompleted).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

	purchased, err := repo.HasUserPurchasedBook(context.Background(), userID, bookID)
	require.NoError(t, err)
	require.True(t, purchased)

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(userID, bookID, domain.OrderCompleted).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(false))

	purchased, err = repo.HasUserPurchasedBook(context.Background(), userID, bookID)
	require.NoError(t, err)
	require.False(t, purchased)

	require.NoError(t, mock.ExpectationsWereMet())
}
