package routers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routines"

	contexts "github.com/HiIamJeff67/notegic-backend/runtimes/core/contexts"
	routineservices "github.com/HiIamJeff67/notegic-backend/runtimes/core/services/routines"
	endpoints "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/endpoints"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/middlewares"
)

type RoutineRouterDependencies struct {
	Service          routineservices.RoutineServiceInterface
	AuthMiddleware   gin.HandlerFunc
	APIKeyMiddleware gin.HandlerFunc
}

func configureRoutineRoutes(
	router *gin.RouterGroup,
	deps RoutineRouterDependencies,
) {
	authMiddleware := deps.AuthMiddleware
	apiKeyMiddleware := deps.APIKeyMiddleware
	endpoint := endpoints.NewRoutineEndpoint(deps.Service)
	apiCompatibleAuthMiddleware := middlewares.EitherMiddleware(
		[]gin.HandlerFunc{authMiddleware},
		[]gin.HandlerFunc{apiKeyMiddleware},
		func(ctx *gin.Context) bool { return contexts.IsClientGateway(ctx.Request.Context()) },
	)[0]

	routineRoutes := router.Group("/routines")
	{
		routineRoutes.POST(
			"/get-by-id",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMyRoutineByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetMyRoutineById,
		)
		routineRoutes.POST(
			"/get-by-station-id",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMyRoutinesByStationIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetMyRoutinesByStationId,
		)
		routineRoutes.POST(
			"/get-all-by-time-range",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetAllMyRoutinesByTimeRangeOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetAllMyRoutinesByTimeRange,
		)
		routineRoutes.POST(
			"/create-by-station-id",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.CreateRoutineByStationIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.CreateRoutineByStationId,
		)
		routineRoutes.POST(
			"/create-many-by-station-ids",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.CreateRoutinesByStationIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.CreateRoutinesByStationIds,
		)
		routineRoutes.POST(
			"/update",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.UpdateMyRoutineByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.UpdateMyRoutineById,
		)
		routineRoutes.POST(
			"/update-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.UpdateMyRoutinesByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.UpdateMyRoutinesByIds,
		)
		routineRoutes.POST(
			"/link-tag",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.LinkRoutineTagByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.LinkRoutineTagById,
		)
		routineRoutes.POST(
			"/link-tags",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.LinkRoutineTagsByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.LinkRoutineTagsByIds,
		)
		routineRoutes.POST(
			"/link-item",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.LinkRoutineItemByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.LinkRoutineItemById,
		)
		routineRoutes.POST(
			"/link-items",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.LinkRoutineItemsByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.LinkRoutineItemsByIds,
		)
		routineRoutes.POST(
			"/restore",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.RestoreMyRoutineByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.RestoreMyRoutineById,
		)
		routineRoutes.POST(
			"/restore-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.RestoreMyRoutinesByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.RestoreMyRoutinesByIds,
		)
		routineRoutes.POST(
			"/delete",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.DeleteMyRoutineByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.DeleteMyRoutineById,
		)
		routineRoutes.POST(
			"/delete-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.DeleteMyRoutinesByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.DeleteMyRoutinesByIds,
		)
		routineRoutes.POST(
			"/hard-delete",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.HardDeleteMyRoutineByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.HardDeleteMyRoutineById,
		)
		routineRoutes.POST(
			"/hard-delete-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.HardDeleteMyRoutinesByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.HardDeleteMyRoutinesByIds,
		)
	}
	visualizationRoutes := router.Group("/routines/visualizations")
	{
		visualizationRoutes.POST(
			"/status-count",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.VisualizeMyRoutineStatusCountOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.VisualizeMyRoutineStatusCount,
		)
		visualizationRoutes.POST(
			"/period-count",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.VisualizeMyRoutinePeriodCountOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.VisualizeMyRoutinePeriodCount,
		)
		visualizationRoutes.POST(
			"/scheduled-start-at-count",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.VisualizeMyRoutineScheduledStartAtCountOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.VisualizeMyRoutineScheduledStartAtCount,
		)
		visualizationRoutes.POST(
			"/scheduled-end-at-count",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.VisualizeMyRoutineScheduledEndAtCountOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.VisualizeMyRoutineScheduledEndAtCount,
		)
	}
	router.POST(
		"/routines/graphql/search",
		middlewares.DelegationAuthenticatedMiddleware(
			capi.SearchRoutinesOperation,
		),
		apiCompatibleAuthMiddleware,
		endpoint.SearchRoutines,
	)
}
