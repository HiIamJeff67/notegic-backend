package developmentroutes

import (
	"time"

	"github.com/gin-gonic/gin"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	ratelimit "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/ratelimit"
	binders "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/api/binders"
	controllers "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/api/controllers"
	interceptors "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/api/interceptors"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/api/middlewares"
	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/core/adapters"
)

type BlockRouteDependencies struct {
	CoreAdapter             *coreadapters.CoreAdapter
	UnauthorizedRateLimiter *ratelimit.HybridRateLimiter
}

func configureDevelopmentBlockRoutes(
	router *gin.RouterGroup,
	deps BlockRouteDependencies,
) {
	coreAdapter, unauthorizedRateLimiter := deps.CoreAdapter, deps.UnauthorizedRateLimiter

	blockBinder := binders.NewBlockBinder()
	blockController := controllers.NewBlockController(coreAdapter)
	blockRoutes := router.Group("/blocks")
	defaultMiddlewares := []gin.HandlerFunc{
		middlewares.UnauthorizedRateLimitMiddleware(unauthorizedRateLimiter),
		middlewares.TimeoutMiddleware(3 * time.Second),
		interceptors.ShareableResponseWriterInterceptor(
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
