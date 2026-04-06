package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserPreferences stores notification preferences for a user.
type UserPreferences struct {
	UserID uuid.UUID `gorm:"type:uuid;primaryKey" db:"user_id" json:"user_id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440004"`
	UseSMS bool      `gorm:"column:use_sms;not null;default:false" db:"use_sms" json:"use_sms" example:"true"`
	BaseEntity
}

type UserProfile struct {
	ID             uuid.UUID `json:"id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440004"`
	FirstName      string    `json:"first_name" example:"Rakshitha"`
	LastName       string    `json:"last_name" example:"Koli"`
	Email          string    `json:"email" example:"rakshitha@example.com"`
	Mobile         string    `json:"mobile" example:"+919876543210"`
	Role           UserRole  `json:"role" enums:"USER,ADMIN" example:"USER"`
	EmailVerified  bool      `json:"email_verified" example:"true"`
	MobileVerified bool      `json:"mobile_verified" example:"true"`
	UseSMS         bool      `json:"use_sms" example:"true"`
	CreatedAt      time.Time `json:"created_at" format:"date-time" example:"2026-04-06T10:30:00Z"`
}

type UserPreferencesInput struct {
	UseSMS bool `json:"use_sms" example:"true"`
}
