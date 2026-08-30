package middlewares

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// SanitizeXForwardedForMiddleware strips an optional port from the first X-Forwarded-For entry.
// Example: "118.163.65.239:56675, 10.224.0.10" -> "118.163.65.239"
// This helps Gin's ClientIP() work correctly behind proxies that append a source port.
func SanitizeXForwardedForMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		xForwardedFor := ctx.GetHeader("X-Forwarded-For")
		if xForwardedFor != "" {
			first := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
			sanitized := first

			// Handle bracketed IPv6 with port: "[2001:db8::1]:1234"
			if strings.HasPrefix(first, "[") {
				if host, _, err := net.SplitHostPort(first); err == nil {
					sanitized = host
				}
			} else if strings.Count(first, ":") == 1 {
				// Likely IPv4:port
				if host, _, err := net.SplitHostPort(first); err == nil {
					sanitized = host
				}
			}

			// Overwrite with sanitized value so gin.ClientIP() can parse it.
			ctx.Request.Header.Set("X-Forwarded-For", sanitized)

			// Optional convenience: also populate X-Real-IP if missing.
			if ctx.GetHeader("X-Real-IP") == "" {
				ctx.Request.Header.Set("X-Real-IP", sanitized)
			}
		}

		ctx.Next()
	}
}
