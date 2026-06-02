package domain

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EmbeddingVector stores a pgvector embedding.
//
// Why pgvector instead of JSONB:
// - pgvector supports efficient distance operators like `<->` for nearest-neighbor search.
// - pgvector can use specialized indexes (e.g., ivfflat/hnsw) for semantic similarity queries.
// - JSONB requires custom distance computation and does not provide vector indexes or operators.
type EmbeddingVector []float64

func (v EmbeddingVector) Value() (driver.Value, error) {
	if v == nil {
		// pgvector accepts a text representation like `[0.1, 0.2, ...]`.
		return "[]", nil
	}
	encoded, err := json.Marshal([]float64(v))
	if err != nil {
		return nil, err
	}
	return string(encoded), nil
}

func (v *EmbeddingVector) Scan(src any) error {
	switch value := src.(type) {
	case nil:
		*v = EmbeddingVector{}
		return nil
	case []byte:
		return json.Unmarshal(value, v)
	case string:
		return json.Unmarshal([]byte(value), v)
	default:
		return fmt.Errorf("unsupported embedding vector type %T", src)
	}
}

type BookEmbedding struct {
	BookID    uuid.UUID       `gorm:"type:uuid;primaryKey" json:"book_id" format:"uuid"`
	Embedding EmbeddingVector `gorm:"type:vector(1536);not null" json:"-"`
	CreatedAt time.Time       `json:"created_at" format:"date-time"`
	UpdatedAt time.Time       `json:"updated_at" format:"date-time"`
} // @name BookEmbedding

type BookEmbeddingRepository interface {
	CreateEmbedding(ctx context.Context, embedding *BookEmbedding) error
	UpdateEmbedding(ctx context.Context, embedding *BookEmbedding) error
	GetEmbedding(ctx context.Context, bookID uuid.UUID) (*BookEmbedding, error)

	// UpsertEmbedding is used by background refresh jobs to avoid read-before-write.
	UpsertEmbedding(ctx context.Context, embedding *BookEmbedding) error

	// GetEmbeddingsByBookIDs fetches embeddings for a set of book IDs in one query.
	GetEmbeddingsByBookIDs(ctx context.Context, bookIDs []uuid.UUID) ([]BookEmbedding, error)

	// SearchNearestBooks returns the most semantically similar books to the query embedding.
	// excludeIDs skips books already purchased by the user.
	// Implementation uses pgvector distance operator: ORDER BY embedding <-> $1.
	SearchNearestBooks(ctx context.Context, query EmbeddingVector, limit int, excludeIDs []uuid.UUID) ([]Book, error)
}

type BookEmbeddingService interface {
	GenerateBookEmbedding(ctx context.Context, book Book) error
}
