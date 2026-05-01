// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// inlineableMIMEPrefixes lists Content-Type prefixes that are safe to display
// inline in the browser (images and audio only).
var inlineableMIMEPrefixes = []string{
	"image/",
	"audio/",
}

// StaticFileSecurity returns a middleware that hardens responses served from the
// public file endpoint:
//   - X-Content-Type-Options: nosniff (prevents MIME-sniffing attacks).
//   - Content-Disposition: attachment for non-image/audio files (forces download
//     instead of in-browser execution).
func StaticFileSecurity() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")

		c.Next()

		ct := c.Writer.Header().Get("Content-Type")
		if ct == "" {
			ext := strings.ToLower(filepath.Ext(c.Request.URL.Path))
			if ext == "" {
				ct = "application/octet-stream"
			}
		}

		if !isInlineableMIME(ct) {
			basename := filepath.Base(c.Request.URL.Path)
			c.Header("Content-Disposition", "attachment; filename=\""+basename+"\"")
		}
	}
}

func isInlineableMIME(ct string) bool {
	lower := strings.ToLower(ct)
	for _, prefix := range inlineableMIMEPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return false
}
