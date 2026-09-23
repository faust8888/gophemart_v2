package handler

import (
	"errors"
	"net/http"

	"github.com/faust8888/gophemart_v2/internal/model"
	"github.com/faust8888/gophemart_v2/internal/service"
)

// Register регистрирует пользователя по паре логин/пароль и сразу аутентифицирует его.
// Возможные коды ответа: 200, 400, 409, 500.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	req, err := decodeCredentials(w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, token, err := h.users.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, model.ErrLoginTaken):
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	setAuth(w, token)
	w.WriteHeader(http.StatusOK)
}
