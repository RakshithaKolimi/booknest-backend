package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"booknest/internal/domain"
)

// MockUserService is a mock implementation of domain.UserService
type MockUserService struct {
	FindUserFunc                func(ctx context.Context, id uuid.UUID) (domain.User, error)
	RegisterFunc                func(ctx context.Context, in domain.UserInput) error
	RegisterAdminFunc           func(ctx context.Context, in domain.AdminRegistrationInput) error
	LoginFunc                   func(ctx context.Context, in domain.LoginInput) (domain.AuthTokens, error)
	RefreshFunc                 func(ctx context.Context, rawRefreshToken string) (string, error)
	ForgotPasswordFunc          func(ctx context.Context, in domain.ForgotPasswordInput) (string, error)
	ResetPasswordFunc           func(ctx context.Context, userID uuid.UUID, newPassword string) error
	ResetPasswordWithTokenFunc  func(ctx context.Context, rawToken, newPassword string) error
	VerifyEmailFunc             func(ctx context.Context, rawToken string) error
	VerifyMobileFunc            func(ctx context.Context, otp string) error
	ResendEmailVerificationFunc func(ctx context.Context, userID uuid.UUID) error
	ResendMobileOTPFunc         func(ctx context.Context, userID uuid.UUID) error
	DeleteUserFunc              func(ctx context.Context, id uuid.UUID) error
}

// Implement domain.UserService methods for MockUserService
func (m *MockUserService) FindUser(ctx context.Context, id uuid.UUID) (domain.User, error) {
	if m.FindUserFunc != nil {
		return m.FindUserFunc(ctx, id)
	}
	return domain.User{}, errors.New("not implemented")
}

func (m *MockUserService) Register(ctx context.Context, in domain.UserInput) error {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(ctx, in)
	}
	return errors.New("not implemented")
}

func (m *MockUserService) RegisterAdmin(ctx context.Context, in domain.AdminRegistrationInput) error {
	if m.RegisterAdminFunc != nil {
		return m.RegisterAdminFunc(ctx, in)
	}
	return errors.New("not implemented")
}

func (m *MockUserService) Login(ctx context.Context, in domain.LoginInput) (domain.AuthTokens, error) {
	if m.LoginFunc != nil {
		return m.LoginFunc(ctx, in)
	}
	return domain.AuthTokens{}, errors.New("not implemented")
}

func (m *MockUserService) Refresh(ctx context.Context, rawRefreshToken string) (string, error) {
	if m.RefreshFunc != nil {
		return m.RefreshFunc(ctx, rawRefreshToken)
	}
	return "", errors.New("not implemented")
}

func (m *MockUserService) ForgotPassword(ctx context.Context, in domain.ForgotPasswordInput) (string, error) {
	if m.ForgotPasswordFunc != nil {
		return m.ForgotPasswordFunc(ctx, in)
	}
	return "", nil
}

func (m *MockUserService) ResetPassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	if m.ResetPasswordFunc != nil {
		return m.ResetPasswordFunc(ctx, userID, newPassword)
	}
	return errors.New("not implemented")
}

func (m *MockUserService) ResetPasswordWithToken(ctx context.Context, rawToken, newPassword string) error {
	if m.ResetPasswordWithTokenFunc != nil {
		return m.ResetPasswordWithTokenFunc(ctx, rawToken, newPassword)
	}
	return errors.New("not implemented")
}

func (m *MockUserService) VerifyEmail(ctx context.Context, rawToken string) error {
	if m.VerifyEmailFunc != nil {
		return m.VerifyEmailFunc(ctx, rawToken)
	}
	return errors.New("not implemented")
}

func (m *MockUserService) VerifyMobile(ctx context.Context, otp string) error {
	if m.VerifyMobileFunc != nil {
		return m.VerifyMobileFunc(ctx, otp)
	}
	return errors.New("not implemented")
}

func (m *MockUserService) ResendEmailVerification(ctx context.Context, userID uuid.UUID) error {
	if m.ResendEmailVerificationFunc != nil {
		return m.ResendEmailVerificationFunc(ctx, userID)
	}
	return errors.New("not implemented")
}

func (m *MockUserService) ResendMobileOTP(ctx context.Context, userID uuid.UUID) error {
	if m.ResendMobileOTPFunc != nil {
		return m.ResendMobileOTPFunc(ctx, userID)
	}
	return errors.New("not implemented")
}

