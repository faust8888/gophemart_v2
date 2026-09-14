package postgres

import (
	"context"

	"github.com/faust8888/gophemart_v2/internal/model"
)

// CreateWithdrawal регистрирует списание баллов в счёт оплаты заказа.
// Операция выполняется в транзакции: строка пользователя блокируется,
// проверяется достаточность средств, затем создаётся запись списания.
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
		return err
	}
	return tx.Commit(ctx)
}
