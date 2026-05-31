package book_service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"booknest/internal/domain"
	"booknest/internal/service/ai_service"
)

type bookService struct {
	repo         domain.BookRepository
	categoryRepo domain.CategoryRepository
	ai           domain.AIService
}

func NewBookService(
	repo domain.BookRepository,
	categoryRepo domain.CategoryRepository,
	ai ...domain.AIService,
) domain.BookService {
	var aiSvc domain.AIService
	if len(ai) > 0 {
		aiSvc = ai[0]
	}
	return &bookService{
		repo:         repo,
		categoryRepo: categoryRepo,
		ai:           aiSvc,
	}
}

func (s *bookService) CreateBook(ctx context.Context, input domain.BookInput) (*domain.Book, error) {
	book, err := s.repo.CreateWithRelations(ctx, input)
	if err != nil {
		return nil, err
	}

	updated := book

	// Best-effort: don't block book creation on AI availability.
	if summaryUpdated, err := s.GenerateSummary(ctx, book.ID); err == nil {
		updated = summaryUpdated
	}

	// Best-effort: if no categories were provided, try generating them.
	if len(input.CategoryIDs) == 0 && s.ai != nil {
		if categoriesUpdated, err := s.GenerateCategories(ctx, book.ID); err == nil {
			updated = categoriesUpdated
		}
	}

	return updated, nil
}

func (s *bookService) GetBook(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	book, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(book.Summary) == "" && s.ai != nil {
		if updated, err := s.generateAndStoreSummary(ctx, book); err == nil {
			return updated, nil
		}
	}

	if len(book.Categories) == 0 && s.ai != nil {
		if updated, err := s.generateAndStoreCategories(ctx, book); err == nil {
			return updated, nil
		}
	}

	return book, nil
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
	book, err := s.repo.UpdateWithRelations(ctx, id, input)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(book.Summary) == "" && s.ai != nil {
		// Best-effort: keep serving updates even if AI is down/unconfigured.
		if updated, err := s.GenerateSummary(ctx, book.ID); err == nil {
			return updated, nil
		}
	}

	if len(input.CategoryIDs) == 0 && len(book.Categories) == 0 && s.ai != nil {
		// Best-effort: only generate categories if none were provided.
		if updated, err := s.GenerateCategories(ctx, book.ID); err == nil {
			return updated, nil
		}
	}

	return book, nil
}

func (s *bookService) GenerateSummary(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	if s.ai == nil {
		return nil, ai_service.ErrProviderUnavailable
	}

	book, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.generateAndStoreSummary(ctx, book)
}

func (s *bookService) GenerateCategories(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	if s.ai == nil {
		return nil, ai_service.ErrProviderUnavailable
	}
	if s.categoryRepo == nil {
		return nil, errors.New("category repository is unavailable")
	}

	book, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.generateAndStoreCategories(ctx, book)
}

func (s *bookService) DeleteBook(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *bookService) generateAndStoreSummary(ctx context.Context, book *domain.Book) (*domain.Book, error) {
	if s.ai == nil {
		return nil, ai_service.ErrProviderUnavailable
	}

	author := strings.TrimSpace(book.Author.Name)
	title := strings.TrimSpace(book.Name)
	description := strings.TrimSpace(book.Description)

	prompt := buildSummaryPrompt(title, author, description)
	resp, err := s.ai.Chat(ctx, domain.AIChatRequest{Message: prompt})
	if err != nil {
		return nil, err
	}

	book.Summary = strings.TrimSpace(resp.Message)
	if err := s.repo.Update(ctx, book); err != nil {
		return nil, err
	}
	return book, nil
}

func (s *bookService) generateAndStoreCategories(ctx context.Context, book *domain.Book) (*domain.Book, error) {
	if s.ai == nil {
		return nil, ai_service.ErrProviderUnavailable
	}
	if s.categoryRepo == nil {
		return nil, errors.New("category repository is unavailable")
	}

	author := strings.TrimSpace(book.Author.Name)
	title := strings.TrimSpace(book.Name)
	description := strings.TrimSpace(book.Description)
	summary := strings.TrimSpace(book.Summary)

	prompt := buildCategoriesPrompt(title, author, description, summary)
	resp, err := s.ai.Chat(ctx, domain.AIChatRequest{Message: prompt})
	if err != nil {
		return nil, err
	}

	categoryNames, err := parseCategoryNames(resp.Message)
	if err != nil {
		return nil, err
	}

	categoryIDs := make([]uuid.UUID, 0, len(categoryNames))
	for _, name := range categoryNames {
		if strings.TrimSpace(name) == "" {
			continue
		}

		existing, err := s.categoryRepo.FindByName(ctx, name)
		if err == nil {
			categoryIDs = append(categoryIDs, existing.ID)
			continue
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		category := &domain.Category{
			ID:   uuid.New(),
			Name: name,
		}
		if err := s.categoryRepo.Create(ctx, category); err != nil {
			return nil, err
		}
		categoryIDs = append(categoryIDs, category.ID)
	}

	if err := s.repo.ReplaceCategories(ctx, book.ID, categoryIDs); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, book.ID)
}

func buildSummaryPrompt(title, author, description string) string {
	var b strings.Builder
	b.WriteString("Write a concise book summary for a bookstore listing.\n")
	b.WriteString("Rules: 2-3 sentences, plain text, no spoilers, no quotes, no markdown.\n")
	b.WriteString("Title: ")
	b.WriteString(title)
	b.WriteString("\nAuthor: ")
	b.WriteString(author)
	b.WriteString("\nDescription: ")
	b.WriteString(description)
	return b.String()
}

func buildCategoriesPrompt(title, author, description, summary string) string {
	var b strings.Builder
	b.WriteString("Create 5-10 concise bookstore categories for this book.\n")
	b.WriteString("Rules: return ONLY a valid JSON array of strings. No markdown, no extra text.\n")
	b.WriteString("Each category: 2-30 chars, Title Case where appropriate, no duplicates.\n")
	b.WriteString("Use broad shelf categories (e.g., Fiction, Mystery, Self-Help), not plot details.\n")
	b.WriteString("Title: ")
	b.WriteString(title)
	b.WriteString("\nAuthor: ")
	b.WriteString(author)
	b.WriteString("\nDescription: ")
	b.WriteString(description)
	if summary != "" {
		b.WriteString("\nSummary: ")
		b.WriteString(summary)
	}
	return b.String()
}

func parseCategoryNames(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("empty AI response")
	}

	// Be forgiving if the model adds extra text: extract the first JSON array.
	start := strings.Index(raw, "[")
	end := strings.LastIndex(raw, "]")
	if start >= 0 && end > start {
		raw = raw[start : end+1]
	}

	var names []string
	if err := json.Unmarshal([]byte(raw), &names); err != nil {
		return nil, err
	}

	seen := make(map[string]struct{}, len(names))
	out := make([]string, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if len(name) > 30 {
			name = strings.TrimSpace(name[:30])
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, name)
	}

	if len(out) < 1 {
		return nil, errors.New("no categories parsed from AI response")
	}

	if len(out) > 10 {
		out = out[:10]
	}

	return out, nil
}
