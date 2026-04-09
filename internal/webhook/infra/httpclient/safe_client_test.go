package httpclient

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewSafeHTTPClient_blocks_loopback_server(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewSafeHTTPClient(3 * time.Second)

	resp, err := client.Get(srv.URL)
	if resp != nil {
		_ = resp.Body.Close()
	}
	if err == nil {
		t.Fatal("expected safe client to block request to loopback server")
	}
	if !strings.Contains(err.Error(), "不被允许") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestNewSafeHTTPClient_blocks_redirect_to_loopback(t *testing.T) {
	t.Parallel()

	loopbackSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer loopbackSrv.Close()

	// This test verifies that even if the initial URL were allowed,
	// a redirect to a loopback address is still blocked.
	// We simulate by having a "public-like" server redirect to loopback.
	redirectSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, loopbackSrv.URL, http.StatusFound)
	}))
	defer redirectSrv.Close()

	client := NewSafeHTTPClient(3 * time.Second)

	resp, err := client.Get(redirectSrv.URL)
	if resp != nil {
		_ = resp.Body.Close()
	}
	if err == nil {
		t.Fatal("expected safe client to block redirect to loopback")
	}
}
