package order_service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpcstatus "google.golang.org/grpc/status"

	orderv1 "booknest-order-service/gen/order/v1"
	"booknest/internal/domain"
)

type grpcOrderClient interface {
	CreateOrder(ctx context.Context, in *orderv1.CreateOrderRequest, opts ...grpc.CallOption) (*orderv1.CreateOrderResponse, error)
	GetOrder(ctx context.Context, in *orderv1.GetOrderRequest, opts ...grpc.CallOption) (*orderv1.GetOrderResponse, error)
	ListOrders(ctx context.Context, in *orderv1.ListOrdersRequest, opts ...grpc.CallOption) (*orderv1.ListOrdersResponse, error)
	ConfirmPayment(ctx context.Context, in *orderv1.ConfirmPaymentRequest, opts ...grpc.CallOption) (*orderv1.ConfirmPaymentResponse, error)
	CancelOrder(ctx context.Context, in *orderv1.CancelOrderRequest, opts ...grpc.CallOption) (*orderv1.GetOrderResponse, error)
	AdminUpdateOrderStatus(ctx context.Context, in *orderv1.AdminUpdateOrderStatusRequest, opts ...grpc.CallOption) (*orderv1.GetOrderResponse, error)
	ListAllOrders(ctx context.Context, in *orderv1.ListAllOrdersRequest, opts ...grpc.CallOption) (*orderv1.ListAllOrdersResponse, error)
}

// remoteOrderService keeps order lifecycle ownership on the remote gRPC service,
// while still depending on local repositories for enrichment and side effects.
type remoteOrderService struct {
	client       grpcOrderClient
	conn         *grpc.ClientConn
	fallback     domain.OrderService
	cartRepo     domain.CartRepository
	bookRepo     domain.BookRepository
	userRepo     domain.UserRepository
	notification domain.NotificationService
}

// NewRemoteOrderService creates a new OrderService that forwards calls to a remote gRPC order service.
func NewRemoteOrderService(
	addr string,
	fallback domain.OrderService,
	cartRepo domain.CartRepository,
	bookRepo domain.BookRepository,
	userRepo domain.UserRepository,
	notification domain.NotificationService,
) (domain.OrderService, error) {
	// Trim spaces from the address to avoid connection issues due to accidental whitespace
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return nil, errors.New("ORDER_SERVICE_GRPC_ADDR is required when microservice mode is enabled")
	}

	// Use insecure credentials for simplicity, but in production you should use proper TLS credentials
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("connect order microservice: %w", err)
	}

	return &remoteOrderService{
		client:       orderv1.NewOrderServiceClient(conn),
		conn:         conn,
		fallback:     fallback,
		cartRepo:     cartRepo,
		bookRepo:     bookRepo,
		userRepo:     userRepo,
		notification: notification,
	}, nil
}

func newRemoteOrderServiceForTest(
	client grpcOrderClient,
	fallback domain.OrderService,
	cartRepo domain.CartRepository,
	bookRepo domain.BookRepository,
	userRepo domain.UserRepository,
	notification domain.NotificationService,
) domain.OrderService {
	return &remoteOrderService{
		client:       client,
		fallback:     fallback,
		cartRepo:     cartRepo,
		bookRepo:     bookRepo,
		userRepo:     userRepo,
		notification: notification,
	}
}

func (s *remoteOrderService) Checkout(
	ctx context.Context,
	userID uuid.UUID,
	input domain.CheckoutInput,
) (domain.OrderView, error) {
	// Get cart items from the local cart repository to include in the order creation request
	items, err := s.cartRepo.GetCartItemRecords(ctx, userID)
	if err != nil {
		return domain.OrderView{}, err
	}

	if len(items) == 0 {
		return domain.OrderView{}, errors.New("cart is empty")
	}

	pbItems := make([]*orderv1.OrderItem, 0, len(items))
	for _, item := range items {
		if item.AvailableStock < item.Count {
			return domain.OrderView{}, fmt.Errorf("insufficient stock for book %s", item.BookID)
		}

		// The remote service owns the order lifecycle, so we send a fully expanded
		// snapshot of the cart items instead of expecting it to re-read local cart state.
		pbItems = append(pbItems, &orderv1.OrderItem{
			BookId:        item.BookID.String(),
			PurchaseCount: int32(item.Count),
			PurchasePrice: item.UnitPrice,
			TotalPrice:    item.UnitPrice * float64(item.Count),
		})
	}

	// Forward the checkout request to the remote order service via gRPC
	resp, err := s.client.CreateOrder(ctx, &orderv1.CreateOrderRequest{
		UserId:        userID.String(),
		PaymentMethod: string(input.PaymentMethod),
		Items:         pbItems,
	})
	if err != nil {
		return domain.OrderView{}, mapRemoteError("create order", err)
	}

	slog.Info("forwarded checkout to order microservice", "userID", userID.String(), "orderID", resp.GetOrderId())
	return s.fetchOrderView(ctx, resp.GetOrderId())
}

