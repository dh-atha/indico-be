package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func Connect(ctx context.Context) (*sqlx.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5432/appdb?sslmode=disable"
	}

	database, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	database.SetMaxOpenConns(50)
	database.SetMaxIdleConns(10)
	database.SetConnMaxLifetime(0)

	if err := database.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("db ping failed: %w", err)
	}

	return database, nil
}
