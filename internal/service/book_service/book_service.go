package book_service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"booknest/internal/domain"
	"booknest/internal/service/ai_service"
)

type bookService struct {
	repo          domain.BookRepository
	categoryRepo  domain.CategoryRepository
	ai            domain.AIService
	embeddingSvc  domain.BookEmbeddingService
	embeddingRepo domain.BookEmbeddingRepository
	orderRepo     domain.OrderRepository
	embeddingJobs sync.Map
}

func NewBookService(
	repo domain.BookRepository,
	categoryRepo domain.CategoryRepository,
	embeddingSvc domain.BookEmbeddingService,
	embeddingRepo domain.BookEmbeddingRepository,
	orderRepo domain.OrderRepository,
	ai ...domain.AIService,
) domain.BookService {
	var aiSvc domain.AIService
	if len(ai) > 0 {
		aiSvc = ai[0]
	}
	svc := &bookService{
		repo:          repo,
		categoryRepo:  categoryRepo,
		ai:            aiSvc,
		embeddingSvc:  embeddingSvc,
		embeddingRepo: embeddingRepo,
		orderRepo:     orderRepo,
	}

	// Start a background sync to generate embeddings for existing books
	// that don't have embeddings yet.
	if aiSvc != nil && embeddingSvc != nil {
		svc.startEmbeddingSyncOnStartup()
	}

	return svc
}

// startEmbeddingSyncOnStartup will scan existing books without embeddings
// and generate embeddings for them. It's started automatically when the
// service is created and an AI provider is available.
func (s *bookService) startEmbeddingSyncOnStartup() {
	if s.ai == nil || s.embeddingSvc == nil {
		return
	}

	go func() {
		slog.Info("starting startup embedding sync")
		defer slog.Info("startup embedding sync completed")

		const pageSize = 50
		offset := 0
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			books, err := s.repo.ListBooksWithoutEmbeddings(ctx, pageSize, offset)
			cancel()
			if err != nil {
				slog.Warn("failed listing books without embeddings", "error", err)
				return
			}
			if len(books) == 0 {
				return
			}

			for _, b := range books {
				// Best-effort: generate embeddings but don't stop the loop on error.
				if err := s.generateAndStoreEmbeddings(context.Background(), &b); err != nil {
					if errors.Is(err, ai_service.ErrProviderUnavailable) {
						slog.Warn("AI provider unavailable during startup embedding sync")
						return
					}
					slog.Warn("failed to generate embeddings for book", "book_id", b.ID.String(), "error", err)
				} else {
					slog.Info("generated embeddings for book", "book_id", b.ID.String())
				}
			}

			if len(books) < pageSize {
				return
			}
			offset += pageSize
		}
	}()
}

func (s *bookService) CreateBook(ctx context.Context, input domain.BookInput) (*domain.Book, error) {
	book, err := s.repo.CreateWithRelations(ctx, input)
	if err != nil {
		return nil, err
	}

	book, err = s.repo.FindByID(ctx, book.ID)
	if err != nil {
		return nil, err
	}

	updated := book

	// Best-effort: don't block book creation on AI availability.
	if summaryUpdated, err := s.generateAndStoreSummary(ctx, book); err == nil {
		updated = summaryUpdated
		book = summaryUpdated
	}

	// Best-effort: if no categories were provided, try generating them.
	if len(input.CategoryIDs) == 0 && s.ai != nil {
		if categoriesUpdated, err := s.generateAndStoreCategories(ctx, book); err == nil {
			updated = categoriesUpdated
			book = categoriesUpdated
		}
	}

	s.scheduleEmbeddingRefresh(book.ID)

	return updated, nil
}

func (s *bookService) GetBook(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	book, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	embeddingNeeded := false
	if strings.TrimSpace(book.Summary) == "" && s.ai != nil {
		if summaryUpdated, err := s.generateAndStoreSummary(ctx, book); err == nil {
			book = summaryUpdated
			embeddingNeeded = true
		}
	}

	if len(book.Categories) == 0 && s.ai != nil {
		if categoriesUpdated, err := s.generateAndStoreCategories(ctx, book); err == nil {
			book = categoriesUpdated
			embeddingNeeded = true
		}
	}

	if embeddingNeeded {
		s.scheduleEmbeddingRefresh(book.ID)
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
		if summaryUpdated, err := s.generateAndStoreSummary(ctx, book); err == nil {
			book = summaryUpdated
		}
	}

	if len(input.CategoryIDs) == 0 && len(book.Categories) == 0 && s.ai != nil {
		// Best-effort: only generate categories if none were provided.
		if categoriesUpdated, err := s.generateAndStoreCategories(ctx, book); err == nil {
			book = categoriesUpdated
		}
	}

	s.scheduleEmbeddingRefresh(book.ID)

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

	updated, err := s.generateAndStoreSummary(ctx, book)
	if err != nil {
		return nil, err
	}

	s.scheduleEmbeddingRefresh(book.ID)
	return updated, nil
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

	updated, err := s.generateAndStoreCategories(ctx, book)
	if err != nil {
		return nil, err
	}

	s.scheduleEmbeddingRefresh(book.ID)
	return updated, nil
}

func (s *bookService) GenerateEmbeddings(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	if s.ai == nil {
		return nil, ai_service.ErrProviderUnavailable
	}

	book, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.generateAndStoreEmbeddings(ctx, book); err != nil {
		return nil, err
	}

	return book, nil
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

	prompt := ai_service.BuildSummaryPrompt(title, author, description)
	summary, err := s.ai.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	book.Summary = strings.TrimSpace(summary)
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

	prompt := ai_service.BuildCategoriesPrompt(title, author, description, summary)
	response, err := s.ai.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	categoryNames, err := parseCategoryNames(response)
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

func (s *bookService) generateAndStoreEmbeddings(ctx context.Context, book *domain.Book) error {
	if s.embeddingSvc == nil {
		return errors.New("embedding service is required")
	}
	if s.ai == nil {
		return ai_service.ErrProviderUnavailable
	}

	return s.embeddingSvc.GenerateBookEmbedding(ctx, *book)
}

func (s *bookService) scheduleEmbeddingRefresh(bookID uuid.UUID) {
	if s.ai == nil {
		return
	}

	if _, loaded := s.embeddingJobs.LoadOrStore(bookID, struct{}{}); loaded {
		return
	}

	go func() {
		defer s.embeddingJobs.Delete(bookID)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		if err := s.refreshEmbeddings(ctx, bookID); err != nil && !errors.Is(err, ai_service.ErrProviderUnavailable) {
			slog.Warn("book embedding refresh failed", "book_id", bookID.String(), "error", err)
		}
	}()
}

func (s *bookService) refreshEmbeddings(ctx context.Context, bookID uuid.UUID) error {
	book, err := s.repo.FindByID(ctx, bookID)
	if err != nil {
		return err
	}

	return s.generateAndStoreEmbeddings(ctx, book)
}

func (s *bookService) RecommendBooks(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Book, error) {
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

func (s *bookService) SemanticSearch(ctx context.Context, query string, limit int) ([]domain.Book, error) {
	if s.ai == nil || s.embeddingRepo == nil {
		return nil, ai_service.ErrProviderUnavailable
	}
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("query is required")
	}
	if limit <= 0 {
		limit = 10
	}

	vectors, err := s.ai.Embed(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	if len(vectors) != 1 || len(vectors[0]) == 0 {
		return nil, errors.New("unexpected embedding response")
	}

	return s.embeddingRepo.SearchNearestBooks(ctx, domain.EmbeddingVector(vectors[0]), limit, nil)
}

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
