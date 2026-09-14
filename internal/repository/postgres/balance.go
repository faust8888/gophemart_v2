package postgres

import (
	"context"

	"github.com/faust8888/gophemart_v2/internal/model"
	"github.com/jackc/pgx/v5"
)

const balanceQuery = `
	SELECT
		COALESCE((
			SELECT SUM(accrual)
			FROM orders
			WHERE user_id = $1::uuid
			  AND status = $2
			  AND accrual IS NOT NULL
		), 0)
		-
		COALESCE((
			SELECT SUM(amount)
			FROM withdrawals
			WHERE user_id = $1::uuid
		), 0) AS current,
		COALESCE((
			SELECT SUM(amount)
			FROM withdrawals
			WHERE user_id = $1::uuid
		), 0) AS withdrawn
`

type queryRower interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// GetBalance возвращает текущий остаток баллов и сумму списаний пользователя.
// Остаток считается как сумма начислений по обработанным заказам минус сумма списаний.
func (s *Storage) GetBalance(ctx context.Context, userID string) (*model.Balance, error) {
	return scanBalance(ctx, s.pool, userID)
}

func scanBalance(ctx context.Context, q queryRower, userID string) (*model.Balance, error) {
	var balance model.Balance
	err := q.QueryRow(ctx, balanceQuery, userID, model.OrderStatusProcessed).Scan(
		&balance.Current,
		&balance.Withdrawn,
	)
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

func hasInsufficientFunds(current, amount float64) bool {
	return current < amount
}
