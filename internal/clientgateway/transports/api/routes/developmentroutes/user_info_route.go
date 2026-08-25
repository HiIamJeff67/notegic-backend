package developmentroutes

import (
	"time"

	"github.com/gin-gonic/gin"

	scookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"

	binders "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/binders"
	controllers "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/controllers"
	interceptors "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/interceptors"
	middlewares "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/middlewares"
	coreadapters "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/core/adapters"
)

type UserInfoRouteDependencies struct {
	CoreAdapter               *coreadapters.CoreAdapter
	AccessTokenCookieHandler  *scookies.CookieHandler
	RefreshTokenCookieHandler *scookies.CookieHandler
	RateLimiters              RateLimiters
}

func configureDevelopmentUserInfoRoutes(
	router *gin.RouterGroup,
	deps UserInfoRouteDependencies,
) {
	coreAdapter, accessTokenCookieHandler, refreshTokenCookieHandler, rateLimiters := deps.CoreAdapter, deps.AccessTokenCookieHandler, deps.RefreshTokenCookieHandler, deps.RateLimiters
	if router == nil {
		router = DevelopmentAPIRouterGroup
	}

	userInfoBinder := binders.NewUserInfoBinder()
	userInfoController := controllers.NewUserInfoController(coreAdapter)

	userInfoRoutes := router.Group("/me/info")
	defaultsMiddlewares := []gin.HandlerFunc{
		middlewares.UnauthorizedRateLimitMiddleware(rateLimiters.Unauthorized),
		middlewares.TimeoutMiddleware(1 * time.Second),
		middlewares.GatewayAuthenticationMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler),
		interceptors.ShareableResponseWriterInterceptor(
			interceptors.RefreshTokenInterceptor(accessTokenCookieHandler),
			interceptors.EmbeddedInterceptor,
		),
	}
	{
		userInfoRoutes.GET(
			"/",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getMyInfo"),
					middlewares.ApplyMeterMiddleware("server.requests.userInfo.getMyInfo"),
				},
				defaultsMiddlewares,
				userInfoBinder.BindGetMyInfo(userInfoController.GetMyInfo),
			)...,
		)
		userInfoRoutes.PUT(
			"/",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("updateMyInfo"),
					middlewares.ApplyMeterMiddleware("server.requests.userInfo.updateMyInfo"),
				},
				defaultsMiddlewares,
				userInfoBinder.BindUpdateMyInfo(userInfoController.UpdateMyInfo),
			)...,
		)
	}
}
