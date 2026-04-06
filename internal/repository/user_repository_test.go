package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	pgxmock "github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"

	"booknest/internal/domain"
)

func TestUserRepo_FindByID(t *testing.T) {
	db := setupTestDB(t, &domain.User{}, &domain.UserPreferences{})

	user := domain.User{
		ID:        uuid.New(),
		Email:     "test@booknest.com",
		FirstName: "Test",
		IsActive:  true,
	}

	require.NoError(t, db.Create(&user).Error)

	repo := &userRepo{gorm: db}

	found, err := repo.FindByID(context.Background(), user.ID)

	require.NoError(t, err)
	require.Equal(t, user.ID, found.ID)
	require.Equal(t, user.Email, found.Email)
}

func TestUserRepo_FindByEmail(t *testing.T) {
	db := setupTestDB(t, &domain.User{}, &domain.UserPreferences{})

	user := domain.User{
		ID:        uuid.New(),
		Email:     "test@booknest.com",
		FirstName: "Test",
		IsActive:  true,
	}

	require.NoError(t, db.Create(&user).Error)

	repo := &userRepo{gorm: db}

	found, err := repo.FindByEmail(context.Background(), user.Email)

	require.NoError(t, err)
	require.Equal(t, user.ID, found.ID)
	require.Equal(t, user.Email, found.Email)
}

func TestUserRepo_FindByMobile(t *testing.T) {
	db := setupTestDB(t, &domain.User{}, &domain.UserPreferences{})

	user := domain.User{
		ID:        uuid.New(),
		Email:     "test@booknest.com",
		Mobile:    "+911111100000",
		FirstName: "Test",
		IsActive:  true,
	}

	require.NoError(t, db.Create(&user).Error)

	repo := &userRepo{gorm: db}

	found, err := repo.FindByMobile(context.Background(), user.Mobile)

	require.NoError(t, err)
	require.Equal(t, user.ID, found.ID)
	require.Equal(t, user.Email, found.Email)
}

func TestUserRepo_Create(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &userRepo{
		db: mock,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	user := &domain.User{
		ID:             uuid.New(),
		FirstName:      "Test",
		LastName:       "User",
		Email:          "test@booknest.com",
		Mobile:         "9999999999",
		Password:       "hashed",
		Role:           "user",
		IsActive:       true,
		EmailVerified:  false,
		MobileVerified: false,
	}

	mock.ExpectQuery("INSERT INTO users").
		WithArgs(
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
			pgxmock.AnyArg(),
		).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "created_at", "updated_at"}).
				AddRow(user.ID, time.Now(), time.Now()),
		)

	mock.ExpectExec("INSERT INTO user_preferences").
		WithArgs(user.ID, false).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.Create(context.Background(), user)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_GetPreferencesByUserID(t *testing.T) {
	db := setupTestDB(t, &domain.User{}, &domain.UserPreferences{})

	userID := uuid.New()
	prefs := domain.UserPreferences{
		UserID: userID,
		UseSMS: false,
	}

	require.NoError(t, db.Create(&prefs).Error)

	repo := &userRepo{gorm: db}

	found, err := repo.GetPreferencesByUserID(context.Background(), userID)

	require.NoError(t, err)
	require.Equal(t, userID, found.UserID)
	require.False(t, found.UseSMS)
}

func TestUserRepo_GetPreferencesByUserID_DefaultsWhenMissing(t *testing.T) {
	db := setupTestDB(t, &domain.User{}, &domain.UserPreferences{})

	userID := uuid.New()
	repo := &userRepo{gorm: db}

	found, err := repo.GetPreferencesByUserID(context.Background(), userID)

	require.NoError(t, err)
	require.Equal(t, userID, found.UserID)
	require.False(t, found.UseSMS)
}

func TestUserRepo_Update(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &userRepo{
		db: mock,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	user := &domain.User{
		ID:             uuid.New(),
		FirstName:      "Test",
		LastName:       "User",
		Email:          "test@booknest.com",
		Mobile:         "9999999999",
		Password:       "hashed",
		Role:           "user",
		IsActive:       true,
		EmailVerified:  false,
		MobileVerified: false,
	}

	mock.ExpectQuery("UPDATE users").
		WithArgs(
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
			pgxmock.AnyArg(), pgxmock.AnyArg(),
		).
		WillReturnRows(
			pgxmock.NewRows([]string{"updated_at"}).
				AddRow(time.Now()),
		)

	err = repo.Update(context.Background(), user)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_Delete(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &userRepo{
		db: mock,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	user := &domain.User{
		ID: uuid.New(),
	}

	mock.ExpectQuery("UPDATE users").
		WithArgs(
			pgxmock.AnyArg(),
		).
		WillReturnRows(
			pgxmock.NewRows([]string{"deleted_at"}).
				AddRow(time.Now()),
		)

	err = repo.Delete(context.Background(), user.ID)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
