package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/faust8888/gophemart_v2/internal/auth"
	"github.com/faust8888/gophemart_v2/internal/model"
	"github.com/faust8888/gophemart_v2/internal/service"
)

func TestUploadOrder_Accepted(t *testing.T) {
	h := newOrderHandler(nil, nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodPost, "/api/user/orders", "12345678903"))
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
}

func TestUploadOrder_AlreadyUploaded(t *testing.T) {
	h := newOrderHandler(&mockOrders{err: model.ErrOrderAlreadyUploaded}, nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodPost, "/api/user/orders", "12345678903"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestUploadOrder_Conflict(t *testing.T) {
	h := newOrderHandler(&mockOrders{err: model.ErrOrderConflict}, nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodPost, "/api/user/orders", "12345678903"))
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestUploadOrder_BadRequest(t *testing.T) {
	h := newOrderHandler(&mockOrders{err: service.ErrInvalidInput}, nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodPost, "/api/user/orders", ""))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestUploadOrder_UnprocessableEntity(t *testing.T) {
	h := newOrderHandler(&mockOrders{err: service.ErrInvalidOrderNumber}, nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodPost, "/api/user/orders", "123"))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}
}

func TestUploadOrder_UnauthorizedNoToken(t *testing.T) {
	h := newOrderHandler(nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("12345678903"))
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestUploadOrder_UnauthorizedInvalidToken(t *testing.T) {
	h := newOrderHandler(nil, &mockAuth{err: auth.ErrInvalidToken})
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodPost, "/api/user/orders", "12345678903"))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestUploadOrder_BearerToken(t *testing.T) {
	h := newOrderHandler(nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("12345678903"))
	req.Header.Set("Authorization", "Bearer jwt-token")
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
}

func TestUploadOrder_UnauthorizedCredentials(t *testing.T) {
	h := newOrderHandler(&mockOrders{err: service.ErrInvalidCredentials}, nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodPost, "/api/user/orders", "12345678903"))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestUploadOrder_InternalError(t *testing.T) {
	h := newOrderHandler(&mockOrders{err: errors.New("db down")}, nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodPost, "/api/user/orders", "12345678903"))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestUploadOrder_MissingUserInContext(t *testing.T) {
	h := newOrderHandler(nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("12345678903"))
	rec := httptest.NewRecorder()
	h.UploadOrder(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestListOrders_Success(t *testing.T) {
	accrual := 500.0
	uploaded := time.Date(2020, 12, 10, 15, 15, 45, 0, time.FixedZone("MSK", 3*3600))
	h := newOrderHandler(&mockOrders{orders: []model.Order{
		{
			Number:     "9278923470",
			Status:     model.OrderStatusProcessed,
			Accrual:    &accrual,
			UploadedAt: uploaded,
		},
		{
			Number:     "12345678903",
			Status:     model.OrderStatusProcessing,
			UploadedAt: time.Date(2020, 12, 10, 15, 12, 1, 0, time.FixedZone("MSK", 3*3600)),
		},
	}}, nil)

	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodGet, "/api/user/orders", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var got []orderResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Number != "9278923470" || got[0].Status != model.OrderStatusProcessed {
		t.Fatalf("first order = %+v", got[0])
	}
	if got[0].Accrual == nil || *got[0].Accrual != 500 {
		t.Fatalf("accrual = %v, want 500", got[0].Accrual)
	}
	if got[0].UploadedAt != "2020-12-10T15:15:45+03:00" {
		t.Fatalf("uploaded_at = %q, want RFC3339", got[0].UploadedAt)
	}
	if got[1].Accrual != nil {
		t.Fatalf("processing order accrual = %v, want omitted", got[1].Accrual)
	}
}

func TestListOrders_NoContent(t *testing.T) {
	h := newOrderHandler(&mockOrders{orders: nil}, nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodGet, "/api/user/orders", ""))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestListOrders_Unauthorized(t *testing.T) {
	h := newOrderHandler(nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestListOrders_UnauthorizedCredentials(t *testing.T) {
	h := newOrderHandler(&mockOrders{err: service.ErrInvalidCredentials}, nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodGet, "/api/user/orders", ""))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestListOrders_InternalError(t *testing.T) {
	h := newOrderHandler(&mockOrders{err: errors.New("db down")}, nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodGet, "/api/user/orders", ""))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestListOrders_MissingUserInContext(t *testing.T) {
	h := newOrderHandler(nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	rec := httptest.NewRecorder()
	h.ListOrders(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestToOrderResponse_OmitsEmptyAccrual(t *testing.T) {
	got := toOrderResponse(model.Order{
		Number:     "123",
		Status:     model.OrderStatusInvalid,
		UploadedAt: time.Date(2020, 12, 9, 16, 9, 53, 0, time.FixedZone("MSK", 3*3600)),
	})
	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if strings.Contains(string(raw), "accrual") {
		t.Fatalf("json = %s, accrual should be omitted", raw)
	}
}

func TestExtractToken(t *testing.T) {
	t.Run("cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.AddCookie(&http.Cookie{Name: AuthCookieName, Value: "from-cookie"})
		req.Header.Set("Authorization", "Bearer from-header")
		if got := extractToken(req); got != "from-cookie" {
			t.Fatalf("extractToken() = %q, want from-cookie", got)
		}
	})

	t.Run("bearer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Authorization", "Bearer from-header")
		if got := extractToken(req); got != "from-header" {
			t.Fatalf("extractToken() = %q, want from-header", got)
		}
	})

	t.Run("empty", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		if got := extractToken(req); got != "" {
			t.Fatalf("extractToken() = %q, want empty", got)
		}
	})
}
