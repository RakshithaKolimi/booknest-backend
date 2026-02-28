package publisher_service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"booknest/internal/domain"
)

// MockPublisherRepository is a mock implementation of domain.PublisherRepository
type MockPublisherRepository struct {
	FindByIDFunc  func(ctx context.Context, id uuid.UUID) (domain.Publisher, error)
	ListFunc      func(ctx context.Context, limit, offset int, search string) ([]domain.Publisher, error)
	CreateFunc    func(ctx context.Context, publisher *domain.Publisher) error
	UpdateFunc    func(ctx context.Context, publisher *domain.Publisher) error
	SetActiveFunc func(ctx context.Context, id uuid.UUID, active bool) error
	DeleteFunc    func(ctx context.Context, id uuid.UUID) error
}

func (m *MockPublisherRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.Publisher, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return domain.Publisher{}, errors.New("not implemented")
}

func (m *MockPublisherRepository) Create(ctx context.Context, user *domain.Publisher) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	return nil
}

func (m *MockPublisherRepository) List(ctx context.Context, limit, offset int, search string) ([]domain.Publisher, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, limit, offset, search)
	}
	return []domain.Publisher{}, nil
}

func (m *MockPublisherRepository) Update(ctx context.Context, user *domain.Publisher) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, user)
	}
	return nil
}

func (m *MockPublisherRepository) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	if m.SetActiveFunc != nil {
		return m.SetActiveFunc(ctx, id, active)
	}
	return nil
}

func (m *MockPublisherRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}
