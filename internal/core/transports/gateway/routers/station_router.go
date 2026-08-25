package routers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/stations"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	routineservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/routines"
	endpoints "github.com/HiIamJeff67/notegic-backend/internal/core/transports/gateway/endpoints"
	middlewares "github.com/HiIamJeff67/notegic-backend/internal/core/transports/gateway/middlewares"
)

type StationRouterDependencies struct {
	Service          routineservices.StationServiceInterface
	AuthMiddleware   gin.HandlerFunc
	APIKeyMiddleware gin.HandlerFunc
}

func configureStationRoutes(
	router *gin.RouterGroup,
	deps StationRouterDependencies,
) {
	authMiddleware := deps.AuthMiddleware
	apiKeyMiddleware := deps.APIKeyMiddleware
	endpoint := endpoints.NewStationEndpoint(deps.Service)
	apiCompatibleAuthMiddleware := middlewares.EitherMiddleware(
		[]gin.HandlerFunc{authMiddleware},
		[]gin.HandlerFunc{apiKeyMiddleware},
		func(ctx *gin.Context) bool { return contexts.IsClientGateway(ctx.Request.Context()) },
	)[0]

	stationRoutes := router.Group("/stations")
	{
		stationRoutes.POST(
			"/get-by-id",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMyStationByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetMyStationById,
		)
		stationRoutes.POST(
			"/get-all",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetAllMyStationsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetAllMyStations,
		)
		stationRoutes.POST(
			"/create",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.CreateStationOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.CreateStation,
		)
		stationRoutes.POST(
			"/create-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.CreateStationsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.CreateStations,
		)
		stationRoutes.POST(
			"/update",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.UpdateMyStationByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.UpdateMyStationById,
		)
		stationRoutes.POST(
			"/update-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.UpdateMyStationsByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.UpdateMyStationsByIds,
		)
		stationRoutes.POST(
			"/restore",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.RestoreMyStationByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.RestoreMyStationById,
		)
		stationRoutes.POST(
			"/restore-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.RestoreMyStationsByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.RestoreMyStationsByIds,
		)
		stationRoutes.POST(
			"/delete",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.DeleteMyStationByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.DeleteMyStationById,
		)
		stationRoutes.POST(
			"/delete-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.DeleteMyStationsByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.DeleteMyStationsByIds,
		)
		stationRoutes.POST(
			"/hard-delete",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.HardDeleteMyStationByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.HardDeleteMyStationById,
		)
		stationRoutes.POST(
			"/hard-delete-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.HardDeleteMyStationsByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.HardDeleteMyStationsByIds,
		)
		stationRoutes.POST(
			"/visualizations/total-count",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.VisualizeMyTotalCountOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.VisualizeMyTotalCount,
		)
		stationRoutes.POST(
			"/permissions/get",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMyStationPermissionOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetMyStationPermission,
		)
		stationRoutes.POST(
			"/permissions/create",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.CreateMyStationPermissionOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.CreateMyStationPermission,
		)
		stationRoutes.POST(
			"/permissions/upsert",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.UpsertMyStationPermissionOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.UpsertMyStationPermission,
		)
		stationRoutes.POST(
			"/permissions/upsert-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.UpsertMyStationPermissionsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.UpsertMyStationPermissions,
		)
		stationRoutes.POST(
			"/permissions/update",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.UpdateMyStationPermissionOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.UpdateMyStationPermission,
		)
		stationRoutes.POST(
			"/ownership/transfer",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.TransferMyStationOwnershipOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.TransferMyStationOwnership,
		)
		stationRoutes.POST(
			"/permissions/delete",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.DeleteMyStationPermissionOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.DeleteMyStationPermission,
		)
		stationRoutes.POST(
			"/permissions/delete-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.DeleteMyStationPermissionsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.DeleteMyStationPermissions,
		)
		stationRoutes.POST(
			"/memberships/leave",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.LeaveMyStationOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.LeaveMyStation,
		)
		stationRoutes.POST(
			"/memberships/leave-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.LeaveMyStationsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.LeaveMyStations,
		)
		stationRoutes.POST(
			"/graphql/search",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.SearchStationsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.SearchStations,
		)
	}
}
