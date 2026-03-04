package repository

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"booknest/internal/domain"
)

var allowedBookSortColumns = map[string]string{
	domain.SortByCreatedAt: "b.created_at",
	domain.SortByPrice:     "b.price",
	domain.SortByName:      "b.name",
	domain.SortByStock:     "b.available_stock",
}

type bookRepository struct {
	db  *gorm.DB
	sql *sql.DB
}

func NewBookRepository(db *gorm.DB, sql *sql.DB) domain.BookRepository {
	return &bookRepository{db: db, sql: sql}
}

// Use GORM to create a book
func (r *bookRepository) Create(ctx context.Context, book *domain.Book) error {
	return r.db.WithContext(ctx).Create(book).Error
}

func (r *bookRepository) CreateWithRelations(
	ctx context.Context,
	input domain.BookInput,
) (*domain.Book, error) {
	book := &domain.Book{
		ID:   uuid.New(),
		Name: input.Name,
	}

	categoryIDs := uniqueCategoryIDs(input.CategoryIDs)

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		authorID, err := resolveAuthorID(tx, input.AuthorID, input.AuthorName)
		if err != nil {
			return err
		}

		book.AuthorID = authorID
		book.AvailableStock = input.AvailableStock
		book.ImageURL = input.ImageURL
		book.IsActive = input.IsActive
		book.Description = input.Description
		book.ISBN = input.ISBN
		book.Price = input.Price
		book.DiscountPercentage = input.DiscountPercentage
		book.PublisherID = input.PublisherID

		if err := tx.Create(book).Error; err != nil {
			return err
		}

		if len(categoryIDs) == 0 {
			return nil
		}

		if err := validateCategories(tx, categoryIDs); err != nil {
			return err
		}

		for _, cid := range categoryIDs {
			bc := domain.BookCategory{
				BookID:     book.ID,
				CategoryID: cid,
			}
			if err := tx.Create(&bc).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return book, nil
}

func (r *bookRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	var book domain.Book
	err := r.db.WithContext(ctx).
		Preload("Author").
		Preload("Publisher").
		Preload("Categories").
		First(&book, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &book, nil
}

func (r *bookRepository) FilterByCriteria(ctx context.Context, filter domain.BookFilter, q domain.QueryOptions) ([]domain.Book, int64, error) {
	books, total, _, _, err := r.QueryBooks(ctx, filter, q)
	return books, total, err
}

func (r *bookRepository) QueryBooks(
	ctx context.Context,
	filter domain.BookFilter,
	q domain.QueryOptions,
) ([]domain.Book, int64, *string, bool, error) {
	limit := q.Limit
	if limit == 0 {
		limit = 10
	}

	// Build a book query
	dataQuery := buildBookBaseQuery()
	// Apply filters
	dataQuery = applyBookFilters(dataQuery, filter)

	// Check if cursor exists
	if q.Cursor != nil && *q.Cursor != "" {
		// Get the cursor created_at and cursor id 
		cursorCreatedAt, cursorID, err := decodeBookCursor(*q.Cursor)
		if err != nil {
			return nil, 0, nil, false, err
		}

		// get books next to cursor
		dataQuery = dataQuery.Where(
			"(b.created_at < ? OR (b.created_at = ? AND b.id < ?))",
			cursorCreatedAt,
			cursorCreatedAt,
			cursorID,
		)
		dataQuery = dataQuery.OrderBy("b.created_at DESC", "b.id DESC")
	} else {
		// If cursor does not exist, get by limit and offset
		dataQuery = applyBookSorting(dataQuery, q.Sort)
		dataQuery = dataQuery.OrderBy("b.created_at DESC", "b.id DESC")
		dataQuery = dataQuery.Offset(q.Offset)
	}

	dataQuery = dataQuery.
		Limit(limit + 1).
		PlaceholderFormat(sq.Dollar)

	// Get the query and arguments
	sqlQuery, args, err := dataQuery.ToSql()
	if err != nil {
		return nil, 0, nil, false, err
	}

	rows, err := r.sql.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, 0, nil, false, err
	}
	defer rows.Close()

	books := make([]domain.Book, 0)
	for rows.Next() {
		var book domain.Book
		err := rows.Scan(
			&book.ID,
			&book.Name,
			&book.AuthorID,
			&book.AvailableStock,
			&book.ImageURL,
			&book.IsActive,
			&book.Description,
			&book.ISBN,
			&book.Price,
			&book.DiscountPercentage,
			&book.PublisherID,
			&book.CreatedAt,
			&book.UpdatedAt,
		)
		if err != nil {
			return nil, 0, nil, false, err
		}
		books = append(books, book)
	}

	hasMore := len(books) > int(limit)
	if hasMore {
		books = books[:limit]
	}

	countQuery := sq.
		Select("COUNT(DISTINCT b.id)").
		From("books b").
		Join("authors a ON a.id = b.author_id").
		Join("publishers p ON p.id = b.publisher_id").
		Where("b.deleted_at IS NULL")

	countQuery = applyBookFilters(countQuery, filter).
		PlaceholderFormat(sq.Dollar)

	countSQL, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, nil, false, err
	}

	var total int64
	if err := r.sql.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, nil, false, err
	}

	if len(books) > 0 {
		if err := r.hydrateBooksWithRelations(ctx, books); err != nil {
			return nil, 0, nil, false, err
		}
	}

	var nextCursor *string
	// if we have more books, move to next cursor
	if hasMore && len(books) > 0 {
		cursor := encodeBookCursor(books[len(books)-1].CreatedAt, books[len(books)-1].ID)
		nextCursor = &cursor
	}

	return books, total, nextCursor, hasMore, nil
}

