package notification_service

import "booknest/internal/domain"

type noopEmailProvider struct{}

func (p *noopEmailProvider) SendEmail(to string, subject string, body string) (domain.EmailSendResult, error) {
	return domain.EmailSendResult{
		Provider: "NOOP",
		Response: "notification delivery disabled",
	}, nil
}
