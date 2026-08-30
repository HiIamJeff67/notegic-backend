package status

import (
	"net/http"

	gin "github.com/gin-gonic/gin"
)

func ConfigureStartedRouter(router gin.IRouter, isStarted func() bool) {
	router.GET("/startedz", func(ctx *gin.Context) {
		if !isStarted() {
			ctx.Status(http.StatusServiceUnavailable)
			return
		}

		ctx.Status(http.StatusOK)
	})
}
