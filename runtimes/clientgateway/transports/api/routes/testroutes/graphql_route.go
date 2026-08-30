package testroutes

import (
	"github.com/gin-gonic/gin"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	scookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"

	graphql "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/graphql"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/middlewares"
	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/core/adapters"
)

type GraphQLRouteDependencies struct {
	CoreAdapter               *coreadapters.CoreAdapter
	AccessTokenCookieHandler  *scookies.CookieHandler
	RefreshTokenCookieHandler *scookies.CookieHandler
}

func ConfigureTestGraphQLRoutes(
	routerGroup *gin.RouterGroup,
	deps GraphQLRouteDependencies,
) {
	coreAdapter, accessTokenCookieHandler, refreshTokenCookieHandler := deps.CoreAdapter, deps.AccessTokenCookieHandler, deps.RefreshTokenCookieHandler
	graphqlRoutes := routerGroup.Group("/graphql")

	graphqlRoutes.Use(
		middlewares.JWTMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler),
		middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
	)
	{
		graphqlRoutes.POST("/", graphql.GraphQLHandler(coreAdapter))
		if gin.Mode() == gin.DebugMode {
			graphqlRoutes.GET("/", graphql.PlaygroundHandler())
		}
	}
}
