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
	ID        uuid.UUID             `gorm:"type:uuid;primaryKey" db:"id" json:"id"`
	UserID    uuid.UUID             `gorm:"type:uuid;index" db:"user_id" json:"user_id"`
	Type      VerificationTokenType `gorm:"type:varchar(32);index" db:"type" json:"type"`
	TokenHash string                `gorm:"not null" db:"token_hash" json:"token_hash"`
	ExpiresAt time.Time             `gorm:"index" db:"expires_at" json:"expires_at"`
	IsUsed    bool                  `gorm:"default:false" db:"is_used" json:"is_used"`

	UsedAt   *time.Time        `db:"used_at" json:"used_at"`

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
