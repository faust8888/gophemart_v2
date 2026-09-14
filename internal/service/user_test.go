package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/faust8888/gophemart_v2/internal/auth"
	"github.com/faust8888/gophemart_v2/internal/model"
	"golang.org/x/crypto/bcrypt"
)

type fakeRepo struct {
	users map[string]*model.User
	err   error
}

func (f *fakeRepo) Create(_ context.Context, login, passwordHash string) (*model.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	if _, exists := f.users[login]; exists {
		return nil, model.ErrLoginTaken
	}
	user := &model.User{
		ID:           "user-1",
		Login:        login,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}
	f.users[login] = user
	return user, nil
}

func (f *fakeRepo) GetByLogin(_ context.Context, login string) (*model.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	user, ok := f.users[login]
	if !ok {
		return nil, model.ErrNotFound
	}
	return user, nil
}

func newTestService(t *testing.T, repo UserRepository) *UserService {
	t.Helper()
	tokens, err := auth.NewManager("test-secret", time.Hour)
	if err != nil {
		t.Fatalf("auth.NewManager() error = %v", err)
	}
	svc := NewUserService(repo, tokens)
	svc.hashCost = bcrypt.MinCost
	return svc
}

func TestRegister_Success(t *testing.T) {
	repo := &fakeRepo{users: make(map[string]*model.User)}
	svc := newTestService(t, repo)

	user, token, err := svc.Register(context.Background(), " alice ", "secret")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if user == nil || user.Login != "alice" {
		t.Fatalf("Register() user = %+v, want login alice", user)
	}
	if token == "" {
		t.Fatal("Register() token is empty")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("secret")); err != nil {
		t.Fatalf("password hash does not match: %v", err)
	}
}

func TestRegister_InvalidInput(t *testing.T) {
	repo := &fakeRepo{users: make(map[string]*model.User)}
	svc := newTestService(t, repo)

	cases := []struct {
		name     string
		login    string
		password string
	}{
		{name: "empty login", login: "", password: "secret"},
		{name: "whitespace login", login: "   ", password: "secret"},
		{name: "empty password", login: "alice", password: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := svc.Register(context.Background(), tc.login, tc.password)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("Register() error = %v, want %v", err, ErrInvalidInput)
			}
		})
	}
}

func TestRegister_LoginTaken(t *testing.T) {
	repo := &fakeRepo{users: map[string]*model.User{
		"alice": {ID: "1", Login: "alice"},
	}}
	svc := newTestService(t, repo)

	_, _, err := svc.Register(context.Background(), "alice", "secret")
	if !errors.Is(err, model.ErrLoginTaken) {
		t.Fatalf("Register() error = %v, want %v", err, model.ErrLoginTaken)
	}
}

func TestRegister_HashError(t *testing.T) {
	repo := &fakeRepo{users: make(map[string]*model.User)}
	svc := newTestService(t, repo)
	svc.hashCost = 100

	_, _, err := svc.Register(context.Background(), "alice", "secret")
	if err == nil {
		t.Fatal("Register() error = nil, want hashing error")
	}
}

type failIssuer struct{}

func (failIssuer) IssueToken(string, string) (string, error) {
	return "", errors.New("sign failed")
}

func TestRegister_TokenError(t *testing.T) {
	repo := &fakeRepo{users: make(map[string]*model.User)}
	svc := NewUserService(repo, failIssuer{})
	svc.hashCost = bcrypt.MinCost

	_, _, err := svc.Register(context.Background(), "alice", "secret")
	if err == nil {
		t.Fatal("Register() error = nil, want token error")
	}
}

func TestRegister_RepositoryError(t *testing.T) {
	want := errors.New("db down")
	repo := &fakeRepo{users: make(map[string]*model.User), err: want}
	svc := newTestService(t, repo)

	_, _, err := svc.Register(context.Background(), "alice", "secret")
	if !errors.Is(err, want) {
		t.Fatalf("Register() error = %v, want %v", err, want)
	}
}

func mustHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}
	return string(hash)
}

func TestLogin_Success(t *testing.T) {
	repo := &fakeRepo{users: map[string]*model.User{
		"alice": {ID: "user-1", Login: "alice", PasswordHash: mustHash(t, "secret")},
	}}
	svc := newTestService(t, repo)

	user, token, err := svc.Login(context.Background(), " alice ", "secret")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if user == nil || user.Login != "alice" {
		t.Fatalf("Login() user = %+v, want login alice", user)
	}
	if token == "" {
		t.Fatal("Login() token is empty")
	}
}

func TestLogin_InvalidInput(t *testing.T) {
	repo := &fakeRepo{users: make(map[string]*model.User)}
	svc := newTestService(t, repo)

	cases := []struct {
		name     string
		login    string
		password string
	}{
		{name: "empty login", login: "", password: "secret"},
		{name: "whitespace login", login: "   ", password: "secret"},
		{name: "empty password", login: "alice", password: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := svc.Login(context.Background(), tc.login, tc.password)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("Login() error = %v, want %v", err, ErrInvalidInput)
			}
		})
	}
}

func TestLogin_UnknownUser(t *testing.T) {
	repo := &fakeRepo{users: make(map[string]*model.User)}
	svc := newTestService(t, repo)

	_, _, err := svc.Login(context.Background(), "alice", "secret")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := &fakeRepo{users: map[string]*model.User{
		"alice": {ID: "user-1", Login: "alice", PasswordHash: mustHash(t, "secret")},
	}}
	svc := newTestService(t, repo)

	_, _, err := svc.Login(context.Background(), "alice", "wrong")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestLogin_RepositoryError(t *testing.T) {
	want := errors.New("db down")
	repo := &fakeRepo{users: make(map[string]*model.User), err: want}
	svc := newTestService(t, repo)

	_, _, err := svc.Login(context.Background(), "alice", "secret")
	if !errors.Is(err, want) {
		t.Fatalf("Login() error = %v, want %v", err, want)
	}
}

func TestLogin_TokenError(t *testing.T) {
	repo := &fakeRepo{users: map[string]*model.User{
		"alice": {ID: "user-1", Login: "alice", PasswordHash: mustHash(t, "secret")},
	}}
	svc := NewUserService(repo, failIssuer{})
	svc.hashCost = bcrypt.MinCost

	_, _, err := svc.Login(context.Background(), "alice", "secret")
	if err == nil {
		t.Fatal("Login() error = nil, want token error")
	}
}
