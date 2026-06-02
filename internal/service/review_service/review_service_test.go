package review_service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"booknest/internal/domain"
)

type stubReviewRepo struct {
	listFunc   func(ctx context.Context, bookID uuid.UUID) ([]domain.Review, domain.ReviewSummary, error)
	upsertFunc func(ctx context.Context, bookID, userID uuid.UUID, input domain.ReviewInput) (*domain.Review, error)
}

func (s *stubReviewRepo) ListByBookID(ctx context.Context, bookID uuid.UUID) ([]domain.Review, domain.ReviewSummary, error) {
	return s.listFunc(ctx, bookID)
}

func (s *stubReviewRepo) Upsert(ctx context.Context, bookID, userID uuid.UUID, input domain.ReviewInput) (*domain.Review, error) {
	return s.upsertFunc(ctx, bookID, userID, input)
}

type stubOrderRepo struct {
	hasPurchasedFunc func(ctx context.Context, userID, bookID uuid.UUID) (bool, error)
}

func (s *stubOrderRepo) CreateOrder(ctx context.Context, order *domain.Order) error { return nil }
func (s *stubOrderRepo) CreateOrderItems(ctx context.Context, items []domain.OrderItem) error {
	return nil
}
func (s *stubOrderRepo) ListOrdersByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.OrderView, error) {
	return nil, nil
}
func (s *stubOrderRepo) ListOrders(ctx context.Context, limit, offset int) ([]domain.OrderView, error) {
	return nil, nil
}
func (s *stubOrderRepo) HasUserPurchasedBook(ctx context.Context, userID, bookID uuid.UUID) (bool, error) {
	return s.hasPurchasedFunc(ctx, userID, bookID)
}
func (s *stubOrderRepo) GetOrderByID(ctx context.Context, orderID uuid.UUID) (domain.Order, error) {
	return domain.Order{}, nil
}
func (s *stubOrderRepo) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItemDetail, error) {
	return nil, nil
}
func (s *stubOrderRepo) UpdateOrderPayment(ctx context.Context, orderID uuid.UUID, status domain.PaymentStatus, method domain.PaymentMethod) error {
	return nil
}
func (s *stubOrderRepo) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status domain.OrderStatus, cancellationReason *string) error {
	return nil
}
func (s *stubOrderRepo) DecrementStock(ctx context.Context, items []domain.OrderItem) error {
	return nil
}
func (s *stubOrderRepo) GetPurchasedBookIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

func TestListBookReviews(t *testing.T) {
	bookID := uuid.New()
	service := NewReviewService(&stubReviewRepo{
		listFunc: func(ctx context.Context, gotBookID uuid.UUID) ([]domain.Review, domain.ReviewSummary, error) {
			if gotBookID != bookID {
				t.Fatalf("expected %s, got %s", bookID, gotBookID)
			}
			return []domain.Review{{ID: uuid.New(), BookID: bookID, Rating: 5}}, domain.ReviewSummary{
				AverageRating: 5,
				TotalReviews:  1,
			}, nil
		},
		upsertFunc: func(ctx context.Context, bookID, userID uuid.UUID, input domain.ReviewInput) (*domain.Review, error) {
			return nil, nil
		},
	}, &stubOrderRepo{
		hasPurchasedFunc: func(ctx context.Context, userID, bookID uuid.UUID) (bool, error) {
			return true, nil
		},
	})

	result, err := service.ListBookReviews(context.Background(), bookID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary.TotalReviews != 1 {
		t.Fatalf("expected 1 review, got %d", result.Summary.TotalReviews)
	}
}

func TestUpsertBookReview(t *testing.T) {
	bookID := uuid.New()
	userID := uuid.New()
	service := NewReviewService(&stubReviewRepo{
		listFunc: func(ctx context.Context, bookID uuid.UUID) ([]domain.Review, domain.ReviewSummary, error) {
			return nil, domain.ReviewSummary{}, nil
		},
		upsertFunc: func(ctx context.Context, gotBookID, gotUserID uuid.UUID, input domain.ReviewInput) (*domain.Review, error) {
			if gotBookID != bookID || gotUserID != userID {
				t.Fatalf("unexpected ids")
			}
			if input.Rating != 4 || input.Comment != "Solid read" {
				t.Fatalf("unexpected input: %+v", input)
			}
			return &domain.Review{
				ID:      uuid.New(),
				BookID:  bookID,
				UserID:  userID,
				Rating:  input.Rating,
				Comment: input.Comment,
			}, nil
		},
	}, &stubOrderRepo{
		hasPurchasedFunc: func(ctx context.Context, gotUserID, gotBookID uuid.UUID) (bool, error) {
			if gotBookID != bookID || gotUserID != userID {
				t.Fatalf("unexpected ids for purchase check")
			}
			return true, nil
		},
	})

	result, err := service.UpsertBookReview(context.Background(), bookID, userID, domain.ReviewInput{
		Rating:  4,
		Comment: "Solid read",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Rating != 4 {
		t.Fatalf("expected rating 4, got %d", result.Rating)
	}
}

func TestUpsertBookReviewRequiresPurchase(t *testing.T) {
	bookID := uuid.New()
	userID := uuid.New()
	service := NewReviewService(&stubReviewRepo{
		listFunc: func(ctx context.Context, bookID uuid.UUID) ([]domain.Review, domain.ReviewSummary, error) {
			return nil, domain.ReviewSummary{}, nil
		},
		upsertFunc: func(ctx context.Context, bookID, userID uuid.UUID, input domain.ReviewInput) (*domain.Review, error) {
			t.Fatalf("upsert should not be called when purchase check fails")
			return nil, nil
		},
	}, &stubOrderRepo{
		hasPurchasedFunc: func(ctx context.Context, gotUserID, gotBookID uuid.UUID) (bool, error) {
			if gotBookID != bookID || gotUserID != userID {
				t.Fatalf("unexpected ids for purchase check")
			}
			return false, nil
		},
	})

	_, err := service.UpsertBookReview(context.Background(), bookID, userID, domain.ReviewInput{
		Rating:  4,
		Comment: "Solid read",
	})
	if err != domain.ErrReviewRequiresPurchase {
		t.Fatalf("expected purchase requirement error, got %v", err)
	}
}
