package server

import (
	"context"
	"testing"
)

func TestServerStopWithoutHTTPServer(t *testing.T) {
	s := &Server{}
	if err := s.Stop(context.Background()); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
}
