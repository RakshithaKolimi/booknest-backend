package user_service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"booknest/internal/domain"
)

type userService struct {
	txm domain.TransactionManager
	r   domain.UserRepository
	vtr domain.VerificationTokenRepository
}

func NewUserService(txm domain.TransactionManager, r domain.UserRepository, vtr domain.VerificationTokenRepository) domain.UserService {
	return &userService{
		txm: txm,
		r:   r,
		vtr: vtr,
	}
}

func (s *userService) FindUser(
	ctx context.Context,
	id uuid.UUID,
) (domain.User, error) {
	return s.r.FindByID(ctx, id)
}

func (s *userService) Register(
	ctx context.Context,
	in domain.UserInput,
) error {
	var emailToken *domain.VerificationToken
	var mobileOTP string

	// Use transaction for user registration
	err := s.txm.InTransaction(ctx, func(txCtx context.Context) error {
		// Create an user domain
		user := &domain.User{
			ID:        uuid.New(),
			FirstName: in.FirstName,
			LastName:  in.LastName,
			Email:     in.Email,
			Mobile:    in.Mobile,
			Password:  s.hashPassword(in.Password),
			Role:      in.Role,
			IsActive:  true,
		}

		// Create the user
		if err := s.r.Create(txCtx, user); err != nil {
			return err
		}

		// Create Verification code to send the email
		verificationCode, err := s.generateRawToken()
		if err != nil {
			return err
		}
		// Create email verification token
		token := &domain.VerificationToken{
			UserID:    user.ID,
			Type:      domain.VerificationEmail,
			TokenHash: s.generateTokenHash(verificationCode),
			ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hours expiry
		}

		// Add token to the database
		err = s.vtr.Create(txCtx, token)
		if err != nil {
			return err
		}
		// Assign to outer variable for sending email after transaction
		emailToken = token

		// Create mobile OTP
		mobileOTP = s.generateOTP(6)
		mobileToken := &domain.VerificationToken{
			UserID:    user.ID,
			Type:      domain.VerificationMobile,
			TokenHash: s.generateTokenHash(mobileOTP),
			ExpiresAt: time.Now().Add(5 * time.Minute), // 5 minutes expiry
		}

		// Add mobile token to the database
		return s.vtr.Create(txCtx, mobileToken)
	})

	if err != nil {
		return err
	}

	go func() {
		s.sendEmailVerification(in.Email, *emailToken)
	}()

	go func() {
		s.sendMobileVerification(in.Mobile, mobileOTP)
	}()

	return err
}

func (s *userService) Login(
	ctx context.Context,
	in domain.LoginInput,
) (domain.AuthTokens, error) {

	var user domain.User
	var err error

	// Get user by email or mobile
	if in.Email != "" {
		user, err = s.r.FindByEmail(ctx, in.Email)
	} else {
		user, err = s.r.FindByMobile(ctx, in.Mobile)
	}
	if err != nil {
		return domain.AuthTokens{}, err
	}

	// Validate the password
	if !s.comparePassword(user.Password, in.Password) {
		return domain.AuthTokens{}, errors.New("invalid credentials")
	}

	var tokens domain.AuthTokens
	err = s.txm.InTransaction(ctx, func(txCtx context.Context) error {
		// Update last login
		now := time.Now()
		user.LastLogin = &now
		if err := s.r.Update(txCtx, &user); err != nil {
			return err
		}

		// Rotate old refresh tokens on every fresh login.
		if err := s.vtr.InvalidateByUserAndType(txCtx, user.ID, domain.RefreshToken); err != nil {
			return err
		}

		// Get the access token
		accessToken, err := s.generateJWT(user)
		if err != nil {
			return err
		}

		// Get refresh token
		refreshToken, err := s.generateRefreshToken(user)
		if err != nil {
			return err
		}

		// Add refresh token to verification token table
		refreshTokenRecord := &domain.VerificationToken{
			UserID:    user.ID,
			Type:      domain.RefreshToken,
			TokenHash: s.generateTokenHash(refreshToken),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}
		if err := s.vtr.Create(txCtx, refreshTokenRecord); err != nil {
			return err
		}

		// Return the Tokens
		tokens = domain.AuthTokens{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}
		return nil
	})
	if err != nil {
		return domain.AuthTokens{}, err
	}
	return tokens, nil
}

func (s *userService) Refresh(
	ctx context.Context,
	rawRefreshToken string,
) (string, error) {
	// Verify the refresh token
	claims, err := s.verifyRefreshJWT(rawRefreshToken)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}

	// get user ID
	userID, err := s.userIDFromClaims(claims)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}

	// Generate Token
	tokenHash := s.generateTokenHash(rawRefreshToken)
	var accessToken string

	err = s.txm.InTransaction(ctx, func(txCtx context.Context) error {
		// Get the token from hash
		storedToken, err := s.vtr.FindByHashAndType(txCtx, tokenHash, domain.RefreshToken)
		if err != nil {
			return errors.New("invalid refresh token")
		}

		// If user ID mismatches, return error
		if storedToken.UserID != userID {
			return errors.New("invalid refresh token")
		}

		// Find the user by ID
		user, err := s.r.FindByID(txCtx, userID)
		if err != nil {
			return err
		}

		// Generate a new access token
		generatedAccessToken, err := s.generateJWT(user)
		if err != nil {
			return err
		}
		accessToken = generatedAccessToken
		return nil
	})
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func (s *userService) ForgotPassword(
	ctx context.Context,
	in domain.ForgotPasswordInput,
) (string, error) {
	var rawToken string

	err := s.txm.InTransaction(ctx, func(txCtx context.Context) error {
		var user domain.User
		var err error

		if in.Email != "" {
			user, err = s.r.FindByEmail(txCtx, in.Email)
		} else {
			user, err = s.r.FindByMobile(txCtx, in.Mobile)
		}
		if err != nil {
			// Avoid user enumeration in forgot-password flows.
			return nil
		}

		if err := s.vtr.InvalidateByUserAndType(txCtx, user.ID, domain.PasswordReset); err != nil {
			return err
		}

		rawToken, err = s.generateRawToken()
		if err != nil {
			return err
		}

		token := &domain.VerificationToken{
			UserID:    user.ID,
			Type:      domain.PasswordReset,
			TokenHash: s.generateTokenHash(rawToken),
			ExpiresAt: time.Now().Add(30 * time.Minute),
		}

		return s.vtr.Create(txCtx, token)
	})
	if err != nil {
		return "", err
	}

	return rawToken, nil
}

