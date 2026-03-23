package repository

import (
	"context"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	pgxmock "github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"
)

func TestNewCartRepo(t *testing.T) {
	repo := NewCartRepo(nil)
	require.NotNil(t, repo)
}

func TestCartRepo_GetOrCreateCart(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &cartRepo{db: mock, sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
	userID := uuid.New()
	cartID := uuid.New()

	mock.ExpectQuery("WITH inserted AS").
		WithArgs(pgxmock.AnyArg(), userID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id"}).AddRow(cartID, userID))

	cart, err := repo.GetOrCreateCart(context.Background(), userID)
	require.NoError(t, err)
	require.Equal(t, cartID, cart.ID)
	require.Equal(t, userID, cart.UserID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCartRepo_UpsertAndRemoveAndClear(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &cartRepo{db: mock, sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
	cartID := uuid.New()
	bookID := uuid.New()

	mock.ExpectExec("INSERT INTO cart_items").
		WithArgs(cartID, bookID, 2, 99.5).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	require.NoError(t, repo.UpsertCartItem(context.Background(), cartID, bookID, 2, 99.5))

	mock.ExpectExec("UPDATE cart_items").
		WithArgs(cartID, bookID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	require.NoError(t, repo.RemoveCartItem(context.Background(), cartID, bookID))

	mock.ExpectExec("UPDATE cart_items").
		WithArgs(cartID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 2))
	require.NoError(t, repo.ClearCart(context.Background(), cartID))

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCartRepo_GetCartItemsAndRecords(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &cartRepo{db: mock, sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
	userID := uuid.New()
	bookID := uuid.New()
	imageURL := "https://example.com/book.jpg"

	mock.ExpectQuery("SELECT(.|\n)*FROM carts c(.|\n)*JOIN cart_items ci").
		WithArgs(userID).
		WillReturnRows(
			pgxmock.NewRows([]string{"book_id", "name", "author_name", "image_url", "unit_price", "count", "line_total"}).
				AddRow(bookID, "Book", "Author", &imageURL, 15.0, 2, 30.0),
		)

	items, err := repo.GetCartItems(context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "Author", items[0].AuthorName)

	mock.ExpectQuery("SELECT(.|\n)*FROM carts c(.|\n)*JOIN cart_items ci").
		WithArgs(userID).
		WillReturnRows(
			pgxmock.NewRows([]string{"book_id", "count", "unit_price", "available_stock"}).
				AddRow(bookID, 2, 15.0, 10),
		)

	records, err := repo.GetCartItemRecords(context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, records, 1)
	require.Equal(t, 10, records[0].AvailableStock)
	require.NoError(t, mock.ExpectationsWereMet())
}
