package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"booknest/internal/domain"
)

type userRepo struct {
	db   domain.DBExecer
	gorm *gorm.DB
	sb   squirrel.StatementBuilderType
}

func NewUserRepo(db *pgxpool.Pool, gormDB *gorm.DB) domain.UserRepository {
	return &userRepo{
		db:   db,
		gorm: gormDB,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *userRepo) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.User, error) {

	var user domain.User

	err := r.gorm.
		WithContext(ctx).
		Where("id = ?", id).
		First(&user).
		Error

	return user, err
}

func (r *userRepo) FindByEmail(
	ctx context.Context,
	email string,
) (domain.User, error) {

	var user domain.User

	err := r.gorm.
		WithContext(ctx).
		Where("email = ?", email).
		First(&user).
		Error

	return user, err
}

func (r *userRepo) FindByMobile(
	ctx context.Context,
	mobile string,
) (domain.User, error) {

	var user domain.User

	err := r.gorm.
		WithContext(ctx).
		Where("mobile = ?", mobile).
		First(&user).
		Error

	return user, err
}

func (r *userRepo) GetPreferencesByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (domain.UserPreferences, error) {
	var prefs domain.UserPreferences

	if r.db != nil {
		query := `
			SELECT user_id, use_sms, created_at, updated_at, deleted_at
			FROM user_preferences
			WHERE user_id = $1 AND deleted_at IS NULL;
		`

		row := queryRowWithTx(ctx, r.db, query, userID)
		err := row.Scan(
			&prefs.UserID,
			&prefs.UseSMS,
			&prefs.CreatedAt,
			&prefs.UpdatedAt,
			&prefs.DeletedAt,
		)
		if err == sql.ErrNoRows {
			return domain.UserPreferences{
				UserID: userID,
				UseSMS: false,
			}, nil
		}
		return prefs, err
	}

	err := r.gorm.
		WithContext(ctx).
		Where("user_id = ?", userID).
		First(&prefs).
		Error
	if err == gorm.ErrRecordNotFound {
		return domain.UserPreferences{
			UserID: userID,
			UseSMS: false,
		}, nil
	}

	return prefs, err
}

func (r *userRepo) UpdatePreferences(
	ctx context.Context,
	prefs *domain.UserPreferences,
) error {
	if r.db != nil {
		query := `
			INSERT INTO user_preferences (user_id, use_sms, created_at, updated_at)
			VALUES ($1, $2, NOW(), NOW())
			ON CONFLICT (user_id)
			DO UPDATE SET use_sms = EXCLUDED.use_sms, updated_at = NOW();
		`

		if err := execWithTx(ctx, r.db, query, prefs.UserID, prefs.UseSMS); err != nil {
			return err
		}
		return nil
	}

	return r.gorm.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"use_sms", "updated_at"}),
	}).Create(prefs).Error
}

func (r *userRepo) Create(ctx context.Context, user *domain.User) error {
	query, args, err := r.sb.
		Insert("users").
		Columns(
			"id",
			"first_name",
			"last_name",
			"email",
			"mobile",
			"password",
			"role",
			"is_active",
			"email_verified",
			"mobile_verified",
		).
		Values(
			user.ID,
			user.FirstName,
			user.LastName,
			user.Email,
			user.Mobile,
			user.Password,
			user.Role,
			user.IsActive,
			user.EmailVerified,
			user.MobileVerified,
		).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return err
	}

	row := queryRowWithTx(ctx, r.db, query, args...)
	if err := row.Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return err
	}

	if err := execWithTx(
		ctx,
		r.db,
		`INSERT INTO user_preferences (user_id, use_sms, created_at, updated_at) VALUES ($1, $2, NOW(), NOW())`,
		user.ID,
		false,
	); err != nil {
		return err
	}

	return nil
}

func (r *userRepo) Update(ctx context.Context, user *domain.User) error {
	query, args, err := r.sb.
		Update("users").
		Set("first_name", user.FirstName).
		Set("last_name", user.LastName).
		Set("password", user.Password).
		Set("last_login", user.LastLogin).
		Set("is_active", user.IsActive).
		Set("email_verified", user.EmailVerified).
		Set("mobile_verified", user.MobileVerified).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": user.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return err
	}

	row := queryRowWithTx(ctx, r.db, query, args...)
	return row.Scan(&user.UpdatedAt)
}

func (r *userRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users
		SET deleted_at = NOW()
		WHERE id = $1
		RETURNING deleted_at;
	`

	row := queryRowWithTx(ctx, r.db, query, id)

	var deletedAt time.Time
	return row.Scan(&deletedAt)
}