func (s *userService) ResetPassword(
	ctx context.Context,
	userID uuid.UUID,
	newPassword string,
) error {

	return s.txm.InTransaction(ctx, func(txCtx context.Context) error {
		// Get user by ID
		user, err := s.r.FindByID(txCtx, userID)
		if err != nil {
			return err
		}

		// Add hashed password to the user
		user.Password = s.hashPassword(newPassword)

		// Update the user
		return s.r.Update(txCtx, &user)
	})
}

func (s *userService) ResetPasswordWithToken(
	ctx context.Context,
	rawToken string,
	newPassword string,
) error {
	tokenHash := s.generateTokenHash(rawToken)

	return s.txm.InTransaction(ctx, func(txCtx context.Context) error {
		token, err := s.vtr.FindByHashAndType(txCtx, tokenHash, domain.PasswordReset)
		if err != nil {
			return errors.New("invalid or expired token")
		}
		if token.IsUsed || time.Now().After(token.ExpiresAt) {
			return errors.New("invalid or expired token")
		}

		user, err := s.r.FindByID(txCtx, token.UserID)
		if err != nil {
			return err
		}

		user.Password = s.hashPassword(newPassword)
		if err := s.r.Update(txCtx, &user); err != nil {
			return err
		}

		now := time.Now()
		token.IsUsed = true
		token.UsedAt = &now
		if err := s.vtr.Update(txCtx, token); err != nil {
			return err
		}

		return s.vtr.InvalidateByUserAndType(txCtx, user.ID, domain.PasswordReset)
	})
}

func (s *userService) VerifyEmail(
	ctx context.Context,
	rawToken string,
) error {
	return s.verifyToken(ctx, rawToken, domain.VerificationEmail)
}

func (s *userService) VerifyMobile(
	ctx context.Context,
	otp string,
) error {
	return s.verifyToken(ctx, otp, domain.VerificationMobile)
}

func (s *userService) ResendEmailVerification(
	ctx context.Context,
	userID uuid.UUID,
) error {

	var emailToken *domain.VerificationToken
	var email string

	err := s.txm.InTransaction(ctx, func(txCtx context.Context) error {
		// Fetch user
		user, err := s.r.FindByID(txCtx, userID)
		if err != nil {
			return err
		}

		// Check if already verified
		if user.EmailVerified {
			return errors.New("email already verified")
		}

		// Invalidate old tokens
		if err := s.vtr.InvalidateByUserAndType(
			txCtx,
			user.ID,
			domain.VerificationEmail,
		); err != nil {
			return err
		}

		// Create new verification token
		rawToken, err := s.generateRawToken()
		if err != nil {
			return err
		}

		// Create email verification token
		token := &domain.VerificationToken{
			UserID:    user.ID,
			Type:      domain.VerificationEmail,
			TokenHash: s.generateTokenHash(rawToken),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		// Add token to the database
		if err := s.vtr.Create(txCtx, token); err != nil {
			return err
		}

		// Assign to outer variables for sending email after transaction
		emailToken = token
		email = user.Email
		return nil
	})

	if err != nil {
		return err
	}

	// Send verification email
	go s.sendEmailVerification(email, *emailToken)
	return nil
}

func (s *userService) ResendMobileOTP(
	ctx context.Context,
	userID uuid.UUID,
) error {

	var otp string
	var mobile string

	err := s.txm.InTransaction(ctx, func(txCtx context.Context) error {
		// Fetch user
		user, err := s.r.FindByID(txCtx, userID)
		if err != nil {
			return err
		}

		// Check if already verified
		if user.MobileVerified {
			return errors.New("mobile already verified")
		}

		// Invalidate old OTPs
		if err := s.vtr.InvalidateByUserAndType(
			txCtx,
			user.ID,
			domain.VerificationMobile, // or LoginOTP if you haven't split yet
		); err != nil {
			return err
		}

		// Generate new OTP
		otp = s.generateOTP(6)

		// Create mobile verification token
		token := &domain.VerificationToken{
			UserID:    user.ID,
			Type:      domain.VerificationMobile,
			TokenHash: s.generateTokenHash(otp),
			ExpiresAt: time.Now().Add(5 * time.Minute),
		}

		// Add token to the database
		if err := s.vtr.Create(txCtx, token); err != nil {
			return err
		}

		// Assign to outer variable for sending SMS after transaction
		mobile = user.Mobile
		return nil
	})

	if err != nil {
		return err
	}

	// Send mobile OTP
	go s.sendMobileVerification(mobile, otp)
	return nil
}

func (s *userService) DeleteUser(
	ctx context.Context,
	id uuid.UUID,
) error {
	return s.r.Delete(ctx, id)
}