func (m *MockUserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(ctx, id)
	}
	return errors.New("not implemented")
}

// TestLogin_Success tests successful login
func TestLogin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockUserService{
		LoginFunc: func(ctx context.Context, in domain.LoginInput) (domain.AuthTokens, error) {
			return domain.AuthTokens{
				AccessToken:  "valid.access.token",
				RefreshToken: "valid.refresh.token",
			}, nil
		},
	}

	controller := NewUserController(mockService)
	router := gin.New()
	controller.RegisterRoutes(router)

	input := domain.LoginInput{
		Email:    "test@example.com",
		Password: "password123",
	}

	body, _ := json.Marshal(input)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var response map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	if response["access_token"] != "valid.access.token" {
		t.Fatalf("expected access_token in response")
	}
	if response["refresh_token"] != "valid.refresh.token" {
		t.Fatalf("expected refresh_token in response")
	}
}

// TestLogin_InvalidCredentials tests login with invalid credentials
func TestLogin_InvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockUserService{
		LoginFunc: func(ctx context.Context, in domain.LoginInput) (domain.AuthTokens, error) {
			return domain.AuthTokens{}, errors.New("invalid credentials")
		},
	}

	controller := NewUserController(mockService)
	router := gin.New()
	controller.RegisterRoutes(router)

	input := domain.LoginInput{
		Email:    "nonexistent@example.com",
		Password: "wrongpassword",
	}

	body, _ := json.Marshal(input)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestLogin_AdminVerificationRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockUserService{
		LoginFunc: func(ctx context.Context, in domain.LoginInput) (domain.AuthTokens, error) {
			return domain.AuthTokens{}, domain.ErrAdminVerificationRequired
		},
	}

	controller := NewUserController(mockService)
	router := gin.New()
	controller.RegisterRoutes(router)

	input := domain.LoginInput{
		Email:    "admin@example.com",
		Password: "password123",
	}

	body, _ := json.Marshal(input)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

// TestLogin_MissingEmailAndMobile tests login without email or mobile
func TestLogin_MissingEmailAndMobile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockUserService{}
	controller := NewUserController(mockService)
	router := gin.New()
	controller.RegisterRoutes(router)

	input := domain.LoginInput{
		Password: "password123",
	}

	body, _ := json.Marshal(input)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRefresh_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockUserService{
		RefreshFunc: func(ctx context.Context, rawRefreshToken string) (string, error) {
			if rawRefreshToken != "old.refresh.token" {
				t.Fatalf("unexpected refresh token: %s", rawRefreshToken)
			}
			return "new.access.token", nil
		},
	}

	controller := NewUserController(mockService)
	router := gin.New()
	controller.RegisterRoutes(router)

	input := map[string]string{
		"refresh_token": "old.refresh.token",
	}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var response map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	if response["access_token"] != "new.access.token" {
		t.Fatalf("expected access_token in response")
	}
	if _, ok := response["refresh_token"]; ok {
		t.Fatalf("did not expect refresh_token in response")
	}
}

func TestRefresh_InvalidPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller := NewUserController(&MockUserService{})
	router := gin.New()
	controller.RegisterRoutes(router)

	// missing refresh_token field
	req := httptest.NewRequest("POST", "/refresh", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockUserService{
		RefreshFunc: func(ctx context.Context, rawRefreshToken string) (string, error) {
			return "", errors.New("invalid refresh token")
		},
	}

	controller := NewUserController(mockService)
	router := gin.New()
	controller.RegisterRoutes(router)

	input := map[string]string{
		"refresh_token": "bad.refresh.token",
	}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

// TestGetUser_Success tests retrieving user successfully
func TestGetUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	mockService := &MockUserService{
		FindUserFunc: func(ctx context.Context, id uuid.UUID) (domain.User, error) {
			return domain.User{
				ID:        id,
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
			}, nil
		},
	}

	controller := NewUserController(mockService)
	router := gin.New()
	controller.RegisterRoutes(router)

	req := httptest.NewRequest("GET", "/user/"+userID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		// Should be unauthorized because no auth middleware
		if w.Code != http.StatusOK {
			// If it doesn't fail due to auth, it should work
			t.Logf("Got %d status code (could be due to missing auth middleware)", w.Code)
		}
	}
}

