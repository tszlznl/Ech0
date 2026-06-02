// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package egress

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

// clientConfig holds the resolved options for NewClient.
type clientConfig struct {
	timeout time.Duration
	guard   bool
}

// Option configures a client built by NewClient.
type Option func(*clientConfig)

// Timeout sets the whole-request timeout (http.Client.Timeout). It covers the
// entire exchange including reading the response body, so it is appropriate for
// short request/response calls but not for long-lived streaming responses.
func Timeout(d time.Duration) Option {
	return func(c *clientConfig) { c.timeout = d }
}

// Guard enables SSRF protection: a dialer that rejects private/reserved
// destination IPs (defending against DNS rebinding) plus redirect validation.
// Use it only for user/peer-supplied URLs. Admin-configured infra endpoints
// (OIDC providers, LLM backends, captcha verifiers) may legitimately live on
// loopback or private networks, so they must NOT enable it.
func Guard() Option {
	return func(c *clientConfig) { c.guard = true }
}

// NewClient builds an *http.Client with a hardened transport (TLS >= 1.2, a
// bounded idle-connection pool, and outbound request logging). Apply Guard()
// to add SSRF protection. The returned *http.Client also satisfies the
// Do(*http.Request) interfaces expected by the OpenAI / Anthropic SDKs.
func NewClient(opts ...Option) *http.Client {
	cfg := clientConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	transport := &http.Transport{
		TLSClientConfig:     &tls.Config{MinVersion: tls.VersionTLS12},
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
	}

	client := &http.Client{
		Timeout:   cfg.timeout,
		Transport: &loggingRoundTripper{base: transport},
	}

	if cfg.guard {
		transport.DialContext = secureDialContext(cfg.timeout)
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxSafeRedirects {
				return errors.New("too many redirects")
			}
			return Validate(req.URL.String())
		}
	}

	return client
}

// loggingRoundTripper logs each outbound request (method/host/status/latency)
// at debug level. Errors are logged then returned unchanged for the caller to
// handle, so no error information is swallowed.
type loggingRoundTripper struct {
	base http.RoundTripper
}

func (l *loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := l.base.RoundTrip(req)
	latencyMs := time.Since(start).Milliseconds()
	if err != nil {
		logUtil.Debug(
			"egress request failed",
			zap.String("module", "egress"),
			zap.String("method", req.Method),
			zap.String("host", req.URL.Host),
			zap.Int64("latency_ms", latencyMs),
			zap.Error(err),
		)
		return resp, err
	}
	logUtil.Debug(
		"egress request",
		zap.String("module", "egress"),
		zap.String("method", req.Method),
		zap.String("host", req.URL.Host),
		zap.Int("status", resp.StatusCode),
		zap.Int64("latency_ms", latencyMs),
	)
	return resp, err
}

// Retry runs fn up to maxAttempts times, sleeping with exponential backoff
// between attempts (starting at initialBackoff). It returns nil on the first
// success and the last error otherwise. It does not sleep after the final
// attempt.
func Retry(maxAttempts int, initialBackoff time.Duration, fn func() error) error {
	var err error
	delay := initialBackoff
	for i := range maxAttempts {
		if err = fn(); err == nil {
			return nil
		}
		if i < maxAttempts-1 {
			time.Sleep(delay)
			delay *= 2
		}
	}
	return err
}

// Header is an optional single request header for Fetch.
type Header struct {
	Header  string
	Content string
}

// Fetch performs an SSRF-guarded request and returns the response body,
// size-limited to 1 MiB. The default timeout is 2s unless overridden.
func Fetch(url, method string, h Header, timeout ...time.Duration) ([]byte, error) {
	if err := Validate(url); err != nil {
		return nil, err
	}

	clientTimeout := 2 * time.Second
	if len(timeout) > 0 {
		clientTimeout = timeout[0]
	}

	client := NewClient(Guard(), Timeout(clientTimeout))

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	if h.Header != "" && h.Content != "" {
		req.Header.Set(h.Header, h.Content)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求发送失败: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logUtil.Warn(
				"close response body failed",
				zap.String("module", "egress"),
				zap.Error(closeErr),
			)
		}
	}()

	return readBodyWithLimit(resp.Body, defaultSafeResponseBodyLimitBytes)
}
