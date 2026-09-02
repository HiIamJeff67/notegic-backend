package routers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routine-task-dependencies"

	contexts "github.com/HiIamJeff67/notegic-backend/runtimes/core/contexts"
	routineservices "github.com/HiIamJeff67/notegic-backend/runtimes/core/services/routines"
	endpoints "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/endpoints"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/middlewares"
)

type RoutineTaskDependencyRouterDependencies struct {
	Service          routineservices.RoutineTaskDependencyServiceInterface
	AuthMiddleware   gin.HandlerFunc
	APIKeyMiddleware gin.HandlerFunc
}

func configureRoutineTaskDependencyRoutes(
	router *gin.RouterGroup,
	deps RoutineTaskDependencyRouterDependencies,
) {
	authMiddleware := deps.AuthMiddleware
	apiKeyMiddleware := deps.APIKeyMiddleware
	endpoint := endpoints.NewRoutineTaskDependencyEndpoint(deps.Service)
	apiCompatibleAuthMiddleware := middlewares.EitherMiddleware(
		[]gin.HandlerFunc{authMiddleware},
		[]gin.HandlerFunc{apiKeyMiddleware},
		func(ctx *gin.Context) bool { return contexts.IsClientGateway(ctx.Request.Context()) },
	)[0]

	routes := router.Group("/routine-task-dependencies")
	{
		routes.POST(
			"/get-by-routine-id",
			middlewares.DelegationAuthenticatedMiddleware(capi.GetRoutineTaskDependenciesByRoutineIdOperation),
			apiCompatibleAuthMiddleware,
			endpoint.GetRoutineTaskDependenciesByRoutineId,
		)
		routes.POST(
			"/create-by-routine-id",
			middlewares.DelegationAuthenticatedMiddleware(capi.CreateRoutineTaskDependencyByRoutineIdOperation),
			apiCompatibleAuthMiddleware,
			endpoint.CreateRoutineTaskDependencyByRoutineId,
		)
		routes.POST(
			"/create-many-by-routine-id",
			middlewares.DelegationAuthenticatedMiddleware(capi.CreateRoutineTaskDependenciesByRoutineIdOperation),
			apiCompatibleAuthMiddleware,
			endpoint.CreateRoutineTaskDependenciesByRoutineId,
		)
		routes.POST(
			"/update-by-routine-id",
			middlewares.DelegationAuthenticatedMiddleware(capi.UpdateRoutineTaskDependencyByRoutineIdOperation),
			apiCompatibleAuthMiddleware,
			endpoint.UpdateRoutineTaskDependencyByRoutineId,
		)
		routes.POST(
			"/update-many-by-routine-id",
			middlewares.DelegationAuthenticatedMiddleware(capi.UpdateRoutineTaskDependenciesByRoutineIdOperation),
			apiCompatibleAuthMiddleware,
			endpoint.UpdateRoutineTaskDependenciesByRoutineId,
		)
		routes.POST(
			"/delete-by-routine-id",
			middlewares.DelegationAuthenticatedMiddleware(capi.DeleteRoutineTaskDependencyByRoutineIdOperation),
			apiCompatibleAuthMiddleware,
			endpoint.DeleteRoutineTaskDependencyByRoutineId,
		)
		routes.POST(
			"/delete-many-by-routine-id",
			middlewares.DelegationAuthenticatedMiddleware(capi.DeleteRoutineTaskDependenciesByRoutineIdOperation),
			apiCompatibleAuthMiddleware,
			endpoint.DeleteRoutineTaskDependenciesByRoutineId,
		)
	}
}
