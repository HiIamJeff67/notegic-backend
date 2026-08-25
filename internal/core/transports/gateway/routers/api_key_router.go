package routers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/api-keys"

	apikeyservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/apikey"
	endpoints "github.com/HiIamJeff67/notegic-backend/internal/core/transports/gateway/endpoints"
	middlewares "github.com/HiIamJeff67/notegic-backend/internal/core/transports/gateway/middlewares"
)

type APIKeyRouterDependencies struct {
	Service        apikeyservices.APIKeyServiceInterface
	AuthMiddleware gin.HandlerFunc
}

func configureAPIKeyRoutes(
	router *gin.RouterGroup,
	deps APIKeyRouterDependencies,
) {
	authMiddleware := deps.AuthMiddleware
	endpoint := endpoints.NewAPIKeyEndpoint(deps.Service)
	routes := router.Group("/api-keys")
	{
		routes.POST(
			"/create",
			middlewares.DelegationAuthenticatedMiddleware(capi.CreateMyAPIKeyOperation),
			authMiddleware,
			endpoint.CreateMyAPIKey,
		)
		routes.POST(
			"/list",
			middlewares.DelegationAuthenticatedMiddleware(capi.ListMyAPIKeysOperation),
			authMiddleware,
			endpoint.ListMyAPIKeys,
		)
		routes.POST(
			"/revoke",
			middlewares.DelegationAuthenticatedMiddleware(capi.RevokeMyAPIKeyOperation),
			authMiddleware,
			endpoint.RevokeMyAPIKey,
		)
	}
}
