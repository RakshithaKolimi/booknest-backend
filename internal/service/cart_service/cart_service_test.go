package cart_service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"booknest/internal/domain"
)

type mockCartRepository struct {
	getOrCreateCartFunc   func(ctx context.Context, userID uuid.UUID) (domain.Cart, error)
	getCartItemsFunc      func(ctx context.Context, userID uuid.UUID) ([]domain.CartItemDetail, error)
	getCartItemRecordsFun func(ctx context.Context, userID uuid.UUID) ([]domain.CartItemRecord, error)
	upsertCartItemFunc    func(ctx context.Context, cartID uuid.UUID, bookID uuid.UUID, count int, unitPrice float64) error
	removeCartItemFunc    func(ctx context.Context, cartID uuid.UUID, bookID uuid.UUID) error
	clearCartFunc         func(ctx context.Context, cartID uuid.UUID) error
}

func (m *mockCartRepository) GetOrCreateCart(ctx context.Context, userID uuid.UUID) (domain.Cart, error) {
	if m.getOrCreateCartFunc != nil {
		return m.getOrCreateCartFunc(ctx, userID)
	}
	return domain.Cart{}, errors.New("not implemented")
}

func (m *mockCartRepository) GetCartItems(ctx context.Context, userID uuid.UUID) ([]domain.CartItemDetail, error) {
	if m.getCartItemsFunc != nil {
		return m.getCartItemsFunc(ctx, userID)
	}
	return []domain.CartItemDetail{}, nil
}

func (m *mockCartRepository) GetCartItemRecords(ctx context.Context, userID uuid.UUID) ([]domain.CartItemRecord, error) {
	if m.getCartItemRecordsFun != nil {
		return m.getCartItemRecordsFun(ctx, userID)
	}
	return []domain.CartItemRecord{}, nil
}

func (m *mockCartRepository) UpsertCartItem(ctx context.Context, cartID uuid.UUID, bookID uuid.UUID, count int, unitPrice float64) error {
	if m.upsertCartItemFunc != nil {
		return m.upsertCartItemFunc(ctx, cartID, bookID, count, unitPrice)
	}
	return nil
}

func (m *mockCartRepository) RemoveCartItem(ctx context.Context, cartID uuid.UUID, bookID uuid.UUID) error {
	if m.removeCartItemFunc != nil {
		return m.removeCartItemFunc(ctx, cartID, bookID)
	}
	return nil
}

func (m *mockCartRepository) ClearCart(ctx context.Context, cartID uuid.UUID) error {
	if m.clearCartFunc != nil {
		return m.clearCartFunc(ctx, cartID)
	}
	return nil
}

type mockBookRepository struct {
	findByIDFunc func(ctx context.Context, id uuid.UUID) (*domain.Book, error)
}

func (m *mockBookRepository) Create(ctx context.Context, book *domain.Book) error { return nil }
func (m *mockBookRepository) CreateWithRelations(ctx context.Context, input domain.BookInput) (*domain.Book, error) {
	return nil, nil
}
func (m *mockBookRepository) List(ctx context.Context, limit, offset int) ([]domain.Book, error) {
	return nil, nil
}
func (m *mockBookRepository) FilterByCriteria(ctx context.Context, filter domain.BookFilter, pagination domain.QueryOptions) ([]domain.Book, int64, error) {
	return nil, 0, nil
}
func (m *mockBookRepository) Update(ctx context.Context, book *domain.Book) error { return nil }
func (m *mockBookRepository) UpdateWithRelations(ctx context.Context, id uuid.UUID, input domain.BookInput) (*domain.Book, error) {
	return nil, nil
}
func (m *mockBookRepository) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockBookRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func TestBuildCartViewComputesTotals(t *testing.T) {
	cartID := uuid.New()
	userID := uuid.New()
	view := buildCartView(domain.Cart{ID: cartID, UserID: userID}, []domain.CartItemDetail{{Count: 2, LineTotal: 150.50}, {Count: 1, LineTotal: 49.50}})

	if view.CartID != cartID || view.UserID != userID {
		t.Fatalf("unexpected cart identity in view: %+v", view)
	}
	if view.TotalItems != 3 {
		t.Fatalf("expected 3 items, got %d", view.TotalItems)
	}
	if view.Subtotal != 200 {
		t.Fatalf("expected subtotal 200, got %f", view.Subtotal)
	}
}

func TestUpsertItemValidationFailures(t *testing.T) {
	svc := &cartService{}
	userID := uuid.New()
	bookID := uuid.New()

	_, err := svc.upsertItem(context.Background(), userID, domain.CartItemInput{BookID: bookID, Count: 0})
	if err == nil || err.Error() != "count must be greater than 0" {
		t.Fatalf("expected invalid count error, got %v", err)
	}

	svc.bookRepo = &mockBookRepository{findByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
		return nil, errors.New("missing")
	}}
	_, err = svc.upsertItem(context.Background(), userID, domain.CartItemInput{BookID: bookID, Count: 1})
	if err == nil || err.Error() != "book not found" {
		t.Fatalf("expected book not found error, got %v", err)
	}
}

