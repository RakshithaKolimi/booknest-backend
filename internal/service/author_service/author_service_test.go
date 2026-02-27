package author_service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"booknest/internal/domain"
)

type mockAuthorRepository struct {
	findByIDFunc   func(ctx context.Context, id uuid.UUID) (domain.Author, error)
	findByNameFunc func(ctx context.Context, name string) (domain.Author, error)
	listFunc       func(ctx context.Context, limit, offset int, search string) ([]domain.Author, error)
	createFunc     func(ctx context.Context, author *domain.Author) error
	updateFunc     func(ctx context.Context, author *domain.Author) error
	deleteFunc     func(ctx context.Context, id uuid.UUID) error
}

func (m *mockAuthorRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.Author, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return domain.Author{}, errors.New("not implemented")
}

func (m *mockAuthorRepository) FindByName(ctx context.Context, name string) (domain.Author, error) {
	if m.findByNameFunc != nil {
		return m.findByNameFunc(ctx, name)
	}
	return domain.Author{}, gorm.ErrRecordNotFound
}

func (m *mockAuthorRepository) List(ctx context.Context, limit, offset int, search string) ([]domain.Author, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, limit, offset, search)
	}
	return []domain.Author{}, nil
}

func (m *mockAuthorRepository) Create(ctx context.Context, author *domain.Author) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, author)
	}
	return nil
}

func (m *mockAuthorRepository) Update(ctx context.Context, author *domain.Author) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, author)
	}
	return nil
}

func (m *mockAuthorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func TestCreateAuthorSuccess(t *testing.T) {
	repo := &mockAuthorRepository{
		findByNameFunc: func(ctx context.Context, name string) (domain.Author, error) {
			if name != "Jane Austen" {
				t.Fatalf("expected trimmed name, got %q", name)
			}
			return domain.Author{}, gorm.ErrRecordNotFound
		},
		createFunc: func(ctx context.Context, author *domain.Author) error {
			if author.ID == uuid.Nil {
				t.Fatalf("expected non-nil id")
			}
			if author.Name != "Jane Austen" {
				t.Fatalf("unexpected author name: %q", author.Name)
			}
			return nil
		},
	}

	svc := NewAuthorService(repo)
	author, err := svc.Create(context.Background(), domain.AuthorInput{Name: "  Jane Austen  "})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if author == nil || author.Name != "Jane Austen" {
		t.Fatalf("unexpected author: %+v", author)
	}
}

func TestCreateAuthorValidationAndDuplicate(t *testing.T) {
	svc := NewAuthorService(&mockAuthorRepository{})
	_, err := svc.Create(context.Background(), domain.AuthorInput{Name: "   "})
	if err == nil || err.Error() != "author name is required" {
		t.Fatalf("expected required-name error, got %v", err)
	}

	repo := &mockAuthorRepository{
		findByNameFunc: func(ctx context.Context, name string) (domain.Author, error) {
			return domain.Author{ID: uuid.New(), Name: name}, nil
		},
	}
	svc = NewAuthorService(repo)
	_, err = svc.Create(context.Background(), domain.AuthorInput{Name: "Jane Austen"})
	if err == nil || err.Error() != "author name already exists" {
		t.Fatalf("expected duplicate error, got %v", err)
	}
}

func TestUpdateAuthorSuccessAndConflict(t *testing.T) {
	authorID := uuid.New()
	repo := &mockAuthorRepository{
		findByIDFunc: func(ctx context.Context, id uuid.UUID) (domain.Author, error) {
			return domain.Author{ID: authorID, Name: "Old"}, nil
		},
		findByNameFunc: func(ctx context.Context, name string) (domain.Author, error) {
			if name != "New Name" {
				t.Fatalf("expected trimmed update name, got %q", name)
			}
			return domain.Author{}, gorm.ErrRecordNotFound
		},
		updateFunc: func(ctx context.Context, author *domain.Author) error {
			if author.ID != authorID || author.Name != "New Name" {
				t.Fatalf("unexpected update payload: %+v", author)
			}
			return nil
		},
	}

	svc := NewAuthorService(repo)
	updated, err := svc.Update(context.Background(), authorID, domain.AuthorInput{Name: " New Name "})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Name != "New Name" {
		t.Fatalf("unexpected updated author: %+v", updated)
	}

	conflictRepo := &mockAuthorRepository{
		findByIDFunc: func(ctx context.Context, id uuid.UUID) (domain.Author, error) {
			return domain.Author{ID: authorID, Name: "Old"}, nil
		},
		findByNameFunc: func(ctx context.Context, name string) (domain.Author, error) {
			return domain.Author{ID: uuid.New(), Name: name}, nil
		},
	}

	svc = NewAuthorService(conflictRepo)
	_, err = svc.Update(context.Background(), authorID, domain.AuthorInput{Name: "New Name"})
	if err == nil || err.Error() != "author name already exists" {
		t.Fatalf("expected duplicate-name error, got %v", err)
	}
}

func TestAuthorReadAndDeletePassThrough(t *testing.T) {
	authorID := uuid.New()
	expected := domain.Author{ID: authorID, Name: "Test"}
	repo := &mockAuthorRepository{
		findByIDFunc: func(ctx context.Context, id uuid.UUID) (domain.Author, error) {
			return expected, nil
		},
		listFunc: func(ctx context.Context, limit, offset int, search string) ([]domain.Author, error) {
			if limit != 5 || offset != 10 || search != "test" {
				t.Fatalf("unexpected list params: %d/%d %q", limit, offset, search)
			}
			return []domain.Author{expected}, nil
		},
		deleteFunc: func(ctx context.Context, id uuid.UUID) error {
			if id != authorID {
				t.Fatalf("unexpected id: %s", id)
			}
			return nil
		},
	}

	svc := NewAuthorService(repo)
	author, err := svc.FindByID(context.Background(), authorID)
	if err != nil || author.ID != authorID {
		t.Fatalf("unexpected find result: %+v, err=%v", author, err)
	}

	list, err := svc.List(context.Background(), 5, 10, "test")
	if err != nil || len(list) != 1 {
		t.Fatalf("unexpected list result: %+v, err=%v", list, err)
	}

	if err := svc.Delete(context.Background(), authorID); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}
}
