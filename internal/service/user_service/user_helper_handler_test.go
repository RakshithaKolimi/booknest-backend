package user_service

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"booknest/internal/domain"
)

// MockUserRepository is a mock implementation of domain.UserRepository
type MockUserRepository struct {
	CreateFunc       func(ctx context.Context, user *domain.User) error
	FindByIDFunc     func(ctx context.Context, id uuid.UUID) (domain.User, error)
	FindByEmailFunc  func(ctx context.Context, email string) (domain.User, error)
	FindByMobileFunc func(ctx context.Context, mobile string) (domain.User, error)
	UpdateFunc       func(ctx context.Context, user *domain.User) error
	DeleteFunc       func(ctx context.Context, id uuid.UUID) error
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return domain.User{}, errors.New("not implemented")
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	if m.FindByEmailFunc != nil {
		return m.FindByEmailFunc(ctx, email)
	}
	return domain.User{}, errors.New("not implemented")
}

func (m *MockUserRepository) FindByMobile(ctx context.Context, mobile string) (domain.User, error) {
	if m.FindByMobileFunc != nil {
		return m.FindByMobileFunc(ctx, mobile)
	}
	return domain.User{}, errors.New("not implemented")
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// MockVerificationTokenRepository is a mock implementation
type MockVerificationTokenRepository struct {
	CreateFunc                  func(ctx context.Context, token *domain.VerificationToken) error
	FindByHashAndTypeFunc       func(ctx context.Context, tokenHash string, tokenType domain.VerificationTokenType) (*domain.VerificationToken, error)
	FindByUserIDAndTypeFunc     func(ctx context.Context, userID uuid.UUID, tokenType domain.VerificationTokenType) (*domain.VerificationToken, error)
	UpdateFunc                  func(ctx context.Context, token *domain.VerificationToken) error
	InvalidateByUserAndTypeFunc func(ctx context.Context, userID uuid.UUID, tokenType domain.VerificationTokenType) error
	DeleteFunc                  func(ctx context.Context, id uuid.UUID) error
}

func (m *MockVerificationTokenRepository) Create(ctx context.Context, token *domain.VerificationToken) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, token)
	}
	return nil
}

func (m *MockVerificationTokenRepository) FindByHashAndType(ctx context.Context, tokenHash string, tokenType domain.VerificationTokenType) (*domain.VerificationToken, error) {
	if m.FindByHashAndTypeFunc != nil {
		return m.FindByHashAndTypeFunc(ctx, tokenHash, tokenType)
	}
	return nil, errors.New("not implemented")
}

func (m *MockVerificationTokenRepository) FindByUserIDAndType(ctx context.Context, userID uuid.UUID, tokenType domain.VerificationTokenType) (*domain.VerificationToken, error) {
	if m.FindByUserIDAndTypeFunc != nil {
		return m.FindByUserIDAndTypeFunc(ctx, userID, tokenType)
	}
	return nil, errors.New("not implemented")
}

func (m *MockVerificationTokenRepository) Update(ctx context.Context, token *domain.VerificationToken) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, token)
	}
	return nil
}

func (m *MockVerificationTokenRepository) InvalidateByUserAndType(ctx context.Context, userID uuid.UUID, tokenType domain.VerificationTokenType) error {
	if m.InvalidateByUserAndTypeFunc != nil {
		return m.InvalidateByUserAndTypeFunc(ctx, userID, tokenType)
	}
	return nil
}

