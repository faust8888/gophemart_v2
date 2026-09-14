package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/faust8888/gophemart_v2/internal/model"
	"github.com/faust8888/gophemart_v2/internal/service"
)

func TestRegister_Success(t *testing.T) {
	h := newHandler(&mockUsers{
		user:  &model.User{ID: "1", Login: "alice"},
		token: "jwt-token",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader(`{"login":"alice","password":"secret"}`))
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusOK)
	}
	if got := res.Header.Get("Authorization"); got != "Bearer jwt-token" {
		t.Errorf("Authorization = %q, want Bearer jwt-token", got)
	}

	var found bool
	for _, c := range res.Cookies() {
		if c.Name == AuthCookieName && c.Value == "jwt-token" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("auth cookie was not set")
	}
}

func TestRegister_BadRequest(t *testing.T) {
	h := newHandler(&mockUsers{})

	cases := []struct {
		name string
		body string
	}{
		{name: "invalid json", body: "{"},
		{name: "empty body", body: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader(tc.body))
			rec := httptest.NewRecorder()
			h.Routes().ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestRegister_InvalidInputFromService(t *testing.T) {
	h := newHandler(&mockUsers{err: service.ErrInvalidInput})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader(`{"login":"","password":""}`))
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestRegister_Conflict(t *testing.T) {
	h := newHandler(&mockUsers{err: model.ErrLoginTaken})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader(`{"login":"alice","password":"secret"}`))
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestRegister_InternalError(t *testing.T) {
	h := newHandler(&mockUsers{err: errors.New("db down")})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString(`{"login":"alice","password":"secret"}`))
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestRegister_MethodNotAllowed(t *testing.T) {
	h := newHandler(&mockUsers{})
	req := httptest.NewRequest(http.MethodGet, "/api/user/register", nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestRecoverMiddleware(t *testing.T) {
	h := recoverMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
