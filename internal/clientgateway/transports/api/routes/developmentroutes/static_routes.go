package developmentroutes

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"

	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"

	middlewares "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/middlewares"
)

type StaticRouteDependencies struct {
	RateLimiters RateLimiters
}

func configureStaticRoutes(router *gin.RouterGroup, deps StaticRouteDependencies) {
	rateLimiters := deps.RateLimiters
	if router == nil {
		router = DevelopmentAPIRouterGroup
	}

	staticGroup := router.Group("/static")
	{
		globalImagesGroup := staticGroup.Group("/global-images")
		globalImagesGroup.Use(
			middlewares.UnauthorizedRateLimitMiddleware(rateLimiters.Unauthorized),
		)
		{
			// configure avatars
			globalImagesGroup.GET("/avatars/:id", func(ctx *gin.Context) {
				ctx.Header("Cross-Origin-Resource-Policy", "cross-origin")
				avatarId := ctx.Param("id")
				filePath := fmt.Sprintf("./global/images/avatars/userAvatar%s.png", avatarId)

				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					filePath = "./global/images/avatars/userAvatar1.png"
				}
				slogs.NotegicLogger.Info(ctx.Request.Context(), "download file")

				ctx.File(filePath)
			})

			// configure brand icon here in the future
		}
	}
}
