package domain

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Cart defines the model for Cart
type Cart struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440006"`
	UserID    uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440004"`
	User      User      `gorm:"foreignKey:UserID"`
	CreatedAt time.Time `json:"created_at" format:"date-time" example:"2026-04-06T10:30:00Z"`
	UpdatedAt time.Time `json:"updated_at" format:"date-time" example:"2026-04-06T11:45:00Z"`
} // @name Cart

// CartItem defines model for cart item
type CartItem struct {
	CartID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"cart_id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440006"`
	BookID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"book_id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440003"`
	Count     int       `gorm:"check:count > 0" json:"count" example:"2"`
	CartPrice float64   `gorm:"type:numeric(10,2)" json:"cart_price" example:"449.99"`
	Book      Book      `gorm:"foreignKey:BookID"`
	Cart      Cart      `gorm:"foreignKey:CartID"`
	BaseEntity
} // @name CartItem

// CartItemInput defines input model for CartItem
type CartItemInput struct {
	BookID uuid.UUID `json:"book_id" binding:"required" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440003"`
	Count  int       `json:"count" binding:"required,min=1" example:"2"`
} // @name CartItemInput

// CartItemDetail defines model for CartItemDetail
type CartItemDetail struct {
	BookID     uuid.UUID `json:"book_id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440003"`
	Name       string    `json:"name" example:"1984"`
	AuthorName string    `json:"author_name" example:"George Orwell"`
	ImageURL   *string   `json:"image_url,omitempty" example:"https://cdn.booknest.example/books/1984.jpg"`
	UnitPrice  float64   `json:"unit_price" example:"449.99"`
	Count      int       `json:"count" example:"2"`
	LineTotal  float64   `json:"line_total" example:"899.98"`
} // @name CartItemDetail

// CartItemRecord defines model for CartItemRecord
type CartItemRecord struct {
	BookID         uuid.UUID `json:"book_id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440003"`
	Count          int       `json:"count" example:"2"`
	UnitPrice      float64   `json:"unit_price" example:"449.99"`
	AvailableStock int       `json:"available_stock" example:"24"`
} // @name CartItemRecord

// CartView defines model for CartView
type CartView struct {
	CartID     uuid.UUID        `json:"cart_id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440006"`
	UserID     uuid.UUID        `json:"user_id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440004"`
	Items      []CartItemDetail `json:"items"`
	Subtotal   float64          `json:"subtotal" example:"899.98"`
	TotalItems int              `json:"total_items" example:"2"`
} // @name CartView

type CartRepository interface {
	GetOrCreateCart(ctx context.Context, userID uuid.UUID) (Cart, error)
	GetCartItems(ctx context.Context, userID uuid.UUID) ([]CartItemDetail, error)
	GetCartItemRecords(ctx context.Context, userID uuid.UUID) ([]CartItemRecord, error)
	UpsertCartItem(ctx context.Context, cartID uuid.UUID, bookID uuid.UUID, count int, unitPrice float64) error
	RemoveCartItem(ctx context.Context, cartID uuid.UUID, bookID uuid.UUID) error
	ClearCart(ctx context.Context, cartID uuid.UUID) error
}

type CartService interface {
	GetCart(ctx context.Context, userID uuid.UUID) (CartView, error)
	AddItem(ctx context.Context, userID uuid.UUID, input CartItemInput) (CartView, error)
	UpdateItem(ctx context.Context, userID uuid.UUID, input CartItemInput) (CartView, error)
	RemoveItem(ctx context.Context, userID uuid.UUID, bookID uuid.UUID) (CartView, error)
	Clear(ctx context.Context, userID uuid.UUID) error
}

type CartController interface {
	RegisterRoutes(r gin.IRouter)
}
