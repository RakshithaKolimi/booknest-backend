package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	pgxmock "github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"

	"booknest/internal/domain"
)

func TestNotificationRepo_Create(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &notificationRepo{
		db: mock,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	referenceID := "ref-123"
	messageID := "msg-123"
	response := "{\"message_id\":\"msg-123\"}"
	notification := &domain.Notification{
		Channel:           domain.NotificationChannelEmail,
		Type:              domain.NotificationTypeVerificationEmail,
		Recipient:         "user@example.com",
		Subject:           "Verify",
		Body:              "<p>Verify</p>",
		Provider:          domain.EmailNotificationProviderSES,
		Status:            domain.NotificationStatusSent,
		ReferenceID:       &referenceID,
		ProviderMessageID: &messageID,
		ProviderResponse:  &response,
	}

	mock.ExpectQuery("INSERT INTO notifications").
		WithArgs(
			pgxmock.AnyArg(),
			notification.Channel,
			notification.Type,
			notification.Recipient,
			notification.Subject,
			notification.Body,
			notification.Provider,
			notification.Status,
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			notification.ErrorMessage,
		).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "created_at", "updated_at"}).
				AddRow(uuid.New(), time.Now(), time.Now()),
		)

	err = repo.Create(context.Background(), notification)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
