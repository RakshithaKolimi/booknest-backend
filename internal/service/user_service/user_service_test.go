package user_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"booknest/internal/domain"
)

type noopTransactionManager struct{}

func (n *noopTransactionManager) InTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

// TestFindUser_Success tests successful user retrieval
func TestFindUser_Success(t *testing.T) {
	userID := uuid.New()
	expectedUser := domain.User{
		ID:        userID,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Mobile:    "1234567890",
	}

	mockUserRepo := &MockUserRepository{
		FindByIDFunc: func(ctx context.Context, id uuid.UUID) (domain.User, error) {
			if id == userID {
				return expectedUser, nil
			}
			return domain.User{}, errors.New("user not found")
		},
	}

	service := &userService{
		txm: &noopTransactionManager{},
		r:   mockUserRepo,
		vtr: &MockVerificationTokenRepository{},
	}
	user, err := service.FindUser(context.Background(), userID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.ID != userID {
		t.Fatalf("expected user ID %s, got %s", userID, user.ID)
	}
}

// TestFindUser_NotFound tests user retrieval when user doesn't exist
func TestFindUser_NotFound(t *testing.T) {
	mockUserRepo := &MockUserRepository{
		FindByIDFunc: func(ctx context.Context, id uuid.UUID) (domain.User, error) {
			return domain.User{}, errors.New("user not found")
		},
	}

	service := &userService{
		txm: &noopTransactionManager{},
		r:   mockUserRepo,
		vtr: &MockVerificationTokenRepository{},
	}
	_, err := service.FindUser(context.Background(), uuid.New())

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

// TestLogin_SuccessByEmail tests successful login with email
func TestLogin_SuccessByEmail(t *testing.T) {
	userID := uuid.New()
	password := "password123"

	mockUserRepo := &MockUserRepository{
		FindByEmailFunc: func(ctx context.Context, email string) (domain.User, error) {
			service := &userService{}
			hashedPassword := service.hashPassword(password)
			return domain.User{
				ID:       userID,
				Email:    email,
				Password: hashedPassword,
				Role:     domain.UserRoleUser,
			}, nil
		},
		UpdateFunc: func(ctx context.Context, user *domain.User) error {
			return nil
		},
	}

	service := &userService{
		txm: &noopTransactionManager{},
		r:   mockUserRepo,
		vtr: &MockVerificationTokenRepository{},
	}
	input := domain.LoginInput{
		Email:    "test@example.com",
		Password: password,
	}

	tokens, err := service.Login(context.Background(), input)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("expected non-empty access and refresh tokens")
	}
}

// TestLogin_SuccessByMobile tests successful login with mobile
func TestLogin_SuccessByMobile(t *testing.T) {
	userID := uuid.New()
	password := "password123"

	mockUserRepo := &MockUserRepository{
		FindByMobileFunc: func(ctx context.Context, mobile string) (domain.User, error) {
			service := &userService{}
			hashedPassword := service.hashPassword(password)
			return domain.User{
				ID:       userID,
				Mobile:   mobile,
				Password: hashedPassword,
				Role:     domain.UserRoleUser,
			}, nil
		},
		UpdateFunc: func(ctx context.Context, user *domain.User) error {
			return nil
		},
	}

	service := &userService{
		txm: &noopTransactionManager{},
		r:   mockUserRepo,
		vtr: &MockVerificationTokenRepository{},
	}
	input := domain.LoginInput{
		Mobile:   "1234567890",
		Password: password,
	}

	tokens, err := service.Login(context.Background(), input)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("expected non-empty access and refresh tokens")
	}
}

