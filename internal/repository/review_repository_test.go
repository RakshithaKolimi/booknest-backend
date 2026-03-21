package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"booknest/internal/domain"
)

func TestReviewRepository_UpsertAndListByBookID(t *testing.T) {
	db := setupTestDB(t,
		&domain.Author{},
		&domain.Publisher{},
		&domain.Book{},
		&domain.User{},
		&domain.Review{},
	)

	repo := NewReviewRepository(db)
	ctx := context.Background()

	authorID := uuid.New()
	publisherID := uuid.New()
	bookID := uuid.New()
	userOneID := uuid.New()
	userTwoID := uuid.New()

	require.NoError(t, db.Create(&domain.Author{
		ID:   authorID,
		Name: "Author One",
	}).Error)
	require.NoError(t, db.Create(&domain.Publisher{
		ID:          publisherID,
		LegalName:   "Legal Name",
		TradingName: "Trading Name",
		Email:       "publisher@example.com",
		Mobile:      "+911111111119",
		Address:     "Address",
		City:        "Hyderabad",
		State:       "Telangana",
		Country:     "India",
		Zipcode:     "500001",
	}).Error)
	require.NoError(t, db.Create(&domain.Book{
		ID:             bookID,
		Name:           "Reviewable Book",
		AuthorID:       authorID,
		PublisherID:    publisherID,
		AvailableStock: 10,
		Price:          299,
	}).Error)
	require.NoError(t, db.Create(&domain.User{
		ID:             userOneID,
		FirstName:      "Alice",
		LastName:       "Reader",
		Email:          "alice@example.com",
		Mobile:         "+911111111117",
		Password:       "secret123",
		Role:           domain.UserRoleUser,
		IsActive:       true,
		EmailVerified:  true,
		MobileVerified: true,
	}).Error)
	require.NoError(t, db.Create(&domain.User{
		ID:             userTwoID,
		FirstName:      "Bob",
		LastName:       "Reader",
		Email:          "bob@example.com",
		Mobile:         "+911111111118",
		Password:       "secret123",
		Role:           domain.UserRoleUser,
		IsActive:       true,
		EmailVerified:  true,
		MobileVerified: true,
	}).Error)

	created, err := repo.Upsert(ctx, bookID, userOneID, domain.ReviewInput{
		Rating:  5,
		Comment: "Excellent",
	})
	require.NoError(t, err)
	require.Equal(t, 5, created.Rating)
	require.Equal(t, "Alice", created.User.FirstName)

	second, err := repo.Upsert(ctx, bookID, userTwoID, domain.ReviewInput{
		Rating:  3,
		Comment: "Decent",
	})
	require.NoError(t, err)
	require.Equal(t, 3, second.Rating)

	updated, err := repo.Upsert(ctx, bookID, userOneID, domain.ReviewInput{
		Rating:  4,
		Comment: "Better on a second read",
	})
	require.NoError(t, err)
	require.Equal(t, created.ID, updated.ID)
	require.Equal(t, 4, updated.Rating)
	require.Equal(t, "Better on a second read", updated.Comment)

	reviews, summary, err := repo.ListByBookID(ctx, bookID)
	require.NoError(t, err)
	require.Len(t, reviews, 2)
	require.Equal(t, int64(2), summary.TotalReviews)
	require.InDelta(t, 3.5, summary.AverageRating, 0.001)
	require.NotEmpty(t, reviews[0].User.Email)
}

func TestReviewRepository_UpsertMissingBook(t *testing.T) {
	db := setupTestDB(t, &domain.Book{}, &domain.User{}, &domain.Review{})
	repo := NewReviewRepository(db)

	_, err := repo.Upsert(context.Background(), uuid.New(), uuid.New(), domain.ReviewInput{
		Rating:  5,
		Comment: "Missing book",
	})
	require.Error(t, err)
}
