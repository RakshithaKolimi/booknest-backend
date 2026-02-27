package repository

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/gorm"

	"booknest/internal/domain"
)

type publisherRepo struct {
	db   domain.DBExecer
	gorm *gorm.DB
	sb   squirrel.StatementBuilderType
}

func NewPublisherRepo(db *pgxpool.Pool, gormDB *gorm.DB) domain.PublisherRepository {
	return &publisherRepo{
		db:   db,
		gorm: gormDB,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *publisherRepo) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.Publisher, error) {

	var publisher domain.Publisher

	err := r.gorm.
		WithContext(ctx).
		Where("id = ?", id).
		First(&publisher).
		Error

	return publisher, err
}

func (r *publisherRepo) List(
	ctx context.Context,
	limit, offset int,
	search string,
) ([]domain.Publisher, error) {
	var publishers []domain.Publisher

	query := r.gorm.WithContext(ctx).
		Where("deleted_at IS NULL")
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("(trading_name ILIKE ? OR legal_name ILIKE ?)", like, like)
	}

	err := query.
		Order("LOWER(trading_name) ASC").
		Limit(limit).
		Offset(offset).
		Find(&publishers).Error
	return publishers, err
}

func (r *publisherRepo) Create(
	ctx context.Context,
	publisher *domain.Publisher,
) error {

	query, args, err := r.sb.
		Insert("publishers").
		Columns(
			"legal_name",
			"trading_name",
			"email",
			"mobile",
			"address",
			"city",
			"state",
			"country",
			"zipcode",
		).
		Values(
			publisher.LegalName,
			publisher.TradingName,
			publisher.Email,
			publisher.Mobile,
			publisher.Address,
			publisher.City,
			publisher.State,
			publisher.Country,
			publisher.Zipcode,
		).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()

	if err != nil {
		return err
	}

	row := queryRowWithTx(ctx, r.db, query, args...)

	return row.Scan(
		&publisher.ID,
		&publisher.CreatedAt,
		&publisher.UpdatedAt,
	)
}

func (r *publisherRepo) Update(
	ctx context.Context,
	publisher *domain.Publisher,
) error {

	query, args, err := r.sb.
		Update("publishers").
		Set("legal_name", publisher.LegalName).
		Set("trading_name", publisher.TradingName).
		Set("email", publisher.Email).
		Set("mobile", publisher.Mobile).
		Set("address", publisher.Address).
		Set("city", publisher.City).
		Set("state", publisher.State).
		Set("country", publisher.Country).
		Set("zipcode", publisher.Zipcode).
		Set("is_active", publisher.IsActive).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": publisher.ID}).
		Suffix("RETURNING updated_at").
		ToSql()

	if err != nil {
		return err
	}

	row := queryRowWithTx(ctx, r.db, query, args...)
	return row.Scan(&publisher.UpdatedAt)
}

func (r *publisherRepo) SetActive(
	ctx context.Context,
	id uuid.UUID,
	active bool,
) error {

	query, args, err := r.sb.
		Update("publishers").
		Set("is_active", active).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": id}).
		ToSql()

	if err != nil {
		return err
	}

	return execWithTx(ctx, r.db, query, args...)
}

func (r *publisherRepo) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {

	query := `
		UPDATE publishers
		SET deleted_at = NOW(),
    		is_active = FALSE
		WHERE id = $1
		RETURNING deleted_at;
	`

	row := queryRowWithTx(ctx, r.db, query, id)

	var deletedAt time.Time
	return row.Scan(&deletedAt)
}
