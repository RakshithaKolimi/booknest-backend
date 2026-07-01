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

var (
	ErrProviderUnavailable = errors.New("AI provider is unavailable")
	ErrInvalidSessionID    = errors.New("invalid session_id")
	ErrForbiddenSession    = errors.New("chat session does not belong to user")
)

type aiService struct {
	provider aiprovider.Provider

	bookRepo      domain.BookRepository
	embeddingRepo domain.BookEmbeddingRepository
	orderRepo     domain.OrderRepository
	chatRepo      domain.ChatRepository

	intentHandlers map[string]intentHandler
}

type intentHandler func(ctx context.Context, toolCall domain.AIIntentToolCall, userID string) (*domain.AIChatResponse, error)

func NewAIService(
	provider aiprovider.Provider,
	bookRepo domain.BookRepository,
	embeddingRepo domain.BookEmbeddingRepository,
	orderRepo domain.OrderRepository,
	chatRepo ...domain.ChatRepository,
) domain.AIService {
	service := &aiService{
		provider:      provider,
		bookRepo:      bookRepo,
		embeddingRepo: embeddingRepo,
		orderRepo:     orderRepo,
	}
	if len(chatRepo) > 0 {
		service.chatRepo = chatRepo[0]
	}
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

	session, history, err := s.prepareChatSession(ctx, input, userID, message)
	if err != nil {
		return nil, err
	}

	intentInput := message
	if len(history) > 0 {
		intentInput = conversationPrompt(history, message)
	}

	toolCall, err := s.GetTool(ctx, intentInput)
	if err != nil {
		return nil, err
	}

	handler, ok := s.intentHandlers[toolCall.Tool]
	if !ok {
		return nil, fmt.Errorf("unknown intent: %s", toolCall.Tool)
	}
	if toolCall.Tool == string(domain.IntentChat) {
		toolCall.Query = message
		if len(history) > 0 {
			toolCall.Query = conversationPrompt(history, message)
		}
	}

	response, err := handler(ctx, toolCall, userID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return response, nil
	}

	response.SessionID = session.ID.String()
	if err := s.saveChatExchange(ctx, session.ID, message, response); err != nil {
		return nil, err
	}

	return response, nil
}

func (s *aiService) Generate(ctx context.Context, prompt string) (string, error) {
	if s.provider == nil {
		return "", ErrProviderUnavailable
	}

	response, err := s.provider.Generate(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("generate response: %w", err)
	}

	return response, nil
}

func (s *aiService) newIntentHandlers() map[string]intentHandler {
	return map[string]intentHandler{
		string(domain.IntentGetBook):            s.handleGetBook,
		string(domain.IntentGetBooksByCategory): s.handleCategorySearch,
		string(domain.IntentSemanticSearch):     s.handleSemanticSearch,
		string(domain.IntentRecommendation):     s.handleRecommendations,
		string(domain.IntentChat):               s.handleChat,
		string(domain.IntentNotRelated):         s.handleOutOfScope,
		string(domain.IntentSummary):            s.handleSummary,
	}
}

func chatMessage(input domain.AIChatRequest) string {
	if message := strings.TrimSpace(input.Message); message != "" {
		return message
	}
	return strings.TrimSpace(input.Prompt)
}

func (s *aiService) prepareChatSession(
	ctx context.Context,
	input domain.AIChatRequest,
	userID string,
	message string,
) (*domain.ChatSession, []domain.ChatMessage, error) {
	if s.chatRepo == nil || strings.TrimSpace(userID) == "" {
		return nil, nil, nil
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, nil, fmt.Errorf("parse user ID: %w", err)
	}

	sessionID := strings.TrimSpace(input.SessionID)
	if sessionID == "" {
		session, err := s.chatRepo.CreateSession(ctx, parsedUserID, chatSessionTitle(message))
		if err != nil {
			return nil, nil, fmt.Errorf("create chat session: %w", err)
		}
		return session, nil, nil
	}

	parsedSessionID, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrInvalidSessionID, err)
	}

	session, err := s.chatRepo.GetSession(ctx, parsedSessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("get chat session: %w", err)
	}
	if session.UserID != parsedUserID {
		return nil, nil, ErrForbiddenSession
	}

	messages, err := s.chatRepo.GetMessages(ctx, parsedSessionID, 20)
	if err != nil {
		return nil, nil, fmt.Errorf("get chat history: %w", err)
	}

	return session, messages, nil
}

func (s *aiService) saveChatExchange(ctx context.Context, sessionID uuid.UUID, userMessage string, response *domain.AIChatResponse) error {
	if err := s.chatRepo.SaveMessage(ctx, domain.ChatMessage{
		SessionID: sessionID,
		Role:      domain.ChatMessageRoleUser,
		Content:   userMessage,
	}); err != nil {
		return err
	}

	if response == nil || strings.TrimSpace(response.Message) == "" {
		return nil
	}

	referenceTitles := bookTitles(response.References)
	if err := s.chatRepo.SaveMessage(ctx, domain.ChatMessage{
		SessionID:       sessionID,
		Role:            domain.ChatMessageRoleAssistant,
		Content:         response.Message,
		ReferenceTitles: referenceTitles,
	}); err != nil {
		return err
	}

	return nil
}

func chatSessionTitle(message string) string {
	const maxTitleLength = 80

	title := strings.TrimSpace(message)
	if len(title) <= maxTitleLength {
		return title
	}
	return strings.TrimSpace(title[:maxTitleLength])
}

