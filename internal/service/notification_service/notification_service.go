package notification_service

import (
	"context"
	"fmt"

	"booknest/internal/domain"
)

type notificationService struct {
	emailProvider domain.EmailProvider
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
