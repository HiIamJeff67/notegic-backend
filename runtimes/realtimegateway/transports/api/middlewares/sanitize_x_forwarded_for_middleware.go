package middlewares

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

func SanitizeXForwardedForMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		xForwardedFor := ctx.GetHeader("X-Forwarded-For")
		if xForwardedFor != "" {
			first := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
			sanitized := first
			if strings.HasPrefix(first, "[") {
				if host, _, err := net.SplitHostPort(first); err == nil {
					sanitized = host
				}
			} else if strings.Count(first, ":") == 1 {
				if host, _, err := net.SplitHostPort(first); err == nil {
					sanitized = host
				}
			}
			ctx.Request.Header.Set("X-Forwarded-For", sanitized)
			if ctx.GetHeader("X-Real-IP") == "" {
				ctx.Request.Header.Set("X-Real-IP", sanitized)
			}
		}

		ctx.Next()
	}
}
