package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	pgxmock "github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"

	"booknest/internal/domain"
)

func TestNewOrderRepo(t *testing.T) {
	repo := NewOrderRepo(nil)
	require.NotNil(t, repo)
}

func TestOrderRepo_CreateOrderAndItems(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &orderRepo{db: mock, sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
	orderID := uuid.New()
	userID := uuid.New()
	method := domain.PaymentCOD
	status := domain.PaymentPending
	reason := "Customer cancelled"
	now := time.Now()

	order := &domain.Order{
		ID:                 orderID,
		OrderNumber:        "ORD-1",
		TotalPrice:         42.5,
		UserID:             userID,
		PaymentMethod:      &method,
		PaymentStatus:      &status,
		Status:             domain.OrderCancelled,
		CancellationReason: &reason,
	}

	mock.ExpectQuery("INSERT INTO orders").
		WithArgs(orderID, "ORD-1", 42.5, userID, &method, &status, domain.OrderCancelled, &reason).
		WillReturnRows(pgxmock.NewRows([]string{"created_at", "updated_at"}).AddRow(now, now))

	require.NoError(t, repo.CreateOrder(context.Background(), order))
	require.WithinDuration(t, now, order.CreatedAt, time.Second)
	require.WithinDuration(t, now, order.UpdatedAt, time.Second)

	bookID := uuid.New()
	items := []domain.OrderItem{{
		OrderID:       orderID,
		BookID:        bookID,
		PurchaseCount: 2,
		PurchasePrice: 21.25,
		TotalPrice:    42.5,
	}}
	mock.ExpectExec("INSERT INTO order_items").
		WithArgs(orderID, bookID, 2, 21.25, 42.5).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	require.NoError(t, repo.CreateOrderItems(context.Background(), items))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepo_GetOrderByIDAndItems(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &orderRepo{db: mock, sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
	orderID := uuid.New()
	userID := uuid.New()
	bookID := uuid.New()
	now := time.Now()
	method := domain.PaymentUPI
	status := domain.PaymentPaid
	imageURL := "https://example.com/book.jpg"

	mock.ExpectQuery("SELECT(.|\n)*FROM orders").
		WithArgs(orderID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"id", "order_number", "total_price", "user_id", "payment_method", "payment_status", "status", "cancellation_reason", "created_at", "updated_at",
			}).AddRow(orderID, "ORD-2", 30.0, userID, &method, &status, domain.OrderPending, nil, now, now),
		)

	order, err := repo.GetOrderByID(context.Background(), orderID)
	require.NoError(t, err)
	require.Equal(t, orderID, order.ID)
	require.NotNil(t, order.PaymentMethod)
	require.Equal(t, domain.PaymentUPI, *order.PaymentMethod)

	mock.ExpectQuery("SELECT(.|\n)*FROM order_items oi").
		WithArgs(orderID).
		WillReturnRows(
			pgxmock.NewRows([]string{"book_id", "name", "image_url", "purchase_price", "purchase_count", "total_price"}).
				AddRow(bookID, "Test Book", &imageURL, 15.0, 2, 30.0),
		)

	gotItems, err := repo.GetOrderItems(context.Background(), orderID)
	require.NoError(t, err)
	require.Len(t, gotItems, 1)
	require.Equal(t, "Test Book", gotItems[0].Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepo_ListOrdersByUserAndListOrders(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &orderRepo{db: mock, sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
	userID := uuid.New()
	orderID1 := uuid.New()
	orderID2 := uuid.New()
	bookID := uuid.New()
	now := time.Now()
	reason := "done"

	orderRows := pgxmock.NewRows([]string{
		"id", "order_number", "total_price", "user_id", "payment_method", "payment_status", "status", "cancellation_reason", "created_at", "updated_at",
	}).
		AddRow(orderID1, "ORD-1", 20.0, userID, ptrPaymentMethod(domain.PaymentCOD), ptrPaymentStatus(domain.PaymentPending), domain.OrderPending, nil, now, now)
	mock.ExpectQuery("SELECT(.|\n)*FROM orders(.|\n)*WHERE user_id =").
		WithArgs(userID, 10, 0).
		WillReturnRows(orderRows)
	mock.ExpectQuery("SELECT(.|\n)*FROM order_items oi").
		WithArgs(orderID1).
		WillReturnRows(
			pgxmock.NewRows([]string{"book_id", "name", "image_url", "purchase_price", "purchase_count", "total_price"}).
				AddRow(bookID, "User Book", nil, 20.0, 1, 20.0),
		)

	userOrders, err := repo.ListOrdersByUser(context.Background(), userID, 10, 0)
	require.NoError(t, err)
	require.Len(t, userOrders, 1)
	require.Equal(t, "User Book", userOrders[0].Items[0].Name)

	allRows := pgxmock.NewRows([]string{
		"id", "order_number", "total_price", "user_id", "payment_method", "payment_status", "status", "cancellation_reason", "created_at", "updated_at",
	}).
		AddRow(orderID2, "ORD-2", 50.0, userID, ptrPaymentMethod(domain.PaymentUPI), ptrPaymentStatus(domain.PaymentPaid), domain.OrderCompleted, &reason, now, now)
	mock.ExpectQuery("SELECT(.|\n)*FROM orders(.|\n)*ORDER BY created_at DESC").
		WithArgs(5, 1).
		WillReturnRows(allRows)
	mock.ExpectQuery("SELECT(.|\n)*FROM order_items oi").
		WithArgs(orderID2).
		WillReturnRows(
			pgxmock.NewRows([]string{"book_id", "name", "image_url", "purchase_price", "purchase_count", "total_price"}).
				AddRow(bookID, "Admin Book", nil, 25.0, 2, 50.0),
		)

	allOrders, err := repo.ListOrders(context.Background(), 5, 1)
	require.NoError(t, err)
	require.Len(t, allOrders, 1)
	require.Equal(t, "Admin Book", allOrders[0].Items[0].Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func ptrPaymentMethod(method domain.PaymentMethod) *domain.PaymentMethod {
	return &method
}

func ptrPaymentStatus(status domain.PaymentStatus) *domain.PaymentStatus {
	return &status
}

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
		WithArgs(domain.OrderCompleted, (*string)(nil), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	require.NoError(t, repo.UpdateOrderStatus(context.Background(), orderID, domain.OrderCompleted, nil))

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
