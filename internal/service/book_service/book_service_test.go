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

type mockBookEmbeddingRepository struct {
	createEmbeddingFunc        func(ctx context.Context, embedding *domain.BookEmbedding) error
	updateEmbeddingFunc        func(ctx context.Context, embedding *domain.BookEmbedding) error
	getEmbeddingFunc           func(ctx context.Context, bookID uuid.UUID) (*domain.BookEmbedding, error)
	upsertEmbeddingFunc        func(ctx context.Context, embedding *domain.BookEmbedding) error
	getEmbeddingsByBookIDsFunc func(ctx context.Context, bookIDs []uuid.UUID) ([]domain.BookEmbedding, error)
	searchNearestBooksFunc     func(ctx context.Context, query domain.EmbeddingVector, limit int, excludeIDs []uuid.UUID) ([]domain.Book, error)
}

func (m *mockBookEmbeddingRepository) CreateEmbedding(ctx context.Context, embedding *domain.BookEmbedding) error {
	if m.createEmbeddingFunc != nil {
		return m.createEmbeddingFunc(ctx, embedding)
	}
	return nil
}

func (m *mockBookEmbeddingRepository) UpdateEmbedding(ctx context.Context, embedding *domain.BookEmbedding) error {
	if m.updateEmbeddingFunc != nil {
		return m.updateEmbeddingFunc(ctx, embedding)
	}
	return nil
}

func (m *mockBookEmbeddingRepository) GetEmbedding(ctx context.Context, bookID uuid.UUID) (*domain.BookEmbedding, error) {
	if m.getEmbeddingFunc != nil {
		return m.getEmbeddingFunc(ctx, bookID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockBookEmbeddingRepository) UpsertEmbedding(ctx context.Context, embedding *domain.BookEmbedding) error {
	if m.upsertEmbeddingFunc != nil {
		return m.upsertEmbeddingFunc(ctx, embedding)
	}
	return nil
}

func (m *mockBookEmbeddingRepository) GetEmbeddingsByBookIDs(ctx context.Context, bookIDs []uuid.UUID) ([]domain.BookEmbedding, error) {
	if m.getEmbeddingsByBookIDsFunc != nil {
		return m.getEmbeddingsByBookIDsFunc(ctx, bookIDs)
	}
	return nil, nil
}

func (m *mockBookEmbeddingRepository) SearchNearestBooks(ctx context.Context, query domain.EmbeddingVector, limit int, excludeIDs []uuid.UUID) ([]domain.Book, error) {
	if m.searchNearestBooksFunc != nil {
		return m.searchNearestBooksFunc(ctx, query, limit, excludeIDs)
	}
	return nil, errors.New("not implemented")
}

type mockOrderRepository struct {
	getPurchasedBookIDsFunc func(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

func (m *mockOrderRepository) CreateOrder(ctx context.Context, order *domain.Order) error {
	return nil
}
func (m *mockOrderRepository) CreateOrderItems(ctx context.Context, items []domain.OrderItem) error {
	return nil
}
func (m *mockOrderRepository) ListOrdersByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.OrderView, error) {
	return nil, nil
}
func (m *mockOrderRepository) ListOrders(ctx context.Context, limit, offset int) ([]domain.OrderView, error) {
	return nil, nil
}
func (m *mockOrderRepository) HasUserPurchasedBook(ctx context.Context, userID, bookID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockOrderRepository) GetOrderByID(ctx context.Context, orderID uuid.UUID) (domain.Order, error) {
	return domain.Order{}, nil
}
func (m *mockOrderRepository) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItemDetail, error) {
	return nil, nil
}
func (m *mockOrderRepository) UpdateOrderPayment(ctx context.Context, orderID uuid.UUID, status domain.PaymentStatus, method domain.PaymentMethod) error {
	return nil
}
func (m *mockOrderRepository) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status domain.OrderStatus, cancellationReason *string) error {
	return nil
}
func (m *mockOrderRepository) DecrementStock(ctx context.Context, items []domain.OrderItem) error {
	return nil
}
func (m *mockOrderRepository) GetPurchasedBookIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	if m.getPurchasedBookIDsFunc != nil {
		return m.getPurchasedBookIDsFunc(ctx, userID)
	}
	return nil, nil
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

	svc := NewBookService(repo, nil, nil, nil, nil)

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

	svc := NewBookService(repo, nil, nil, nil, nil)

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

	svc := NewBookService(repo, nil, nil, nil, nil)
	result, err := svc.QueryBooks(context.Background(), filter, query)
	if err != nil {
		t.Fatalf("unexpected query error: %v", err)
	}
	if result.Total != 1 || !result.HasMore || result.NextCursor == nil || *result.NextCursor != nextCursor {
		t.Fatalf("unexpected query result: %+v", result)
	}
}

type mockAIService struct {
	chatFunc     func(ctx context.Context, input domain.AIChatRequest, userID string) (*domain.AIChatResponse, error)
	generateFunc func(ctx context.Context, prompt string) (string, error)
	embedFunc    func(ctx context.Context, inputs []string) ([][]float64, error)
}

