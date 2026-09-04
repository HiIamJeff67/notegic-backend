package developmentroutes

import (
	"time"

	"github.com/gin-gonic/gin"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	scookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"

	binders "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/binders"
	controllers "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/controllers"
	interceptors "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/interceptors"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/middlewares"
	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/core/adapters"
)

type RoutineTaskRecordRouteDependencies struct {
	CoreAdapter               *coreadapters.CoreAdapter
	AccessTokenCookieHandler  *scookies.CookieHandler
	RefreshTokenCookieHandler *scookies.CookieHandler
	RateLimiters              RateLimiters
}

func configureDevelopmentRoutineTaskRecordRoutes(
	router *gin.RouterGroup,
	deps RoutineTaskRecordRouteDependencies,
) {
	coreAdapter, accessTokenCookieHandler, refreshTokenCookieHandler, rateLimiters := deps.CoreAdapter, deps.AccessTokenCookieHandler, deps.RefreshTokenCookieHandler, deps.RateLimiters
	if router == nil {
		router = DevelopmentAPIRouterGroup
	}

	routineTaskRecordBinder := binders.NewRoutineTaskRecordBinder()
	routineTaskRecordController := controllers.NewRoutineTaskRecordController(coreAdapter)

	routineTaskRecordRouterGroup := router.Group("/routine-task-records")
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
		routineTaskRecordRouterGroup.GET(
			"/routine-task/:routine-task-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getMyRoutineTaskRecordsByRoutineTaskId"),
					middlewares.ApplyMeterMiddleware("server.requests.routineTaskRecord.getMyRoutineTaskRecordsByRoutineTaskId"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				routineTaskRecordBinder.BindGetMyRoutineTaskRecordsByRoutineTaskId(routineTaskRecordController.GetMyRoutineTaskRecordsByRoutineTaskId),
			)...,
		)
	}

	/* ============================== Routes for Visualization ============================== */

	visualizationRoutes := router.Group("/routine-task-records/visualizations")
	visualizationMiddlewares := []gin.HandlerFunc{
		middlewares.UnauthorizedRateLimitMiddleware(rateLimiters.Unauthorized),
		middlewares.TimeoutMiddleware(3 * time.Second),
		middlewares.GatewayAuthenticationMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler),
		interceptors.ShareableResponseWriterInterceptor(
			interceptors.RefreshTokenInterceptor(accessTokenCookieHandler),
			interceptors.EmbeddedInterceptor,
		),
	}
	{
		visualizationRoutes.GET(
			"/status-count",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("visualizeMyRoutineTaskRecordStatusCount"),
					middlewares.ApplyMeterMiddleware("server.requests.routineTaskRecord.visualizeMyRoutineTaskRecordStatusCount"),
				},
				append(
					visualizationMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				routineTaskRecordBinder.BindVisualizeMyRoutineTaskRecordStatusCount(routineTaskRecordController.VisualizeMyRoutineTaskRecordStatusCount),
			)...,
		)
		visualizationRoutes.GET(
			"/purpose-count",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("visualizeMyRoutineTaskRecordPurposeCount"),
					middlewares.ApplyMeterMiddleware("server.requests.routineTaskRecord.visualizeMyRoutineTaskRecordPurposeCount"),
				},
				append(
					visualizationMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				routineTaskRecordBinder.BindVisualizeMyRoutineTaskRecordPurposeCount(routineTaskRecordController.VisualizeMyRoutineTaskRecordPurposeCount),
			)...,
		)
		visualizationRoutes.GET(
			"/scheduled-at-count",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("visualizeMyRoutineTaskRecordScheduledAtCount"),
					middlewares.ApplyMeterMiddleware("server.requests.routineTaskRecord.visualizeMyRoutineTaskRecordScheduledAtCount"),
				},
				append(
					visualizationMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				routineTaskRecordBinder.BindVisualizeMyRoutineTaskRecordScheduledAtCount(routineTaskRecordController.VisualizeMyRoutineTaskRecordScheduledAtCount),
			)...,
		)
		visualizationRoutes.GET(
			"/actual-started-at-count",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("visualizeMyRoutineTaskRecordActualStartedAtCount"),
					middlewares.ApplyMeterMiddleware("server.requests.routineTaskRecord.visualizeMyRoutineTaskRecordActualStartedAtCount"),
				},
				append(
					visualizationMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				routineTaskRecordBinder.BindVisualizeMyRoutineTaskRecordActualStartedAtCount(routineTaskRecordController.VisualizeMyRoutineTaskRecordActualStartedAtCount),
			)...,
		)
		visualizationRoutes.GET(
			"/actual-ended-at-count",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("visualizeMyRoutineTaskRecordActualEndedAtCount"),
					middlewares.ApplyMeterMiddleware("server.requests.routineTaskRecord.visualizeMyRoutineTaskRecordActualEndedAtCount"),
				},
				append(
					visualizationMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				routineTaskRecordBinder.BindVisualizeMyRoutineTaskRecordActualEndedAtCount(routineTaskRecordController.VisualizeMyRoutineTaskRecordActualEndedAtCount),
			)...,
		)
	}
}
