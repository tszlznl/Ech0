package viewer

import (
	"context"
	"testing"
)

func TestWithAndFromContext(t *testing.T) {
	ctx := context.Background()
	v := NewUserViewer("u1")
	ctx = WithContext(ctx, v)

	got, ok := FromContext(ctx)
	if !ok {
		t.Fatalf("expected viewer in context")
	}
	if got.UserID() != "u1" {
		t.Fatalf("unexpected user id: %s", got.UserID())
	}
}

func TestMustFromContextFallback(t *testing.T) {
	got := MustFromContext(context.Background())
	if got.IsAuthenticated() {
		t.Fatalf("expected unauthenticated fallback viewer")
	}
}

func TestIsAdminByRole(t *testing.T) {
	user := NewUserViewer("u1", WithRoles([]string{"admin"}))
	if !user.IsAdmin() {
		t.Fatalf("expected admin viewer")
	}

	guest := NewUserViewer("u2", WithRoles([]string{"user"}))
	if guest.IsAdmin() {
		t.Fatalf("expected non-admin viewer")
	}
}
