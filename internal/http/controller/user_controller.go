package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"booknest/internal/domain"
)

type userController struct {
	service domain.UserService
}

// NewUserController creates a new user controller instance
func NewUserController(service domain.UserService) domain.UserController {
	return &userController{service: service}
}

// RegisterRoutes registers all user routes
func (c *userController) RegisterRoutes(r gin.IRouter) {
	RegisterUserRoutes(r, getJWTConfig(), c)
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account and sends email & mobile verification
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body  domain.UserInput  true  "User registration input"
// @Success      201  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /auth/register [post]
func (c *userController) Register(ctx *gin.Context) {

	var input domain.UserInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.Register(ctx, input); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully. Please verify your email and mobile.",
	})
}

// Login godoc
// @Summary      Login user
// @Description  Login using email or mobile and password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body  domain.LoginInput  true  "Login credentials"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /auth/login [post]
func (c *userController) Login(ctx *gin.Context) {
	var input domain.LoginInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate that either email or mobile is provided
	if input.Email == "" && input.Mobile == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "email or mobile is required"})
		return
	}

	tokens, err := c.service.Login(ctx, input)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"message":       "Login successful",
	})
}

// Refresh godoc
// @Summary      Refresh access token
// @Description  Exchanges a valid refresh token for a new access token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body  map[string]string  true  "Refresh token payload"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /auth/refresh [post]
func (c *userController) Refresh(ctx *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, err := c.service.Refresh(ctx, input.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
	})
}

// GetUser godoc
// @Summary      Get user by ID
// @Description  Fetch user details by user ID
// @Tags         Users
// @Produce      json
// @Param        id   path  string  true  "User ID"
// @Success      200  {object}  domain.User
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Security     BearerAuth
// @Router       /users/{id} [get]
func (c *userController) GetUser(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user, err := c.service.FindUser(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// DeleteUser godoc
// @Summary      Delete user account
// @Description  User can delete only their own account
// @Tags         Users
// @Produce      json
// @Param        id   path  string  true  "User ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /users/{id} [delete]
func (c *userController) DeleteUser(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	// Verify that the user can only delete their own account
	userIDFromCtx, err := getUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if userIDFromCtx != id {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "you can only delete your own account"})
		return
	}

	if err := c.service.DeleteUser(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// ForgotPassword godoc
// @Summary      Forgot password
// @Description  Sends password reset link or OTP to email/mobile
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body  domain.ForgotPasswordInput  true  "Forgot password input"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Router       /auth/forgot-password [post]
func (c *userController) ForgotPassword(ctx *gin.Context) {
	var input domain.ForgotPasswordInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Email == "" && input.Mobile == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "email or mobile is required"})
		return
	}

	resetToken, err := c.service.ForgotPassword(ctx, input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":     "If the account exists, a password reset link has been sent",
		"reset_token": resetToken,
	})
}

// ResetPasswordWithToken godoc
// @Summary      Reset password with token
// @Description  Reset password using forgot-password reset token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body  map[string]string  true  "Token and new password"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /auth/reset-password/confirm [post]
func (c *userController) ResetPasswordWithToken(ctx *gin.Context) {
	var input struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.ResetPasswordWithToken(ctx, input.Token, input.NewPassword); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}

// VerifyEmail godoc
// @Summary      Verify email
// @Description  Verifies email using token
// @Tags         Verification
// @Accept       json
// @Produce      json
// @Param        payload  body  map[string]string  true  "Verification token"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Router       /auth/verify-email [post]
func (c *userController) VerifyEmail(ctx *gin.Context) {
	var input struct {
		Token string `json:"token" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.VerifyEmail(ctx, input.Token); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Email verified successfully",
	})
}

// VerifyMobile godoc
// @Summary      Verify mobile
// @Description  Verifies mobile using OTP
// @Tags         Verification
// @Accept       json
// @Produce      json
// @Param        payload  body  map[string]string  true  "OTP"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Router       /auth/verify-mobile [post]
func (c *userController) VerifyMobile(ctx *gin.Context) {
	var input struct {
		OTP string `json:"otp" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.VerifyMobile(ctx, input.OTP); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired OTP"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Mobile verified successfully",
	})
}

// ResendEmailVerification godoc
// @Summary      Resend email verification
// @Description  Resends verification email if not verified
// @Tags         Verification
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/resend-email-verification [post]
func (c *userController) ResendEmailVerification(ctx *gin.Context) {
	userIDFromCtx, err := getUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := c.service.ResendEmailVerification(ctx, userIDFromCtx); err != nil {
		if errors.Is(err, errors.New("email already verified")) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "email already verified"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Verification email sent successfully",
	})
}

// ResendMobileOTP godoc
// @Summary      Resend mobile OTP
// @Description  Resends mobile OTP if not verified
// @Tags         Verification
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/resend-mobile-otp [post]
func (c *userController) ResendMobileOTP(ctx *gin.Context) {
	userIDFromCtx, err := getUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := c.service.ResendMobileOTP(ctx, userIDFromCtx); err != nil {
		if errors.Is(err, errors.New("mobile already verified")) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "mobile already verified"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Mobile OTP sent successfully",
	})
}

// ResetPassword godoc
// @Summary      Reset password
// @Description  Reset password for authenticated user
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body  map[string]string  true  "New password"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/reset-password [post]
func (c *userController) ResetPassword(ctx *gin.Context) {
	userIDFromCtx, err := getUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input struct {
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.ResetPassword(ctx, userIDFromCtx, input.NewPassword); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}
