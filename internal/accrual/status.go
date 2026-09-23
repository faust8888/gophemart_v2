package accrual

import (
	"github.com/faust8888/gophemart_v2/internal/model"
)

// LocalStatus преобразует статус системы расчёта в статус заказа Гофермарта.
// Второй результат равен false, если статус неизвестен.
func LocalStatus(accrualStatus string) (string, bool) {
	switch accrualStatus {
	case StatusRegistered, StatusProcessing:
		return model.OrderStatusProcessing, true
	case StatusInvalid:
		return model.OrderStatusInvalid, true
	case StatusProcessed:
		return model.OrderStatusProcessed, true
	default:
		return "", false
	}
}
