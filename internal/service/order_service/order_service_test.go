package order_service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"booknest/internal/domain"
)

type mockOrderRepository struct {
	createOrderFunc        func(ctx context.Context, order *domain.Order) error
	createOrderItemsFunc   func(ctx context.Context, items []domain.OrderItem) error
	getOrderByIDFunc       func(ctx context.Context, orderID uuid.UUID) (domain.Order, error)
	getOrderItemsFunc      func(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItemDetail, error)
	updateOrderPaymentFunc func(ctx context.Context, orderID uuid.UUID, status domain.PaymentStatus, method domain.PaymentMethod) error
	updateOrderStatusFunc  func(ctx context.Context, orderID uuid.UUID, status domain.OrderStatus) error
	decrementStockFunc     func(ctx context.Context, items []domain.OrderItem) error
	listOrdersByUserFunc   func(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.OrderView, error)
	listOrdersFunc         func(ctx context.Context, limit, offset int) ([]domain.OrderView, error)
}

func (m *mockOrderRepository) CreateOrder(ctx context.Context, order *domain.Order) error {
	if m.createOrderFunc != nil {
		return m.createOrderFunc(ctx, order)
	}
	return nil
}
func (m *mockOrderRepository) CreateOrderItems(ctx context.Context, items []domain.OrderItem) error {
	if m.createOrderItemsFunc != nil {
		return m.createOrderItemsFunc(ctx, items)
	}
	return nil
}
func (m *mockOrderRepository) GetOrderByID(ctx context.Context, orderID uuid.UUID) (domain.Order, error) {
	if m.getOrderByIDFunc != nil {
		return m.getOrderByIDFunc(ctx, orderID)
	}
	return domain.Order{}, errors.New("not implemented")
}
func (m *mockOrderRepository) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItemDetail, error) {
	if m.getOrderItemsFunc != nil {
		return m.getOrderItemsFunc(ctx, orderID)
	}
	return nil, errors.New("not implemented")
}
func (m *mockOrderRepository) UpdateOrderPayment(ctx context.Context, orderID uuid.UUID, status domain.PaymentStatus, method domain.PaymentMethod) error {
	if m.updateOrderPaymentFunc != nil {
		return m.updateOrderPaymentFunc(ctx, orderID, status, method)
	}
	return nil
}
func (m *mockOrderRepository) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status domain.OrderStatus) error {
	if m.updateOrderStatusFunc != nil {
		return m.updateOrderStatusFunc(ctx, orderID, status)
	}
	return nil
}
func (m *mockOrderRepository) DecrementStock(ctx context.Context, items []domain.OrderItem) error {
	if m.decrementStockFunc != nil {
		return m.decrementStockFunc(ctx, items)
	}
	return nil
}
func (m *mockOrderRepository) ListOrdersByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.OrderView, error) {
	if m.listOrdersByUserFunc != nil {
		return m.listOrdersByUserFunc(ctx, userID, limit, offset)
	}
	return []domain.OrderView{}, nil
}
func (m *mockOrderRepository) ListOrders(ctx context.Context, limit, offset int) ([]domain.OrderView, error) {
	if m.listOrdersFunc != nil {
		return m.listOrdersFunc(ctx, limit, offset)
	}
	return []domain.OrderView{}, nil
}

type noopCartRepository struct{}

func (n *noopCartRepository) GetOrCreateCart(ctx context.Context, userID uuid.UUID) (domain.Cart, error) {
	return domain.Cart{}, nil
}
func (n *noopCartRepository) GetCartItems(ctx context.Context, userID uuid.UUID) ([]domain.CartItemDetail, error) {
	return nil, nil
}
func (n *noopCartRepository) GetCartItemRecords(ctx context.Context, userID uuid.UUID) ([]domain.CartItemRecord, error) {
	return nil, nil
}
func (n *noopCartRepository) UpsertCartItem(ctx context.Context, cartID uuid.UUID, bookID uuid.UUID, count int, unitPrice float64) error {
	return nil
}
func (n *noopCartRepository) RemoveCartItem(ctx context.Context, cartID uuid.UUID, bookID uuid.UUID) error {
	return nil
}
func (n *noopCartRepository) ClearCart(ctx context.Context, cartID uuid.UUID) error { return nil }

