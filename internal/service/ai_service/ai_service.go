package ai_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	aiprovider "booknest/internal/ai/provider"
	"booknest/internal/domain"

	"github.com/google/uuid"
)

var ErrProviderUnavailable = errors.New("AI provider is unavailable")

type aiService struct {
	provider aiprovider.Provider

	bookRepo      domain.BookRepository
	embeddingRepo domain.BookEmbeddingRepository
	orderRepo     domain.OrderRepository

	intentHandlers map[string]intentHandler
}

type intentHandler func(ctx context.Context, toolCall domain.AIIntentToolCall, userID string) (*domain.AIChatResponse, error)

func NewAIService(provider aiprovider.Provider, bookRepo domain.BookRepository, embeddingRepo domain.BookEmbeddingRepository, orderRepo domain.OrderRepository) domain.AIService {
	service := &aiService{provider: provider, bookRepo: bookRepo, embeddingRepo: embeddingRepo, orderRepo: orderRepo}
	service.intentHandlers = service.newIntentHandlers()
	return service
}

func (s *aiService) Chat(ctx context.Context, input domain.AIChatRequest, userID string) (*domain.AIChatResponse, error) {
	message := chatMessage(input)
	if message == "" {
		return nil, errors.New("message is required")
	}
	if s.provider == nil {
		return nil, ErrProviderUnavailable
	}

	toolCall, err := s.GetTool(ctx, message)
	if err != nil {
		return nil, err
	}

	handler, ok := s.intentHandlers[toolCall.Tool]
	if !ok {
		return nil, fmt.Errorf("unknown intent: %s", toolCall.Tool)
	}
	if toolCall.Tool == string(domain.IntentChat) {
		toolCall.Query = message
	}

	return handler(ctx, toolCall, userID)
}

func (s *aiService) newIntentHandlers() map[string]intentHandler {
	return map[string]intentHandler{
		string(domain.IntentGetBook):            s.handleGetBook,
		string(domain.IntentGetBooksByCategory): s.handleCategorySearch,
		string(domain.IntentSemanticSearch):     s.handleSemanticSearch,
		string(domain.IntentRecommendation):     s.handleRecommendations,
		string(domain.IntentChat):               s.handleChat,
		string(domain.IntentNotRelated):         s.handleOutOfScope,
	}
}

func chatMessage(input domain.AIChatRequest) string {
	if message := strings.TrimSpace(input.Message); message != "" {
		return message
	}
	return strings.TrimSpace(input.Prompt)
}

func (s *aiService) handleGetBook(ctx context.Context, toolCall domain.AIIntentToolCall, userID string) (*domain.AIChatResponse, error) {
	if s.bookRepo == nil {
		return nil, errors.New("book repository is not configured")
	}

	query := searchQuery(toolCall)
	books, _, err := s.bookRepo.FilterByCriteria(ctx, domain.BookFilter{
		Search: &query,
	}, domain.QueryOptions{
		Limit: 5,
	})
	if err != nil {
		return nil, fmt.Errorf("search books: %w", err)
	}

	return buildBookResponse(
		books,
		fmt.Sprintf("Here are the books I found for %q.", query),
		query,
	), nil
}

func (s *aiService) handleCategorySearch(ctx context.Context, toolCall domain.AIIntentToolCall, userID string) (*domain.AIChatResponse, error) {
	query := searchQuery(toolCall)
	books, err := s.SemanticSearch(ctx, query, 5)
	if err != nil {
		return nil, fmt.Errorf("search books by category: %w", err)
	}

	return buildBookResponse(
		books,
		fmt.Sprintf("Here are the books I found in %q.", query),
		query,
	), nil
}

func (s *aiService) handleSemanticSearch(ctx context.Context, toolCall domain.AIIntentToolCall, userID string) (*domain.AIChatResponse, error) {
	query := strings.TrimSpace(toolCall.Query)
	books, err := s.SemanticSearch(ctx, query, 5)
	if err != nil {
		return nil, fmt.Errorf("semantic search books: %w", err)
	}

	return buildBookResponse(
		books,
		fmt.Sprintf("Here are some books I found for %q.", query),
		query,
	), nil
}

func (s *aiService) handleRecommendations(ctx context.Context, toolCall domain.AIIntentToolCall, userID string) (*domain.AIChatResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("parse user ID: %w", err)
	}

	books, err := s.RecommendBooks(ctx, id, 5)
	if err != nil {
		return nil, fmt.Errorf("recommend books: %w", err)
	}

	return buildRecommendationResponse(books), nil
}

func (s *aiService) handleChat(ctx context.Context, toolCall domain.AIIntentToolCall, userID string) (*domain.AIChatResponse, error) {
	response, err := s.provider.Generate(ctx, toolCall.Query)
	if err != nil {
		return nil, fmt.Errorf("generate chat response: %w", err)
	}

	return &domain.AIChatResponse{Message: response}, nil
}

