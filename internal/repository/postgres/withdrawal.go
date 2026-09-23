package postgres

import (
	"context"

	"github.com/faust8888/gophemart_v2/internal/model"
)

// CreateWithdrawal регистрирует списание баллов в счёт оплаты заказа.
// Операция выполняется в транзакции: строка пользователя блокируется,
// проверяется достаточность средств, затем создаётся запись списания.
// Повторное списание по тому же order_number возвращает [model.ErrOrderAlreadyWithdrawn].
func (s *Storage) CreateWithdrawal(ctx context.Context, userID, orderNumber string, amount float64) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var lockedUserID string
	err = tx.QueryRow(ctx, `SELECT id::text FROM users WHERE id = $1::uuid FOR UPDATE`, userID).Scan(&lockedUserID)
	if err != nil {
		if isNoRows(err) {
			return model.ErrNotFound
		}
		return err
	}

	balance, err := scanBalance(ctx, tx, userID)
	if err != nil {
		return err
	}
	if hasInsufficientFunds(balance.Current, amount) {
		return model.ErrInsufficientFunds
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO withdrawals (user_id, order_number, amount)
		VALUES ($1::uuid, $2, $3)
	`, userID, orderNumber, amount)
	if err != nil {
		return withdrawalInsertError(err)
	}
	return tx.Commit(ctx)
}

func withdrawalInsertError(err error) error {
	if isUniqueViolation(err) {
		return model.ErrOrderAlreadyWithdrawn
	}
	return err
}

// ListWithdrawals возвращает списания пользователя, отсортированные по времени
// обработки от самых новых к самым старым.
func (s *Storage) ListWithdrawals(ctx context.Context, userID string) ([]model.Withdrawal, error) {
	const q = `
		SELECT order_number, user_id::text, amount, processed_at
		FROM withdrawals
		WHERE user_id = $1::uuid
		ORDER BY processed_at DESC
	`

	rows, err := s.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	withdrawals := make([]model.Withdrawal, 0)
	for rows.Next() {
		var item model.Withdrawal
		if err := rows.Scan(
			&item.OrderNumber,
			&item.UserID,
			&item.Amount,
			&item.ProcessedAt,
		); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return withdrawals, nil
}
