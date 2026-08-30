package routers

import (
	"github.com/gin-gonic/gin"

	cnotifications "github.com/HiIamJeff67/notegic-backend/contracts/notification/v1/api"

	endpoints "github.com/HiIamJeff67/notegic-backend/runtimes/notification/transports/gateway/endpoints"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/notification/transports/gateway/middlewares"
)

func ConfigureNotificationRoutes(
	router *gin.RouterGroup,
	endpoint *endpoints.NotificationEndpoint,
) {
	notificationRoutes := router.Group("/notifications")
	notificationRoutes.POST(
		"/search",
		middlewares.DelegationAuthenticatedMiddleware(cnotifications.SearchPrivateNotificationsOperation),
		endpoint.Search,
	)
	notificationRoutes.POST(
		"/unread-count",
		middlewares.DelegationAuthenticatedMiddleware(cnotifications.CountMyUnreadNotificationsOperation),
		endpoint.CountUnread,
	)
	notificationRoutes.POST(
		"/read",
		middlewares.DelegationAuthenticatedMiddleware(cnotifications.MarkMyNotificationsReadOperation),
		endpoint.MarkRead,
	)
	notificationRoutes.POST(
		"/delete",
		middlewares.DelegationAuthenticatedMiddleware(cnotifications.DeleteMyNotificationsOperation),
		endpoint.Delete,
	)
}
