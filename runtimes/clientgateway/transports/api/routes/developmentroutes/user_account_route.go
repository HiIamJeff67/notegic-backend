package developmentroutes

import (
	"time"

	"github.com/gin-gonic/gin"

	scookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"

	ratelimit "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/ratelimit"
	binders "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/binders"
	controllers "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/controllers"
	interceptors "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/interceptors"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/middlewares"
	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/core/adapters"
)

type UserAccountRouteDependencies struct {
	CoreAdapter               *coreadapters.CoreAdapter
	AccessTokenCookieHandler  *scookies.CookieHandler
	RefreshTokenCookieHandler *scookies.CookieHandler
	UnauthorizedRateLimiter   *ratelimit.HybridRateLimiter
}

func configureDevelopmentUserAccountRoutes(
	router *gin.RouterGroup,
	deps UserAccountRouteDependencies,
) {
	coreAdapter, accessTokenCookieHandler, refreshTokenCookieHandler, unauthorizedRateLimiter := deps.CoreAdapter, deps.AccessTokenCookieHandler, deps.RefreshTokenCookieHandler, deps.UnauthorizedRateLimiter

	userAccountBinder := binders.NewUserAccountBinder()
	userAccountController := controllers.NewUserAccountController(coreAdapter)

	userAccountRoutes := router.Group("/me/account")
	defaultMiddlewares := []gin.HandlerFunc{
		middlewares.UnauthorizedRateLimitMiddleware(unauthorizedRateLimiter),
		middlewares.TimeoutMiddleware(3 * time.Second),
		middlewares.GatewayAuthenticationMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler),
		interceptors.ShareableResponseWriterInterceptor(
			interceptors.RefreshTokenInterceptor(accessTokenCookieHandler),
			interceptors.EmbeddedInterceptor,
		),
	}
	{
		userAccountRoutes.GET(
			"/",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getMyAccount"),
					middlewares.ApplyMeterMiddleware("server.requests.userAccount.getMyAccount"),
				},
				defaultMiddlewares,
				userAccountBinder.BindGetMyAccount(userAccountController.GetMyAccount),
			)...,
		)
		userAccountRoutes.PUT(
			"/",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("updateMyAccount"),
					middlewares.ApplyMeterMiddleware("server.requests.userAccount.updateMyAccount"),
				},
				defaultMiddlewares,
				userAccountBinder.BindUpdateMyAccount(userAccountController.UpdateMyAccount),
			)...,
		)
		userAccountRoutes.PUT(
			"/google",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("bindGoogleAccount"),
					middlewares.ApplyMeterMiddleware("server.requests.userAccount.bindGoogleAccount"),
				},
				defaultMiddlewares,
				userAccountBinder.BindBindGoogleAccount(userAccountController.BindGoogleAccount),
			)...,
		)
		userAccountRoutes.DELETE(
			"/google",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("unbindGoogleAccount"),
					middlewares.ApplyMeterMiddleware("server.requests.userAccount.unbindGoogleAccount"),
				},
				defaultMiddlewares,
				userAccountBinder.BindUnbindGoogleAccount(userAccountController.UnbindGoogleAccount),
			)...,
		)
	}
}
