package controllers

import (
	"github.com/gin-gonic/gin"

	cnotifications "github.com/HiIamJeff67/notegic-backend/contracts/notification/v1/api"
	sharedcontexts "github.com/HiIamJeff67/notegic-backend/shared/lib/contexts"

	exceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	gatewaycontexts "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/contexts"
	notificationadapters "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/notification/adapters"
)

type NotificationControllerInterface interface {
	Search(ctx *gin.Context, requestDto *cnotifications.SearchPrivateNotificationsRequestDto)
	CountUnread(ctx *gin.Context, requestDto *cnotifications.CountUnreadNotificationsRequestDto)
	MarkRead(ctx *gin.Context, requestDto *cnotifications.MarkNotificationsReadRequestDto)
	Delete(ctx *gin.Context, requestDto *cnotifications.DeleteNotificationsRequestDto)
}

type NotificationController struct {
	notificationClient *notificationadapters.NotificationAdapter
}

func NewNotificationController(
	notificationClient *notificationadapters.NotificationAdapter,
) NotificationControllerInterface {
	return &NotificationController{notificationClient: notificationClient}
}

func (c *NotificationController) Search(
	ctx *gin.Context,
	requestDto *cnotifications.SearchPrivateNotificationsRequestDto,
) {
	recipientUserPublicId, exception := gatewaycontexts.GetAndConvertContextFieldToUUID(
		ctx,
		sharedcontexts.ContextFieldName_User_PublicId,
	)
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	requestDto.RecipientUserPublicId = *recipientUserPublicId

	response, exception := notificationadapters.CallSecurly[
		cnotifications.SearchPrivateNotificationsRequestDto,
		cnotifications.SearchPrivateNotificationsResponseDto,
	](ctx, c.notificationClient, requestDto, cnotifications.SearchPrivateNotificationsOperation, "/internal/v1/notifications/search")
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	writeClientResponse(ctx, response.Data)
}

func (c *NotificationController) CountUnread(
	ctx *gin.Context,
	requestDto *cnotifications.CountUnreadNotificationsRequestDto,
) {
	recipientUserPublicId, exception := gatewaycontexts.GetAndConvertContextFieldToUUID(
		ctx,
		sharedcontexts.ContextFieldName_User_PublicId,
	)
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	requestDto.RecipientUserPublicId = *recipientUserPublicId

	response, exception := notificationadapters.CallSecurly[
		cnotifications.CountUnreadNotificationsRequestDto,
		cnotifications.CountUnreadNotificationsResponseDto,
	](ctx, c.notificationClient, requestDto, cnotifications.CountMyUnreadNotificationsOperation, "/internal/v1/notifications/unread-count")
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	writeClientResponse(ctx, response.Data)
}

func (c *NotificationController) MarkRead(
	ctx *gin.Context,
	requestDto *cnotifications.MarkNotificationsReadRequestDto,
) {
	recipientUserPublicId, exception := gatewaycontexts.GetAndConvertContextFieldToUUID(
		ctx,
		sharedcontexts.ContextFieldName_User_PublicId,
	)
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	requestDto.RecipientUserPublicId = *recipientUserPublicId

	response, exception := notificationadapters.CallSecurly[
		cnotifications.MarkNotificationsReadRequestDto,
		cnotifications.MarkNotificationsReadResponseDto,
	](ctx, c.notificationClient, requestDto, cnotifications.MarkMyNotificationsReadOperation, "/internal/v1/notifications/read")
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	writeClientResponse(ctx, response.Data)
}

func (c *NotificationController) Delete(
	ctx *gin.Context,
	requestDto *cnotifications.DeleteNotificationsRequestDto,
) {
	recipientUserPublicId, exception := gatewaycontexts.GetAndConvertContextFieldToUUID(
		ctx,
		sharedcontexts.ContextFieldName_User_PublicId,
	)
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	requestDto.RecipientUserPublicId = *recipientUserPublicId

	response, exception := notificationadapters.CallSecurly[
		cnotifications.DeleteNotificationsRequestDto,
		cnotifications.DeleteNotificationsResponseDto,
	](ctx, c.notificationClient, requestDto, cnotifications.DeleteMyNotificationsOperation, "/internal/v1/notifications/delete")
	if exception != nil {
		exceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}
	writeClientResponse(ctx, response.Data)
}