func (s *aiService) handleOutOfScope(ctx context.Context, toolCall domain.AIIntentToolCall, userID string) (*domain.AIChatResponse, error) {
	return &domain.AIChatResponse{Message: "> 🚫 This request is outside my scope.\n\n**I can help with:**\n- Book recommendations\n- Book discovery\n- Categories and genres\n- Author information\n- Book details\n\n💡 **Examples**\n- Recommend sci-fi books\n- Show romance books\n- Tell me about The Hobbit\n- Suggest books similar to Dune"}, nil
}

func searchQuery(toolCall domain.AIIntentToolCall) string {
	if query := strings.TrimSpace(toolCall.Query); query != "" {
		return query
	}
	if category := strings.TrimSpace(toolCall.Category); category != "" {
		return category
	}
	return strings.TrimSpace(toolCall.BookName)
}

func buildBookResponse(books []domain.Book, successMessage string, query string) *domain.AIChatResponse {
	if len(books) == 0 {
		return &domain.AIChatResponse{Message: bookSearchEmptyMessage(query)}
	}

	return &domain.AIChatResponse{
		Message:    successMessage,
		References: books,
	}
}

func buildRecommendationResponse(books []domain.Book) *domain.AIChatResponse {
	if len(books) == 0 {
		return &domain.AIChatResponse{Message: "I couldn't generate recommendations yet.\n\n💡 Try rating or interacting with more books."}
	}

	return &domain.AIChatResponse{
		Message:    "Here are some book recommendations for you.",
		References: books,
	}
}

func bookSearchEmptyMessage(query string) string {
	if query == "" {
		return "I couldn't find any books matching your request.\n\n💡 Try:\n- Different keywords\n- Another author\n- A genre name\n- A partial title"
	}

	return fmt.Sprintf("I couldn't find any books matching %q.\n\n💡 Try:\n- Different keywords\n- Another author\n- A genre name\n- A partial title", query)
}

func (s *aiService) Embed(ctx context.Context, inputs []string) ([][]float64, error) {
	if len(inputs) == 0 {
		return nil, errors.New("inputs are required")
	}
	if s.provider == nil {
		return nil, ErrProviderUnavailable
	}

	for _, input := range inputs {
		if strings.TrimSpace(input) == "" {
			return nil, errors.New("inputs must not contain empty strings")
		}
	}

	vectors, err := s.provider.Embed(ctx, inputs)
	if err != nil {
		return nil, err
	}

	return vectors, nil
}

func (s *aiService) GetTool(ctx context.Context, input string) (domain.AIIntentToolCall, error) {
	if input == "" {
		return domain.AIIntentToolCall{}, errors.New("input is required")
	}
	if s.provider == nil {
		return domain.AIIntentToolCall{}, ErrProviderUnavailable
	}

	prompt := fmt.Sprintf(IntentDetectionPrompt, input)
	response, err := s.provider.Generate(ctx, prompt)
	if err != nil {
		return domain.AIIntentToolCall{}, err
	}

	bytes := []byte(response)
	var toolCall domain.AIIntentToolCall
	if err := json.Unmarshal(bytes, &toolCall); err != nil {
		return domain.AIIntentToolCall{}, fmt.Errorf("parse provider response: %w", err)
	}

	return toolCall, nil
}

func (s *aiService) SemanticSearch(ctx context.Context, query string, limit int) ([]domain.Book, error) {
	if s.provider == nil || s.embeddingRepo == nil {
		return nil, ErrProviderUnavailable
	}
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("query is required")
	}
	if limit <= 0 {
		limit = 10
	}

	vectors, err := s.provider.Embed(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	if len(vectors) != 1 || len(vectors[0]) == 0 {
		return nil, errors.New("unexpected embedding response")
	}

	return s.embeddingRepo.SearchNearestBooks(ctx, domain.EmbeddingVector(vectors[0]), limit, nil)
}

func (s *aiService) RecommendBooks(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Book, error) {
	if s.embeddingRepo == nil || s.orderRepo == nil {
		return nil, errors.New("recommendation dependencies are not configured")
	}
	if limit <= 0 {
		limit = 10
	}

	purchasedIDs, err := s.orderRepo.GetPurchasedBookIDs(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(purchasedIDs) == 0 {
		return []domain.Book{}, nil
	}

	embeddings, err := s.embeddingRepo.GetEmbeddingsByBookIDs(ctx, purchasedIDs)
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return []domain.Book{}, nil
	}

	avgVector := averageEmbeddings(embeddings)
	return s.embeddingRepo.SearchNearestBooks(ctx, avgVector, limit, purchasedIDs)
}

// averageEmbeddings computes the element-wise average of the given book embeddings to create a single embedding vector representing the user's overall preferences.
func averageEmbeddings(embeddings []domain.BookEmbedding) domain.EmbeddingVector {
	if len(embeddings) == 0 {
		return nil
	}
	dim := len(embeddings[0].Embedding)
	avg := make(domain.EmbeddingVector, dim)
	for _, emb := range embeddings {
		for i, v := range emb.Embedding {
			if i < dim {
				avg[i] += v
			}
		}
	}
	n := float64(len(embeddings))
	for i := range avg {
		avg[i] /= n
	}
	return avg
}
