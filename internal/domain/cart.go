package domain

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Cart defines the model for Cart
type Cart struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
} // @name Cart

// CartItem defines model for cart item
type CartItem struct {
	CartID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"cart_id"`
	BookID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"book_id"`
	Count     int       `gorm:"check:count > 0" json:"count"`
	CartPrice float64   `gorm:"type:numeric(10,2)" json:"cart_price"`
	Book      Book      `gorm:"foreignKey:BookID"`
	Cart      Cart      `gorm:"foreignKey:CartID"`
	BaseEntity
} // @name CartItem

type CartItemInput struct {
	BookID uuid.UUID `json:"book_id" binding:"required"`
	Count  int       `json:"count" binding:"required,min=1"`
}

type CartItemDetail struct {
	BookID     uuid.UUID `json:"book_id"`
	Name       string    `json:"name"`
	AuthorName string    `json:"author_name"`
	ImageURL   *string   `json:"image_url,omitempty"`
	UnitPrice  float64   `json:"unit_price"`
	Count      int       `json:"count"`
	LineTotal  float64   `json:"line_total"`
}

type CartItemRecord struct {
	BookID         uuid.UUID
	Count          int
	UnitPrice      float64
	AvailableStock int
}

type CartView struct {
	CartID     uuid.UUID        `json:"cart_id"`
	UserID     uuid.UUID        `json:"user_id"`
	Items      []CartItemDetail `json:"items"`
	Subtotal   float64          `json:"subtotal"`
	TotalItems int              `json:"total_items"`
}

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
