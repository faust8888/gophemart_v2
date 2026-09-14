package model

import "time"

// Withdrawal представляет списание баллов лояльности в счёт оплаты заказа.
type Withdrawal struct {
	// OrderNumber — номер заказа, в счёт которого списаны баллы.
	OrderNumber string
	// UserID — идентификатор пользователя.
	UserID string
	// Amount — сумма списанных баллов.
	Amount float64
	// ProcessedAt — время списания.
	ProcessedAt time.Time
}
