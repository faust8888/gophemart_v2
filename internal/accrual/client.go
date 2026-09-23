// Package accrual реализует клиент внешней системы расчёта баллов лояльности.
package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// StatusRegistered означает, что заказ зарегистрирован, но начисление не рассчитано.
	StatusRegistered = "REGISTERED"
	// StatusProcessing означает, что расчёт начисления выполняется.
	StatusProcessing = "PROCESSING"
	// StatusInvalid означает, что заказ не принят к расчёту.
	StatusInvalid = "INVALID"
	// StatusProcessed означает, что расчёт начисления завершён.
	StatusProcessed = "PROCESSED"

	defaultTimeout    = 10 * time.Second
	defaultRetryAfter = 60 * time.Second
	maxRetryAfter     = 60 * time.Second
)

// ErrNotRegistered возвращается, если заказ ещё не появился в системе расчёта.
var ErrNotRegistered = errors.New("order is not registered in accrual system")

// RateLimitError означает, что система расчёта отклонила запрос из-за лимита частоты.
type RateLimitError struct {
	// RetryAfter — пауза до следующей попытки.
	RetryAfter time.Duration
}

// Error реализует интерфейс error.
func (e *RateLimitError) Error() string {
	if e == nil {
		return "accrual rate limit"
	}
	return fmt.Sprintf("accrual rate limit, retry after %s", e.RetryAfter)
}

// OrderInfo содержит сведения о расчёте начисления по заказу.
type OrderInfo struct {
	// Number — номер заказа в системе расчёта.
	Number string
	// Status — статус расчёта начисления.
	Status string
	// Accrual — рассчитанные баллы; nil, если начисления нет.
	Accrual *float64
}

type orderResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual"`
}

// Client обращается к HTTP API системы расчёта начислений.
type Client struct {
	baseURL string
	http    *http.Client
}

// NewClient создаёт клиент системы расчёта. Если httpClient равен nil, используется
// клиент с таймаутом 10 секунд.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}
	return &Client{
		baseURL: normalizeBaseURL(baseURL),
		http:    httpClient,
	}
}

// GetOrder запрашивает информацию о расчёте начисления для номера заказа.
func (c *Client) GetOrder(ctx context.Context, number string) (*OrderInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.orderURL(number), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var body orderResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			return nil, fmt.Errorf("decode accrual response: %w", err)
		}
		return &OrderInfo{
			Number:  body.Order,
			Status:  body.Status,
			Accrual: body.Accrual,
		}, nil
	case http.StatusNoContent:
		return nil, ErrNotRegistered
	case http.StatusTooManyRequests:
		return nil, &RateLimitError{RetryAfter: parseRetryAfter(resp.Header.Get("Retry-After"))}
	default:
		return nil, fmt.Errorf("accrual system status %d", resp.StatusCode)
	}
}

func (c *Client) orderURL(number string) string {
	return c.baseURL + "/api/orders/" + url.PathEscape(number)
}

func normalizeBaseURL(addr string) string {
	addr = strings.TrimSpace(strings.TrimRight(addr, "/"))
	if addr == "" {
		return ""
	}
	if !strings.Contains(addr, "://") {
		return "http://" + addr
	}
	return addr
}

func parseRetryAfter(value string) time.Duration {
	sec, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || sec <= 0 {
		return defaultRetryAfter
	}
	d := time.Duration(sec) * time.Second
	if d > maxRetryAfter {
		return maxRetryAfter
	}
	return d
}
