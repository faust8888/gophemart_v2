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
	Login(ctx context.Context, login, password string) (*model.User, string, error)
}

type orderService interface {
	Upload(ctx context.Context, userID, number string) error
	List(ctx context.Context, userID string) ([]model.Order, error)
}

// Handler обрабатывает HTTP-запросы API системы лояльности.
type Handler struct {
	users  userService
	orders orderService
	tokens tokenParser
}

// New создаёт HTTP-обработчик API с сервисами пользователей, заказов и проверкой токенов.
func New(users userService, orders orderService, tokens tokenParser) *Handler {
	return &Handler{
		users:  users,
		orders: orders,
		tokens: tokens,
	}
}

// Routes возвращает маршрутизатор HTTP API.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/user/register", h.Register)
	mux.HandleFunc("POST /api/user/login", h.Login)
	mux.Handle("POST /api/user/orders", h.requireAuth(http.HandlerFunc(h.UploadOrder)))
	mux.Handle("GET /api/user/orders", h.requireAuth(http.HandlerFunc(h.ListOrders)))
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