type mockCartRepository struct {
	getOrCreateCartFunc  func(ctx context.Context, userID uuid.UUID) (domain.Cart, error)
	getCartItemsFunc     func(ctx context.Context, userID uuid.UUID) ([]domain.CartItemDetail, error)
	getCartItemRecordsFn func(ctx context.Context, userID uuid.UUID) ([]domain.CartItemRecord, error)
	clearCartFunc        func(ctx context.Context, cartID uuid.UUID) error
}

func (m *mockCartRepository) GetOrCreateCart(ctx context.Context, userID uuid.UUID) (domain.Cart, error) {
	if m.getOrCreateCartFunc != nil {
		return m.getOrCreateCartFunc(ctx, userID)
	}
	return domain.Cart{}, nil
}
func (m *mockCartRepository) GetCartItems(ctx context.Context, userID uuid.UUID) ([]domain.CartItemDetail, error) {
	if m.getCartItemsFunc != nil {
		return m.getCartItemsFunc(ctx, userID)
	}
	return nil, nil
}
func (m *mockCartRepository) GetCartItemRecords(ctx context.Context, userID uuid.UUID) ([]domain.CartItemRecord, error) {
	if m.getCartItemRecordsFn != nil {
		return m.getCartItemRecordsFn(ctx, userID)
	}
	return nil, nil
}
func (m *mockCartRepository) UpsertCartItem(ctx context.Context, cartID uuid.UUID, bookID uuid.UUID, count int, unitPrice float64) error {
	return nil
}
func (m *mockCartRepository) RemoveCartItem(ctx context.Context, cartID uuid.UUID, bookID uuid.UUID) error {
	return nil
}
func (m *mockCartRepository) ClearCart(ctx context.Context, cartID uuid.UUID) error {
	if m.clearCartFunc != nil {
		return m.clearCartFunc(ctx, cartID)
	}
	return nil
}

type noopTransactionManager struct{}

func (n *noopTransactionManager) InTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func TestPtrPaymentStatus(t *testing.T) {
	status := ptrPaymentStatus(domain.PaymentPaid)
	if status == nil || *status != domain.PaymentPaid {
		t.Fatalf("unexpected pointer status value: %+v", status)
	}
}

func TestOrderListPassThrough(t *testing.T) {
	userID := uuid.New()
	orders := []domain.OrderView{{Order: domain.Order{ID: uuid.New()}}}
	repo := &mockOrderRepository{
		listOrdersByUserFunc: func(ctx context.Context, gotUserID uuid.UUID, limit, offset int) ([]domain.OrderView, error) {
			if gotUserID != userID || limit != 10 || offset != 5 {
				t.Fatalf("unexpected user list params")
			}
			return orders, nil
		},
		listOrdersFunc: func(ctx context.Context, limit, offset int) ([]domain.OrderView, error) {
			if limit != 20 || offset != 10 {
				t.Fatalf("unexpected list params")
			}
			return orders, nil
		},
	}

	svc := NewOrderService(&noopTransactionManager{}, repo, &noopCartRepository{})

	userOrders, err := svc.ListUserOrders(context.Background(), userID, 10, 5)
	if err != nil || len(userOrders) != 1 {
		t.Fatalf("unexpected ListUserOrders result: %+v, err=%v", userOrders, err)
	}

	allOrders, err := svc.ListAllOrders(context.Background(), 20, 10)
	if err != nil || len(allOrders) != 1 {
		t.Fatalf("unexpected ListAllOrders result: %+v, err=%v", allOrders, err)
	}
}

func TestValidateOrderForPaymentConfirmation(t *testing.T) {
	paid := domain.PaymentPaid
	pending := domain.PaymentPending

	tests := []struct {
		name    string
		order   domain.Order
		wantErr bool
	}{
		{
			name: "allows pending order with pending payment",
			order: domain.Order{
				Status:        domain.OrderPending,
				PaymentStatus: &pending,
			},
			wantErr: false,
		},
		{
			name: "allows pending order with nil payment status",
			order: domain.Order{
				Status:        domain.OrderPending,
				PaymentStatus: nil,
			},
			wantErr: false,
		},
		{
			name: "rejects completed order",
			order: domain.Order{
				Status:        domain.OrderCompleted,
				PaymentStatus: &pending,
			},
			wantErr: true,
		},
		{
			name: "rejects cancelled order",
			order: domain.Order{
				Status:        domain.OrderCancelled,
				PaymentStatus: &pending,
			},
			wantErr: true,
		},
		{
			name: "rejects pending order with paid status",
			order: domain.Order{
				Status:        domain.OrderPending,
				PaymentStatus: &paid,
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateOrderForPaymentConfirmation(tc.order)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error but got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected nil error but got %v", err)
			}
		})
	}
}

