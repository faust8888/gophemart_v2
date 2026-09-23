package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/faust8888/gophemart_v2/internal/auth"
	"github.com/faust8888/gophemart_v2/internal/model"
	"github.com/faust8888/gophemart_v2/internal/service"
)

// GetBalance возвращает текущий баланс баллов лояльности пользователя.
// Возможные коды ответа: 200, 401, 500.
func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	balance, err := h.balances.Get(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, balance)
}

type withdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

// Withdraw списывает баллы с накопительного счёта в счёт оплаты нового заказа.
// Возможные коды ответа: 200, 400, 401, 402, 422, 500.
func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	var req withdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := h.balances.Withdraw(r.Context(), userID, req.Order, req.Sum)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, service.ErrInvalidInput):
		w.WriteHeader(http.StatusBadRequest)
	case errors.Is(err, service.ErrInvalidCredentials):
		w.WriteHeader(http.StatusUnauthorized)
	case errors.Is(err, model.ErrInsufficientFunds):
		w.WriteHeader(http.StatusPaymentRequired)
	case errors.Is(err, service.ErrInvalidOrderNumber), errors.Is(err, model.ErrOrderAlreadyWithdrawn):
		w.WriteHeader(http.StatusUnprocessableEntity)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

type withdrawalResponse struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

// ListWithdrawals возвращает информацию о списаниях баллов пользователя.
// Возможные коды ответа: 200, 204, 401, 500.
func (h *Handler) ListWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	items, err := h.balances.ListWithdrawals(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(items) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp := make([]withdrawalResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toWithdrawalResponse(item))
	}
	writeJSON(w, http.StatusOK, resp)
}

func toWithdrawalResponse(item model.Withdrawal) withdrawalResponse {
	return withdrawalResponse{
		Order:       item.OrderNumber,
		Sum:         item.Amount,
		ProcessedAt: item.ProcessedAt.Format(time.RFC3339),
	}
}
