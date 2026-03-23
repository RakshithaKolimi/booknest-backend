package publisher_service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"booknest/internal/domain"
)

// TestFindPublisher_Success tests successful publisher retrieval
func TestFindPublisher_Success(t *testing.T) {
	publisherID := uuid.New()
	expectedPublisher := domain.Publisher{
		ID:          publisherID,
		LegalName:   "Legal",
		TradingName: "Trading",
		Email:       "test@mail.com",
		Mobile:      "+911234567890",
		Address:     "Addr",
		City:        "City",
		State:       "State",
		Country:     "Country",
		Zipcode:     "123456",
	}

	mockPublisherRepo := &MockPublisherRepository{
		FindByIDFunc: func(ctx context.Context, id uuid.UUID) (domain.Publisher, error) {
			if id == publisherID {
				return expectedPublisher, nil
			}
			return domain.Publisher{}, errors.New("publisher not found")
		},
	}

	service := &publisherService{r: mockPublisherRepo}
	publisher, err := service.FindByID(context.Background(), publisherID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if publisher.ID != publisherID {
		t.Fatalf("expected publisher ID %s, got %s", publisherID, publisher.ID)
	}
}

// TestFindPublisher_NotFound tests publisher retrieval when publisher doesn't exist
func TestFindPublisher_NotFound(t *testing.T) {
	mockPublisherRepo := &MockPublisherRepository{
		FindByIDFunc: func(ctx context.Context, id uuid.UUID) (domain.Publisher, error) {
			return domain.Publisher{}, errors.New("publisher not found")
		},
	}

	service := &publisherService{r: mockPublisherRepo}
	_, err := service.FindByID(context.Background(), uuid.New())

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestCreatePublisher_Success(t *testing.T) {
	mockRepo := &MockPublisherRepository{
		CreateFunc: func(ctx context.Context, publisher *domain.Publisher) error {
			if publisher.ID == uuid.Nil {
				t.Fatalf("expected ID to be set")
			}
			return nil
		},
	}

	service := &publisherService{r: mockRepo}

	input := domain.PublisherInput{
		LegalName:   "Legal",
		TradingName: "Trading",
		Email:       "test@mail.com",
		Mobile:      "+911234567890",
		Address:     "Addr",
		City:        "City",
		State:       "State",
		Country:     "Country",
		Zipcode:     "123456",
	}

	publisher, err := service.Create(context.Background(), input)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if publisher.LegalName != input.LegalName {
		t.Fatalf("expected legal name %s, got %s", input.LegalName, publisher.LegalName)
	}
}

func TestUpdatePublisher_Success(t *testing.T) {
	id := uuid.New()

	existing := domain.Publisher{
		ID: id,
	}

	mockRepo := &MockPublisherRepository{
		FindByIDFunc: func(ctx context.Context, publisherID uuid.UUID) (domain.Publisher, error) {
			return existing, nil
		},
		UpdateFunc: func(ctx context.Context, publisher *domain.Publisher) error {
			if publisher.ID != id {
				t.Fatalf("unexpected publisher ID")
			}
			return nil
		},
	}

	service := &publisherService{r: mockRepo}

	input := domain.PublisherInput{
		LegalName:   "Updated Legal",
		TradingName: "Updated Trading",
		Email:       "updated@mail.com",
		Mobile:      "+911111111111",
		Address:     "Updated Addr",
		City:        "City",
		State:       "State",
		Country:     "Country",
		Zipcode:     "654321",
	}

	publisher, err := service.Update(context.Background(), id, input)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if publisher.LegalName != input.LegalName {
		t.Fatalf("expected updated legal name")
	}
}

func TestSetActivePublisher_Success(t *testing.T) {
	id := uuid.New()

	mockRepo := &MockPublisherRepository{
		SetActiveFunc: func(ctx context.Context, publisherID uuid.UUID, active bool) error {
			if publisherID != id {
				t.Fatalf("unexpected publisher ID")
			}
			if active != false {
				t.Fatalf("expected active=false")
			}
			return nil
		},
	}

	service := &publisherService{r: mockRepo}

	err := service.SetActive(context.Background(), id, false)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDeletePublisher_Success(t *testing.T) {
	id := uuid.New()

	mockRepo := &MockPublisherRepository{
		DeleteFunc: func(ctx context.Context, publisherID uuid.UUID) error {
			if publisherID != id {
				t.Fatalf("unexpected publisher ID")
			}
			return nil
		},
	}

	service := &publisherService{r: mockRepo}

	err := service.Delete(context.Background(), id)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUpdatePublisher_NotFound(t *testing.T) {
	mockRepo := &MockPublisherRepository{
		FindByIDFunc: func(ctx context.Context, id uuid.UUID) (domain.Publisher, error) {
			return domain.Publisher{}, errors.New("publisher not found")
		},
	}

	service := &publisherService{r: mockRepo}

	_, err := service.Update(context.Background(), uuid.New(), domain.PublisherInput{})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestNewPublisherServiceAndList(t *testing.T) {
	expected := []domain.Publisher{{ID: uuid.New(), LegalName: "Legal"}}
	mockRepo := &MockPublisherRepository{
		ListFunc: func(ctx context.Context, limit, offset int, search string) ([]domain.Publisher, error) {
			if limit != 10 || offset != 5 || search != "pub" {
				t.Fatalf("unexpected list args: %d %d %q", limit, offset, search)
			}
			return expected, nil
		},
	}

	service := NewPublisherService(mockRepo)
	if service == nil {
		t.Fatal("expected non-nil service")
	}

	publishers, err := service.List(context.Background(), 10, 5, "pub")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(publishers) != 1 || publishers[0].ID != expected[0].ID {
		t.Fatalf("unexpected publishers: %+v", publishers)
	}
}
