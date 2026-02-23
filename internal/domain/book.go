package domain

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	SortByCreatedAt = "created_at"
	SortByPrice     = "price"
	SortByName      = "name"
	SortByStock     = "available_stock"
)

// Book defines model for Book
type Book struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Name               string     `gorm:"not null" json:"name"`
	AuthorID           uuid.UUID  `gorm:"type:uuid;not null;index" json:"author_id"`
	Author             Author     `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	AvailableStock     int        `gorm:"check:available_stock >= 0" json:"available_stock"`
	ImageURL           *string    `json:"image_url,omitempty"`
	IsActive           bool       `gorm:"default:false" json:"is_active"`
	Description        string     `gorm:"default:''" json:"description"`
	ISBN               *string    `gorm:"uniqueIndex" json:"isbn,omitempty"`
	Price              float64    `gorm:"type:numeric(10,2)" json:"price"`
	DiscountPercentage float64    `gorm:"type:numeric(10,2);check:discount_percentage >= 0 AND discount_percentage <= 100" json:"discount_percentage"`
	PublisherID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"publisher_id"`
	Publisher          Publisher  `gorm:"foreignKey:PublisherID"`
	Categories         []Category `gorm:"many2many:book_categories;" json:"categories,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
} // @name Book

// BookCategory defines model for BookCategory
type BookCategory struct {
	BookID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"book_id"`
	CategoryID uuid.UUID `gorm:"type:uuid;primaryKey" json:"category_id"`
	Book       Book      `gorm:"foreignKey:BookID"`
	Category   Category  `gorm:"foreignKey:CategoryID"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	DeletedAt  time.Time `json:"deleted_at,omitempty"`
} // @name BookCategory

// BookInput is used for create/update requests
type BookInput struct {
	Name               string      `json:"name" binding:"required"`
	AuthorName         string      `json:"author_name" binding:"required"`
	AuthorID           *uuid.UUID  `json:"author_id,omitempty"`
	AvailableStock     int         `json:"available_stock"`
	ImageURL           *string     `json:"image_url,omitempty"`
	IsActive           bool        `json:"is_active"`
	Description        string      `json:"description"`
	ISBN               *string     `json:"isbn,omitempty"`
	Price              float64     `json:"price"`
	DiscountPercentage float64     `json:"discount_percentage"`
	PublisherID        uuid.UUID   `json:"publisher_id" binding:"required"`
	CategoryIDs        []uuid.UUID `json:"category_ids,omitempty"`
}

type BookFilter struct {
	Search       *string // name / author / isbn
	MinPrice     *float64
	MaxPrice     *float64
	IsActive     *bool
	IDs          []uuid.UUID
	AuthorIDs    []uuid.UUID
	PublisherIDs []uuid.UUID
	CategoryIDs  []uuid.UUID
	MinStock     *int
}

type BookSearchResult struct {
	Items  []Book `json:"items"`
	Total  int64  `json:"total"`
	Limit  uint64 `json:"limit"`
	Offset uint64 `json:"offset"`
}

type BookRepository interface {
	Create(ctx context.Context, book *Book) error
	CreateWithRelations(ctx context.Context, input BookInput) (*Book, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Book, error)
	List(ctx context.Context, limit, offset int) ([]Book, error)
	FilterByCriteria(ctx context.Context, filter BookFilter, pagination QueryOptions) ([]Book, int64, error)
	Update(ctx context.Context, book *Book) error
	UpdateWithRelations(ctx context.Context, id uuid.UUID, input BookInput) (*Book, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type BookService interface {
	CreateBook(ctx context.Context, input BookInput) (*Book, error)
	GetBook(ctx context.Context, id uuid.UUID) (*Book, error)
	ListBooks(ctx context.Context, limit, offset int) ([]Book, error)
	FilterByCriteria(ctx context.Context, filter BookFilter, q QueryOptions) (*BookSearchResult, error)
	UpdateBook(ctx context.Context, id uuid.UUID, input BookInput) (*Book, error)
	DeleteBook(ctx context.Context, id uuid.UUID) error
}

type BookController interface {
	RegisterRoutes(r *gin.Engine)
}
