package binders

import (
	"github.com/gin-gonic/gin"

	cnotifications "github.com/HiIamJeff67/notegic-backend/contracts/notification/v1/api"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	controllers "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/controllers"
)

type NotificationBinderInterface interface {
	BindSearch(controllers.Func[*cnotifications.SearchPrivateNotificationsRequestDto]) gin.HandlerFunc
	BindCountUnread(controllers.Func[*cnotifications.CountUnreadNotificationsRequestDto]) gin.HandlerFunc
	BindMarkRead(controllers.Func[*cnotifications.MarkNotificationsReadRequestDto]) gin.HandlerFunc
	BindDelete(controllers.Func[*cnotifications.DeleteNotificationsRequestDto]) gin.HandlerFunc
}

type NotificationBinder struct{}

func NewNotificationBinder() NotificationBinderInterface {
	return &NotificationBinder{}
}

func (b *NotificationBinder) BindSearch(
	controllerFunc controllers.Func[*cnotifications.SearchPrivateNotificationsRequestDto],
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &cnotifications.SearchPrivateNotificationsRequestDto{}
		if err := ctx.ShouldBindQuery(requestDto); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Notification").WithOrigin(err), ctx)
			return
		}
		controllerFunc(ctx, requestDto)
	}
}

func (b *NotificationBinder) BindCountUnread(
	controllerFunc controllers.Func[*cnotifications.CountUnreadNotificationsRequestDto],
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		controllerFunc(ctx, &cnotifications.CountUnreadNotificationsRequestDto{})
	}
}

func (b *NotificationBinder) BindMarkRead(
	controllerFunc controllers.Func[*cnotifications.MarkNotificationsReadRequestDto],
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &cnotifications.MarkNotificationsReadRequestDto{}
		if err := ctx.ShouldBindJSON(requestDto); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Notification").WithOrigin(err), ctx)
			return
		}
		controllerFunc(ctx, requestDto)
	}
}

func (b *NotificationBinder) BindDelete(
	controllerFunc controllers.Func[*cnotifications.DeleteNotificationsRequestDto],
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestDto := &cnotifications.DeleteNotificationsRequestDto{}
		if err := ctx.ShouldBindJSON(requestDto); err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidDto("Notification").WithOrigin(err), ctx)
			return
		}
		controllerFunc(ctx, requestDto)
	}
}