func (m *mockAIService) Chat(ctx context.Context, input domain.AIChatRequest, userID string) (*domain.AIChatResponse, error) {
	if m.chatFunc != nil {
		return m.chatFunc(ctx, input, userID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockAIService) Generate(ctx context.Context, prompt string) (string, error) {
	if m.generateFunc != nil {
		return m.generateFunc(ctx, prompt)
	}
	return "", errors.New("not implemented")
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
		generateFunc: func(ctx context.Context, prompt string) (string, error) {
			if prompt == "" {
				t.Fatalf("expected prompt to be set")
			}
			return "Generated summary.", nil
		},
	}

	svc := NewBookService(repo, nil, nil, nil, nil, ai)
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
		generateFunc: func(ctx context.Context, prompt string) (string, error) {
			return "Generated summary.", nil
		},
	}

	svc := NewBookService(repo, nil, nil, nil, nil, ai)
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
	svc := NewBookService(repo, nil, embeddingSvc, nil, nil, ai)
	got, err := svc.GenerateEmbeddings(context.Background(), bookID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != bookID {
		t.Fatalf("unexpected book returned: %+v", got)
	}
}

func TestRecommendBooksNoPurchases(t *testing.T) {
	userID := uuid.New()
	orderRepo := &mockOrderRepository{
		getPurchasedBookIDsFunc: func(ctx context.Context, uid uuid.UUID) ([]uuid.UUID, error) {
			if uid != userID {
				t.Fatalf("unexpected userID: %s", uid)
			}
			return []uuid.UUID{}, nil
		},
	}

	svc := NewBookService(&mockBookRepository{}, nil, nil, &mockBookEmbeddingRepository{}, orderRepo)
	books, err := svc.RecommendBooks(context.Background(), userID, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(books) != 0 {
		t.Fatalf("expected empty slice for user with no purchases, got %d items", len(books))
	}
}

func TestRecommendBooksNilDepsReturnsError(t *testing.T) {
	svc := NewBookService(&mockBookRepository{}, nil, nil, nil, nil)
	_, err := svc.RecommendBooks(context.Background(), uuid.New(), 10)
	if err == nil {
		t.Fatal("expected error when deps are nil")
	}
}

func TestRecommendBooksNoPurchasedEmbeddings(t *testing.T) {
	userID := uuid.New()
	bookID := uuid.New()

	orderRepo := &mockOrderRepository{
		getPurchasedBookIDsFunc: func(ctx context.Context, uid uuid.UUID) ([]uuid.UUID, error) {
			return []uuid.UUID{bookID}, nil
		},
	}
	embeddingRepo := &mockBookEmbeddingRepository{
		getEmbeddingsByBookIDsFunc: func(ctx context.Context, bookIDs []uuid.UUID) ([]domain.BookEmbedding, error) {
			return []domain.BookEmbedding{}, nil
		},
	}

	svc := NewBookService(&mockBookRepository{}, nil, nil, embeddingRepo, orderRepo)
	books, err := svc.RecommendBooks(context.Background(), userID, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(books) != 0 {
		t.Fatalf("expected empty slice when no embeddings exist, got %d items", len(books))
	}
}

func TestRecommendBooksHappyPath(t *testing.T) {
	userID := uuid.New()
	purchasedID1 := uuid.New()
	purchasedID2 := uuid.New()
	recommendedID := uuid.New()

	vec1 := domain.EmbeddingVector{1.0, 0.0}
	vec2 := domain.EmbeddingVector{0.0, 1.0}
	wantAvg := domain.EmbeddingVector{0.5, 0.5}

	orderRepo := &mockOrderRepository{
		getPurchasedBookIDsFunc: func(ctx context.Context, uid uuid.UUID) ([]uuid.UUID, error) {
			return []uuid.UUID{purchasedID1, purchasedID2}, nil
		},
	}
	embeddingRepo := &mockBookEmbeddingRepository{
		getEmbeddingsByBookIDsFunc: func(ctx context.Context, bookIDs []uuid.UUID) ([]domain.BookEmbedding, error) {
			if len(bookIDs) != 2 {
				t.Fatalf("expected 2 book IDs, got %d", len(bookIDs))
			}
			return []domain.BookEmbedding{
				{BookID: purchasedID1, Embedding: vec1},
				{BookID: purchasedID2, Embedding: vec2},
			}, nil
		},
		searchNearestBooksFunc: func(ctx context.Context, query domain.EmbeddingVector, limit int, excludeIDs []uuid.UUID) ([]domain.Book, error) {
			if len(query) != 2 || query[0] != wantAvg[0] || query[1] != wantAvg[1] {
				t.Fatalf("unexpected averaged query vector: %v", query)
			}
			if limit != 5 {
				t.Fatalf("unexpected limit: %d", limit)
			}
			if len(excludeIDs) != 2 {
				t.Fatalf("expected 2 exclude IDs, got %d", len(excludeIDs))
			}
			return []domain.Book{{ID: recommendedID, Name: "Recommended Book"}}, nil
		},
	}

	svc := NewBookService(&mockBookRepository{}, nil, nil, embeddingRepo, orderRepo)
	books, err := svc.RecommendBooks(context.Background(), userID, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(books) != 1 || books[0].ID != recommendedID {
		t.Fatalf("unexpected recommendations: %+v", books)
	}
}

func TestAverageEmbeddings(t *testing.T) {
	embeddings := []domain.BookEmbedding{
		{Embedding: domain.EmbeddingVector{1.0, 2.0, 3.0}},
		{Embedding: domain.EmbeddingVector{3.0, 4.0, 5.0}},
	}
	avg := averageEmbeddings(embeddings)
	want := domain.EmbeddingVector{2.0, 3.0, 4.0}
	for i := range want {
		if avg[i] != want[i] {
			t.Fatalf("avg[%d] = %f, want %f", i, avg[i], want[i])
		}
	}
}

func TestAverageEmbeddingsEmpty(t *testing.T) {
	if got := averageEmbeddings(nil); got != nil {
		t.Fatalf("expected nil for empty input, got %v", got)
	}
}
