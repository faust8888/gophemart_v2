package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/faust8888/gophemart_v2/internal/auth"
	"github.com/faust8888/gophemart_v2/internal/model"
)

type mockUsers struct {
	user  *model.User
	token string
	err   error
}

func (m *mockUsers) Register(_ context.Context, _, _ string) (*model.User, string, error) {
	return m.user, m.token, m.err
}

func (m *mockUsers) Login(_ context.Context, _, _ string) (*model.User, string, error) {
	return m.user, m.token, m.err
}

type mockOrders struct {
	err    error
	orders []model.Order
}

func (m *mockOrders) Upload(_ context.Context, _, _ string) error {
	return m.err
}

func (m *mockOrders) List(_ context.Context, _ string) ([]model.Order, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.orders, nil
}

type mockBalance struct {
	balance *model.Balance
	err     error
}

func (m *mockBalance) Get(_ context.Context, _ string) (*model.Balance, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.balance == nil {
		return &model.Balance{}, nil
	}
	return m.balance, nil
}

func (m *mockBalance) Withdraw(_ context.Context, _, _ string, _ float64) error {
	return m.err
}

type mockAuth struct {
	userID string
	err    error
}

func (m *mockAuth) ParseToken(token string) (string, string, error) {
	if m.err != nil {
		return "", "", m.err
	}
	if token == "" {
		return "", "", auth.ErrInvalidToken
	}
	return m.userID, "alice", nil
}

func newHandler(users userService) *Handler {
	if users == nil {
		users = &mockUsers{}
	}
	return New(users, &mockOrders{}, &mockBalance{}, &mockAuth{userID: "user-1"})
}

func newOrderHandler(orders *mockOrders, tokens *mockAuth) *Handler {
	if orders == nil {
		orders = &mockOrders{}
	}
	if tokens == nil {
		tokens = &mockAuth{userID: "user-1"}
	}
	return New(&mockUsers{}, orders, &mockBalance{}, tokens)
}

func newBalanceHandler(balances *mockBalance, tokens *mockAuth) *Handler {
	if balances == nil {
		balances = &mockBalance{}
	}
	if tokens == nil {
		tokens = &mockAuth{userID: "user-1"}
	}
	return New(&mockUsers{}, &mockOrders{}, balances, tokens)
}

func authorizedRequest(method, path, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.AddCookie(&http.Cookie{Name: AuthCookieName, Value: "jwt-token"})
	return req
}
