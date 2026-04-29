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
	cancelOrderFunc    func(ctx context.Context, in *orderv1.CancelOrderRequest, opts ...grpc.CallOption) (*orderv1.GetOrderResponse, error)
	adminUpdateFunc    func(ctx context.Context, in *orderv1.AdminUpdateOrderStatusRequest, opts ...grpc.CallOption) (*orderv1.GetOrderResponse, error)
	listAllOrdersFunc  func(ctx context.Context, in *orderv1.ListAllOrdersRequest, opts ...grpc.CallOption) (*orderv1.ListAllOrdersResponse, error)
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

func (m *mockRemoteOrderClient) CancelOrder(ctx context.Context, in *orderv1.CancelOrderRequest, opts ...grpc.CallOption) (*orderv1.GetOrderResponse, error) {
	if m.cancelOrderFunc != nil {
		return m.cancelOrderFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRemoteOrderClient) AdminUpdateOrderStatus(ctx context.Context, in *orderv1.AdminUpdateOrderStatusRequest, opts ...grpc.CallOption) (*orderv1.GetOrderResponse, error) {
	if m.adminUpdateFunc != nil {
		return m.adminUpdateFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRemoteOrderClient) ListAllOrders(ctx context.Context, in *orderv1.ListAllOrdersRequest, opts ...grpc.CallOption) (*orderv1.ListAllOrdersResponse, error) {
	if m.listAllOrdersFunc != nil {
		return m.listAllOrdersFunc(ctx, in, opts...)
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

func TestRemoteOrderServiceCancelUsesRemoteClient(t *testing.T) {
	userID := uuid.New()
	orderID := uuid.New()
	cancellationReason := "customer request"

	svc := newRemoteOrderServiceForTest(&mockRemoteOrderClient{
		cancelOrderFunc: func(ctx context.Context, in *orderv1.CancelOrderRequest, opts ...grpc.CallOption) (*orderv1.GetOrderResponse, error) {
			if in.GetUserId() != userID.String() {
				t.Fatalf("expected user id %s, got %s", userID, in.GetUserId())
			}
			if in.GetOrderId() != orderID.String() {
				t.Fatalf("expected order id %s, got %s", orderID, in.GetOrderId())
			}
			if in.GetCancellationReason() != cancellationReason {
				t.Fatalf("expected cancellation reason %q, got %q", cancellationReason, in.GetCancellationReason())
			}
			return &orderv1.GetOrderResponse{
				OrderId:            orderID.String(),
				OrderNumber:        "BN-3",
				UserId:             userID.String(),
				PaymentStatus:      string(domain.PaymentRefundInitiated),
				Status:             string(domain.OrderCancelled),
				CancellationReason: cancellationReason,
			}, nil
		},
	}, &panicFallbackOrderService{}, &noopCartRepository{}, nil, nil, nil)

	order, err := svc.CancelOrder(context.Background(), userID, domain.OrderCancelInput{
		OrderID:            orderID,
		CancellationReason: cancellationReason,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if order.Order.Status != domain.OrderCancelled {
		t.Fatalf("expected cancelled status, got %s", order.Order.Status)
	}
	if order.Order.CancellationReason == nil || *order.Order.CancellationReason != cancellationReason {
		t.Fatalf("expected cancellation reason %q, got %+v", cancellationReason, order.Order.CancellationReason)
	}
}

func TestRemoteOrderServiceAdminUpdateUsesRemoteClient(t *testing.T) {
	orderID := uuid.New()
	paymentStatus := domain.PaymentRefunded
	cancellationReason := "out of stock"

	svc := newRemoteOrderServiceForTest(&mockRemoteOrderClient{
		adminUpdateFunc: func(ctx context.Context, in *orderv1.AdminUpdateOrderStatusRequest, opts ...grpc.CallOption) (*orderv1.GetOrderResponse, error) {
			if in.GetOrderId() != orderID.String() {
				t.Fatalf("expected order id %s, got %s", orderID, in.GetOrderId())
			}
			if in.GetStatus() != string(domain.OrderCancelled) {
				t.Fatalf("expected status %s, got %s", domain.OrderCancelled, in.GetStatus())
			}
			if in.GetPaymentStatus() != string(paymentStatus) {
				t.Fatalf("expected payment status %s, got %s", paymentStatus, in.GetPaymentStatus())
			}
			if in.GetCancellationReason() != cancellationReason {
				t.Fatalf("expected cancellation reason %q, got %q", cancellationReason, in.GetCancellationReason())
			}
			return &orderv1.GetOrderResponse{
				OrderId:            orderID.String(),
				OrderNumber:        "BN-4",
				PaymentStatus:      string(paymentStatus),
				Status:             string(domain.OrderCancelled),
				CancellationReason: cancellationReason,
			}, nil
		},
	}, &panicFallbackOrderService{}, &noopCartRepository{}, nil, nil, nil)

	order, err := svc.AdminUpdateOrderStatus(context.Background(), domain.AdminOrderStatusUpdateInput{
		OrderID:            orderID,
		Status:             domain.OrderCancelled,
		PaymentStatus:      &paymentStatus,
		CancellationReason: cancellationReason,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if order.Order.PaymentStatus == nil || *order.Order.PaymentStatus != paymentStatus {
		t.Fatalf("expected payment status %s, got %+v", paymentStatus, order.Order.PaymentStatus)
	}
}

func TestRemoteOrderServiceListAllUsesRemoteClient(t *testing.T) {
	orderID := uuid.New()

	svc := newRemoteOrderServiceForTest(&mockRemoteOrderClient{
		listAllOrdersFunc: func(ctx context.Context, in *orderv1.ListAllOrdersRequest, opts ...grpc.CallOption) (*orderv1.ListAllOrdersResponse, error) {
			if in.GetLimit() != 25 {
				t.Fatalf("expected limit 25, got %d", in.GetLimit())
			}
			if in.GetOffset() != 10 {
				t.Fatalf("expected offset 10, got %d", in.GetOffset())
			}
			return &orderv1.ListAllOrdersResponse{
				Orders: []*orderv1.GetOrderResponse{{
					OrderId:       orderID.String(),
					OrderNumber:   "BN-5",
					PaymentStatus: string(domain.PaymentPending),
					Status:        string(domain.OrderPending),
					PaymentMethod: string(domain.PaymentUPI),
				}},
			}, nil
		},
	}, &panicFallbackOrderService{}, &noopCartRepository{}, nil, nil, nil)

	orders, err := svc.ListAllOrders(context.Background(), 25, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orders))
	}
	if orders[0].Order.PaymentMethod == nil || *orders[0].Order.PaymentMethod != domain.PaymentUPI {
		t.Fatalf("expected payment method %s, got %+v", domain.PaymentUPI, orders[0].Order.PaymentMethod)
	}
}