func (m *MockVerificationTokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// TestHashPassword_Success tests successful password hashing
func TestHashPassword_Success(t *testing.T) {
	service := &userService{}
	password := "securepassword123"

	hashed := service.hashPassword(password)

	if hashed == "" {
		t.Fatalf("expected non-empty hashed password, got empty string")
	}

	if hashed == password {
		t.Fatalf("hashed password should not equal raw password")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	if err != nil {
		t.Fatalf("bcrypt comparison failed: %v", err)
	}
}

// TestComparePassword_Success tests successful password comparison
func TestComparePassword_Success(t *testing.T) {
	service := &userService{}
	password := "securepassword123"
	hashed := service.hashPassword(password)

	result := service.comparePassword(hashed, password)

	if !result {
		t.Fatalf("expected password comparison to succeed")
	}
}

// TestComparePassword_Failure tests password comparison with wrong password
func TestComparePassword_Failure(t *testing.T) {
	service := &userService{}
	password := "securepassword123"
	hashed := service.hashPassword(password)

	result := service.comparePassword(hashed, "wrongpassword")

	if result {
		t.Fatalf("expected password comparison to fail")
	}
}

// TestGenerateRawToken_Success tests successful raw token generation
func TestGenerateRawToken_Success(t *testing.T) {
	service := &userService{}

	token1, err := service.generateRawToken()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if token1 == "" {
		t.Fatalf("expected non-empty token")
	}

	token2, err := service.generateRawToken()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if token1 == token2 {
		t.Fatalf("expected different tokens, got same token")
	}

	if len(token1) != 64 {
		t.Fatalf("expected 64 character token, got %d", len(token1))
	}
}

// TestGenerateOTP_Success tests successful OTP generation
func TestGenerateOTP_Success(t *testing.T) {
	service := &userService{}

	otp := service.generateOTP(6)

	if len(otp) != 6 {
		t.Fatalf("expected OTP of length 6, got %d", len(otp))
	}

	for _, ch := range otp {
		if ch < '0' || ch > '9' {
			t.Fatalf("expected digit character, got %c", ch)
		}
	}
}

// TestGenerateOTP_DifferentLengths tests OTP generation with different lengths
func TestGenerateOTP_DifferentLengths(t *testing.T) {
	service := &userService{}

	tests := []int{4, 6, 8}
	for _, length := range tests {
		otp := service.generateOTP(length)
		if len(otp) != length {
			t.Fatalf("expected OTP of length %d, got %d", length, len(otp))
		}
	}
}

// TestGenerateOTP_Randomness tests that OTP is random
func TestGenerateOTP_Randomness(t *testing.T) {
	service := &userService{}

	otp1 := service.generateOTP(6)
	otp2 := service.generateOTP(6)

	if otp1 == otp2 {
		t.Fatalf("expected different OTPs, got same OTP")
	}
}

// TestGenerateOTP_InvalidLength tests OTP generation with invalid length
func TestGenerateOTP_InvalidLength(t *testing.T) {
	service := &userService{}

	otp := service.generateOTP(0)

	if otp != "" {
		t.Fatalf("expected empty OTP for length 0, got %s", otp)
	}
}

// TestGenerateTokenHash_Success tests successful token hash generation
func TestGenerateTokenHash_Success(t *testing.T) {
	service := &userService{}
	token := "test_token_12345"

	hash := service.generateTokenHash(token)

	if hash == "" {
		t.Fatalf("expected non-empty hash, got empty string")
	}

	if len(hash) != 64 {
		t.Fatalf("expected 64 character hash, got %d", len(hash))
	}

	hash2 := service.generateTokenHash(token)
	if hash != hash2 {
		t.Fatalf("expected same hash for same token")
	}

	hash3 := service.generateTokenHash("different_token")
	if hash == hash3 {
		t.Fatalf("expected different hashes for different tokens")
	}
}

// TestGenerateJWT_Success tests successful JWT generation
func TestGenerateJWT_Success(t *testing.T) {
	os.Setenv("JWT_SECRET_V1", "test_secret_key")
	service := &userService{}
	userID := uuid.New()

	user := domain.User{
		ID:    userID,
		Email: "test@example.com",
		Role:  domain.UserRoleUser,
	}

	token, err := service.generateJWT(user)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if token == "" {
		t.Fatalf("expected non-empty token")
	}

	parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte("test_secret_key"), nil
	})

	if err != nil {
		t.Fatalf("expected valid JWT, got error: %v", err)
	}

	if !parsed.Valid {
		t.Fatalf("expected valid token")
	}

	claims := parsed.Claims.(jwt.MapClaims)
	if claims["user_id"] != userID.String() {
		t.Fatalf("expected user_id in claims")
	}
	if claims["email"] != "test@example.com" {
		t.Fatalf("expected email in claims")
	}
}

// TestGenerateJWT_DefaultSecret tests JWT generation uses default secret when env var is missing
func TestGenerateJWT_DefaultSecret(t *testing.T) {
	os.Unsetenv("JWT_SECRET_V1")
	service := &userService{}
	userID := uuid.New()

	user := domain.User{
		ID:    userID,
		Email: "test@example.com",
		Role:  domain.UserRoleUser,
	}

	token, err := service.generateJWT(user)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if token == "" {
		t.Fatalf("expected non-empty token")
	}

	parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte("booknest_secret"), nil
	})

	if err != nil {
		t.Fatalf("expected valid JWT, got error: %v", err)
	}

	if !parsed.Valid {
		t.Fatalf("expected valid token")
	}
}

