package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestNew_EmptyURI(t *testing.T) {
	_, err := New(context.Background(), "")
	if err == nil {
		t.Fatal("New() error = nil, want error")
	}
}

func TestNew_InvalidURI(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := New(ctx, "://bad")
	if err == nil {
		t.Fatal("New() error = nil, want error")
	}
}

func TestMigrationFileNames(t *testing.T) {
	names, err := migrationFileNames()
	if err != nil {
		t.Fatalf("migrationFileNames() error = %v", err)
	}
	if len(names) == 0 {
		t.Fatal("migrationFileNames() returned no files")
	}
	found := false
	for _, name := range names {
		if name == "001_create_users.sql" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("migrationFileNames() = %v, want 001_create_users.sql", names)
	}
}

func TestIsUniqueViolation(t *testing.T) {
	if isUniqueViolation(errors.New("other")) {
		t.Fatal("isUniqueViolation(generic) = true, want false")
	}
	if isUniqueViolation(nil) {
		t.Fatal("isUniqueViolation(nil) = true, want false")
	}

	unique := &pgconn.PgError{Code: uniqueViolationCode}
	if !isUniqueViolation(unique) {
		t.Fatal("isUniqueViolation(unique) = false, want true")
	}

	wrapped := errors.Join(errors.New("wrap"), unique)
	if !isUniqueViolation(wrapped) {
		t.Fatal("isUniqueViolation(wrapped unique) = false, want true")
	}
}

func TestClose_NilSafe(t *testing.T) {
	var s *Storage
	s.Close()

	empty := &Storage{}
	empty.Close()
}