func (r *bookRepository) List(ctx context.Context, limit, offset int) ([]domain.Book, error) {
	var books []domain.Book
	err := r.db.WithContext(ctx).
		Preload("Author").
		Preload("Categories").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&books).Error
	return books, err
}

func (r *bookRepository) Update(ctx context.Context, book *domain.Book) error {
	return r.db.WithContext(ctx).Save(book).Error
}

func (r *bookRepository) UpdateWithRelations(
	ctx context.Context,
	id uuid.UUID,
	input domain.BookInput,
) (*domain.Book, error) {
	book, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	categoryIDs := uniqueCategoryIDs(input.CategoryIDs)

	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		authorID, err := resolveAuthorID(tx, input.AuthorID, input.AuthorName)
		if err != nil {
			return err
		}

		book.Name = input.Name
		book.AuthorID = authorID
		book.AvailableStock = input.AvailableStock
		book.ImageURL = input.ImageURL
		book.IsActive = input.IsActive
		book.Description = input.Description
		book.ISBN = input.ISBN
		book.Price = input.Price
		book.DiscountPercentage = input.DiscountPercentage
		book.PublisherID = input.PublisherID

		if err := tx.Save(book).Error; err != nil {
			return err
		}

		if err := tx.Where("book_id = ?", book.ID).Delete(&domain.BookCategory{}).Error; err != nil {
			return err
		}

		if len(categoryIDs) == 0 {
			return nil
		}

		if err := validateCategories(tx, categoryIDs); err != nil {
			return err
		}

		for _, cid := range categoryIDs {
			bc := domain.BookCategory{
				BookID:     book.ID,
				CategoryID: cid,
			}
			if err := tx.Create(&bc).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return book, nil
}

func (r *bookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Book{}, "id = ?", id).Error
}

func buildBookBaseQuery() sq.SelectBuilder {
	return sq.Select(
		"b.id",
		"b.name",
		"b.author_id",
		"b.available_stock",
		"b.image_url",
		"b.is_active",
		"b.description",
		"b.isbn",
		"b.price",
		"b.discount_percentage",
		"b.publisher_id",
		"b.created_at",
		"b.updated_at",
	).
		Options("DISTINCT").
		From("books b").
		Join("authors a ON a.id = b.author_id").
		Join("publishers p ON p.id = b.publisher_id").
		Where("b.deleted_at IS NULL")
}

func applyBookFilters(
	q sq.SelectBuilder,
	filter domain.BookFilter,
) sq.SelectBuilder {

	if filter.Search != nil {
		search := "%" + *filter.Search + "%"
		q = q.Where(
			sq.Or{
				sq.ILike{"b.name": search},
				sq.ILike{"a.name": search},
				sq.ILike{"b.isbn": search},
				sq.ILike{"p.trading_name": search},
				sq.ILike{"p.legal_name": search},
				sq.Expr(
					`EXISTS (
						SELECT 1
						FROM book_categories bc_search
						JOIN categories c_search ON c_search.id = bc_search.category_id
						WHERE bc_search.book_id = b.id
						  AND c_search.deleted_at IS NULL
						  AND c_search.name ILIKE ?
					)`,
					search,
				),
			},
		)
	}

	if filter.MinPrice != nil {
		q = q.Where(sq.GtOrEq{"b.price": *filter.MinPrice})
	}

	if filter.MaxPrice != nil {
		q = q.Where(sq.LtOrEq{"b.price": *filter.MaxPrice})
	}

	if filter.IsActive != nil {
		q = q.Where(sq.Eq{"b.is_active": *filter.IsActive})
	}

	if filter.MinStock != nil {
		q = q.Where(sq.GtOrEq{"b.available_stock": *filter.MinStock})
	}

	if len(filter.IDs) > 0 {
		q = q.Where(sq.Eq{"b.id": filter.IDs})
	}

	if len(filter.AuthorIDs) > 0 {
		q = q.Where(sq.Eq{"b.author_id": filter.AuthorIDs})
	}

	if len(filter.PublisherIDs) > 0 {
		q = q.Where(sq.Eq{"b.publisher_id": filter.PublisherIDs})
	}

	if len(filter.CategoryIDs) > 0 {
		q = q.
			LeftJoin("book_categories bc_ids ON bc_ids.book_id = b.id").
			Where(sq.Eq{"bc_ids.category_id": filter.CategoryIDs})
	}

	return q
}

