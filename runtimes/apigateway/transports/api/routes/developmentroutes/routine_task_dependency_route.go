package developmentroutes

import (
	"time"

	"github.com/gin-gonic/gin"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	binders "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/api/binders"
	controllers "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/api/controllers"
	interceptors "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/api/interceptors"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/api/middlewares"
	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/core/adapters"
)

type RoutineTaskDependencyRouteDependencies struct {
	CoreAdapter  *coreadapters.CoreAdapter
	RateLimiters RateLimiters
}

func configureDevelopmentRoutineTaskDependencyRoutes(router *gin.RouterGroup, deps RoutineTaskDependencyRouteDependencies) {
	if router == nil {
		router = DevelopmentAPIRouterGroup
	}
	routineTaskDependencyBinder := binders.NewRoutineTaskDependencyBinder()
	controller := controllers.NewRoutineTaskDependencyController(deps.CoreAdapter)
	routes := router.Group("/routine-task-dependencies")
	defaultMiddlewares := []gin.HandlerFunc{
		middlewares.UnauthorizedRateLimitMiddleware(deps.RateLimiters.Unauthorized),
		middlewares.TimeoutMiddleware(3 * time.Second),
		interceptors.ShareableResponseWriterInterceptor(interceptors.EmbeddedInterceptor),
	}
	withPermission := func(permission cenums.AccessControlPermission, handler gin.HandlerFunc) []gin.HandlerFunc {
		middles := append([]gin.HandlerFunc{}, defaultMiddlewares...)
		middles = append(middles, middlewares.AllowedPermissionsAbove(permission))
		return middlewares.Reposition(nil, middles, handler)
	}
	routes.GET(
		"/routine/:routine-id",
		withPermission(
			cenums.AccessControlPermission_Read,
			routineTaskDependencyBinder.BindGetRoutineTaskDependenciesByRoutineId(
				controller.GetRoutineTaskDependenciesByRoutineId,
			),
		)...,
	)
	routes.POST(
		"/routine/:routine-id",
		withPermission(
			cenums.AccessControlPermission_Write,
			routineTaskDependencyBinder.BindCreateRoutineTaskDependencyByRoutineId(
				controller.CreateRoutineTaskDependencyByRoutineId,
			),
		)...,
	)
	routes.POST(
		"/routine/:routine-id/batch",
		withPermission(
			cenums.AccessControlPermission_Write,
			routineTaskDependencyBinder.BindCreateRoutineTaskDependenciesByRoutineId(
				controller.CreateRoutineTaskDependenciesByRoutineId,
			),
		)...,
	)
	routes.PUT(
		"/routine/:routine-id",
		withPermission(
			cenums.AccessControlPermission_Write,
			routineTaskDependencyBinder.BindUpdateRoutineTaskDependencyByRoutineId(
				controller.UpdateRoutineTaskDependencyByRoutineId,
			),
		)...,
	)
	routes.PUT(
		"/routine/:routine-id/batch",
		withPermission(
			cenums.AccessControlPermission_Write,
			routineTaskDependencyBinder.BindUpdateRoutineTaskDependenciesByRoutineId(
				controller.UpdateRoutineTaskDependenciesByRoutineId,
			),
		)...,
	)
	routes.DELETE(
		"/routine/:routine-id",
		withPermission(
			cenums.AccessControlPermission_Write,
			routineTaskDependencyBinder.BindDeleteRoutineTaskDependencyByRoutineId(
				controller.DeleteRoutineTaskDependencyByRoutineId,
			),
		)...,
	)
	routes.DELETE(
		"/routine/:routine-id/batch",
		withPermission(
			cenums.AccessControlPermission_Write,
			routineTaskDependencyBinder.BindDeleteRoutineTaskDependenciesByRoutineId(
				controller.DeleteRoutineTaskDependenciesByRoutineId,
			),
		)...,
	)
}
