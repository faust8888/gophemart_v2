package service

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

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

func (f *fakeOrderRepo) ListByUser(_ context.Context, userID string) ([]model.Order, error) {
	if f.err != nil {
		return nil, f.err
	}
	out := make([]model.Order, 0)
	for _, order := range f.orders {
		if order.UserID == userID {
			out = append(out, *order)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].UploadedAt.After(out[j].UploadedAt)
	})
	return out, nil
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

func TestList_Success(t *testing.T) {
	newer := time.Date(2020, 12, 10, 15, 15, 45, 0, time.FixedZone("MSK", 3*3600))
	older := time.Date(2020, 12, 10, 15, 12, 1, 0, time.FixedZone("MSK", 3*3600))
	accrual := 500.0
	svc := NewOrderService(&fakeOrderRepo{orders: map[string]*model.Order{
		"old":   {Number: "12345678903", UserID: "user-1", Status: model.OrderStatusProcessing, UploadedAt: older},
		"new":   {Number: "9278923470", UserID: "user-1", Status: model.OrderStatusProcessed, Accrual: &accrual, UploadedAt: newer},
		"other": {Number: "346436439", UserID: "other", Status: model.OrderStatusInvalid, UploadedAt: newer},
	}})

	got, err := svc.List(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("List() len = %d, want 2", len(got))
	}
	if got[0].Number != "9278923470" || got[1].Number != "12345678903" {
		t.Fatalf("List() order = %q, %q", got[0].Number, got[1].Number)
	}
}

func TestList_Empty(t *testing.T) {
	svc := NewOrderService(&fakeOrderRepo{orders: make(map[string]*model.Order)})
	got, err := svc.List(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("List() len = %d, want 0", len(got))
	}
}

func TestList_EmptyUser(t *testing.T) {
	svc := NewOrderService(&fakeOrderRepo{orders: make(map[string]*model.Order)})
	if _, err := svc.List(context.Background(), ""); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("List() error = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestList_RepositoryError(t *testing.T) {
	want := errors.New("db down")
	svc := NewOrderService(&fakeOrderRepo{orders: make(map[string]*model.Order), err: want})
	if _, err := svc.List(context.Background(), "user-1"); !errors.Is(err, want) {
		t.Fatalf("List() error = %v, want %v", err, want)
	}
}
