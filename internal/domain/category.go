package domain

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Category defines model for Category
type Category struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string     `gorm:"not null" json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// CategoryInput defines input model for Category
type CategoryInput struct {
	Name string `json:"name" binding:"required,min=2"`
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
