package repository

import (
	"context"
	"strings"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"booknest/internal/domain"
)

func TestBookRepo_FindByIDAndList(t *testing.T) {
	db := setupTestDB(t,
		&domain.Author{},
		&domain.Publisher{},
		&domain.Category{},
		&domain.Book{},
		&domain.BookCategory{},
	)
	sqlDB, err := db.DB()
	require.NoError(t, err)

	repo := &bookRepository{db: db, sql: sqlDB}
	ctx := context.Background()

	authorID := uuid.New()
	publisherID := uuid.New()
	categoryID := uuid.New()
	bookID := uuid.New()

	require.NoError(t, db.Create(&domain.Author{ID: authorID, Name: "Author"}).Error)
	require.NoError(t, db.Create(&domain.Publisher{
		ID:          publisherID,
		LegalName:   "Legal",
		TradingName: "Trading",
		Email:       "a@b.com",
		Mobile:      "+911111111111",
		Address:     "Addr",
		City:        "City",
		State:       "State",
		Country:     "Country",
		Zipcode:     "123456",
	}).Error)
	require.NoError(t, db.Create(&domain.Category{ID: categoryID, Name: "Fiction"}).Error)

	book := &domain.Book{
		ID:             bookID,
		Name:           "Book Name",
		AuthorID:       authorID,
		PublisherID:    publisherID,
		AvailableStock: 3,
		Price:          100,
	}
	require.NoError(t, repo.Create(ctx, book))
	require.NoError(t, db.Create(&domain.BookCategory{BookID: bookID, CategoryID: categoryID}).Error)

	found, err := repo.FindByID(ctx, bookID)
	require.NoError(t, err)
	require.Equal(t, bookID, found.ID)
	require.Equal(t, authorID, found.Author.ID)
	require.Equal(t, publisherID, found.Publisher.ID)
	require.Len(t, found.Categories, 1)

	list, err := repo.List(ctx, 10, 0)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Len(t, list[0].Categories, 1)

	found.Name = "Updated"
	require.NoError(t, repo.Update(ctx, found))

	updated, err := repo.FindByID(ctx, bookID)
	require.NoError(t, err)
	require.Equal(t, "Updated", updated.Name)

	require.NoError(t, repo.Delete(ctx, bookID))
	_, err = repo.FindByID(ctx, bookID)
	require.Error(t, err)
}

func TestBookQueryHelpers_FilterAndSort(t *testing.T) {
	search := "harry"
	minPrice := 50.0
	maxPrice := 300.0
	isActive := true
	minStock := 2
	authorID := uuid.New()
	publisherID := uuid.New()
	categoryID := uuid.New()
	bookID := uuid.New()

	q := sq.Select("b.id").From("books b")
	q = applyBookFilters(q, domain.BookFilter{
		Search:       &search,
		MinPrice:     &minPrice,
		MaxPrice:     &maxPrice,
		IsActive:     &isActive,
		MinStock:     &minStock,
		IDs:          []uuid.UUID{bookID},
		AuthorIDs:    []uuid.UUID{authorID},
		PublisherIDs: []uuid.UUID{publisherID},
		CategoryIDs:  []uuid.UUID{categoryID},
	})
	q = applyBookSorting(q, &domain.SortOptions{Field: domain.SortByPrice, Order: domain.Desc})

	sqlText, args, err := q.PlaceholderFormat(sq.Dollar).ToSql()
	require.NoError(t, err)
	require.True(t, strings.Contains(sqlText, "book_categories bc ON bc.book_id = b.id"))
	require.True(t, strings.Contains(sqlText, "b.price DESC"))
	require.NotEmpty(t, args)

	defaultSorted, _, err := applyBookSorting(
		sq.Select("b.id").From("books b"),
		nil,
	).PlaceholderFormat(sq.Dollar).ToSql()
	require.NoError(t, err)
	require.True(t, strings.Contains(defaultSorted, "b.created_at DESC"))
}

func TestBookRepo_CreateAndUpdateWithRelations(t *testing.T) {
	db := setupTestDB(t,
		&domain.Author{},
		&domain.Publisher{},
		&domain.Category{},
		&domain.Book{},
		&domain.BookCategory{},
	)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	repo := &bookRepository{db: db, sql: sqlDB}
	ctx := context.Background()

	publisherID := uuid.New()
	categoryA := uuid.New()
	categoryB := uuid.New()
	existingAuthorID := uuid.New()
	require.NoError(t, db.Create(&domain.Publisher{
		ID:          publisherID,
		LegalName:   "Legal",
		TradingName: "Trading",
		Email:       "a@b.com",
		Mobile:      "+911111111112",
		Address:     "Addr",
		City:        "City",
		State:       "State",
		Country:     "Country",
		Zipcode:     "123456",
	}).Error)
	require.NoError(t, db.Create(&domain.Category{ID: categoryA, Name: "One"}).Error)
	require.NoError(t, db.Create(&domain.Category{ID: categoryB, Name: "Two"}).Error)
	require.NoError(t, db.Create(&domain.Author{ID: existingAuthorID, Name: "Existing"}).Error)

	created, err := repo.CreateWithRelations(ctx, domain.BookInput{
		Name:               "Created Book",
		AuthorID:           &existingAuthorID,
		AuthorName:         "Ignored",
		AvailableStock:     5,
		IsActive:           true,
		Description:        "d",
		Price:              100,
		DiscountPercentage: 10,
		PublisherID:        publisherID,
		CategoryIDs:        []uuid.UUID{categoryA, categoryA, uuid.Nil, categoryB},
	})
	require.NoError(t, err)
	require.Equal(t, existingAuthorID, created.AuthorID)

	var links []domain.BookCategory
	require.NoError(t, db.Where("book_id = ?", created.ID).Find(&links).Error)
	require.Len(t, links, 2)

	updated, err := repo.UpdateWithRelations(ctx, created.ID, domain.BookInput{
		Name:               "Updated Book",
		AuthorName:         "New Author",
		AvailableStock:     8,
		IsActive:           true,
		Description:        "new",
		Price:              120,
		DiscountPercentage: 0,
		PublisherID:        publisherID,
		CategoryIDs:        []uuid.UUID{categoryA},
	})
	require.NoError(t, err)
	require.Equal(t, "Updated Book", updated.Name)

	var newAuthor domain.Author
	require.NoError(t, db.Where("LOWER(name) = LOWER(?)", "New Author").First(&newAuthor).Error)
	require.NotEqual(t, uuid.Nil, newAuthor.ID)

	links = nil
	require.NoError(t, db.Where("book_id = ?", created.ID).Find(&links).Error)
	require.Len(t, links, 1)
	require.Equal(t, categoryA, links[0].CategoryID)
}
