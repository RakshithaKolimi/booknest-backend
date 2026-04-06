package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type VerificationTokenType string

const (
	VerificationEmail  VerificationTokenType = "EMAIL_VERIFICATION"
	VerificationMobile VerificationTokenType = "MOBILE_VERIFICATION"
	PasswordReset      VerificationTokenType = "PASSWORD_RESET"
	LoginOTP           VerificationTokenType = "LOGIN_OTP"
	RefreshToken       VerificationTokenType = "REFRESH_TOKEN"
)

// VerificationToken defines model for VerificationToken
type VerificationToken struct {
	ID        uuid.UUID             `gorm:"type:uuid;primaryKey" db:"id" json:"id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440009"`
	UserID    uuid.UUID             `gorm:"type:uuid;index" db:"user_id" json:"user_id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440004"`
	Type      VerificationTokenType `gorm:"type:varchar(32);index" db:"type" json:"type" enums:"EMAIL_VERIFICATION,MOBILE_VERIFICATION,PASSWORD_RESET,LOGIN_OTP,REFRESH_TOKEN" example:"EMAIL_VERIFICATION"`
	TokenHash string                `gorm:"not null" db:"token_hash" json:"token_hash" example:"4f3c2e1d9b8a7c6e5d4f3a2b1c0d9e8f"`
	ExpiresAt time.Time             `gorm:"index" db:"expires_at" json:"expires_at" format:"date-time" example:"2026-04-06T12:30:00Z"`
	IsUsed    bool                  `gorm:"default:false" db:"is_used" json:"is_used" example:"false"`

	UsedAt *time.Time `db:"used_at" json:"used_at" format:"date-time" example:"2026-04-06T11:45:00Z"`

	BaseEntity
} // @name VerificationToken

type VerificationTokenRepository interface {
	FindByUserIDAndType(
		ctx context.Context,
		userID uuid.UUID,
		tokenType VerificationTokenType,
	) (*VerificationToken, error)
	FindByHashAndType(ctx context.Context, tokenHash string, tokenType VerificationTokenType) (*VerificationToken, error)

	InvalidateByUserAndType(ctx context.Context, userID uuid.UUID, tokenType VerificationTokenType) error

	Create(ctx context.Context, token *VerificationToken) error
	Update(ctx context.Context, token *VerificationToken) error
	Delete(ctx context.Context, id uuid.UUID) error
}
