// Package store provides database access for the application.
package store

import (
	"context"
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Store wraps the database connection pool.
type Store struct {
	pool *pgxpool.Pool
}

// New creates a new Store with the given connection string.
func New(ctx context.Context, connString string) (*Store, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Store{pool: pool}, nil
}

// Close closes the database connection pool.
func (s *Store) Close() {
	s.pool.Close()
}

// Pool returns the underlying connection pool.
func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

// Migrate runs all database migrations.
func (s *Store) Migrate(ctx context.Context) error {
	migrations := []string{
		"migrations/001_initial.sql",
		"migrations/002_vectors.sql",
	}

	for _, path := range migrations {
		sql, err := migrationsFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", path, err)
		}

		if _, err := s.pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", path, err)
		}
	}

	return nil
}

// MigrateWithSeeds runs migrations and seeds the database.
func (s *Store) MigrateWithSeeds(ctx context.Context) error {
	if err := s.Migrate(ctx); err != nil {
		return err
	}

	sql, err := migrationsFS.ReadFile("migrations/003_seeds.sql")
	if err != nil {
		return fmt.Errorf("failed to read seeds: %w", err)
	}

	if _, err := s.pool.Exec(ctx, string(sql)); err != nil {
		return fmt.Errorf("failed to run seeds: %w", err)
	}

	return nil
}
