package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	// allows test to inject mock behavior
	newPgxPool = pgxpool.NewWithConfig
	pingPgxPool = func(ctx context.Context, pool *pgxpool.Pool) error {
		return pool.Ping(ctx)
	}
)

func Connect() (*pgxpool.Pool, error) {
	// You can load these from environment variables or .env file
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	// Build DSN (Data source name)
	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable&timezone=Asia/Kolkata",
		user, password, host, port, dbName,
	)

	// Create config
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Optional tuning
	config.MaxConns = 20
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.HealthCheckPeriod = time.Minute * 2

	// Connect
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := newPgxPool(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgxpool: %w", err)
	}

	// Ping to verify connection
	if err = pingPgxPool(ctx, pool); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ Connected to PostgreSQL using pgxpool!")
	return pool, nil
}

func ConnectGORM() (*gorm.DB, error) {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Kolkata",
		host, user, password, dbName, port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // change to Silent in prod
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// Get underlying sql.DB to configure pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Pool configuration (similar to pgxpool)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Verify connection
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	log.Println("✅ Connected to PostgreSQL using GORM!")
	return db, nil
}
