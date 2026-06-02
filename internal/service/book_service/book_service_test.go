package book_service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"booknest/internal/domain"
)

type mockBookRepository struct {
	createWithRelationsFunc        func(ctx context.Context, input domain.BookInput) (*domain.Book, error)
	findByIDFunc                   func(ctx context.Context, id uuid.UUID) (*domain.Book, error)
	listBooksWithoutEmbeddingsFunc func(ctx context.Context, limit, offset int) ([]domain.Book, error)
	listFunc                       func(ctx context.Context, limit, offset int) ([]domain.Book, error)
	filterByCriteriaFunc           func(ctx context.Context, filter domain.BookFilter, pagination domain.QueryOptions) ([]domain.Book, int64, error)
	queryBooksFunc                 func(ctx context.Context, filter domain.BookFilter, pagination domain.QueryOptions) ([]domain.Book, int64, *string, bool, error)
	updateFunc                     func(ctx context.Context, book *domain.Book) error
	updateWithRelationsFunc        func(ctx context.Context, id uuid.UUID, input domain.BookInput) (*domain.Book, error)
	replaceCategoriesFunc          func(ctx context.Context, bookID uuid.UUID, categoryIDs []uuid.UUID) error
	upsertEmbeddingsFunc           func(ctx context.Context, embeddings *domain.BookEmbedding) error
	deleteFunc                     func(ctx context.Context, id uuid.UUID) error
}

type mockBookEmbeddingService struct {
	generateBookEmbeddingFunc func(ctx context.Context, book domain.Book) error
}

func (m *mockBookEmbeddingService) GenerateBookEmbedding(ctx context.Context, book domain.Book) error {
	if m.generateBookEmbeddingFunc != nil {
		return m.generateBookEmbeddingFunc(ctx, book)
	}
	return nil
}

func (m *mockBookRepository) Create(ctx context.Context, book *domain.Book) error {
	return nil
}

func (m *mockBookRepository) CreateWithRelations(ctx context.Context, input domain.BookInput) (*domain.Book, error) {
	if m.createWithRelationsFunc != nil {
		return m.createWithRelationsFunc(ctx, input)
	}
	return nil, errors.New("not implemented")
}

func (m *mockBookRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockBookRepository) List(ctx context.Context, limit, offset int) ([]domain.Book, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, limit, offset)
	}
	return []domain.Book{}, nil
}

func (m *mockBookRepository) ListBooksWithoutEmbeddings(ctx context.Context, limit, offset int) ([]domain.Book, error) {
	if m.listBooksWithoutEmbeddingsFunc != nil {
		return m.listBooksWithoutEmbeddingsFunc(ctx, limit, offset)
	}
	return []domain.Book{}, nil
}

func (m *mockBookRepository) FilterByCriteria(ctx context.Context, filter domain.BookFilter, pagination domain.QueryOptions) ([]domain.Book, int64, error) {
	if m.filterByCriteriaFunc != nil {
		return m.filterByCriteriaFunc(ctx, filter, pagination)
	}
	return []domain.Book{}, 0, nil
}

func (m *mockBookRepository) QueryBooks(ctx context.Context, filter domain.BookFilter, pagination domain.QueryOptions) ([]domain.Book, int64, *string, bool, error) {
	if m.queryBooksFunc != nil {
		return m.queryBooksFunc(ctx, filter, pagination)
	}
	return []domain.Book{}, 0, nil, false, nil
}

func (m *mockBookRepository) Update(ctx context.Context, book *domain.Book) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, book)
	}
	return nil
}

func (m *mockBookRepository) UpdateWithRelations(ctx context.Context, id uuid.UUID, input domain.BookInput) (*domain.Book, error) {
	if m.updateWithRelationsFunc != nil {
		return m.updateWithRelationsFunc(ctx, id, input)
	}
	return nil, errors.New("not implemented")
}

func (m *mockBookRepository) ReplaceCategories(ctx context.Context, bookID uuid.UUID, categoryIDs []uuid.UUID) error {
	if m.replaceCategoriesFunc != nil {
		return m.replaceCategoriesFunc(ctx, bookID, categoryIDs)
	}
	return nil
}

func (m *mockBookRepository) UpsertEmbeddings(ctx context.Context, embeddings *domain.BookEmbedding) error {
	if m.upsertEmbeddingsFunc != nil {
		return m.upsertEmbeddingsFunc(ctx, embeddings)
	}
	return nil
}

