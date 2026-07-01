package ai_service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"booknest/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type mockProvider struct {
	prompt  string
	prompts []string
	replies []string
	reply   string
	err     error
	inputs  [][]string
	vectors [][]float64
}

func (m *mockProvider) Generate(ctx context.Context, prompt string) (string, error) {
	m.prompt = prompt
	m.prompts = append(m.prompts, prompt)
	if m.err != nil {
		return "", m.err
	}
	if len(m.replies) > 0 {
		reply := m.replies[0]
		m.replies = m.replies[1:]
		return reply, nil
	}
	return m.reply, nil
}

func (m *mockProvider) Embed(ctx context.Context, inputs []string) ([][]float64, error) {
	m.inputs = append(m.inputs, append([]string(nil), inputs...))
	if m.err != nil {
		return nil, m.err
	}
	if m.vectors != nil {
		return m.vectors, nil
	}
	return [][]float64{{1, 2, 3}}, nil
}

type mockBookEmbeddingRepository struct {
	books      []domain.Book
	embeddings []domain.BookEmbedding
	query      domain.EmbeddingVector
	limit      int
	excludeIDs []uuid.UUID
}

func (m *mockBookEmbeddingRepository) Create(ctx context.Context, book *domain.Book) error {
	return nil
}

func (m *mockBookEmbeddingRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	for _, b := range m.books {
		if b.ID == id {
			return &b, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockBookEmbeddingRepository) List(ctx context.Context, limit, offset int) ([]domain.Book, error) {
	return nil, nil
}

func (m *mockBookEmbeddingRepository) UpsertEmbedding(ctx context.Context, embedding *domain.BookEmbedding) error {
	return nil
}

func (m *mockBookEmbeddingRepository) GetEmbeddingsByBookIDs(ctx context.Context, bookIDs []uuid.UUID) ([]domain.BookEmbedding, error) {
	return m.embeddings, nil
}

func (m *mockBookEmbeddingRepository) SearchNearestBooks(ctx context.Context, query domain.EmbeddingVector, limit int, excludeIDs []uuid.UUID) ([]domain.Book, error) {
	m.query = query
	m.limit = limit
	m.excludeIDs = append([]uuid.UUID(nil), excludeIDs...)
	return m.books, nil
}

func (m *mockBookEmbeddingRepository) CreateEmbedding(ctx context.Context, embedding *domain.BookEmbedding) error {
	return nil
}

func (m *mockBookEmbeddingRepository) UpdateEmbedding(ctx context.Context, embedding *domain.BookEmbedding) error {
	return nil
}

func (m *mockBookEmbeddingRepository) GetEmbedding(ctx context.Context, bookID uuid.UUID) (*domain.BookEmbedding, error) {
	for _, e := range m.embeddings {
		if e.BookID == bookID {
			return &e, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

type mockOrderRepository struct {
	purchasedIDs []uuid.UUID
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
	return m.purchasedIDs, nil
}

type mockBookRepository struct {
	books []domain.Book
}

func (m *mockBookRepository) Create(ctx context.Context, book *domain.Book) error {
	return nil
}

func (m *mockBookRepository) CreateWithRelations(ctx context.Context, input domain.BookInput) (*domain.Book, error) {
	return nil, nil
}

func (m *mockBookRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	return nil, nil
}

func (m *mockBookRepository) ListBooksWithoutEmbeddings(ctx context.Context, limit, offset int) ([]domain.Book, error) {
	return nil, nil
}

func (m *mockBookRepository) List(ctx context.Context, limit, offset int) ([]domain.Book, error) {
	return nil, nil
}

func (m *mockBookRepository) FilterByCriteria(ctx context.Context, filter domain.BookFilter, pagination domain.QueryOptions) ([]domain.Book, int64, error) {
	return m.books, int64(len(m.books)), nil
}

func (m *mockBookRepository) QueryBooks(ctx context.Context, filter domain.BookFilter, pagination domain.QueryOptions) ([]domain.Book, int64, *string, bool, error) {
	return nil, 0, nil, false, nil
}

func (m *mockBookRepository) Update(ctx context.Context, book *domain.Book) error {
	return nil
}

func (m *mockBookRepository) UpdateWithRelations(ctx context.Context, id uuid.UUID, input domain.BookInput) (*domain.Book, error) {
	return nil, nil
}

func (m *mockBookRepository) ReplaceCategories(ctx context.Context, bookID uuid.UUID, categoryIDs []uuid.UUID) error {
	return nil
}

func (m *mockBookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

type mockChatRepository struct {
	session       *domain.ChatSession
	messages      []domain.ChatMessage
	savedMessages []domain.ChatMessage
	createdTitle  string
}

func (m *mockChatRepository) CreateSession(ctx context.Context, userID uuid.UUID, title string) (*domain.ChatSession, error) {
	m.createdTitle = title
	if m.session == nil {
		m.session = &domain.ChatSession{
			ID:        uuid.New(),
			UserID:    userID,
			Title:     title,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}
	return m.session, nil
}

func (m *mockChatRepository) SaveMessage(ctx context.Context, msg domain.ChatMessage) error {
	m.savedMessages = append(m.savedMessages, msg)
	return nil
}

func (m *mockChatRepository) GetMessages(ctx context.Context, sessionID uuid.UUID, limit int) ([]domain.ChatMessage, error) {
	return m.messages, nil
}

func (m *mockChatRepository) GetSession(ctx context.Context, sessionID uuid.UUID) (*domain.ChatSession, error) {
	return m.session, nil
}

func TestAIServiceChat(t *testing.T) {
	provider := &mockProvider{replies: []string{`{"tool":"chat"}`, "A useful answer."}}
	service := NewAIService(provider, nil, nil, nil)

	response, err := service.Chat(context.Background(), domain.AIChatRequest{Message: " hello "}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if response.Message != "A useful answer." {
		t.Fatalf("unexpected response: %+v", response)
	}
	if len(provider.prompts) != 2 || provider.prompts[1] != "hello" {
		t.Fatalf("expected trimmed chat prompt, got %+v", provider.prompts)
	}
}

func TestAIServiceChatCreatesSessionAndSavesExchange(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	provider := &mockProvider{replies: []string{`{"tool":"chat"}`, "A useful answer."}}
	chatRepo := &mockChatRepository{
		session: &domain.ChatSession{ID: sessionID, UserID: userID},
	}
	service := NewAIService(provider, nil, nil, nil, chatRepo)

	response, err := service.Chat(context.Background(), domain.AIChatRequest{Message: "recommend fantasy books"}, userID.String())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if response.SessionID != sessionID.String() {
		t.Fatalf("expected session ID %s, got %q", sessionID, response.SessionID)
	}
	if chatRepo.createdTitle != "recommend fantasy books" {
		t.Fatalf("expected session title from first message, got %q", chatRepo.createdTitle)
	}
	if len(chatRepo.savedMessages) != 2 {
		t.Fatalf("expected user and assistant messages to be saved, got %+v", chatRepo.savedMessages)
	}
	if chatRepo.savedMessages[0].Role != domain.ChatMessageRoleUser || chatRepo.savedMessages[0].Content != "recommend fantasy books" {
		t.Fatalf("unexpected saved user message: %+v", chatRepo.savedMessages[0])
	}
	if chatRepo.savedMessages[1].Role != domain.ChatMessageRoleAssistant || chatRepo.savedMessages[1].Content != "A useful answer." {
		t.Fatalf("unexpected saved assistant message: %+v", chatRepo.savedMessages[1])
	}
}

func TestAIServiceChatUsesExistingSessionHistory(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	provider := &mockProvider{replies: []string{`{"tool":"chat"}`, "Try The Blade Itself."}}
	chatRepo := &mockChatRepository{
		session: &domain.ChatSession{ID: sessionID, UserID: userID},
		messages: []domain.ChatMessage{
			{SessionID: sessionID, Role: domain.ChatMessageRoleUser, Content: "recommend fantasy books"},
			{SessionID: sessionID, Role: domain.ChatMessageRoleAssistant, Content: "Try The Hobbit."},
		},
	}
	service := NewAIService(provider, nil, nil, nil, chatRepo)

	response, err := service.Chat(context.Background(), domain.AIChatRequest{
		SessionID: sessionID.String(),
		Message:   "something darker",
	}, userID.String())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if response.SessionID != sessionID.String() {
		t.Fatalf("expected session ID %s, got %q", sessionID, response.SessionID)
	}
	if len(provider.prompts) != 2 {
		t.Fatalf("expected intent and chat prompts, got %+v", provider.prompts)
	}
	for _, want := range []string{
		"Conversation so far:",
		"user: recommend fantasy books",
		"assistant: Try The Hobbit.",
		"user: something darker",
	} {
		if !strings.Contains(provider.prompts[1], want) {
			t.Fatalf("expected chat prompt to contain %q, got %q", want, provider.prompts[1])
		}
	}
	if len(chatRepo.savedMessages) != 2 || chatRepo.savedMessages[0].Content != "something darker" {
		t.Fatalf("expected follow-up exchange to be saved, got %+v", chatRepo.savedMessages)
	}
}

func TestAIServiceChatUsesPromptAlias(t *testing.T) {
	provider := &mockProvider{replies: []string{`{"tool":"chat"}`, "A useful answer."}}
	service := NewAIService(provider, nil, nil, nil)

	_, err := service.Chat(context.Background(), domain.AIChatRequest{Prompt: "legacy prompt"}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(provider.prompts) != 2 || provider.prompts[1] != "legacy prompt" {
		t.Fatalf("expected prompt alias, got %+v", provider.prompts)
	}
}

func TestAIServiceGenerate(t *testing.T) {
	provider := &mockProvider{reply: "Generated summary."}
	service := NewAIService(provider, nil, nil, nil)

	prompt := "Write a concise book summary for a bookstore listing.\nTitle: Dune\nAuthor: Frank Herbert\nDescription: Desert planet politics."
	response, err := service.Generate(context.Background(), prompt)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if response != "Generated summary." {
		t.Fatalf("unexpected response: %q", response)
	}
	if len(provider.prompts) != 1 || provider.prompts[0] != prompt {
		t.Fatalf("expected direct generation prompt, got %+v", provider.prompts)
	}
}

func TestAIServiceChatValidation(t *testing.T) {
	service := NewAIService(&mockProvider{}, nil, nil, nil)

	_, err := service.Chat(context.Background(), domain.AIChatRequest{}, "")
	if err == nil || err.Error() != "message is required" {
		t.Fatalf("expected message validation error, got %v", err)
	}
}

func TestAIServiceChatProviderUnavailable(t *testing.T) {
	service := NewAIService(nil, nil, nil, nil)

	_, err := service.Chat(context.Background(), domain.AIChatRequest{Message: "hello"}, "")
	if !errors.Is(err, ErrProviderUnavailable) {
		t.Fatalf("expected provider unavailable error, got %v", err)
	}
}

func TestAIServiceChatSemanticSearchReturnsReferences(t *testing.T) {
	book := domain.Book{ID: uuid.New(), Name: "Dune", Author: domain.Author{Name: "Frank Herbert"}}
	provider := &mockProvider{
		replies: []string{`{"tool":"semantic_search","query":"desert politics"}`},
		vectors: [][]float64{{0.1, 0.2}},
	}
	embeddingRepo := &mockBookEmbeddingRepository{books: []domain.Book{book}}
	service := NewAIService(provider, nil, embeddingRepo, nil)

	response, err := service.Chat(context.Background(), domain.AIChatRequest{Message: "books about desert politics"}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	expectedMsg := `- Dune by Frank Herbert`
	if response.Message != expectedMsg {
		t.Fatalf("unexpected response message:\ngot:  %q\nwant:%s", response.Message, expectedMsg)
	}
	if len(response.References) != 1 || response.References[0].ID != book.ID {
		t.Fatalf("expected semantic search references, got %+v", response.References)
	}
	if embeddingRepo.limit != 5 {
		t.Fatalf("expected semantic search limit 5, got %d", embeddingRepo.limit)
	}
}

func TestAIServiceChatRecommendationUsesUserHistory(t *testing.T) {
	userID := uuid.New()
	purchasedID := uuid.New()
	book := domain.Book{ID: uuid.New(), Name: "The Left Hand of Darkness"}
	provider := &mockProvider{replies: []string{`{"tool":"recommendation"}`}}
	embeddingRepo := &mockBookEmbeddingRepository{
		books: []domain.Book{book},
		embeddings: []domain.BookEmbedding{
			{BookID: purchasedID, Embedding: domain.EmbeddingVector{1, 3}},
		},
	}
	orderRepo := &mockOrderRepository{purchasedIDs: []uuid.UUID{purchasedID}}
	service := NewAIService(provider, nil, embeddingRepo, orderRepo)

	response, err := service.Chat(context.Background(), domain.AIChatRequest{Message: "recommend books for me"}, userID.String())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if response.Message != "Here are some book recommendations for you." {
		t.Fatalf("unexpected response message: %q", response.Message)
	}
	if len(response.References) != 1 || response.References[0].ID != book.ID {
		t.Fatalf("expected recommendation references, got %+v", response.References)
	}
	if len(embeddingRepo.excludeIDs) != 1 || embeddingRepo.excludeIDs[0] != purchasedID {
		t.Fatalf("expected purchased book to be excluded, got %+v", embeddingRepo.excludeIDs)
	}
}

func TestAIServiceChatGetBookEmptyState(t *testing.T) {
	provider := &mockProvider{replies: []string{`{"tool":"get_book","query":"Dune"}`}}
	service := NewAIService(provider, &mockBookRepository{}, nil, nil)

	response, err := service.Chat(context.Background(), domain.AIChatRequest{Message: "find Dune"}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := "I couldn't find any books matching \"Dune\".\n\n💡 Try:\n- Different keywords\n- Another author\n- A genre name\n- A partial title"
	if response.Message != expected {
		t.Fatalf("unexpected empty response: %q", response.Message)
	}
	if len(response.References) != 0 {
		t.Fatalf("expected no references, got %+v", response.References)
	}
}

func TestAIServiceChatSemanticSearchEmptyState(t *testing.T) {
	provider := &mockProvider{
		replies: []string{`{"tool":"semantic_search","query":"desert politics"}`},
		vectors: [][]float64{{0.1, 0.2}},
	}
	service := NewAIService(provider, nil, &mockBookEmbeddingRepository{}, nil)

	response, err := service.Chat(context.Background(), domain.AIChatRequest{Message: "books about desert politics"}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := "I couldn't find any books matching \"desert politics\".\n\n💡 Try:\n- Different keywords\n- Another author\n- A genre name\n- A partial title"
	if response.Message != expected {
		t.Fatalf("unexpected empty response: %q", response.Message)
	}
}

func TestAIServiceChatRecommendationEmptyState(t *testing.T) {
	userID := uuid.New()
	provider := &mockProvider{replies: []string{`{"tool":"recommendation"}`}}
	orderRepo := &mockOrderRepository{}
	service := NewAIService(provider, nil, &mockBookEmbeddingRepository{}, orderRepo)

	response, err := service.Chat(context.Background(), domain.AIChatRequest{Message: "recommend books for me"}, userID.String())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := "I couldn't generate recommendations yet.\n\n💡 Try rating or interacting with more books."
	if response.Message != expected {
		t.Fatalf("unexpected empty response: %q", response.Message)
	}
}

func TestAIServiceEmbed(t *testing.T) {
	provider := &mockProvider{vectors: [][]float64{{1, 2}, {3, 4}}}
	service := NewAIService(provider, nil, nil, nil)

	vectors, err := service.Embed(context.Background(), []string{" title ", "description"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(vectors) != 2 {
		t.Fatalf("expected 2 vectors, got %d", len(vectors))
	}
	if len(provider.inputs) != 1 || provider.inputs[0][0] != " title " {
		t.Fatalf("expected provider to receive original inputs, got %+v", provider.inputs)
	}
}

func TestAIServiceEmbedValidation(t *testing.T) {
	service := NewAIService(&mockProvider{}, nil, nil, nil)

	_, err := service.Embed(context.Background(), []string{"valid", " "})
	if err == nil || err.Error() != "inputs must not contain empty strings" {
		t.Fatalf("expected validation error, got %v", err)
	}
}
