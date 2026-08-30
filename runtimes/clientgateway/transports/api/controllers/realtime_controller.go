package controllers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/realtime"

	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/core/adapters"
)

type RealtimeControllerInterface interface {
	CreateMyRealtimeConnectionTicket(ctx *gin.Context, requestDto *capi.CreateMyRealtimeConnectionTicketRequestDto)
	CreateMyBlockPackChannelTicket(ctx *gin.Context, requestDto *capi.CreateMyBlockPackChannelTicketRequestDto)
}

type RealtimeController struct {
	coreAdapter *coreadapters.CoreAdapter
}

func NewRealtimeController(
	coreAdapter *coreadapters.CoreAdapter,
) RealtimeControllerInterface {
	return &RealtimeController{
		coreAdapter: coreAdapter,
	}
}

func (c *RealtimeController) CreateMyRealtimeConnectionTicket(
	ctx *gin.Context,
	requestDto *capi.CreateMyRealtimeConnectionTicketRequestDto,
) {
	response, exception := coreadapters.CallSecurly[
		capi.CreateMyRealtimeConnectionTicketRequestDto,
		capi.CreateMyRealtimeConnectionTicketResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.CreateMyRealtimeConnectionTicketOperation,
		"/core/v1/realtime/connection-ticket/create",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeCreatedClientResponse(ctx, response.Data)
}

func (c *RealtimeController) CreateMyBlockPackChannelTicket(
	ctx *gin.Context,
	requestDto *capi.CreateMyBlockPackChannelTicketRequestDto,
) {
	response, exception := coreadapters.CallSecurly[
		capi.CreateMyBlockPackChannelTicketRequestDto,
		capi.CreateMyBlockPackChannelTicketResponseDto,
	](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.CreateMyBlockPackChannelTicketOperation,
		"/core/v1/realtime/block-pack-channel-ticket/create",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeCreatedClientResponse(ctx, response.Data)
}
