package developmentroutes

import (
	"time"

	"github.com/gin-gonic/gin"

	scookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"

	ratelimit "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/ratelimit"
	binder "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/binders"
	controllers "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/controllers"
	interceptors "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/interceptors"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/middlewares"
	notificationadapters "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/notification/adapters"
)

type NotificationRouteDependencies struct {
	NotificationClient        *notificationadapters.NotificationAdapter
	AccessTokenCookieHandler  *scookies.CookieHandler
	RefreshTokenCookieHandler *scookies.CookieHandler
	UnauthorizedRateLimiter   *ratelimit.HybridRateLimiter
}

func configureDevelopmentNotificationRoutes(
	router *gin.RouterGroup,
	deps NotificationRouteDependencies,
) {
	notificationClient, accessTokenCookieHandler, refreshTokenCookieHandler, unauthorizedRateLimiter := deps.NotificationClient, deps.AccessTokenCookieHandler, deps.RefreshTokenCookieHandler, deps.UnauthorizedRateLimiter

	notificationBinder := binder.NewNotificationBinder()
	notificationController := controllers.NewNotificationController(notificationClient)
	notificationRoutes := router.Group("/notifications")
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
		notificationRoutes.GET(
			"/",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("searchPrivateNotifications"),
					middlewares.ApplyMeterMiddleware("server.requests.notifications.search"),
				},
				defaultMiddlewares,
				notificationBinder.BindSearch(notificationController.SearchPrivateNotifications),
			)...,
		)
		notificationRoutes.GET(
			"/unread-count",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("countMyUnreadNotifications"),
					middlewares.ApplyMeterMiddleware("server.requests.notifications.unreadCount"),
				},
				defaultMiddlewares,
				notificationBinder.BindCountUnread(notificationController.CountMyUnreadNotifications),
			)...,
		)
		notificationRoutes.PATCH(
			"/read",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("markMyNotificationsRead"),
					middlewares.ApplyMeterMiddleware("server.requests.notifications.read"),
				},
				defaultMiddlewares,
				notificationBinder.BindMarkRead(notificationController.MarkMyNotificationsRead),
			)...,
		)
		notificationRoutes.DELETE(
			"/",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("deleteMyNotifications"),
					middlewares.ApplyMeterMiddleware("server.requests.notifications.delete"),
				},
				defaultMiddlewares,
				notificationBinder.BindDelete(notificationController.DeleteMyNotifications),
			)...,
		)
	}
}
