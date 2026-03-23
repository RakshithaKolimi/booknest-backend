package order_service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"booknest/internal/domain"
)

type orderService struct {
	txm       domain.TransactionManager
	orderRepo domain.OrderRepository
	cartRepo  domain.CartRepository
}

func NewOrderService(
	txm domain.TransactionManager,
	orderRepo domain.OrderRepository,
	cartRepo domain.CartRepository,
) domain.OrderService {
	return &orderService{
		txm:       txm,
		orderRepo: orderRepo,
		cartRepo:  cartRepo,
	}
}

func (s *orderService) Checkout(
	ctx context.Context,
	userID uuid.UUID,
	input domain.CheckoutInput,
) (domain.OrderView, error) {
	var orderView domain.OrderView

	err := s.txm.InTransaction(ctx, func(txCtx context.Context) error {
		cart, err := s.cartRepo.GetOrCreateCart(txCtx, userID)
		if err != nil {
			return err
		}

		items, err := s.cartRepo.GetCartItemRecords(txCtx, userID)
		if err != nil {
			return err
		}

		if len(items) == 0 {
			return errors.New("cart is empty")
		}

		total := 0.0
		orderItems := make([]domain.OrderItem, 0, len(items))
		for _, item := range items {
			if item.AvailableStock < item.Count {
				return fmt.Errorf("insufficient stock for book %s", item.BookID)
			}
			lineTotal := item.UnitPrice * float64(item.Count)
			total += lineTotal
			orderItems = append(orderItems, domain.OrderItem{
				OrderID:       uuid.Nil,
				BookID:        item.BookID,
				PurchaseCount: item.Count,
				PurchasePrice: item.UnitPrice,
				TotalPrice:    lineTotal,
			})
		}

		order := &domain.Order{
			ID:            uuid.New(),
			OrderNumber:   fmt.Sprintf("BN-%d", time.Now().UnixNano()),
			TotalPrice:    total,
			UserID:        userID,
			PaymentMethod: &input.PaymentMethod,
			PaymentStatus: ptrPaymentStatus(domain.PaymentPending),
			Status:        domain.OrderPending,
		}

		if err := s.orderRepo.CreateOrder(txCtx, order); err != nil {
			return err
		}

		for i := range orderItems {
			orderItems[i].OrderID = order.ID
		}

		if err := s.orderRepo.CreateOrderItems(txCtx, orderItems); err != nil {
			return err
		}

		orderItemsView, err := s.orderRepo.GetOrderItems(txCtx, order.ID)
		if err != nil {
			return err
		}

		orderView = domain.OrderView{
			Order: *order,
			Items: orderItemsView,
		}

		_ = cart
		return nil
	})

	return orderView, err
}

func (s *orderService) ConfirmPayment(
	ctx context.Context,
	userID uuid.UUID,
	input domain.PaymentConfirmInput,
) (domain.OrderView, error) {
	var orderView domain.OrderView

	err := s.txm.InTransaction(ctx, func(txCtx context.Context) error {
		order, err := s.orderRepo.GetOrderByID(txCtx, input.OrderID)
		if err != nil {
			return errors.New("order not found")
		}

		if order.UserID != userID {
			return errors.New("not authorized to confirm this order")
		}
		if err := validateOrderForPaymentConfirmation(order); err != nil {
			return err
		}

		paymentMethod := domain.PaymentCOD
		if order.PaymentMethod != nil {
			paymentMethod = *order.PaymentMethod
		}

		items, err := s.orderRepo.GetOrderItems(txCtx, order.ID)
		if err != nil {
			return err
		}

		if input.Success {
			if err := s.orderRepo.UpdateOrderPayment(txCtx, order.ID, domain.PaymentPaid, paymentMethod); err != nil {
				return err
			}

			stockItems := make([]domain.OrderItem, 0, len(items))
			for _, item := range items {
				stockItems = append(stockItems, domain.OrderItem{
					BookID:        item.BookID,
					PurchaseCount: item.Count,
				})
			}
			if err := s.orderRepo.DecrementStock(txCtx, stockItems); err != nil {
				return err
			}

			cart, err := s.cartRepo.GetOrCreateCart(txCtx, userID)
			if err != nil {
				return err
			}
			if err := s.cartRepo.ClearCart(txCtx, cart.ID); err != nil {
				return err
			}
		} else {
			if err := s.orderRepo.UpdateOrderPayment(txCtx, order.ID, domain.PaymentFailed, paymentMethod); err != nil {
				return err
			}
			if err := s.orderRepo.UpdateOrderStatus(txCtx, order.ID, domain.OrderFailed, nil); err != nil {
				return err
			}
		}

		orderView, err = s.getOrderView(txCtx, order.ID)
		return err
	})

	return orderView, err
}

func (s *orderService) CancelOrder(
	ctx context.Context,
	userID uuid.UUID,
	input domain.OrderCancelInput,
) (domain.OrderView, error) {
	var orderView domain.OrderView

	err := s.txm.InTransaction(ctx, func(txCtx context.Context) error {
		order, err := s.orderRepo.GetOrderByID(txCtx, input.OrderID)
		if err != nil {
			return errors.New("order not found")
		}
		if order.UserID != userID {
			return errors.New("not authorized to cancel this order")
		}

		reason := strings.TrimSpace(input.CancellationReason)
		if err := validateOrderForCancellation(order, reason); err != nil {
			return err
		}

		if shouldStartRefund(order.PaymentStatus) {
			paymentMethod := paymentMethodOrDefault(order.PaymentMethod)
			if err := s.orderRepo.UpdateOrderPayment(txCtx, order.ID, domain.PaymentRefundInitiated, paymentMethod); err != nil {
				return err
			}
		}

		if err := s.orderRepo.UpdateOrderStatus(txCtx, order.ID, domain.OrderCancelled, &reason); err != nil {
			return err
		}

		orderView, err = s.getOrderView(txCtx, order.ID)
		return err
	})

	return orderView, err
}

