package routers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/blocks"

	contexts "github.com/HiIamJeff67/notegic-backend/runtimes/core/contexts"
	blockservices "github.com/HiIamJeff67/notegic-backend/runtimes/core/services/blocks"
	endpoints "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/endpoints"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/middlewares"
)

type BlockRouterDependencies struct {
	Service          blockservices.BlockServiceInterface
	AuthMiddleware   gin.HandlerFunc
	APIKeyMiddleware gin.HandlerFunc
}

func configureBlockRoutes(
	router *gin.RouterGroup,
	deps BlockRouterDependencies,
) {
	authMiddleware := deps.AuthMiddleware
	apiKeyMiddleware := deps.APIKeyMiddleware
	endpoint := endpoints.NewBlockEndpoint(deps.Service)
	apiCompatibleAuthMiddleware := middlewares.EitherMiddleware(
		[]gin.HandlerFunc{authMiddleware},
		[]gin.HandlerFunc{apiKeyMiddleware},
		func(ctx *gin.Context) bool { return contexts.IsClientGateway(ctx.Request.Context()) },
	)[0]

	blockRoutes := router.Group("/blocks")
	{
		blockRoutes.POST(
			"/get-by-id",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMyBlockByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetMyBlockById,
		)
		blockRoutes.POST(
			"/get-by-ids",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMyBlocksByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetMyBlocksByIds,
		)
		blockRoutes.POST(
			"/get-by-block-pack-id",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMyBlocksByBlockPackIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetMyBlocksByBlockPackId,
		)
		blockRoutes.POST(
			"/graphql/search",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.SearchBlocksOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.SearchBlocks,
		)
	}
}
