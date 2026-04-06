package domain

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Category defines model for Category
type Category struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440001"`
	Name      string     `gorm:"not null" json:"name" example:"Fiction"`
	CreatedAt time.Time  `json:"created_at" format:"date-time" example:"2026-04-06T10:30:00Z"`
	UpdatedAt time.Time  `json:"updated_at" format:"date-time" example:"2026-04-06T11:45:00Z"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" format:"date-time" example:"2026-04-07T09:00:00Z"`
} // @name Category

// CategoryInput defines input model for Category
type CategoryInput struct {
	Name string `json:"name" binding:"required,min=2" example:"Fiction"`
} // @name CategoryInput

type CategoryRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (Category, error)
	FindByName(ctx context.Context, name string) (Category, error)
	List(ctx context.Context, limit, offset int) ([]Category, error)
	Create(ctx context.Context, category *Category) error
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type CategoryService interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Category, error)
	List(ctx context.Context, limit, offset int) ([]Category, error)
	Create(ctx context.Context, input CategoryInput) (*Category, error)
	Update(ctx context.Context, id uuid.UUID, input CategoryInput) (*Category, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type CategoryController interface {
	RegisterRoutes(r gin.IRouter)
}
