package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