func TestCheckout_Success(t *testing.T) {
	userID := uuid.New()
	cartID := uuid.New()
	bookID := uuid.New()
	createdOrderID := uuid.Nil
	createdItems := 0

	svc := NewOrderService(
		&noopTransactionManager{},
		&mockOrderRepository{
			createOrderFunc: func(ctx context.Context, order *domain.Order) error {
				createdOrderID = order.ID
				return nil
			},
			createOrderItemsFunc: func(ctx context.Context, items []domain.OrderItem) error {
				createdItems = len(items)
				return nil
			},
			getOrderItemsFunc: func(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItemDetail, error) {
				if orderID != createdOrderID {
					t.Fatalf("unexpected order id")
				}
				return []domain.OrderItemDetail{{BookID: bookID, Count: 2, UnitPrice: 10, LineTotal: 20}}, nil
			},
		},
		&mockCartRepository{
			getOrCreateCartFunc: func(ctx context.Context, gotUserID uuid.UUID) (domain.Cart, error) {
				if gotUserID != userID {
					t.Fatalf("unexpected user id")
				}
				return domain.Cart{ID: cartID, UserID: userID}, nil
			},
			getCartItemRecordsFn: func(ctx context.Context, gotUserID uuid.UUID) ([]domain.CartItemRecord, error) {
				return []domain.CartItemRecord{{BookID: bookID, Count: 2, UnitPrice: 10, AvailableStock: 5}}, nil
			},
		},
	)

	view, err := svc.Checkout(context.Background(), userID, domain.CheckoutInput{PaymentMethod: domain.PaymentCOD})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if createdOrderID == uuid.Nil {
		t.Fatalf("expected created order id")
	}
	if createdItems != 1 {
		t.Fatalf("expected one created item, got %d", createdItems)
	}
	if view.Order.ID == uuid.Nil || len(view.Items) != 1 {
		t.Fatalf("unexpected checkout response: %+v", view)
	}
}

func TestCheckout_FailsOnEmptyCart(t *testing.T) {
	userID := uuid.New()
	svc := NewOrderService(
		&noopTransactionManager{},
		&mockOrderRepository{},
		&mockCartRepository{
			getOrCreateCartFunc: func(ctx context.Context, gotUserID uuid.UUID) (domain.Cart, error) {
				return domain.Cart{ID: uuid.New(), UserID: gotUserID}, nil
			},
			getCartItemRecordsFn: func(ctx context.Context, gotUserID uuid.UUID) ([]domain.CartItemRecord, error) {
				return nil, nil
			},
		},
	)

	_, err := svc.Checkout(context.Background(), userID, domain.CheckoutInput{PaymentMethod: domain.PaymentUPI})
	if err == nil || err.Error() != "cart is empty" {
		t.Fatalf("expected cart empty error, got %v", err)
	}
}