func (m *mockBookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func TestBookServiceReadAndFilterPassThrough(t *testing.T) {
	bookID := uuid.New()
	expected := &domain.Book{ID: bookID, Name: "Book"}
	filter := domain.BookFilter{}
	query := domain.QueryOptions{Limit: 10, Offset: 5}

	repo := &mockBookRepository{
		findByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
			if id != bookID {
				t.Fatalf("unexpected ID: %s", id)
			}
			return expected, nil
		},
		listFunc: func(ctx context.Context, limit, offset int) ([]domain.Book, error) {
			if limit != 10 || offset != 20 {
				t.Fatalf("unexpected pagination: %d/%d", limit, offset)
			}
			return []domain.Book{*expected}, nil
		},
		filterByCriteriaFunc: func(ctx context.Context, gotFilter domain.BookFilter, pagination domain.QueryOptions) ([]domain.Book, int64, error) {
			if pagination.Limit != query.Limit || pagination.Offset != query.Offset {
				t.Fatalf("unexpected query options: %+v", pagination)
			}
			return []domain.Book{*expected}, 1, nil
		},
		deleteFunc: func(ctx context.Context, id uuid.UUID) error {
			if id != bookID {
				t.Fatalf("unexpected delete ID: %s", id)
			}
			return nil
		},
	}

	svc := NewBookService(repo, nil, nil)

	book, err := svc.GetBook(context.Background(), bookID)
	if err != nil || book.ID != bookID {
		t.Fatalf("unexpected GetBook result: %+v, err=%v", book, err)
	}

	books, err := svc.ListBooks(context.Background(), 10, 20)
	if err != nil || len(books) != 1 {
		t.Fatalf("unexpected ListBooks result: %+v, err=%v", books, err)
	}

	result, err := svc.FilterByCriteria(context.Background(), filter, query)
	if err != nil {
		t.Fatalf("unexpected filter error: %v", err)
	}
	if result.Total != 1 || result.Limit != query.Limit || result.Offset != query.Offset || len(result.Items) != 1 {
		t.Fatalf("unexpected filter result: %+v", result)
	}

	if err := svc.DeleteBook(context.Background(), bookID); err != nil {
		t.Fatalf("unexpected DeleteBook error: %v", err)
	}
}

func TestBookServiceCreateAndUpdatePassThrough(t *testing.T) {
	bookID := uuid.New()
	input := domain.BookInput{
		Name:        "Book",
		AuthorName:  "Author",
		PublisherID: uuid.New(),
	}
	createCalled := false
	updateCalled := false

	repo := &mockBookRepository{
		createWithRelationsFunc: func(ctx context.Context, got domain.BookInput) (*domain.Book, error) {
			createCalled = true
			return &domain.Book{ID: bookID, Name: got.Name}, nil
		},
		findByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
			if id != bookID {
				t.Fatalf("unexpected id")
			}
			return &domain.Book{ID: id, Name: input.Name}, nil
		},
		updateWithRelationsFunc: func(ctx context.Context, id uuid.UUID, got domain.BookInput) (*domain.Book, error) {
			updateCalled = true
			if id != bookID {
				t.Fatalf("unexpected id")
			}
			return &domain.Book{ID: id, Name: got.Name}, nil
		},
	}

	svc := NewBookService(repo, nil, nil)

	created, err := svc.CreateBook(context.Background(), input)
	if err != nil || created.ID != bookID {
		t.Fatalf("unexpected create result: %+v err=%v", created, err)
	}

	updated, err := svc.UpdateBook(context.Background(), bookID, input)
	if err != nil || updated.ID != bookID {
		t.Fatalf("unexpected update result: %+v err=%v", updated, err)
	}

	if !createCalled || !updateCalled {
		t.Fatalf("expected both create and update repository calls")
	}
}

func TestBookServiceQueryBooksPassThrough(t *testing.T) {
	nextCursor := "next-cursor"
	filter := domain.BookFilter{}
	query := domain.QueryOptions{Limit: 12, Offset: 3, Cursor: &nextCursor}
	repo := &mockBookRepository{
		queryBooksFunc: func(ctx context.Context, gotFilter domain.BookFilter, pagination domain.QueryOptions) ([]domain.Book, int64, *string, bool, error) {
			if pagination.Limit != query.Limit || pagination.Offset != query.Offset || pagination.Cursor == nil || *pagination.Cursor != nextCursor {
				t.Fatalf("unexpected query options: %+v", pagination)
			}
			return []domain.Book{{ID: uuid.New(), Name: "Book"}}, 1, &nextCursor, true, nil
		},
	}

	svc := NewBookService(repo, nil, nil)
	result, err := svc.QueryBooks(context.Background(), filter, query)
	if err != nil {
		t.Fatalf("unexpected query error: %v", err)
	}
	if result.Total != 1 || !result.HasMore || result.NextCursor == nil || *result.NextCursor != nextCursor {
		t.Fatalf("unexpected query result: %+v", result)
	}
}

