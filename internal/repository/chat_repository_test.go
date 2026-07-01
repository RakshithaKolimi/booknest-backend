package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	pgxmock "github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"

	"booknest/internal/domain"
)

func TestNewChatRepo(t *testing.T) {
	repo := NewChatRepo(nil)
	require.NotNil(t, repo)
}

func TestChatRepo_CreateSession(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &chatRepo{db: mock}
	userID := uuid.New()
	sessionID := uuid.New()
	now := time.Now()

	mock.ExpectQuery("INSERT INTO chat_sessions").
		WithArgs(pgxmock.AnyArg(), userID, "Fantasy recommendations").
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "user_id", "title", "created_at", "updated_at"}).
				AddRow(sessionID, userID, "Fantasy recommendations", now, now),
		)

	session, err := repo.CreateSession(context.Background(), userID, "Fantasy recommendations")
	require.NoError(t, err)
	require.Equal(t, sessionID, session.ID)
	require.Equal(t, userID, session.UserID)
	require.Equal(t, "Fantasy recommendations", session.Title)
	require.WithinDuration(t, now, session.CreatedAt, time.Second)
	require.WithinDuration(t, now, session.UpdatedAt, time.Second)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestChatRepo_SaveMessage(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &chatRepo{db: mock}
	sessionID := uuid.New()

	mock.ExpectExec("INSERT INTO chat_messages").
		WithArgs(
			pgxmock.AnyArg(),
			sessionID,
			domain.ChatMessageRoleUser,
			"Can you recommend books like Dune?",
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.SaveMessage(context.Background(), domain.ChatMessage{
		SessionID: sessionID,
		Role:      domain.ChatMessageRoleUser,
		Content:   "Can you recommend books like Dune?",
	})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestChatRepo_GetMessages(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &chatRepo{db: mock}
	sessionID := uuid.New()
	firstID := uuid.New()
	secondID := uuid.New()
	now := time.Now()

	mock.ExpectQuery("SELECT id, session_id, role, content, created_at(.|\n)*FROM chat_messages").
		WithArgs(sessionID, 2).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "session_id", "role", "content", "created_at"}).
				AddRow(firstID, sessionID, "user", "I want a space opera.", now).
				AddRow(secondID, sessionID, "assistant", "Try Hyperion.", now.Add(time.Second)),
		)

	messages, err := repo.GetMessages(context.Background(), sessionID, 2)
	require.NoError(t, err)
	require.Len(t, messages, 2)
	require.Equal(t, firstID, messages[0].ID)
	require.Equal(t, domain.ChatMessageRoleUser, messages[0].Role)
	require.Equal(t, "I want a space opera.", messages[0].Content)
	require.Equal(t, secondID, messages[1].ID)
	require.Equal(t, domain.ChatMessageRoleAssistant, messages[1].Role)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestChatRepo_GetMessagesDefaultLimit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &chatRepo{db: mock}
	sessionID := uuid.New()

	mock.ExpectQuery("SELECT id, session_id, role, content, created_at(.|\n)*FROM chat_messages").
		WithArgs(sessionID, 50).
		WillReturnRows(pgxmock.NewRows([]string{"id", "session_id", "role", "content", "created_at"}))

	messages, err := repo.GetMessages(context.Background(), sessionID, 0)
	require.NoError(t, err)
	require.Empty(t, messages)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestChatRepo_GetSession(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &chatRepo{db: mock}
	sessionID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	mock.ExpectQuery("SELECT id, user_id, title, created_at, updated_at(.|\n)*FROM chat_sessions").
		WithArgs(sessionID).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "user_id", "title", "created_at", "updated_at"}).
				AddRow(sessionID, userID, nil, now, now),
		)

	session, err := repo.GetSession(context.Background(), sessionID)
	require.NoError(t, err)
	require.Equal(t, sessionID, session.ID)
	require.Equal(t, userID, session.UserID)
	require.Empty(t, session.Title)
	require.WithinDuration(t, now, session.CreatedAt, time.Second)
	require.WithinDuration(t, now, session.UpdatedAt, time.Second)
	require.NoError(t, mock.ExpectationsWereMet())
}
