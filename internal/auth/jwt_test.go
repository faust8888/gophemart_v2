package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestNewManager_EmptySecret(t *testing.T) {
	_, err := NewManager("", time.Hour)
	if err == nil {
		t.Fatal("NewManager() error = nil, want error")
	}
}

func TestNewManager_DefaultTTL(t *testing.T) {
	m, err := NewManager("secret", 0)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	if m.ttl != 24*time.Hour {
		t.Errorf("ttl = %v, want 24h", m.ttl)
	}
}

func TestIssueAndParseToken(t *testing.T) {
	m, err := NewManager("secret", time.Hour)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	token, err := m.IssueToken("user-1", "alice")
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}
	if token == "" {
		t.Fatal("IssueToken() returned empty token")
	}

	userID, login, err := m.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if userID != "user-1" {
		t.Errorf("userID = %q, want user-1", userID)
	}
	if login != "alice" {
		t.Errorf("login = %q, want alice", login)
	}
}

func TestParseToken_Invalid(t *testing.T) {
	m, err := NewManager("secret", time.Hour)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	cases := []string{"", "not-a-token", "a.b.c"}
	for _, token := range cases {
		if _, _, err := m.ParseToken(token); err == nil {
			t.Errorf("ParseToken(%q) error = nil, want error", token)
		}
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	issuer, err := NewManager("issuer", time.Hour)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	parser, err := NewManager("other", time.Hour)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	token, err := issuer.IssueToken("user-1", "alice")
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}
	if _, _, err := parser.ParseToken(token); err == nil {
		t.Fatal("ParseToken() with wrong secret succeeded, want error")
	}
}

func TestParseToken_Expired(t *testing.T) {
	m, err := NewManager("secret", time.Hour)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	expired := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		UserID: "user-1",
		Login:  "alice",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	})
	token, err := expired.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	if _, _, err := m.ParseToken(token); err == nil {
		t.Fatal("ParseToken() for expired token succeeded, want error")
	}
}

func TestParseToken_EmptyUserID(t *testing.T) {
	m, err := NewManager("secret", time.Hour)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	unsigned := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		UserID: "",
		Login:  "alice",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	token, err := unsigned.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}
	if _, _, err := m.ParseToken(token); err == nil {
		t.Fatal("ParseToken() with empty user id succeeded, want error")
	}
}
