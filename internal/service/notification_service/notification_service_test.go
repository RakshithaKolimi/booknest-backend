package notification_service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"booknest/internal/domain"
)

type mockEmailProvider struct {
	sendEmailFunc func(to string, subject string, body string) (domain.EmailSendResult, error)
}

func (m *mockEmailProvider) SendEmail(to string, subject string, body string) (domain.EmailSendResult, error) {
	if m.sendEmailFunc != nil {
		return m.sendEmailFunc(to, subject, body)
	}
	return domain.EmailSendResult{}, nil
}

func TestSendVerificationEmail(t *testing.T) {
	saved := &domain.Notification{}
	provider := &mockEmailProvider{
		sendEmailFunc: func(to string, subject string, body string) (domain.EmailSendResult, error) {
			if to != "user@example.com" {
				t.Fatalf("unexpected recipient: %q", to)
			}
			if subject != "Verify your BookNest account" {
				t.Fatalf("unexpected subject: %q", subject)
			}
			if !strings.Contains(body, "https://booknest.example/verify") {
				t.Fatalf("expected verification link in body, got %q", body)
			}
			return domain.EmailSendResult{
				Provider:  domain.EmailNotificationProviderSES,
				MessageID: "msg-1",
				Response:  "{\"message_id\":\"msg-1\"}",
			}, nil
		},
	}

	repo := &recordingNotificationRepository{createFunc: func(notification *domain.Notification) error {
		*saved = *notification
		return nil
	}}

	svc := NewNotificationServiceWithRepository(provider, repo)
	if err := svc.SendVerificationEmail("user@example.com", "https://booknest.example/verify"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if saved.Status != domain.NotificationStatusSent || saved.Type != domain.NotificationTypeVerificationEmail {
		t.Fatalf("unexpected saved notification: %+v", saved)
	}
}

func TestSendPasswordReset(t *testing.T) {
	provider := &mockEmailProvider{
		sendEmailFunc: func(to string, subject string, body string) (domain.EmailSendResult, error) {
			if to != "user@example.com" {
				t.Fatalf("unexpected recipient: %q", to)
			}
			if subject != "Reset your BookNest password" {
				t.Fatalf("unexpected subject: %q", subject)
			}
			if !strings.Contains(body, "https://booknest.example/reset") {
				t.Fatalf("expected reset link in body, got %q", body)
			}
			return domain.EmailSendResult{Provider: domain.EmailNotificationProviderSES}, nil
		},
	}

	svc := NewNotificationService(provider)
	if err := svc.SendPasswordReset("user@example.com", "https://booknest.example/reset"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSendOrderReceipt(t *testing.T) {
	provider := &mockEmailProvider{
		sendEmailFunc: func(to string, subject string, body string) (domain.EmailSendResult, error) {
			if to != "user@example.com" {
				t.Fatalf("unexpected recipient: %q", to)
			}
			if subject != "Your BookNest order receipt: ORDER-123" {
				t.Fatalf("unexpected subject: %q", subject)
			}
			if !strings.Contains(body, "ORDER-123") {
				t.Fatalf("expected order id in body, got %q", body)
			}
			return domain.EmailSendResult{Provider: domain.EmailNotificationProviderSES}, nil
		},
	}

	svc := NewNotificationService(provider)
	if err := svc.SendOrderReceipt("user@example.com", "ORDER-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSendVerificationEmail_RecordsFailure(t *testing.T) {
	saved := &domain.Notification{}
	provider := &mockEmailProvider{
		sendEmailFunc: func(to string, subject string, body string) (domain.EmailSendResult, error) {
			return domain.EmailSendResult{
				Provider: domain.EmailNotificationProviderSES,
				Response: "{\"error\":\"send failed\"}",
			}, errors.New("send failed")
		},
	}
	repo := &recordingNotificationRepository{createFunc: func(notification *domain.Notification) error {
		*saved = *notification
		return nil
	}}

	svc := NewNotificationServiceWithRepository(provider, repo)
	err := svc.SendVerificationEmail("user@example.com", "https://booknest.example/verify")
	if err == nil || err.Error() != "send failed" {
		t.Fatalf("expected send failure, got %v", err)
	}
	if saved.Status != domain.NotificationStatusFailed {
		t.Fatalf("expected failed notification status, got %+v", saved)
	}
	if saved.ErrorMessage == nil || *saved.ErrorMessage != "send failed" {
		t.Fatalf("expected error message to be captured, got %+v", saved)
	}
}

func TestSendOrderReceipt_ReturnsErrorWhenProviderMissing(t *testing.T) {
	svc := NewNotificationService(nil)

	err := svc.SendOrderReceipt("user@example.com", "ORDER-123")
	if err == nil || !strings.Contains(err.Error(), "email provider is not configured") {
		t.Fatalf("expected missing provider error, got %v", err)
	}
}

type recordingNotificationRepository struct {
	createFunc func(notification *domain.Notification) error
}

func (m *recordingNotificationRepository) Create(_ context.Context, notification *domain.Notification) error {
	if m.createFunc != nil {
		return m.createFunc(notification)
	}
	return nil
}
