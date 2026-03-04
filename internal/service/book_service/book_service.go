package book_service

import (
	"context"

	"github.com/google/uuid"

	"booknest/internal/domain"
)

type bookService struct {
	repo domain.BookRepository
}

func NewBookService(repo domain.BookRepository) domain.BookService {
	return &bookService{
		repo: repo,
	}
}

func (s *bookService) CreateBook(ctx context.Context, input domain.BookInput) (*domain.Book, error) {
	return s.repo.CreateWithRelations(ctx, input)
}

func (s *bookService) GetBook(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *bookService) ListBooks(ctx context.Context, limit, offset int) ([]domain.Book, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *bookService) FilterByCriteria(
	ctx context.Context,
	filter domain.BookFilter,
	q domain.QueryOptions,
) (*domain.BookSearchResult, error) {

	books, total, err := s.repo.FilterByCriteria(ctx, filter, q)
	if err != nil {
		return nil, err
	}

	return &domain.BookSearchResult{
		Items:  books,
		Total:  total,
		Limit:  q.Limit,
		Offset: q.Offset,
	}, nil
}

func (s *bookService) QueryBooks(
	ctx context.Context,
	filter domain.BookFilter,
	q domain.QueryOptions,
) (*domain.BookSearchResult, error) {
	books, total, nextCursor, hasMore, err := s.repo.QueryBooks(ctx, filter, q)
	if err != nil {
		return nil, err
	}

	return &domain.BookSearchResult{
		Items:      books,
		Total:      total,
		Limit:      q.Limit,
		Offset:     q.Offset,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (s *bookService) UpdateBook(
	ctx context.Context,
	id uuid.UUID,
	input domain.BookInput,
) (*domain.Book, error) {
	return s.repo.UpdateWithRelations(ctx, id, input)
}

func (s *bookService) DeleteBook(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
