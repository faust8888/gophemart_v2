package service

import (
	"context"
	"errors"
	"testing"

	"github.com/faust8888/gophemart_v2/internal/model"
)

type fakeBalanceRepo struct {
	balance *model.Balance
	err     error
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