type mockAIService struct {
	chatFunc  func(ctx context.Context, input domain.AIChatRequest) (*domain.AIChatResponse, error)
	embedFunc func(ctx context.Context, inputs []string) ([][]float64, error)
}

func (m *mockAIService) Chat(ctx context.Context, input domain.AIChatRequest) (*domain.AIChatResponse, error) {
	if m.chatFunc != nil {
		return m.chatFunc(ctx, input)
	}
	return nil, errors.New("not implemented")
}

func (m *mockAIService) Embed(ctx context.Context, inputs []string) ([][]float64, error) {
	if m.embedFunc != nil {
		return m.embedFunc(ctx, inputs)
	}
	return [][]float64{{1, 2, 3}}, nil
}

func TestBookServiceGenerateSummaryStoresResult(t *testing.T) {
	bookID := uuid.New()
	repo := &mockBookRepository{
		findByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
			if id != bookID {
				t.Fatalf("unexpected id: %s", id)
			}
			return &domain.Book{
				ID:          id,
				Name:        "Test Book",
				Description: "A test description.",
				Author:      domain.Author{Name: "Test Author"},
			}, nil
		},
		updateFunc: func(ctx context.Context, book *domain.Book) error {
			if book.ID != bookID {
				t.Fatalf("unexpected book id: %s", book.ID)
			}
			if book.Summary != "Generated summary." {
				t.Fatalf("expected summary to be stored, got %q", book.Summary)
			}
			return nil
		},
	}

	ai := &mockAIService{
		chatFunc: func(ctx context.Context, input domain.AIChatRequest) (*domain.AIChatResponse, error) {
			if input.Message == "" {
				t.Fatalf("expected prompt to be set")
			}
			return &domain.AIChatResponse{Message: "Generated summary."}, nil
		},
	}

	svc := NewBookService(repo, nil, nil, ai)
	got, err := svc.GenerateSummary(context.Background(), bookID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Summary != "Generated summary." {
		t.Fatalf("expected returned summary, got %q", got.Summary)
	}
}

func TestBookServiceGetBookGeneratesSummaryWhenMissing(t *testing.T) {
	bookID := uuid.New()
	repo := &mockBookRepository{
		findByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
			if id != bookID {
				t.Fatalf("unexpected id: %s", id)
			}
			return &domain.Book{
				ID:          id,
				Name:        "Test Book",
				Description: "A test description.",
				Summary:     "",
				Author:      domain.Author{Name: "Test Author"},
			}, nil
		},
		updateFunc: func(ctx context.Context, book *domain.Book) error {
			if book.ID != bookID {
				t.Fatalf("unexpected book id: %s", book.ID)
			}
			if book.Summary != "Generated summary." {
				t.Fatalf("expected summary to be stored, got %q", book.Summary)
			}
			return nil
		},
	}

	ai := &mockAIService{
		chatFunc: func(ctx context.Context, input domain.AIChatRequest) (*domain.AIChatResponse, error) {
			return &domain.AIChatResponse{Message: "Generated summary."}, nil
		},
	}

	svc := NewBookService(repo, nil, nil, ai)
	got, err := svc.GetBook(context.Background(), bookID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Summary != "Generated summary." {
		t.Fatalf("expected returned summary, got %q", got.Summary)
	}
}

func TestBookServiceGenerateEmbeddingsStoresVectors(t *testing.T) {
	bookID := uuid.New()
	repo := &mockBookRepository{
		findByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
			return &domain.Book{
				ID:          id,
				Name:        "Test Book",
				Description: "A description.",
				Summary:     "A summary.",
				Categories:  []domain.Category{{Name: "Fiction"}, {Name: "Adventure"}},
			}, nil
		},
	}

	embeddingSvc := &mockBookEmbeddingService{
		generateBookEmbeddingFunc: func(ctx context.Context, book domain.Book) error {
			if book.ID != bookID {
				t.Fatalf("unexpected book id: %s", book.ID)
			}
			if book.Name == "" || book.Description == "" {
				t.Fatalf("expected book fields to be present, got %+v", book)
			}
			return nil
		},
	}

	ai := &mockAIService{}
	svc := NewBookService(repo, nil, embeddingSvc, ai)
	got, err := svc.GenerateEmbeddings(context.Background(), bookID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != bookID {
		t.Fatalf("unexpected book returned: %+v", got)
	}
}
