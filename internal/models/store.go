package models

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store provides database access
type Store struct {
	DB *pgxpool.Pool
}

// NewStore creates a new database store
func NewStore(databaseURL string) (*Store, error) {
	if databaseURL == "" {
		databaseURL = os.Getenv("CONTINUUM_DATABASE_URL")
	}
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL not provided")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &Store{DB: pool}, nil
}

// Close closes the database connection
func (s *Store) Close() {
	s.DB.Close()
}

// RunMigrations executes all migration files
func (s *Store) RunMigrations(ctx context.Context) error {
	migrations := []string{
		"001_initial_schema.sql",
		"002_row_level_security.sql",
	}

	for _, migration := range migrations {
		sql, err := os.ReadFile(fmt.Sprintf("internal/models/migrations/%s", migration))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", migration, err)
		}

		if _, err := s.DB.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("execute migration %s: %w", migration, err)
		}
	}

	return nil
}
