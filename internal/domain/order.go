package domain

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Order defines the model for Order
type Order struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	OrderNumber   string         `gorm:"uniqueIndex;not null" json:"order_number"`
	TotalPrice    float64        `gorm:"type:numeric(10,2)" json:"total_price"`
	UserID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	User          User           `gorm:"foreignKey:UserID"`
	PaymentMethod *PaymentMethod `gorm:"type:payment_method" json:"payment_method,omitempty"`
	PaymentStatus *PaymentStatus `gorm:"type:payment_status" json:"payment_status,omitempty"`
	Status        OrderStatus    `gorm:"type:order_status;default:PENDING" json:"status"`
	BaseEntity
} // @name Order

// OrderItem defines model for OrderItem
type OrderItem struct {
	OrderID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"order_id"`
	BookID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"book_id"`
	PurchaseCount int       `gorm:"check:purchase_count > 0" json:"purchase_count"`
	PurchasePrice float64   `gorm:"type:numeric(10,2)" json:"purchase_price"`
	TotalPrice    float64   `gorm:"type:numeric(10,2)" json:"total_price"`
	Book          Book      `gorm:"foreignKey:BookID"`
	Order         Order     `gorm:"foreignKey:OrderID"`
	BaseEntity
} // @name OrderItem

type OrderItemDetail struct {
	BookID     uuid.UUID `json:"book_id"`
	Name       string    `json:"name"`
	ImageURL   *string   `json:"image_url,omitempty"`
	UnitPrice  float64   `json:"unit_price"`
	Count      int       `json:"count"`
	LineTotal  float64   `json:"line_total"`
}

type OrderView struct {
	Order Order             `json:"order"`
	Items []OrderItemDetail `json:"items"`
}

type CheckoutInput struct {
	PaymentMethod PaymentMethod `json:"payment_method" binding:"required"`
}

type PaymentConfirmInput struct {
	OrderID uuid.UUID `json:"order_id" binding:"required"`
	Success bool      `json:"success"`
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *Order) error
	CreateOrderItems(ctx context.Context, items []OrderItem) error
	ListOrdersByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]OrderView, error)
	ListOrders(ctx context.Context, limit, offset int) ([]OrderView, error)
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (Order, error)
	GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]OrderItemDetail, error)
	UpdateOrderPayment(ctx context.Context, orderID uuid.UUID, status PaymentStatus, method PaymentMethod) error
	UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status OrderStatus) error
	DecrementStock(ctx context.Context, items []OrderItem) error
}

type OrderService interface {
	Checkout(ctx context.Context, userID uuid.UUID, input CheckoutInput) (OrderView, error)
	ConfirmPayment(ctx context.Context, userID uuid.UUID, input PaymentConfirmInput) (OrderView, error)
	ListUserOrders(ctx context.Context, userID uuid.UUID, limit, offset int) ([]OrderView, error)
	ListAllOrders(ctx context.Context, limit, offset int) ([]OrderView, error)
}

type OrderController interface {
	RegisterRoutes(r gin.IRouter)
}
