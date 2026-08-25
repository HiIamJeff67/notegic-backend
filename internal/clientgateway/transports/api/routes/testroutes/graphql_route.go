package testroutes

import (
	"github.com/gin-gonic/gin"

	cookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	graphql "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/graphql"
	middlewares "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/middlewares"
	coreadapters "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/core/adapters"
)

type GraphQLRouteDependencies struct {
	CoreAdapter               *coreadapters.CoreAdapter
	AccessTokenCookieHandler  *cookies.CookieHandler
	RefreshTokenCookieHandler *cookies.CookieHandler
}

func ConfigureTestGraphQLRoutes(
	routerGroup *gin.RouterGroup,
	deps GraphQLRouteDependencies,
) {
	coreAdapter, accessTokenCookieHandler, refreshTokenCookieHandler := deps.CoreAdapter, deps.AccessTokenCookieHandler, deps.RefreshTokenCookieHandler
	graphqlRoutes := routerGroup.Group("/graphql")

	graphqlRoutes.Use(
		middlewares.JWTMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler),
		middlewares.AllowedPermissionsAbove(enums.AccessControlPermission_Read),
	)
	{
		graphqlRoutes.POST("/", graphql.GraphQLHandler(coreAdapter))
		if gin.Mode() == gin.DebugMode {
			graphqlRoutes.GET("/", graphql.PlaygroundHandler())
		}
	}
}
