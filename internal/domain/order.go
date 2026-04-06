package domain

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Order defines the model for Order
type Order struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440005"`
	OrderNumber        string         `gorm:"uniqueIndex;not null" json:"order_number" example:"ORD-20260406-0001"`
	TotalPrice         float64        `gorm:"type:numeric(10,2)" json:"total_price" example:"899.98"`
	UserID             uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440004"`
	User               User           `gorm:"foreignKey:UserID"`
	PaymentMethod      *PaymentMethod `gorm:"type:payment_method" json:"payment_method,omitempty" enums:"COD,CREDIT_CARD,DEBIT_CARD,NET_BANKING,UPI" example:"UPI"`
	PaymentStatus      *PaymentStatus `gorm:"type:payment_status" json:"payment_status,omitempty" enums:"PENDING,PAID,REFUND_INITIATED,REFUNDED,FAILED" example:"PAID"`
	Status             OrderStatus    `gorm:"type:order_status;default:PENDING" json:"status" enums:"PENDING,FAILED,CANCELLED,COMPLETED" example:"COMPLETED"`
	CancellationReason *string        `gorm:"type:text" json:"cancellation_reason,omitempty" example:"Customer requested cancellation before dispatch."`
	BaseEntity
} // @name Order

// OrderItem defines model for OrderItem
type OrderItem struct {
	OrderID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"order_id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440005"`
	BookID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"book_id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440003"`
	PurchaseCount int       `gorm:"check:purchase_count > 0" json:"purchase_count" example:"2"`
	PurchasePrice float64   `gorm:"type:numeric(10,2)" json:"purchase_price" example:"449.99"`
	TotalPrice    float64   `gorm:"type:numeric(10,2)" json:"total_price" example:"899.98"`
	Book          Book      `gorm:"foreignKey:BookID"`
	Order         Order     `gorm:"foreignKey:OrderID"`
	BaseEntity
} // @name OrderItem

// OrderItemDetail defines model for OrderItemDetail
type OrderItemDetail struct {
	BookID    uuid.UUID `json:"book_id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440003"`
	Name      string    `json:"name" example:"1984"`
	ImageURL  *string   `json:"image_url,omitempty" example:"https://cdn.booknest.example/books/1984.jpg"`
	UnitPrice float64   `json:"unit_price" example:"449.99"`
	Count     int       `json:"count" example:"2"`
	LineTotal float64   `json:"line_total" example:"899.98"`
} // @name OrderItemDetail

// OrderView defines model for OrderView
type OrderView struct {
	Order Order             `json:"order"`
	Items []OrderItemDetail `json:"items"`
} // @name OrderView

// CheckoutInput defines input model for Checkout
type CheckoutInput struct {
	PaymentMethod PaymentMethod `json:"payment_method" binding:"required" enums:"COD,CREDIT_CARD,DEBIT_CARD,NET_BANKING,UPI" example:"UPI"`
} // @name CheckoutInput

// PaymentConfirmInput defines input model for PaymentConfirm
type PaymentConfirmInput struct {
	OrderID uuid.UUID `json:"order_id" binding:"required" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440005"`
	Success bool      `json:"success" example:"true"`
} // @name PaymentConfirmInput

// OrderCancelInput defines input model for user order cancellation
type OrderCancelInput struct {
	OrderID            uuid.UUID `json:"order_id" binding:"required" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440005"`
	CancellationReason string    `json:"cancellation_reason" binding:"required" example:"Customer requested cancellation before dispatch."`
} // @name OrderCancelInput

// AdminOrderStatusUpdateInput defines input model for admin order status updates
type AdminOrderStatusUpdateInput struct {
	OrderID            uuid.UUID      `json:"order_id" binding:"required" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440005"`
	Status             OrderStatus    `json:"status,omitempty" enums:"PENDING,FAILED,CANCELLED,COMPLETED" example:"COMPLETED"`
	PaymentStatus      *PaymentStatus `json:"payment_status,omitempty" enums:"PENDING,PAID,REFUND_INITIATED,REFUNDED,FAILED" example:"PAID"`
	CancellationReason string         `json:"cancellation_reason,omitempty" example:"Out of stock after payment failure."`
} // @name AdminOrderStatusUpdateInput

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *Order) error
	CreateOrderItems(ctx context.Context, items []OrderItem) error
	ListOrdersByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]OrderView, error)
	ListOrders(ctx context.Context, limit, offset int) ([]OrderView, error)
	HasUserPurchasedBook(ctx context.Context, userID, bookID uuid.UUID) (bool, error)
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (Order, error)
	GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]OrderItemDetail, error)
	UpdateOrderPayment(ctx context.Context, orderID uuid.UUID, status PaymentStatus, method PaymentMethod) error
	UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status OrderStatus, cancellationReason *string) error
	DecrementStock(ctx context.Context, items []OrderItem) error
}

type OrderService interface {
	Checkout(ctx context.Context, userID uuid.UUID, input CheckoutInput) (OrderView, error)
	ConfirmPayment(ctx context.Context, userID uuid.UUID, input PaymentConfirmInput) (OrderView, error)
	CancelOrder(ctx context.Context, userID uuid.UUID, input OrderCancelInput) (OrderView, error)
	AdminUpdateOrderStatus(ctx context.Context, input AdminOrderStatusUpdateInput) (OrderView, error)
	ListUserOrders(ctx context.Context, userID uuid.UUID, limit, offset int) ([]OrderView, error)
	ListAllOrders(ctx context.Context, limit, offset int) ([]OrderView, error)
}

type OrderController interface {
	RegisterRoutes(r gin.IRouter)
}
