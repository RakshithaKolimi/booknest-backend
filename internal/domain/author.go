package domain

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Author defines model for Author
type Author struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name string    `gorm:"not null;uniqueIndex" json:"name"`
	BaseEntity
} // @name Author

// AuthorInput defines input model for Author
type AuthorInput struct {
	Name string `json:"name" binding:"required,min=2"`
} // @name AuthorInput

type AuthorRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (Author, error)
	FindByName(ctx context.Context, name string) (Author, error)
	List(ctx context.Context, limit, offset int) ([]Author, error)
	Create(ctx context.Context, author *Author) error
	Update(ctx context.Context, author *Author) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type AuthorService interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Author, error)
	List(ctx context.Context, limit, offset int) ([]Author, error)
	Create(ctx context.Context, input AuthorInput) (*Author, error)
	Update(ctx context.Context, id uuid.UUID, input AuthorInput) (*Author, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type AuthorController interface {
	RegisterRoutes(r gin.IRouter)
}
