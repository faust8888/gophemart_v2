package postgres

import (
	"context"

	"github.com/faust8888/gophemart_v2/internal/model"
)

// CreateOrder сохраняет новый номер заказа со статусом NEW.
// При повторной загрузке тем же пользователем возвращает [model.ErrOrderAlreadyUploaded],
// при загрузке другим пользователем — [model.ErrOrderConflict].
func (s *Storage) CreateOrder(ctx context.Context, userID, number string) (*model.Order, error) {
	const q = `
		INSERT INTO orders (number, user_id, status)
		VALUES ($1, $2::uuid, $3)
		RETURNING number, user_id::text, status, uploaded_at
	`

	var order model.Order
	err := s.pool.QueryRow(ctx, q, number, userID, model.OrderStatusNew).Scan(
		&order.Number,
		&order.UserID,
		&order.Status,
		&order.UploadedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			existing, getErr := s.getOrderByNumber(ctx, number)
			if getErr != nil {
				return nil, getErr
			}
			return existing, orderOwnerConflict(existing.UserID, userID)
		}
		return nil, err
	}
	return &order, nil
}

func (s *Storage) getOrderByNumber(ctx context.Context, number string) (*model.Order, error) {
	const q = `
		SELECT number, user_id::text, status, uploaded_at
		FROM orders
		WHERE number = $1
	`

	var order model.Order
	err := s.pool.QueryRow(ctx, q, number).Scan(
		&order.Number,
		&order.UserID,
		&order.Status,
		&order.UploadedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return &order, nil
}

func orderOwnerConflict(existingUserID, userID string) error {
	if existingUserID == userID {
		return model.ErrOrderAlreadyUploaded
	}
	return model.ErrOrderConflict
}

// ListByUser возвращает заказы пользователя, отсортированные по времени загрузки
// от самых новых к самым старым.
func (s *Storage) ListByUser(ctx context.Context, userID string) ([]model.Order, error) {
	const q = `
		SELECT number, user_id::text, status, accrual, uploaded_at
		FROM orders
		WHERE user_id = $1::uuid
		ORDER BY uploaded_at DESC
	`

	rows, err := s.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]model.Order, 0)
	for rows.Next() {
		var order model.Order
		if err := rows.Scan(
			&order.Number,
			&order.UserID,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}
