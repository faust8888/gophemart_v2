package service

import (
	"context"
	"errors"
	"strings"

	"github.com/faust8888/gophemart_v2/internal/luhn"
	"github.com/faust8888/gophemart_v2/internal/model"
)

// ErrInvalidOrderNumber возвращается, если номер заказа не проходит проверку Луна.
var ErrInvalidOrderNumber = errors.New("invalid order number")

// OrderRepository описывает сохранение номеров заказов.
type OrderRepository interface {
	// CreateOrder сохраняет номер заказа, привязанный к пользователю.
	CreateOrder(ctx context.Context, userID, number string) (*model.Order, error)
	// ListByUser возвращает заказы пользователя, от новых к старым.
	ListByUser(ctx context.Context, userID string) ([]model.Order, error)
}

// OrderService реализует сценарии загрузки номеров заказов.
type OrderService struct {
	repo OrderRepository
}

// NewOrderService создаёт сервис загрузки номеров заказов.
func NewOrderService(repo OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

// Upload принимает номер заказа от аутентифицированного пользователя.
// Пустой номер считается неверным форматом запроса, номер, не проходящий
// алгоритм Луна, — неверным форматом номера.
func (s *OrderService) Upload(ctx context.Context, userID, number string) error {
	if userID == "" {
		return ErrInvalidCredentials
	}

	number = strings.TrimSpace(number)
	if number == "" {
		return ErrInvalidInput
	}
	if !luhn.Valid(number) {
		return ErrInvalidOrderNumber
	}

	_, err := s.repo.CreateOrder(ctx, userID, number)
	return err
}

// List возвращает заказы пользователя, отсортированные от новых к старым.
func (s *OrderService) List(ctx context.Context, userID string) ([]model.Order, error) {
	if userID == "" {
		return nil, ErrInvalidCredentials
	}
	return s.repo.ListByUser(ctx, userID)
}
