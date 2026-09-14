package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNormalizeBaseURL(t *testing.T) {
	cases := map[string]string{
		"":                     "",
		"http://accrual:8080":  "http://accrual:8080",
		"http://accrual:8080/": "http://accrual:8080",
		"accrual:8080":         "http://accrual:8080",
	}
	for in, want := range cases {
		if got := normalizeBaseURL(in); got != want {
			t.Errorf("normalizeBaseURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseRetryAfter(t *testing.T) {
	if got := parseRetryAfter("15"); got != 15*time.Second {
		t.Errorf("parseRetryAfter(15) = %s, want 15s", got)
	}
	if got := parseRetryAfter(""); got != defaultRetryAfter {
		t.Errorf("parseRetryAfter(empty) = %s, want default", got)
	}
	if got := parseRetryAfter("3600"); got != maxRetryAfter {
		t.Errorf("parseRetryAfter(3600) = %s, want capped", got)
	}
}

func TestGetOrder_Processed(t *testing.T) {
	accrual := 500.0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/orders/12345678903" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(orderResponse{
			Order:   "12345678903",
			Status:  StatusProcessed,
			Accrual: &accrual,
		})
	}))
	defer srv.Close()

	info, err := NewClient(srv.URL, srv.Client()).GetOrder(context.Background(), "12345678903")
	if err != nil {
		t.Fatalf("GetOrder() error = %v", err)
	}
	if info.Status != StatusProcessed || info.Accrual == nil || *info.Accrual != 500 {
		t.Fatalf("GetOrder() = %+v", info)
	}
}

func TestGetOrder_NoAccrual(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"order":"1","status":"PROCESSED"}`))
	}))
	defer srv.Close()

	info, err := NewClient(srv.URL, srv.Client()).GetOrder(context.Background(), "1")
	if err != nil {
		t.Fatalf("GetOrder() error = %v", err)
	}
	if info.Accrual != nil {
		t.Fatalf("accrual = %v, want nil", info.Accrual)
	}
}

func TestGetOrder_NotRegistered(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	_, err := NewClient(srv.URL, srv.Client()).GetOrder(context.Background(), "1")
	if !errors.Is(err, ErrNotRegistered) {
		t.Fatalf("GetOrder() error = %v, want %v", err, ErrNotRegistered)
	}
}

func TestGetOrder_RateLimited(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "12")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	_, err := NewClient(srv.URL, srv.Client()).GetOrder(context.Background(), "1")
	var rate *RateLimitError
	if !errors.As(err, &rate) {
		t.Fatalf("GetOrder() error = %v, want RateLimitError", err)
	}
	if rate.RetryAfter != 12*time.Second {
		t.Fatalf("RetryAfter = %s, want 12s", rate.RetryAfter)
	}
}

func TestGetOrder_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	if _, err := NewClient(srv.URL, srv.Client()).GetOrder(context.Background(), "1"); err == nil {
		t.Fatal("GetOrder() error = nil, want error")
	}
}

func TestGetOrder_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{"))
	}))
	defer srv.Close()

	if _, err := NewClient(srv.URL, srv.Client()).GetOrder(context.Background(), "1"); err == nil {
		t.Fatal("GetOrder() error = nil, want error")
	}
}

func TestRateLimitError_Nil(t *testing.T) {
	var err *RateLimitError
	if err.Error() == "" {
		t.Fatal("nil RateLimitError.Error() is empty")
	}
}
