package routers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routine-tags"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	routineservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/routines"
	endpoints "github.com/HiIamJeff67/notegic-backend/internal/core/transports/gateway/endpoints"
	middlewares "github.com/HiIamJeff67/notegic-backend/internal/core/transports/gateway/middlewares"
)

type RoutineTagRouterDependencies struct {
	Service          routineservices.RoutineTagServiceInterface
	AuthMiddleware   gin.HandlerFunc
	APIKeyMiddleware gin.HandlerFunc
}

func configureRoutineTagRoutes(
	router *gin.RouterGroup,
	deps RoutineTagRouterDependencies,
) {
	authMiddleware := deps.AuthMiddleware
	apiKeyMiddleware := deps.APIKeyMiddleware
	endpoint := endpoints.NewRoutineTagEndpoint(deps.Service)
	apiCompatibleAuthMiddleware := middlewares.EitherMiddleware(
		[]gin.HandlerFunc{authMiddleware},
		[]gin.HandlerFunc{apiKeyMiddleware},
		func(ctx *gin.Context) bool { return contexts.IsClientGateway(ctx.Request.Context()) },
	)[0]

	routineTagRoutes := router.Group("/routine-tags")
	{
		routineTagRoutes.POST(
			"/get-by-id",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMyRoutineTagByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetMyRoutineTagById,
		)
		routineTagRoutes.POST(
			"/get-all",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetAllMyRoutineTagsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.GetAllMyRoutineTags,
		)
		routineTagRoutes.POST(
			"/create",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.CreateRoutineTagOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.CreateRoutineTag,
		)
		routineTagRoutes.POST(
			"/create-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.CreateRoutineTagsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.CreateRoutineTags,
		)
		routineTagRoutes.POST(
			"/update",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.UpdateMyRoutineTagByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.UpdateMyRoutineTagById,
		)
		routineTagRoutes.POST(
			"/update-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.UpdateMyRoutineTagsByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.UpdateMyRoutineTagsByIds,
		)
		routineTagRoutes.POST(
			"/hard-delete",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.HardDeleteMyRoutineTagByIdOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.HardDeleteMyRoutineTagById,
		)
		routineTagRoutes.POST(
			"/hard-delete-many",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.HardDeleteMyRoutineTagsByIdsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.HardDeleteMyRoutineTagsByIds,
		)
		routineTagRoutes.POST(
			"/graphql/search",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.SearchRoutineTagsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.SearchRoutineTags,
		)
	}
}
