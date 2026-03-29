package domain

import (
	"context"

	"github.com/google/uuid"
)

type NotificationChannel string
type NotificationType string
type NotificationStatus string

const (
	NotificationChannelEmail NotificationChannel = "EMAIL"

	NotificationTypeVerificationEmail NotificationType = "VERIFICATION_EMAIL"
	NotificationTypePasswordReset     NotificationType = "PASSWORD_RESET"
	NotificationTypeOrderReceipt      NotificationType = "ORDER_RECEIPT"

	NotificationStatusSent   NotificationStatus = "SENT"
	NotificationStatusFailed NotificationStatus = "FAILED"

	EmailNotificationProviderSES = "SES"
)

type EmailSendResult struct {
	Provider  string
	MessageID string
	Response  string
}

type Notification struct {
	ID                uuid.UUID           `gorm:"type:uuid;primaryKey" db:"id" json:"id"`
	Channel           NotificationChannel `gorm:"type:varchar(20);not null" db:"channel" json:"channel"`
	Type              NotificationType    `gorm:"type:varchar(40);not null" db:"type" json:"type"`
	Recipient         string              `gorm:"not null" db:"recipient" json:"recipient"`
	Subject           string              `gorm:"not null" db:"subject" json:"subject"`
	Body              string              `gorm:"not null" db:"body" json:"body"`
	Provider          string              `gorm:"not null" db:"provider" json:"provider"`
	Status            NotificationStatus  `gorm:"type:varchar(20);not null" db:"status" json:"status"`
	ReferenceID       *string             `db:"reference_id" json:"reference_id,omitempty"`
	ProviderMessageID *string             `db:"provider_message_id" json:"provider_message_id,omitempty"`
	ProviderResponse  *string             `db:"provider_response" json:"provider_response,omitempty"`
	ErrorMessage      *string             `db:"error_message" json:"error_message,omitempty"`
	BaseEntity
}

type NotificationRepository interface {
	Create(ctx context.Context, notification *Notification) error
}

type SMSProvider interface {
	SendSMS(to string, message string) error
}

type EmailProvider interface {
	SendEmail(to string, subject string, body string) (EmailSendResult, error)
}

type NotificationService interface {
	SendVerificationEmail(email string, link string) error
	SendPasswordReset(email string, link string) error
	SendOrderReceipt(email string, orderID string) error
}