// TestGenerateJWT_ExpirationTime tests JWT has correct expiration time
func TestGenerateJWT_ExpirationTime(t *testing.T) {
	os.Setenv("JWT_SECRET_V1", "test_secret_key")
	service := &userService{}
	user := domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  domain.UserRoleUser,
	}

	token, err := service.generateJWT(user)
	afterGeneration := time.Now()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	parsed, _ := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte("test_secret_key"), nil
	})

	claims := parsed.Claims.(jwt.MapClaims)
	exp := int64(claims["exp"].(float64))
	expTime := time.Unix(exp, 0)

	expectedExpiration := afterGeneration.Add(15 * time.Minute)
	diffSeconds := expTime.Sub(expectedExpiration).Seconds()

	if diffSeconds < -1 || diffSeconds > 1 {
		t.Fatalf("expected expiration in ~15 minutes, got difference of %v seconds", diffSeconds)
	}
}

func TestGenerateRefreshToken_Success(t *testing.T) {
	t.Setenv("JWT_REFRESH_V1", "refresh-secret")
	service := &userService{}
	userID := uuid.New()

	user := domain.User{
		ID:    userID,
		Email: "refresh@example.com",
		Role:  domain.UserRoleUser,
	}

	token, err := service.generateRefreshToken(user)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatalf("expected non-empty token")
	}

	parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if token.Header["kid"] != domain.CurrentRefreshKeyID {
			t.Fatalf("expected refresh kid %s, got %v", domain.CurrentRefreshKeyID, token.Header["kid"])
		}
		return []byte("refresh-secret"), nil
	})
	if err != nil || !parsed.Valid {
		t.Fatalf("expected valid refresh token, got err=%v valid=%v", err, parsed != nil && parsed.Valid)
	}
}

func TestVerifyRefreshJWT_Success(t *testing.T) {
	t.Setenv("JWT_REFRESH_V1", "refresh-secret")
	service := &userService{}
	user := domain.User{
		ID:    uuid.New(),
		Email: "refresh@example.com",
		Role:  domain.UserRoleUser,
	}

	raw, err := service.generateRefreshToken(user)
	if err != nil {
		t.Fatalf("generateRefreshToken failed: %v", err)
	}

	claims, err := service.verifyRefreshJWT(raw)
	if err != nil {
		t.Fatalf("expected verify success, got %v", err)
	}
	if claims["user_id"] != user.ID.String() {
		t.Fatalf("expected user_id claim to match")
	}
}

func TestVerifyRefreshJWT_UsesFallbackSecret(t *testing.T) {
	os.Unsetenv("JWT_REFRESH_V1")
	os.Unsetenv("JWT_REFRESH_V0")

	service := &userService{}
	user := domain.User{
		ID:    uuid.New(),
		Email: "refresh@example.com",
		Role:  domain.UserRoleUser,
	}

	raw, err := service.generateRefreshToken(user)
	if err != nil {
		t.Fatalf("generateRefreshToken failed: %v", err)
	}

	if _, err := service.verifyRefreshJWT(raw); err != nil {
		t.Fatalf("expected fallback verify success, got %v", err)
	}
}

func TestVerifyRefreshJWT_InvalidKid(t *testing.T) {
	t.Setenv("JWT_REFRESH_V1", "refresh-secret")
	service := &userService{}

	claims := jwt.MapClaims{
		"user_id": uuid.New().String(),
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = "UNKNOWN_REFRESH_KID"

	raw, err := token.SignedString([]byte("refresh-secret"))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	if _, err := service.verifyRefreshJWT(raw); err == nil {
		t.Fatalf("expected invalid kid error")
	}
}

func TestUserIDFromClaims(t *testing.T) {
	service := &userService{}
	userID := uuid.New()

	got, err := service.userIDFromClaims(jwt.MapClaims{"user_id": userID.String()})
	if err != nil || got != userID {
		t.Fatalf("expected parsed user id, got %v err=%v", got, err)
	}

	if _, err := service.userIDFromClaims(jwt.MapClaims{}); err == nil {
		t.Fatalf("expected error when user_id is missing")
	}

	if _, err := service.userIDFromClaims(jwt.MapClaims{"user_id": "bad-uuid"}); err == nil {
		t.Fatalf("expected error for invalid UUID")
	}
}

// Note: TestVerifyToken tests are not included here as they require
// a database pool and transaction context which cannot be easily mocked.
// These should be tested through integration tests with a test database.
