package developmentroutes

import (
	"time"

	"github.com/gin-gonic/gin"

	scookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"

	binders "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/binders"
	controllers "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/controllers"
	interceptors "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/interceptors"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/middlewares"
	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/core/adapters"
)

type UserRouteDependencies struct {
	CoreAdapter               *coreadapters.CoreAdapter
	AccessTokenCookieHandler  *scookies.CookieHandler
	RefreshTokenCookieHandler *scookies.CookieHandler
	RateLimiters              RateLimiters
}

func configureDevelopmentUserRoutes(
	router *gin.RouterGroup,
	deps UserRouteDependencies,
) {
	coreAdapter, accessTokenCookieHandler, refreshTokenCookieHandler, rateLimiters := deps.CoreAdapter, deps.AccessTokenCookieHandler, deps.RefreshTokenCookieHandler, deps.RateLimiters
	if router == nil {
		router = DevelopmentAPIRouterGroup
	}

	userBinder := binders.NewUserBinder()
	userController := controllers.NewUserController(coreAdapter)

	userRoutes := router.Group("/users")
	defaultMiddlewares := []gin.HandlerFunc{
		middlewares.UnauthorizedRateLimitMiddleware(rateLimiters.Unauthorized),
		middlewares.TimeoutMiddleware(1 * time.Second),
		middlewares.GatewayAuthenticationMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler),
		interceptors.ShareableResponseWriterInterceptor(
			interceptors.RefreshTokenInterceptor(accessTokenCookieHandler),
			interceptors.EmbeddedInterceptor,
		),
	}
	{
		userRoutes.GET(
			"/data",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getUserData"),
					middlewares.ApplyMeterMiddleware("server.requests.user.getUserData"),
				},
				defaultMiddlewares,
				userBinder.BindGetUserData(userController.GetUserData),
			)...,
		)
		userRoutes.GET(
			"/me",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getMe"),
					middlewares.ApplyMeterMiddleware("server.requests.user.getMe"),
				},
				defaultMiddlewares,
				userBinder.BindGetMe(userController.GetMe),
			)...,
		)
		userRoutes.PUT(
			"/me",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("updateMe"),
					middlewares.ApplyMeterMiddleware("server.requests.user.updateMe"),
				},
				defaultMiddlewares,
				userBinder.BindUpdateMe(userController.UpdateMe),
			)...,
		)
	}
}
