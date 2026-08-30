package status

import (
	"net/http"

	gin "github.com/gin-gonic/gin"
)

func ConfigureHealthRouter(router gin.IRouter, isHealthy func() bool) {
	router.GET("/healthz", func(ctx *gin.Context) {
		if !isHealthy() {
			ctx.Status(http.StatusServiceUnavailable)
			return
		}

		ctx.Status(http.StatusOK)
	})
}
