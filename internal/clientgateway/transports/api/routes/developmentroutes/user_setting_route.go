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

type UserSettingRouteDependencies struct {
	CoreAdapter               *coreadapters.CoreAdapter
	AccessTokenCookieHandler  *scookies.CookieHandler
	RefreshTokenCookieHandler *scookies.CookieHandler
	RateLimiters              RateLimiters
}

func configureUserSettingRoutes(
	router *gin.RouterGroup,
	deps UserSettingRouteDependencies,
) {
	coreAdapter, accessTokenCookieHandler, refreshTokenCookieHandler, rateLimiters := deps.CoreAdapter, deps.AccessTokenCookieHandler, deps.RefreshTokenCookieHandler, deps.RateLimiters
	if router == nil {
		router = DevelopmentAPIRouterGroup
	}

	userSettingBinder := binders.NewUserSettingBinder()
	userSettingController := controllers.NewUserSettingController(coreAdapter)

	userSettingRoutes := router.Group("/me/settings")
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
