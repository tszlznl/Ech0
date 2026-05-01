// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package httpclient

import (
	"crypto/tls"
	"errors"
	"net/http"
	"time"

	httpUtil "github.com/lin-snow/ech0/internal/util/http"
)

// NewSafeHTTPClient builds an http.Client that blocks requests to
// private/reserved IP addresses (even when hidden behind public hostnames)
// and validates redirect targets against the same policy.
func NewSafeHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
			DialContext:         httpUtil.SecureDialContext(timeout),
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     30 * time.Second,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= httpUtil.MaxSafeRedirects {
				return errors.New("too many redirects")
			}
			return httpUtil.ValidatePublicHTTPURL(req.URL.String())
		},
	}
}
