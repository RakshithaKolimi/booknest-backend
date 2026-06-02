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
	ID                 uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440003"`
	Name               string     `gorm:"not null" json:"name" example:"1984"`
	AuthorID           uuid.UUID  `gorm:"type:uuid;not null;index" json:"author_id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Author             Author     `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	AvailableStock     int        `gorm:"check:available_stock >= 0" json:"available_stock" example:"24"`
	ImageURL           *string    `json:"image_url,omitempty" example:"https://cdn.booknest.example/books/1984.jpg"`
	IsActive           bool       `gorm:"default:false" json:"is_active" example:"true"`
	Description        string     `gorm:"default:''" json:"description" example:"A dystopian novel set in a totalitarian society ruled by Big Brother."`
	Summary            string     `gorm:"default:''" json:"summary" example:"A gripping dystopian classic about surveillance and control, following Winston Smith as he questions the Party's reality. Orwell's chilling vision explores truth, freedom, and resistance under totalitarian rule."`
	ISBN               *string    `gorm:"uniqueIndex" json:"isbn,omitempty" example:"9780451524935"`
	Price              float64    `gorm:"type:numeric(10,2)" json:"price" example:"499.99"`
	DiscountPercentage float64    `gorm:"type:numeric(10,2);check:discount_percentage >= 0 AND discount_percentage <= 100" json:"discount_percentage" example:"10"`
	PublisherID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"publisher_id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440002"`
	Publisher          Publisher  `gorm:"foreignKey:PublisherID"`
	Categories         []Category `gorm:"many2many:book_categories;" json:"categories,omitempty"`
	AverageRating      float64    `gorm:"-" json:"average_rating" example:"4.7"`
	TotalReviews       int64      `gorm:"-" json:"total_reviews" example:"128"`
	CreatedAt          time.Time  `json:"created_at" format:"date-time" example:"2026-04-06T10:30:00Z"`
	UpdatedAt          time.Time  `json:"updated_at" format:"date-time" example:"2026-04-06T11:45:00Z"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty" format:"date-time" example:"2026-04-07T09:00:00Z"`
} // @name Book

// BookCategory defines model for BookCategory
type BookCategory struct {
	BookID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"book_id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440003"`
	CategoryID uuid.UUID `gorm:"type:uuid;primaryKey" json:"category_id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440001"`
	Book       Book      `gorm:"foreignKey:BookID"`
	Category   Category  `gorm:"foreignKey:CategoryID"`
	CreatedAt  time.Time `json:"created_at" format:"date-time" example:"2026-04-06T10:30:00Z"`
	UpdatedAt  time.Time `json:"updated_at" format:"date-time" example:"2026-04-06T11:45:00Z"`
	DeletedAt  time.Time `json:"deleted_at,omitempty" format:"date-time" example:"2026-04-07T09:00:00Z"`
} // @name BookCategory

// BookInput is used for create/update requests
type BookInput struct {
	Name               string      `json:"name" binding:"required" example:"1984"`
	AuthorName         string      `json:"author_name" binding:"required" example:"George Orwell"`
	AuthorID           *uuid.UUID  `json:"author_id,omitempty" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	AvailableStock     int         `json:"available_stock" example:"24"`
	ImageURL           *string     `json:"image_url,omitempty" example:"https://cdn.booknest.example/books/1984.jpg"`
	IsActive           bool        `json:"is_active" example:"true"`
	Description        string      `json:"description" example:"A dystopian novel set in a totalitarian society ruled by Big Brother."`
	ISBN               *string     `json:"isbn,omitempty" example:"9780451524935"`
	Price              float64     `json:"price" example:"499.99"`
	DiscountPercentage float64     `json:"discount_percentage" example:"10"`
	PublisherID        uuid.UUID   `json:"publisher_id" binding:"required" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440002"`
	CategoryIDs        []uuid.UUID `json:"category_ids,omitempty" example:"550e8400-e29b-41d4-a716-446655440001"`
} // @name BookInput

// BookFilter defines filter model for Book
type BookFilter struct {
	Search       *string     `json:"search,omitempty" form:"search" example:"orwell"` // name / author / isbn
	MinPrice     *float64    `json:"min_price,omitempty" form:"min_price" example:"100"`
	MaxPrice     *float64    `json:"max_price,omitempty" form:"max_price" example:"500"`
	IsActive     *bool       `json:"is_active,omitempty" form:"is_active" example:"true"`
	IDs          []uuid.UUID `json:"ids,omitempty" form:"ids" example:"550e8400-e29b-41d4-a716-446655440003"`
	AuthorIDs    []uuid.UUID `json:"author_ids,omitempty" form:"author_ids" example:"550e8400-e29b-41d4-a716-446655440000"`
	PublisherIDs []uuid.UUID `json:"publisher_ids,omitempty" form:"publisher_ids" example:"550e8400-e29b-41d4-a716-446655440002"`
	CategoryIDs  []uuid.UUID `json:"category_ids,omitempty" form:"category_ids" example:"550e8400-e29b-41d4-a716-446655440001"`
	MinStock     *int        `json:"min_stock,omitempty" form:"min_stock" example:"1"`
} // @name BookFilter

// BookSearchResult defines model for BookSearchResult
type BookSearchResult struct {
	Items      []Book  `json:"items"`
	Total      int64   `json:"total" example:"42"`
	Limit      uint64  `json:"limit" example:"10"`
	Offset     uint64  `json:"offset" example:"0"`
	NextCursor *string `json:"next_cursor,omitempty" example:"MjAyNi0wNC0wNlQxMTo0NTowMFo="`
	HasMore    bool    `json:"has_more" example:"true"`
} // @name BookSearchResult

type BookRepository interface {
	Create(ctx context.Context, book *Book) error
	CreateWithRelations(ctx context.Context, input BookInput) (*Book, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Book, error)
	ListBooksWithoutEmbeddings(ctx context.Context, limit, offset int) ([]Book, error)
	List(ctx context.Context, limit, offset int) ([]Book, error)
	FilterByCriteria(ctx context.Context, filter BookFilter, pagination QueryOptions) ([]Book, int64, error)
	QueryBooks(ctx context.Context, filter BookFilter, pagination QueryOptions) ([]Book, int64, *string, bool, error)
	Update(ctx context.Context, book *Book) error
	UpdateWithRelations(ctx context.Context, id uuid.UUID, input BookInput) (*Book, error)
	ReplaceCategories(ctx context.Context, bookID uuid.UUID, categoryIDs []uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type BookService interface {
	CreateBook(ctx context.Context, input BookInput) (*Book, error)
	GetBook(ctx context.Context, id uuid.UUID) (*Book, error)
	ListBooks(ctx context.Context, limit, offset int) ([]Book, error)
	FilterByCriteria(ctx context.Context, filter BookFilter, q QueryOptions) (*BookSearchResult, error)
	QueryBooks(ctx context.Context, filter BookFilter, q QueryOptions) (*BookSearchResult, error)
	UpdateBook(ctx context.Context, id uuid.UUID, input BookInput) (*Book, error)
	GenerateSummary(ctx context.Context, id uuid.UUID) (*Book, error)
	GenerateCategories(ctx context.Context, id uuid.UUID) (*Book, error)
	GenerateEmbeddings(ctx context.Context, id uuid.UUID) (*Book, error)
	DeleteBook(ctx context.Context, id uuid.UUID) error
	RecommendBooks(ctx context.Context, userID uuid.UUID, limit int) ([]Book, error)
}

type BookController interface {
	RegisterRoutes(r gin.IRouter)
}