func TestUpsertItemBookStateFailures(t *testing.T) {
	userID := uuid.New()
	bookID := uuid.New()

	svc := &cartService{
		bookRepo: &mockBookRepository{findByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
			return &domain.Book{ID: bookID, IsActive: false, AvailableStock: 10, Price: 50}, nil
		}},
	}

	_, err := svc.upsertItem(context.Background(), userID, domain.CartItemInput{BookID: bookID, Count: 1})
	if err == nil || err.Error() != "book is not active" {
		t.Fatalf("expected inactive book error, got %v", err)
	}

	svc.bookRepo = &mockBookRepository{findByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
		return &domain.Book{ID: bookID, IsActive: true, AvailableStock: 1, Price: 50}, nil
	}}
	_, err = svc.upsertItem(context.Background(), userID, domain.CartItemInput{BookID: bookID, Count: 2})
	if err == nil || err.Error() != "insufficient stock" {
		t.Fatalf("expected insufficient stock error, got %v", err)
	}
}

func TestUpsertItemSuccessRoundsDiscountedPrice(t *testing.T) {
	userID := uuid.New()
	cartID := uuid.New()
	bookID := uuid.New()
	called := false

	svc := &cartService{
		cartRepo: &mockCartRepository{
			getOrCreateCartFunc: func(ctx context.Context, gotUserID uuid.UUID) (domain.Cart, error) {
				if gotUserID != userID {
					t.Fatalf("unexpected userID: %s", gotUserID)
				}
				return domain.Cart{ID: cartID, UserID: userID}, nil
			},
			upsertCartItemFunc: func(ctx context.Context, gotCartID uuid.UUID, gotBookID uuid.UUID, count int, unitPrice float64) error {
				called = true
				if gotCartID != cartID || gotBookID != bookID || count != 2 {
					t.Fatalf("unexpected upsert payload")
				}
				if unitPrice != 84.99 {
					t.Fatalf("expected rounded unit price 84.99, got %f", unitPrice)
				}
				return nil
			},
			getCartItemsFunc: func(ctx context.Context, gotUserID uuid.UUID) ([]domain.CartItemDetail, error) {
				return []domain.CartItemDetail{{BookID: bookID, Count: 2, UnitPrice: 84.99, LineTotal: 169.98}}, nil
			},
		},
		bookRepo: &mockBookRepository{findByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
			return &domain.Book{ID: bookID, IsActive: true, AvailableStock: 10, Price: 99.99, DiscountPercentage: 15}, nil
		}},
	}

	view, err := svc.upsertItem(context.Background(), userID, domain.CartItemInput{BookID: bookID, Count: 2})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !called {
		t.Fatalf("expected upsert to be called")
	}
	if view.TotalItems != 2 || view.Subtotal != 169.98 {
		t.Fatalf("unexpected cart view: %+v", view)
	}
}

func TestGetCartAndRemoveAndClear(t *testing.T) {
	userID := uuid.New()
	cartID := uuid.New()
	bookID := uuid.New()
	removed := false
	cleared := false

	repo := &mockCartRepository{
		getOrCreateCartFunc: func(ctx context.Context, gotUserID uuid.UUID) (domain.Cart, error) {
			return domain.Cart{ID: cartID, UserID: userID}, nil
		},
		getCartItemsFunc: func(ctx context.Context, gotUserID uuid.UUID) ([]domain.CartItemDetail, error) {
			return []domain.CartItemDetail{{BookID: bookID, Count: 1, LineTotal: 99}}, nil
		},
		removeCartItemFunc: func(ctx context.Context, gotCartID uuid.UUID, gotBookID uuid.UUID) error {
			removed = true
			return nil
		},
		clearCartFunc: func(ctx context.Context, gotCartID uuid.UUID) error {
			cleared = true
			return nil
		},
	}

	svc := NewCartService(repo, &mockBookRepository{})
	view, err := svc.GetCart(context.Background(), userID)
	if err != nil || view.TotalItems != 1 {
		t.Fatalf("unexpected GetCart result: %+v err=%v", view, err)
	}

	_, err = svc.RemoveItem(context.Background(), userID, bookID)
	if err != nil {
		t.Fatalf("unexpected RemoveItem error: %v", err)
	}
	if !removed {
		t.Fatalf("expected remove call")
	}

	err = svc.Clear(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected Clear error: %v", err)
	}
	if !cleared {
		t.Fatalf("expected clear call")
	}
}

func TestAddAndUpdateItemDelegateToUpsert(t *testing.T) {
	userID := uuid.New()
	bookID := uuid.New()
	cartID := uuid.New()
	calls := 0

	svc := NewCartService(
		&mockCartRepository{
			getOrCreateCartFunc: func(ctx context.Context, gotUserID uuid.UUID) (domain.Cart, error) {
				return domain.Cart{ID: cartID, UserID: gotUserID}, nil
			},
			upsertCartItemFunc: func(ctx context.Context, gotCartID uuid.UUID, gotBookID uuid.UUID, count int, unitPrice float64) error {
				calls++
				return nil
			},
			getCartItemsFunc: func(ctx context.Context, gotUserID uuid.UUID) ([]domain.CartItemDetail, error) {
				return []domain.CartItemDetail{{BookID: gotUserID, Count: 1, LineTotal: 10}}, nil
			},
		},
		&mockBookRepository{
			findByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
				return &domain.Book{ID: bookID, IsActive: true, AvailableStock: 100, Price: 10}, nil
			},
		},
	)

	_, err := svc.AddItem(context.Background(), userID, domain.CartItemInput{BookID: bookID, Count: 1})
	if err != nil {
		t.Fatalf("unexpected AddItem error: %v", err)
	}
	_, err = svc.UpdateItem(context.Background(), userID, domain.CartItemInput{BookID: bookID, Count: 1})
	if err != nil {
		t.Fatalf("unexpected UpdateItem error: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected two upsert calls, got %d", calls)
	}
}
