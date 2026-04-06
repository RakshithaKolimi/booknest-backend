package domain

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Author defines model for Author
type Author struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name string    `gorm:"not null;uniqueIndex" json:"name" example:"George Orwell"`
	BaseEntity
} // @name Author

// AuthorInput defines input model for Author
type AuthorInput struct {
	Name string `json:"name" binding:"required,min=2" example:"George Orwell"`
} // @name AuthorInput

type AuthorRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (Author, error)
	FindByName(ctx context.Context, name string) (Author, error)
	List(ctx context.Context, limit, offset int, search string) ([]Author, error)
	Create(ctx context.Context, author *Author) error
	Update(ctx context.Context, author *Author) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type AuthorService interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Author, error)
	List(ctx context.Context, limit, offset int, search string) ([]Author, error)
	Create(ctx context.Context, input AuthorInput) (*Author, error)
	Update(ctx context.Context, id uuid.UUID, input AuthorInput) (*Author, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type AuthorController interface {
	RegisterRoutes(r gin.IRouter)
}
