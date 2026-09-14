package postgres

import (
	"context"
	"errors"

	"github.com/faust8888/gophemart_v2/internal/model"
	"github.com/jackc/pgx/v5/pgconn"
)

const uniqueViolationCode = "23505"

// Create сохраняет нового пользователя с уникальным логином.
// При конфликте логина возвращает [model.ErrLoginTaken].
func (s *Storage) Create(ctx context.Context, login, passwordHash string) (*model.User, error) {
	const q = `
		INSERT INTO users (login, password_hash)
		VALUES ($1, $2)
		RETURNING id::text, login, password_hash, created_at
	`

	var user model.User
	err := s.pool.QueryRow(ctx, q, login, passwordHash).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, model.ErrLoginTaken
		}
		return nil, err
	}
	return &user, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode
}
