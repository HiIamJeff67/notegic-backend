package routers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/auth"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	userdata "github.com/HiIamJeff67/notegic-backend/internal/core/data/redis/userdata"
	authservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/auth"
	endpoints "github.com/HiIamJeff67/notegic-backend/internal/core/transports/gateway/endpoints"
	middlewares "github.com/HiIamJeff67/notegic-backend/internal/core/transports/gateway/middlewares"
)

type AuthRouterDependencies struct {
	Service             authservices.AuthServiceInterface
	AuthMiddleware      gin.HandlerFunc
	UserDataCacheClient *userdata.UserDataCacheClient
}

func configureAnonymousAuthRoutes(
	router *gin.RouterGroup,
	deps AuthRouterDependencies,
) {
	endpoint := endpoints.NewAuthEndpoint(deps.Service)
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST(
			"/register",
			middlewares.DelegationMiddleware(
				capi.RegisterOperation,
			),
			endpoint.Register,
		)
		authRoutes.POST(
			"/register/google",
			middlewares.DelegationMiddleware(
				capi.RegisterViaGoogleOperation,
			),
			endpoint.RegisterViaGoogle,
		)
		authRoutes.POST(
			"/login",
			middlewares.DelegationMiddleware(
				capi.LoginOperation,
			),
			endpoint.Login,
		)
		authRoutes.POST(
			"/login/google",
			middlewares.DelegationMiddleware(
				capi.LoginViaGoogleOperation,
			),
			endpoint.LoginViaGoogle,
		)
		authRoutes.POST(
			"/email/code",
			middlewares.DelegationMiddleware(
				capi.SendAuthCodeOperation,
			),
			endpoint.SendAuthCode,
		)
		authRoutes.PUT(
			"/password/forget",
			middlewares.DelegationMiddleware(
				capi.ForgetPasswordOperation,
			),
			endpoint.ForgetPassword,
		)
	}
}

func configureAuthenticatedAuthRoutes(
	router *gin.RouterGroup,
	deps AuthRouterDependencies,
) {
	authMiddleware := deps.AuthMiddleware
	endpoint := endpoints.NewAuthEndpoint(deps.Service)
	userDataCacheClient := deps.UserDataCacheClient
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST(
			"/logout",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.LogoutOperation,
			),
			authMiddleware,
			endpoint.Logout,
		)
		authRoutes.PUT(
			"/email/validate",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.ValidateEmailOperation,
			),
			authMiddleware,
			middlewares.CSRFMiddleware(userDataCacheClient),
			endpoint.ValidateEmail,
		)
		authRoutes.PUT(
			"/email/reset",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.ResetEmailOperation,
			),
			authMiddleware,
			middlewares.UserRoleMiddleware(enums.UserRole_Normal),
			middlewares.CSRFMiddleware(userDataCacheClient),
			endpoint.ResetEmail,
		)
		authRoutes.PUT(
			"/me/reset",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.ResetMeOperation,
			),
			authMiddleware,
			middlewares.CSRFMiddleware(userDataCacheClient),
			endpoint.ResetMe,
		)
		authRoutes.DELETE(
			"/me/delete",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.DeleteMeOperation,
			),
			authMiddleware,
			middlewares.CSRFMiddleware(userDataCacheClient),
			endpoint.DeleteMe,
		)
	}
}
