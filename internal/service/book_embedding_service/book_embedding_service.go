package book_embedding_service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"booknest/internal/ai"
	"booknest/internal/domain"
	"booknest/internal/service/ai_service"
)

type service struct {
	ai   domain.AIService
	repo domain.BookEmbeddingRepository
}

func New(aiSvc domain.AIService, repo domain.BookEmbeddingRepository) domain.BookEmbeddingService {
	return &service{ai: aiSvc, repo: repo}
}

func (s *service) GenerateBookEmbedding(ctx context.Context, book domain.Book) error {
	if s.ai == nil {
		return ai_service.ErrProviderUnavailable
	}
	if s.repo == nil {
		return errors.New("book embedding repository is required")
	}
	if book.ID == uuid.Nil {
		return errors.New("book id is required")
	}

	text := ai.BuildEmbeddingText(book)
	vectors, err := s.ai.Embed(ctx, []string{text})
	if err != nil {
		return err
	}
	if len(vectors) != 1 || len(vectors[0]) == 0 {
		return fmt.Errorf("unexpected embedding response size: %d", len(vectors))
	}

	embedding := &domain.BookEmbedding{
		BookID:    book.ID,
		Embedding: domain.EmbeddingVector(vectors[0]),
	}

	return s.repo.UpsertEmbedding(ctx, embedding)
}

