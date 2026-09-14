package handler

import (
	"errors"
	"net/http"

	"github.com/faust8888/gophemart_v2/internal/service"
)

// Login аутентифицирует пользователя по паре логин/пароль.
// Возможные коды ответа: 200, 400, 401, 500.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	req, err := decodeCredentials(w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, token, err := h.users.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, service.ErrInvalidCredentials):
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	setAuth(w, token)
	w.WriteHeader(http.StatusOK)
}
