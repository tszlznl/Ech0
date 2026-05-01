// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRateLimitBasic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimit(2, 2))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	for i := range 2 {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("request %d: status = %d, want %d", i, rec.Code, http.StatusOK)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("rate limited request: status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
}
