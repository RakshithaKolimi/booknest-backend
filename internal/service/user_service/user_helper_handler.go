package user_service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"booknest/internal/domain"
)

func (s userService) hashPassword(p string) string {
	// Generate bcrypt hash
	hashed, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.MinCost)
	if err != nil {
		slog.Error("Cannot hash the password", "error", err)
		return ""
	}

	// Return the hashed password as a string
	return string(hashed)
}

func (s userService) comparePassword(hash, raw string) bool {
	// Compare the hashed password with the raw password
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(raw)) == nil
}

func (s userService) generateRawToken() (string, error) {
	// Generate a secure random token
	b := make([]byte, 32) // 256-bit
	// Read random bytes
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	// Encode to hex string
	// example: "4f3c2e1d..."
	return hex.EncodeToString(b), nil
}

func (s userService) generateOTP(length int) string {
	// Generate a numeric OTP of specified length
	const digits = "1234567890"
	if length <= 0 {
		return ""
	}
	// Create a byte slice to hold the OTP characters
	buffer := make([]byte, length)
	// Read random bytes
	n, err := io.ReadAtLeast(rand.Reader, buffer, length)
	if n != length || err != nil {
		return ""
	}
	// Map bytes to digits
	for i := 0; i < len(buffer); i++ {
		buffer[i] = digits[int(buffer[i])%len(digits)]
	}

	// Return the OTP as a string
	// example: "493027"
	return string(buffer)
}

func (s userService) generateTokenHash(rawToken string) string {
	// Generate SHA-256 hash of the raw token
	hash := sha256.Sum256([]byte(rawToken))
	// Return the hash as a hex string
	// example: "9f2c4e5d..."
	return hex.EncodeToString(hash[:])
}

func (s userService) generateJWT(user domain.User) (string, error) {
	// Generate JWT token for the user
	secret := os.Getenv("JWT_SECRET_V1")
	if secret == "" {
		secret = "booknest_secret" // fallback for local dev
	}

	// Create JWT claims
	claims := jwt.MapClaims{
		"user_id":   user.ID.String(),
		"user_role": user.Role,
		"email":     user.Email,
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
		"iat":       time.Now().Unix(),
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = domain.CurrentKeyID

	// Sign the token with the secret
	return token.SignedString([]byte(secret))
}

func (s *userService) verifyToken(
	ctx context.Context,
	rawToken string,
	tokenType domain.VerificationTokenType,
) error {
	// Generate token hash
	tokenHash := s.generateTokenHash(rawToken)

	// Use transaction for the verification process
	return s.txm.InTransaction(ctx, func(txCtx context.Context) error {

		// 1. Find token
		token, err := s.vtr.FindByHashAndType(txCtx, tokenHash, tokenType)
		if err != nil {
			return errors.New("invalid or expired token")
		}

		// 2. Validate token
		if token.IsUsed {
			return errors.New("token already used")
		}

		if time.Now().After(token.ExpiresAt) {
			return errors.New("token expired")
		}

		// 3. Fetch user
		user, err := s.r.FindByID(txCtx, token.UserID)
		if err != nil {
			return err
		}

		// 4. Update verification flags
		switch tokenType {
		case domain.VerificationEmail:
			user.EmailVerified = true
		case domain.VerificationMobile:
			user.MobileVerified = true
		}

		if err := s.r.Update(txCtx, &user); err != nil {
			return err
		}

		// 5. Mark token as used
		now := time.Now()
		token.IsUsed = true
		token.UsedAt = &now

		return s.vtr.Update(txCtx, token)
	})
}

func (s *userService) sendEmailVerification(email string, token domain.VerificationToken) {
	slog.Debug("Sending Email...", "email", email, "token:", token.TokenHash)
}

func (s *userService) sendMobileVerification(mobile, otp string) {
	slog.Debug("Sending OTP...", "mobile:", mobile, "otp:", otp)
}
