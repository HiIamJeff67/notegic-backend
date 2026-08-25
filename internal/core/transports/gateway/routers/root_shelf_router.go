package routers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/root-shelves"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	shelfservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/shelves"
	endpoints "github.com/HiIamJeff67/notegic-backend/internal/core/transports/gateway/endpoints"
	middlewares "github.com/HiIamJeff67/notegic-backend/internal/core/transports/gateway/middlewares"
)

type RootShelfRouterDependencies struct {
	Service          shelfservices.RootShelfServiceInterface
	AuthMiddleware   gin.HandlerFunc
	APIKeyMiddleware gin.HandlerFunc
}

func configureRootShelfRoutes(
	router *gin.RouterGroup,
	deps RootShelfRouterDependencies,
) {
	authMiddleware := deps.AuthMiddleware
	apiKeyMiddleware := deps.APIKeyMiddleware
	endpoint := endpoints.NewRootShelfEndpoint(deps.Service)
	apiCompatibleAuthMiddleware := middlewares.EitherMiddleware(
		[]gin.HandlerFunc{authMiddleware},
		[]gin.HandlerFunc{apiKeyMiddleware},
		func(ctx *gin.Context) bool { return contexts.IsClientGateway(ctx.Request.Context()) },
	)[0]

	rootShelfRoutes := router.Group("/root-shelves")
	{
		rootShelfRoutes.POST(
			"/get-by-id",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMyRootShelfByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetMyRootShelfById,
		)
		rootShelfRoutes.POST(
			"/create",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.CreateRootShelfOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.CreateRootShelf,
		)
		rootShelfRoutes.POST(
			"/create-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.CreateRootShelvesOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.CreateRootShelves,
		)
		rootShelfRoutes.POST(
			"/update",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.UpdateMyRootShelfByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.UpdateMyRootShelfById,
		)
		rootShelfRoutes.POST(
			"/update-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.UpdateMyRootShelvesByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.UpdateMyRootShelvesByIds,
		)
		rootShelfRoutes.POST(
			"/restore",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.RestoreMyRootShelfByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.RestoreMyRootShelfById,
		)
		rootShelfRoutes.POST(
			"/restore-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.RestoreMyRootShelvesByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.RestoreMyRootShelvesByIds,
		)
		rootShelfRoutes.POST(
			"/delete",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.DeleteMyRootShelfByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.DeleteMyRootShelfById,
		)
		rootShelfRoutes.POST(
			"/delete-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.DeleteMyRootShelvesByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.DeleteMyRootShelvesByIds,
		)
		rootShelfRoutes.POST(
			"/permissions/get",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMyRootShelfPermissionOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetMyRootShelfPermission,
		)
		rootShelfRoutes.POST(
			"/permissions/create",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.CreateMyRootShelfPermissionOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.CreateMyRootShelfPermission,
		)
		rootShelfRoutes.POST(
			"/permissions/upsert",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.UpsertMyRootShelfPermissionOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.UpsertMyRootShelfPermission,
		)
		rootShelfRoutes.POST(
			"/permissions/upsert-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.UpsertMyRootShelfPermissionsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.UpsertMyRootShelfPermissions,
		)
		rootShelfRoutes.POST(
			"/permissions/update",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.UpdateMyRootShelfPermissionOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.UpdateMyRootShelfPermission,
		)
		rootShelfRoutes.POST(
			"/ownership/transfer",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.TransferMyRootShelfOwnershipOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.TransferMyRootShelfOwnership,
		)
		rootShelfRoutes.POST(
			"/permissions/delete",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.DeleteMyRootShelfPermissionOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.DeleteMyRootShelfPermission,
		)
		rootShelfRoutes.POST(
			"/permissions/delete-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.DeleteMyRootShelfPermissionsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.DeleteMyRootShelfPermissions,
		)
		rootShelfRoutes.POST(
			"/memberships/leave",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.LeaveMyRootShelfOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.LeaveMyRootShelf,
		)
		rootShelfRoutes.POST(
			"/memberships/leave-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.LeaveMyRootShelvesOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.LeaveMyRootShelves,
		)
		rootShelfRoutes.POST(
			"/graphql/search",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.SearchRootShelvesOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.SearchRootShelves,
		)
	}
}
