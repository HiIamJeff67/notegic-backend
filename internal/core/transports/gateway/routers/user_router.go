package routers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/users"

	userservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/user"
	endpoints "github.com/HiIamJeff67/notegic-backend/internal/core/transports/gateway/endpoints"
	middlewares "github.com/HiIamJeff67/notegic-backend/internal/core/transports/gateway/middlewares"
)

type UserRouterDependencies struct {
	Service        userservices.UserServiceInterface
	AuthMiddleware gin.HandlerFunc
}

func configureUserRoutes(
	router *gin.RouterGroup,
	deps UserRouterDependencies,
) {
	authMiddleware := deps.AuthMiddleware
	endpoint := endpoints.NewUserEndpoint(deps.Service)
	routes := router.Group("/users")
	{
		routes.POST(
			"/data",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetUserDataOperation,
			),
			authMiddleware,
			endpoint.GetUserData,
		)
		routes.POST(
			"/me",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMeOperation,
			),
			authMiddleware,
			endpoint.GetMe,
		)
		routes.POST(
			"/me/update",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.UpdateMeOperation,
			),
			authMiddleware,
			endpoint.UpdateMe,
		)
		routes.POST(
			"/graphql/search",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.SearchUsersOperation,
			),
			authMiddleware,
			endpoint.SearchUsers,
		)
		routes.POST(
			"/graphql/load-theme-authors",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.LoadThemeAuthorsOperation,
			),
			authMiddleware,
			endpoint.LoadThemeAuthors,
		)
	}
}
