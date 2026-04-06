package order_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"booknest/internal/domain"
)

type mockOrderRepository struct {
	createOrderFunc        func(ctx context.Context, order *domain.Order) error
	createOrderItemsFunc   func(ctx context.Context, items []domain.OrderItem) error
	getOrderByIDFunc       func(ctx context.Context, orderID uuid.UUID) (domain.Order, error)
	getOrderItemsFunc      func(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItemDetail, error)
	updateOrderPaymentFunc func(ctx context.Context, orderID uuid.UUID, status domain.PaymentStatus, method domain.PaymentMethod) error
	updateOrderStatusFunc  func(ctx context.Context, orderID uuid.UUID, status domain.OrderStatus, cancellationReason *string) error
	decrementStockFunc     func(ctx context.Context, items []domain.OrderItem) error
	listOrdersByUserFunc   func(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.OrderView, error)
	listOrdersFunc         func(ctx context.Context, limit, offset int) ([]domain.OrderView, error)
	hasUserPurchasedBookFn func(ctx context.Context, userID, bookID uuid.UUID) (bool, error)
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
func (m *mockOrderRepository) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status domain.OrderStatus, cancellationReason *string) error {
	if m.updateOrderStatusFunc != nil {
		return m.updateOrderStatusFunc(ctx, orderID, status, cancellationReason)
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
func (m *mockOrderRepository) HasUserPurchasedBook(ctx context.Context, userID, bookID uuid.UUID) (bool, error) {
	if m.hasUserPurchasedBookFn != nil {
		return m.hasUserPurchasedBookFn(ctx, userID, bookID)
	}
	return false, nil
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

type mockUserRepository struct {
	findByIDFunc               func(ctx context.Context, id uuid.UUID) (domain.User, error)
	getPreferencesByUserIDFunc func(ctx context.Context, userID uuid.UUID) (domain.UserPreferences, error)
	updatePreferencesFunc      func(ctx context.Context, prefs *domain.UserPreferences) error
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error { return nil }
func (m *mockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return domain.User{}, errors.New("not implemented")
}
func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	return domain.User{}, errors.New("not implemented")
}
func (m *mockUserRepository) FindByMobile(ctx context.Context, mobile string) (domain.User, error) {
	return domain.User{}, errors.New("not implemented")
}
func (m *mockUserRepository) GetPreferencesByUserID(ctx context.Context, userID uuid.UUID) (domain.UserPreferences, error) {
	if m.getPreferencesByUserIDFunc != nil {
		return m.getPreferencesByUserIDFunc(ctx, userID)
	}
	return domain.UserPreferences{}, errors.New("not implemented")
}

func (m *mockUserRepository) UpdatePreferences(ctx context.Context, prefs *domain.UserPreferences) error {
	if m.updatePreferencesFunc != nil {
		return m.updatePreferencesFunc(ctx, prefs)
	}
	return nil
}
func (m *mockUserRepository) Update(ctx context.Context, user *domain.User) error { return nil }
func (m *mockUserRepository) Delete(ctx context.Context, id uuid.UUID) error      { return nil }

type mockNotificationService struct {
	sendOTPFunc               func(phone string, otp string) error
	sendLoginAlertFunc        func(phone string, device string, location string) error
	sendOrderConfirmationFunc func(phone string, orderID string) error
	sendOrderCancellationFunc func(phone string, orderID string, reason string) error
	sendVerificationEmailFunc func(email string, link string) error
	sendPasswordResetFunc     func(email string, link string) error
	sendOrderReceiptFunc      func(email string, orderID string) error
}

func (m *mockNotificationService) SendOTP(phone string, otp string) error {
	if m.sendOTPFunc != nil {
		return m.sendOTPFunc(phone, otp)
	}
	return nil
}
func (m *mockNotificationService) SendLoginAlert(phone string, device string, location string) error {
	if m.sendLoginAlertFunc != nil {
		return m.sendLoginAlertFunc(phone, device, location)
	}
	return nil
}
func (m *mockNotificationService) SendOrderConfirmation(phone string, orderID string) error {
	if m.sendOrderConfirmationFunc != nil {
		return m.sendOrderConfirmationFunc(phone, orderID)
	}
	return nil
}
func (m *mockNotificationService) SendOrderCancellation(phone string, orderID string, reason string) error {
	if m.sendOrderCancellationFunc != nil {
		return m.sendOrderCancellationFunc(phone, orderID, reason)
	}
	return nil
}
func (m *mockNotificationService) SendVerificationEmail(email string, link string) error {
	if m.sendVerificationEmailFunc != nil {
		return m.sendVerificationEmailFunc(email, link)
	}
	return nil
}
func (m *mockNotificationService) SendPasswordReset(email string, link string) error {
	if m.sendPasswordResetFunc != nil {
		return m.sendPasswordResetFunc(email, link)
	}
	return nil
}
func (m *mockNotificationService) SendOrderReceipt(email string, orderID string) error {
	if m.sendOrderReceiptFunc != nil {
		return m.sendOrderReceiptFunc(email, orderID)
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
	failed := domain.PaymentFailed

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
			name: "allows pending order with failed payment",
			order: domain.Order{
				Status:        domain.OrderPending,
				PaymentStatus: &failed,
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
			name: "rejects failed order",
			order: domain.Order{
				Status:        domain.OrderFailed,
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
			updateOrderStatusFunc: func(ctx context.Context, gotOrderID uuid.UUID, status domain.OrderStatus, cancellationReason *string) error {
				updateStatusCalls++
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
	if updatePaymentCalls != 1 || updateStatusCalls != 0 || decrementCalls != 1 || clearCalls != 1 {
		t.Fatalf("unexpected call counts: payment=%d status=%d decrement=%d clear=%d", updatePaymentCalls, updateStatusCalls, decrementCalls, clearCalls)
	}
}

func TestConfirmPayment_FailurePathFailsOrder(t *testing.T) {
	userID := uuid.New()
	orderID := uuid.New()
	pending := domain.PaymentPending
	paymentFailedCalled := false
	orderFailedCalled := false

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
					paymentFailedCalled = true
				}
				return nil
			},
			updateOrderStatusFunc: func(ctx context.Context, gotOrderID uuid.UUID, status domain.OrderStatus, cancellationReason *string) error {
				if status == domain.OrderFailed {
					orderFailedCalled = true
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
	if !paymentFailedCalled || !orderFailedCalled {
		t.Fatalf("expected payment failed and order failed updates")
	}
}

func TestSendOrderNotifications_SendsEmailAndSMS(t *testing.T) {
	userID := uuid.New()
	emailCalled := false
	smsCalled := false

	svc := &orderService{
		userRepo: &mockUserRepository{
			findByIDFunc: func(ctx context.Context, id uuid.UUID) (domain.User, error) {
				if id != userID {
					t.Fatalf("unexpected user id: %s", id)
				}
				return domain.User{
					ID:     userID,
					Email:  "user@example.com",
					Mobile: "+15555550123",
				}, nil
			},
			getPreferencesByUserIDFunc: func(ctx context.Context, gotUserID uuid.UUID) (domain.UserPreferences, error) {
				if gotUserID != userID {
					t.Fatalf("unexpected user id for prefs: %s", gotUserID)
				}
				return domain.UserPreferences{UserID: userID, UseSMS: true}, nil
			},
		},
		notification: &mockNotificationService{
			sendOrderReceiptFunc: func(email string, orderID string) error {
				emailCalled = true
				if email != "user@example.com" || orderID != "BN-123" {
					t.Fatalf("unexpected email notification payload: %s %s", email, orderID)
				}
				return nil
			},
			sendOrderConfirmationFunc: func(phone string, orderID string) error {
				smsCalled = true
				if phone != "+15555550123" || orderID != "BN-123" {
					t.Fatalf("unexpected sms notification payload: %s %s", phone, orderID)
				}
				return nil
			},
		},
	}

	svc.sendOrderNotifications(context.Background(), userID, "BN-123")

	if !emailCalled || !smsCalled {
		t.Fatalf("expected both email and sms notifications, got email=%v sms=%v", emailCalled, smsCalled)
	}
}

func TestSendOrderNotifications_SkipsSMSWhenPreferenceDisabled(t *testing.T) {
	userID := uuid.New()
	emailCalled := false
	smsCalled := false

	svc := &orderService{
		userRepo: &mockUserRepository{
			findByIDFunc: func(ctx context.Context, id uuid.UUID) (domain.User, error) {
				return domain.User{
					ID:     userID,
					Email:  "user@example.com",
					Mobile: "+15555550123",
				}, nil
			},
			getPreferencesByUserIDFunc: func(ctx context.Context, gotUserID uuid.UUID) (domain.UserPreferences, error) {
				return domain.UserPreferences{UserID: gotUserID, UseSMS: false}, nil
			},
		},
		notification: &mockNotificationService{
			sendOrderReceiptFunc: func(email string, orderID string) error {
				emailCalled = true
				return nil
			},
			sendOrderConfirmationFunc: func(phone string, orderID string) error {
				smsCalled = true
				return nil
			},
		},
	}

	svc.sendOrderNotifications(context.Background(), userID, "BN-123")

	if !emailCalled {
		t.Fatalf("expected order receipt email to be sent")
	}
	if smsCalled {
		t.Fatalf("expected sms to be skipped when preference is disabled")
	}
}

func TestCancelOrder_StartsRefundForPaidOrder(t *testing.T) {
	userID := uuid.New()
	orderID := uuid.New()
	reason := "Need to cancel"
	paid := domain.PaymentPaid
	refundInitiatedCalled := false
	cancelCalled := false
	smsCalled := make(chan struct{}, 1)

	svc := NewOrderServiceWithNotification(
		&noopTransactionManager{},
		&mockOrderRepository{
			getOrderByIDFunc: func(ctx context.Context, gotOrderID uuid.UUID) (domain.Order, error) {
				return domain.Order{
					ID:            orderID,
					UserID:        userID,
					OrderNumber:   "BN-200",
					Status:        domain.OrderPending,
					PaymentMethod: ptrPaymentMethod(domain.PaymentUPI),
					PaymentStatus: &paid,
				}, nil
			},
			updateOrderPaymentFunc: func(ctx context.Context, gotOrderID uuid.UUID, status domain.PaymentStatus, method domain.PaymentMethod) error {
				if status == domain.PaymentRefundInitiated && method == domain.PaymentUPI {
					refundInitiatedCalled = true
				}
				return nil
			},
			updateOrderStatusFunc: func(ctx context.Context, gotOrderID uuid.UUID, status domain.OrderStatus, cancellationReason *string) error {
				if status == domain.OrderCancelled && cancellationReason != nil && *cancellationReason == reason {
					cancelCalled = true
				}
				return nil
			},
			getOrderItemsFunc: func(ctx context.Context, gotOrderID uuid.UUID) ([]domain.OrderItemDetail, error) {
				return []domain.OrderItemDetail{}, nil
			},
		},
		&mockCartRepository{},
		&mockUserRepository{
			findByIDFunc: func(ctx context.Context, gotUserID uuid.UUID) (domain.User, error) {
				if gotUserID != userID {
					t.Fatalf("unexpected user id: %s", gotUserID)
				}
				return domain.User{ID: userID, Mobile: "+15555550123"}, nil
			},
			getPreferencesByUserIDFunc: func(ctx context.Context, gotUserID uuid.UUID) (domain.UserPreferences, error) {
				return domain.UserPreferences{UserID: gotUserID, UseSMS: true}, nil
			},
		},
		&mockNotificationService{
			sendOrderCancellationFunc: func(phone string, orderNumber string, gotReason string) error {
				if phone != "+15555550123" || orderNumber != "BN-200" || gotReason != reason {
					t.Fatalf("unexpected cancellation sms payload: %s %s %s", phone, orderNumber, gotReason)
				}
				smsCalled <- struct{}{}
				return nil
			},
		},
	)

	_, err := svc.CancelOrder(context.Background(), userID, domain.OrderCancelInput{
		OrderID:            orderID,
		CancellationReason: reason,
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !refundInitiatedCalled || !cancelCalled {
		t.Fatalf("expected refund initiation and cancel status update")
	}
	select {
	case <-smsCalled:
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected cancellation sms to be sent")
	}
}

func TestCancelOrder_SkipsCancellationSMSWhenPreferenceDisabled(t *testing.T) {
	userID := uuid.New()
	orderID := uuid.New()
	reason := "Need to cancel"
	paid := domain.PaymentPaid
	smsCalled := make(chan struct{}, 1)

	svc := NewOrderServiceWithNotification(
		&noopTransactionManager{},
		&mockOrderRepository{
			getOrderByIDFunc: func(ctx context.Context, gotOrderID uuid.UUID) (domain.Order, error) {
				return domain.Order{
					ID:            orderID,
					UserID:        userID,
					OrderNumber:   "BN-201",
					Status:        domain.OrderPending,
					PaymentMethod: ptrPaymentMethod(domain.PaymentUPI),
					PaymentStatus: &paid,
				}, nil
			},
			updateOrderPaymentFunc: func(ctx context.Context, gotOrderID uuid.UUID, status domain.PaymentStatus, method domain.PaymentMethod) error {
				return nil
			},
			updateOrderStatusFunc: func(ctx context.Context, gotOrderID uuid.UUID, status domain.OrderStatus, cancellationReason *string) error {
				return nil
			},
			getOrderItemsFunc: func(ctx context.Context, gotOrderID uuid.UUID) ([]domain.OrderItemDetail, error) {
				return []domain.OrderItemDetail{}, nil
			},
		},
		&mockCartRepository{},
		&mockUserRepository{
			findByIDFunc: func(ctx context.Context, gotUserID uuid.UUID) (domain.User, error) {
				return domain.User{ID: userID, Mobile: "+15555550123"}, nil
			},
			getPreferencesByUserIDFunc: func(ctx context.Context, gotUserID uuid.UUID) (domain.UserPreferences, error) {
				return domain.UserPreferences{UserID: gotUserID, UseSMS: false}, nil
			},
		},
		&mockNotificationService{
			sendOrderCancellationFunc: func(phone string, orderNumber string, gotReason string) error {
				smsCalled <- struct{}{}
				return nil
			},
		},
	)

	_, err := svc.CancelOrder(context.Background(), userID, domain.OrderCancelInput{
		OrderID:            orderID,
		CancellationReason: reason,
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	select {
	case <-smsCalled:
		t.Fatalf("expected cancellation sms to be skipped")
	case <-time.After(200 * time.Millisecond):
	}
}

func TestValidateOrderForCancellation(t *testing.T) {
	tests := []struct {
		name    string
		order   domain.Order
		reason  string
		wantErr string
	}{
		{
			name:    "requires reason",
			order:   domain.Order{Status: domain.OrderPending},
			reason:  "",
			wantErr: "cancellation reason is required",
		},
		{
			name:    "rejects completed order",
			order:   domain.Order{Status: domain.OrderCompleted},
			reason:  "Changed plans",
			wantErr: "completed orders cannot be cancelled",
		},
		{
			name:    "rejects already cancelled order",
			order:   domain.Order{Status: domain.OrderCancelled},
			reason:  "Changed plans",
			wantErr: "order is already cancelled",
		},
		{
			name:   "allows failed order",
			order:  domain.Order{Status: domain.OrderFailed},
			reason: "Changed plans",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateOrderForCancellation(tc.order, tc.reason)
			if tc.wantErr == "" && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if tc.wantErr != "" && (err == nil || err.Error() != tc.wantErr) {
				t.Fatalf("expected error %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestValidateAdminOrderStatusUpdate(t *testing.T) {
	refundInitiated := domain.PaymentRefundInitiated
	refunded := domain.PaymentRefunded

	tests := []struct {
		name              string
		order             domain.Order
		nextStatus        domain.OrderStatus
		nextPaymentStatus *domain.PaymentStatus
		reason            string
		wantErr           string
	}{
		{
			name:    "requires at least one update",
			order:   domain.Order{Status: domain.OrderPending},
			wantErr: "status or payment status is required",
		},
		{
			name:       "rejects unsupported admin status",
			order:      domain.Order{Status: domain.OrderPending},
			nextStatus: domain.OrderFailed,
			wantErr:    "admin can only set order status to COMPLETED or CANCELLED",
		},
		{
			name:       "rejects finalized order status update",
			order:      domain.Order{Status: domain.OrderCompleted},
			nextStatus: domain.OrderCancelled,
			reason:     "Out of stock",
			wantErr:    "order is already finalized",
		},
		{
			name:       "requires cancellation reason",
			order:      domain.Order{Status: domain.OrderPending},
			nextStatus: domain.OrderCancelled,
			wantErr:    "cancellation reason is required",
		},
		{
			name:              "rejects invalid refund transition",
			order:             domain.Order{Status: domain.OrderCancelled},
			nextPaymentStatus: &refunded,
			wantErr:           "refund can only be completed after refund is initiated",
		},
		{
			name:              "allows refund completion after initiation",
			order:             domain.Order{Status: domain.OrderCancelled, PaymentStatus: &refundInitiated},
			nextPaymentStatus: &refunded,
		},
		{
			name:       "allows completion status",
			order:      domain.Order{Status: domain.OrderPending},
			nextStatus: domain.OrderCompleted,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateAdminOrderStatusUpdate(tc.order, tc.nextStatus, tc.nextPaymentStatus, tc.reason)
			if tc.wantErr == "" && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if tc.wantErr != "" && (err == nil || err.Error() != tc.wantErr) {
				t.Fatalf("expected error %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestPaymentMethodOrDefault(t *testing.T) {
	if got := paymentMethodOrDefault(nil); got != domain.PaymentCOD {
		t.Fatalf("expected COD default, got %s", got)
	}

	if got := paymentMethodOrDefault(ptrPaymentMethod(domain.PaymentCreditCard)); got != domain.PaymentCreditCard {
		t.Fatalf("expected card payment method, got %s", got)
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

func TestCancelOrder_RejectsUnauthorizedUser(t *testing.T) {
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

	_, err := svc.CancelOrder(context.Background(), userID, domain.OrderCancelInput{
		OrderID:            orderID,
		CancellationReason: "Changed my mind",
	})
	if err == nil || err.Error() != "not authorized to cancel this order" {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}

func TestCancelOrder_UserCanCancelFailedOrder(t *testing.T) {
	userID := uuid.New()
	orderID := uuid.New()
	reason := "Customer requested cancellation"
	cancelCalled := false
	smsCalled := make(chan struct{}, 1)

	svc := NewOrderServiceWithNotification(
		&noopTransactionManager{},
		&mockOrderRepository{
			getOrderByIDFunc: func(ctx context.Context, gotOrderID uuid.UUID) (domain.Order, error) {
				return domain.Order{ID: orderID, UserID: userID, OrderNumber: "BN-201", Status: domain.OrderFailed}, nil
			},
			updateOrderStatusFunc: func(ctx context.Context, gotOrderID uuid.UUID, status domain.OrderStatus, cancellationReason *string) error {
				cancelCalled = true
				if status != domain.OrderCancelled || cancellationReason == nil || *cancellationReason != reason {
					t.Fatalf("unexpected cancel payload")
				}
				return nil
			},
			getOrderItemsFunc: func(ctx context.Context, gotOrderID uuid.UUID) ([]domain.OrderItemDetail, error) {
				return []domain.OrderItemDetail{}, nil
			},
		},
		&mockCartRepository{},
		&mockUserRepository{
			findByIDFunc: func(ctx context.Context, gotUserID uuid.UUID) (domain.User, error) {
				return domain.User{ID: userID, Mobile: "+15555550123"}, nil
			},
			getPreferencesByUserIDFunc: func(ctx context.Context, gotUserID uuid.UUID) (domain.UserPreferences, error) {
				return domain.UserPreferences{UserID: gotUserID, UseSMS: true}, nil
			},
		},
		&mockNotificationService{
			sendOrderCancellationFunc: func(phone string, orderNumber string, gotReason string) error {
				if phone != "+15555550123" || orderNumber != "BN-201" || gotReason != reason {
					t.Fatalf("unexpected cancellation sms payload: %s %s %s", phone, orderNumber, gotReason)
				}
				smsCalled <- struct{}{}
				return nil
			},
		},
	)

	_, err := svc.CancelOrder(context.Background(), userID, domain.OrderCancelInput{
		OrderID:            orderID,
		CancellationReason: reason,
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !cancelCalled {
		t.Fatalf("expected cancel status update")
	}
	select {
	case <-smsCalled:
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected cancellation sms to be sent")
	}
}

func TestCancelOrder_UsesTrimmedReasonWithoutRefundWhenUnpaid(t *testing.T) {
	userID := uuid.New()
	orderID := uuid.New()
	pending := domain.PaymentPending
	paymentUpdated := false

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
			updateOrderPaymentFunc: func(ctx context.Context, gotOrderID uuid.UUID, status domain.PaymentStatus, method domain.PaymentMethod) error {
				paymentUpdated = true
				return nil
			},
			updateOrderStatusFunc: func(ctx context.Context, gotOrderID uuid.UUID, status domain.OrderStatus, cancellationReason *string) error {
				if status != domain.OrderCancelled {
					t.Fatalf("expected cancelled status, got %s", status)
				}
				if cancellationReason == nil || *cancellationReason != "Need to cancel" {
					t.Fatalf("expected trimmed cancellation reason, got %+v", cancellationReason)
				}
				return nil
			},
			getOrderItemsFunc: func(ctx context.Context, gotOrderID uuid.UUID) ([]domain.OrderItemDetail, error) {
				return []domain.OrderItemDetail{}, nil
			},
		},
		&mockCartRepository{},
	)

	_, err := svc.CancelOrder(context.Background(), userID, domain.OrderCancelInput{
		OrderID:            orderID,
		CancellationReason: "  Need to cancel  ",
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if paymentUpdated {
		t.Fatalf("did not expect refund initiation for unpaid order")
	}
}

func TestAdminUpdateOrderStatus_RequiresCancellationReason(t *testing.T) {
	orderID := uuid.New()

	svc := NewOrderService(
		&noopTransactionManager{},
		&mockOrderRepository{
			getOrderByIDFunc: func(ctx context.Context, gotOrderID uuid.UUID) (domain.Order, error) {
				return domain.Order{ID: orderID, Status: domain.OrderPending}, nil
			},
		},
		&mockCartRepository{},
	)

	_, err := svc.AdminUpdateOrderStatus(context.Background(), domain.AdminOrderStatusUpdateInput{
		OrderID: orderID,
		Status:  domain.OrderCancelled,
	})
	if err == nil || err.Error() != "cancellation reason is required" {
		t.Fatalf("expected cancellation reason error, got %v", err)
	}
}

func TestAdminUpdateOrderStatus_CompletesRefund(t *testing.T) {
	orderID := uuid.New()
	refundInitiated := domain.PaymentRefundInitiated
	refundedCalled := false

	svc := NewOrderService(
		&noopTransactionManager{},
		&mockOrderRepository{
			getOrderByIDFunc: func(ctx context.Context, gotOrderID uuid.UUID) (domain.Order, error) {
				return domain.Order{
					ID:            orderID,
					Status:        domain.OrderCancelled,
					PaymentMethod: ptrPaymentMethod(domain.PaymentCOD),
					PaymentStatus: &refundInitiated,
				}, nil
			},
			updateOrderPaymentFunc: func(ctx context.Context, gotOrderID uuid.UUID, status domain.PaymentStatus, method domain.PaymentMethod) error {
				if status == domain.PaymentRefunded && method == domain.PaymentCOD {
					refundedCalled = true
				}
				return nil
			},
			getOrderItemsFunc: func(ctx context.Context, gotOrderID uuid.UUID) ([]domain.OrderItemDetail, error) {
				return []domain.OrderItemDetail{}, nil
			},
		},
		&mockCartRepository{},
	)

	_, err := svc.AdminUpdateOrderStatus(context.Background(), domain.AdminOrderStatusUpdateInput{
		OrderID:       orderID,
		PaymentStatus: ptrPaymentStatus(domain.PaymentRefunded),
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !refundedCalled {
		t.Fatalf("expected refund completion update")
	}
}

func TestAdminUpdateOrderStatus_CancelsPaidOrderAndDefaultsPaymentMethod(t *testing.T) {
	orderID := uuid.New()
	paid := domain.PaymentPaid
	refundInitiatedCalled := false
	cancelCalled := false

	svc := NewOrderService(
		&noopTransactionManager{},
		&mockOrderRepository{
			getOrderByIDFunc: func(ctx context.Context, gotOrderID uuid.UUID) (domain.Order, error) {
				return domain.Order{
					ID:            orderID,
					Status:        domain.OrderPending,
					PaymentStatus: &paid,
				}, nil
			},
			updateOrderPaymentFunc: func(ctx context.Context, gotOrderID uuid.UUID, status domain.PaymentStatus, method domain.PaymentMethod) error {
				if status == domain.PaymentRefundInitiated && method == domain.PaymentCOD {
					refundInitiatedCalled = true
				}
				return nil
			},
			updateOrderStatusFunc: func(ctx context.Context, gotOrderID uuid.UUID, status domain.OrderStatus, cancellationReason *string) error {
				if status == domain.OrderCancelled && cancellationReason != nil && *cancellationReason == "Out of stock" {
					cancelCalled = true
				}
				return nil
			},
			getOrderItemsFunc: func(ctx context.Context, gotOrderID uuid.UUID) ([]domain.OrderItemDetail, error) {
				return []domain.OrderItemDetail{}, nil
			},
		},
		&mockCartRepository{},
	)

	_, err := svc.AdminUpdateOrderStatus(context.Background(), domain.AdminOrderStatusUpdateInput{
		OrderID:            orderID,
		Status:             domain.OrderCancelled,
		CancellationReason: "  Out of stock  ",
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !refundInitiatedCalled || !cancelCalled {
		t.Fatalf("expected refund initiation and cancel status update")
	}
}

func TestAdminUpdateOrderStatus_CompletesOrder(t *testing.T) {
	orderID := uuid.New()
	completedCalled := false

	svc := NewOrderService(
		&noopTransactionManager{},
		&mockOrderRepository{
			getOrderByIDFunc: func(ctx context.Context, gotOrderID uuid.UUID) (domain.Order, error) {
				return domain.Order{ID: orderID, Status: domain.OrderPending}, nil
			},
			updateOrderStatusFunc: func(ctx context.Context, gotOrderID uuid.UUID, status domain.OrderStatus, cancellationReason *string) error {
				completedCalled = true
				if status != domain.OrderCompleted || cancellationReason != nil {
					t.Fatalf("unexpected completion update payload")
				}
				return nil
			},
			getOrderItemsFunc: func(ctx context.Context, gotOrderID uuid.UUID) ([]domain.OrderItemDetail, error) {
				return []domain.OrderItemDetail{}, nil
			},
		},
		&mockCartRepository{},
	)

	_, err := svc.AdminUpdateOrderStatus(context.Background(), domain.AdminOrderStatusUpdateInput{
		OrderID: orderID,
		Status:  domain.OrderCompleted,
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !completedCalled {
		t.Fatalf("expected completed status update")
	}
}

func ptrPaymentMethod(method domain.PaymentMethod) *domain.PaymentMethod {
	return &method
}
