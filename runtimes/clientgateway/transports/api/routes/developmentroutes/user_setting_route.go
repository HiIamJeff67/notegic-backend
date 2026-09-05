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

type UserSettingRouteDependencies struct {
	CoreAdapter               *coreadapters.CoreAdapter
	AccessTokenCookieHandler  *scookies.CookieHandler
	RefreshTokenCookieHandler *scookies.CookieHandler
	UnauthorizedRateLimiter   *ratelimit.HybridRateLimiter
}

func configureUserSettingRoutes(
	router *gin.RouterGroup,
	deps UserSettingRouteDependencies,
) {
	coreAdapter, accessTokenCookieHandler, refreshTokenCookieHandler, unauthorizedRateLimiter := deps.CoreAdapter, deps.AccessTokenCookieHandler, deps.RefreshTokenCookieHandler, deps.UnauthorizedRateLimiter

	userSettingBinder := binders.NewUserSettingBinder()
	userSettingController := controllers.NewUserSettingController(coreAdapter)

	userSettingRoutes := router.Group("/me/settings")
	defaultMiddlewares := []gin.HandlerFunc{
		middlewares.UnauthorizedRateLimitMiddleware(unauthorizedRateLimiter),
		middlewares.TimeoutMiddleware(1 * time.Second),
		middlewares.GatewayAuthenticationMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler),
		interceptors.ShareableResponseWriterInterceptor(
			interceptors.RefreshTokenInterceptor(accessTokenCookieHandler),
			interceptors.EmbeddedInterceptor,
		),
	}
	{
		userSettingRoutes.GET(
			"/",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getMySetting"),
					middlewares.ApplyMeterMiddleware("server.requests.userSetting.getMySetting"),
				},
				defaultMiddlewares,
				userSettingBinder.BindGetMySetting(userSettingController.GetMySetting),
			)...,
		)
		userSettingRoutes.PUT(
			"/",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("updateMySetting"),
					middlewares.ApplyMeterMiddleware("server.requests.userSetting.updateMySetting"),
				},
				defaultMiddlewares,
				userSettingBinder.BindUpdateMySetting(userSettingController.UpdateMySetting),
			)...,
		)
	}
}