func conversationPrompt(history []domain.ChatMessage, message string) string {
	var builder strings.Builder
	builder.WriteString("Conversation so far:\n")
	for _, msg := range history {
		builder.WriteString(string(msg.Role))
		builder.WriteString(": ")
		builder.WriteString(msg.Content)
		if len(msg.ReferenceTitles) > 0 {
			builder.WriteString("\nreferences: ")
			builder.WriteString(strings.Join(msg.ReferenceTitles, ", "))
		}
		builder.WriteString("\n")
	}
	builder.WriteString("user: ")
	builder.WriteString(message)
	return builder.String()
}

func bookTitles(books []domain.Book) []string {
	if len(books) == 0 {
		return nil
	}

	titles := make([]string, 0, len(books))
	for _, book := range books {
		title := strings.TrimSpace(book.Name)
		if title != "" {
			titles = append(titles, title)
		}
	}
	if len(titles) == 0 {
		return nil
	}
	return titles
}

func (s *aiService) handleGetBook(ctx context.Context, toolCall domain.AIIntentToolCall, userID string) (*domain.AIChatResponse, error) {
	if s.bookRepo == nil {
		return nil, errors.New("book repository is not configured")
	}

	query := searchQuery(toolCall)
	books, _, err := s.bookRepo.FilterByCriteria(ctx, domain.BookFilter{
		Search: &toolCall.BookName,
	}, domain.QueryOptions{
		Limit: 5,
	})
	if err != nil {
		return nil, fmt.Errorf("search books: %w", err)
	}

	message := buildBookMessage(books, query)
	return buildBookResponse(
		books,
		message,
		query,
	), nil
}

func (s *aiService) handleCategorySearch(ctx context.Context, toolCall domain.AIIntentToolCall, userID string) (*domain.AIChatResponse, error) {
	query := searchQuery(toolCall)
	books, err := s.SemanticSearch(ctx, query, 5)
	if err != nil {
		return nil, fmt.Errorf("search books by category: %w", err)
	}

	message := buildBookMessage(books, query)
	return buildBookResponse(
		books,
		message,
		query,
	), nil
}

func (s *aiService) handleSemanticSearch(ctx context.Context, toolCall domain.AIIntentToolCall, userID string) (*domain.AIChatResponse, error) {
	query := strings.TrimSpace(toolCall.Query)
	books, err := s.SemanticSearch(ctx, query, 5)
	if err != nil {
		return nil, fmt.Errorf("semantic search books: %w", err)
	}

	message := buildBookMessage(books, query)
	return buildBookResponse(
		books,
		message,
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

func buildBookMessage(books []domain.Book, query string) string {
	if len(books) == 0 {
		return bookSearchEmptyMessage(query)
	}

	var parts []string
	for i, book := range books {
		// Add a decorative header for each book
		header := fmt.Sprintf("📚 **%s**", truncateString(book.Name, 60))
		author := fmt.Sprintf("👤 Author: %s", truncateString(book.Author.Name, 40))

		bookSection := header + "\n" + author

		// Add category if available
		if len(book.Categories) > 0 {
			categories := make([]string, len(book.Categories))
			for i, cat := range book.Categories {
				categories[i] = cat.Name
			}
			bookSection += fmt.Sprintf("\n🏷️ Category: %s", strings.Join(categories, ", "))
		}

		// Add summary for the first book with more detail, briefer for others
		if book.Summary != "" {
			if i == 0 {
				bookSection += fmt.Sprintf("\n\n📖 Summary:\n%s", formatSummary(book.Summary))
			} else {
				// For subsequent books, show a shorter summary
				shortSummary := truncateString(book.Summary, 150)
				if len(book.Summary) > 150 {
					shortSummary += "..."
				}
				bookSection += fmt.Sprintf("\n📖 Summary: %s", shortSummary)
			}
		}

		parts = append(parts, bookSection)
	}

	return strings.Join(parts, "\n\n---\n\n")
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

// truncateString truncates a string to the specified length, adding "..." if truncated.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// formatSummary formats and optionally truncates a book summary for display in chat responses.
func formatSummary(summary string) string {
	const maxLength = 2000

	if len(summary) <= maxLength {
		return summary
	}

	return fmt.Sprintf("%s...", summary[:maxLength])
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

func (s *aiService) handleSummary(ctx context.Context, toolCall domain.AIIntentToolCall, userID string) (*domain.AIChatResponse, error) {
	if s.bookRepo == nil {
		return nil, errors.New("book repository is not configured")
	}

	bookName := searchQuery(toolCall)

	if bookName == "" {
		return &domain.AIChatResponse{Message: "I couldn't find a specific book to summarize. Please provide a book title."}, nil
	}

	books, _, err := s.bookRepo.FilterByCriteria(ctx, domain.BookFilter{
		Search: &bookName,
	}, domain.QueryOptions{
		Limit: 1,
	})
	if err != nil {
		return nil, fmt.Errorf("search books: %w", err)
	}

	if len(books) == 0 {
		return &domain.AIChatResponse{Message: fmt.Sprintf("I couldn't find the book %q to summarize.", bookName)}, nil
	}

	book := books[0]
	if book.Summary == "" {
		return &domain.AIChatResponse{
			Message:    fmt.Sprintf("I found the book %q, but I don't have a summary available for it.", book.Name),
			References: []domain.Book{book},
		}, nil
	}

	response := fmt.Sprintf("Here's a summary of %q:\n\n%s", book.Name, book.Summary)
	return &domain.AIChatResponse{
		Message:    response,
		References: []domain.Book{book},
	}, nil
}
