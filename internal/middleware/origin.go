// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func OriginGuard(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = struct{}{}
	}
	return func(c *gin.Context) {
		if len(allowed) == 0 {
			c.Next()
			return
		}
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}
		if _, ok := allowed[origin]; !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "origin not allowed"})
			c.Abort()
			return
		}
		c.Next()
	}
}
