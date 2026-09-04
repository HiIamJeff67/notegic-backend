package routers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/block-packs"

	contexts "github.com/HiIamJeff67/notegic-backend/runtimes/core/contexts"
	blockservices "github.com/HiIamJeff67/notegic-backend/runtimes/core/services/blocks"
	endpoints "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/endpoints"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/middlewares"
)

type BlockPackRouterDependencies struct {
	Service          blockservices.BlockPackServiceInterface
	AuthMiddleware   gin.HandlerFunc
	APIKeyMiddleware gin.HandlerFunc
}

func configureBlockPackRoutes(
	router *gin.RouterGroup,
	deps BlockPackRouterDependencies,
) {
	authMiddleware := deps.AuthMiddleware
	apiKeyMiddleware := deps.APIKeyMiddleware
	endpoint := endpoints.NewBlockPackEndpoint(deps.Service)
	apiCompatibleAuthMiddleware := middlewares.EitherMiddleware(
		[]gin.HandlerFunc{authMiddleware},
		[]gin.HandlerFunc{apiKeyMiddleware},
		func(ctx *gin.Context) bool { return contexts.IsClientGateway(ctx.Request.Context()) },
	)[0]

	blockPackRoutes := router.Group("/block-packs")
	{
		blockPackRoutes.POST(
			"/get-by-id",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMyBlockPackByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetMyBlockPackById,
		)
		blockPackRoutes.POST(
			"/get-and-parent-by-id",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMyBlockPackAndItsParentByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetMyBlockPackAndItsParentById,
		)
		blockPackRoutes.POST(
			"/get-by-parent-sub-shelf-id",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMyBlockPacksByParentSubShelfIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetMyBlockPacksByParentSubShelfId,
		)
		blockPackRoutes.POST(
			"/get-by-root-shelf-id",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMyBlockPacksByRootShelfIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetMyBlockPacksByRootShelfId,
		)
		blockPackRoutes.POST(
			"/create",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.CreateBlockPackOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.CreateBlockPack,
		)
		blockPackRoutes.POST(
			"/create-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.CreateBlockPacksOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.CreateBlockPacks,
		)
		blockPackRoutes.POST(
			"/update",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.UpdateMyBlockPackByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.UpdateMyBlockPackById,
		)
		blockPackRoutes.POST(
			"/update-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.UpdateMyBlockPacksByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.UpdateMyBlockPacksByIds,
		)
		blockPackRoutes.POST(
			"/move",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.MoveMyBlockPackByParentSubShelfIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.MoveMyBlockPackByParentSubShelfId,
		)
		blockPackRoutes.POST(
			"/move-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.MoveMyBlockPacksByParentSubShelfIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.MoveMyBlockPacksByParentSubShelfId,
		)
		blockPackRoutes.POST(
			"/move-many-by-parent-sub-shelves",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.MoveMyBlockPacksByParentSubShelfIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.MoveMyBlockPacksByParentSubShelfIds,
		)
		blockPackRoutes.POST(
			"/restore",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.RestoreMyBlockPackByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.RestoreMyBlockPackById,
		)
		blockPackRoutes.POST(
			"/restore-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.RestoreMyBlockPacksByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.RestoreMyBlockPacksByIds,
		)
		blockPackRoutes.POST(
			"/delete",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.DeleteMyBlockPackByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.DeleteMyBlockPackById,
		)
		blockPackRoutes.POST(
			"/delete-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.DeleteMyBlockPacksByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.DeleteMyBlockPacksByIds,
		)
		blockPackRoutes.POST(
			"/graphql/search",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.SearchBlockPacksOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.SearchBlockPacks,
		)
	}
}
