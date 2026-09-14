package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/faust8888/gophemart_v2/internal/model"
	"github.com/faust8888/gophemart_v2/internal/service"
)

func TestGetBalance_Success(t *testing.T) {
	h := newBalanceHandler(&mockBalance{balance: &model.Balance{Current: 500.5, Withdrawn: 42}}, nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodGet, "/api/user/balance", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var got model.Balance
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if got.Current != 500.5 || got.Withdrawn != 42 {
		t.Fatalf("balance = %+v, want current=500.5 withdrawn=42", got)
	}
}

func TestGetBalance_Zero(t *testing.T) {
	h := newBalanceHandler(&mockBalance{balance: &model.Balance{}}, nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodGet, "/api/user/balance", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got model.Balance
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if got.Current != 0 || got.Withdrawn != 0 {
		t.Fatalf("balance = %+v, want zeros", got)
	}
}

func TestGetBalance_Unauthorized(t *testing.T) {
	h := newBalanceHandler(nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestGetBalance_UnauthorizedCredentials(t *testing.T) {
	h := newBalanceHandler(&mockBalance{err: service.ErrInvalidCredentials}, nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodGet, "/api/user/balance", ""))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestGetBalance_InternalError(t *testing.T) {
	h := newBalanceHandler(&mockBalance{err: errors.New("db down")}, nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodGet, "/api/user/balance", ""))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestGetBalance_MissingUserInContext(t *testing.T) {
	h := newBalanceHandler(nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	rec := httptest.NewRecorder()
	h.GetBalance(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestWithdraw_Success(t *testing.T) {
	h := newBalanceHandler(nil, nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodPost, "/api/user/balance/withdraw", `{"order":"12345678903","sum":751}`))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestWithdraw_BadRequest(t *testing.T) {
	h := newBalanceHandler(nil, nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodPost, "/api/user/balance/withdraw", `{`))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestWithdraw_InvalidInput(t *testing.T) {
	h := newBalanceHandler(&mockBalance{err: service.ErrInvalidInput}, nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodPost, "/api/user/balance/withdraw", `{"order":"12345678903","sum":0}`))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestWithdraw_Unauthorized(t *testing.T) {
	h := newBalanceHandler(nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", strings.NewReader(`{"order":"12345678903","sum":751}`))
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestWithdraw_PaymentRequired(t *testing.T) {
	h := newBalanceHandler(&mockBalance{err: model.ErrInsufficientFunds}, nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodPost, "/api/user/balance/withdraw", `{"order":"12345678903","sum":751}`))
	if rec.Code != http.StatusPaymentRequired {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusPaymentRequired)
	}
}

func TestWithdraw_UnprocessableEntity(t *testing.T) {
	h := newBalanceHandler(&mockBalance{err: service.ErrInvalidOrderNumber}, nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodPost, "/api/user/balance/withdraw", `{"order":"123","sum":751}`))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}
}

func TestWithdraw_UnauthorizedCredentials(t *testing.T) {
	h := newBalanceHandler(&mockBalance{err: service.ErrInvalidCredentials}, nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodPost, "/api/user/balance/withdraw", `{"order":"12345678903","sum":751}`))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestWithdraw_InternalError(t *testing.T) {
	h := newBalanceHandler(&mockBalance{err: errors.New("db down")}, nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodPost, "/api/user/balance/withdraw", `{"order":"12345678903","sum":751}`))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestWithdraw_MissingUserInContext(t *testing.T) {
	h := newBalanceHandler(nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", strings.NewReader(`{"order":"12345678903","sum":751}`))
	rec := httptest.NewRecorder()
	h.Withdraw(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestListWithdrawals_Success(t *testing.T) {
	processed := time.Date(2020, 12, 9, 16, 9, 57, 0, time.FixedZone("MSK", 3*3600))
	h := newBalanceHandler(&mockBalance{withdrawals: []model.Withdrawal{
		{OrderNumber: "2377225624", Amount: 500, ProcessedAt: processed},
	}}, nil)

	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodGet, "/api/user/withdrawals", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var got []withdrawalResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Order != "2377225624" || got[0].Sum != 500 {
		t.Fatalf("withdrawal = %+v", got[0])
	}
	if got[0].ProcessedAt != "2020-12-09T16:09:57+03:00" {
		t.Fatalf("processed_at = %q, want RFC3339", got[0].ProcessedAt)
	}
}

func TestListWithdrawals_NoContent(t *testing.T) {
	h := newBalanceHandler(&mockBalance{withdrawals: nil}, nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodGet, "/api/user/withdrawals", ""))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestListWithdrawals_Unauthorized(t *testing.T) {
	h := newBalanceHandler(nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestListWithdrawals_UnauthorizedCredentials(t *testing.T) {
	h := newBalanceHandler(&mockBalance{err: service.ErrInvalidCredentials}, nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodGet, "/api/user/withdrawals", ""))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestListWithdrawals_InternalError(t *testing.T) {
	h := newBalanceHandler(&mockBalance{err: errors.New("db down")}, nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, authorizedRequest(http.MethodGet, "/api/user/withdrawals", ""))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestListWithdrawals_MissingUserInContext(t *testing.T) {
	h := newBalanceHandler(nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	rec := httptest.NewRecorder()
	h.ListWithdrawals(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
