package handler

import (
	"net/http"
	"strings"

	"github.com/faust8888/gophemart_v2/internal/auth"
)

const bearerPrefix = "Bearer "

type tokenParser interface {
	ParseToken(tokenString string) (userID string, login string, err error)
}

func (h *Handler) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userID, _, err := h.tokens.ParseToken(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := auth.ContextWithUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractToken(r *http.Request) string {
	if cookie, err := r.Cookie(AuthCookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(header), strings.ToLower(bearerPrefix)) {
		return strings.TrimSpace(header[len(bearerPrefix):])
	}
	return header
}
