package developmentroutes

import (
	"time"

	"github.com/gin-gonic/gin"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	scookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"

	binders "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/binders"
	controllers "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/controllers"
	interceptors "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/interceptors"
	middlewares "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/middlewares"
	coreadapters "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/core/adapters"
)

type BlockRouteDependencies struct {
	CoreAdapter               *coreadapters.CoreAdapter
	AccessTokenCookieHandler  *scookies.CookieHandler
	RefreshTokenCookieHandler *scookies.CookieHandler
	RateLimiters              RateLimiters
}

func configureDevelopmentBlockRoutes(
	router *gin.RouterGroup,
	deps BlockRouteDependencies,
) {
	coreAdapter, accessTokenCookieHandler, refreshTokenCookieHandler, rateLimiters := deps.CoreAdapter, deps.AccessTokenCookieHandler, deps.RefreshTokenCookieHandler, deps.RateLimiters
	if router == nil {
		router = DevelopmentAPIRouterGroup
	}

	blockBinder := binders.NewBlockBinder()
	blockController := controllers.NewBlockController(coreAdapter)
	blockRoutes := router.Group("/blocks")
	defaultMiddlewares := []gin.HandlerFunc{
		middlewares.UnauthorizedRateLimitMiddleware(rateLimiters.Unauthorized),
		middlewares.TimeoutMiddleware(3 * time.Second),
		middlewares.GatewayAuthenticationMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler),
		interceptors.ShareableResponseWriterInterceptor(
			interceptors.RefreshTokenInterceptor(accessTokenCookieHandler),
			interceptors.EmbeddedInterceptor,
		),
	}
	{
		blockRoutes.GET(
			"/:block-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getMyBlockById"),
					middlewares.ApplyMeterMiddleware("server.requests.block.getMyBlockById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				blockBinder.BindGetMyBlockById(blockController.GetMyBlockById),
			)...,
		)
		blockRoutes.GET(
			"/batch",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getMyBlocksByIds"),
					middlewares.ApplyMeterMiddleware("server.requests.block.getMyBlocksByIds"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				blockBinder.BindGetMyBlocksByIds(blockController.GetMyBlocksByIds),
			)...,
		)
		blockRoutes.GET(
			"/block-pack/:block-pack-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getMyBlocksByBlockPackId"),
					middlewares.ApplyMeterMiddleware("server.requests.block.getMyBlocksByBlockPackId"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				blockBinder.BindGetMyBlocksByBlockPackId(blockController.GetMyBlocksByBlockPackId),
			)...,
		)
	}
}
