package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/faust8888/gophemart_v2/internal/auth"
	"github.com/faust8888/gophemart_v2/internal/model"
	"github.com/faust8888/gophemart_v2/internal/service"
)

// UploadOrder принимает номер заказа от аутентифицированного пользователя.
// Возможные коды ответа: 200, 202, 400, 401, 409, 422, 500.
func (h *Handler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.orders.Upload(r.Context(), userID, string(body))
	switch {
	case err == nil:
		w.WriteHeader(http.StatusAccepted)
	case errors.Is(err, service.ErrInvalidInput):
		w.WriteHeader(http.StatusBadRequest)
	case errors.Is(err, service.ErrInvalidCredentials):
		w.WriteHeader(http.StatusUnauthorized)
	case errors.Is(err, service.ErrInvalidOrderNumber):
		w.WriteHeader(http.StatusUnprocessableEntity)
	case errors.Is(err, model.ErrOrderAlreadyUploaded):
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, model.ErrOrderConflict):
		w.WriteHeader(http.StatusConflict)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}
