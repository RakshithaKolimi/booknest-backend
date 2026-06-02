package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"booknest/internal/domain"
)

type bookEmbeddingRepository struct {
	db  *gorm.DB
	sql *sql.DB
}

func NewBookEmbeddingRepository(db *gorm.DB, sqlDB *sql.DB) domain.BookEmbeddingRepository {
	return &bookEmbeddingRepository{db: db, sql: sqlDB}
}

func (r *bookEmbeddingRepository) CreateEmbedding(ctx context.Context, embedding *domain.BookEmbedding) error {
	if embedding == nil || embedding.BookID == uuid.Nil {
		return nil
	}
	return r.db.WithContext(ctx).Create(embedding).Error
}

func (r *bookEmbeddingRepository) UpdateEmbedding(ctx context.Context, embedding *domain.BookEmbedding) error {
	if embedding == nil || embedding.BookID == uuid.Nil {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&domain.BookEmbedding{}).
		Where("book_id = ?", embedding.BookID).
		Updates(map[string]any{
			"embedding":   embedding.Embedding,
			"updated_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *bookEmbeddingRepository) GetEmbedding(ctx context.Context, bookID uuid.UUID) (*domain.BookEmbedding, error) {
	if bookID == uuid.Nil {
		return nil, errors.New("book id is required")
	}
	var emb domain.BookEmbedding
	if err := r.db.WithContext(ctx).First(&emb, "book_id = ?", bookID).Error; err != nil {
		return nil, err
	}
	return &emb, nil
}

func (r *bookEmbeddingRepository) UpsertEmbedding(ctx context.Context, embedding *domain.BookEmbedding) error {
	if embedding == nil || embedding.BookID == uuid.Nil {
		return nil
	}

	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "book_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"embedding",
			"updated_at",
		}),
	}).Create(embedding).Error
}

func (r *bookEmbeddingRepository) SearchNearestBooks(
	ctx context.Context,
	query domain.EmbeddingVector,
	limit int,
) ([]domain.Book, error) {
	if limit <= 0 {
		limit = 10
	}
	if len(query) == 0 {
		return nil, errors.New("query embedding is required")
	}
	if r.sql == nil {
		return nil, errors.New("sql db handle is required")
	}

	// Use pgvector distance operator (<->) for nearest-neighbor semantic search.
	const sqlQuery = `
SELECT
  b.id,
  b.name,
  b.author_id,
  b.available_stock,
  b.image_url,
  b.is_active,
  b.description,
  b.summary,
  b.isbn,
  b.price,
  b.discount_percentage,
  b.publisher_id,
  b.created_at,
  b.updated_at
FROM books b
JOIN book_embeddings be ON be.book_id = b.id
WHERE b.deleted_at IS NULL
ORDER BY be.embedding <-> $1
LIMIT $2
`

	rows, err := r.sql.QueryContext(ctx, sqlQuery, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	books := make([]domain.Book, 0, limit)
	for rows.Next() {
		var book domain.Book
		if err := rows.Scan(
			&book.ID,
			&book.Name,
			&book.AuthorID,
			&book.AvailableStock,
			&book.ImageURL,
			&book.IsActive,
			&book.Description,
			&book.Summary,
			&book.ISBN,
			&book.Price,
			&book.DiscountPercentage,
			&book.PublisherID,
			&book.CreatedAt,
			&book.UpdatedAt,
		); err != nil {
			return nil, err
		}
		books = append(books, book)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Note: These Book records are minimally hydrated. If the caller needs
	// Author/Publisher/Categories preloaded, fetch by IDs in a second step.
	return books, nil
}
