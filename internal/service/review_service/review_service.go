package review_service

import (
	"context"

	"github.com/google/uuid"

	"booknest/internal/domain"
)

type reviewService struct {
	repo      domain.ReviewRepository
	orderRepo domain.OrderRepository
}

func NewReviewService(
	repo domain.ReviewRepository,
	orderRepo domain.OrderRepository,
) domain.ReviewService {
	return &reviewService{repo: repo, orderRepo: orderRepo}
}

func (s *reviewService) ListBookReviews(
	ctx context.Context,
	bookID uuid.UUID,
) (*domain.ReviewListResponse, error) {
	reviews, summary, err := s.repo.ListByBookID(ctx, bookID)
	if err != nil {
		return nil, err
	}

	return &domain.ReviewListResponse{
		Items:   reviews,
		Summary: summary,
	}, nil
}

func (s *reviewService) UpsertBookReview(
	ctx context.Context,
	bookID, userID uuid.UUID,
	input domain.ReviewInput,
) (*domain.Review, error) {
	purchased, err := s.orderRepo.HasUserPurchasedBook(ctx, userID, bookID)
	if err != nil {
		return nil, err
	}
	if !purchased {
		return nil, domain.ErrReviewRequiresPurchase
	}

	return s.repo.Upsert(ctx, bookID, userID, input)
}
