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

func TestLogin_Success(t *testing.T) {
	h := New(&mockUsers{
		user:  &model.User{ID: "1", Login: "alice"},
		token: "jwt-token",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(`{"login":"alice","password":"secret"}`))
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

func TestLogin_BadRequest(t *testing.T) {
	h := New(&mockUsers{})

	cases := []struct {
		name string
		body string
	}{
		{name: "invalid json", body: "{"},
		{name: "empty body", body: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(tc.body))
			rec := httptest.NewRecorder()
			h.Routes().ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestLogin_InvalidInputFromService(t *testing.T) {
	h := New(&mockUsers{err: service.ErrInvalidInput})
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(`{"login":"","password":""}`))
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestLogin_Unauthorized(t *testing.T) {
	h := New(&mockUsers{err: service.ErrInvalidCredentials})
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(`{"login":"alice","password":"wrong"}`))
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestLogin_InternalError(t *testing.T) {
	h := New(&mockUsers{err: errors.New("db down")})
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBufferString(`{"login":"alice","password":"secret"}`))
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestLogin_MethodNotAllowed(t *testing.T) {
	h := New(&mockUsers{})
	req := httptest.NewRequest(http.MethodGet, "/api/user/login", nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}
