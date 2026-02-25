package domain

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Publisher defines model for Publisher
type Publisher struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	LegalName   string    `gorm:"not null" json:"legal_name"`
	TradingName string    `gorm:"not null" json:"trading_name"`
	Email       string    `gorm:"not null" json:"email"`
	Mobile      string    `gorm:"not null" json:"mobile"`
	Address     string    `gorm:"type:text" json:"address"`
	City        string    `gorm:"not null" json:"city"`
	State       string    `gorm:"not null" json:"state"`
	Country     string    `gorm:"not null" json:"country"`
	Zipcode     string    `gorm:"not null" json:"zipcode"`
	IsActive    bool      `gorm:"default:false" json:"is_active"`
	BaseEntity
} // @name Publisher

// PublisherInput defines input model for Publisher
type PublisherInput struct {
	LegalName   string `json:"legal_name" binding:"required,min=3"`
	TradingName string `json:"trading_name" binding:"required,min=3"`
	Email       string `json:"email" binding:"required,email"`
	Mobile      string `json:"mobile" binding:"required,e164"`
	Address     string `json:"address" binding:"required,min=3"`
	City        string `json:"city" binding:"required"`
	State       string `json:"state" binding:"required"`
	Country     string `json:"country" binding:"required"`
	Zipcode     string `json:"zipcode" binding:"required"`
} // @name PublisherInput

type PublisherRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (Publisher, error)
	List(ctx context.Context, limit, offset int) ([]Publisher, error)
	Create(ctx context.Context, publisher *Publisher) error
	Update(ctx context.Context, publisher *Publisher) error
	SetActive(ctx context.Context, id uuid.UUID, active bool) error
	Delete(ctx context.Context, id uuid.UUID) error
}
type PublisherService interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Publisher, error)
	List(ctx context.Context, limit, offset int) ([]Publisher, error)
	Create(ctx context.Context, input PublisherInput) (*Publisher, error)
	Update(ctx context.Context, id uuid.UUID, input PublisherInput) (*Publisher, error)
	SetActive(ctx context.Context, id uuid.UUID, active bool) error
	Delete(ctx context.Context, id uuid.UUID) error
}
type PublisherController interface {
	RegisterRoutes(r gin.IRouter)
}