func applyBookSorting(
	q sq.SelectBuilder,
	sort *domain.SortOptions,
) sq.SelectBuilder {

	// Default sort
	if sort == nil {
		return q.OrderBy("b.created_at DESC")
	}

	column, ok := allowedBookSortColumns[sort.Field]
	if !ok {
		return q.OrderBy("b.created_at DESC")
	}

	order := "ASC"
	if sort.Order == domain.Desc {
		order = "DESC"
	}

	return q.OrderBy(column + " " + order)
}

func uniqueCategoryIDs(categoryIDs []uuid.UUID) []uuid.UUID {
	if len(categoryIDs) == 0 {
		return nil
	}

	seen := make(map[uuid.UUID]struct{}, len(categoryIDs))
	unique := make([]uuid.UUID, 0, len(categoryIDs))
	for _, id := range categoryIDs {
		if id == uuid.Nil {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}

	return unique
}

func validateCategories(tx *gorm.DB, categoryIDs []uuid.UUID) error {
	var count int64
	if err := tx.Model(&domain.Category{}).Where("id IN ?", categoryIDs).Count(&count).Error; err != nil {
		return err
	}

	if int(count) != len(categoryIDs) {
		return fmt.Errorf("one or more categories are invalid")
	}

	return nil
}

func resolveAuthorID(tx *gorm.DB, authorID *uuid.UUID, authorName string) (uuid.UUID, error) {
	if authorID != nil && *authorID != uuid.Nil {
		var author domain.Author
		if err := tx.First(&author, "id = ?", *authorID).Error; err != nil {
			return uuid.Nil, err
		}
		return *authorID, nil
	}

	var author domain.Author
	err := tx.Where("LOWER(name) = LOWER(?)", authorName).First(&author).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return uuid.Nil, err
		}

		author = domain.Author{
			ID:   uuid.New(),
			Name: authorName,
		}

		if err := tx.Create(&author).Error; err != nil {
			return uuid.Nil, err
		}
	}

	return author.ID, nil
}

func (r *bookRepository) hydrateBooksWithRelations(ctx context.Context, books []domain.Book) error {
	if len(books) == 0 {
		return nil
	}

	ids := make([]uuid.UUID, 0, len(books))
	for _, book := range books {
		ids = append(ids, book.ID)
	}

	var hydrated []domain.Book
	if err := r.db.WithContext(ctx).
		Preload("Author").
		Preload("Publisher").
		Preload("Categories").
		Where("id IN ?", ids).
		Find(&hydrated).Error; err != nil {
		return err
	}

	indexByID := make(map[uuid.UUID]domain.Book, len(hydrated))
	for _, book := range hydrated {
		indexByID[book.ID] = book
	}

	for i := range books {
		if hydratedBook, ok := indexByID[books[i].ID]; ok {
			books[i].Author = hydratedBook.Author
			books[i].Publisher = hydratedBook.Publisher
			books[i].Categories = hydratedBook.Categories
		}
	}

	return nil
}

func encodeBookCursor(createdAt time.Time, id uuid.UUID) string {
	raw := createdAt.UTC().Format(time.RFC3339Nano) + "|" + id.String()
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeBookCursor(cursor string) (time.Time, uuid.UUID, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, uuid.Nil, fmt.Errorf("invalid cursor")
	}

	parts := strings.Split(string(decoded), "|")
	if len(parts) != 2 {
		return time.Time{}, uuid.Nil, fmt.Errorf("invalid cursor")
	}

	createdAt, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, uuid.Nil, fmt.Errorf("invalid cursor")
	}

	id, err := uuid.Parse(parts[1])
	if err != nil {
		return time.Time{}, uuid.Nil, fmt.Errorf("invalid cursor")
	}

	return createdAt, id, nil
}