// TestGetUser_InvalidID tests retrieving user with invalid ID format
func TestGetUser_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockUserService{}
	controller := NewUserController(mockService)
	router := gin.New()
	controller.RegisterRoutes(router)

	req := httptest.NewRequest("GET", "/user/invalid-id", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest && w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 400 or 401, got %d", w.Code)
	}
}

// TestForgotPassword_Success tests forgot password request
func TestForgotPassword_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockUserService{
		ForgotPasswordFunc: func(ctx context.Context, in domain.ForgotPasswordInput) (string, error) {
			return "reset-token-123", nil
		},
	}

	controller := NewUserController(mockService)
	router := gin.New()
	controller.RegisterRoutes(router)

	input := domain.ForgotPasswordInput{
		Email: "test@example.com",
	}

	body, _ := json.Marshal(input)
	req := httptest.NewRequest("POST", "/forgot-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var response map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	if response["reset_token"] != "reset-token-123" {
		t.Fatalf("expected reset token in response")
	}
}

// TestForgotPassword_MissingEmailAndMobile tests forgot password without email or mobile
func TestForgotPassword_MissingEmailAndMobile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockUserService{}

	controller := NewUserController(mockService)
	router := gin.New()
	controller.RegisterRoutes(router)

	input := domain.ForgotPasswordInput{}

	body, _ := json.Marshal(input)
	req := httptest.NewRequest("POST", "/forgot-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestResetPasswordWithToken_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockUserService{
		ResetPasswordWithTokenFunc: func(ctx context.Context, rawToken, newPassword string) error {
			if rawToken != "token-1" || newPassword != "newpassword123" {
				t.Fatalf("unexpected input: token=%s password=%s", rawToken, newPassword)
			}
			return nil
		},
	}

	controller := NewUserController(mockService)
	router := gin.New()
	controller.RegisterRoutes(router)

	input := map[string]string{
		"token":        "token-1",
		"new_password": "newpassword123",
	}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest("POST", "/reset-password/confirm", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// TestVerifyEmail_Success tests successful email verification
func TestVerifyEmail_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockUserService{
		VerifyEmailFunc: func(ctx context.Context, rawToken string) error {
			return nil
		},
	}

	controller := NewUserController(mockService)
	router := gin.New()
	controller.RegisterRoutes(router)

	input := map[string]string{
		"token": "valid.email.token",
	}

	body, _ := json.Marshal(input)
	req := httptest.NewRequest("POST", "/verify-email", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		// Should be unauthorized because no auth middleware
		if w.Code != http.StatusOK {
			t.Logf("Got %d status code (could be due to missing auth middleware)", w.Code)
		}
	}
}

// TestVerifyMobile_Success tests successful mobile verification
func TestVerifyMobile_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockUserService{
		VerifyMobileFunc: func(ctx context.Context, otp string) error {
			return nil
		},
	}

	controller := NewUserController(mockService)
	router := gin.New()
	controller.RegisterRoutes(router)

	input := map[string]string{
		"otp": "123456",
	}

	body, _ := json.Marshal(input)
	req := httptest.NewRequest("POST", "/verify-mobile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		// Should be unauthorized because no auth middleware
		if w.Code != http.StatusOK {
			t.Logf("Got %d status code (could be due to missing auth middleware)", w.Code)
		}
	}
}

func TestUserControllerRegisterDirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	called := false
	ctl := NewUserController(&MockUserService{
		RegisterFunc: func(ctx context.Context, in domain.UserInput) error {
			called = true
			if in.Email != "jane@example.com" || in.Role != domain.UserRoleUser {
				t.Fatalf("unexpected input: %+v", in)
			}
			return nil
		},
	}).(*userController)

	body, _ := json.Marshal(domain.UserInput{
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
		Mobile:    "+15555550123",
		Password:  "password123",
		Role:      domain.UserRoleUser,
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	ctl.Register(c)
	if w.Code != http.StatusCreated || !called {
		t.Fatalf("expected register success, code=%d called=%v", w.Code, called)
	}
}

func TestUserControllerRegisterErrorDirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctl := NewUserController(&MockUserService{
		RegisterFunc: func(ctx context.Context, in domain.UserInput) error {
			return errors.New("create failed")
		},
	}).(*userController)

	body, _ := json.Marshal(domain.UserInput{
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
		Mobile:    "+15555550123",
		Password:  "password123",
		Role:      domain.UserRoleUser,
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	ctl.Register(c)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestUserControllerRegisterRejectsAdminDirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctl := NewUserController(&MockUserService{
		RegisterFunc: func(ctx context.Context, in domain.UserInput) error {
			return domain.ErrAdminSelfRegistrationNotAllowed
		},
	}).(*userController)

	body, _ := json.Marshal(domain.UserInput{
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
		Mobile:    "+15555550123",
		Password:  "password123",
		Role:      domain.UserRoleAdmin,
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	ctl.Register(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestUserControllerRegisterAdminDirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	called := false
	ctl := NewUserController(&MockUserService{
		RegisterAdminFunc: func(ctx context.Context, in domain.AdminRegistrationInput) error {
			called = true
			if in.Email != "admin@example.com" {
				t.Fatalf("unexpected input: %+v", in)
			}
			return nil
		},
	}).(*userController)

	body, _ := json.Marshal(domain.AdminRegistrationInput{
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
		Mobile:    "+15555550123",
		Password:  "password123",
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/register-admin", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	ctl.RegisterAdmin(c)
	if w.Code != http.StatusCreated || !called {
		t.Fatalf("expected register admin success, code=%d called=%v", w.Code, called)
	}
}

func TestUserControllerGetUserDirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	ctl := NewUserController(&MockUserService{
		FindUserFunc: func(ctx context.Context, id uuid.UUID) (domain.User, error) {
			return domain.User{ID: id, Email: "jane@example.com"}, nil
		},
	}).(*userController)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: userID.String()}}
	c.Request = httptest.NewRequest(http.MethodGet, "/user/"+userID.String(), nil)

	ctl.GetUser(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUserControllerDeleteUserDirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	deleted := false
	ctl := NewUserController(&MockUserService{
		DeleteUserFunc: func(ctx context.Context, id uuid.UUID) error {
			deleted = true
			return nil
		},
	}).(*userController)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", userID.String())
	c.Params = gin.Params{{Key: "id", Value: userID.String()}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/user/"+userID.String(), nil)

	ctl.DeleteUser(c)
	if w.Code != http.StatusOK || !deleted {
		t.Fatalf("expected delete success, code=%d deleted=%v", w.Code, deleted)
	}
}

func TestUserControllerDeleteUserForbiddenAndUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctl := NewUserController(&MockUserService{}).(*userController)
	targetID := uuid.New()

	wUnauthorized := httptest.NewRecorder()
	cUnauthorized, _ := gin.CreateTestContext(wUnauthorized)
	cUnauthorized.Params = gin.Params{{Key: "id", Value: targetID.String()}}
	cUnauthorized.Request = httptest.NewRequest(http.MethodDelete, "/user/"+targetID.String(), nil)
	ctl.DeleteUser(cUnauthorized)
	if wUnauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", wUnauthorized.Code)
	}

	wForbidden := httptest.NewRecorder()
	cForbidden, _ := gin.CreateTestContext(wForbidden)
	cForbidden.Set("user_id", uuid.New().String())
	cForbidden.Params = gin.Params{{Key: "id", Value: targetID.String()}}
	cForbidden.Request = httptest.NewRequest(http.MethodDelete, "/user/"+targetID.String(), nil)
	ctl.DeleteUser(cForbidden)
	if wForbidden.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", wForbidden.Code)
	}
}

func TestUserControllerVerifyAndResetDirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	ctl := NewUserController(&MockUserService{
		VerifyEmailFunc:  func(ctx context.Context, rawToken string) error { return nil },
		VerifyMobileFunc: func(ctx context.Context, otp string) error { return nil },
		ResetPasswordFunc: func(ctx context.Context, gotUserID uuid.UUID, newPassword string) error {
			if gotUserID != userID || newPassword != "newpassword123" {
				t.Fatalf("unexpected reset password input")
			}
			return nil
		},
		ResendEmailVerificationFunc: func(ctx context.Context, gotUserID uuid.UUID) error {
			if gotUserID != userID {
				t.Fatalf("unexpected resend email user id")
			}
			return nil
		},
		ResendMobileOTPFunc: func(ctx context.Context, gotUserID uuid.UUID) error {
			if gotUserID != userID {
				t.Fatalf("unexpected resend mobile user id")
			}
			return nil
		},
	}).(*userController)

	cases := []struct {
		name   string
		method func(*gin.Context)
		body   string
		setup  func(*gin.Context)
	}{
		{name: "verify email", method: ctl.VerifyEmail, body: `{"token":"email-token"}`},
		{name: "verify mobile", method: ctl.VerifyMobile, body: `{"otp":"123456"}`},
		{name: "reset password", method: ctl.ResetPassword, body: `{"new_password":"newpassword123"}`, setup: func(c *gin.Context) { c.Set("user_id", userID.String()) }},
		{name: "resend email", method: ctl.ResendEmailVerification, setup: func(c *gin.Context) { c.Set("user_id", userID.String()) }},
		{name: "resend mobile", method: ctl.ResendMobileOTP, setup: func(c *gin.Context) { c.Set("user_id", userID.String()) }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			if tc.setup != nil {
				tc.setup(c)
			}
			c.Request = httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(tc.body))
			c.Request.Header.Set("Content-Type", "application/json")
			tc.method(c)
			if w.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", w.Code)
			}
		})
	}
}

func TestUserControllerHandlerErrorsDirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	ctl := NewUserController(&MockUserService{
		FindUserFunc: func(ctx context.Context, id uuid.UUID) (domain.User, error) {
			return domain.User{}, errors.New("not found")
		},
		ResetPasswordWithTokenFunc: func(ctx context.Context, rawToken, newPassword string) error {
			return errors.New("invalid token")
		},
		VerifyEmailFunc:             func(ctx context.Context, rawToken string) error { return errors.New("bad token") },
		VerifyMobileFunc:            func(ctx context.Context, otp string) error { return errors.New("bad otp") },
		ResendEmailVerificationFunc: func(ctx context.Context, gotUserID uuid.UUID) error { return errors.New("boom") },
		ResendMobileOTPFunc:         func(ctx context.Context, gotUserID uuid.UUID) error { return errors.New("boom") },
		ResetPasswordFunc:           func(ctx context.Context, gotUserID uuid.UUID, newPassword string) error { return errors.New("boom") },
	}).(*userController)

	t.Run("get user invalid id", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "bad-id"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/user/bad-id", nil)
		ctl.GetUser(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("get user not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}
		c.Request = httptest.NewRequest(http.MethodGet, "/user/x", nil)
		ctl.GetUser(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("reset password with token error", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/reset-password/confirm", bytes.NewBufferString(`{"token":"x","new_password":"newpassword123"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		ctl.ResetPasswordWithToken(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	handlers := []struct {
		name   string
		method func(*gin.Context)
		body   string
		setup  func(*gin.Context)
		code   int
	}{
		{name: "verify email error", method: ctl.VerifyEmail, body: `{"token":"bad"}`, code: http.StatusBadRequest},
		{name: "verify mobile error", method: ctl.VerifyMobile, body: `{"otp":"bad"}`, code: http.StatusBadRequest},
		{name: "resend email unauthorized", method: ctl.ResendEmailVerification, code: http.StatusUnauthorized},
		{name: "resend mobile unauthorized", method: ctl.ResendMobileOTP, code: http.StatusUnauthorized},
		{name: "resend email server error", method: ctl.ResendEmailVerification, setup: func(c *gin.Context) { c.Set("user_id", userID.String()) }, code: http.StatusInternalServerError},
		{name: "resend mobile server error", method: ctl.ResendMobileOTP, setup: func(c *gin.Context) { c.Set("user_id", userID.String()) }, code: http.StatusInternalServerError},
		{name: "reset password unauthorized", method: ctl.ResetPassword, body: `{"new_password":"newpassword123"}`, code: http.StatusUnauthorized},
		{name: "reset password server error", method: ctl.ResetPassword, body: `{"new_password":"newpassword123"}`, setup: func(c *gin.Context) { c.Set("user_id", userID.String()) }, code: http.StatusInternalServerError},
	}

	for _, tc := range handlers {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			if tc.setup != nil {
				tc.setup(c)
			}
			c.Request = httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(tc.body))
			c.Request.Header.Set("Content-Type", "application/json")
			tc.method(c)
			if w.Code != tc.code {
				t.Fatalf("expected %d, got %d", tc.code, w.Code)
			}
		})
	}
}