func (s *remoteOrderService) ConfirmPayment(
	ctx context.Context,
	userID uuid.UUID,
	input domain.PaymentConfirmInput,
) (domain.OrderView, error) {
	_, err := s.client.ConfirmPayment(ctx, &orderv1.ConfirmPaymentRequest{
		UserId:  userID.String(),
		OrderId: input.OrderID.String(),
		Success: input.Success,
	})
	if err != nil {
		return domain.OrderView{}, mapRemoteError("confirm payment", err)
	}

	slog.Info("forwarded payment confirmation to order microservice", "userID", userID.String(), "orderID", input.OrderID.String(), "success", input.Success)
	orderView, err := s.fetchOrderView(ctx, input.OrderID.String())
	if err != nil {
		return domain.OrderView{}, err
	}

	if input.Success {
		// These side effects are intentionally best-effort because payment has
		// already been confirmed by the remote service at this point.
		s.clearCartBestEffort(ctx, userID)
		go s.sendOrderNotifications(ctx, orderView.Order.UserID, orderView.Order.OrderNumber)
	}

	return orderView, nil
}

func (s *remoteOrderService) CancelOrder(
	ctx context.Context,
	userID uuid.UUID,
	input domain.OrderCancelInput,
) (domain.OrderView, error) {
	resp, err := s.client.CancelOrder(ctx, &orderv1.CancelOrderRequest{
		UserId:             userID.String(),
		OrderId:            input.OrderID.String(),
		CancellationReason: input.CancellationReason,
	})
	if err != nil {
		return domain.OrderView{}, mapRemoteError("cancel order", err)
	}

	slog.Info("forwarded order cancellation to order microservice", "userID", userID.String(), "orderID", input.OrderID.String())
	return s.mapOrderView(ctx, resp), nil
}

func (s *remoteOrderService) AdminUpdateOrderStatus(
	ctx context.Context,
	input domain.AdminOrderStatusUpdateInput,
) (domain.OrderView, error) {
	req := &orderv1.AdminUpdateOrderStatusRequest{
		OrderId:            input.OrderID.String(),
		Status:             string(input.Status),
		CancellationReason: input.CancellationReason,
	}
	if input.PaymentStatus != nil {
		req.PaymentStatus = string(*input.PaymentStatus)
	}

	resp, err := s.client.AdminUpdateOrderStatus(ctx, req)
	if err != nil {
		return domain.OrderView{}, mapRemoteError("admin update order status", err)
	}

	slog.Info("forwarded admin order update to order microservice", "orderID", input.OrderID.String())
	return s.mapOrderView(ctx, resp), nil
}

