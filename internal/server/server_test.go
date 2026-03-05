package server

import (
	"context"
	"testing"
)

func TestServerStopCallsCacheCleanupWithoutHTTPServer(t *testing.T) {
	called := false
	s := &Server{
		cacheCleanup: func() error {
			called = true
			return nil
		},
	}

	if err := s.Stop(context.Background()); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
	if !called {
		t.Fatalf("expected cache cleanup to be called")
	}
}