// TestLogin_InvalidPassword tests login with incorrect password
func TestLogin_InvalidPassword(t *testing.T) {
	password := "correctpassword123"

	mockUserRepo := &MockUserRepository{
		FindByEmailFunc: func(ctx context.Context, email string) (domain.User, error) {
			service := &userService{}
			hashedPassword := service.hashPassword(password)
			return domain.User{
				ID:       uuid.New(),
				Email:    email,
				Password: hashedPassword,
				Role:     domain.UserRoleUser,
			}, nil
		},
	}

	service := &userService{
		txm: &noopTransactionManager{},
		r:   mockUserRepo,
		vtr: &MockVerificationTokenRepository{},
	}
	input := domain.LoginInput{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	_, err := service.Login(context.Background(), input)

	if err == nil {
		t.Fatalf("expected error for invalid password")
	}

	if err.Error() != "invalid credentials" {
		t.Fatalf("expected 'invalid credentials' error, got %v", err)
	}
}

func TestLogin_AdminRequiresVerification(t *testing.T) {
	password := "password123"

	service := &userService{
		txm: &noopTransactionManager{},
		r: &MockUserRepository{
			FindByEmailFunc: func(ctx context.Context, email string) (domain.User, error) {
				return domain.User{
					ID:             uuid.New(),
					Email:          email,
					Password:       (&userService{}).hashPassword(password),
					Role:           domain.UserRoleAdmin,
					EmailVerified:  false,
					MobileVerified: false,
				}, nil
			},
			UpdateFunc: func(ctx context.Context, user *domain.User) error {
				t.Fatal("admin should not be updated when verification gate blocks login")
				return nil
			},
		},
		vtr: &MockVerificationTokenRepository{
			CreateFunc: func(ctx context.Context, token *domain.VerificationToken) error {
				t.Fatal("refresh token should not be created for blocked admin login")
				return nil
			},
		},
	}

	_, err := service.Login(context.Background(), domain.LoginInput{
		Email:    "admin@example.com",
		Password: password,
	})
	if !errors.Is(err, domain.ErrAdminVerificationRequired) {
		t.Fatalf("expected admin verification error, got %v", err)
	}
}

func TestLogin_AdminAllowedAfterEmailVerification(t *testing.T) {
	password := "password123"

	service := &userService{
		txm: &noopTransactionManager{},
		r: &MockUserRepository{
			FindByEmailFunc: func(ctx context.Context, email string) (domain.User, error) {
				return domain.User{
					ID:             uuid.New(),
					Email:          email,
					Password:       (&userService{}).hashPassword(password),
					Role:           domain.UserRoleAdmin,
					EmailVerified:  true,
					MobileVerified: false,
				}, nil
			},
			UpdateFunc: func(ctx context.Context, user *domain.User) error { return nil },
		},
		vtr: &MockVerificationTokenRepository{
			CreateFunc: func(ctx context.Context, token *domain.VerificationToken) error { return nil },
		},
	}

	tokens, err := service.Login(context.Background(), domain.LoginInput{
		Email:    "admin@example.com",
		Password: password,
	})
	if err != nil {
		t.Fatalf("expected login success, got %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("expected access and refresh tokens")
	}
}

func TestLogin_AdminAllowedAfterMobileVerification(t *testing.T) {
	password := "password123"

	service := &userService{
		txm: &noopTransactionManager{},
		r: &MockUserRepository{
			FindByEmailFunc: func(ctx context.Context, email string) (domain.User, error) {
				return domain.User{
					ID:             uuid.New(),
					Email:          email,
					Password:       (&userService{}).hashPassword(password),
					Role:           domain.UserRoleAdmin,
					EmailVerified:  false,
					MobileVerified: true,
				}, nil
			},
			UpdateFunc: func(ctx context.Context, user *domain.User) error { return nil },
		},
		vtr: &MockVerificationTokenRepository{
			CreateFunc: func(ctx context.Context, token *domain.VerificationToken) error { return nil },
		},
	}

	tokens, err := service.Login(context.Background(), domain.LoginInput{
		Email:    "admin@example.com",
		Password: password,
	})
	if err != nil {
		t.Fatalf("expected login success, got %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("expected access and refresh tokens")
	}
}

// TestLogin_UserNotFound tests login when user doesn't exist
func TestLogin_UserNotFound(t *testing.T) {
	mockUserRepo := &MockUserRepository{
		FindByEmailFunc: func(ctx context.Context, email string) (domain.User, error) {
			return domain.User{}, errors.New("user not found")
		},
	}

	service := &userService{
		txm: &noopTransactionManager{},
		r:   mockUserRepo,
		vtr: &MockVerificationTokenRepository{},
	}
	input := domain.LoginInput{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	_, err := service.Login(context.Background(), input)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

// TestLogin_UpdatesLastLogin tests that login updates last_login timestamp
func TestLogin_UpdatesLastLogin(t *testing.T) {
	userID := uuid.New()
	password := "password123"
	updateCalled := false
	var capturedUser *domain.User

	mockUserRepo := &MockUserRepository{
		FindByEmailFunc: func(ctx context.Context, email string) (domain.User, error) {
			service := &userService{}
			hashedPassword := service.hashPassword(password)
			return domain.User{
				ID:       userID,
				Email:    email,
				Password: hashedPassword,
				Role:     domain.UserRoleUser,
			}, nil
		},
		UpdateFunc: func(ctx context.Context, user *domain.User) error {
			updateCalled = true
			capturedUser = user
			return nil
		},
	}

	service := &userService{
		txm: &noopTransactionManager{},
		r:   mockUserRepo,
		vtr: &MockVerificationTokenRepository{},
	}
	input := domain.LoginInput{
		Email:    "test@example.com",
		Password: password,
	}

	tokens, err := service.Login(context.Background(), input)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("expected non-empty access and refresh tokens")
	}

	if !updateCalled {
		t.Fatalf("expected Update to be called")
	}

	if capturedUser.LastLogin == nil {
		t.Fatalf("expected LastLogin to be set")
	}
}

func TestRegister_Success(t *testing.T) {
	createdUsers := 0
	createdTokens := 0

	service := &userService{
		txm: &noopTransactionManager{},
		r: &MockUserRepository{
			CreateFunc: func(ctx context.Context, user *domain.User) error {
				createdUsers++
				if user.ID == uuid.Nil || user.Email != "jane@example.com" || user.Mobile != "+15555550123" {
					t.Fatalf("unexpected user payload: %+v", user)
				}
				if user.Password == "" || user.Password == "password123" {
					t.Fatalf("expected hashed password, got %q", user.Password)
				}
				return nil
			},
		},
		vtr: &MockVerificationTokenRepository{
			CreateFunc: func(ctx context.Context, token *domain.VerificationToken) error {
				createdTokens++
				if token.UserID == uuid.Nil {
					t.Fatalf("expected token user id")
				}
				if token.Type != domain.VerificationEmail && token.Type != domain.VerificationMobile {
					t.Fatalf("unexpected token type: %s", token.Type)
				}
				if token.TokenHash == "" {
					t.Fatalf("expected token hash")
				}
				return nil
			},
		},
	}

	err := service.Register(context.Background(), domain.UserInput{
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
		Mobile:    "+15555550123",
		Password:  "password123",
		Role:      domain.UserRoleUser,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if createdUsers != 1 || createdTokens != 2 {
		t.Fatalf("expected 1 user and 2 tokens, got users=%d tokens=%d", createdUsers, createdTokens)
	}
}

func TestRegister_RejectsAdminSelfRegistration(t *testing.T) {
	service := &userService{
		txm: &noopTransactionManager{},
		r: &MockUserRepository{
			CreateFunc: func(ctx context.Context, user *domain.User) error {
				t.Fatal("user should not be created for public admin registration")
				return nil
			},
		},
		vtr: &MockVerificationTokenRepository{},
	}

	err := service.Register(context.Background(), domain.UserInput{
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
		Mobile:    "+15555550123",
		Password:  "password123",
		Role:      domain.UserRoleAdmin,
	})
	if !errors.Is(err, domain.ErrAdminSelfRegistrationNotAllowed) {
		t.Fatalf("expected admin self registration error, got %v", err)
	}
}

func TestRegisterAdmin_SetsAdminRole(t *testing.T) {
	createdUsers := 0

	service := &userService{
		txm: &noopTransactionManager{},
		r: &MockUserRepository{
			CreateFunc: func(ctx context.Context, user *domain.User) error {
				createdUsers++
				if user.Role != domain.UserRoleAdmin {
					t.Fatalf("expected admin role, got %s", user.Role)
				}
				return nil
			},
		},
		vtr: &MockVerificationTokenRepository{
			CreateFunc: func(ctx context.Context, token *domain.VerificationToken) error {
				return nil
			},
		},
	}

	err := service.RegisterAdmin(context.Background(), domain.AdminRegistrationInput{
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
		Mobile:    "+15555550123",
		Password:  "password123",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if createdUsers != 1 {
		t.Fatalf("expected admin user create call, got %d", createdUsers)
	}
}

func TestRegister_StopsWhenUserCreateFails(t *testing.T) {
	service := &userService{
		txm: &noopTransactionManager{},
		r: &MockUserRepository{
			CreateFunc: func(ctx context.Context, user *domain.User) error {
				return errors.New("create failed")
			},
		},
		vtr: &MockVerificationTokenRepository{
			CreateFunc: func(ctx context.Context, token *domain.VerificationToken) error {
				t.Fatal("verification token should not be created when user creation fails")
				return nil
			},
		},
	}

	err := service.Register(context.Background(), domain.UserInput{
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
		Mobile:    "+15555550123",
		Password:  "password123",
		Role:      domain.UserRoleUser,
	})
	if err == nil || err.Error() != "create failed" {
		t.Fatalf("expected create failure, got %v", err)
	}
}

func TestRegister_StopsWhenVerificationCreateFails(t *testing.T) {
	createCalls := 0
	service := &userService{
		txm: &noopTransactionManager{},
		r: &MockUserRepository{
			CreateFunc: func(ctx context.Context, user *domain.User) error {
				return nil
			},
		},
		vtr: &MockVerificationTokenRepository{
			CreateFunc: func(ctx context.Context, token *domain.VerificationToken) error {
				createCalls++
				if createCalls == 1 {
					return errors.New("token create failed")
				}
				return nil
			},
		},
	}

	err := service.Register(context.Background(), domain.UserInput{
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
		Mobile:    "+15555550123",
		Password:  "password123",
		Role:      domain.UserRoleUser,
	})
	if err == nil || err.Error() != "token create failed" {
		t.Fatalf("expected token create failure, got %v", err)
	}
}

// TestResetPassword_Success tests successful password reset
// Note: This test requires a real DB pool and transaction support,
// so it's skipped and should be tested through integration tests
func TestResetPassword_Success(t *testing.T) {
	t.Skip("ResetPassword requires database transaction, tested through integration tests")
}

// TestDeleteUser_Success tests successful user deletion
func TestDeleteUser_Success(t *testing.T) {
	userID := uuid.New()
	deleteCalled := false

	mockUserRepo := &MockUserRepository{
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			if id == userID {
				deleteCalled = true
				return nil
			}
			return errors.New("user not found")
		},
	}

	service := &userService{
		txm: &noopTransactionManager{},
		r:   mockUserRepo,
		vtr: &MockVerificationTokenRepository{},
	}
	err := service.DeleteUser(context.Background(), userID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !deleteCalled {
		t.Fatalf("expected Delete to be called")
	}
}

// TestDeleteUser_NotFound tests deletion of non-existent user
func TestDeleteUser_NotFound(t *testing.T) {
	mockUserRepo := &MockUserRepository{
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return errors.New("user not found")
		},
	}

	service := &userService{
		txm: &noopTransactionManager{},
		r:   mockUserRepo,
		vtr: &MockVerificationTokenRepository{},
	}
	err := service.DeleteUser(context.Background(), uuid.New())

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

// TestVerifyEmail_CallsVerifyToken tests that VerifyEmail delegates to verifyToken
func TestVerifyEmail_CallsVerifyToken(t *testing.T) {
	// This test would require mocking the verifyToken method
	// Since verifyToken requires DB transaction, it's better tested through integration tests
	t.Skip("VerifyEmail requires database transaction, tested through integration tests")
}

// TestVerifyMobile_CallsVerifyToken tests that VerifyMobile delegates to verifyToken
func TestVerifyMobile_CallsVerifyToken(t *testing.T) {
	// This test would require mocking the verifyToken method
	// Since verifyToken requires DB transaction, it's better tested through integration tests
	t.Skip("VerifyMobile requires database transaction, tested through integration tests")
}

// TestResendEmailVerification_EmailAlreadyVerified tests resending when email already verified
// Note: This test requires a real DB pool and transaction support,
// so it's skipped and should be tested through integration tests
func TestResendEmailVerification_EmailAlreadyVerified(t *testing.T) {
	t.Skip("ResendEmailVerification requires database transaction, tested through integration tests")
}

// TestResendMobileOTP_MobileAlreadyVerified tests resending when mobile already verified
// Note: This test requires a real DB pool and transaction support,
// so it's skipped and should be tested through integration tests
func TestResendMobileOTP_MobileAlreadyVerified(t *testing.T) {
	t.Skip("ResendMobileOTP requires database transaction, tested through integration tests")
}

// TestNewUserService tests service initialization
func TestNewUserService(t *testing.T) {
	mockUserRepo := &MockUserRepository{}
	mockVerificationRepo := &MockVerificationTokenRepository{}
	txm := &noopTransactionManager{}

	service := NewUserService(txm, mockUserRepo, mockVerificationRepo)

	if service == nil {
		t.Fatalf("expected non-nil service")
	}

	userService, ok := service.(*userService)
	if !ok {
		t.Fatalf("expected *userService type")
	}

	if userService.r != mockUserRepo {
		t.Fatalf("expected user repository to be set")
	}

	if userService.vtr != mockVerificationRepo {
		t.Fatalf("expected verification token repository to be set")
	}

	if userService.txm != txm {
		t.Fatalf("expected transaction manager to be set")
	}
}

// TestLogin_TrimsContextualEmailAndMobile tests login prefers email when both provided
func TestLogin_PrefersEmailOverMobile(t *testing.T) {
	userID := uuid.New()
	password := "password123"
	emailCalled := false

	mockUserRepo := &MockUserRepository{
		FindByEmailFunc: func(ctx context.Context, email string) (domain.User, error) {
			emailCalled = true
			service := &userService{}
			hashedPassword := service.hashPassword(password)
			return domain.User{
				ID:       userID,
				Email:    email,
				Password: hashedPassword,
				Role:     domain.UserRoleUser,
			}, nil
		},
		FindByMobileFunc: func(ctx context.Context, mobile string) (domain.User, error) {
			t.Fatalf("should not call FindByMobile when email is provided")
			return domain.User{}, errors.New("should not be called")
		},
		UpdateFunc: func(ctx context.Context, user *domain.User) error {
			return nil
		},
	}

	service := &userService{
		txm: &noopTransactionManager{},
		r:   mockUserRepo,
		vtr: &MockVerificationTokenRepository{},
	}
	input := domain.LoginInput{
		Email:    "test@example.com",
		Mobile:   "1234567890",
		Password: password,
	}

	_, err := service.Login(context.Background(), input)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !emailCalled {
		t.Fatalf("expected FindByEmail to be called")
	}
}

// TestFindUser_PassesContext tests that FindUser passes context properly
func TestFindUser_PassesContext(t *testing.T) {
	userID := uuid.New()
	contextPassed := false

	mockUserRepo := &MockUserRepository{
		FindByIDFunc: func(ctx context.Context, id uuid.UUID) (domain.User, error) {
			if ctx == nil {
				t.Fatalf("expected non-nil context")
			}
			contextPassed = true
			return domain.User{ID: id}, nil
		},
	}

	service := &userService{r: mockUserRepo}
	user, err := service.FindUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID != userID {
		t.Fatalf("expected user ID %s, got %s", userID, user.ID)
	}

	if !contextPassed {
		t.Fatalf("expected context to be passed to repository")
	}
}

// TestDeleteUser_PassesContext tests that DeleteUser passes context properly
func TestDeleteUser_PassesContext(t *testing.T) {
	userID := uuid.New()
	contextPassed := false

	mockUserRepo := &MockUserRepository{
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			if ctx == nil {
				t.Fatalf("expected non-nil context")
			}
			contextPassed = true
			return nil
		},
	}

	service := &userService{r: mockUserRepo}
	err := service.DeleteUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !contextPassed {
		t.Fatalf("expected context to be passed to repository")
	}
}

func TestForgotPassword_ReturnsTokenAndInvalidatesOldOnEmail(t *testing.T) {
	userID := uuid.New()
	created := 0
	invalidated := 0

	service := &userService{
		txm: &noopTransactionManager{},
		r: &MockUserRepository{
			FindByEmailFunc: func(ctx context.Context, email string) (domain.User, error) {
				return domain.User{ID: userID, Email: email}, nil
			},
		},
		vtr: &MockVerificationTokenRepository{
			InvalidateByUserAndTypeFunc: func(ctx context.Context, gotUserID uuid.UUID, tokenType domain.VerificationTokenType) error {
				invalidated++
				if gotUserID != userID || tokenType != domain.PasswordReset {
					t.Fatalf("unexpected invalidation params")
				}
				return nil
			},
			CreateFunc: func(ctx context.Context, token *domain.VerificationToken) error {
				created++
				if token.UserID != userID || token.Type != domain.PasswordReset {
					t.Fatalf("unexpected token payload")
				}
				return nil
			},
		},
	}

	token, err := service.ForgotPassword(context.Background(), domain.ForgotPasswordInput{Email: "user@test.com"})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if token == "" {
		t.Fatalf("expected non-empty token")
	}
	if created != 1 || invalidated != 1 {
		t.Fatalf("expected create+invalidate once, got create=%d invalidate=%d", created, invalidated)
	}
}

func TestResetPassword_SuccessUnit(t *testing.T) {
	userID := uuid.New()
	var updatedUser domain.User

	service := &userService{
		txm: &noopTransactionManager{},
		r: &MockUserRepository{
			FindByIDFunc: func(ctx context.Context, id uuid.UUID) (domain.User, error) {
				return domain.User{ID: userID, Password: "old"}, nil
			},
			UpdateFunc: func(ctx context.Context, user *domain.User) error {
				updatedUser = *user
				return nil
			},
		},
		vtr: &MockVerificationTokenRepository{},
	}

	err := service.ResetPassword(context.Background(), userID, "newpassword123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updatedUser.ID != userID || updatedUser.Password == "" || updatedUser.Password == "newpassword123" {
		t.Fatalf("expected hashed password update, got %+v", updatedUser)
	}
}

func TestResetPasswordWithToken_Success(t *testing.T) {
	userID := uuid.New()
	rawToken := "reset-token"
	expiresAt := time.Now().Add(30 * time.Minute)
	updateTokenCalled := false
	invalidatedCalled := false

	service := &userService{
		txm: &noopTransactionManager{},
		r: &MockUserRepository{
			FindByIDFunc: func(ctx context.Context, id uuid.UUID) (domain.User, error) {
				return domain.User{ID: userID, Password: "old"}, nil
			},
			UpdateFunc: func(ctx context.Context, user *domain.User) error {
				if user.Password == "" || user.Password == "newpassword123" {
					t.Fatalf("expected hashed password, got %q", user.Password)
				}
				return nil
			},
		},
		vtr: &MockVerificationTokenRepository{
			FindByHashAndTypeFunc: func(ctx context.Context, tokenHash string, tokenType domain.VerificationTokenType) (*domain.VerificationToken, error) {
				if tokenHash == "" || tokenType != domain.PasswordReset {
					t.Fatalf("unexpected token lookup args")
				}
				return &domain.VerificationToken{UserID: userID, Type: domain.PasswordReset, ExpiresAt: expiresAt}, nil
			},
			UpdateFunc: func(ctx context.Context, token *domain.VerificationToken) error {
				updateTokenCalled = true
				if !token.IsUsed || token.UsedAt == nil {
					t.Fatalf("expected token to be marked used")
				}
				return nil
			},
			InvalidateByUserAndTypeFunc: func(ctx context.Context, gotUserID uuid.UUID, tokenType domain.VerificationTokenType) error {
				invalidatedCalled = true
				if gotUserID != userID || tokenType != domain.PasswordReset {
					t.Fatalf("unexpected invalidation args")
				}
				return nil
			},
		},
	}

	err := service.ResetPasswordWithToken(context.Background(), rawToken, "newpassword123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !updateTokenCalled || !invalidatedCalled {
		t.Fatalf("expected token update and invalidation")
	}
}

func TestResetPasswordWithToken_InvalidOrExpired(t *testing.T) {
	userID := uuid.New()
	tests := []struct {
		name  string
		token *domain.VerificationToken
		err   error
	}{
		{name: "missing token", err: errors.New("not found")},
		{name: "used token", token: &domain.VerificationToken{UserID: userID, Type: domain.PasswordReset, IsUsed: true, ExpiresAt: time.Now().Add(time.Hour)}},
		{name: "expired token", token: &domain.VerificationToken{UserID: userID, Type: domain.PasswordReset, ExpiresAt: time.Now().Add(-time.Hour)}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := &userService{
				txm: &noopTransactionManager{},
				r:   &MockUserRepository{},
				vtr: &MockVerificationTokenRepository{
					FindByHashAndTypeFunc: func(ctx context.Context, tokenHash string, tokenType domain.VerificationTokenType) (*domain.VerificationToken, error) {
						if tc.err != nil {
							return nil, tc.err
						}
						return tc.token, nil
					},
				},
			}

			err := service.ResetPasswordWithToken(context.Background(), "reset-token", "newpassword123")
			if err == nil || err.Error() != "invalid or expired token" {
				t.Fatalf("expected invalid or expired token error, got %v", err)
			}
		})
	}
}

func TestResendEmailVerification_SuccessAndAlreadyVerified(t *testing.T) {
	userID := uuid.New()
	tests := []struct {
		name       string
		user       domain.User
		wantErr    string
		wantCreate bool
	}{
		{
			name:       "success",
			user:       domain.User{ID: userID, Email: "jane@example.com"},
			wantCreate: true,
		},
		{
			name:    "already verified",
			user:    domain.User{ID: userID, EmailVerified: true},
			wantErr: "email already verified",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			created := 0
			invalidated := 0
			service := &userService{
				txm: &noopTransactionManager{},
				r: &MockUserRepository{
					FindByIDFunc: func(ctx context.Context, id uuid.UUID) (domain.User, error) {
						return tc.user, nil
					},
					GetPreferencesByUserIDFunc: func(ctx context.Context, userID uuid.UUID) (domain.UserPreferences, error) {
						return domain.UserPreferences{UserID: userID, UseSMS: tc.wantCreate}, nil
					},
				},
				vtr: &MockVerificationTokenRepository{
					InvalidateByUserAndTypeFunc: func(ctx context.Context, gotUserID uuid.UUID, tokenType domain.VerificationTokenType) error {
						invalidated++
						return nil
					},
					CreateFunc: func(ctx context.Context, token *domain.VerificationToken) error {
						created++
						if token.Type != domain.VerificationEmail {
							t.Fatalf("unexpected token type: %s", token.Type)
						}
						return nil
					},
				},
			}

			err := service.ResendEmailVerification(context.Background(), userID)
			if tc.wantErr == "" && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if tc.wantErr != "" && (err == nil || err.Error() != tc.wantErr) {
				t.Fatalf("expected error %q, got %v", tc.wantErr, err)
			}
			if tc.wantCreate && (created != 1 || invalidated != 1) {
				t.Fatalf("expected create and invalidate once, got create=%d invalidate=%d", created, invalidated)
			}
			if !tc.wantCreate && tc.wantErr == "" && (created != 0 || invalidated != 0) {
				t.Fatalf("expected no otp work when sms is disabled, got create=%d invalidate=%d", created, invalidated)
			}
		})
	}
}

func TestResendMobileOTP_SuccessAndAlreadyVerified(t *testing.T) {
	userID := uuid.New()
	tests := []struct {
		name       string
		user       domain.User
		wantErr    string
		wantCreate bool
	}{
		{
			name:       "success",
			user:       domain.User{ID: userID, Mobile: "+15555550123"},
			wantCreate: true,
		},
		{
			name:    "already verified",
			user:    domain.User{ID: userID, MobileVerified: true},
			wantErr: "mobile already verified",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			created := 0
			invalidated := 0
			service := &userService{
				txm: &noopTransactionManager{},
				r: &MockUserRepository{
					FindByIDFunc: func(ctx context.Context, id uuid.UUID) (domain.User, error) {
						return tc.user, nil
					},
					GetPreferencesByUserIDFunc: func(ctx context.Context, userID uuid.UUID) (domain.UserPreferences, error) {
						return domain.UserPreferences{UserID: userID, UseSMS: tc.wantCreate}, nil
					},
				},
				vtr: &MockVerificationTokenRepository{
					InvalidateByUserAndTypeFunc: func(ctx context.Context, gotUserID uuid.UUID, tokenType domain.VerificationTokenType) error {
						invalidated++
						return nil
					},
					CreateFunc: func(ctx context.Context, token *domain.VerificationToken) error {
						created++
						if token.Type != domain.VerificationMobile {
							t.Fatalf("unexpected token type: %s", token.Type)
						}
						return nil
					},
				},
			}

			err := service.ResendMobileOTP(context.Background(), userID)
			if tc.wantErr == "" && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if tc.wantErr != "" && (err == nil || err.Error() != tc.wantErr) {
				t.Fatalf("expected error %q, got %v", tc.wantErr, err)
			}
			if tc.wantCreate && (created != 1 || invalidated != 1) {
				t.Fatalf("expected create and invalidate once, got create=%d invalidate=%d", created, invalidated)
			}
		})
	}
}

func TestForgotPassword_AvoidsEnumerationWhenUserMissing(t *testing.T) {
	service := &userService{
		txm: &noopTransactionManager{},
		r: &MockUserRepository{
			FindByEmailFunc: func(ctx context.Context, email string) (domain.User, error) {
				return domain.User{}, errors.New("not found")
			},
		},
		vtr: &MockVerificationTokenRepository{},
	}

	token, err := service.ForgotPassword(context.Background(), domain.ForgotPasswordInput{Email: "missing@test.com"})
	if err != nil {
		t.Fatalf("expected no error for enumeration protection, got %v", err)
	}
	if token != "" {
		t.Fatalf("expected empty token when user missing")
	}
}

func TestResetPassword_UpdatesHashedPassword(t *testing.T) {
	userID := uuid.New()
	var updated *domain.User
	service := &userService{
		txm: &noopTransactionManager{},
		r: &MockUserRepository{
			FindByIDFunc: func(ctx context.Context, id uuid.UUID) (domain.User, error) {
				return domain.User{ID: id, Password: "old"}, nil
			},
			UpdateFunc: func(ctx context.Context, user *domain.User) error {
				updated = user
				return nil
			},
		},
	}

	err := service.ResetPassword(context.Background(), userID, "new-password")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if updated == nil || updated.Password == "" || updated.Password == "new-password" {
		t.Fatalf("expected hashed password to be persisted")
	}
}

func TestResetPasswordWithToken_MarksTokenUsed(t *testing.T) {
	userID := uuid.New()
	token := &domain.VerificationToken{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      domain.PasswordReset,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	updatedToken := false
	invalidated := false

	service := &userService{
		txm: &noopTransactionManager{},
		r: &MockUserRepository{
			FindByIDFunc: func(ctx context.Context, id uuid.UUID) (domain.User, error) {
				return domain.User{ID: id, Password: "old"}, nil
			},
			UpdateFunc: func(ctx context.Context, user *domain.User) error {
				return nil
			},
		},
		vtr: &MockVerificationTokenRepository{
			FindByHashAndTypeFunc: func(ctx context.Context, tokenHash string, tokenType domain.VerificationTokenType) (*domain.VerificationToken, error) {
				return token, nil
			},
			UpdateFunc: func(ctx context.Context, got *domain.VerificationToken) error {
				updatedToken = got.IsUsed && got.UsedAt != nil
				return nil
			},
			InvalidateByUserAndTypeFunc: func(ctx context.Context, gotUserID uuid.UUID, tokenType domain.VerificationTokenType) error {
				invalidated = true
				return nil
			},
		},
	}

	err := service.ResetPasswordWithToken(context.Background(), "raw-token", "new-password")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !updatedToken || !invalidated {
		t.Fatalf("expected token update + invalidation")
	}
}

func TestVerifyEmailAndMobile_UpdateUserFlags(t *testing.T) {
	userID := uuid.New()
	emailToken := &domain.VerificationToken{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      domain.VerificationEmail,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	mobileToken := &domain.VerificationToken{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      domain.VerificationMobile,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	emailVerified := false
	mobileVerified := false

	service := &userService{
		txm: &noopTransactionManager{},
		r: &MockUserRepository{
			FindByIDFunc: func(ctx context.Context, id uuid.UUID) (domain.User, error) {
				return domain.User{ID: id}, nil
			},
			UpdateFunc: func(ctx context.Context, user *domain.User) error {
				emailVerified = emailVerified || user.EmailVerified
				mobileVerified = mobileVerified || user.MobileVerified
				return nil
			},
		},
		vtr: &MockVerificationTokenRepository{
			FindByHashAndTypeFunc: func(ctx context.Context, tokenHash string, tokenType domain.VerificationTokenType) (*domain.VerificationToken, error) {
				if tokenType == domain.VerificationEmail {
					return emailToken, nil
				}
				return mobileToken, nil
			},
			UpdateFunc: func(ctx context.Context, token *domain.VerificationToken) error {
				return nil
			},
		},
	}

	if err := service.VerifyEmail(context.Background(), "email-raw"); err != nil {
		t.Fatalf("expected email verification success, got %v", err)
	}
	if err := service.VerifyMobile(context.Background(), "123456"); err != nil {
		t.Fatalf("expected mobile verification success, got %v", err)
	}
	if !emailVerified || !mobileVerified {
		t.Fatalf("expected both verification flags to be updated")
	}
}

func TestResendEmailVerification_CreatesNewToken(t *testing.T) {
	userID := uuid.New()
	created := 0

	service := &userService{
		txm: &noopTransactionManager{},
		r: &MockUserRepository{
			FindByIDFunc: func(ctx context.Context, id uuid.UUID) (domain.User, error) {
				return domain.User{ID: id, Email: "user@test.com"}, nil
			},
		},
		vtr: &MockVerificationTokenRepository{
			InvalidateByUserAndTypeFunc: func(ctx context.Context, id uuid.UUID, tokenType domain.VerificationTokenType) error {
				return nil
			},
			CreateFunc: func(ctx context.Context, token *domain.VerificationToken) error {
				created++
				if token.Type != domain.VerificationEmail {
					t.Fatalf("expected email token type")
				}
				return nil
			},
		},
	}

	err := service.ResendEmailVerification(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if created != 1 {
		t.Fatalf("expected one token creation, got %d", created)
	}
}

func TestResendMobileOTP_CreatesNewOTPToken(t *testing.T) {
	userID := uuid.New()
	created := 0

	service := &userService{
		txm: &noopTransactionManager{},
		r: &MockUserRepository{
			FindByIDFunc: func(ctx context.Context, id uuid.UUID) (domain.User, error) {
				return domain.User{ID: id, Mobile: "+911111111111"}, nil
			},
			GetPreferencesByUserIDFunc: func(ctx context.Context, userID uuid.UUID) (domain.UserPreferences, error) {
				return domain.UserPreferences{UserID: userID, UseSMS: true}, nil
			},
		},
		vtr: &MockVerificationTokenRepository{
			InvalidateByUserAndTypeFunc: func(ctx context.Context, id uuid.UUID, tokenType domain.VerificationTokenType) error {
				return nil
			},
			CreateFunc: func(ctx context.Context, token *domain.VerificationToken) error {
				created++
				if token.Type != domain.VerificationMobile {
					t.Fatalf("expected mobile token type")
				}
				return nil
			},
		},
	}

	err := service.ResendMobileOTP(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if created != 1 {
		t.Fatalf("expected one token creation, got %d", created)
	}
}

func TestResendMobileOTP_SkipsWhenSMSPreferenceDisabled(t *testing.T) {
	userID := uuid.New()
	created := 0
	invalidated := 0

	service := &userService{
		txm: &noopTransactionManager{},
		r: &MockUserRepository{
			FindByIDFunc: func(ctx context.Context, id uuid.UUID) (domain.User, error) {
				return domain.User{ID: id, Mobile: "+911111111111"}, nil
			},
			GetPreferencesByUserIDFunc: func(ctx context.Context, userID uuid.UUID) (domain.UserPreferences, error) {
				return domain.UserPreferences{UserID: userID, UseSMS: false}, nil
			},
		},
		vtr: &MockVerificationTokenRepository{
			InvalidateByUserAndTypeFunc: func(ctx context.Context, id uuid.UUID, tokenType domain.VerificationTokenType) error {
				invalidated++
				return nil
			},
			CreateFunc: func(ctx context.Context, token *domain.VerificationToken) error {
				created++
				return nil
			},
		},
	}

	err := service.ResendMobileOTP(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if created != 0 || invalidated != 0 {
		t.Fatalf("expected no otp work when sms is disabled, got create=%d invalidate=%d", created, invalidated)
	}
}

func TestLogin_InvalidatesAndStoresRefreshToken(t *testing.T) {
	t.Setenv("JWT_SECRET_V1", "access-secret")
	t.Setenv("JWT_REFRESH_V1", "refresh-secret")

	userID := uuid.New()
	password := "password123"
	updateCalled := false
	invalidated := false
	created := false

	service := &userService{
		txm: &noopTransactionManager{},
		r: &MockUserRepository{
			FindByEmailFunc: func(ctx context.Context, email string) (domain.User, error) {
				return domain.User{
					ID:       userID,
					Email:    email,
					Password: (&userService{}).hashPassword(password),
					Role:     domain.UserRoleUser,
				}, nil
			},
			UpdateFunc: func(ctx context.Context, user *domain.User) error {
				updateCalled = true
				if user.LastLogin == nil {
					t.Fatalf("expected last_login to be updated")
				}
				return nil
			},
		},
		vtr: &MockVerificationTokenRepository{
			FindByUserIDAndTypeFunc: func(ctx context.Context, id uuid.UUID, tokenType domain.VerificationTokenType) (*domain.VerificationToken, error) {
				if id != userID || tokenType != domain.RefreshToken {
					t.Fatalf("unexpected find refresh token args")
				}
				return &domain.VerificationToken{
					ID:     uuid.New(),
					UserID: userID,
					Type:   domain.RefreshToken,
				}, nil
			},
			InvalidateByUserAndTypeFunc: func(ctx context.Context, id uuid.UUID, tokenType domain.VerificationTokenType) error {
				invalidated = true
				if id != userID || tokenType != domain.RefreshToken {
					t.Fatalf("unexpected invalidation args")
				}
				return nil
			},
			CreateFunc: func(ctx context.Context, token *domain.VerificationToken) error {
				created = true
				if token.UserID != userID || token.Type != domain.RefreshToken {
					t.Fatalf("unexpected refresh token record")
				}
				if token.TokenHash == "" {
					t.Fatalf("expected hashed refresh token to be stored")
				}
				return nil
			},
		},
	}

	tokens, err := service.Login(context.Background(), domain.LoginInput{
		Email:    "user@test.com",
		Password: password,
	})
	if err != nil {
		t.Fatalf("expected login success, got %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("expected access and refresh tokens")
	}
	if !updateCalled || !invalidated || !created {
		t.Fatalf("expected update=%v invalidate=%v create=%v", updateCalled, invalidated, created)
	}
}

func TestLogin_ReturnsErrorWhenRefreshInvalidationFails(t *testing.T) {
	userID := uuid.New()
	password := "password123"
	createCalled := false

	service := &userService{
		txm: &noopTransactionManager{},
		r: &MockUserRepository{
			FindByEmailFunc: func(ctx context.Context, email string) (domain.User, error) {
				return domain.User{
					ID:       userID,
					Email:    email,
					Password: (&userService{}).hashPassword(password),
					Role:     domain.UserRoleUser,
				}, nil
			},
			UpdateFunc: func(ctx context.Context, user *domain.User) error { return nil },
		},
		vtr: &MockVerificationTokenRepository{
			FindByUserIDAndTypeFunc: func(ctx context.Context, id uuid.UUID, tokenType domain.VerificationTokenType) (*domain.VerificationToken, error) {
				return &domain.VerificationToken{
					ID:     uuid.New(),
					UserID: userID,
					Type:   domain.RefreshToken,
				}, nil
			},
			InvalidateByUserAndTypeFunc: func(ctx context.Context, id uuid.UUID, tokenType domain.VerificationTokenType) error {
				return errors.New("invalidation failed")
			},
			CreateFunc: func(ctx context.Context, token *domain.VerificationToken) error {
				createCalled = true
				return nil
			},
		},
	}

	_, err := service.Login(context.Background(), domain.LoginInput{
		Email:    "user@test.com",
		Password: password,
	})
	if err == nil {
		t.Fatalf("expected error when invalidation fails")
	}
	if createCalled {
		t.Fatalf("did not expect refresh token create when invalidation fails")
	}
}

func TestLogin_ReturnsErrorWhenRefreshCreateFails(t *testing.T) {
	userID := uuid.New()
	password := "password123"

	service := &userService{
		txm: &noopTransactionManager{},
		r: &MockUserRepository{
			FindByEmailFunc: func(ctx context.Context, email string) (domain.User, error) {
				return domain.User{
					ID:       userID,
					Email:    email,
					Password: (&userService{}).hashPassword(password),
					Role:     domain.UserRoleUser,
				}, nil
			},
			UpdateFunc: func(ctx context.Context, user *domain.User) error { return nil },
		},
		vtr: &MockVerificationTokenRepository{
			InvalidateByUserAndTypeFunc: func(ctx context.Context, id uuid.UUID, tokenType domain.VerificationTokenType) error {
				return nil
			},
			CreateFunc: func(ctx context.Context, token *domain.VerificationToken) error {
				return errors.New("create failed")
			},
		},
	}

	_, err := service.Login(context.Background(), domain.LoginInput{
		Email:    "user@test.com",
		Password: password,
	})
	if err == nil {
		t.Fatalf("expected error when refresh token create fails")
	}
}

func TestRefresh_Success_RotatesToken(t *testing.T) {
	t.Setenv("JWT_SECRET_V1", "access-secret")
	t.Setenv("JWT_REFRESH_V1", "refresh-secret")

	user := domain.User{
		ID:    uuid.New(),
		Email: "user@test.com",
		Role:  domain.UserRoleUser,
	}

	rawRefreshToken, err := (&userService{}).generateRefreshToken(user)
	if err != nil {
		t.Fatalf("failed to create refresh token for test: %v", err)
	}
	tokenHash := (&userService{}).generateTokenHash(rawRefreshToken)

	service := &userService{
		txm: &noopTransactionManager{},
		r: &MockUserRepository{
			FindByIDFunc: func(ctx context.Context, id uuid.UUID) (domain.User, error) {
				if id != user.ID {
					t.Fatalf("unexpected user id")
				}
				return user, nil
			},
		},
		vtr: &MockVerificationTokenRepository{
			FindByHashAndTypeFunc: func(ctx context.Context, hash string, tokenType domain.VerificationTokenType) (*domain.VerificationToken, error) {
				if hash != tokenHash || tokenType != domain.RefreshToken {
					t.Fatalf("unexpected hash lookup")
				}
				return &domain.VerificationToken{
					ID:        uuid.New(),
					UserID:    user.ID,
					Type:      domain.RefreshToken,
					TokenHash: hash,
					ExpiresAt: time.Now().Add(time.Hour),
				}, nil
			},
		},
	}

	accessToken, err := service.Refresh(context.Background(), rawRefreshToken)
	if err != nil {
		t.Fatalf("expected refresh success, got %v", err)
	}
	if accessToken == "" {
		t.Fatalf("expected access token")
	}
}

func TestRefresh_InvalidWhenStoredUserMismatch(t *testing.T) {
	t.Setenv("JWT_REFRESH_V1", "refresh-secret")

	user := domain.User{
		ID:    uuid.New(),
		Email: "user@test.com",
		Role:  domain.UserRoleUser,
	}
	rawRefreshToken, err := (&userService{}).generateRefreshToken(user)
	if err != nil {
		t.Fatalf("failed to create refresh token for test: %v", err)
	}

	service := &userService{
		txm: &noopTransactionManager{},
		r:   &MockUserRepository{},
		vtr: &MockVerificationTokenRepository{
			FindByHashAndTypeFunc: func(ctx context.Context, hash string, tokenType domain.VerificationTokenType) (*domain.VerificationToken, error) {
				return &domain.VerificationToken{
					ID:        uuid.New(),
					UserID:    uuid.New(),
					Type:      domain.RefreshToken,
					TokenHash: hash,
					ExpiresAt: time.Now().Add(time.Hour),
				}, nil
			},
		},
	}

	_, err = service.Refresh(context.Background(), rawRefreshToken)
	if err == nil || err.Error() != "invalid refresh token" {
		t.Fatalf("expected invalid refresh token error, got %v", err)
	}
}

func TestRefresh_InvalidWhenTokenMalformed(t *testing.T) {
	service := &userService{
		txm: &noopTransactionManager{},
		r:   &MockUserRepository{},
		vtr: &MockVerificationTokenRepository{},
	}

	_, err := service.Refresh(context.Background(), "not-a-jwt")
	if err == nil || err.Error() != "invalid refresh token" {
		t.Fatalf("expected invalid refresh token error, got %v", err)
	}
}
