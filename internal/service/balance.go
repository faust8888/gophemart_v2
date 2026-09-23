package service

import (
	"context"
	"strings"

	"github.com/faust8888/gophemart_v2/internal/luhn"
	"github.com/faust8888/gophemart_v2/internal/model"
)

// BalanceRepository описывает получение баланса и регистрацию списаний баллов лояльности.
type BalanceRepository interface {
	// GetBalance возвращает текущий остаток и сумму списаний пользователя.
	GetBalance(ctx context.Context, userID string) (*model.Balance, error)
	// CreateWithdrawal регистрирует списание баллов в счёт оплаты заказа.
	CreateWithdrawal(ctx context.Context, userID, orderNumber string, amount float64) error
	// ListWithdrawals возвращает списания пользователя, от новых к старым.
	ListWithdrawals(ctx context.Context, userID string) ([]model.Withdrawal, error)
}

// BalanceService реализует сценарии работы с балансом баллов лояльности.
type BalanceService struct {
	repo BalanceRepository
}

// NewBalanceService создаёт сервис баланса баллов лояльности.
func NewBalanceService(repo BalanceRepository) *BalanceService {
	return &BalanceService{repo: repo}
}

// Get возвращает текущий остаток баллов и сумму списаний за весь период регистрации.
func (s *BalanceService) Get(ctx context.Context, userID string) (*model.Balance, error) {
	if userID == "" {
		return nil, ErrInvalidCredentials
	}
	return s.repo.GetBalance(ctx, userID)
}

// Withdraw списывает баллы с накопительного счёта в счёт оплаты заказа.
func (s *BalanceService) Withdraw(ctx context.Context, userID, orderNumber string, amount float64) error {
	if userID == "" {
		return ErrInvalidCredentials
	}

	orderNumber = strings.TrimSpace(orderNumber)
	if amount <= 0 {
		return ErrInvalidInput
	}
	if !luhn.Valid(orderNumber) {
		return ErrInvalidOrderNumber
	}

	return s.repo.CreateWithdrawal(ctx, userID, orderNumber, amount)
}

// ListWithdrawals возвращает списания пользователя, отсортированные от новых к старым.
func (s *BalanceService) ListWithdrawals(ctx context.Context, userID string) ([]model.Withdrawal, error) {
	if userID == "" {
		return nil, ErrInvalidCredentials
	}
	return s.repo.ListWithdrawals(ctx, userID)
}
