package service

import (
	"context"
	"errors"
	"testing"

	"github.com/faust8888/gophemart_v2/internal/model"
)

type fakeOrderRepo struct {
	orders map[string]*model.Order
	err    error
}

func (f *fakeOrderRepo) CreateOrder(_ context.Context, userID, number string) (*model.Order, error) {
	if f.err != nil {
		return nil, f.err
	}
	if existing, ok := f.orders[number]; ok {
		if existing.UserID == userID {
			return existing, model.ErrOrderAlreadyUploaded
		}
		return existing, model.ErrOrderConflict
	}
	order := &model.Order{
		Number: number,
		UserID: userID,
		Status: model.OrderStatusNew,
	}
	f.orders[number] = order
	return order, nil
}

func TestUpload_Success(t *testing.T) {
	svc := NewOrderService(&fakeOrderRepo{orders: make(map[string]*model.Order)})
	if err := svc.Upload(context.Background(), "user-1", " 12345678903 "); err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
}

func TestUpload_InvalidInput(t *testing.T) {
	svc := NewOrderService(&fakeOrderRepo{orders: make(map[string]*model.Order)})
	if err := svc.Upload(context.Background(), "user-1", "   "); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Upload() error = %v, want %v", err, ErrInvalidInput)
	}
}

func TestUpload_EmptyUser(t *testing.T) {
	svc := NewOrderService(&fakeOrderRepo{orders: make(map[string]*model.Order)})
	if err := svc.Upload(context.Background(), "", "12345678903"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Upload() error = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestUpload_InvalidNumber(t *testing.T) {
	svc := NewOrderService(&fakeOrderRepo{orders: make(map[string]*model.Order)})
	if err := svc.Upload(context.Background(), "user-1", "123"); !errors.Is(err, ErrInvalidOrderNumber) {
		t.Fatalf("Upload() error = %v, want %v", err, ErrInvalidOrderNumber)
	}
}

func TestUpload_AlreadyUploaded(t *testing.T) {
	svc := NewOrderService(&fakeOrderRepo{orders: map[string]*model.Order{
		"12345678903": {Number: "12345678903", UserID: "user-1"},
	}})
	if err := svc.Upload(context.Background(), "user-1", "12345678903"); !errors.Is(err, model.ErrOrderAlreadyUploaded) {
		t.Fatalf("Upload() error = %v, want %v", err, model.ErrOrderAlreadyUploaded)
	}
}

func TestUpload_Conflict(t *testing.T) {
	svc := NewOrderService(&fakeOrderRepo{orders: map[string]*model.Order{
		"12345678903": {Number: "12345678903", UserID: "other"},
	}})
	if err := svc.Upload(context.Background(), "user-1", "12345678903"); !errors.Is(err, model.ErrOrderConflict) {
		t.Fatalf("Upload() error = %v, want %v", err, model.ErrOrderConflict)
	}
}

func TestUpload_RepositoryError(t *testing.T) {
	want := errors.New("db down")
	svc := NewOrderService(&fakeOrderRepo{orders: make(map[string]*model.Order), err: want})
	if err := svc.Upload(context.Background(), "user-1", "12345678903"); !errors.Is(err, want) {
		t.Fatalf("Upload() error = %v, want %v", err, want)
	}
}
