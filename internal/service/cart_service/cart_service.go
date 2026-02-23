package cart_service

import (
	"context"
	"errors"
	"math"

	"github.com/google/uuid"

	"booknest/internal/domain"
)

type cartService struct {
	cartRepo domain.CartRepository
	bookRepo domain.BookRepository
}

func NewCartService(
	cartRepo domain.CartRepository,
	bookRepo domain.BookRepository,
) domain.CartService {
	return &cartService{
		cartRepo: cartRepo,
		bookRepo: bookRepo,
	}
}

func (s *cartService) GetCart(
	ctx context.Context,
	userID uuid.UUID,
) (domain.CartView, error) {
	cart, err := s.cartRepo.GetOrCreateCart(ctx, userID)
	if err != nil {
		return domain.CartView{}, err
	}

	items, err := s.cartRepo.GetCartItems(ctx, userID)
	if err != nil {
		return domain.CartView{}, err
	}

	return buildCartView(cart, items), nil
}

func (s *cartService) AddItem(
	ctx context.Context,
	userID uuid.UUID,
	input domain.CartItemInput,
) (domain.CartView, error) {
	return s.upsertItem(ctx, userID, input)
}

func (s *cartService) UpdateItem(
	ctx context.Context,
	userID uuid.UUID,
	input domain.CartItemInput,
) (domain.CartView, error) {
	return s.upsertItem(ctx, userID, input)
}

func (s *cartService) RemoveItem(
	ctx context.Context,
	userID uuid.UUID,
	bookID uuid.UUID,
) (domain.CartView, error) {
	cart, err := s.cartRepo.GetOrCreateCart(ctx, userID)
	if err != nil {
		return domain.CartView{}, err
	}

	if err := s.cartRepo.RemoveCartItem(ctx, cart.ID, bookID); err != nil {
		return domain.CartView{}, err
	}

	items, err := s.cartRepo.GetCartItems(ctx, userID)
	if err != nil {
		return domain.CartView{}, err
	}

	return buildCartView(cart, items), nil
}

func (s *cartService) Clear(ctx context.Context, userID uuid.UUID) error {
	cart, err := s.cartRepo.GetOrCreateCart(ctx, userID)
	if err != nil {
		return err
	}
	return s.cartRepo.ClearCart(ctx, cart.ID)
}

func (s *cartService) upsertItem(
	ctx context.Context,
	userID uuid.UUID,
	input domain.CartItemInput,
) (domain.CartView, error) {
	if input.Count <= 0 {
		return domain.CartView{}, errors.New("count must be greater than 0")
	}

	book, err := s.bookRepo.FindByID(ctx, input.BookID)
	if err != nil {
		return domain.CartView{}, errors.New("book not found")
	}

	if !book.IsActive {
		return domain.CartView{}, errors.New("book is not active")
	}

	if book.AvailableStock < input.Count {
		return domain.CartView{}, errors.New("insufficient stock")
	}

	unitPrice := book.Price * (1 - (book.DiscountPercentage / 100))
	unitPrice = math.Round(unitPrice*100) / 100

	cart, err := s.cartRepo.GetOrCreateCart(ctx, userID)
	if err != nil {
		return domain.CartView{}, err
	}

	if err := s.cartRepo.UpsertCartItem(ctx, cart.ID, book.ID, input.Count, unitPrice); err != nil {
		return domain.CartView{}, err
	}

	items, err := s.cartRepo.GetCartItems(ctx, userID)
	if err != nil {
		return domain.CartView{}, err
	}

	return buildCartView(cart, items), nil
}

func buildCartView(
	cart domain.Cart,
	items []domain.CartItemDetail,
) domain.CartView {
	totalItems := 0
	subtotal := 0.0
	for i := range items {
		totalItems += items[i].Count
		subtotal += items[i].LineTotal
	}

	return domain.CartView{
		CartID:     cart.ID,
		UserID:     cart.UserID,
		Items:      items,
		Subtotal:   subtotal,
		TotalItems: totalItems,
	}
}
