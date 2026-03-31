package notification_service

import (
	"context"
	"fmt"

	"booknest/internal/domain"
)

type notificationService struct {
	emailProvider domain.EmailProvider
	smsProvider   domain.SMSProvider
	repo          domain.NotificationRepository
}

func NewNotificationService(emailProvider domain.EmailProvider) domain.NotificationService {
	return &notificationService{
		emailProvider: emailProvider,
	}
}

func NewNotificationServiceWithRepository(
	emailProvider domain.EmailProvider,
	repo domain.NotificationRepository,
) domain.NotificationService {
	return &notificationService{
		emailProvider: emailProvider,
		repo:          repo,
	}
}

func NewNotificationServiceWithProviders(
	emailProvider domain.EmailProvider,
	smsProvider domain.SMSProvider,
) domain.NotificationService {
	return &notificationService{
		emailProvider: emailProvider,
		smsProvider:   smsProvider,
	}
}

func NewNotificationServiceWithProvidersAndRepository(
	emailProvider domain.EmailProvider,
	smsProvider domain.SMSProvider,
	repo domain.NotificationRepository,
) domain.NotificationService {
	return &notificationService{
		emailProvider: emailProvider,
		smsProvider:   smsProvider,
		repo:          repo,
	}
}

func (s *notificationService) SendVerificationEmail(email string, link string) error {
	subject := "Verify your BookNest account"
	body, err := renderTemplate(templateKeyVerificationEmail, struct {
		Link string
	}{
		Link: link,
	})
	if err != nil {
		return err
	}

	return s.sendAndRecord(
		context.Background(),
		domain.NotificationTypeVerificationEmail,
		email,
		subject,
		body,
		&link,
	)
}

func (s *notificationService) SendPasswordReset(email string, link string) error {
	subject := "Reset your BookNest password"
	body, err := renderTemplate(templateKeyPasswordReset, struct {
		Link string
	}{
		Link: link,
	})
	if err != nil {
		return err
	}

	return s.sendAndRecord(
		context.Background(),
		domain.NotificationTypePasswordReset,
		email,
		subject,
		body,
		&link,
	)
}

func (s *notificationService) SendOrderReceipt(email string, orderID string) error {
	subject := fmt.Sprintf("Your BookNest order receipt: %s", orderID)
	body, err := renderTemplate(templateKeyOrderReceipt, struct {
		OrderID string
	}{
		OrderID: orderID,
	})
	if err != nil {
		return err
	}

	return s.sendAndRecord(
		context.Background(),
		domain.NotificationTypeOrderReceipt,
		email,
		subject,
		body,
		&orderID,
	)
}

func (s *notificationService) SendOTP(phone string, otp string) error {
	subject := "Your BookNest OTP"
	body := fmt.Sprintf("Your BookNest verification code is %s. It expires soon. Do not share this code with anyone.", otp)

	return s.sendSMSAndRecord(
		context.Background(),
		domain.NotificationTypeOTP,
		phone,
		subject,
		body,
		&otp,
	)
}

func (s *notificationService) SendLoginAlert(phone string, device string, location string) error {
	subject := "BookNest login alert"
	body := fmt.Sprintf("New BookNest login detected from %s in %s. If this was not you, secure your account right away.", device, location)

	return s.sendSMSAndRecord(
		context.Background(),
		domain.NotificationTypeLoginAlert,
		phone,
		subject,
		body,
		nil,
	)
}

func (s *notificationService) SendOrderConfirmation(phone string, orderID string) error {
	subject := "BookNest order confirmation"
	body := fmt.Sprintf("Your BookNest order %s has been confirmed. We will notify you when it ships.", orderID)

	return s.sendSMSAndRecord(
		context.Background(),
		domain.NotificationTypeOrderConfirmation,
		phone,
		subject,
		body,
		&orderID,
	)
}

func (s *notificationService) SendOrderCancellation(phone string, orderID string, reason string) error {
	subject := "BookNest order cancellation"
	body := fmt.Sprintf("Your BookNest order %s has been cancelled. Reason: %s.", orderID, reason)

	return s.sendSMSAndRecord(
		context.Background(),
		domain.NotificationTypeOrderCancellation,
		phone,
		subject,
		body,
		&orderID,
	)
}

func (s *notificationService) sendAndRecord(
	ctx context.Context,
	notificationType domain.NotificationType,
	recipient string,
	subject string,
	body string,
	referenceID *string,
) error {
	if s.emailProvider == nil {
		return fmt.Errorf("email provider is not configured")
	}

	result, err := s.emailProvider.SendEmail(recipient, subject, body)

	notification := &domain.Notification{
		Channel:     domain.NotificationChannelEmail,
		Type:        notificationType,
		Recipient:   recipient,
		Subject:     subject,
		Body:        body,
		Provider:    result.Provider,
		ReferenceID: referenceID,
	}

	if result.MessageID != "" {
		notification.ProviderMessageID = &result.MessageID
	}
	if result.Response != "" {
		notification.ProviderResponse = &result.Response
	}

	if err != nil {
		notification.Status = domain.NotificationStatusFailed
		errorMessage := err.Error()
		notification.ErrorMessage = &errorMessage
		if repoErr := s.persist(ctx, notification); repoErr != nil {
			return repoErr
		}
		return err
	}

	notification.Status = domain.NotificationStatusSent
	return s.persist(ctx, notification)
}

func (s *notificationService) persist(ctx context.Context, notification *domain.Notification) error {
	if s.repo == nil {
		return nil
	}

	return s.repo.Create(ctx, notification)
}

func (s *notificationService) sendSMSAndRecord(
	ctx context.Context,
	notificationType domain.NotificationType,
	recipient string,
	subject string,
	body string,
	referenceID *string,
) error {
	if s.smsProvider == nil {
		return fmt.Errorf("sms provider is not configured")
	}

	err := s.smsProvider.SendSMS(recipient, body)
	notification := &domain.Notification{
		Channel:     domain.NotificationChannelSMS,
		Type:        notificationType,
		Recipient:   recipient,
		Subject:     subject,
		Body:        body,
		Provider:    domain.SMSNotificationProviderSNS,
		ReferenceID: referenceID,
	}

	if err != nil {
		notification.Status = domain.NotificationStatusFailed
		errorMessage := err.Error()
		notification.ErrorMessage = &errorMessage
		if repoErr := s.persist(ctx, notification); repoErr != nil {
			return repoErr
		}
		return err
	}

	notification.Status = domain.NotificationStatusSent
	return s.persist(ctx, notification)
}
