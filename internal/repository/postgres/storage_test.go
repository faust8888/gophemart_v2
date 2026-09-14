package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/faust8888/gophemart_v2/internal/model"
	"github.com/jackc/pgx/v5"
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
	found = false
	for _, name := range names {
		if name == "002_create_orders.sql" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("migrationFileNames() = %v, want 002_create_orders.sql", names)
	}
	found = false
	for _, name := range names {
		if name == "003_create_withdrawals.sql" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("migrationFileNames() = %v, want 003_create_withdrawals.sql", names)
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

func TestIsNoRows(t *testing.T) {
	if isNoRows(errors.New("other")) {
		t.Fatal("isNoRows(generic) = true, want false")
	}
	if isNoRows(nil) {
		t.Fatal("isNoRows(nil) = true, want false")
	}
	if !isNoRows(pgx.ErrNoRows) {
		t.Fatal("isNoRows(pgx.ErrNoRows) = false, want true")
	}
}

func TestClose_NilSafe(t *testing.T) {
	var s *Storage
	s.Close()

	empty := &Storage{}
	empty.Close()
}

func TestOrderOwnerConflict(t *testing.T) {
	if err := orderOwnerConflict("user-1", "user-1"); !errors.Is(err, model.ErrOrderAlreadyUploaded) {
		t.Fatalf("same user error = %v, want %v", err, model.ErrOrderAlreadyUploaded)
	}
	if err := orderOwnerConflict("other", "user-1"); !errors.Is(err, model.ErrOrderConflict) {
		t.Fatalf("other user error = %v, want %v", err, model.ErrOrderConflict)
	}
}

func TestHasInsufficientFunds(t *testing.T) {
	if hasInsufficientFunds(751, 751) {
		t.Fatal("equal amounts should be sufficient")
	}
	if !hasInsufficientFunds(10, 751) {
		t.Fatal("smaller current should be insufficient")
	}
	if hasInsufficientFunds(800, 751) {
		t.Fatal("greater current should be sufficient")
	}
}
