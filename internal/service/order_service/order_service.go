package order_service

import (
	"context"
	"errors"
	"fmt"
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
			if err := s.orderRepo.UpdateOrderStatus(txCtx, order.ID, domain.OrderCompleted); err != nil {
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
			if err := s.orderRepo.UpdateOrderStatus(txCtx, order.ID, domain.OrderCancelled); err != nil {
				return err
			}
		}

		updated, err := s.orderRepo.GetOrderByID(txCtx, order.ID)
		if err != nil {
			return err
		}

		updatedItems, err := s.orderRepo.GetOrderItems(txCtx, order.ID)
		if err != nil {
			return err
		}

		orderView = domain.OrderView{
			Order: updated,
			Items: updatedItems,
		}
		return nil
	})

	return orderView, err
}

func validateOrderForPaymentConfirmation(order domain.Order) error {
	if order.Status != domain.OrderPending {
		return errors.New("order is already finalized")
	}

	if order.PaymentStatus != nil && *order.PaymentStatus != domain.PaymentPending {
		return errors.New("payment is already finalized")
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

func ptrPaymentStatus(status domain.PaymentStatus) *domain.PaymentStatus {
	return &status
}
