package service

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/faust8888/gophemart_v2/internal/model"
)

type fakeBalanceRepo struct {
	balance     *model.Balance
	withdrawals []model.Withdrawal
	err         error
}

func (f *fakeBalanceRepo) GetBalance(_ context.Context, _ string) (*model.Balance, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.balance, nil
}

func (f *fakeBalanceRepo) CreateWithdrawal(_ context.Context, _, _ string, amount float64) error {
	if f.err != nil {
		return f.err
	}
	if f.balance == nil || f.balance.Current < amount {
		return model.ErrInsufficientFunds
	}
	f.balance.Current -= amount
	f.balance.Withdrawn += amount
	return nil
}

func (f *fakeBalanceRepo) ListWithdrawals(_ context.Context, userID string) ([]model.Withdrawal, error) {
	if f.err != nil {
		return nil, f.err
	}
	out := make([]model.Withdrawal, 0)
	for _, item := range f.withdrawals {
		if item.UserID == userID {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ProcessedAt.After(out[j].ProcessedAt)
	})
	return out, nil
}

func TestBalanceGet_Success(t *testing.T) {
	want := &model.Balance{Current: 500.5, Withdrawn: 42}
	svc := NewBalanceService(&fakeBalanceRepo{balance: want})

	got, err := svc.Get(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Current != want.Current || got.Withdrawn != want.Withdrawn {
		t.Fatalf("Get() = %+v, want %+v", got, want)
	}
}

func TestBalanceGet_EmptyUser(t *testing.T) {
	svc := NewBalanceService(&fakeBalanceRepo{balance: &model.Balance{}})
	if _, err := svc.Get(context.Background(), ""); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Get() error = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestBalanceGet_RepositoryError(t *testing.T) {
	want := errors.New("db down")
	svc := NewBalanceService(&fakeBalanceRepo{err: want})
	if _, err := svc.Get(context.Background(), "user-1"); !errors.Is(err, want) {
		t.Fatalf("Get() error = %v, want %v", err, want)
	}
}

func TestWithdraw_Success(t *testing.T) {
	repo := &fakeBalanceRepo{balance: &model.Balance{Current: 800, Withdrawn: 10}}
	svc := NewBalanceService(repo)

	if err := svc.Withdraw(context.Background(), "user-1", " 12345678903 ", 751); err != nil {
		t.Fatalf("Withdraw() error = %v", err)
	}
	if repo.balance.Current != 49 || repo.balance.Withdrawn != 761 {
		t.Fatalf("balance after withdraw = %+v", repo.balance)
	}
}

func TestWithdraw_EmptyUser(t *testing.T) {
	svc := NewBalanceService(&fakeBalanceRepo{balance: &model.Balance{Current: 800}})
	if err := svc.Withdraw(context.Background(), "", "12345678903", 10); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Withdraw() error = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestWithdraw_InvalidAmount(t *testing.T) {
	svc := NewBalanceService(&fakeBalanceRepo{balance: &model.Balance{Current: 800}})
	if err := svc.Withdraw(context.Background(), "user-1", "12345678903", 0); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Withdraw() error = %v, want %v", err, ErrInvalidInput)
	}
}

func TestWithdraw_InvalidOrder(t *testing.T) {
	svc := NewBalanceService(&fakeBalanceRepo{balance: &model.Balance{Current: 800}})
	if err := svc.Withdraw(context.Background(), "user-1", "123", 10); !errors.Is(err, ErrInvalidOrderNumber) {
		t.Fatalf("Withdraw() error = %v, want %v", err, ErrInvalidOrderNumber)
	}
}

func TestWithdraw_InsufficientFunds(t *testing.T) {
	svc := NewBalanceService(&fakeBalanceRepo{balance: &model.Balance{Current: 10}})
	if err := svc.Withdraw(context.Background(), "user-1", "12345678903", 751); !errors.Is(err, model.ErrInsufficientFunds) {
		t.Fatalf("Withdraw() error = %v, want %v", err, model.ErrInsufficientFunds)
	}
}

func TestWithdraw_RepositoryError(t *testing.T) {
	want := errors.New("db down")
	svc := NewBalanceService(&fakeBalanceRepo{balance: &model.Balance{Current: 800}, err: want})
	if err := svc.Withdraw(context.Background(), "user-1", "12345678903", 10); !errors.Is(err, want) {
		t.Fatalf("Withdraw() error = %v, want %v", err, want)
	}
}

func TestListWithdrawals_Success(t *testing.T) {
	newer := time.Date(2020, 12, 9, 16, 9, 57, 0, time.FixedZone("MSK", 3*3600))
	older := time.Date(2020, 12, 8, 10, 0, 0, 0, time.FixedZone("MSK", 3*3600))
	svc := NewBalanceService(&fakeBalanceRepo{withdrawals: []model.Withdrawal{
		{OrderNumber: "111", UserID: "user-1", Amount: 100, ProcessedAt: older},
		{OrderNumber: "2377225624", UserID: "user-1", Amount: 500, ProcessedAt: newer},
		{OrderNumber: "999", UserID: "other", Amount: 50, ProcessedAt: newer},
	}})

	got, err := svc.ListWithdrawals(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("ListWithdrawals() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("ListWithdrawals() len = %d, want 2", len(got))
	}
	if got[0].OrderNumber != "2377225624" || got[1].OrderNumber != "111" {
		t.Fatalf("ListWithdrawals() order = %q, %q", got[0].OrderNumber, got[1].OrderNumber)
	}
}

func TestListWithdrawals_Empty(t *testing.T) {
	svc := NewBalanceService(&fakeBalanceRepo{})
	got, err := svc.ListWithdrawals(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("ListWithdrawals() error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("ListWithdrawals() len = %d, want 0", len(got))
	}
}

func TestListWithdrawals_EmptyUser(t *testing.T) {
	svc := NewBalanceService(&fakeBalanceRepo{})
	if _, err := svc.ListWithdrawals(context.Background(), ""); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("ListWithdrawals() error = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestListWithdrawals_RepositoryError(t *testing.T) {
	want := errors.New("db down")
	svc := NewBalanceService(&fakeBalanceRepo{err: want})
	if _, err := svc.ListWithdrawals(context.Background(), "user-1"); !errors.Is(err, want) {
		t.Fatalf("ListWithdrawals() error = %v, want %v", err, want)
	}
}
