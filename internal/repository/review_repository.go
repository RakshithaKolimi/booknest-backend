package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"booknest/internal/domain"
)

type reviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) domain.ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) ListByBookID(
	ctx context.Context,
	bookID uuid.UUID,
) ([]domain.Review, domain.ReviewSummary, error) {
	var reviews []domain.Review
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("book_id = ?", bookID).
		Order("created_at DESC").
		Find(&reviews).Error
	if err != nil {
		return nil, domain.ReviewSummary{}, err
	}

	var summary domain.ReviewSummary
	type reviewAggregate struct {
		AverageRating *float64
		TotalReviews  int64
	}

	var aggregate reviewAggregate
	err = r.db.WithContext(ctx).
		Model(&domain.Review{}).
		Select("AVG(rating) AS average_rating, COUNT(*) AS total_reviews").
		Where("book_id = ?", bookID).
		Scan(&aggregate).Error
	if err != nil {
		return nil, domain.ReviewSummary{}, err
	}

	if aggregate.AverageRating != nil {
		summary.AverageRating = *aggregate.AverageRating
	}
	summary.TotalReviews = aggregate.TotalReviews

	return reviews, summary, nil
}

func (r *reviewRepository) Upsert(
	ctx context.Context,
	bookID, userID uuid.UUID,
	input domain.ReviewInput,
) (*domain.Review, error) {
	var review domain.Review

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&domain.Book{}, "id = ?", bookID).Error; err != nil {
			return err
		}

		err := tx.Where("book_id = ? AND user_id = ?", bookID, userID).
			First(&review).Error
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			}

			review = domain.Review{
				ID:      uuid.New(),
				BookID:  bookID,
				UserID:  userID,
				Rating:  input.Rating,
				Comment: input.Comment,
			}
			return tx.Create(&review).Error
		}

		review.Rating = input.Rating
		review.Comment = input.Comment
		return tx.Save(&review).Error
	})
	if err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).
		Preload("User").
		First(&review, "id = ?", review.ID).Error; err != nil {
		return nil, err
	}

	return &review, nil
}
