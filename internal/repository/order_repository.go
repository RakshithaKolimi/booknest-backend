package repository

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"booknest/internal/domain"
)

type orderRepo struct {
	db domain.DBExecer
	sb squirrel.StatementBuilderType
}

func NewOrderRepo(db *pgxpool.Pool) domain.OrderRepository {
	return &orderRepo{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *orderRepo) CreateOrder(ctx context.Context, order *domain.Order) error {
	query := `
		INSERT INTO orders (
			id,
			order_number,
			total_price,
			user_id,
			payment_method,
			payment_status,
			status,
			cancellation_reason,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING created_at, updated_at;
	`

	row := queryRowWithTx(
		ctx,
		r.db,
		query,
		order.ID,
		order.OrderNumber,
		order.TotalPrice,
		order.UserID,
		order.PaymentMethod,
		order.PaymentStatus,
		order.Status,
		order.CancellationReason,
	)

	return row.Scan(&order.CreatedAt, &order.UpdatedAt)
}

func (r *orderRepo) CreateOrderItems(ctx context.Context, items []domain.OrderItem) error {
	query := `
		INSERT INTO order_items (
			order_id,
			book_id,
			purchase_count,
			purchase_price,
			total_price,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, NOW(), NOW());
	`

	for i := range items {
		item := items[i]
		if err := execWithTx(
			ctx,
			r.db,
			query,
			item.OrderID,
			item.BookID,
			item.PurchaseCount,
			item.PurchasePrice,
			item.TotalPrice,
		); err != nil {
			return err
		}
	}

	return nil
}

func (r *orderRepo) ListOrdersByUser(
	ctx context.Context,
	userID uuid.UUID,
	limit,
	offset int,
) ([]domain.OrderView, error) {
	query := `
		SELECT
			id,
			order_number,
			total_price,
			user_id,
			payment_method,
			payment_status,
			status,
			cancellation_reason,
			created_at,
			updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3;
	`

	rows, err := queryWithTx(ctx, r.db, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]domain.OrderView, 0)
	for rows.Next() {
		var order domain.Order
		if err := rows.Scan(
			&order.ID,
			&order.OrderNumber,
			&order.TotalPrice,
			&order.UserID,
			&order.PaymentMethod,
			&order.PaymentStatus,
			&order.Status,
			&order.CancellationReason,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, err
		}

		items, err := r.GetOrderItems(ctx, order.ID)
		if err != nil {
			return nil, err
		}

		orders = append(orders, domain.OrderView{
			Order: order,
			Items: items,
		})
	}

	return orders, rows.Err()
}

func (r *orderRepo) ListOrders(
	ctx context.Context,
	limit,
	offset int,
) ([]domain.OrderView, error) {
	query := `
		SELECT
			id,
			order_number,
			total_price,
			user_id,
			payment_method,
			payment_status,
			status,
			cancellation_reason,
			created_at,
			updated_at
		FROM orders
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2;
	`

	rows, err := queryWithTx(ctx, r.db, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]domain.OrderView, 0)
	for rows.Next() {
		var order domain.Order
		if err := rows.Scan(
			&order.ID,
			&order.OrderNumber,
			&order.TotalPrice,
			&order.UserID,
			&order.PaymentMethod,
			&order.PaymentStatus,
			&order.Status,
			&order.CancellationReason,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, err
		}

		items, err := r.GetOrderItems(ctx, order.ID)
		if err != nil {
			return nil, err
		}

		orders = append(orders, domain.OrderView{
			Order: order,
			Items: items,
		})
	}

	return orders, rows.Err()
}

func (r *orderRepo) HasUserPurchasedBook(
	ctx context.Context,
	userID, bookID uuid.UUID,
) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM orders o
			JOIN order_items oi ON oi.order_id = o.id
			WHERE o.user_id = $1
				AND oi.book_id = $2
				AND o.status = $3
				AND o.deleted_at IS NULL
				AND oi.deleted_at IS NULL
		);
	`

	var exists bool
	row := queryRowWithTx(ctx, r.db, query, userID, bookID, domain.OrderCompleted)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *orderRepo) GetOrderByID(
	ctx context.Context,
	orderID uuid.UUID,
) (domain.Order, error) {
	query := `
		SELECT
			id,
			order_number,
			total_price,
			user_id,
			payment_method,
			payment_status,
			status,
			cancellation_reason,
			created_at,
			updated_at
		FROM orders
		WHERE id = $1;
	`
	var order domain.Order
	row := queryRowWithTx(ctx, r.db, query, orderID)
	err := row.Scan(
		&order.ID,
		&order.OrderNumber,
		&order.TotalPrice,
		&order.UserID,
		&order.PaymentMethod,
		&order.PaymentStatus,
		&order.Status,
		&order.CancellationReason,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	return order, err
}

func (r *orderRepo) GetOrderItems(
	ctx context.Context,
	orderID uuid.UUID,
) ([]domain.OrderItemDetail, error) {
	query := `
		SELECT
			oi.book_id,
			b.name,
			b.image_url,
			oi.purchase_price,
			oi.purchase_count,
			oi.total_price
		FROM order_items oi
		JOIN books b ON b.id = oi.book_id
		WHERE oi.order_id = $1
		ORDER BY oi.created_at ASC;
	`

	rows, err := queryWithTx(ctx, r.db, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.OrderItemDetail, 0)
	for rows.Next() {
		var item domain.OrderItemDetail
		if err := rows.Scan(
			&item.BookID,
			&item.Name,
			&item.ImageURL,
			&item.UnitPrice,
			&item.Count,
			&item.LineTotal,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *orderRepo) UpdateOrderPayment(
	ctx context.Context,
	orderID uuid.UUID,
	status domain.PaymentStatus,
	method domain.PaymentMethod,
) error {
	query, args, err := r.sb.
		Update("orders").
		Set("payment_status", status).
		Set("payment_method", method).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": orderID}).
		ToSql()
	if err != nil {
		return err
	}
	return execWithTx(ctx, r.db, query, args...)
}

func (r *orderRepo) UpdateOrderStatus(
	ctx context.Context,
	orderID uuid.UUID,
	status domain.OrderStatus,
	cancellationReason *string,
) error {
	query, args, err := r.sb.
		Update("orders").
		Set("status", status).
		Set("cancellation_reason", cancellationReason).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": orderID}).
		ToSql()
	if err != nil {
		return err
	}
	return execWithTx(ctx, r.db, query, args...)
}

func (r *orderRepo) DecrementStock(
	ctx context.Context,
	items []domain.OrderItem,
) error {
	query := `
		UPDATE books
		SET available_stock = available_stock - $1,
		    updated_at = NOW()
		WHERE id = $2 AND available_stock >= $1;
	`

	for i := range items {
		item := items[i]
		tag, err := execWithTxTag(ctx, r.db, query, item.PurchaseCount, item.BookID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("insufficient stock for book %s", item.BookID)
		}
	}
	return nil
}
