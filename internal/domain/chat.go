package domain

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ChatMessageRole string

const (
	ChatMessageRoleUser      ChatMessageRole = "user"
	ChatMessageRoleAssistant ChatMessageRole = "assistant"
	ChatMessageRoleSystem    ChatMessageRole = "system"
)

type ChatSession struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" db:"id" json:"id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440009"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" db:"user_id" json:"user_id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440004"`
	Title     string    `gorm:"type:varchar(255)" db:"title" json:"title" example:"Finding a new fantasy series"`
	CreatedAt time.Time `db:"created_at" json:"created_at" format:"date-time" example:"2026-06-22T10:30:00Z"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at" format:"date-time" example:"2026-06-22T11:45:00Z"`
} // @name ChatSession

type ChatMessage struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey" db:"id" json:"id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440010"`
	SessionID       uuid.UUID       `gorm:"type:uuid;not null;index" db:"session_id" json:"session_id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440009"`
	Role            ChatMessageRole `gorm:"type:varchar(20);not null" db:"role" json:"role" enums:"user,assistant,system" example:"user"`
	Content         string          `gorm:"type:text;not null" db:"content" json:"content" example:"Can you recommend books like Dune?"`
	ReferenceTitles StringSlice     `gorm:"type:jsonb;not null;default:'[]'" db:"reference_titles" json:"reference_titles,omitempty" example:"[\"The Hobbit\",\"A Wizard of Earthsea\"]"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at" format:"date-time" example:"2026-06-22T10:31:00Z"`
} // @name ChatMessage

// StringSlice stores a JSON array of strings in the database.
type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}

	encoded, err := json.Marshal([]string(s))
	if err != nil {
		return nil, err
	}

	return string(encoded), nil
}

func (s *StringSlice) Scan(src any) error {
	switch value := src.(type) {
	case nil:
		*s = StringSlice{}
		return nil
	case []byte:
		return json.Unmarshal(value, s)
	case string:
		return json.Unmarshal([]byte(value), s)
	default:
		return fmt.Errorf("unsupported string slice type %T", src)
	}
}

type ChatRepository interface {
	CreateSession(
		ctx context.Context,
		userID uuid.UUID,
		title string,
	) (*ChatSession, error)

	SaveMessage(
		ctx context.Context,
		msg ChatMessage,
	) error

	GetMessages(
		ctx context.Context,
		sessionID uuid.UUID,
		limit int,
	) ([]ChatMessage, error)

	GetSession(
		ctx context.Context,
		sessionID uuid.UUID,
	) (*ChatSession, error)
}
