package order_service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	orderv1 "booknest-order-service/gen/order/v1"
	"booknest/internal/domain"
)

type mockRemoteOrderClient struct {
	createOrderFunc    func(ctx context.Context, in *orderv1.CreateOrderRequest, opts ...grpc.CallOption) (*orderv1.CreateOrderResponse, error)
	getOrderFunc       func(ctx context.Context, in *orderv1.GetOrderRequest, opts ...grpc.CallOption) (*orderv1.GetOrderResponse, error)
	listOrdersFunc     func(ctx context.Context, in *orderv1.ListOrdersRequest, opts ...grpc.CallOption) (*orderv1.ListOrdersResponse, error)
	confirmPaymentFunc func(ctx context.Context, in *orderv1.ConfirmPaymentRequest, opts ...grpc.CallOption) (*orderv1.ConfirmPaymentResponse, error)
}

func (m *mockRemoteOrderClient) CreateOrder(ctx context.Context, in *orderv1.CreateOrderRequest, opts ...grpc.CallOption) (*orderv1.CreateOrderResponse, error) {
	if m.createOrderFunc != nil {
		return m.createOrderFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRemoteOrderClient) GetOrder(ctx context.Context, in *orderv1.GetOrderRequest, opts ...grpc.CallOption) (*orderv1.GetOrderResponse, error) {
	if m.getOrderFunc != nil {
		return m.getOrderFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRemoteOrderClient) ListOrders(ctx context.Context, in *orderv1.ListOrdersRequest, opts ...grpc.CallOption) (*orderv1.ListOrdersResponse, error) {
	if m.listOrdersFunc != nil {
		return m.listOrdersFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRemoteOrderClient) ConfirmPayment(ctx context.Context, in *orderv1.ConfirmPaymentRequest, opts ...grpc.CallOption) (*orderv1.ConfirmPaymentResponse, error) {
	if m.confirmPaymentFunc != nil {
		return m.confirmPaymentFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

type panicFallbackOrderService struct{}

func (p *panicFallbackOrderService) Checkout(ctx context.Context, userID uuid.UUID, input domain.CheckoutInput) (domain.OrderView, error) {
	panic("unexpected Checkout fallback call")
}

func (p *panicFallbackOrderService) ConfirmPayment(ctx context.Context, userID uuid.UUID, input domain.PaymentConfirmInput) (domain.OrderView, error) {
	panic("unexpected ConfirmPayment fallback call")
}

func (p *panicFallbackOrderService) CancelOrder(ctx context.Context, userID uuid.UUID, input domain.OrderCancelInput) (domain.OrderView, error) {
	panic("unexpected CancelOrder fallback call")
}

func (p *panicFallbackOrderService) AdminUpdateOrderStatus(ctx context.Context, input domain.AdminOrderStatusUpdateInput) (domain.OrderView, error) {
	panic("unexpected AdminUpdateOrderStatus fallback call")
}

func (p *panicFallbackOrderService) ListUserOrders(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.OrderView, error) {
	panic("unexpected ListUserOrders fallback call")
}

func (p *panicFallbackOrderService) ListAllOrders(ctx context.Context, limit, offset int) ([]domain.OrderView, error) {
	panic("unexpected ListAllOrders fallback call")
}

type mockFallbackOrderService struct {
	cancelOrderFunc func(ctx context.Context, userID uuid.UUID, input domain.OrderCancelInput) (domain.OrderView, error)
}

func (m *mockFallbackOrderService) Checkout(ctx context.Context, userID uuid.UUID, input domain.CheckoutInput) (domain.OrderView, error) {
	return domain.OrderView{}, errors.New("not implemented")
}

func (m *mockFallbackOrderService) ConfirmPayment(ctx context.Context, userID uuid.UUID, input domain.PaymentConfirmInput) (domain.OrderView, error) {
	return domain.OrderView{}, errors.New("not implemented")
}

func (m *mockFallbackOrderService) CancelOrder(ctx context.Context, userID uuid.UUID, input domain.OrderCancelInput) (domain.OrderView, error) {
	if m.cancelOrderFunc != nil {
		return m.cancelOrderFunc(ctx, userID, input)
	}
	return domain.OrderView{}, nil
}

func (m *mockFallbackOrderService) AdminUpdateOrderStatus(ctx context.Context, input domain.AdminOrderStatusUpdateInput) (domain.OrderView, error) {
	return domain.OrderView{}, errors.New("not implemented")
}

func (m *mockFallbackOrderService) ListUserOrders(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.OrderView, error) {
	return nil, errors.New("not implemented")
}

func (m *mockFallbackOrderService) ListAllOrders(ctx context.Context, limit, offset int) ([]domain.OrderView, error) {
	return nil, errors.New("not implemented")
}

func TestRemoteOrderServiceCheckoutUsesRemoteClient(t *testing.T) {
	userID := uuid.New()
	bookID := uuid.New()
	orderID := uuid.New()

	svc := newRemoteOrderServiceForTest(
		&mockRemoteOrderClient{
			createOrderFunc: func(ctx context.Context, in *orderv1.CreateOrderRequest, opts ...grpc.CallOption) (*orderv1.CreateOrderResponse, error) {
				if in.GetUserId() != userID.String() {
					t.Fatalf("expected user id %s, got %s", userID, in.GetUserId())
				}
				if got := len(in.GetItems()); got != 1 {
					t.Fatalf("expected 1 item, got %d", got)
				}
				return &orderv1.CreateOrderResponse{OrderId: orderID.String()}, nil
			},
			getOrderFunc: func(ctx context.Context, in *orderv1.GetOrderRequest, opts ...grpc.CallOption) (*orderv1.GetOrderResponse, error) {
				return &orderv1.GetOrderResponse{
					OrderId:       orderID.String(),
					OrderNumber:   "BN-1",
					UserId:        userID.String(),
					TotalPrice:    123.45,
					PaymentStatus: string(domain.PaymentPending),
					Status:        string(domain.OrderPending),
					Items: []*orderv1.OrderItem{{
						BookId:        bookID.String(),
						PurchaseCount: 2,
						PurchasePrice: 61.725,
						TotalPrice:    123.45,
					}},
				}, nil
			},
		},
		&panicFallbackOrderService{},
		&mockCartRepository{
			getCartItemRecordsFn: func(ctx context.Context, uid uuid.UUID) ([]domain.CartItemRecord, error) {
				return []domain.CartItemRecord{{
					BookID:         bookID,
					Count:          2,
					UnitPrice:      61.725,
					AvailableStock: 5,
				}}, nil
			},
		},
		nil,
		nil,
		nil,
	)

	order, err := svc.Checkout(context.Background(), userID, domain.CheckoutInput{PaymentMethod: domain.PaymentUPI})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if order.Order.ID != orderID {
		t.Fatalf("expected order id %s, got %s", orderID, order.Order.ID)
	}
	if len(order.Items) != 1 {
		t.Fatalf("expected 1 order item, got %d", len(order.Items))
	}
}

func TestRemoteOrderServiceConfirmPaymentClearsCart(t *testing.T) {
	userID := uuid.New()
	orderID := uuid.New()
	cartID := uuid.New()
	cleared := false

	svc := newRemoteOrderServiceForTest(
		&mockRemoteOrderClient{
			confirmPaymentFunc: func(ctx context.Context, in *orderv1.ConfirmPaymentRequest, opts ...grpc.CallOption) (*orderv1.ConfirmPaymentResponse, error) {
				return &orderv1.ConfirmPaymentResponse{OrderId: orderID.String()}, nil
			},
			getOrderFunc: func(ctx context.Context, in *orderv1.GetOrderRequest, opts ...grpc.CallOption) (*orderv1.GetOrderResponse, error) {
				return &orderv1.GetOrderResponse{
					OrderId:       orderID.String(),
					OrderNumber:   "BN-2",
					UserId:        userID.String(),
					TotalPrice:    50,
					PaymentStatus: string(domain.PaymentPaid),
					Status:        string(domain.OrderPending),
				}, nil
			},
		},
		&panicFallbackOrderService{},
		&mockCartRepository{
			getOrCreateCartFunc: func(ctx context.Context, uid uuid.UUID) (domain.Cart, error) {
				return domain.Cart{ID: cartID, UserID: userID}, nil
			},
			clearCartFunc: func(ctx context.Context, gotCartID uuid.UUID) error {
				if gotCartID != cartID {
					t.Fatalf("expected cart id %s, got %s", cartID, gotCartID)
				}
				cleared = true
				return nil
			},
		},
		nil,
		nil,
		nil,
	)

	_, err := svc.ConfirmPayment(context.Background(), userID, domain.PaymentConfirmInput{
		OrderID: orderID,
		Success: true,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !cleared {
		t.Fatal("expected cart to be cleared")
	}
}

func TestRemoteOrderServiceCancelFallsBackToMonolith(t *testing.T) {
	userID := uuid.New()
	orderID := uuid.New()
	called := false

	fallback := &mockFallbackOrderService{
		cancelOrderFunc: func(ctx context.Context, uid uuid.UUID, input domain.OrderCancelInput) (domain.OrderView, error) {
			called = true
			if uid != userID {
				t.Fatalf("expected user id %s, got %s", userID, uid)
			}
			if input.OrderID != orderID {
				t.Fatalf("expected order id %s, got %s", orderID, input.OrderID)
			}
			return domain.OrderView{}, nil
		},
	}

	svc := newRemoteOrderServiceForTest(&mockRemoteOrderClient{}, fallback, &noopCartRepository{}, nil, nil, nil)
	_, err := svc.CancelOrder(context.Background(), userID, domain.OrderCancelInput{
		OrderID:            orderID,
		CancellationReason: "customer request",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatal("expected monolith fallback to be called")
	}
}
