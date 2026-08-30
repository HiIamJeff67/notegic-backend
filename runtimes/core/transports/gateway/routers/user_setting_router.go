package routers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/user-settings"

	userservices "github.com/HiIamJeff67/notegic-backend/runtimes/core/services/user"
	endpoints "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/endpoints"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/middlewares"
)

type UserSettingRouterDependencies struct {
	Service        userservices.UserSettingServiceInterface
	AuthMiddleware gin.HandlerFunc
}

func configureUserSettingRoutes(
	router *gin.RouterGroup,
	deps UserSettingRouterDependencies,
) {
	authMiddleware := deps.AuthMiddleware
	endpoint := endpoints.NewUserSettingEndpoint(deps.Service)
	userSettingRoutes := router.Group("/user-settings")
	{
		userSettingRoutes.POST(
			"/get",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMySettingOperation,
			),
			authMiddleware,
			endpoint.GetMySetting,
		)
		userSettingRoutes.POST(
			"/update",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.UpdateMySettingOperation,
			),
			authMiddleware,
			endpoint.UpdateMySetting,
		)
	}
}
