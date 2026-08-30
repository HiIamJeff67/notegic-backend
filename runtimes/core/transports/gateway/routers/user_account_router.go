package routers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/user-accounts"

	userservices "github.com/HiIamJeff67/notegic-backend/runtimes/core/services/user"
	endpoints "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/endpoints"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/middlewares"
)

type UserAccountRouterDependencies struct {
	Service        userservices.UserAccountServiceInterface
	AuthMiddleware gin.HandlerFunc
}

func configureUserAccountRoutes(
	router *gin.RouterGroup,
	deps UserAccountRouterDependencies,
) {
	authMiddleware := deps.AuthMiddleware
	endpoint := endpoints.NewUserAccountEndpoint(deps.Service)
	userAccountRoutes := router.Group("/user-accounts")
	{
		userAccountRoutes.POST(
			"/get",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMyAccountOperation,
			),
			authMiddleware,
			endpoint.GetMyAccount,
		)
		userAccountRoutes.POST(
			"/update",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.UpdateMyAccountOperation,
			),
			authMiddleware,
			endpoint.UpdateMyAccount,
		)
		userAccountRoutes.POST(
			"/google/bind",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.BindGoogleAccountOperation,
			),
			authMiddleware,
			endpoint.BindGoogleAccount,
		)
		userAccountRoutes.POST(
			"/google/unbind",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.UnbindGoogleAccountOperation,
			),
			authMiddleware,
			endpoint.UnbindGoogleAccount,
		)
	}
}
