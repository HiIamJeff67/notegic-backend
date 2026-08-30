package routers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routine-tasks"

	contexts "github.com/HiIamJeff67/notegic-backend/runtimes/core/contexts"
	routineservices "github.com/HiIamJeff67/notegic-backend/runtimes/core/services/routines"
	endpoints "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/endpoints"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/middlewares"
)

type RoutineTaskRouterDependencies struct {
	Service          routineservices.RoutineTaskServiceInterface
	AuthMiddleware   gin.HandlerFunc
	APIKeyMiddleware gin.HandlerFunc
}

func configureRoutineTaskRoutes(
	router *gin.RouterGroup,
	deps RoutineTaskRouterDependencies,
) {
	authMiddleware := deps.AuthMiddleware
	apiKeyMiddleware := deps.APIKeyMiddleware
	endpoint := endpoints.NewRoutineTaskEndpoint(deps.Service)
	apiCompatibleAuthMiddleware := middlewares.EitherMiddleware(
		[]gin.HandlerFunc{authMiddleware},
		[]gin.HandlerFunc{apiKeyMiddleware},
		func(ctx *gin.Context) bool { return contexts.IsClientGateway(ctx.Request.Context()) },
	)[0]

	routineTaskRoutes := router.Group("/routine-tasks")
	{
		routineTaskRoutes.POST(
			"/get-by-id",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMyRoutineTaskByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetMyRoutineTaskById,
		)
		routineTaskRoutes.POST(
			"/get-all-by-routine-ids",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetAllMyRoutineTasksByRoutineIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetAllMyRoutineTasksByRoutineIds,
		)
		routineTaskRoutes.POST(
			"/get-all",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetAllMyRoutineTasksOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetAllMyRoutineTasks,
		)
		routineTaskRoutes.POST(
			"/create-by-routine-id",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.CreateRoutineTaskByRoutineIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.CreateRoutineTaskByRoutineId,
		)
		routineTaskRoutes.POST(
			"/update",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.UpdateMyRoutineTaskByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.UpdateMyRoutineTaskById,
		)
		routineTaskRoutes.POST(
			"/pause",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.PauseMyRoutineTaskByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.PauseMyRoutineTaskById,
		)
		routineTaskRoutes.POST(
			"/resume",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.ResumeMyRoutineTaskByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.ResumeMyRoutineTaskById,
		)
		routineTaskRoutes.POST(
			"/hard-delete",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.HardDeleteMyRoutineTaskByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.HardDeleteMyRoutineTaskById,
		)
		routineTaskRoutes.POST(
			"/hard-delete-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.HardDeleteMyRoutineTasksByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.HardDeleteMyRoutineTasksByIds,
		)
	}
	visualizationRoutes := router.Group("/routine-tasks/visualizations")
	{
		visualizationRoutes.POST(
			"/visualize-status-count",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.VisualizeMyRoutineTaskStatusCountOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.VisualizeMyRoutineTaskStatusCount,
		)
		visualizationRoutes.POST(
			"/visualize-purpose-count",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.VisualizeMyRoutineTaskPurposeCountOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.VisualizeMyRoutineTaskPurposeCount,
		)
		visualizationRoutes.POST(
			"/visualize-scheduled-at-count",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.VisualizeMyRoutineTaskScheduledAtCountOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.VisualizeMyRoutineTaskScheduledAtCount,
		)
		visualizationRoutes.POST(
			"/visualize-actual-started-at-count",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.VisualizeMyRoutineTaskActualStartedAtCountOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.VisualizeMyRoutineTaskActualStartedAtCount,
		)
		visualizationRoutes.POST(
			"/visualize-actual-ended-at-count",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.VisualizeMyRoutineTaskActualEndedAtCountOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.VisualizeMyRoutineTaskActualEndedAtCount,
		)
	}
	router.POST(
		"/routine-tasks/graphql/search",
		middlewares.DelegationAuthenticatedMiddleware(
			capi.SearchRoutineTasksOperation,
		),
		apiCompatibleAuthMiddleware,
		endpoint.SearchRoutineTasks,
	)
}
