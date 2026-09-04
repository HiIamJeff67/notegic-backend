package routers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/sub-shelves"

	contexts "github.com/HiIamJeff67/notegic-backend/runtimes/core/contexts"
	shelfservices "github.com/HiIamJeff67/notegic-backend/runtimes/core/services/shelves"
	endpoints "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/endpoints"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/middlewares"
)

type SubShelfRouterDependencies struct {
	Service          shelfservices.SubShelfServiceInterface
	AuthMiddleware   gin.HandlerFunc
	APIKeyMiddleware gin.HandlerFunc
}

func configureSubShelfRoutes(
	router *gin.RouterGroup,
	deps SubShelfRouterDependencies,
) {
	authMiddleware := deps.AuthMiddleware
	apiKeyMiddleware := deps.APIKeyMiddleware
	endpoint := endpoints.NewSubShelfEndpoint(deps.Service)
	apiCompatibleAuthMiddleware := middlewares.EitherMiddleware(
		[]gin.HandlerFunc{authMiddleware},
		[]gin.HandlerFunc{apiKeyMiddleware},
		func(ctx *gin.Context) bool { return contexts.IsClientGateway(ctx.Request.Context()) },
	)[0]

	subShelfRoutes := router.Group("/sub-shelves")
	{
		subShelfRoutes.POST(
			"/get-by-id",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMySubShelfByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetMySubShelfById,
		)
		subShelfRoutes.POST(
			"/get-by-prev-sub-shelf-id",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMySubShelvesByPrevSubShelfIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetMySubShelvesByPrevSubShelfId,
		)
		subShelfRoutes.POST(
			"/get-by-root-shelf-id",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMySubShelvesByRootShelfIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetMySubShelvesByRootShelfId,
		)
		subShelfRoutes.POST(
			"/get-and-items-by-prev-sub-shelf-id",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMySubShelvesAndItemsByPrevSubShelfIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetMySubShelvesAndItemsByPrevSubShelfId,
		)
		subShelfRoutes.POST(
			"/create",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.CreateSubShelfByRootShelfIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.CreateSubShelfByRootShelfId,
		)
		subShelfRoutes.POST(
			"/create-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.CreateSubShelvesByRootShelfIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.CreateSubShelvesByRootShelfIds,
		)
		subShelfRoutes.POST(
			"/update",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.UpdateMySubShelfByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.UpdateMySubShelfById,
		)
		subShelfRoutes.POST(
			"/update-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.UpdateMySubShelvesByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.UpdateMySubShelvesByIds,
		)
		subShelfRoutes.POST(
			"/move",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.MoveMySubShelfByRootShelfIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.MoveMySubShelfByRootShelfId,
		)
		subShelfRoutes.POST(
			"/move-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.MoveMySubShelvesByRootShelfIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.MoveMySubShelvesByRootShelfId,
		)
		subShelfRoutes.POST(
			"/move-many-by-root-shelves",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.MoveMySubShelvesByRootShelfIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.MoveMySubShelvesByRootShelfIds,
		)
		subShelfRoutes.POST(
			"/restore",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.RestoreMySubShelfByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.RestoreMySubShelfById,
		)
		subShelfRoutes.POST(
			"/restore-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.RestoreMySubShelvesByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.RestoreMySubShelvesByIds,
		)
		subShelfRoutes.POST(
			"/delete",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.DeleteMySubShelfByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.DeleteMySubShelfById,
		)
		subShelfRoutes.POST(
			"/delete-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.DeleteMySubShelvesByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.DeleteMySubShelvesByIds,
		)
		subShelfRoutes.POST(
			"/graphql/search",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.SearchSubShelvesOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.SearchSubShelves,
		)
	}
}
