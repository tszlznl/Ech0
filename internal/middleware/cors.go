package middleware

import (
	"net/http"
	"net/url"
	"net"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/config"
)

// Cors 跨域配置中间件
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		if origin != "" && isAllowedOrigin(origin) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		c.Header(
			"Access-Control-Allow-Headers",
			"Content-Type,AccessToken,X-CSRF-Token,Authorization,Token,x-token,X-Timezone",
		)
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PATCH, PUT")
		c.Header(
			"Access-Control-Expose-Headers",
			"Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type",
		)
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func isAllowedOrigin(origin string) bool {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return false
	}

	u, err := url.Parse(origin)
	if err != nil || u.Scheme == "" || u.Hostname() == "" {
		return false
	}
	if isLocalDevHost(u.Hostname()) {
		return true
	}

	allowedOrigins := config.Config().Web.CORS.AllowedOrigins
	if len(allowedOrigins) == 0 {
		// 未配置时退化为允许当前请求 Origin，保证可通过 Panel 首次完成配置。
		return true
	}
	for _, allowed := range allowedOrigins {
		allowedURL, err := url.Parse(strings.TrimSpace(allowed))
		if err != nil || allowedURL.Scheme == "" || allowedURL.Hostname() == "" {
			continue
		}
		if strings.EqualFold(u.Scheme, allowedURL.Scheme) && strings.EqualFold(u.Host, allowedURL.Host) {
			return true
		}
	}
	return false
}

func isLocalDevHost(hostname string) bool {
	h := strings.TrimSpace(strings.ToLower(hostname))
	if h == "localhost" || h == "::1" {
		return true
	}
	ip := net.ParseIP(h)
	return ip != nil && ip.IsLoopback()
}
