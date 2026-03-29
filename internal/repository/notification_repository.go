package repository

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"booknest/internal/domain"
)

type notificationRepo struct {
	db domain.DBExecer
	sb squirrel.StatementBuilderType
}

func NewNotificationRepo(db *pgxpool.Pool) domain.NotificationRepository {
	return &notificationRepo{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *notificationRepo) Create(ctx context.Context, notification *domain.Notification) error {
	if notification.ID == uuid.Nil {
		notification.ID = uuid.New()
	}

	query, args, err := r.sb.
		Insert("notifications").
		Columns(
			"id",
			"channel",
			"type",
			"recipient",
			"subject",
			"body",
			"provider",
			"status",
			"reference_id",
			"provider_message_id",
			"provider_response",
			"error_message",
		).
		Values(
			notification.ID,
			notification.Channel,
			notification.Type,
			notification.Recipient,
			notification.Subject,
			notification.Body,
			notification.Provider,
			notification.Status,
			notification.ReferenceID,
			notification.ProviderMessageID,
			notification.ProviderResponse,
			notification.ErrorMessage,
		).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return err
	}

	row := queryRowWithTx(ctx, r.db, query, args...)
	return row.Scan(&notification.ID, &notification.CreatedAt, &notification.UpdatedAt)
}
