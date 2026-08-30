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

type APIKeyRouteDependencies struct {
	CoreAdapter               *coreadapters.CoreAdapter
	AccessTokenCookieHandler  *scookies.CookieHandler
	RefreshTokenCookieHandler *scookies.CookieHandler
	RateLimiters              RateLimiters
}

func configureDevelopmentAPIKeyRoutes(
	router *gin.RouterGroup,
	deps APIKeyRouteDependencies,
) {
	coreAdapter, accessTokenCookieHandler, refreshTokenCookieHandler, rateLimiters := deps.CoreAdapter, deps.AccessTokenCookieHandler, deps.RefreshTokenCookieHandler, deps.RateLimiters
	binder := binders.NewAPIKeyBinder()
	controller := controllers.NewAPIKeyController(coreAdapter)
	routes := router.Group("/me/api-keys")
	defaultMiddlewares := []gin.HandlerFunc{
		middlewares.UnauthorizedRateLimitMiddleware(rateLimiters.Unauthorized),
		middlewares.TimeoutMiddleware(3 * time.Second),
		middlewares.GatewayAuthenticationMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler),
		interceptors.ShareableResponseWriterInterceptor(
			interceptors.RefreshTokenInterceptor(accessTokenCookieHandler),
			interceptors.EmbeddedInterceptor,
		),
	}
	{
		routes.POST(
			"/create",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("createMyAPIKey"),
					middlewares.ApplyMeterMiddleware("server.requests.apiKey.create"),
				},
				defaultMiddlewares,
				binder.BindCreateMyAPIKey(controller.CreateMyAPIKey),
			)...,
		)
		routes.GET(
			"/",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("listMyAPIKeys"),
					middlewares.ApplyMeterMiddleware("server.requests.apiKey.list"),
				},
				defaultMiddlewares,
				binder.BindListMyAPIKeys(controller.ListMyAPIKeys),
			)...,
		)
		routes.DELETE(
			"/:api-key-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("revokeMyAPIKey"),
					middlewares.ApplyMeterMiddleware("server.requests.apiKey.revoke"),
				},
				defaultMiddlewares,
				binder.BindRevokeMyAPIKey(controller.RevokeMyAPIKey),
			)...,
		)
	}
}
