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
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	BookID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"book_id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	User      User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Book      Book       `gorm:"foreignKey:BookID" json:"book,omitempty"`
	Rating    int        `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	Comment   string     `gorm:"type:text;default:''" json:"comment"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
} // @name Review

type ReviewInput struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment"`
} // @name ReviewInput

type ReviewSummary struct {
	AverageRating float64 `json:"average_rating"`
	TotalReviews  int64   `json:"total_reviews"`
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
