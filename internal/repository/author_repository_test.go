package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"booknest/internal/domain"
)

func TestAuthorRepo_CRUDAndList(t *testing.T) {
	db := setupTestDB(t, &domain.Author{})
	repo := &authorRepo{gorm: db}

	ctx := context.Background()
	author := &domain.Author{ID: uuid.New(), Name: "Author A"}
	require.NoError(t, repo.Create(ctx, author))

	foundByID, err := repo.FindByID(ctx, author.ID)
	require.NoError(t, err)
	require.Equal(t, author.Name, foundByID.Name)

	foundByName, err := repo.FindByName(ctx, "author a")
	require.NoError(t, err)
	require.Equal(t, author.ID, foundByName.ID)

	second := &domain.Author{ID: uuid.New(), Name: "Author B"}
	require.NoError(t, repo.Create(ctx, second))

	list, err := repo.List(ctx, 1, 0, "")
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "Author A", list[0].Name)

	author.Name = "Author A Updated"
	require.NoError(t, repo.Update(ctx, author))
	updated, err := repo.FindByID(ctx, author.ID)
	require.NoError(t, err)
	require.Equal(t, "Author A Updated", updated.Name)

	require.NoError(t, repo.Delete(ctx, author.ID))
	_, err = repo.FindByID(ctx, author.ID)
	require.Error(t, err)
}
