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

type RoutineTaskRouteDependencies struct {
	CoreAdapter               *coreadapters.CoreAdapter
	AccessTokenCookieHandler  *scookies.CookieHandler
	RefreshTokenCookieHandler *scookies.CookieHandler
	RateLimiters              RateLimiters
}

func configureDevelopmentRoutineTaskRoutes(
	router *gin.RouterGroup,
	deps RoutineTaskRouteDependencies,
) {
	coreAdapter, accessTokenCookieHandler, refreshTokenCookieHandler, rateLimiters := deps.CoreAdapter, deps.AccessTokenCookieHandler, deps.RefreshTokenCookieHandler, deps.RateLimiters
	if router == nil {
		router = DevelopmentAPIRouterGroup
	}

	routineTaskBinder := binders.NewRoutineTaskBinder()
	routineTaskController := controllers.NewRoutineTaskController(coreAdapter)

	routineTaskRoutes := router.Group("/routine-tasks")
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
		routineTaskRoutes.GET(
			"/:routine-task-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getMyRoutineTaskById"),
					middlewares.ApplyMeterMiddleware("server.requests.routineTask.getMyRoutineTaskById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				routineTaskBinder.BindGetMyRoutineTaskById(routineTaskController.GetMyRoutineTaskById),
			)...,
		)
		routineTaskRoutes.GET(
			"/routines",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getAllMyRoutineTasksByRoutineIds"),
					middlewares.ApplyMeterMiddleware("server.requests.routineTask.getAllMyRoutineTasksByRoutineIds"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				routineTaskBinder.BindGetAllMyRoutineTasksByRoutineIds(routineTaskController.GetAllMyRoutineTasksByRoutineIds),
			)...,
		)
		routineTaskRoutes.GET(
			"/",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getAllMyRoutineTasks"),
					middlewares.ApplyMeterMiddleware("server.requests.routineTask.getAllMyRoutineTasks"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				routineTaskBinder.BindGetAllMyRoutineTasks(routineTaskController.GetAllMyRoutineTasks),
			)...,
		)
		routineTaskRoutes.POST(
			"/routine/:routine-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("createRoutineTaskByRoutineId"),
					middlewares.ApplyMeterMiddleware("server.requests.routineTask.createRoutineTaskByRoutineId"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				routineTaskBinder.BindCreateRoutineTaskByRoutineId(routineTaskController.CreateRoutineTaskByRoutineId),
			)...,
		)
		routineTaskRoutes.PUT(
			"/:routine-task-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("updateMyRoutineTaskById"),
					middlewares.ApplyMeterMiddleware("server.requests.routineTask.updateMyRoutineTaskById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				routineTaskBinder.BindUpdateMyRoutineTaskById(routineTaskController.UpdateMyRoutineTaskById),
			)...,
		)
		routineTaskRoutes.DELETE(
			"/:routine-task-id/permanently",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("hardDeleteMyRoutineTaskById"),
					middlewares.ApplyMeterMiddleware("server.requests.routineTask.hardDeleteMyRoutineTaskById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				routineTaskBinder.BindHardDeleteMyRoutineTaskById(routineTaskController.HardDeleteMyRoutineTaskById),
			)...,
		)
		routineTaskRoutes.DELETE(
			"/batch/permanently",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("hardDeleteMyRoutineTasksByIds"),
					middlewares.ApplyMeterMiddleware("server.requests.routineTask.hardDeleteMyRoutineTasksByIds"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				routineTaskBinder.BindHardDeleteMyRoutineTasksByIds(routineTaskController.HardDeleteMyRoutineTasksByIds),
			)...,
		)
	}

	/* ============================== Routes for Visualization ============================== */

	visualizationRoutes := router.Group("/routine-tasks/visualizations")
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
			"/purpose-count",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("visualizeMyRoutineTaskPurposeCount"),
					middlewares.ApplyMeterMiddleware("server.requests.routineTask.visualizeMyRoutineTaskPurposeCount"),
				},
				append(
					visualizationMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				routineTaskBinder.BindVisualizeMyRoutineTaskPurposeCount(routineTaskController.VisualizeMyRoutineTaskPurposeCount),
			)...,
		)
	}
}
