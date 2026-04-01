package util

import (
	"bytes"
	"testing"
)

func TestValidatePublicHTTPURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rawURL  string
		wantErr bool
	}{
		{
			name:    "allow_public_https_domain",
			rawURL:  "https://example.com",
			wantErr: false,
		},
		{
			name:    "allow_public_http_ip",
			rawURL:  "http://93.184.216.34",
			wantErr: false,
		},
		{
			name:    "reject_non_http_scheme",
			rawURL:  "ftp://example.com",
			wantErr: true,
		},
		{
			name:    "reject_loopback_ipv4",
			rawURL:  "http://127.0.0.1",
			wantErr: true,
		},
		{
			name:    "reject_loopback_ipv6",
			rawURL:  "http://[::1]",
			wantErr: true,
		},
		{
			name:    "reject_localhost",
			rawURL:  "http://localhost",
			wantErr: true,
		},
		{
			name:    "reject_private_ipv4",
			rawURL:  "http://10.0.0.1",
			wantErr: true,
		},
		{
			name:    "reject_link_local_metadata",
			rawURL:  "http://169.254.169.254",
			wantErr: true,
		},
		{
			name:    "reject_docker_host_alias",
			rawURL:  "http://host.docker.internal",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidatePublicHTTPURL(tt.rawURL)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidatePublicHTTPURL(%q) error = %v, wantErr=%v", tt.rawURL, err, tt.wantErr)
			}
		})
	}
}

func TestReadBodyWithLimit(t *testing.T) {
	t.Parallel()

	t.Run("allow_small_body", func(t *testing.T) {
		t.Parallel()
		body := bytes.Repeat([]byte("a"), int(defaultSafeResponseBodyLimitBytes-1))
		got, err := readBodyWithLimit(bytes.NewReader(body), defaultSafeResponseBodyLimitBytes)
		if err != nil {
			t.Fatalf("readBodyWithLimit unexpected error: %v", err)
		}
		if len(got) != len(body) {
			t.Fatalf("unexpected body len: got=%d want=%d", len(got), len(body))
		}
	})

	t.Run("reject_oversized_body", func(t *testing.T) {
		t.Parallel()
		body := bytes.Repeat([]byte("b"), int(defaultSafeResponseBodyLimitBytes+1))
		_, err := readBodyWithLimit(bytes.NewReader(body), defaultSafeResponseBodyLimitBytes)
		if err == nil {
			t.Fatal("expected oversize body error")
		}
	})
}
