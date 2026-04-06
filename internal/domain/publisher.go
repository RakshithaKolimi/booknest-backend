package domain

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Publisher defines model for Publisher
type Publisher struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440002"`
	LegalName   string    `gorm:"not null" json:"legal_name" example:"Penguin Random House LLC"`
	TradingName string    `gorm:"not null" json:"trading_name" example:"Penguin Random House"`
	Email       string    `gorm:"not null" json:"email" example:"contact@penguinrandomhouse.com"`
	Mobile      string    `gorm:"not null" json:"mobile" example:"+919876543210"`
	Address     string    `gorm:"type:text" json:"address" example:"1745 Broadway, New York, NY"`
	City        string    `gorm:"not null" json:"city" example:"New York"`
	State       string    `gorm:"not null" json:"state" example:"New York"`
	Country     string    `gorm:"not null" json:"country" example:"USA"`
	Zipcode     string    `gorm:"not null" json:"zipcode" example:"10019"`
	IsActive    bool      `gorm:"default:false" json:"is_active" example:"true"`
	BaseEntity
} // @name Publisher

// PublisherInput defines input model for Publisher
type PublisherInput struct {
	LegalName   string `json:"legal_name" binding:"required,min=3" example:"Penguin Random House LLC"`
	TradingName string `json:"trading_name" binding:"required,min=3" example:"Penguin Random House"`
	Email       string `json:"email" binding:"required,email" example:"contact@penguinrandomhouse.com"`
	Mobile      string `json:"mobile" binding:"required,e164" example:"+919876543210"`
	Address     string `json:"address" binding:"required,min=3" example:"1745 Broadway, New York, NY"`
	City        string `json:"city" binding:"required" example:"New York"`
	State       string `json:"state" binding:"required" example:"New York"`
	Country     string `json:"country" binding:"required" example:"USA"`
	Zipcode     string `json:"zipcode" binding:"required" example:"10019"`
} // @name PublisherInput

type PublisherRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (Publisher, error)
	List(ctx context.Context, limit, offset int, search string) ([]Publisher, error)
	Create(ctx context.Context, publisher *Publisher) error
	Update(ctx context.Context, publisher *Publisher) error
	SetActive(ctx context.Context, id uuid.UUID, active bool) error
	Delete(ctx context.Context, id uuid.UUID) error
}
type PublisherService interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Publisher, error)
	List(ctx context.Context, limit, offset int, search string) ([]Publisher, error)
	Create(ctx context.Context, input PublisherInput) (*Publisher, error)
	Update(ctx context.Context, id uuid.UUID, input PublisherInput) (*Publisher, error)
	SetActive(ctx context.Context, id uuid.UUID, active bool) error
	Delete(ctx context.Context, id uuid.UUID) error
}
type PublisherController interface {
	RegisterRoutes(r gin.IRouter)
}