func TestConfirmPayment_SuccessPathClearsCart(t *testing.T) {
	userID := uuid.New()
	orderID := uuid.New()
	cartID := uuid.New()
	pending := domain.PaymentPending
	updatePaymentCalls := 0
	updateStatusCalls := 0
	decrementCalls := 0
	clearCalls := 0

	svc := NewOrderService(
		&noopTransactionManager{},
		&mockOrderRepository{
			getOrderByIDFunc: func(ctx context.Context, gotOrderID uuid.UUID) (domain.Order, error) {
				if gotOrderID != orderID {
					t.Fatalf("unexpected order id")
				}
				return domain.Order{
					ID:            orderID,
					UserID:        userID,
					Status:        domain.OrderPending,
					PaymentMethod: ptrPaymentMethod(domain.PaymentCOD),
					PaymentStatus: &pending,
				}, nil
			},
			getOrderItemsFunc: func(ctx context.Context, gotOrderID uuid.UUID) ([]domain.OrderItemDetail, error) {
				return []domain.OrderItemDetail{{BookID: uuid.New(), Count: 1, UnitPrice: 50, LineTotal: 50}}, nil
			},
			updateOrderPaymentFunc: func(ctx context.Context, gotOrderID uuid.UUID, status domain.PaymentStatus, method domain.PaymentMethod) error {
				updatePaymentCalls++
				if status != domain.PaymentPaid {
					t.Fatalf("expected paid status")
				}
				return nil
			},
			updateOrderStatusFunc: func(ctx context.Context, gotOrderID uuid.UUID, status domain.OrderStatus) error {
				updateStatusCalls++
				if status != domain.OrderCompleted {
					t.Fatalf("expected completed status")
				}
				return nil
			},
			decrementStockFunc: func(ctx context.Context, items []domain.OrderItem) error {
				decrementCalls++
				return nil
			},
		},
		&mockCartRepository{
			getOrCreateCartFunc: func(ctx context.Context, gotUserID uuid.UUID) (domain.Cart, error) {
				return domain.Cart{ID: cartID, UserID: gotUserID}, nil
			},
			clearCartFunc: func(ctx context.Context, gotCartID uuid.UUID) error {
				clearCalls++
				if gotCartID != cartID {
					t.Fatalf("unexpected cart id")
				}
				return nil
			},
		},
	)

	_, err := svc.ConfirmPayment(context.Background(), userID, domain.PaymentConfirmInput{OrderID: orderID, Success: true})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if updatePaymentCalls != 1 || updateStatusCalls != 1 || decrementCalls != 1 || clearCalls != 1 {
		t.Fatalf("unexpected call counts: payment=%d status=%d decrement=%d clear=%d", updatePaymentCalls, updateStatusCalls, decrementCalls, clearCalls)
	}
}

func TestConfirmPayment_FailurePathCancelsOrder(t *testing.T) {
	userID := uuid.New()
	orderID := uuid.New()
	pending := domain.PaymentPending
	paidCalled := false
	cancelCalled := false

	svc := NewOrderService(
		&noopTransactionManager{},
		&mockOrderRepository{
			getOrderByIDFunc: func(ctx context.Context, gotOrderID uuid.UUID) (domain.Order, error) {
				return domain.Order{
					ID:            orderID,
					UserID:        userID,
					Status:        domain.OrderPending,
					PaymentStatus: &pending,
				}, nil
			},
			getOrderItemsFunc: func(ctx context.Context, gotOrderID uuid.UUID) ([]domain.OrderItemDetail, error) {
				return []domain.OrderItemDetail{}, nil
			},
			updateOrderPaymentFunc: func(ctx context.Context, gotOrderID uuid.UUID, status domain.PaymentStatus, method domain.PaymentMethod) error {
				if status == domain.PaymentFailed {
					paidCalled = true
				}
				return nil
			},
			updateOrderStatusFunc: func(ctx context.Context, gotOrderID uuid.UUID, status domain.OrderStatus) error {
				if status == domain.OrderCancelled {
					cancelCalled = true
				}
				return nil
			},
		},
		&mockCartRepository{},
	)

	_, err := svc.ConfirmPayment(context.Background(), userID, domain.PaymentConfirmInput{OrderID: orderID, Success: false})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !paidCalled || !cancelCalled {
		t.Fatalf("expected payment failed and cancel updates")
	}
}

func TestConfirmPayment_RejectsUnauthorizedUser(t *testing.T) {
	userID := uuid.New()
	orderID := uuid.New()
	svc := NewOrderService(
		&noopTransactionManager{},
		&mockOrderRepository{
			getOrderByIDFunc: func(ctx context.Context, gotOrderID uuid.UUID) (domain.Order, error) {
				return domain.Order{ID: orderID, UserID: uuid.New(), Status: domain.OrderPending}, nil
			},
		},
		&mockCartRepository{},
	)

	_, err := svc.ConfirmPayment(context.Background(), userID, domain.PaymentConfirmInput{OrderID: orderID, Success: true})
	if err == nil || err.Error() != "not authorized to confirm this order" {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}

func ptrPaymentMethod(method domain.PaymentMethod) *domain.PaymentMethod {
	return &method
}