func (s *remoteOrderService) ListUserOrders(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
) ([]domain.OrderView, error) {
	resp, err := s.client.ListOrders(ctx, &orderv1.ListOrdersRequest{
		UserId: userID.String(),
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, mapRemoteError("list orders", err)
	}

	orders := make([]domain.OrderView, 0, len(resp.GetOrders()))
	for _, order := range resp.GetOrders() {
		orders = append(orders, s.mapOrderView(ctx, order))
	}

	return orders, nil
}

func (s *remoteOrderService) ListAllOrders(
	ctx context.Context,
	limit, offset int,
) ([]domain.OrderView, error) {
	resp, err := s.client.ListAllOrders(ctx, &orderv1.ListAllOrdersRequest{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, mapRemoteError("list all orders", err)
	}

	orders := make([]domain.OrderView, 0, len(resp.GetOrders()))
	for _, order := range resp.GetOrders() {
		orders = append(orders, s.mapOrderView(ctx, order))
	}

	slog.Info("listed all orders via order microservice", "limit", limit, "offset", offset, "count", len(orders))
	return orders, nil
}

func (s *remoteOrderService) fetchOrderView(ctx context.Context, orderID string) (domain.OrderView, error) {
	resp, err := s.client.GetOrder(ctx, &orderv1.GetOrderRequest{OrderId: orderID})
	if err != nil {
		return domain.OrderView{}, mapRemoteError("get order", err)
	}

	return s.mapOrderView(ctx, resp), nil
}

func (s *remoteOrderService) mapOrderView(ctx context.Context, resp *orderv1.GetOrderResponse) domain.OrderView {
	// Ignore parse failures here and let zero values flow through so a malformed
	// remote payload does not prevent callers from seeing the rest of the order.
	orderID, _ := uuid.Parse(resp.GetOrderId())
	userID, _ := uuid.Parse(resp.GetUserId())

	order := domain.Order{
		ID:                 orderID,
		OrderNumber:        resp.GetOrderNumber(),
		TotalPrice:         resp.GetTotalPrice(),
		UserID:             userID,
		PaymentMethod:      parsePaymentMethodPtr(resp.GetPaymentMethod()),
		PaymentStatus:      parsePaymentStatusPtr(resp.GetPaymentStatus()),
		Status:             parseOrderStatus(resp.GetStatus()),
		CancellationReason: parseOptionalString(resp.GetCancellationReason()),
	}

	items := make([]domain.OrderItemDetail, 0, len(resp.GetItems()))
	for _, item := range resp.GetItems() {
		bookID, _ := uuid.Parse(item.GetBookId())

		detail := domain.OrderItemDetail{
			BookID:    bookID,
			UnitPrice: item.GetPurchasePrice(),
			Count:     int(item.GetPurchaseCount()),
			LineTotal: item.GetTotalPrice(),
		}

		// Book metadata remains local to this service, so enrich the remote order
		// response when the repository is available.
		if s.bookRepo != nil {
			if book, err := s.bookRepo.FindByID(ctx, bookID); err == nil && book != nil {
				detail.Name = book.Name
				detail.ImageURL = book.ImageURL
			}
		}

		items = append(items, detail)
	}

	return domain.OrderView{
		Order: order,
		Items: items,
	}
}

func parseOptionalString(raw string) *string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func (s *remoteOrderService) clearCartBestEffort(ctx context.Context, userID uuid.UUID) {
	// Cart cleanup should never turn a successful payment flow into an error path.
	cart, err := s.cartRepo.GetOrCreateCart(ctx, userID)
	if err != nil {
		slog.Warn("clear local cart after remote payment failed", "userID", userID.String(), "err", err)
		return
	}

	if err := s.cartRepo.ClearCart(ctx, cart.ID); err != nil {
		slog.Warn("clear local cart after remote payment failed", "userID", userID.String(), "err", err)
	}
}

func (s *remoteOrderService) sendOrderNotifications(ctx context.Context, userID uuid.UUID, orderID string) {
	// Notification delivery is intentionally silent on failure to avoid retry-style
	// behavior affecting the main checkout response path.
	if s.userRepo == nil || s.notification == nil || orderID == "" {
		return
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return
	}

	if err := s.notification.SendOrderReceipt(user.Email, orderID); err != nil {
		return
	}

	if strings.TrimSpace(user.Mobile) == "" {
		return
	}

	prefs, err := s.userRepo.GetPreferencesByUserID(ctx, userID)
	if err != nil || !prefs.UseSMS {
		return
	}

	_ = s.notification.SendOrderConfirmation(user.Mobile, orderID)
}

func mapRemoteError(action string, err error) error {
	if err == nil {
		return nil
	}

	// Prefer the gRPC status message so callers see the domain-level error
	// produced by the remote service instead of a transport-heavy wrapper.
	st, ok := grpcstatus.FromError(err)
	if ok && st.Message() != "" {
		return fmt.Errorf("%s: %s", action, st.Message())
	}

	return fmt.Errorf("%s: %w", action, err)
}

func parsePaymentStatusPtr(raw string) *domain.PaymentStatus {
	// Normalize remote enum strings defensively because protobuf JSON/text
	// representations may vary in case or contain extra whitespace.
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case string(domain.PaymentPending):
		status := domain.PaymentPending
		return &status
	case string(domain.PaymentPaid):
		status := domain.PaymentPaid
		return &status
	case string(domain.PaymentRefundInitiated):
		status := domain.PaymentRefundInitiated
		return &status
	case string(domain.PaymentRefunded):
		status := domain.PaymentRefunded
		return &status
	case string(domain.PaymentFailed):
		status := domain.PaymentFailed
		return &status
	default:
		return nil
	}
}

func parsePaymentMethodPtr(raw string) *domain.PaymentMethod {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case string(domain.PaymentCOD):
		method := domain.PaymentCOD
		return &method
	case string(domain.PaymentCreditCard):
		method := domain.PaymentCreditCard
		return &method
	case string(domain.PaymentDebitCard):
		method := domain.PaymentDebitCard
		return &method
	case string(domain.PaymentNetBanking):
		method := domain.PaymentNetBanking
		return &method
	case string(domain.PaymentUPI):
		method := domain.PaymentUPI
		return &method
	default:
		return nil
	}
}

func parseOrderStatus(raw string) domain.OrderStatus {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case string(domain.OrderPending):
		return domain.OrderPending
	case string(domain.OrderFailed):
		return domain.OrderFailed
	case string(domain.OrderCancelled):
		return domain.OrderCancelled
	case string(domain.OrderCompleted):
		return domain.OrderCompleted
	default:
		return ""
	}
}
