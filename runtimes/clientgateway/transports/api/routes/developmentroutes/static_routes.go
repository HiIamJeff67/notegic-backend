package developmentroutes

import (
	"github.com/gin-gonic/gin"
)

func configureStaticRoutes(router *gin.RouterGroup) {
	staticGroup := router.Group("/static")
	{
	staticGroup.GET("/logo", func(ctx *gin.Context) {
		ctx.Header("Cross-Origin-Resource-Policy", "cross-origin")
		ctx.File("./runtimes/email/templates/assets/common.svg")
	})
	}
}
