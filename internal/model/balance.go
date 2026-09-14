package model

// Balance содержит текущий остаток баллов лояльности и сумму списаний за всё время.
type Balance struct {
	// Current — доступный остаток баллов.
	Current float64 `json:"current"`
	// Withdrawn — сумма списанных баллов за весь период регистрации.
	Withdrawn float64 `json:"withdrawn"`
}
