// Package auth выпускает и проверяет JWT-токены аутентификации пользователей.
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken возвращается, если токен отсутствует, повреждён или просрочен.
var ErrInvalidToken = errors.New("invalid auth token")

type claims struct {
	UserID string `json:"user_id"`
	Login  string `json:"login"`
	jwt.RegisteredClaims
}

// Manager выпускает и валидирует JWT-токены доступа.
type Manager struct {
	secret []byte
	ttl    time.Duration
}

// NewManager создаёт менеджер JWT с указанным секретом и временем жизни токена.
// Если ttl неположительный, используется срок жизни 24 часа.
func NewManager(secret string, ttl time.Duration) (*Manager, error) {
	if secret == "" {
		return nil, errors.New("auth secret is empty")
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &Manager{
		secret: []byte(secret),
		ttl:    ttl,
	}, nil
}

// IssueToken выпускает подписанный JWT для пользователя с идентификатором userID и логином login.
func (m *Manager) IssueToken(userID, login string) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		UserID: userID,
		Login:  login,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	})
	return token.SignedString(m.secret)
}

// ParseToken проверяет подпись и срок действия токена и возвращает идентификатор и логин пользователя.
func (m *Manager) ParseToken(tokenString string) (userID string, login string, err error) {
	if tokenString == "" {
		return "", "", ErrInvalidToken
	}

	parsed, err := jwt.ParseWithClaims(tokenString, &claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return "", "", ErrInvalidToken
	}

	c, ok := parsed.Claims.(*claims)
	if !ok || !parsed.Valid || c.UserID == "" {
		return "", "", ErrInvalidToken
	}
	return c.UserID, c.Login, nil
}
