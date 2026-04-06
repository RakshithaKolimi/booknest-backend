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
	NotificationChannelSMS   NotificationChannel = "SMS"

	NotificationTypeVerificationEmail NotificationType = "VERIFICATION_EMAIL"
	NotificationTypePasswordReset     NotificationType = "PASSWORD_RESET"
	NotificationTypeOrderReceipt      NotificationType = "ORDER_RECEIPT"
	NotificationTypeOTP               NotificationType = "OTP"
	NotificationTypeLoginAlert        NotificationType = "LOGIN_ALERT"
	NotificationTypeOrderConfirmation NotificationType = "ORDER_CONFIRMATION"
	NotificationTypeOrderCancellation NotificationType = "ORDER_CANCELLATION"

	NotificationStatusSent   NotificationStatus = "SENT"
	NotificationStatusFailed NotificationStatus = "FAILED"

	EmailNotificationProviderSES = "SES"

	SMSNotificationProviderSNS = "SNS"
)

type EmailSendResult struct {
	Provider  string
	MessageID string
	Response  string
}

type Notification struct {
	ID                uuid.UUID           `gorm:"type:uuid;primaryKey" db:"id" json:"id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440008"`
	Channel           NotificationChannel `gorm:"type:varchar(20);not null" db:"channel" json:"channel" enums:"EMAIL,SMS" example:"EMAIL"`
	Type              NotificationType    `gorm:"type:varchar(40);not null" db:"type" json:"type" enums:"VERIFICATION_EMAIL,PASSWORD_RESET,ORDER_RECEIPT,OTP,LOGIN_ALERT,ORDER_CONFIRMATION,ORDER_CANCELLATION" example:"ORDER_RECEIPT"`
	Recipient         string              `gorm:"not null" db:"recipient" json:"recipient" example:"rakshitha@example.com"`
	Subject           string              `gorm:"not null" db:"subject" json:"subject" example:"Your BookNest order receipt"`
	Body              string              `gorm:"not null" db:"body" json:"body" example:"Thank you for your purchase. Your order has been confirmed."`
	Provider          string              `gorm:"not null" db:"provider" json:"provider" example:"SES"`
	Status            NotificationStatus  `gorm:"type:varchar(20);not null" db:"status" json:"status" enums:"SENT,FAILED" example:"SENT"`
	ReferenceID       *string             `db:"reference_id" json:"reference_id,omitempty" example:"ORD-20260406-0001"`
	ProviderMessageID *string             `db:"provider_message_id" json:"provider_message_id,omitempty" example:"0102019487abc123-456def"`
	ProviderResponse  *string             `db:"provider_response" json:"provider_response,omitempty" example:"Queued successfully"`
	ErrorMessage      *string             `db:"error_message" json:"error_message,omitempty" example:"Email provider timeout"`
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
	SendOTP(phone string, otp string) error
	SendLoginAlert(phone string, device string, location string) error
	SendOrderConfirmation(phone string, orderID string) error
	SendOrderCancellation(phone string, orderID string, reason string) error

	SendVerificationEmail(email string, link string) error
	SendPasswordReset(email string, link string) error
	SendOrderReceipt(email string, orderID string) error
}
