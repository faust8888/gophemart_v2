package model

import "time"

// Статусы обработки заказа в системе лояльности.
const (
	// OrderStatusNew означает, что заказ загружен, но ещё не попал в обработку.
	OrderStatusNew = "NEW"
	// OrderStatusProcessing означает, что вознаграждение за заказ рассчитывается.
	OrderStatusProcessing = "PROCESSING"
	// OrderStatusInvalid означает, что система расчёта отказала в начислении.
	OrderStatusInvalid = "INVALID"
	// OrderStatusProcessed означает, что расчёт успешно получен.
	OrderStatusProcessed = "PROCESSED"
)

// Order представляет загруженный пользователем номер заказа.
type Order struct {
	// Number — уникальный номер заказа.
	Number string
	// UserID — идентификатор пользователя, загрузившего заказ.
	UserID string
	// Status — статус обработки расчёта баллов.
	Status string
	// Accrual — начисленные баллы; отсутствует, если начисления нет.
	Accrual *float64
	// UploadedAt — время загрузки номера заказа.
	UploadedAt time.Time
}