func (s *orderService) AdminUpdateOrderStatus(
	ctx context.Context,
	input domain.AdminOrderStatusUpdateInput,
) (domain.OrderView, error) {
	var orderView domain.OrderView

	err := s.txm.InTransaction(ctx, func(txCtx context.Context) error {
		order, err := s.orderRepo.GetOrderByID(txCtx, input.OrderID)
		if err != nil {
			return errors.New("order not found")
		}

		reason := strings.TrimSpace(input.CancellationReason)
		if err := validateAdminOrderStatusUpdate(order, input.Status, input.PaymentStatus, reason); err != nil {
			return err
		}

		if input.Status != "" {
			var reasonPtr *string
			if input.Status == domain.OrderCancelled {
				if shouldStartRefund(order.PaymentStatus) {
					paymentMethod := paymentMethodOrDefault(order.PaymentMethod)
					if err := s.orderRepo.UpdateOrderPayment(txCtx, order.ID, domain.PaymentRefundInitiated, paymentMethod); err != nil {
						return err
					}
				}
				reasonPtr = &reason
			}

			if err := s.orderRepo.UpdateOrderStatus(txCtx, order.ID, input.Status, reasonPtr); err != nil {
				return err
			}
		}

		if input.PaymentStatus != nil {
			paymentMethod := paymentMethodOrDefault(order.PaymentMethod)
			if err := s.orderRepo.UpdateOrderPayment(txCtx, order.ID, *input.PaymentStatus, paymentMethod); err != nil {
				return err
			}
		}

		orderView, err = s.getOrderView(txCtx, order.ID)
		return err
	})

	return orderView, err
}

func validateOrderForPaymentConfirmation(order domain.Order) error {
	if order.Status != domain.OrderPending {
		return errors.New("order is already finalized")
	}

	if order.PaymentStatus != nil &&
		*order.PaymentStatus != domain.PaymentPending &&
		*order.PaymentStatus != domain.PaymentFailed {
		return errors.New("payment is already finalized")
	}

	return nil
}

func validateOrderForCancellation(order domain.Order, reason string) error {
	if reason == "" {
		return errors.New("cancellation reason is required")
	}
	if order.Status == domain.OrderCompleted {
		return errors.New("completed orders cannot be cancelled")
	}
	if order.Status == domain.OrderCancelled {
		return errors.New("order is already cancelled")
	}
	return nil
}

func validateAdminOrderStatusUpdate(
	order domain.Order,
	nextStatus domain.OrderStatus,
	nextPaymentStatus *domain.PaymentStatus,
	reason string,
) error {
	if nextStatus == "" && nextPaymentStatus == nil {
		return errors.New("status or payment status is required")
	}

	if nextStatus != "" {
		if nextStatus != domain.OrderCompleted && nextStatus != domain.OrderCancelled {
			return errors.New("admin can only set order status to COMPLETED or CANCELLED")
		}
		if order.Status == domain.OrderCompleted || order.Status == domain.OrderCancelled {
			return errors.New("order is already finalized")
		}
		if nextStatus == domain.OrderCancelled && reason == "" {
			return errors.New("cancellation reason is required")
		}
	}

	if nextPaymentStatus != nil {
		if *nextPaymentStatus != domain.PaymentRefunded {
			return errors.New("admin can only set payment status to REFUNDED")
		}
		if order.PaymentStatus == nil || *order.PaymentStatus != domain.PaymentRefundInitiated {
			return errors.New("refund can only be completed after refund is initiated")
		}
	}

	return nil
}

func (s *orderService) ListUserOrders(
	ctx context.Context,
	userID uuid.UUID,
	limit,
	offset int,
) ([]domain.OrderView, error) {
	return s.orderRepo.ListOrdersByUser(ctx, userID, limit, offset)
}

func (s *orderService) ListAllOrders(
	ctx context.Context,
	limit,
	offset int,
) ([]domain.OrderView, error) {
	return s.orderRepo.ListOrders(ctx, limit, offset)
}

func (s *orderService) getOrderView(ctx context.Context, orderID uuid.UUID) (domain.OrderView, error) {
	updated, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return domain.OrderView{}, err
	}

	updatedItems, err := s.orderRepo.GetOrderItems(ctx, orderID)
	if err != nil {
		return domain.OrderView{}, err
	}

	return domain.OrderView{
		Order: updated,
		Items: updatedItems,
	}, nil
}

func ptrPaymentStatus(status domain.PaymentStatus) *domain.PaymentStatus {
	return &status
}

func paymentMethodOrDefault(method *domain.PaymentMethod) domain.PaymentMethod {
	if method != nil {
		return *method
	}
	return domain.PaymentCOD
}

func shouldStartRefund(paymentStatus *domain.PaymentStatus) bool {
	return paymentStatus != nil && *paymentStatus == domain.PaymentPaid
}
