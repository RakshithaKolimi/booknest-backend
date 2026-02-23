package repository

import (
	"context"
	"database/sql"
	"fmt"

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

	// ---------- DATA QUERY ----------
	dataQuery := buildBookBaseQuery()
	dataQuery = applyBookFilters(dataQuery, filter)
	dataQuery = applyBookSorting(dataQuery, q.Sort)

	dataQuery = dataQuery.
		OrderBy("b.created_at DESC").
		Limit(q.Limit).
		Offset(q.Offset).
		PlaceholderFormat(sq.Dollar)

	sqlQuery, args, err := dataQuery.ToSql()
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.sql.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, 0, err
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
			return nil, 0, err
		}
		books = append(books, book)
	}

	// ---------- COUNT QUERY ----------
	countQuery := sq.
		Select("COUNT(DISTINCT b.id)").
		From("books b").
		Where("b.deleted_at IS NULL")

	countQuery = applyBookFilters(countQuery, filter).
		PlaceholderFormat(sq.Dollar)

	countSQL, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, err
	}

	var total int64
	if err := r.sql.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	return books, total, nil
}

func (r *bookRepository) List(ctx context.Context, limit, offset int) ([]domain.Book, error) {
	var books []domain.Book
	err := r.db.WithContext(ctx).
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
		"b.author_name",
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
		From("books b").
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
				sq.ILike{"b.author_name": search},
				sq.ILike{"b.isbn": search},
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
			Join("book_categories bc ON bc.book_id = b.id").
			Where("bc.deleted_at IS NULL").
			Where(sq.Eq{"bc.category_id": filter.CategoryIDs})
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
