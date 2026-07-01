package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"booknest/internal/domain"
)

type chatRepo struct {
	db domain.DBExecer
}

func NewChatRepo(db *pgxpool.Pool) domain.ChatRepository {
	return &chatRepo{db: db}
}

// CreateSession creates a new chat session for the given user.
func (r *chatRepo) CreateSession(ctx context.Context, userID uuid.UUID, title string) (*domain.ChatSession, error) {
	session := &domain.ChatSession{
		ID:     uuid.New(),
		UserID: userID,
		Title:  title,
	}

	query := `
		INSERT INTO chat_sessions (
			id,
			user_id,
			title,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, user_id, title, created_at, updated_at;
	`

	row := queryRowWithTx(ctx, r.db, query, session.ID, session.UserID, session.Title)
	if err := scanChatSession(row, session); err != nil {
		return nil, fmt.Errorf("create chat session: %w", err)
	}

	return session, nil
}

func (r *chatRepo) SaveMessage(ctx context.Context, msg domain.ChatMessage) error {
	if msg.ID == uuid.Nil {
		msg.ID = uuid.New()
	}

	query := `
		INSERT INTO chat_messages (
			id,
			session_id,
			role,
			content,
			created_at
		) VALUES ($1, $2, $3, $4, NOW());
	`

	if err := execWithTx(ctx, r.db, query, msg.ID, msg.SessionID, msg.Role, msg.Content); err != nil {
		return fmt.Errorf("save chat message: %w", err)
	}

	return nil
}

func (r *chatRepo) GetMessages(
	ctx context.Context,
	sessionID uuid.UUID,
	limit int,
) ([]domain.ChatMessage, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, session_id, role, content, created_at
		FROM (
			SELECT id, session_id, role, content, created_at
			FROM chat_messages
			WHERE session_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		) recent_messages
		ORDER BY created_at ASC;
	`

	rows, err := queryWithTx(ctx, r.db, query, sessionID, limit)
	if err != nil {
		return nil, fmt.Errorf("get chat messages: %w", err)
	}
	defer rows.Close()

	messages := make([]domain.ChatMessage, 0)
	for rows.Next() {
		var message domain.ChatMessage
		var role string

		if err := rows.Scan(
			&message.ID,
			&message.SessionID,
			&role,
			&message.Content,
			&message.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan chat message: %w", err)
		}

		message.Role = domain.ChatMessageRole(role)
		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chat messages: %w", err)
	}

	return messages, nil
}

func (r *chatRepo) GetSession(
	ctx context.Context,
	sessionID uuid.UUID,
) (*domain.ChatSession, error) {
	session := &domain.ChatSession{}

	query := `
		SELECT id, user_id, title, created_at, updated_at
		FROM chat_sessions
		WHERE id = $1;
	`

	row := queryRowWithTx(ctx, r.db, query, sessionID)
	if err := scanChatSession(row, session); err != nil {
		return nil, fmt.Errorf("get chat session: %w", err)
	}

	return session, nil
}

type chatSessionRow interface {
	Scan(dest ...any) error
}

func scanChatSession(row chatSessionRow, session *domain.ChatSession) error {
	var title sql.NullString

	if err := row.Scan(
		&session.ID,
		&session.UserID,
		&title,
		&session.CreatedAt,
		&session.UpdatedAt,
	); err != nil {
		return err
	}

	session.Title = title.String
	return nil
}
