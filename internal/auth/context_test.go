package auth

import (
	"context"
	"testing"
)

func TestUserIDFromContext(t *testing.T) {
	if _, ok := UserIDFromContext(context.Background()); ok {
		t.Fatal("UserIDFromContext(empty) ok = true, want false")
	}

	ctx := ContextWithUserID(context.Background(), "user-1")
	got, ok := UserIDFromContext(ctx)
	if !ok {
		t.Fatal("UserIDFromContext() ok = false, want true")
	}
	if got != "user-1" {
		t.Fatalf("UserIDFromContext() = %q, want user-1", got)
	}

	if _, ok := UserIDFromContext(ContextWithUserID(context.Background(), "")); ok {
		t.Fatal("UserIDFromContext(empty id) ok = true, want false")
	}
}
