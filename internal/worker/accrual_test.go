package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/faust8888/gophemart_v2/internal/accrual"
	"github.com/faust8888/gophemart_v2/internal/model"
)

type fakeOrders struct {
	items   []model.Order
	updates []updateAccrual
	listErr error
	updErr  error
}

type updateAccrual struct {
	number  string
	status  string
	accrual *float64
}

func (f *fakeOrders) ListUnprocessed(context.Context) ([]model.Order, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.items, nil
}

func (f *fakeOrders) UpdateAccrual(_ context.Context, number, status string, value *float64) error {
	if f.updErr != nil {
		return f.updErr
	}
	f.updates = append(f.updates, updateAccrual{number: number, status: status, accrual: value})
	return nil
}

type fakeClient struct {
	info map[string]*accrual.OrderInfo
	err  map[string]error
}

func (f *fakeClient) GetOrder(_ context.Context, number string) (*accrual.OrderInfo, error) {
	if err := f.err[number]; err != nil {
		return nil, err
	}
	return f.info[number], nil
}

func TestNewAccrualPoller_DefaultInterval(t *testing.T) {
	p := NewAccrualPoller(&fakeOrders{}, &fakeClient{}, 0)
	if p.interval != DefaultPollInterval {
		t.Fatalf("interval = %s, want %s", p.interval, DefaultPollInterval)
	}
}

func TestTick_Processed(t *testing.T) {
	accrualValue := 500.0
	orders := &fakeOrders{items: []model.Order{{Number: "1", Status: model.OrderStatusNew}}}
	p := NewAccrualPoller(orders, &fakeClient{info: map[string]*accrual.OrderInfo{
		"1": {Number: "1", Status: accrual.StatusProcessed, Accrual: &accrualValue},
	}}, time.Second)

	p.tick(context.Background())
	if len(orders.updates) != 1 {
		t.Fatalf("updates = %d, want 1", len(orders.updates))
	}
	if orders.updates[0].status != model.OrderStatusProcessed {
		t.Fatalf("status = %s, want PROCESSED", orders.updates[0].status)
	}
	if orders.updates[0].accrual == nil || *orders.updates[0].accrual != 500 {
		t.Fatalf("accrual = %v, want 500", orders.updates[0].accrual)
	}
}

func TestTick_RegisteredMapsToProcessing(t *testing.T) {
	orders := &fakeOrders{items: []model.Order{{Number: "1"}}}
	p := NewAccrualPoller(orders, &fakeClient{info: map[string]*accrual.OrderInfo{
		"1": {Status: accrual.StatusRegistered},
	}}, time.Second)

	p.tick(context.Background())
	if len(orders.updates) != 1 || orders.updates[0].status != model.OrderStatusProcessing {
		t.Fatalf("updates = %+v, want PROCESSING", orders.updates)
	}
}

func TestTick_Invalid(t *testing.T) {
	orders := &fakeOrders{items: []model.Order{{Number: "1"}}}
	p := NewAccrualPoller(orders, &fakeClient{info: map[string]*accrual.OrderInfo{
		"1": {Status: accrual.StatusInvalid},
	}}, time.Second)

	p.tick(context.Background())
	if len(orders.updates) != 1 || orders.updates[0].status != model.OrderStatusInvalid {
		t.Fatalf("updates = %+v, want INVALID", orders.updates)
	}
}

func TestTick_NotRegistered(t *testing.T) {
	orders := &fakeOrders{items: []model.Order{{Number: "1"}}}
	p := NewAccrualPoller(orders, &fakeClient{err: map[string]error{
		"1": accrual.ErrNotRegistered,
	}}, time.Second)

	p.tick(context.Background())
	if len(orders.updates) != 0 {
		t.Fatalf("updates = %d, want 0", len(orders.updates))
	}
}

func TestTick_UnknownStatus(t *testing.T) {
	orders := &fakeOrders{items: []model.Order{{Number: "1"}}}
	p := NewAccrualPoller(orders, &fakeClient{info: map[string]*accrual.OrderInfo{
		"1": {Status: "OTHER"},
	}}, time.Second)

	p.tick(context.Background())
	if len(orders.updates) != 0 {
		t.Fatalf("updates = %d, want 0", len(orders.updates))
	}
}

func TestTick_RateLimitStopsBatch(t *testing.T) {
	orders := &fakeOrders{items: []model.Order{{Number: "1"}, {Number: "2"}}}
	var slept time.Duration
	p := NewAccrualPoller(orders, &fakeClient{
		err: map[string]error{"1": &accrual.RateLimitError{RetryAfter: 3 * time.Second}},
		info: map[string]*accrual.OrderInfo{
			"2": {Status: accrual.StatusProcessed},
		},
	}, time.Second)
	p.sleep = func(context.Context, time.Duration) {
		slept = 3 * time.Second
	}

	p.tick(context.Background())
	if slept != 3*time.Second {
		t.Fatalf("slept = %s, want 3s", slept)
	}
	if len(orders.updates) != 0 {
		t.Fatalf("updates = %d, want 0 after rate limit", len(orders.updates))
	}
}

func TestTick_ListError(t *testing.T) {
	orders := &fakeOrders{listErr: errors.New("db down")}
	p := NewAccrualPoller(orders, &fakeClient{}, time.Second)
	p.tick(context.Background())
	if len(orders.updates) != 0 {
		t.Fatalf("updates = %d, want 0", len(orders.updates))
	}
}

func TestTick_CancelledContext(t *testing.T) {
	orders := &fakeOrders{items: []model.Order{{Number: "1"}}}
	p := NewAccrualPoller(orders, &fakeClient{info: map[string]*accrual.OrderInfo{
		"1": {Status: accrual.StatusProcessed},
	}}, time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p.tick(ctx)
	if len(orders.updates) != 0 {
		t.Fatalf("updates = %d, want 0", len(orders.updates))
	}
}

func TestRun_StopsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	p := NewAccrualPoller(&fakeOrders{}, &fakeClient{}, time.Hour)
	done := make(chan struct{})
	go func() {
		defer close(done)
		p.Run(ctx)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run() did not stop after cancel")
	}
}

func TestSleepContext_Zero(t *testing.T) {
	sleepContext(context.Background(), 0)
}

func TestProcess_ClientError(t *testing.T) {
	p := NewAccrualPoller(&fakeOrders{items: []model.Order{{Number: "1"}}}, &fakeClient{err: map[string]error{
		"1": errors.New("network"),
	}}, time.Second)
	if err := p.process(context.Background(), model.Order{Number: "1"}); err == nil {
		t.Fatal("process() error = nil, want error")
	}
}
