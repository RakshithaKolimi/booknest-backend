package domain

import (
	"context"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440004"`
	FirstName      string     `gorm:"not null" json:"first_name" example:"Rakshitha"`
	LastName       string     `gorm:"not null" json:"last_name" example:"Koli"`
	Email          string     `gorm:"uniqueIndex;not null" json:"email" example:"rakshitha@example.com"`
	Mobile         string     `gorm:"uniqueIndex;not null" json:"mobile" example:"+919876543210"`
	Password       string     `gorm:"not null" json:"-"`
	LastLogin      *time.Time `json:"last_login,omitempty" format:"date-time" example:"2026-04-06T11:45:00Z"`
	Role           UserRole   `gorm:"type:user_role;default:USER" json:"role" enums:"USER,ADMIN" example:"USER"`
	IsActive       bool       `gorm:"default:false" json:"is_active" example:"true"`
	EmailVerified  bool       `gorm:"default:false" json:"email_verified" example:"true"`
	MobileVerified bool       `gorm:"default:false" json:"mobile_verified" example:"true"`
	BaseEntity
} // @name User

// UserInput is used for creating or updating users
type UserInput struct {
	FirstName string   `json:"first_name" binding:"required,min=3" example:"Rakshitha"`
	LastName  string   `json:"last_name" binding:"required" example:"Koli"`
	Email     string   `json:"email" binding:"required,email" example:"rakshitha@example.com"`
	Mobile    string   `json:"mobile" binding:"required,e164" example:"+919876543210"`
	Password  string   `json:"password" binding:"required,min=6" example:"Password@123"`
	Role      UserRole `json:"role" binding:"required" enums:"USER,ADMIN" example:"USER"`
} // @name UserInput

type AdminRegistrationInput struct {
	FirstName string `json:"first_name" binding:"required,min=3" example:"Admin"`
	LastName  string `json:"last_name" binding:"required" example:"User"`
	Email     string `json:"email" binding:"required,email" example:"admin@booknest.com"`
	Mobile    string `json:"mobile" binding:"required,e164" example:"+919876543211"`
	Password  string `json:"password" binding:"required,min=6" example:"Admin@123"`
} // @name AdminRegistrationInput

var (
	ErrAdminSelfRegistrationNotAllowed = errors.New("admin registration is not allowed from the public register portal")
	ErrAdminVerificationRequired       = errors.New("admin must verify email or mobile before login")
)

// ForgotPasswordInput is used for forgot password
type ForgotPasswordInput struct {
	Email  string `json:"email" example:"rakshitha@example.com"`
	Mobile string `json:"mobile" example:"+919876543210"`
} // @name ForgotPasswordInput

// LoginInput is used for login
type LoginInput struct {
	Email    string `json:"email" example:"rakshitha@example.com"`
	Mobile   string `json:"mobile" example:"+919876543210"`
	Password string `json:"password" binding:"required" example:"Password@123"`
} // @name LoginInput

type AuthTokens struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.access.payload"`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.refresh.payload"`
} // @name AuthTokens

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id uuid.UUID) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByMobile(ctx context.Context, mobile string) (User, error)
	GetPreferencesByUserID(ctx context.Context, userID uuid.UUID) (UserPreferences, error)
	UpdatePreferences(ctx context.Context, prefs *UserPreferences) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type UserService interface {
	FindUser(ctx context.Context, id uuid.UUID) (User, error)
	GetUserProfile(ctx context.Context, id uuid.UUID) (UserProfile, error)
	Register(ctx context.Context, in UserInput) error
	RegisterAdmin(ctx context.Context, in AdminRegistrationInput) error
	Login(ctx context.Context, in LoginInput) (tokens AuthTokens, err error)
	Refresh(ctx context.Context, rawRefreshToken string) (accessToken string, err error)
	ForgotPassword(ctx context.Context, in ForgotPasswordInput) (string, error)
	ResetPassword(ctx context.Context, userID uuid.UUID, newPassword string) error
	ResetPasswordWithToken(ctx context.Context, rawToken, newPassword string) error
	VerifyEmail(ctx context.Context, rawToken string) error
	VerifyMobile(ctx context.Context, otp string) error
	ResendEmailVerification(ctx context.Context, userID uuid.UUID) error
	ResendEmailVerificationByEmail(ctx context.Context, email string) error
	ResendMobileOTP(ctx context.Context, userID uuid.UUID) error
	UpdateUserPreferences(ctx context.Context, userID uuid.UUID, in UserPreferencesInput) (UserPreferences, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type UserController interface {
	RegisterRoutes(r gin.IRouter)
}
