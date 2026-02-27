package author_service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"booknest/internal/domain"
)

type authorService struct {
	r domain.AuthorRepository
}

func NewAuthorService(r domain.AuthorRepository) domain.AuthorService {
	return &authorService{
		r: r,
	}
}

func (s *authorService) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Author, error) {
	author, err := s.r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &author, nil
}

func (s *authorService) List(
	ctx context.Context,
	limit, offset int,
	search string,
) ([]domain.Author, error) {
	return s.r.List(ctx, limit, offset, search)
}

func (s *authorService) Create(
	ctx context.Context,
	input domain.AuthorInput,
) (*domain.Author, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, errors.New("author name is required")
	}

	_, err := s.r.FindByName(ctx, name)
	if err == nil {
		return nil, errors.New("author name already exists")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	author := &domain.Author{
		ID:   uuid.New(),
		Name: name,
	}

	if err := s.r.Create(ctx, author); err != nil {
		return nil, err
	}

	return author, nil
}

func (s *authorService) Update(
	ctx context.Context,
	id uuid.UUID,
	input domain.AuthorInput,
) (*domain.Author, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, errors.New("author name is required")
	}

	author, err := s.r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	existing, err := s.r.FindByName(ctx, name)
	if err == nil && existing.ID != author.ID {
		return nil, errors.New("author name already exists")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	author.Name = name
	if err := s.r.Update(ctx, &author); err != nil {
		return nil, err
	}

	return &author, nil
}

func (s *authorService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.r.Delete(ctx, id)
}
