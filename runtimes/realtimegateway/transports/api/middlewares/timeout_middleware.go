package middlewares

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		timeoutContext, cancel := context.WithTimeout(ctx.Request.Context(), timeout)
		defer cancel()
		ctx.Request = ctx.Request.WithContext(timeoutContext)
		ctx.Next()
	}
}
