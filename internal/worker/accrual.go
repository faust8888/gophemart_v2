// Package worker выполняет фоновые задачи накопительной системы лояльности.
package worker

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/faust8888/gophemart_v2/internal/accrual"
	"github.com/faust8888/gophemart_v2/internal/model"
)

const (
	// DefaultPollInterval — период опроса системы расчёта начислений.
	DefaultPollInterval = time.Second
)

// OrderRepository описывает выборку и обновление заказов, ожидающих начисления.
type OrderRepository interface {
	// ListUnprocessed возвращает заказы в статусах NEW и PROCESSING.
	ListUnprocessed(ctx context.Context) ([]model.Order, error)
	// UpdateAccrual сохраняет статус и сумму начисления по номеру заказа.
	UpdateAccrual(ctx context.Context, number, status string, accrual *float64) error
}

// AccrualClient запрашивает сведения о начислении у внешней системы расчёта.
type AccrualClient interface {
	// GetOrder возвращает информацию о расчёте начисления для номера заказа.
	GetOrder(ctx context.Context, number string) (*accrual.OrderInfo, error)
}

// AccrualPoller периодически опрашивает систему расчёта и обновляет статусы заказов.
type AccrualPoller struct {
	orders   OrderRepository
	client   AccrualClient
	interval time.Duration
	sleep    func(ctx context.Context, d time.Duration)
}

// NewAccrualPoller создаёт воркер опроса системы расчёта начислений.
// Если interval неположителен, используется [DefaultPollInterval].
func NewAccrualPoller(orders OrderRepository, client AccrualClient, interval time.Duration) *AccrualPoller {
	if interval <= 0 {
		interval = DefaultPollInterval
	}
	return &AccrualPoller{
		orders:   orders,
		client:   client,
		interval: interval,
		sleep:    sleepContext,
	}
}

// Run опрашивает систему расчёта до отмены контекста.
func (p *AccrualPoller) Run(ctx context.Context) {
	p.tick(ctx)

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.tick(ctx)
		}
	}
}

func (p *AccrualPoller) tick(ctx context.Context) {
	orders, err := p.orders.ListUnprocessed(ctx)
	if err != nil {
		slog.Error("list unprocessed orders", slog.Any("err", err))
		return
	}

	for _, order := range orders {
		if ctx.Err() != nil {
			return
		}
		if err := p.process(ctx, order); err != nil {
			var rate *accrual.RateLimitError
			if errors.As(err, &rate) {
				p.wait(ctx, rate.RetryAfter)
				return
			}
			slog.Error("poll accrual order", slog.String("number", order.Number), slog.Any("err", err))
		}
	}
}

func (p *AccrualPoller) process(ctx context.Context, order model.Order) error {
	info, err := p.client.GetOrder(ctx, order.Number)
	if err != nil {
		if errors.Is(err, accrual.ErrNotRegistered) {
			return nil
		}
		return err
	}

	status, ok := accrual.LocalStatus(info.Status)
	if !ok {
		return nil
	}
	return p.orders.UpdateAccrual(ctx, order.Number, status, info.Accrual)
}

func (p *AccrualPoller) wait(ctx context.Context, d time.Duration) {
	if p.sleep != nil {
		p.sleep(ctx, d)
		return
	}
	sleepContext(ctx, d)
}

func sleepContext(ctx context.Context, d time.Duration) {
	if d <= 0 {
		return
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}
