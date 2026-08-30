package routers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/materials"

	contexts "github.com/HiIamJeff67/notegic-backend/runtimes/core/contexts"
	materialservices "github.com/HiIamJeff67/notegic-backend/runtimes/core/services/material"
	endpoints "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/endpoints"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/middlewares"
)

type MaterialRouterDependencies struct {
	Service          materialservices.MaterialServiceInterface
	AuthMiddleware   gin.HandlerFunc
	APIKeyMiddleware gin.HandlerFunc
}

func configureMaterialRoutes(
	router *gin.RouterGroup,
	deps MaterialRouterDependencies,
) {
	authMiddleware := deps.AuthMiddleware
	apiKeyMiddleware := deps.APIKeyMiddleware
	endpoint := endpoints.NewMaterialEndpoint(deps.Service)
	apiCompatibleAuthMiddleware := middlewares.EitherMiddleware(
		[]gin.HandlerFunc{authMiddleware},
		[]gin.HandlerFunc{apiKeyMiddleware},
		func(ctx *gin.Context) bool { return contexts.IsClientGateway(ctx.Request.Context()) },
	)[0]

	materialRoutes := router.Group("/materials")
	{
		materialRoutes.POST(
			"/get-by-id",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMyMaterialByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetMyMaterialById,
		)
		materialRoutes.POST(
			"/get-and-parent-by-id",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMyMaterialAndItsParentByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetMyMaterialAndItsParentById,
		)
		materialRoutes.POST(
			"/get-by-parent-sub-shelf-id",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMyMaterialsByParentSubShelfIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetMyMaterialsByParentSubShelfId,
		)
		materialRoutes.POST(
			"/get-all-by-root-shelf-id",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetAllMyMaterialsByRootShelfIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetAllMyMaterialsByRootShelfId,
		)
		materialRoutes.POST(
			"/create",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.CreateMyMaterialOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.CreateMyMaterial,
		)
		materialRoutes.POST(
			"/update",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.UpdateMyMaterialByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.UpdateMyMaterialById,
		)
		materialRoutes.POST(
			"/save",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.SaveMyMaterialByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.SaveMyMaterialById,
		)
		materialRoutes.POST(
			"/move",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.MoveMyMaterialByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.MoveMyMaterialById,
		)
		materialRoutes.POST(
			"/move-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.MoveMyMaterialsByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.MoveMyMaterialsByIds,
		)
		materialRoutes.POST(
			"/restore",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.RestoreMyMaterialByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.RestoreMyMaterialById,
		)
		materialRoutes.POST(
			"/restore-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.RestoreMyMaterialsByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.RestoreMyMaterialsByIds,
		)
		materialRoutes.POST(
			"/delete",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.DeleteMyMaterialByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.DeleteMyMaterialById,
		)
		materialRoutes.POST(
			"/delete-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.DeleteMyMaterialsByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.DeleteMyMaterialsByIds,
		)
		materialRoutes.POST(
			"/graphql/search",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.SearchMaterialsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.SearchMaterials,
		)
	}
}
