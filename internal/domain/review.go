package domain

import (
	"context"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var ErrReviewRequiresPurchase = errors.New("only customers who purchased this book can leave a review")

type Review struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440007"`
	BookID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"book_id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440003"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440004"`
	User      User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Book      Book       `gorm:"foreignKey:BookID" json:"book,omitempty"`
	Rating    int        `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating" example:"5"`
	Comment   string     `gorm:"type:text;default:''" json:"comment" example:"Thought-provoking and timeless."`
	CreatedAt time.Time  `json:"created_at" format:"date-time" example:"2026-04-06T10:30:00Z"`
	UpdatedAt time.Time  `json:"updated_at" format:"date-time" example:"2026-04-06T11:45:00Z"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" format:"date-time" example:"2026-04-07T09:00:00Z"`
} // @name Review

type ReviewInput struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5" example:"5"`
	Comment string `json:"comment" example:"Thought-provoking and timeless."`
} // @name ReviewInput

type ReviewSummary struct {
	AverageRating float64 `json:"average_rating" example:"4.7"`
	TotalReviews  int64   `json:"total_reviews" example:"128"`
} // @name ReviewSummary

type ReviewListResponse struct {
	Items   []Review      `json:"items"`
	Summary ReviewSummary `json:"summary"`
} // @name ReviewListResponse

type ReviewRepository interface {
	ListByBookID(ctx context.Context, bookID uuid.UUID) ([]Review, ReviewSummary, error)
	Upsert(ctx context.Context, bookID, userID uuid.UUID, input ReviewInput) (*Review, error)
}

type ReviewService interface {
	ListBookReviews(ctx context.Context, bookID uuid.UUID) (*ReviewListResponse, error)
	UpsertBookReview(ctx context.Context, bookID, userID uuid.UUID, input ReviewInput) (*Review, error)
}

type ReviewController interface {
	RegisterRoutes(r gin.IRouter)
}
