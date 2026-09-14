// Package handler предоставляет HTTP-обработчики API накопительной системы лояльности.
package handler

import (
	"context"
	"net/http"

	"github.com/faust8888/gophemart_v2/internal/model"
)

const (
	// AuthCookieName — имя cookie с токеном аутентификации.
	AuthCookieName = "gophermart_auth"
	maxBodyBytes   = 1 << 20
)

type userService interface {
	Register(ctx context.Context, login, password string) (*model.User, string, error)
}

// Handler обрабатывает HTTP-запросы API системы лояльности.
type Handler struct {
	users userService
}

// New создаёт HTTP-обработчик, использующий переданный сервис пользователей.
func New(users userService) *Handler {
	return &Handler{users: users}
}

// Routes возвращает маршрутизатор HTTP API.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/user/register", h.Register)
	return recoverMiddleware(mux)
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func setAuth(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     AuthCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.Header().Set("Authorization", "Bearer "+token)
}
