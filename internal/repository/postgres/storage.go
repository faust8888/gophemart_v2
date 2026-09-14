// Package postgres реализует хранение данных накопительной системы лояльности в PostgreSQL.
package postgres

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Storage реализует доступ к PostgreSQL.
type Storage struct {
	pool *pgxpool.Pool
}

// New создаёт пул соединений с PostgreSQL и применяет миграции схемы.
func New(ctx context.Context, databaseURI string) (*Storage, error) {
	if databaseURI == "" {
		return nil, fmt.Errorf("database URI is empty")
	}

	pool, err := pgxpool.New(ctx, databaseURI)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	s := &Storage{pool: pool}
	if err := s.migrate(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return s, nil
}

// Close закрывает пул соединений с базой данных.
func (s *Storage) Close() {
	if s == nil || s.pool == nil {
		return
	}
	s.pool.Close()
}

func (s *Storage) migrate(ctx context.Context) error {
	names, err := migrationFileNames()
	if err != nil {
		return err
	}

	for _, name := range names {
		sql, err := migrationFiles.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if _, err := s.pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
	}
	return nil
}

func migrationFileNames() ([]string, error) {
	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names, nil
}
