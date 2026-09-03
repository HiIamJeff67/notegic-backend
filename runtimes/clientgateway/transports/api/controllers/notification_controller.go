package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"
	cnotifications "github.com/HiIamJeff67/notegic-backend/contracts/notification/v1/api"

	sharedcontexts "github.com/HiIamJeff67/notegic-backend/shared/lib/contexts"
	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	gatewaycontexts "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/contexts"
	notificationadapters "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/notification/adapters"
)

type NotificationControllerInterface interface {
	SearchPrivateNotifications(ctx *gin.Context, requestDto *cnotifications.SearchPrivateNotificationsRequestDto)
	CountMyUnreadNotifications(ctx *gin.Context, requestDto *cnotifications.CountUnreadNotificationsRequestDto)
	MarkMyNotificationsRead(ctx *gin.Context, requestDto *cnotifications.MarkNotificationsReadRequestDto)
	DeleteMyNotifications(ctx *gin.Context, requestDto *cnotifications.DeleteNotificationsRequestDto)
}

type NotificationController struct {
	notificationClient *notificationadapters.NotificationAdapter
}

func NewNotificationController(
	notificationClient *notificationadapters.NotificationAdapter,
) NotificationControllerInterface {
	return &NotificationController{notificationClient: notificationClient}
}

func (c *NotificationController) SearchPrivateNotifications(
	ctx *gin.Context,
	requestDto *cnotifications.SearchPrivateNotificationsRequestDto,
) {
	recipientUserPublicId, exception := gatewaycontexts.GetAndConvertContextFieldToUUID(
		ctx,
		sharedcontexts.ContextFieldName_User_PublicId,
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	requestDto.RecipientUserPublicId = *recipientUserPublicId

	response, exception := notificationadapters.CallSecurly[
		cnotifications.SearchPrivateNotificationsRequestDto,
		cnotifications.SearchPrivateNotificationsResponseDto,
	](ctx, c.notificationClient, requestDto, cnotifications.SearchPrivateNotificationsOperation, "/runtimes/v1/notifications/search")
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	ctx.JSON(
		http.StatusOK,
		cgateway.ClientResponse[any]{
			Success: true,
			Data:    response.Data,
		},
	)
}

func (c *NotificationController) CountMyUnreadNotifications(
	ctx *gin.Context,
	requestDto *cnotifications.CountUnreadNotificationsRequestDto,
) {
	recipientUserPublicId, exception := gatewaycontexts.GetAndConvertContextFieldToUUID(
		ctx,
		sharedcontexts.ContextFieldName_User_PublicId,
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	requestDto.RecipientUserPublicId = *recipientUserPublicId

	response, exception := notificationadapters.CallSecurly[
		cnotifications.CountUnreadNotificationsRequestDto,
		cnotifications.CountUnreadNotificationsResponseDto,
	](ctx, c.notificationClient, requestDto, cnotifications.CountMyUnreadNotificationsOperation, "/runtimes/v1/notifications/unread-count")
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	ctx.JSON(
		http.StatusOK,
		cgateway.ClientResponse[any]{
			Success: true,
			Data:    response.Data,
		},
	)
}

func (c *NotificationController) MarkMyNotificationsRead(
	ctx *gin.Context,
	requestDto *cnotifications.MarkNotificationsReadRequestDto,
) {
	recipientUserPublicId, exception := gatewaycontexts.GetAndConvertContextFieldToUUID(
		ctx,
		sharedcontexts.ContextFieldName_User_PublicId,
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	requestDto.RecipientUserPublicId = *recipientUserPublicId

	response, exception := notificationadapters.CallSecurly[
		cnotifications.MarkNotificationsReadRequestDto,
		cnotifications.MarkNotificationsReadResponseDto,
	](ctx, c.notificationClient, requestDto, cnotifications.MarkMyNotificationsReadOperation, "/runtimes/v1/notifications/read")
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	ctx.JSON(
		http.StatusOK,
		cgateway.ClientResponse[any]{
			Success: true,
			Data:    response.Data,
		},
	)
}

func (c *NotificationController) DeleteMyNotifications(
	ctx *gin.Context,
	requestDto *cnotifications.DeleteNotificationsRequestDto,
) {
	recipientUserPublicId, exception := gatewaycontexts.GetAndConvertContextFieldToUUID(
		ctx,
		sharedcontexts.ContextFieldName_User_PublicId,
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	requestDto.RecipientUserPublicId = *recipientUserPublicId

	response, exception := notificationadapters.CallSecurly[
		cnotifications.DeleteNotificationsRequestDto,
		cnotifications.DeleteNotificationsResponseDto,
	](ctx, c.notificationClient, requestDto, cnotifications.DeleteMyNotificationsOperation, "/runtimes/v1/notifications/delete")
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	ctx.JSON(
		http.StatusOK,
		cgateway.ClientResponse[any]{
			Success: true,
			Data:    response.Data,
		},
	)
}
